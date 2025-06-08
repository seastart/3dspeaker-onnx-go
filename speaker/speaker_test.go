package speaker

import (
	"testing"
)

// TestSpeakerEmbeddingExtraction 测试从PCM数据中提取嵌入向量
func TestSpeakerEmbeddingExtraction(t *testing.T) {
	// 此处需要根据实际情况设置模型路径和配置文件路径
	modelPath := "../../onnxruntime/model.onnx"
	configPath := "../../onnxruntime/assets/fbank_config.json"

	// 创建Speaker实例
	speaker, err := New(modelPath, configPath)
	if err != nil {
		t.Skipf("跳过测试：无法加载模型: %v", err)
		return
	}
	defer speaker.Close()

	// 生成测试用的PCM数据（1秒16kHz的音频）
	sampleRate := 16000
	pcmData := make([]int16, sampleRate)
	// 简单的正弦波
	for i := 0; i < len(pcmData); i++ {
		// 生成440Hz的正弦波
		pcmData[i] = int16(10000.0 * float64(i%36) / 36.0)
	}

	// 提取嵌入向量
	embedding, err := speaker.ExtractEmbedding(pcmData)
	if err != nil {
		t.Fatalf("提取嵌入向量失败: %v", err)
	}

	// 检查嵌入向量维度
	dimension := embedding.GetEmbeddingDimension()
	if dimension <= 0 {
		t.Fatalf("嵌入向量维度应大于0，实际为: %d", dimension)
	}

	t.Logf("成功提取嵌入向量，维度: %d", dimension)
}

// TestSpeakerComparison 测试比较两段音频的相似度
func TestSpeakerComparison(t *testing.T) {
	// 此处需要根据实际情况设置模型路径和配置文件路径
	modelPath := "../../onnxruntime/model.onnx"
	configPath := "../../onnxruntime/assets/fbank_config.json"

	// 创建Speaker实例
	speaker, err := New(modelPath, configPath)
	if err != nil {
		t.Skipf("跳过测试：无法加载模型: %v", err)
		return
	}
	defer speaker.Close()

	// 生成测试用的PCM数据
	sampleRate := 16000
	pcm1 := make([]int16, sampleRate)
	pcm2 := make([]int16, sampleRate)

	// 生成两段相似的音频
	for i := 0; i < len(pcm1); i++ {
		// 440Hz的正弦波
		pcm1[i] = int16(10000.0 * float64(i%36) / 36.0)
		// 略微不同的442Hz正弦波
		pcm2[i] = int16(10000.0 * float64(i%35) / 35.0)
	}

	// 比较音频相似度
	similarity, err := speaker.CompareSpeakers(pcm1, pcm2)
	if err != nil {
		t.Fatalf("比较音频相似度失败: %v", err)
	}

	t.Logf("音频相似度: %f", similarity)

	// 判断是否为同一说话人
	isSame, score, err := speaker.IsSameSpeaker(pcm1, pcm2, 0.70)
	if err != nil {
		t.Fatalf("判断说话人失败: %v", err)
	}

	t.Logf("是否为同一说话人: %v, 相似度分数: %f", isSame, score)
}
