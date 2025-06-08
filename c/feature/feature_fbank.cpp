//
// Created by shuli on 2024/3/4.
//

#include "feature_fbank.h"
#include "feature_functions.h"


speakerlab::FbankComputer::FbankComputer(const speakerlab::FbankOptions &opts) : opts_(opts),
                                                                                 frame_preprocessor_(opts.frame_opts),
                                                                                 log_energy_floor_(0.0),
                                                                                 mel_bank_processor_(opts.mel_opts) {
    int frame_length = opts_.frame_opts.compute_window_size();
    const int fft_n = round_up_to_nearest_power_of_two(frame_length);
    init_sin_tbl(sin_tbl_, fft_n);
    init_bit_reverse_index(bit_rev_index_, fft_n);

    int padded_window_length = opts_.frame_opts.padded_window_size();
    mel_bank_processor_.init_mel_bins(opts.frame_opts.sample_freq, padded_window_length);
}

// 直接从PCM数据提取特征（新接口）
speakerlab::Feature speakerlab::FbankComputer::compute_feature_from_pcm(const short* pcm_data, int pcm_length) {
    // 检查PCM数据有效性
    if (!check_pcm_data(pcm_length)) {
        throw std::invalid_argument("PCM data is invalid");
    }
    
    // 计算相关参数
    int frame_length = opts_.compute_window_size();
    int frame_shift = opts_.compute_window_shift();
    int fft_n = round_up_to_nearest_power_of_two(frame_length);
    
    // 将int16类型的PCM数据转换为float类型
    Wave wav_data(pcm_length);
    for (int i = 0; i < pcm_length; i++) {
        // 将int16转换为float，并进行归一化（除以32768）
        wav_data[i] = static_cast<float>(pcm_data[i]) / 32768.0f;
    }
    
    int num_samples = wav_data.size();
    int num_frames = 1 + ((num_samples - frame_length) / frame_shift);
    
    // 输出调试信息
    // std::cout << "PCM data: " << pcm_length << " samples, "  
    //           << "computing " << num_frames << " frames" 
    //           << std::endl;
    
    // 初始化特征矩阵
    Feature feature;
    feature.resize(num_frames);

    float epsilon = std::numeric_limits<float>::epsilon();
    int fbank_num_bins = opts_.get_fbank_num_bins();
    std::vector<std::pair<int, std::vector<float>>> mel_bins = mel_bank_processor_.get_mel_bins();
    
    for (int i = 0; i < num_frames; i++) {
        std::vector<float> cur_wav_data(wav_data.data() + i * frame_shift,
                                        wav_data.data() + i * frame_shift + frame_length);
        // 包含抖动、预处理等
        frame_preprocessor_.frame_pre_process(cur_wav_data);

        // 构建FFT
        std::vector<std::complex<float>> cur_window_data(fft_n);
        for (int j = 0; j < fft_n; j++) {
            if (j < frame_length) {
                cur_window_data[j] = std::complex<float>(cur_wav_data[j], 0.0);
            } else {
                cur_window_data[j] = std::complex<float>(0.0, 0.0);
            }
        }
        custom_fft(bit_rev_index_, sin_tbl_, cur_window_data);
        std::vector<float> power(fft_n / 2);
        for (int j = 0; j < fft_n / 2; j++) {
            power[j] = cur_window_data[j].real() * cur_window_data[j].real() +
                       cur_window_data[j].imag() * cur_window_data[j].imag();
        }
        if (!opts_.use_power) {
            for (int j = 0; j < fft_n / 2; j++) {
                power[j] = powf(power[j], 0.5);
            }
        }
        // 梯度滤波
        feature[i].resize(opts_.get_fbank_num_bins());
        for (int j = 0; j < fbank_num_bins; j++) {
            float mel_energy = 0.0;
            int start_index = mel_bins[j].first;
            for (int k = 0; k < mel_bins[j].second.size(); k++) {
                mel_energy += mel_bins[j].second[k] * power[k + start_index];
            }
            if (opts_.use_log_fbank) {
                if (mel_energy < epsilon) mel_energy = epsilon;
                mel_energy = logf(mel_energy);
            }
            feature[i][j] = mel_energy;
        }
    }
    return feature;
}

// 检查PCM数据是否有效
// 简化版，只需要确保PCM数据长度足够
// 我们假设输入的PCM数据已经是16kHz单声道的
// 所以不再需要检查采样率和通道数
bool speakerlab::FbankComputer::check_pcm_data(int pcm_length) {
    int window_size = opts_.compute_window_size();
    int window_shift = opts_.compute_window_shift();
    
    // 检查数据长度是否足够
    if (window_size < 2 || window_size > pcm_length) {
        std::cerr << "Choose a window size " << window_size << " that is [2, " << pcm_length << "]"
                  << std::endl;
        return false;
    }
    if (window_shift <= 0) {
        std::cerr << "Window shift " << window_shift << " must be greater than 0" << std::endl;
        return false;
    }
    
    // 其他检查可以从原来的check_wav_and_config函数中复用
    int padded_window_size = opts_.paddle_window_size();
    if (padded_window_size % 2 == 1) {
        std::cerr << "The padded `window_size` must be divisible by two.";
        return false;
    }
    if (opts_.frame_opts.pre_emphasis_coefficient < 0.0 || opts_.frame_opts.pre_emphasis_coefficient > 1.0) {
        std::cerr << "Pre-emphasis coefficient " << opts_.frame_opts.pre_emphasis_coefficient
                  << " must be between [0, 1]" << std::endl;
        return false;
    }
    
    return true;
}
