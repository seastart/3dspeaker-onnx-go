//go:build !nobuild
// +build !nobuild

package speaker

import (
	"fmt"
	"io"
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
	// fmt.Println("executableDir", getExecutableRoot(), "moduleSrcRoot", getModuleSrcRoot())

	// 尝试寻找预编译库
	libName := getLibName(osType)
	executableLibPath := filepath.Join(getExecutableRoot(), "lib")
	prebuiltPath := filepath.Join(getModuleSrcRoot(), "speaker", "lib", osType, archType)
	builtPath := filepath.Join(getModuleSrcRoot(), "c", "build")

	// 检查预编译库是否存在
	executableLibExists := checkLibExists(executableLibPath, libName)
	prebuiltExists := checkLibExists(prebuiltPath, libName)
	builtExists := checkLibExists(builtPath, libName)

	// 如果库已经存在（优先使用可执行文件目录lib，然后是预编译库或已构建库），直接返回
	if executableLibExists {
		// fmt.Printf("已找到 %s/%s 库文件\n", osType, archType)
		return
	} else if prebuiltExists {
		// 自动拷贝预编译库到可执行文件目录
		if err := copyLibFile(prebuiltPath, executableLibPath, libName); err == nil {
			fmt.Printf("3d-speaker C++库已拷贝至: %s\n", executableLibPath)
		}
		return
	} else if builtExists {
		// 自动拷贝已构建库到可执行文件目录
		if err := copyLibFile(builtPath, executableLibPath, libName); err == nil {
			fmt.Printf("3d-speaker C++库已拷贝至: %s\n", executableLibPath)
		}
		return
	}

	// 如果库不存在，尝试构建
	fmt.Printf("正在为 %s/%s 构建3d-speaker C++库...\n", osType, archType)
	if err := buildLib(); err != nil {
		fmt.Printf("3d-speaker C++库构建失败: %v\n", err)
		fmt.Println("请查看 https://github.com/seastart/3dspeaker-onnx-go 获取预编译库或手动构建说明")
		// 不直接退出，让用户决定如何处理
	} else {
		// 自动拷贝文件到可执行文件目录
		if err := copyLibFile(builtPath, executableLibPath, libName); err != nil {
			fmt.Printf("拷贝库文件失败: %v\n", err)
			fmt.Println("3d-speaker C++库构建成功，但拷贝失败，请手动拷贝至: ", executableLibPath)
		} else {
			fmt.Printf("3d-speaker C++库构建成功，并已拷贝至: %s\n", executableLibPath)
		}
	}
}

// 获取可执行文件所在目录
func getExecutableRoot() string {
	executableDir, _ := os.Executable()
	return filepath.Dir(executableDir)
}

// getModuleSrcRoot 获取模块源码根目录
func getModuleSrcRoot() string {
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
	rootDir := getModuleSrcRoot()

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

// copyLibFile 拷贝库文件到目标目录
func copyLibFile(srcDir, destDir, libName string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 构建源文件和目标文件路径
	srcFile := filepath.Join(srcDir, libName)
	destFile := filepath.Join(destDir, libName)

	// 检查源文件是否存在
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return fmt.Errorf("源文件不存在: %s", srcFile)
	}

	// 打开源文件
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dest.Close()

	// 拷贝文件内容
	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("拷贝文件失败: %w", err)
	}

	// 复制文件权限
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}
	if err := os.Chmod(destFile, srcInfo.Mode()); err != nil {
		return fmt.Errorf("设置文件权限失败: %w", err)
	}

	return nil
}
