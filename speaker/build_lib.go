// +build !nobuild

package speaker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:generate make -C ../c

// init 初始化函数，在包首次导入时执行
// 负责检查和确保C++库已正确编译
func init() {
	// 获取当前操作系统和CPU架构
	osType := runtime.GOOS
	archType := runtime.GOARCH

	// 尝试寻找预编译库
	libName := getLibName(osType)
	prebuiltPath := filepath.Join(getModuleRoot(), "speaker", "lib", osType, archType)
	builtPath := filepath.Join(getModuleRoot(), "c", "build")
	
	// 检查预编译库是否存在
	prebuiltExists := checkLibExists(prebuiltPath, libName)
	builtExists := checkLibExists(builtPath, libName)

	// 如果库已经存在（无论是预编译的还是已构建的），直接返回
	if prebuiltExists || builtExists {
		// fmt.Printf("已找到 %s/%s 库文件\n", osType, archType)
		return
	}

	// 如果库不存在，尝试构建
	fmt.Printf("正在为 %s/%s 构建C++库...\n", osType, archType)
	if err := buildLib(); err != nil {
		fmt.Printf("库构建失败: %v\n", err)
		fmt.Println("请查看 https://github.com/seastart/3dspeaker-onnx-go 获取预编译库或手动构建说明")
		// 不直接退出，让用户决定如何处理
	} else {
		fmt.Println("C++库构建成功")
	}
}

// getModuleRoot 获取模块根目录
func getModuleRoot() string {
	// 获取当前文件所在目录
	_, currentFile, _, _ := runtime.Caller(0)
	// 返回上级目录作为模块根目录
	return filepath.Dir(filepath.Dir(currentFile))
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

// buildLib 构建C++库
func buildLib() error {
	rootDir := getModuleRoot()
	
	// 确保构建目录存在
	buildDir := filepath.Join(rootDir, "c", "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("创建构建目录失败: %w", err)
	}
	
	// 执行make命令
	cmd := exec.Command("make", "-C", filepath.Join(rootDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make命令执行失败: %w", err)
	}
	
	return nil
}
