# 定义编译器和编译选项
CXXFLAGS = -fPIC -std=c++17

# 检测操作系统类型
OS := $(shell uname -s)
ifeq ($(OS),Darwin)
    OS_LOWER = darwin
else
    OS_LOWER = linux
endif

# 检测CPU架构
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
    GO_ARCH = amd64
else ifeq ($(ARCH),arm64)
    GO_ARCH = arm64
else ifeq ($(ARCH),aarch64)
    GO_ARCH = arm64
else
    GO_ARCH = $(ARCH)
endif

# 根据操作系统设置不同的编译选项
ifeq ($(OS),Darwin)
    # macOS 系统设置
    CXX = clang++
    
    # 检查是否有自定义ONNX路径
    ifneq ($(ONNX_PATH),)
        INCLUDES = -I. -I$(ONNX_PATH)/include
        LDFLAGS = -L$(ONNX_PATH)/lib -lonnxruntime -lstdc++
    else
        INCLUDES = -I. -I/opt/homebrew/include/onnxruntime/
        LDFLAGS = -L/opt/homebrew/lib -lonnxruntime -lstdc++
    endif
    
    LIB_EXT = dylib
else
    # Linux系统设置
    CXX = g++
    
    # 检查是否有自定义ONNX路径
    ifneq ($(ONNX_PATH),)
        INCLUDES = -I. -I$(ONNX_PATH)/include
        LDFLAGS = -L$(ONNX_PATH)/lib -lonnxruntime -lstdc++
    else
        INCLUDES = -I. -I/usr/local/lib/onnxruntime/include
        LDFLAGS = -L/usr/local/lib/onnxruntime/lib -lonnxruntime -lstdc++
    endif
    
    LIB_EXT = so
endif

# 定义源文件和目标文件路径
SRC_DIR = ./c
BUILD_DIR = ./c/build

# 源文件列表
SRCS = $(SRC_DIR)/feature/feature_basic.cpp \
	   $(SRC_DIR)/feature/feature_common.cpp \
	   $(SRC_DIR)/feature/feature_fbank.cpp \
	   $(SRC_DIR)/feature/feature_functions.cpp \
	   $(SRC_DIR)/model/speaker_embedding_model.cpp \
	   $(SRC_DIR)/speaker_wrapper.cpp

# 目标文件列表
OBJS = $(BUILD_DIR)/feature_basic.o \
	   $(BUILD_DIR)/feature_common.o \
	   $(BUILD_DIR)/feature_fbank.o \
	   $(BUILD_DIR)/feature_functions.o \
	   $(BUILD_DIR)/speaker_embedding_model.o \
	   $(BUILD_DIR)/speaker_wrapper.o

# 库文件名
LIB = $(BUILD_DIR)/libspeaker_wrapper.$(LIB_EXT)

# 默认目标
all: $(LIB)

# 创建构建目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# 编译C++源文件为目标文件

$(BUILD_DIR)/feature_basic.o: $(SRC_DIR)/feature/feature_basic.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/feature_common.o: $(SRC_DIR)/feature/feature_common.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/feature_fbank.o: $(SRC_DIR)/feature/feature_fbank.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/feature_functions.o: $(SRC_DIR)/feature/feature_functions.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/speaker_embedding_model.o: $(SRC_DIR)/model/speaker_embedding_model.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

$(BUILD_DIR)/speaker_wrapper.o: $(SRC_DIR)/speaker_wrapper.cpp | $(BUILD_DIR)
	$(CXX) $(CXXFLAGS) $(INCLUDES) -c $< -o $@

# 链接目标文件生成共享库
$(LIB): $(OBJS)
	$(CXX) -shared -o $@ $^ $(LDFLAGS)

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)

# 安装目标 - 将库安装到预编译目录
install: $(LIB)
	@echo "安装库到 speaker/lib/$(OS_LOWER)/$(GO_ARCH)/"
	@mkdir -p ./speaker/lib/$(OS_LOWER)/$(GO_ARCH)/
	@cp $(LIB) ./speaker/lib/$(OS_LOWER)/$(GO_ARCH)/

# 显示当前构建环境信息
info:
	@echo "操作系统: $(OS) ($(OS_LOWER))"
	@echo "CPU架构: $(ARCH) (Go架构: $(GO_ARCH))"
	@echo "编译器: $(CXX)"
	@echo "库扩展名: $(LIB_EXT)"

# 发布目标 - 生成发布包
dist: clean install
	@echo "生成适用于 $(OS_LOWER)/$(GO_ARCH) 的发布包..."
	@tar -czf 3dspeaker-onnx-go-$(OS_LOWER)-$(GO_ARCH).tar.gz ./speaker/lib/$(OS_LOWER)/$(GO_ARCH)/
	@echo "发布包已生成: 3dspeaker-onnx-go-$(OS_LOWER)-$(GO_ARCH).tar.gz"

.PHONY: all clean install info dist