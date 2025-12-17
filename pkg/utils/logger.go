package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	// Define colors
	Info    = color.New(color.FgCyan).PrintfFunc()
	Success = color.New(color.FgGreen).PrintfFunc()
	Warning = color.New(color.FgHiBlack).PrintfFunc() // Gray/Dark Gray
	Error   = color.New(color.FgRed).PrintfFunc()
	Debug   = color.New(color.FgHiBlack).PrintfFunc()

	// Bold helper
	Bold = color.New(color.Bold).SprintFunc()

	logFile *os.File
)

// ... (init/close/logToFile same)

func LogDownload(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("DOWNLOAD", msg)
	// Indented with Tab, Yellow color
	// User asked for "Yellow with an indention"
	color.New(color.FgYellow).Printf("    [+] %s\n", msg)
}

func InitLogger(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func logToFile(level string, msg string) {
	if logFile != nil {
		ts := time.Now().Format("2006/01/02 15:04:05")
		// Strip newlines from msg for cleaner log? No, keep as is.
		fmt.Fprintf(logFile, "%s [%s] %s\n", ts, level, strings.TrimSpace(msg))
	}
}

func LogInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("INFO", msg)
	Info("[INFO] " + msg + "\n")
}

func LogSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("SUCCESS", msg)
	Success("[+] " + msg + "\n")
}

func LogWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("WARNING", msg)
	Warning("[!] " + msg + "\n")
}

func LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("ERROR", msg)
	Error("[-] " + msg + "\n")
}

func LogDebug(format string, a ...interface{}) {
	// Debug goes to file mostly if we want deep traces, but user asked for "timestamps of things that get downloaded".
	// Debug might be too noisy for default log?
	// The user said "include a log file... This should include timestamps of things that get downloaded".
	// So LogSuccess/Info is most key.
	// We'll log Debug to file only if verbose or debug env is set?
	// Or maybe just always log it but keep console clean?
	// Let's log to file if it happens.

	if os.Getenv("DEBUG") == "true" {
		msg := fmt.Sprintf(format, a...)
		logToFile("DEBUG", msg)
		Debug("[DEBUG] " + msg + "\n")
	}
}
