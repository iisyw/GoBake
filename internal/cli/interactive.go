package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gobake/internal/builder"
)

// 获取Go环境变量
func getGoEnv(name string) string {
	cmd := exec.Command("go", "env", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// 如果命令执行失败，返回默认值
		return ""
	}
	return strings.TrimSpace(out.String())
}

// 打印当前环境信息
func printCurrentEnv() {
	fmt.Printf("\n当前系统环境\n")
	fmt.Printf("====================================\n")
	
	// 使用go env命令获取实际环境设置，而不是从环境变量读取
	goos := getGoEnv("GOOS")
	if goos == "" {
		goos = runtime.GOOS
	}
	
	goarch := getGoEnv("GOARCH")
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	
	cgoEnabled := getGoEnv("CGO_ENABLED")
	if cgoEnabled == "" {
		// Go默认值
		if runtime.GOOS == "windows" {
			cgoEnabled = "0"
		} else {
			cgoEnabled = "1"
		}
	}
	
	fmt.Printf("GOOS        = %s\n", goos)
	fmt.Printf("GOARCH      = %s\n", goarch)
	fmt.Printf("CGO_ENABLED = %s\n", cgoEnabled)
	fmt.Printf("====================================\n")
}

// StartInteractiveBuild 启动交互式构建过程
func StartInteractiveBuild() error {
	// 在开始时立即显示当前环境
	fmt.Println("\n===== GoBake 多平台构建工具 =====")
	fmt.Println("版本: 1.0.0")
	printCurrentEnv()
	
	config := builder.BuildConfig{
		OutputDir: "./build",
	}

	// 获取当前目录名作为包名
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}
	config.PackageName = filepath.Base(currentDir)
	fmt.Printf("\n当前项目目录: %s\n", currentDir)
	fmt.Printf("提取的包名: %s\n", config.PackageName)

	reader := bufio.NewReader(os.Stdin)

	// 1. 询问是否使用默认输出目录
	fmt.Printf("\n是否使用默认输出目录? (当前: %s, Y/n): ", config.OutputDir)
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(answer)) == "n" {
		fmt.Print("请输入自定义输出目录路径: ")
		if customDir, err := reader.ReadString('\n'); err == nil {
			config.OutputDir = strings.TrimSpace(customDir)
		}
	}
	fmt.Printf("输出目录设置为: %s\n", config.OutputDir)

	// 2. 询问是否启用 CGO
	// 获取当前CGO状态
	currentCGO := getGoEnv("CGO_ENABLED")
	defaultChoice := "n"
	if currentCGO == "1" {
		defaultChoice = "y"
		config.CGOEnabled = true
	}
	
	// 正确显示默认选项：大写字母表示默认选项
	var promptFormat string
	if defaultChoice == "y" {
		promptFormat = "Y/n"  // CGO默认启用
	} else {
		promptFormat = "y/N"  // CGO默认禁用
	}
	
	fmt.Printf("\n是否启用 CGO? (当前: %s, %s): ", currentCGO, promptFormat)
		
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(answer) != "" {
		if strings.TrimSpace(strings.ToLower(answer)) == "y" {
			config.CGOEnabled = true
			fmt.Println("CGO 已启用")
		} else if strings.TrimSpace(strings.ToLower(answer)) == "n" {
			config.CGOEnabled = false
			fmt.Println("CGO 已禁用")
		} else {
			// 保持默认选项
			if config.CGOEnabled {
				fmt.Println("CGO 已启用 (默认)")
			} else {
				fmt.Println("CGO 已禁用 (默认)")
			}
		}
	} else {
		// 保持默认选项
		if config.CGOEnabled {
			fmt.Println("CGO 已启用 (默认)")
		} else {
			fmt.Println("CGO 已禁用 (默认)")
		}
	}

	// 3. 询问是否构建所有平台
	fmt.Print("\n是否构建所有支持的平台和架构? (Y/n): ")
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(answer)) == "n" {
		config.BuildAllPlatforms = false
		
		// 显示可用平台列表
		fmt.Println("\n可用平台:")
		fmt.Println("1. Windows AMD64")
		fmt.Println("2. Windows ARM64")
		fmt.Println("3. Linux AMD64")
		fmt.Println("4. Linux ARM64")

		// 获取用户选择
		fmt.Print("\n请输入平台编号（用空格分隔，例如 '1 3 4'）: ")
		if numbers, err := reader.ReadString('\n'); err == nil {
			selected := make(map[int]bool)
			for _, num := range strings.Fields(numbers) {
				switch num {
				case "1":
					selected[1] = true
				case "2":
					selected[2] = true
				case "3":
					selected[3] = true
				case "4":
					selected[4] = true
				}
			}

			// 如果没有选择任何平台，默认使用当前平台
			if len(selected) == 0 {
				fmt.Println("[提示] 未选择任何平台，将只构建 Windows AMD64 版本")
				config.Platforms = []builder.Platform{
					{OS: "windows", Arch: "amd64"},
				}
			} else {
				// 根据选择添加平台
				var platforms []builder.Platform
				fmt.Println("\n已选择的平台:")
				if selected[1] {
					platforms = append(platforms, builder.Platform{OS: "windows", Arch: "amd64"})
					fmt.Println("- Windows AMD64")
				}
				if selected[2] {
					platforms = append(platforms, builder.Platform{OS: "windows", Arch: "arm64"})
					fmt.Println("- Windows ARM64")
				}
				if selected[3] {
					platforms = append(platforms, builder.Platform{OS: "linux", Arch: "amd64"})
					fmt.Println("- Linux AMD64")
				}
				if selected[4] {
					platforms = append(platforms, builder.Platform{OS: "linux", Arch: "arm64"})
					fmt.Println("- Linux ARM64")
				}
				config.Platforms = platforms
			}
		}
	} else {
		config.BuildAllPlatforms = true
		fmt.Println("将构建所有支持的平台和架构")
	}

	// 创建并执行构建器
	b := builder.NewBuilder(config)
	
	fmt.Println("\n----- 开始构建过程 -----")
	if err := b.Build(); err != nil {
		return err
	}

	fmt.Printf("\n====================================\n")
	fmt.Printf("    所有版本构建成功\n")
	fmt.Printf("    文件已生成在 %s 目录中\n", config.OutputDir)
	fmt.Printf("====================================\n")
	fmt.Println("\n环境变量已恢复到原始状态")
	fmt.Println("感谢使用 GoBake 多平台构建工具")

	return nil
} 