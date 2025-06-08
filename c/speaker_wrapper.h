#ifndef SPEAKER_WRAPPER_H
#define SPEAKER_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

/**
 * 说话人嵌入模型句柄
 */
typedef void* SpeakerModelHandle;

/**
 * 加载ONNX模型并初始化特征提取器
 * 
 * @param onnx_model_path ONNX模型文件路径
 * @param sample_freq 采样率（例如 16000）
 * @param frame_shift_ms 帧移（毫秒，例如 10.0）
 * @param frame_length_ms 帧长（毫秒，例如 25.0）
 * @param num_bins 梅尔滤波器组数量（例如 80）
 * @param use_log 是否使用对数化FBANK特征（例如 1 表示是）
 * @param dither 抖动参数（例如 0.0 表示不使用抖动）
 * @param use_power 是否使用功率谱（例如 1 表示是）
 * @return 模型句柄，失败时返回NULL
 */
SpeakerModelHandle LoadSpeakerModel(const char* onnx_model_path, 
                                 float sample_freq,
                                 float frame_shift_ms,
                                 float frame_length_ms,
                                 int num_bins,
                                 int use_log,
                                 float dither,
                                 int use_power);

/**
 * 释放模型资源
 * 
 * @param handle 模型句柄
 */
void FreeSpeakerModel(SpeakerModelHandle handle);

/**
 * 从所人数据中提取说话人嵌入向量[16kHz单声道int16 PCM数据]
 * 
 * 注意: 输入数据应已经过滤为16kHz单声道PCM数据
 * 
 * @param handle 模型句柄
 * @param pcm_data PCM数据指针（int16类型数据）
 * @param pcm_length PCM数据长度（样本数）
 * @param embedding 输出的嵌入向量，调用方负责释放内存
 * @param embedding_size 输出的嵌入向量长度
 * @return 成功返回1，失败返回0
 */
int ExtractEmbedding(SpeakerModelHandle handle, 
                     const short* pcm_data, 
                     int pcm_length, 
                     float** embedding, 
                     int* embedding_size);

/**
 * 计算两个嵌入向量的余弦相似度
 * 
 * @param embedding1 第一个嵌入向量
 * @param size1 第一个嵌入向量长度
 * @param embedding2 第二个嵌入向量
 * @param size2 第二个嵌入向量长度
 * @return 余弦相似度[-1,1]，相似度越高值越大
 */
float ComputeCosineSimilarity(const float* embedding1, int size1, 
                              const float* embedding2, int size2);

// 计算两个嵌入向量的L2距离
float ComputeL2Distance(const float* embedding1, int size1,
                       const float* embedding2, int size2);

/**
 * 释放嵌入向量内存
 * 
 * @param embedding 嵌入向量指针
 */
void FreeEmbedding(float* embedding);

#ifdef __cplusplus
}
#endif

#endif // SPEAKER_WRAPPER_H
