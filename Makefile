# 定义编译器和编译选项
CXXFLAGS = -fPIC -std=c++14
# 检测操作系统类型
OS := $(shell uname -s)

# 根据操作系统设置不同的编译选项
ifeq ($(OS),Darwin)
    # macOS 系统设置
    CXX = clang++
    INCLUDES = -I. -I/opt/homebrew/include/onnxruntime/
    LDFLAGS = -L/opt/homebrew/lib -lonnxruntime -lstdc++
    LIB_EXT = dylib
else
    # Linux系统设置
    CXX = g++
    INCLUDES = -I. -I/usr/local/lib/onnxruntime/include
    LDFLAGS = -L/usr/local/lib/onnxruntime/lib -lonnxruntime -lstdc++
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

.PHONY: all clean