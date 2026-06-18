package style

import (
	"fmt"
	"os"
)

var (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	noColor     = false
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	
	stat, _ := os.Stdout.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		noColor = true
	}
}

func format(prefix, color, formatStr string, args ...interface{}) string {
	msg := fmt.Sprintf(formatStr, args...)
	if noColor {
		return fmt.Sprintf("[%s] %s", prefix, msg)
	}
	return fmt.Sprintf("%s[%s]%s %s", color, prefix, colorReset, msg)
}

func OK(formatStr string, args ...interface{}) {
	fmt.Println(format("OK", colorGreen, formatStr, args...))
}

func Error(formatStr string, args ...interface{}) {
	fmt.Println(format("ERROR", colorRed, formatStr, args...))
}

func Info(formatStr string, args ...interface{}) {
	fmt.Println(format("INFO", colorCyan, formatStr, args...))
}

func Warn(formatStr string, args ...interface{}) {
	fmt.Println(format("WARN", colorYellow, formatStr, args...))
}

func Separator() {
	fmt.Println("--------------------------------------------------")
}
