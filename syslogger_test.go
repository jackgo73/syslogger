package syslogger

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestSysLoggerStart(t *testing.T) {

	config := Config{
		LogDirectory:          "/tmp/logger",
		LogFilename:           LogFileMode,
		LogBufferLength:       256,
		LogFileMode:           0600,
		LogTimezone:           "Local",
		LogRotationMin:        1,
		LogRotationMb:         10,
		LogTruncateOnRotation: true,
		LogLinePrefix:         "%f %l %t %n",
		LogOutputMode:         "%f %o",
	}
	Debug("test debug before init!!")
	Init(config)
	Run()
	Debug("test debug!!")
	Log("test log!!")
	Warning("test warning!!")
	Error("test error!!")
	time.Sleep(time.Second)
	sysLogger.loggerChan <- nil
	time.Sleep(time.Second)
}

func TestFileMode(t *testing.T) {
	mode := os.ModeDir
	fmt.Printf("%b\n", mode)
	mode = os.ModeAppend
	fmt.Printf("%b\n", mode)
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now().Format(LogFileMode))
}

func TestIntMax(t *testing.T) {
	const MaxInt64 = int64(^uint64(0) >> 1)
	fmt.Println(MaxInt64)
}

func TestLineNo(t *testing.T) {
	_, file, line, ok := runtime.Caller(0)
	fmt.Println(file)
	fmt.Println(line)
	fmt.Println(ok)
}
