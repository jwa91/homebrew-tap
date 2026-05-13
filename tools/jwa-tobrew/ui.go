package main

import (
	"fmt"
	"os"
)

const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cGray   = "\033[90m"
)

func useColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fi, _ := os.Stderr.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func paint(c, s string) string {
	if !useColor() {
		return s
	}
	return c + s + cReset
}

func info(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cBlue, "→"), fmt.Sprintf(format, a...))
}
func ok(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cGreen, "✓"), fmt.Sprintf(format, a...))
}
func warn(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cYellow, "!"), fmt.Sprintf(format, a...))
}
func fail(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cRed, "✗"), fmt.Sprintf(format, a...))
}
func hint(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cGray, "  "+fmt.Sprintf(format, a...)))
}
func banner(format string, a ...any) {
	fmt.Fprintln(os.Stderr, paint(cBold, fmt.Sprintf(format, a...)))
}
