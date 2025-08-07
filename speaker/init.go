package speaker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// init 初始化函数，在包首次导入时执行
// 负责检查C++库是否存在，如果不存在则提示用户运行 go run
func init() {
	// 获取当前操作系统
	osType := runtime.GOOS

	// 检查工作目录的lib是否存在
	libName := getLibName(osType)
	workDir, _ := os.Getwd()
	workdirLibPath := filepath.Join(workDir, "lib")
	workdirLibExists := checkLibExists(workdirLibPath, libName)

	// 如果库已经存在，直接返回
	if workdirLibExists {
		return
	}

	// 如果库不存在，提示用户运行 go run
	fmt.Printf("警告: 3d-speaker C++库文件不存在\n")
	fmt.Printf("请运行以下命令生成库文件:\n")
	fmt.Printf("  go run github.com/seastart/3dspeaker-onnx-go/cmd/build-lib\n")
	fmt.Printf("或者查看 https://github.com/seastart/3dspeaker-onnx-go 获取预编译库\n")
}

// checkLibExists 检查指定目录中是否存在库文件
func checkLibExists(dirPath, libName string) bool {
	libPath := filepath.Join(dirPath, libName)
	_, err := os.Stat(libPath)
	return err == nil
}

// getLibName 根据操作系统获取动态库名称
func getLibName(osType string) string {
	switch osType {
	case "darwin":
		return "libspeaker_wrapper.dylib"
	case "linux":
		return "libspeaker_wrapper.so"
	default:
		return "libspeaker_wrapper.so" // 默认采用Linux命名规则
	}
}
