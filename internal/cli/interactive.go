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
	"gobake/internal/display"
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
	display.PrintSection("当前系统环境")
	
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
	
	display.PrintFieldValue("GOOS", goos)
	display.PrintFieldValue("GOARCH", goarch)
	display.PrintFieldValue("CGO_ENABLED", cgoEnabled)
}

// StartInteractiveBuild 启动交互式构建过程
func StartInteractiveBuild() error {
	// 在开始时立即显示当前环境
	display.PrintSection("GoBake 多平台构建工具")
	display.PrintHighlight("版本: 1.0.0")
	
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
	
	display.PrintEmptyLine()
	display.PrintHeader("项目信息")
	display.PrintInfo("当前项目目录: %s", currentDir)
	display.PrintInfo("提取的包名: %s", config.PackageName)

	reader := bufio.NewReader(os.Stdin)

	// 1. 询问是否使用默认输出目录
	display.PrintSubSection("配置输出目录")
	display.PrintPrompt(fmt.Sprintf("是否使用默认输出目录? (当前: %s, Y/n): ", config.OutputDir))
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(answer)) == "n" {
		display.PrintInputPrompt("请输入自定义输出目录路径: ")
		if customDir, err := reader.ReadString('\n'); err == nil {
			config.OutputDir = strings.TrimSpace(customDir)
		}
	}
	display.PrintSuccess(fmt.Sprintf("输出目录设置为: %s", config.OutputDir))

	// 2. 询问是否启用 CGO
	display.PrintSubSection("配置CGO选项")
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
	
	display.PrintPrompt(fmt.Sprintf("是否启用 CGO? (当前: %s, %s): ", currentCGO, promptFormat))
		
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(answer) != "" {
		if strings.TrimSpace(strings.ToLower(answer)) == "y" {
			config.CGOEnabled = true
			display.PrintSuccess("CGO 已启用")
		} else if strings.TrimSpace(strings.ToLower(answer)) == "n" {
			config.CGOEnabled = false
			display.PrintSuccess("CGO 已禁用")
		} else {
			// 保持默认选项
			if config.CGOEnabled {
				display.PrintSuccess("CGO 已启用 (默认)")
			} else {
				display.PrintSuccess("CGO 已禁用 (默认)")
			}
		}
	} else {
		// 保持默认选项
		if config.CGOEnabled {
			display.PrintSuccess("CGO 已启用 (默认)")
		} else {
			display.PrintSuccess("CGO 已禁用 (默认)")
		}
	}

	// 3. 询问是否构建所有平台
	display.PrintSubSection("选择目标平台")
	display.PrintPrompt("是否构建所有支持的平台和架构? (Y/n): ")
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(answer)) == "n" {
		config.BuildAllPlatforms = false
		
		// 显示可用平台列表
		display.PrintEmptyLine()
		display.PrintHeader("可用平台:")
		display.PrintInfo("1. Windows AMD64")
		display.PrintInfo("2. Windows ARM64")
		display.PrintInfo("3. Linux AMD64")
		display.PrintInfo("4. Linux ARM64")

		// 获取用户选择
		display.PrintEmptyLine()
		display.PrintInputPrompt("请输入平台编号（用空格分隔，例如 '1 3 4'）: ")
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
				display.PrintWarning("未选择任何平台，将只构建 Windows AMD64 版本")
				config.Platforms = []builder.Platform{
					{OS: "windows", Arch: "amd64"},
				}
			} else {
				// 根据选择添加平台
				var platforms []builder.Platform
				display.PrintEmptyLine()
				display.PrintHeader("已选择的平台:")
				if selected[1] {
					platforms = append(platforms, builder.Platform{OS: "windows", Arch: "amd64"})
					display.PrintInfo("- Windows AMD64")
				}
				if selected[2] {
					platforms = append(platforms, builder.Platform{OS: "windows", Arch: "arm64"})
					display.PrintInfo("- Windows ARM64")
				}
				if selected[3] {
					platforms = append(platforms, builder.Platform{OS: "linux", Arch: "amd64"})
					display.PrintInfo("- Linux AMD64")
				}
				if selected[4] {
					platforms = append(platforms, builder.Platform{OS: "linux", Arch: "arm64"})
					display.PrintInfo("- Linux ARM64")
				}
				config.Platforms = platforms
			}
		}
	} else {
		config.BuildAllPlatforms = true
		display.PrintSuccess("将构建所有支持的平台和架构")
	}

	// 创建并执行构建器
	b := builder.NewBuilder(config)
	
	display.PrintSection("开始构建过程")
	if err := b.Build(); err != nil {
		display.PrintError(err.Error())
		return err
	}

	display.PrintSection("构建完成")
	display.PrintSuccess("所有版本构建成功")
	display.PrintHighlight(fmt.Sprintf("文件已生成在 %s 目录中", config.OutputDir))
	
	display.PrintEmptyLine()
	display.PrintInfo("环境变量已恢复到原始状态")
	display.PrintInfo("感谢使用 GoBake 多平台构建工具")
	display.PrintDivider()

	return nil
} 