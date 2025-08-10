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
		GoCommand: "go", // 默认使用标准的go命令
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
	display.PrintInfo("正在分析项目依赖以检测CGO需求...")
	cgoRequired, err := builder.IsCGORequired(".")
	if err != nil {
		display.PrintWarning(fmt.Sprintf("CGO检测失败: %v", err))
		// 在检测失败时，退回旧的逻辑
		cgoRequired = getGoEnv("CGO_ENABLED") == "1"
	}

	config.CGOEnabled = cgoRequired // 设置默认值

	var promptFormat, promptMessage string
	if cgoRequired {
		promptFormat = "Y/n"
		promptMessage = "检测到项目依赖需要 CGO，是否启用?"
	} else {
		promptFormat = "y/N"
		promptMessage = "未检测到项目依赖需要 CGO，是否启用?"
	}

	display.PrintPrompt(fmt.Sprintf("%s (%s): ", promptMessage, promptFormat))

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

	// 3. 询问使用哪个Go命令
	display.PrintSubSection("配置Go命令")
	display.PrintPrompt(fmt.Sprintf("是否使用默认Go命令? (当前: %s, Y/n): ", config.GoCommand))
	if answer, _ := reader.ReadString('\n'); strings.TrimSpace(strings.ToLower(answer)) == "n" {
		display.PrintInfo("您可以指定使用特定的Go版本命令，例如go120、go123等")
		display.PrintInputPrompt("请输入Go命令: ")
		if customCmd, err := reader.ReadString('\n'); err == nil {
			customCmd = strings.TrimSpace(customCmd)
			if customCmd != "" {
				config.GoCommand = customCmd
			}
		}
	}
	display.PrintSuccess(fmt.Sprintf("Go命令设置为: %s", config.GoCommand))

	// 4. 选择目标平台
	display.PrintSubSection("选择目标平台")
	display.PrintHeader("可用平台:")
	display.PrintInfo("1. Windows AMD64")
	display.PrintInfo("2. Windows ARM64")
	display.PrintInfo("3. Linux AMD64")
	display.PrintInfo("4. Linux ARM64")
	display.PrintInfo("5. Windows AMD64 + Linux AMD64 (常用)")
	display.PrintInfo("6. 全部平台")
	display.PrintEmptyLine()
	display.PrintInputPrompt(fmt.Sprintf("请输入平台编号 (默认: 当前系统 %s/%s): ", runtime.GOOS, runtime.GOARCH))

	numbers, _ := reader.ReadString('\n')
	selections := strings.Fields(numbers)
	selectedPlatforms := make(map[builder.Platform]bool)

	if len(selections) == 0 {
		// 默认情况：使用当前系统平台
		defaultPlatform := builder.Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
		// 确保默认平台是受支持的
		isSupported := false
		for _, p := range builder.SupportedPlatforms {
			if p == defaultPlatform {
				isSupported = true
				break
			}
		}
		if isSupported {
			selectedPlatforms[defaultPlatform] = true
			display.PrintInfo(fmt.Sprintf("未输入，使用默认平台: %s/%s", defaultPlatform.OS, defaultPlatform.Arch))
		} else {
			display.PrintWarning(fmt.Sprintf("当前系统平台 %s/%s 不在支持列表中，请手动选择。", defaultPlatform.OS, defaultPlatform.Arch))
		}
	} else {
		for _, num := range selections {
			switch num {
			case "1":
				selectedPlatforms[builder.Platform{OS: "windows", Arch: "amd64"}] = true
			case "2":
				selectedPlatforms[builder.Platform{OS: "windows", Arch: "arm64"}] = true
			case "3":
				selectedPlatforms[builder.Platform{OS: "linux", Arch: "amd64"}] = true
			case "4":
				selectedPlatforms[builder.Platform{OS: "linux", Arch: "arm64"}] = true
			case "5":
				selectedPlatforms[builder.Platform{OS: "windows", Arch: "amd64"}] = true
				selectedPlatforms[builder.Platform{OS: "linux", Arch: "amd64"}] = true
			case "6":
				config.BuildAllPlatforms = true
			default:
				display.PrintWarning(fmt.Sprintf("无效的平台编号: %s", num))
			}
		}
	}

	if config.BuildAllPlatforms {
		display.PrintSuccess("将构建所有支持的平台和架构")
	} else if len(selectedPlatforms) > 0 {
		var platforms []builder.Platform
		display.PrintEmptyLine()
		display.PrintHeader("已选择的平台:")
		// 按照支持列表的顺序添加，保证输出一致性
		for _, p := range builder.SupportedPlatforms {
			if selectedPlatforms[p] {
				platforms = append(platforms, p)
				display.PrintInfo(fmt.Sprintf("- %s %s", p.OS, p.Arch))
			}
		}
		config.Platforms = platforms
	} else {
		// 如果没有有效选择且不是构建全部，则提示并退出或采取默认行为
		display.PrintWarning("未选择任何有效平台，将不执行任何构建。")
		return nil // 或者可以设置一个默认值
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
