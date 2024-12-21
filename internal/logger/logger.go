package logger

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
)

type Logger struct {
	isDebug bool
}

func New(debug bool) *Logger {
	return &Logger{
		isDebug: debug,
	}
}

func (l *Logger) Error(err string, reason string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println("-=-=-=-=--=-=-=- Error -=-=-=-=--=-=-=-")
	color.Red(fmt.Sprintf("Error: %s", err))
	color.Yellow(fmt.Sprintf("File: %s", file))
	color.Blue(fmt.Sprintf("Line: %d", line))
	color.Green(fmt.Sprintf("Reason: %s", reason))
	fmt.Println("-=-=-=-=--=-=-=-=-=-=--=-=-=-=-=-=-=-")
}

func (l *Logger) Info(info string, reason string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println("-=-=-=-=--=-=-=-=-=-=--=-=-=-=-=-=-=-")
	color.Green(fmt.Sprintf("Info: %s", info))
	color.Yellow(fmt.Sprintf("File: %s", file))
	color.Blue(fmt.Sprintf("Line: %d", line))
	color.Green(fmt.Sprintf("Reason: %s", reason))
	fmt.Println("-=-=-=-=--=-=-=-=-=-=--=-=-=-=-=-=-=-")
}
