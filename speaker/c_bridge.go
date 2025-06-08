package speaker

/*
#cgo CFLAGS: -I${SRCDIR}/../c
#cgo LDFLAGS: -L${SRCDIR}/../c/build -lspeaker_wrapper -lstdc++

#include <stdlib.h>
#include "speaker_wrapper.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"unsafe"
)

// ModelHandle 封装了C语言的模型句柄
type ModelHandle struct {
	handle C.SpeakerModelHandle
}

// FrameExtractionOptions 帧提取选项
type FrameExtractionOptions struct {
	SampleFreq        float32 `json:"sample_freq"`              // 采样率
	FrameShiftMs      float32 `json:"frame_shift_ms"`           // 帧移（毫秒）
	FrameLengthMs     float32 `json:"frame_length_ms"`          // 帧长（毫秒）
	Dither            float32 `json:"dither"`                   // 抖动参数，0表示不使用抖动
	RemoveDcOffset    bool    `json:"remove_dc_offset"`         // 是否移除直流偏移
	PreEmphasisCoeff  float32 `json:"pre_emphasis_coefficient"` // 预加重系数
	WindowType        string  `json:"window_type"`              // 窗函数类型
	RoundToPowerOfTwo bool    `json:"round_to_power_of_two"`    // 是否四舍五入到2的幂
}

// MelBanksOptions 梅尔滤波器组选项
type MelBanksOptions struct {
	NumBins  int     `json:"num_bins"`  // 梅尔滤波器组数量
	LowFreq  float32 `json:"low_freq"`  // 最低频率
	HighFreq float32 `json:"high_freq"` // 最高频率
}

// FbankConfig FBANK特征提取配置结构体
type FbankConfig struct {
	FrameExtractionOptions FrameExtractionOptions `json:"FrameExtractionOptions"` // 帧提取选项
	MelBanksOptions        MelBanksOptions        `json:"MelBanksOptions"`        // 梅尔滤波器组选项
	UsePower               bool                   `json:"use_power"`              // 是否使用功率谱
	UseLogFbank            bool                   `json:"use_log_fbank"`          // 是否使用对数化FBANK
	UseEnergy              bool                   `json:"use_energy"`             // 是否使用能量
	EnergyFloor            float32                `json:"energy_floor"`           // 能量下限
	RawEnergy              bool                   `json:"raw_energy"`             // 是否使用原始能量
}

// 默认的特征提取配置
var defaultFbankConfig = FbankConfig{
	FrameExtractionOptions: FrameExtractionOptions{
		SampleFreq:        16000.0,
		FrameShiftMs:      10.0,
		FrameLengthMs:     25.0,
		Dither:            0.0, // 默认不使用抖动
		RemoveDcOffset:    true,
		PreEmphasisCoeff:  0.97,
		WindowType:        "povey",
		RoundToPowerOfTwo: true,
	},
	MelBanksOptions: MelBanksOptions{
		NumBins:  80,
		LowFreq:  20,
		HighFreq: 0,
	},
	UsePower:    true,
	UseLogFbank: true,
	UseEnergy:   false,
	EnergyFloor: 0.0,
	RawEnergy:   true,
}

// LoadModel 加载说话人识别模型（使用配置文件）
// onnxModelPath: ONNX模型文件路径
// fbankConfigPath: FBANK特征提取配置文件路径
func LoadModel(onnxModelPath, fbankConfigPath string) (*ModelHandle, error) {
	// 尝试从配置文件加载参数
	config, err := loadFbankConfig(fbankConfigPath)
	if err != nil {
		// 配置文件加载失败，使用默认参数
		fmt.Printf("警告: 无法加载配置文件 %s: %v, 将使用默认参数\n", fbankConfigPath, err)
		config = defaultFbankConfig
	}

	// 使用解析出的参数调用新的加载函数
	return LoadModelWithParams(onnxModelPath, config)
}

// LoadModelWithParams 加载说话人识别模型（直接使用参数）
// onnxModelPath: ONNX模型文件路径
// config: FBANK特征提取配置
func LoadModelWithParams(onnxModelPath string, config FbankConfig) (*ModelHandle, error) {
	cOnnxPath := C.CString(onnxModelPath)
	defer C.free(unsafe.Pointer(cOnnxPath))

	// 将Go结构体中的参数传递给C函数
	var useLog C.int
	if config.UseLogFbank {
		useLog = 1
	} else {
		useLog = 0
	}

	var usePower C.int
	if config.UsePower {
		usePower = 1
	} else {
		usePower = 0
	}

	// 打印使用的关键参数
	fmt.Printf("使用传入的参数创建FbankComputer [采样率=%v, 帧移=%vms, 帧长=%vms, 滤波器数=%v, 使用对数=%v, 抖动=%v, 使用功率谱=%v]\n",
		config.FrameExtractionOptions.SampleFreq,
		config.FrameExtractionOptions.FrameShiftMs,
		config.FrameExtractionOptions.FrameLengthMs,
		config.MelBanksOptions.NumBins,
		config.UseLogFbank,
		config.FrameExtractionOptions.Dither,
		config.UsePower)

	// 调用C++函数加载模型 - 现在传递所有参数
	handle := C.LoadSpeakerModel(
		cOnnxPath,
		C.float(config.FrameExtractionOptions.SampleFreq),
		C.float(config.FrameExtractionOptions.FrameShiftMs),
		C.float(config.FrameExtractionOptions.FrameLengthMs),
		C.int(config.MelBanksOptions.NumBins),
		useLog,
		C.float(config.FrameExtractionOptions.Dither),
		usePower,
	)

	if handle == nil {
		return nil, errors.New("加载模型失败")
	}

	m := &ModelHandle{handle: handle}
	// 注册模型释放函数
	runtime.SetFinalizer(m, freeModel)

	return m, nil
}

// loadFbankConfig 从JSON文件加载FBANK配置
func loadFbankConfig(configPath string) (FbankConfig, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultFbankConfig, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 直接解析JSON到结构体
	var config FbankConfig
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("解析JSON到结构体失败: %v，尝试兼容模式\n", err)

		// 尝试兼容模式解析
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return defaultFbankConfig, fmt.Errorf("解析JSON失败: %w", err)
		}

		// 使用默认配置作为基础
		config = defaultFbankConfig

		// 解析FrameExtractionOptions
		if frameOpts, ok := jsonData["FrameExtractionOptions"].(map[string]interface{}); ok {
			if sampleFreq, ok := frameOpts["sample_freq"].(float64); ok {
				config.FrameExtractionOptions.SampleFreq = float32(sampleFreq)
			}
			if frameShift, ok := frameOpts["frame_shift_ms"].(float64); ok {
				config.FrameExtractionOptions.FrameShiftMs = float32(frameShift)
			}
			if frameLength, ok := frameOpts["frame_length_ms"].(float64); ok {
				config.FrameExtractionOptions.FrameLengthMs = float32(frameLength)
			}
			if dither, ok := frameOpts["dither"].(float64); ok {
				config.FrameExtractionOptions.Dither = float32(dither)
			}
			if preEmphasis, ok := frameOpts["pre_emphasis_coefficient"].(float64); ok {
				config.FrameExtractionOptions.PreEmphasisCoeff = float32(preEmphasis)
			}
			if removeDc, ok := frameOpts["remove_dc_offset"].(bool); ok {
				config.FrameExtractionOptions.RemoveDcOffset = removeDc
			}
			if windowType, ok := frameOpts["window_type"].(string); ok {
				config.FrameExtractionOptions.WindowType = windowType
			}
			if roundToPow2, ok := frameOpts["round_to_power_of_two"].(bool); ok {
				config.FrameExtractionOptions.RoundToPowerOfTwo = roundToPow2
			}
		}

		// 解析MelBanksOptions
		if melOpts, ok := jsonData["MelBanksOptions"].(map[string]interface{}); ok {
			if numBins, ok := melOpts["num_bins"].(float64); ok {
				config.MelBanksOptions.NumBins = int(numBins)
			}
			if lowFreq, ok := melOpts["low_freq"].(float64); ok {
				config.MelBanksOptions.LowFreq = float32(lowFreq)
			}
			if highFreq, ok := melOpts["high_freq"].(float64); ok {
				config.MelBanksOptions.HighFreq = float32(highFreq)
			}
		}

		// 解析其他选项
		if usePower, ok := jsonData["use_power"].(bool); ok {
			config.UsePower = usePower
		}
		if useLogFbank, ok := jsonData["use_log_fbank"].(bool); ok {
			config.UseLogFbank = useLogFbank
		}
		if useEnergy, ok := jsonData["use_energy"].(bool); ok {
			config.UseEnergy = useEnergy
		}
		if energyFloor, ok := jsonData["energy_floor"].(float64); ok {
			config.EnergyFloor = float32(energyFloor)
		}
		if rawEnergy, ok := jsonData["raw_energy"].(bool); ok {
			config.RawEnergy = rawEnergy
		}
	}

	// 打印加载的关键参数
	fmt.Printf("成功加载配置: 采样率=%v, 帧移=%v, 帧长=%v, 抖动=%v, 滤波器数=%v, 使用对数=%v\n",
		config.FrameExtractionOptions.SampleFreq,
		config.FrameExtractionOptions.FrameShiftMs,
		config.FrameExtractionOptions.FrameLengthMs,
		config.FrameExtractionOptions.Dither,
		config.MelBanksOptions.NumBins,
		config.UseLogFbank)

	return config, nil
}

// freeModel 释放模型资源
func freeModel(m *ModelHandle) {
	if m.handle != nil {
		C.FreeSpeakerModel(m.handle)
		m.handle = nil
	}
}

// Close 手动释放模型资源
func (m *ModelHandle) Close() {
	if m.handle != nil {
		C.FreeSpeakerModel(m.handle)
		m.handle = nil
		runtime.SetFinalizer(m, nil)
	}
}

// Embedding 表示说话人嵌入向量
type Embedding struct {
	data []float32
}

// ExtractEmbedding 从PCM数据中提取说话人嵌入向量[必须是16khz单声道音频]
// pcmData: PCM数据（int16格式）
func (m *ModelHandle) ExtractEmbedding(pcmData []int16) (*Embedding, error) {
	if m.handle == nil {
		return nil, errors.New("模型已关闭或未初始化")
	}

	if len(pcmData) == 0 {
		return nil, errors.New("PCM数据为空")
	}

	// 注意: 采样率转换已在readPCMFile中完成，这里直接使用转换后的数据

	var cEmbedding *C.float
	var cEmbeddingSize C.int

	// 调用C函数提取嵌入向量
	ret := C.ExtractEmbedding(
		m.handle,
		(*C.short)(unsafe.Pointer(&pcmData[0])),
		C.int(len(pcmData)),
		&cEmbedding,
		&cEmbeddingSize,
	)

	if ret == 0 || cEmbedding == nil {
		return nil, errors.New("提取嵌入向量失败")
	}

	// 创建Go切片并复制数据
	embeddingSize := int(cEmbeddingSize)
	goEmbedding := make([]float32, embeddingSize)

	// 从C内存复制到Go内存
	embeddingPtr := unsafe.Pointer(cEmbedding)
	for i := 0; i < embeddingSize; i++ {
		ptr := (*C.float)(unsafe.Pointer(uintptr(embeddingPtr) + uintptr(i)*unsafe.Sizeof(C.float(0))))
		goEmbedding[i] = float32(*ptr)
	}

	// 释放C分配的内存
	C.FreeEmbedding(cEmbedding)

	return &Embedding{data: goEmbedding}, nil
}

// CosineSimilarity 计算两个嵌入向量的余弦相似度
func CosineSimilarity(emb1, emb2 *Embedding) (float32, error) {
	if emb1 == nil || emb2 == nil {
		return 0, errors.New("嵌入向量为空")
	}

	if len(emb1.data) == 0 || len(emb2.data) == 0 {
		return 0, errors.New("嵌入向量数据为空")
	}

	if len(emb1.data) != len(emb2.data) {
		return 0, fmt.Errorf("嵌入向量维度不匹配: %d vs %d", len(emb1.data), len(emb2.data))
	}

	similarity := C.ComputeCosineSimilarity(
		(*C.float)(unsafe.Pointer(&emb1.data[0])),
		C.int(len(emb1.data)),
		(*C.float)(unsafe.Pointer(&emb2.data[0])),
		C.int(len(emb2.data)),
	)

	return float32(similarity), nil
}

// L2Distance 计算两个嵌入向量的L2距离
func L2Distance(emb1, emb2 *Embedding) (float32, error) {
	if emb1 == nil || emb2 == nil {
		return -1, errors.New("嵌入向量为空")
	}

	if len(emb1.data) == 0 || len(emb2.data) == 0 {
		return -1, errors.New("嵌入向量数据为空")
	}

	if len(emb1.data) != len(emb2.data) {
		return -1, fmt.Errorf("嵌入向量维度不匹配: %d vs %d", len(emb1.data), len(emb2.data))
	}

	// 调用C++函数计算L2距离
	distance := C.ComputeL2Distance(
		(*C.float)(unsafe.Pointer(&emb1.data[0])),
		C.int(len(emb1.data)),
		(*C.float)(unsafe.Pointer(&emb2.data[0])),
		C.int(len(emb2.data)),
	)

	// 检查是否返回错误值（负值）
	if distance < 0 {
		return -1, errors.New("L2距离计算失败")
	}

	return float32(distance), nil
}

// HybridSimilarity 结合余弦相似度和L2距离的混合评分
// 权重范围[0,1]，值越大表示越重视余弦相似度，越小表示越重视L2距离
func HybridSimilarity(emb1, emb2 *Embedding, cosineWeight float32) (float32, error) {
	// 计算余弦相似度
	cosine, err := CosineSimilarity(emb1, emb2)
	if err != nil {
		return 0, fmt.Errorf("计算余弦相似度失败: %w", err)
	}

	// 计算L2距离
	l2, err := L2Distance(emb1, emb2)
	if err != nil {
		return 0, fmt.Errorf("计算L2距离失败: %w", err)
	}

	// 将L2距离转换为相似度得分（距离越小，相似度越高）
	// 使用指数衰减函数将L2距离映射到[0,1]范围
	l2Similarity := float32(math.Exp(-float64(l2)))

	// 确保权重在有效范围内
	if cosineWeight < 0 {
		cosineWeight = 0
	} else if cosineWeight > 1 {
		cosineWeight = 1
	}

	// 计算加权混合得分
	hybridScore := cosineWeight*cosine + (1-cosineWeight)*l2Similarity

	fmt.Printf("混合评分计算: 余弦相似度=%.4f, L2相似度=%.4f, 权重=%.2f, 最终得分=%.4f\n",
		cosine, l2Similarity, cosineWeight, hybridScore)

	return hybridScore, nil
}
