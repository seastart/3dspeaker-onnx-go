package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// main 函数，用于构建和拷贝C++库到用户的工作目录
// 这个工具可以被用户通过 go run github.com/seastart/3dspeaker-onnx-go/cmd/build-lib 调用
func main() {
	// 获取当前操作系统和CPU架构
	osType := runtime.GOOS
	archType := runtime.GOARCH

	fmt.Printf("正在为 %s/%s 构建 3dspeaker C++ 库...\n", osType, archType)
	fmt.Printf("工作目录: %s\n", getWorkDir())
	fmt.Printf("模块源码目录: %s\n", getModuleSrcRoot())

	// 尝试寻找预编译库
	libName := getLibName(osType)
	workdirLibPath := filepath.Join(getWorkDir(), "lib")
	prebuiltPath := filepath.Join(getModuleSrcRoot(), "speaker", "lib", osType, archType)
	builtPath := filepath.Join(getModuleSrcRoot(), "c", "build")

	// 检查预编译库是否存在
	workdirLibExists := checkLibExists(workdirLibPath, libName)
	prebuiltExists := checkLibExists(prebuiltPath, libName)
	builtExists := checkLibExists(builtPath, libName)

	// 如果库已经存在，询问是否重新构建
	if workdirLibExists {
		fmt.Printf("3dspeaker C++ 库已存在: %s\n", filepath.Join(workdirLibPath, libName))
		fmt.Println("如需重新构建，请先删除现有库文件")
		return
	}

	// 尝试从预编译库拷贝
	if prebuiltExists {
		fmt.Printf("发现预编译库，正在拷贝...\n")
		if err := copyLibFile(prebuiltPath, workdirLibPath, libName); err == nil {
			fmt.Printf("✅ 3dspeaker C++ 库已从预编译库拷贝至: %s\n", filepath.Join(workdirLibPath, libName))
			return
		} else {
			fmt.Printf("❌ 拷贝预编译库失败: %v\n", err)
		}
	}

	// 尝试从已构建库拷贝
	if builtExists {
		fmt.Printf("发现已构建库，正在拷贝...\n")
		if err := copyLibFile(builtPath, workdirLibPath, libName); err == nil {
			fmt.Printf("✅ 3dspeaker C++ 库已从构建库拷贝至: %s\n", filepath.Join(workdirLibPath, libName))
			return
		} else {
			fmt.Printf("❌ 拷贝构建库失败: %v\n", err)
		}
	}

	// 如果都没有，尝试构建
	fmt.Printf("🔨 正在从源码构建 3dspeaker C++ 库...\n")
	if err := buildLib(); err != nil {
		fmt.Printf("❌ 3dspeaker C++ 库构建失败: %v\n", err)
		fmt.Println("\n请尝试以下解决方案:")
		fmt.Println("1. 查看 https://github.com/seastart/3dspeaker-onnx-go 获取预编译库")
		fmt.Println("2. 确保系统已安装必要的构建工具 (cmake, make, gcc/clang)")
		fmt.Println("3. 手动构建说明请参考项目 README")
		os.Exit(1)
	}

	// 构建成功后拷贝
	if err := copyLibFile(builtPath, workdirLibPath, libName); err != nil {
		fmt.Printf("❌ 拷贝库文件失败: %v\n", err)
		fmt.Printf("3dspeaker C++ 库构建成功，但拷贝失败\n")
		fmt.Printf("请手动拷贝 %s 至 %s\n",
			filepath.Join(builtPath, libName), filepath.Join(workdirLibPath, libName))
		os.Exit(1)
	}

	fmt.Printf("✅ 3dspeaker C++ 库构建成功，并已拷贝至: %s\n", filepath.Join(workdirLibPath, libName))
	fmt.Println("\n现在您可以正常使用 3dspeaker-onnx-go 库了！")
}

// 获取当前工作目录（用户项目的根目录）
func getWorkDir() string {
	wd, _ := os.Getwd()
	return wd
}

// getModuleSrcRoot 获取 3dspeaker-onnx-go 模块源码根目录
func getModuleSrcRoot() string {
	// 获取当前文件所在目录
	_, currentFile, _, _ := runtime.Caller(0)
	// 返回上上级目录作为模块根目录 (cmd/build-lib -> cmd -> root)
	return filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
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
	case "windows":
		return "libspeaker_wrapper.dll"
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
	cmd := exec.Command("make", "-C", rootDir)
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
