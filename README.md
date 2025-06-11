# 3D-Speaker-ONNX-Go

Golang版本onnxruntime

3D-Speaker模型的ONNX推理实现，支持Go语言调用。本项目可作为Go库被外部项目通过`go get`方式引用。

## 功能特点

- 支持多平台（macOS、Linux）
- 支持多架构（amd64/x86_64、arm64/aarch64）
- 提供自动编译和预编译两种使用方式
- 基于ONNX Runtime进行高效推理

## 安装与使用

### 方式1：自动编译（默认）

直接通过Go模块引入，会自动编译C++库：

```bash
go get github.com/seastart/3dspeaker-onnx-go
```

库将自动编译所需的C++动态库。确保已安装以下依赖：
- C++编译器（GCC 或 Clang）
- ONNX Runtime
- Make 工具

### 方式2：使用预编译库

如果不想自动编译C++库，可以下载对应平台和架构的预编译库：

1. 下载预编译库：

| 操作系统 | 架构 | 下载链接 |
|---------|------|--------|
| macOS   | Intel (amd64) | [下载](https://github.com/seastart/3dspeaker-onnx-go/releases) |
| macOS   | Apple Silicon (arm64) | [下载](https://github.com/seastart/3dspeaker-onnx-go/releases) |
| Linux   | x86_64 (amd64) | [下载](https://github.com/seastart/3dspeaker-onnx-go/releases) |
| Linux   | ARM64 | [下载](https://github.com/seastart/3dspeaker-onnx-go/releases) |

2. 解压到正确的位置：

```bash
# 假设您下载了macOS/arm64版本
tar -xzf 3dspeaker-onnx-go-darwin-arm64.tar.gz -C 您的项目路径
```

3. 构建时使用`nobuild`标签跳过编译：

```bash
go build -tags nobuild
```

## 编译方法

### 依赖
- onnxruntime
    ```sh
    # mac
    brew install onnxruntime
    # ubuntu
    # 手动下载onnxruntime https://github.com/microsoft/onnxruntime/releases 解压到如 /usr/local/lib/onnxruntime/
    ```

### 编译
注意：需要替换下面`MakeFile`及手动编译里的onnxruntime路径为实际路径

```sh
make clean
make
```

或者手动编译
```sh
mkdir -p ./c/build
# 编译C++源码为对象文件
g++ -fPIC -c ./c/feature/feature_basic.cpp -o ./c/build/feature_basic.o -I. -I/usr/local/lib/onnxruntime/include
g++ -fPIC -c ./c/feature/feature_common.cpp -o ./c/build/feature_common.o -I. -I/usr/local/lib/onnxruntime/include
g++ -fPIC -c ./c/feature/feature_fbank.cpp -o ./c/build/feature_fbank.o -I. -I/usr/local/lib/onnxruntime/include
g++ -fPIC -c ./c/feature/feature_functions.cpp -o ./c/build/feature_functions.o -I. -I/usr/local/lib/onnxruntime/include
g++ -fPIC -c ./c/model/speaker_embedding_model.cpp -o ./c/build/speaker_embedding_model.o -I. -I/usr/local/lib/onnxruntime/include
g++ -fPIC -c ./c/speaker_wrapper.cpp -o ./c/build/speaker_wrapper.o -I. -I/usr/local/lib/onnxruntime/include

# 将对象文件链接成共享库，注意链接3D-Speaker的库文件
g++ -shared -o ./c/build/libspeaker_wrapper.so ./c/build/feature_basic.o ./c/build/feature_common.o ./c/build/feature_fbank.o ./c/build/feature_functions.o ./c/build/speaker_embedding_model.o ./c/build/speaker_wrapper.o -L/usr/local/lib/onnxruntime/lib -lonnxruntime -lstdc++
```
## 测试
注意：需要替换下面的onnxruntime路径为实际路径  

```sh
# linux
CGO_ENABLED=1 CGO_CFLAGS="-I/usr/local/lib/onnxruntime/include" CGO_LDFLAGS="-L/usr/local/lib/onnxruntime/lib" go run compare_audio.go -model=./model/model.onnx -config=./model/fbank_config.json -audio1=man1.wav -audio2=man2.wav

# mac
CGO_ENABLED=1 CGO_CFLAGS="-I/opt/homebrew/include/onnxruntime/" CGO_LDFLAGS="-L/opt/homebrew/lib" go run compare_audio.go -model=./model/model.onnx -config=./model/fbank_config.json -audio1=man1.wav -audio2=man2.wav
```

## TODO
- [ ] 无需编译动态库，直接cgo c++源码
- [ ] 无需依赖C++库，直接用Go实现，如[onnxruntime_go](https://github.com/yalue/onnxruntime_go) [onnx-go](https://github.com/oramasearch/onnx-go)
