package syslogger

import (
	"fmt"
	"os"
	"time"
)

const (
	// changeme!
	logFileMode = "syslogger-2006-01-02_150405.log"

	LogDestinationStderr = 1
	LogDestinationCsvlog = 8
)

type Logger struct {
	config Config
	utils  Utils

	location          *time.Location
	filename          string
	syslogFile        *os.File
	LogRotationMin    int
	nextRotationTime  int64
	LogRotationMb     int64
	rotationRequested bool
	timeBasedRotation bool
	sizeRotationFor   int
}

type Config struct {
	logDirectory   string
	logFilename    string
	logFileMode    int64
	logRotationMin int
	logRotationMb  int64
	logLinePrefix  string
	logTimezone    string
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
	l.LogRotationMin = conf.logRotationMin
	l.LogRotationMb = conf.logRotationMb


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
	if l.LogRotationMin <= 0 {
		return
	}
	// convert to seconds
	var rotinterval int64
	rotinterval = int64(l.LogRotationMin) * 60
	now := time.Now().In(l.location).Unix()
	now -= now % rotinterval
	now += rotinterval
	l.nextRotationTime = now
}

func (l *Logger) SysLoggerMain(s string) {

	l.setNextRotationTime()

	for {
		if l.LogRotationMin > 0 {
			now := time.Now().In(l.location).Unix()
			if now >= l.nextRotationTime {
				l.rotationRequested = true
				l.timeBasedRotation = true
			}
		}

		if !l.rotationRequested && l.LogRotationMb > 0 {
			if l.getFileSize(l.filename) >= l.LogRotationMb * 1024 {
				l.rotationRequested = true
				l.sizeRotationFor |= LogDestinationStderr
			}
			//TODO CSV

		}

		if l.rotationRequested {
			l.logfileRotate(l.timeBasedRotation, l.sizeRotationFor)
		}
	}


}

func (l *Logger) logfileRotate(timeBasedRotation bool, sizeRotationFor int){

}

func (l *Logger) getFileSize(path string) int64 {
	file, err := os.Stat(path);
	if err != nil {
		return 0
	}
	return file.Size()
}
