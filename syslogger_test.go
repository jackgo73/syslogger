package syslogger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSysLoggerStart(t *testing.T) {

	logger := Logger{}
	config := Config{
		logDirectory:"/tmp/logger",
		logFilename:logFileMode,
		logFileMode:0600,
		logTimezone:"Local",
		logRotationAge:2,
	}
	logger.SysLoggerInit(config)
}


func TestFileMode(t *testing.T) {
	mode := os.ModeDir
	fmt.Printf("%b\n", mode)
	mode = os.ModeAppend
	fmt.Printf("%b\n", mode)
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now().Format(logFileMode))
}

