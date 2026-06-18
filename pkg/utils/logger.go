package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
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

	// Silent suppresses all console output except matches and downloads
	// (the positive hits). Everything is still written to the log file.
	Silent bool

	// progressBar, when set, is kept pinned at the bottom of the terminal.
	// All console output is funneled through printAboveBar so log lines are
	// drawn above the bar instead of corrupting it. progressMu also serializes
	// concurrent writers (many goroutines log at once).
	progressBar *progressbar.ProgressBar
	progressMu  sync.Mutex
)

// StartProgress creates a progress bar pinned to the bottom of the terminal and
// registers it so log output is rendered above it. Pass the total number of
// units of work. It is a no-op in silent mode.
func StartProgress(total int) {
	if Silent {
		return
	}
	progressMu.Lock()
	defer progressMu.Unlock()
	progressBar = progressbar.NewOptions(total,
		// Render to the same stream the colored logs use so clears/redraws
		// stay in sync with our other output.
		progressbar.OptionSetWriter(color.Output),
		progressbar.OptionSetDescription("Spidering targets"),
		progressbar.OptionSetWidth(20),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(color.Output, "\n")
		}),
		progressbar.OptionFullWidth(),
	)
}

// AdvanceProgress moves the progress bar forward by one unit.
func AdvanceProgress() {
	progressMu.Lock()
	defer progressMu.Unlock()
	if progressBar != nil {
		progressBar.Add(1)
	}
}

// FinishProgress clears the bar from the terminal (used on shutdown).
func FinishProgress() {
	progressMu.Lock()
	defer progressMu.Unlock()
	if progressBar != nil {
		progressBar.Finish()
		progressBar = nil
	}
}

// printAboveBar serializes console writes and ensures the progress bar (if any)
// is cleared before the line is printed and redrawn afterwards, keeping it
// pinned to the bottom of the terminal.
func printAboveBar(emit func()) {
	progressMu.Lock()
	defer progressMu.Unlock()
	if progressBar != nil {
		progressBar.Clear()
	}
	emit()
	if progressBar != nil {
		progressBar.RenderBlank()
	}
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
		fmt.Fprintf(logFile, "%s [%s] %s\n", ts, level, strings.TrimSpace(msg))
	}
}

// LogDownload reports a successful download. This is a "positive hit" and is
// always shown on the console, even in silent mode.
func LogDownload(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("DOWNLOAD", msg)
	printAboveBar(func() {
		// Indented with spaces, Yellow color.
		color.New(color.FgYellow).Printf("    [+] %s\n", msg)
	})
}

// LogSuccess reports a match. This is a "positive hit" and is always shown on
// the console, even in silent mode.
func LogSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("SUCCESS", msg)
	printAboveBar(func() {
		Success("[+] " + msg + "\n")
	})
}

func LogInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("INFO", msg)
	if Silent {
		return
	}
	printAboveBar(func() {
		Info("[INFO] " + msg + "\n")
	})
}

func LogWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("WARNING", msg)
	if Silent {
		return
	}
	printAboveBar(func() {
		Warning("[!] " + msg + "\n")
	})
}

func LogError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	logToFile("ERROR", msg)
	if Silent {
		return
	}
	printAboveBar(func() {
		Error("[-] " + msg + "\n")
	})
}

func LogDebug(format string, a ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		msg := fmt.Sprintf(format, a...)
		logToFile("DEBUG", msg)
		if Silent {
			return
		}
		printAboveBar(func() {
			Debug("[DEBUG] " + msg + "\n")
		})
	}
}
