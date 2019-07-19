package syslogger

import (
	"fmt"
	"os"
	"time"
)

const (
	// changeme!
	logFileMode = "syslogger-2006-01-02_150405.log"
)

type Logger struct {
	config Config
	utils  Utils

	location          *time.Location
	filename          string
	syslogFile        *os.File
	LogRotationAge    int
	nextRotationTime  int64
	rotationRequested bool
	timeBasedRotation bool
}

type Config struct {
	logDirectory    string
	logFilename     string
	logFileMode     int64
	logRotationAge  int
	logRotationSize string
	logLinePrefix   string
	logTimezone     string
}

type Utils struct {
}

func (l *Logger) SysLoggerInit(conf Config) {
	l.config = conf
	l.location, _ = time.LoadLocation(l.config.logTimezone)

	// Create log directory if not present; ignore errors
	_ = os.Mkdir(l.config.logDirectory, os.ModePerm)

	firstSysLoggerFileTime := time.Now().Unix()
	filename := l.logFileGetName(firstSysLoggerFileTime, "")
	syslogFile, err := l.logFileOpen(filename, l.config.logFileMode)
	if err != nil {
		fmt.Printf("could not open log file %s: %s", filename, err.Error())
		return
	}
	l.filename = filename
	l.syslogFile = syslogFile
	l.LogRotationAge = conf.logRotationAge

	l.setNextRotationTime()
	fmt.Println(l.filename)
}

func (l *Logger) SysLoggerClose() {
	_ = l.syslogFile.Close()
}

// construct logfile name using timestamp information
// If suffix isn't NULL, append it to the name, replacing
// any ".log" that may be in the pattern.
func (l *Logger) logFileGetName(timestamp int64, suffix string) string {
	timeStr := time.Now().In(l.location).Format(l.config.logFilename)
	if len(suffix) == 0 {
		return fmt.Sprintf("%s/%s", l.config.logDirectory, timeStr)
	} else {
		return fmt.Sprintf("%s/%s%s", l.config.logDirectory, timeStr[:len(timeStr)-4], suffix)
	}

}

// Open a new logfile with proper permissions and buffering options.
func (l *Logger) logFileOpen(filename string, mode int64) (*os.File, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.FileMode(mode))
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Determine the next planned rotation time, and store in next_rotation_time.
func (l *Logger) setNextRotationTime() {
	if l.LogRotationAge <= 0 {
		return
	}
	// convert to seconds
	var rotinterval int64
	rotinterval = int64(l.LogRotationAge) * 60
	now := time.Now().In(l.location).Unix()
	now -= now % rotinterval
	now += rotinterval
	l.nextRotationTime = now
}

func (l *Logger) write(s string) {

	if l.LogRotationAge > 0 {
		now := time.Now().In(l.location).Unix()
		if now >= l.nextRotationTime {
			l.rotationRequested = true
			l.timeBasedRotation = true
		}
	}
}
