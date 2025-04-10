package builder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Platform 定义目标平台
type Platform struct {
	OS   string
	Arch string
}

// BuildConfig 构建配置
type BuildConfig struct {
	OutputDir         string
	CGOEnabled        bool
	Platforms         []Platform
	PackageName       string
	BuildAllPlatforms bool
}

// 支持的平台列表
var SupportedPlatforms = []Platform{
	{"windows", "amd64"},
	{"windows", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
}

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

// Builder 构建器
type Builder struct {
	config      BuildConfig
	originalEnv map[string]string
}

// NewBuilder 创建新的构建器实例
func NewBuilder(config BuildConfig) *Builder {
	return &Builder{
		config: config,
		originalEnv: map[string]string{
			"GOOS":        getGoEnv("GOOS"),
			"GOARCH":      getGoEnv("GOARCH"),
			"CGO_ENABLED": getGoEnv("CGO_ENABLED"),
		},
	}
}

// printEnvInfo 打印环境信息
func (b *Builder) printEnvInfo(title string, goos, goarch, cgo string) {
	fmt.Printf("\n%s\n", title)
	fmt.Printf("====================================\n")
	fmt.Printf("GOOS        = %s\n", goos)
	fmt.Printf("GOARCH      = %s\n", goarch)
	fmt.Printf("CGO_ENABLED = %s\n", cgo)
	fmt.Printf("====================================\n")
}

// Build 执行构建过程
func (b *Builder) Build() error {
	// 创建输出目录
	if err := os.MkdirAll(b.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 如果选择构建所有平台
	if b.config.BuildAllPlatforms {
		b.config.Platforms = SupportedPlatforms
	}

	// 对每个目标平台执行构建
	for _, platform := range b.config.Platforms {
		if err := b.buildForPlatform(platform); err != nil {
			return fmt.Errorf("构建 %s/%s 失败: %v", platform.OS, platform.Arch, err)
		}
	}

	// 恢复原始环境
	if err := b.restoreEnv(); err != nil {
		return err
	}

	return nil
}

// buildForPlatform 为特定平台执行构建
func (b *Builder) buildForPlatform(platform Platform) error {
	fmt.Printf("\n----- 正在构建 %s %s 版本 -----\n", platform.OS, platform.Arch)

	// 设置环境变量
	os.Setenv("GOOS", platform.OS)
	os.Setenv("GOARCH", platform.Arch)
	cgoValue := "0"
	if b.config.CGOEnabled {
		cgoValue = "1"
	}
	os.Setenv("CGO_ENABLED", cgoValue)

	// 确认环境变量已正确设置
	currentGOOS := getGoEnv("GOOS")
	currentGOARCH := getGoEnv("GOARCH")
	currentCGO := getGoEnv("CGO_ENABLED")
	
	// 输出当前构建环境
	b.printEnvInfo("构建环境配置", currentGOOS, currentGOARCH, currentCGO)

	// 构建输出文件名
	outputName := fmt.Sprintf("%s_%s_%s", b.config.PackageName, platform.OS, platform.Arch)
	if platform.OS == "windows" {
		outputName += ".exe"
	}
	outputPath := filepath.Join(b.config.OutputDir, outputName)

	// 执行构建命令
	cmd := exec.Command("go", "build", "-ldflags", "-w -s", "-o", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("构建失败: %v", err)
	}

	fmt.Printf("[成功] %s/%s 构建完成: %s\n", platform.OS, platform.Arch, outputPath)
	return nil
}

// restoreEnv 恢复原始环境变量
func (b *Builder) restoreEnv() error {
	fmt.Println("\n----- 正在恢复原始环境 -----")
	
	for key, value := range b.originalEnv {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("恢复环境变量 %s 失败: %v", key, err)
		}
	}
	
	// 验证环境已恢复
	currentGOOS := getGoEnv("GOOS")
	currentGOARCH := getGoEnv("GOARCH")
	currentCGO := getGoEnv("CGO_ENABLED")
	
	// 输出恢复后的环境
	fmt.Printf("\n环境已恢复\n")
	fmt.Printf("====================================\n")
	fmt.Printf("GOOS        = %s\n", currentGOOS)
	fmt.Printf("GOARCH      = %s\n", currentGOARCH)
	fmt.Printf("CGO_ENABLED = %s\n", currentCGO)
	fmt.Printf("====================================\n")
	
	return nil
} 