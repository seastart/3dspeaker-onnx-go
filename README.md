Golang版本onnxruntime [3D-Speaker](https://github.com/modelscope/3D-Speaker)

## 编译方法
```sh
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
```sh
CGO_ENABLED=1 CGO_CFLAGS="-I/usr/local/lib/onnxruntime/include" CGO_LDFLAGS="-L/usr/local/lib/onnxruntime/lib" go run compare_audio.go -model=./model/model.onnx -config=./model/fbank_config.json -audio1=man1.wav -audio2=man2.wav
```

## TODO
- [ ] 无需编译动态库，直接cgo c++源码
- [ ] 无需依赖C++库，直接用Go实现，如[onnxruntime_go](https://github.com/yalue/onnxruntime_go) [onnx-go](https://github.com/oramasearch/onnx-go)
