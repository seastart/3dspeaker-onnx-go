// Package speaker 提供了使用3D-Speaker进行说话人识别的Go接口
package speaker

import (
	"errors"
	"fmt"
)

// Speaker 提供了说话人识别的高级API
type Speaker struct {
	model *ModelHandle
}

// New 创建一个新的Speaker实例
//
// 参数:
//   - onnxModelPath: ONNX模型文件路径
//   - fbankConfigPath: FBANK特征配置文件路径
//
// 返回:
//   - Speaker实例和可能的错误
func New(onnxModelPath, fbankConfigPath string) (*Speaker, error) {
	model, err := LoadModel(onnxModelPath, fbankConfigPath)
	if err != nil {
		return nil, fmt.Errorf("创建Speaker实例失败: %w", err)
	}
	return &Speaker{model: model}, nil
}

// Close 关闭Speaker实例并释放资源
func (s *Speaker) Close() error {
	if s.model == nil {
		return errors.New("Speaker实例已关闭或未初始化")
	}
	s.model.Close()
	s.model = nil
	return nil
}

// ExtractEmbedding 从PCM音频数据中提取说话人嵌入向量[必须是16khz单声道音频]
//
// 参数:
//   - pcmData: PCM音频数据，int16格式
//
// 返回:
//   - 嵌入向量和可能的错误
func (s *Speaker) ExtractEmbedding(pcmData []int16) (*Embedding, error) {
	if s.model == nil {
		return nil, errors.New("Speaker实例已关闭或未初始化")
	}
	return s.model.ExtractEmbedding(pcmData)
}

// CompareSpeakers 比较两段音频的说话人相似度[必须是16khz单声道音频]
//
// 参数:
//   - pcm1: 第一段PCM音频数据
//   - pcm2: 第二段PCM音频数据
//
// 返回:
//   - 相似度[-1,1]，越接近1表示越相似
//   - 可能的错误
func (s *Speaker) CompareSpeakers(pcm1, pcm2 []int16) (float32, error) {
	if s.model == nil {
		return 0, errors.New("Speaker实例已关闭或未初始化")
	}

	// 提取第一段音频的嵌入向量
	emb1, err := s.ExtractEmbedding(pcm1)
	if err != nil {
		return 0, fmt.Errorf("提取第一段音频嵌入向量失败: %w", err)
	}

	// 提取第二段音频的嵌入向量
	emb2, err := s.ExtractEmbedding(pcm2)
	if err != nil {
		return 0, fmt.Errorf("提取第二段音频嵌入向量失败: %w", err)
	}

	// 计算余弦相似度
	similarity, err := CosineSimilarity(emb1, emb2)
	if err != nil {
		return 0, fmt.Errorf("计算余弦相似度失败: %w", err)
	}

	return similarity, nil
}

// IsSameSpeaker 判断两段音频是否来自同一说话人[必须是16khz单声道音频]
//
// 参数:
//   - pcm1: 第一段PCM音频数据
//   - pcm2: 第二段PCM音频数据
//   - threshold: 判断阈值，默认为0.70
//
// 返回:
//   - 是否为同一说话人
//   - 相似度分数
//   - 可能的错误
func (s *Speaker) IsSameSpeaker(pcm1, pcm2 []int16, threshold float32) (bool, float32, error) {
	// 如果未指定阈值，使用默认值0.70
	if threshold <= 0 {
		threshold = 0.70
	}

	// 直接使用余弦相似度计算
	similarity, err := s.CompareSpeakers(pcm1, pcm2)
	if err != nil {
		return false, 0, err
	}

	// 输出调试信息
	// fmt.Printf("使用纯余弦相似度计算，得分=%.4f\n", similarity)

	// 判断是否相似
	return similarity >= threshold, similarity, nil
}

// CompareHybrid 使用混合相似度（余弦+L2距离）比较两段音频
//
// 参数:
//   - pcm1: 第一段PCM音频数据
//   - pcm2: 第二段PCM音频数据
//   - cosineWeight: 余弦相似度的权重（0-1），值越大越重视余弦相似度
//
// 返回:
//   - 混合相似度评分（值越高越相似）
//   - 可能的错误
func (s *Speaker) CompareHybrid(pcm1, pcm2 []int16, cosineWeight float32) (float32, error) {
	if s.model == nil {
		return 0, errors.New("Speaker实例已关闭或未初始化")
	}

	// 提取第一段音频的嵌入向量
	emb1, err := s.ExtractEmbedding(pcm1)
	if err != nil {
		return 0, fmt.Errorf("提取第一段音频嵌入向量失败: %w", err)
	}

	// 提取第二段音频的嵌入向量
	emb2, err := s.ExtractEmbedding(pcm2)
	if err != nil {
		return 0, fmt.Errorf("提取第二段音频嵌入向量失败: %w", err)
	}

	// 计算混合相似度
	hybridScore, err := HybridSimilarity(emb1, emb2, cosineWeight)
	if err != nil {
		return 0, fmt.Errorf("计算混合相似度失败: %w", err)
	}

	return hybridScore, nil
}

// GetEmbeddingDimension 获取嵌入向量的维度
func (e *Embedding) GetEmbeddingDimension() int {
	return len(e.data)
}

// GetData 获取嵌入向量数据
func (e *Embedding) GetData() []float32 {
	return e.data
}
