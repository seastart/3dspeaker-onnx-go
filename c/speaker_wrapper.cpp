#include "speaker_wrapper.h"
#include <cmath>
#include <memory>
#include <iostream>
#include <vector>
#include "model/speaker_embedding_model.h"
#include "feature/feature_fbank.h"

// 内部结构体，用于保存模型和特征提取器
struct SpeakerModelWrapper {
    std::unique_ptr<speakerlab::OnnxSpeakerEmbeddingModel> model;
    std::unique_ptr<speakerlab::FbankComputer> feature_extractor;
    
    // 参数化构造函数
    SpeakerModelWrapper(const std::string& onnx_path, float sample_freq, float frame_shift_ms, float frame_length_ms, 
                        int num_bins, bool use_log, float dither, bool use_power) {
        try {
            // 加载ONNX模型
            model = std::make_unique<speakerlab::OnnxSpeakerEmbeddingModel>(onnx_path);
            
            // 直接使用传入的参数创建FbankOptions
            speakerlab::FbankOptions opts;
            
            // 设置基本的FrameExtractionOptions参数
            opts.frame_opts.sample_freq = sample_freq;
            opts.frame_opts.frame_shift_ms = frame_shift_ms;
            opts.frame_opts.frame_length_ms = frame_length_ms;
            opts.frame_opts.dither = dither;  // 设置抖动参数
            
            // 设置梅尔滤波器组参数
            opts.mel_opts.num_bins = num_bins;
            
            // 设置是否使用对数化FBANK和功率谱
            opts.use_log_fbank = use_log;
            opts.use_power = use_power;
            
            // 创建FbankComputer
            feature_extractor = std::make_unique<speakerlab::FbankComputer>(opts);
            
            // 输出参数信息
            std::cout << "使用传入的参数创建FbankComputer [采样率=" << sample_freq 
                      << ", 帧移=" << frame_shift_ms << "ms"
                      << ", 帧长=" << frame_length_ms << "ms"
                      << ", 滤波器数=" << num_bins
                      << ", 使用对数=" << (use_log ? "是" : "否")
                      << ", 抖动=" << dither
                      << ", 使用功率谱=" << (use_power ? "是" : "否") << "]" << std::endl;
                      
        } catch (const std::exception& e) {
            std::cerr << "初始化失败: " << e.what() << std::endl;
            throw; // 重新抛出异常以便调用者处理
        }
    }
    
    // 直接从PCM数据提取特征
    speakerlab::Feature extractFeatureFromPcm(const short* pcm_data, int pcm_length) {
        if (!pcm_data || pcm_length <= 0) {
            throw std::invalid_argument("无效的PCM数据");
        }
        
        // 调用我们重构后的compute_feature_from_pcm函数
        // 这个函数直接从PCM数据计算FBANK特征
        return feature_extractor->compute_feature_from_pcm(pcm_data, pcm_length);
    }
};

extern "C" {

// 实现加载模型函数
SpeakerModelHandle LoadSpeakerModel(const char* onnx_model_path, 
                                 float sample_freq,
                                 float frame_shift_ms,
                                 float frame_length_ms,
                                 int num_bins,
                                 int use_log,
                                 float dither,
                                 int use_power) {
    if (!onnx_model_path) {
        std::cerr << "无效的模型路径" << std::endl;
        return nullptr;
    }
    
    std::string model_path(onnx_model_path);
    
    try {
        // 创建一个新的SpeakerModelWrapper对象
        auto* wrapper = new SpeakerModelWrapper(model_path, sample_freq, frame_shift_ms, frame_length_ms, 
                                              num_bins, use_log, dither, use_power);
        
        // 输出成功信息
        std::cout << "成功加载模型和特征提取器" << std::endl;
        
        return static_cast<SpeakerModelHandle>(wrapper);
    } catch (const std::exception& e) {
        std::cerr << "加载模型出错: " << e.what() << std::endl;
        return nullptr;
    }
}

// 释放模型资源
void FreeSpeakerModel(SpeakerModelHandle handle) {
    if (handle) {
        auto* wrapper = static_cast<SpeakerModelWrapper*>(handle);
        delete wrapper;
    }
}

// 从所人数据中提取说话人嵌入向量
int ExtractEmbedding(SpeakerModelHandle handle, 
                     const short* pcm_data, 
                     int pcm_length, 
                     float** embedding, 
                     int* embedding_size) {
    if (!handle || !pcm_data || !embedding || !embedding_size || pcm_length <= 0) {
        std::cerr << "ExtractEmbedding参数无效" << std::endl;
        return 0;
    }
    
    auto* wrapper = static_cast<SpeakerModelWrapper*>(handle);
    
    try {
        // 打印调试信息
        // std::cout << "调试信息: PCM长度=" << pcm_length << std::endl;
        
        // 直接从PCM数据提取特征，无需通过WavReader
        // std::cout << "开始从PCM数据提取特征..." << std::endl;
        speakerlab::Feature feature = wrapper->extractFeatureFromPcm(pcm_data, pcm_length);
        
        if (feature.empty()) {
            std::cerr << "特征提取失败" << std::endl;
            return 0;
        }
        
        // 提取嵌入向量
        // std::cout << "开始提取嵌入向量..." << std::endl;
        speakerlab::Embedding emb;
        wrapper->model->extract_embedding(feature, emb);
        
        if (emb.empty()) {
            std::cerr << "嵌入向量提取失败" << std::endl;
            return 0;
        }
        
        // 计算向量的L2范数
        float norm = 0.0f;
        for (int i = 0; i < emb.size(); i++) {
            norm += emb[i] * emb[i];
        }
        norm = std::sqrt(norm);
        
        // 当范数为0时的处理
        if (norm < 1e-10) {
            std::cerr << "警告: 嵌入向量范数接近于0，无法归一化" << std::endl;
            norm = 1.0f; // 避免除以0
        }
        
        // 分配内存并复制嵌入向量（进行归一化）
        int size = emb.size();
        float* output = new float[size];
        for (int i = 0; i < size; i++) {
            output[i] = emb[i] / norm; // 归一化
        }
        
        *embedding = output;
        *embedding_size = size;
        
        // std::cout << "成功提取嵌入向量，维度=" << size << ", 归一化前范数=" << norm << std::endl;
        return 1;
    } catch (const std::exception& e) {
        std::cerr << "提取嵌入向量时出错: " << e.what() << std::endl;
        return 0;
    }
}

// 计算两个嵌入向量的余弦相似度
float ComputeCosineSimilarity(const float* embedding1, int size1, 
                              const float* embedding2, int size2) {
    if (!embedding1 || !embedding2 || size1 <= 0 || size2 <= 0) {
        std::cerr << "ComputeCosineSimilarity参数无效" << std::endl;
        return 0.0f;
    }
    
    if (size1 != size2) {
        std::cerr << "嵌入向量维度不匹配: " << size1 << " vs " << size2 << std::endl;
        return 0.0f;
    }
    
    // 计算点积
    float dot_product = 0.0f;
    float norm1 = 0.0f;
    float norm2 = 0.0f;
    
    for (int i = 0; i < size1; i++) {
        dot_product += embedding1[i] * embedding2[i];
        norm1 += embedding1[i] * embedding1[i];
        norm2 += embedding2[i] * embedding2[i];
    }
    
    if (norm1 <= 0.0f || norm2 <= 0.0f) {
        return 0.0f;
    }
    
    float similarity = dot_product / (std::sqrt(norm1) * std::sqrt(norm2));
    // std::cout << "计算的余弦相似度: " << similarity << std::endl;
    return similarity;
}

// 计算两个嵌入向量的L2距离
float ComputeL2Distance(const float* embedding1, int size1,
                       const float* embedding2, int size2) {
    if (!embedding1 || !embedding2 || size1 <= 0 || size2 <= 0) {
        std::cerr << "ComputeL2Distance参数无效" << std::endl;
        return -1.0f; // 返回负值表示错误
    }
    
    if (size1 != size2) {
        std::cerr << "嵌入向量维度不匹配: " << size1 << " vs " << size2 << std::endl;
        return -1.0f;
    }
    
    // 计算欧氏距离的平方和
    float sum_squared_diff = 0.0f;
    
    for (int i = 0; i < size1; i++) {
        float diff = embedding1[i] - embedding2[i];
        sum_squared_diff += diff * diff;
    }
    
    // 计算平方根得到L2距离
    float distance = std::sqrt(sum_squared_diff);
    // std::cout << "计算的L2距离: " << distance << std::endl;
    return distance;
}

// 释放嵌入向量内存
void FreeEmbedding(float* embedding) {
    if (embedding) {
        delete[] embedding;
    }
}

} // extern "C"
