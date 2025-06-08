// 音频相似度比较示例程序
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/seastart/3dspeaker-onnx-go/speaker"
)

func main() {
	// 解析命令行参数
	modelPath := flag.String("model", "", "ONNX模型文件路径")
	configPath := flag.String("config", "", "FBANK特征配置文件路径")
	audio1Path := flag.String("audio1", "", "第一个wav PCM音频文件路径")
	audio2Path := flag.String("audio2", "", "第二个wav PCM音频文件路径")
	threshold := flag.Float64("threshold", 0.70, "判断为同一说话人的阈值")
	flag.Parse()

	// 检查必要参数
	if *modelPath == "" || *configPath == "" || *audio1Path == "" || *audio2Path == "" {
		fmt.Println("用法: compare_audio -model=<模型路径> -config=<配置文件路径> -audio1=<音频文件1> -audio2=<音频文件2> [-threshold=0.70]")
		os.Exit(1)
	}

	// 确保文件存在
	for _, path := range []string{*modelPath, *configPath, *audio1Path, *audio2Path} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("错误: 文件不存在: %s\n", path)
			os.Exit(1)
		}
	}

	// 初始化Speaker
	fmt.Println("正在加载模型...")
	spk, err := speaker.New(*modelPath, *configPath)
	if err != nil {
		fmt.Printf("加载模型失败: %v\n", err)
		os.Exit(1)
	}
	defer spk.Close()

	// 读取PCM文件
	pcm1, err := readPCMFile(*audio1Path)
	if err != nil {
		fmt.Printf("读取音频文件1失败: %v\n", err)
		os.Exit(1)
	}

	pcm2, err := readPCMFile(*audio2Path)
	if err != nil {
		fmt.Printf("读取音频文件2失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("音频1: %s\n", filepath.Base(*audio1Path))
	fmt.Printf("音频2: %s\n", filepath.Base(*audio2Path))

	// 比较说话人
	fmt.Println("正在比较音频...")
	isSame, score, err := spk.IsSameSpeaker(pcm1, pcm2, float32(*threshold))
	if err != nil {
		fmt.Printf("比较音频失败: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	fmt.Printf("\n比较结果:\n")
	fmt.Printf("相似度分数: %.4f (阈值: %.2f)\n", score, *threshold)
	if isSame {
		fmt.Println("判断结果: 两段音频来自同一说话人")
	} else {
		fmt.Println("判断结果: 两段音频来自不同说话人")
	}
}

// WAV文件头部结构
type wavHeader struct {
	RiffID        [4]byte // "RIFF"
	FileSize      uint32  // 文件总大小 - 8
	WaveID        [4]byte // "WAVE"
	FmtID         [4]byte // "fmt "
	FmtSize       uint32  // fmt 块大小
	AudioFormat   uint16  // 音频格式（1 = PCM）
	NumChannels   uint16  // 声道数（1 = 单声道，2 = 立体声）
	SampleRate    uint32  // 采样率（每秒样本数）
	ByteRate      uint32  // 每秒字节数 = SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // 每个采样的字节数 = NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 每个采样的位数（通常16位）
}

// readPCMFile 读取音频文件（PCM或WAV）并返回int16数组
func readPCMFile(filePath string) ([]int16, error) {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 检查是否是WAV文件
	isWav := len(data) > 44 && // WAV头部至少有 44 字节
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WAVE"

	if isWav {
		return readWavFile(data)
	} else {
		return readRawPCM(data)
	}
}

// readWavFile 解析WAV文件并返回16khz单声道int16音频
func readWavFile(data []byte) ([]int16, error) {
	// 解析WAV头部
	var header wavHeader
	reader := bytes.NewReader(data[:44]) // 头部通常44字节
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("解析WAV头部失败: %w", err)
	}

	// 输出音频信息
	fmt.Printf("音频信息: 采样率=%d Hz, 声道数=%d, 位深=%d bit\n",
		header.SampleRate, header.NumChannels, header.BitsPerSample)

	// 定位数据块
	dataOffset := 0
	dataSize := uint32(0)

	for i := 12; i < len(data)-8; i++ {
		if string(data[i:i+4]) == "data" {
			dataSize = binary.LittleEndian.Uint32(data[i+4 : i+8])
			dataOffset = i + 8
			fmt.Printf("数据块位置: 偏移量=%d, 大小=%d 字节\n", dataOffset, dataSize)
			break
		}
	}

	if dataOffset == 0 {
		return nil, fmt.Errorf("找不到WAV文件的data块")
	}

	// 提取PCM数据
	pcmData := data[dataOffset : dataOffset+int(dataSize)]

	var pcm []int16

	// 处理多声道（转换为单声道）
	if header.NumChannels > 1 {
		fmt.Printf("检测到 %d 声道音频，正在转换为单声道...\n", header.NumChannels)
		pcm = convertToMono(pcmData, int(header.NumChannels), int(header.BitsPerSample))
	} else {
		// 单声道处理
		pcm = make([]int16, len(pcmData)/2)
		for i := 0; i < len(pcm); i++ {
			pcm[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2:]))
		}
	}

	// 模型期望的采样率
	expectedSampleRate := 16000

	// 如果采样率不是16kHz，进行采样率转换
	if int(header.SampleRate) != expectedSampleRate {
		fmt.Printf("正在进行采样率转换: %d Hz -> %d Hz\n", header.SampleRate, expectedSampleRate)
		pcm = resamplePcm(pcm, int(header.SampleRate), expectedSampleRate)
	}

	return pcm, nil
}

// readRawPCM 处理原始 PCM 数据
func readRawPCM(data []byte) ([]int16, error) {
	// 确保数据长度是偶数（每个int16需要2个字节）
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("无效的PCM文件: 数据长度不是2的倍数")
	}

	// 转换为int16数组
	pcm := make([]int16, len(data)/2)
	for i := 0; i < len(pcm); i++ {
		// 假设PCM数据是小端字节序
		pcm[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}

	return pcm, nil
}

// resamplePcm 将PCM数据从一个采样率转换为另一个采样率
// inputPcm: 输入PCM数据
// inputSampleRate: 输入采样率
// outputSampleRate: 输出采样率
// 返回采样率转换后的PCM数据
func resamplePcm(inputPcm []int16, inputSampleRate, outputSampleRate int) []int16 {
	// 如果采样率相同，直接返回原始数据
	if inputSampleRate == outputSampleRate {
		return inputPcm
	}

	// 计算输出长度
	outputLength := int(float64(len(inputPcm)) * float64(outputSampleRate) / float64(inputSampleRate))
	outputPcm := make([]int16, outputLength)

	// 线性插值采样率转换
	for i := 0; i < outputLength; i++ {
		// 计算对应的输入索引（浮点数）
		inputIndexFloat := float64(i) * float64(inputSampleRate) / float64(outputSampleRate)

		// 获取整数部分和小数部分
		inputIndex := int(inputIndexFloat)
		fraction := inputIndexFloat - float64(inputIndex)

		// 边界检查
		if inputIndex >= len(inputPcm)-1 {
			// 如果超出范围，使用最后一个样本
			outputPcm[i] = inputPcm[len(inputPcm)-1]
		} else {
			// 线性插值
			sample1 := float64(inputPcm[inputIndex])
			sample2 := float64(inputPcm[inputIndex+1])
			interpolatedSample := sample1*(1-fraction) + sample2*fraction
			outputPcm[i] = int16(interpolatedSample)
		}
	}

	return outputPcm
}

// convertToMono 将多声道音频转换为单声道
func convertToMono(data []byte, channels, bitsPerSample int) []int16 {
	bytesPerSample := bitsPerSample / 8
	samplesPerChannel := len(data) / (channels * bytesPerSample)

	// 创建单声道结果
	monoData := make([]int16, samplesPerChannel)

	for i := 0; i < samplesPerChannel; i++ {
		sum := int32(0)

		// 将每个采样的所有声道汇总
		for ch := 0; ch < channels; ch++ {
			offset := (i*channels + ch) * bytesPerSample
			sample := int32(int16(binary.LittleEndian.Uint16(data[offset:])))
			sum += sample
		}

		// 取平均值
		monoData[i] = int16(sum / int32(channels))
	}

	return monoData
}
