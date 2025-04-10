package display

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// GetTerminalWidth 获取终端宽度
func GetTerminalWidth() int {
	if runtime.GOOS == "windows" {
		return getWindowsTerminalWidth()
	} else {
		return getUnixTerminalWidth()
	}
}

// 获取 Windows 终端宽度
func getWindowsTerminalWidth() int {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleScreenBufferInfo := kernel32.NewProc("GetConsoleScreenBufferInfo")
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return 80 // 默认宽度
	}

	var csbi struct {
		size struct {
			x, y uint16
		}
		cursorPosition struct {
			x, y uint16
		}
		attributes uint16
		window     struct {
			left, top, right, bottom uint16
		}
		maximumWindowSize struct {
			x, y uint16
		}
	}

	ret, _, _ := getConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi)))
	if ret == 0 {
		return 80 // 默认宽度
	}

	return int(csbi.window.right - csbi.window.left + 1)
}

// 获取 Unix 系统终端宽度 (Linux/macOS)
func getUnixTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 80 // 默认宽度
	}

	parts := strings.Split(strings.TrimSpace(string(output)), " ")
	if len(parts) != 2 {
		return 80 // 默认宽度
	}

	width, err := strconv.Atoi(parts[1])
	if err != nil {
		return 80 // 默认宽度
	}

	return width
}

// PrintCenteredTitle 居中显示标题
func PrintCenteredTitle(title string) {
	width := GetTerminalWidth()
	titleWidth := 0
	for _, r := range title {
		if r > 0x7F {
			titleWidth += 2 // 中文字符占用两个字符宽度
		} else {
			titleWidth += 1 // 英文字符占用一个字符宽度
		}
	}

	padding := (width - titleWidth) / 2
	if padding < 0 {
		padding = 0
	}
	paddedTitle := strings.Repeat(" ", padding) + title
	fmt.Println(paddedTitle)
}

// PrintDivider 显示主分隔线
func PrintDivider() {
	width := GetTerminalWidth()
	divider := strings.Repeat("═", width)
	fmt.Println(divider)
}

// PrintSubDivider 显示次级分隔线
func PrintSubDivider() {
	width := GetTerminalWidth()
	divider := strings.Repeat("─", width)
	fmt.Println(divider)
}

// PrintEmptyLine 打印空行
func PrintEmptyLine() {
	fmt.Println()
}

// PrintInfo 打印信息行
func PrintInfo(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// PrintSection 打印带有分隔线的区块
func PrintSection(title string) {
	PrintEmptyLine()
	PrintDivider()
	PrintCenteredTitle(title)
	PrintDivider()
}

// PrintSubSection 打印带有次级分隔线的子区块
func PrintSubSection(title string) {
	PrintEmptyLine()
	PrintSubDivider()
	PrintCenteredTitle(title)
	PrintSubDivider()
}

// PrintSectionEnd 打印区块结束
func PrintSectionEnd() {
	PrintDivider()
}

// PrintFieldValue 打印字段和值
func PrintFieldValue(field, value string) {
	fmt.Printf("%-15s = %s\n", field, value)
}

// PrintSuccess 打印成功信息
func PrintSuccess(message string) {
	fmt.Printf("[成功] %s\n", message)
}

// PrintWarning 打印警告信息
func PrintWarning(message string) {
	fmt.Printf("[警告] %s\n", message)
}

// PrintError 打印错误信息
func PrintError(message string) {
	fmt.Printf("[错误] %s\n", message)
} 