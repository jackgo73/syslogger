package syslogger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var sysLogger *Logger

const (
	LogFileMode          = "syslogger-2006-01-02_150405.log"
	LogBufferLength      = 128
	LogDestinationStderr = 1
	//LogDestinationCsvlog = 8
	MaxInt64 = int64(^uint64(0) >> 1)
	SEP      = " "
	DEBUG    = "DEBUG"
	LOG      = "LOG"
	WARNING  = "WARNING"
	ERROR    = "ERROR"
)

type Logger struct {
	config     Config
	loggerChan chan *LoggerElem
	loggerStop bool

	location               *time.Location
	lastFileName           string
	syslogFile             *os.File
	LogRotationMin         int
	firstSysLoggerFileTime time.Time
	nextRotationTime       time.Time
	LogRotationMb          int64
	rotationRequested      bool
	logTruncateOnRotation  bool
	sizeRotationFor        int
	logLinePrefix          string
	logOutputMode          string
}

type LoggerElem struct {
	prefix string
	level  string
	edata  string
}

type Config struct {
	LogDirectory          string
	LogFilename           string
	LogFileMode           int64
	LogBufferLength       int
	LogRotationMin        int
	LogRotationMb         int64
	LogTruncateOnRotation bool
	LogLinePrefix         string
	LogTimezone           string
	LogOutputMode         string
}

func Init(conf Config) {
	var l Logger
	l.loggerStop = true

	l.config = conf
	// check parameter
	if conf.LogBufferLength == 0 {
		conf.LogBufferLength = LogBufferLength
	} else if len(conf.LogFilename) == 0 {
		conf.LogFilename = LogFileMode
	}

	l.location, _ = time.LoadLocation(l.config.LogTimezone)

	// Create log directory if not present; ignore errors
	_ = os.Mkdir(l.config.LogDirectory, os.ModePerm)

	l.firstSysLoggerFileTime = l.getLcNowTime()
	filename := l.logFileGetName(l.firstSysLoggerFileTime, "")
	syslogFile, err := l.logFileOpen(filename, os.O_RDWR|os.O_CREATE, l.config.LogFileMode)
	if err != nil {
		Debug(fmt.Sprintf("could not open log file %s: %s", filename, err.Error()))
		return
	}
	l.lastFileName = filename
	l.syslogFile = syslogFile
	l.LogRotationMin = conf.LogRotationMin
	l.LogRotationMb = conf.LogRotationMb
	l.logTruncateOnRotation = conf.LogTruncateOnRotation
	l.loggerChan = make(chan *LoggerElem, conf.LogBufferLength)
	l.logLinePrefix = conf.LogLinePrefix
	l.logOutputMode = conf.LogOutputMode

	sysLogger = &l
}

func Run() {
	go sysLogger.SysLoggerMain()
}

func Exit(force bool) {
	if force {
		sysLogger.loggerStop = true
		sysLogger.loggerChan <- nil
	} else {
		sysLogger.loggerChan <- nil
		for !sysLogger.loggerStop {
			time.Sleep(time.Millisecond)
		}
	}
}

func (l *Logger) SysLoggerClose() {
	_ = l.syslogFile.Close()
}

// construct logfile name using timestamp information
// If suffix isn't NULL, append it to the name, replacing
// any ".log" that may be in the pattern.
func (l *Logger) logFileGetName(t time.Time, suffix string) string {
	timeStr := t.Format(l.config.LogFilename)
	if len(suffix) == 0 {
		return fmt.Sprintf("%s/%s", l.config.LogDirectory, timeStr)
	} else {
		return fmt.Sprintf("%s/%s%s", l.config.LogDirectory, timeStr[:len(timeStr)-4], suffix)
	}

}

func (l *Logger) getLcNowTime() time.Time {
	return time.Now().In(l.location)
}

// Open a new logfile with proper permissions and buffering options.
func (l *Logger) logFileOpen(filename string, flag int, mode int64) (*os.File, error) {
	f, err := os.OpenFile(filename, flag, os.FileMode(mode))
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
	now := l.getLcNowTime().Unix()
	now -= now % rotinterval
	now += rotinterval
	l.nextRotationTime = time.Unix(now, 0)
}

func (l *Logger) SysLoggerMain() {

	l.loggerStop = false
	l.lastFileName = l.logFileGetName(l.firstSysLoggerFileTime, "")
	l.setNextRotationTime()

	for {

		var timeBasedRotation bool
		var curTimeoutMs int64
		var loggerElem *LoggerElem

		if l.LogRotationMin > 0 {
			now := time.Now().In(l.location).Unix()
			if now >= l.nextRotationTime.Unix() {
				l.rotationRequested = true
				timeBasedRotation = true
			}
		}

		if !l.rotationRequested && l.LogRotationMb > 0 {
			if l.getFileSize(l.syslogFile) >= l.LogRotationMb*1024 {
				l.rotationRequested = true
				l.sizeRotationFor |= LogDestinationStderr
			}
			//TODO CSV
		}

		if l.rotationRequested {
			l.logfileRotate(timeBasedRotation, l.sizeRotationFor)
		}

		// Calculate time till next time-based rotation, so that we don't
		// sleep longer than that.
		if l.LogRotationMin > 0 {
			delay := l.nextRotationTime.Unix() - l.getLcNowTime().Unix()
			if delay > 0 {
				if delay > MaxInt64/1000 {
					delay = MaxInt64 / 1000
				}
				curTimeoutMs = delay * 1000
			} else {
				curTimeoutMs = 0
			}
		} else {
			curTimeoutMs = -1
		}

		//Sleep until there's something to do
		select {
		case loggerElem = <-l.loggerChan:
			if loggerElem == nil {
				l.loggerStop = true
				break
			}
			l.processPipeInput(loggerElem)
			break
		case <-time.After(time.Duration(curTimeoutMs) * time.Millisecond):
			Debug(fmt.Sprintf("debug: sleep for %d\n", curTimeoutMs))
			break
		}

		// Normal exit from the syslogger
		if l.loggerStop {
			break
		}
	}
}

func (l *Logger) processPipeInput(e *LoggerElem) {
	s := e.prefix + e.level + e.edata + "\n"

	for _, p := range strings.Split(l.logOutputMode, " ") {
		// format error, ignore it
		if !strings.HasPrefix(p, "%") || len(p) != 2 {
			continue
		}
		switch p {
		case "%f":
			_, _ = l.syslogFile.WriteString(s)
		case "%o":
			_, _ = fmt.Fprintf(os.Stdout, s)
		case "%e":
			_, _ = fmt.Fprintf(os.Stderr, s)
		}
	}
}

func (l *Logger) logfileRotate(timeBasedRotation bool, sizeRotationFor int) {

	var fntime time.Time
	var filename string
	var fh *os.File
	var err error

	l.rotationRequested = false

	if timeBasedRotation {
		fntime = l.nextRotationTime
	} else {
		fntime = l.getLcNowTime()
	}
	filename = l.logFileGetName(fntime, "")

	if timeBasedRotation || ((l.sizeRotationFor & LogDestinationStderr) != 0) {
		if l.logTruncateOnRotation && timeBasedRotation && filename != l.lastFileName {
			fh, err = l.logFileOpen(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, l.config.LogFileMode)
		} else {
			fh, err = l.logFileOpen(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, l.config.LogFileMode)
		}

		if err != nil {
			Debug("create new log file failed, keep using the old file till get a new one: " + filename)
			return
		}

		_ = l.syslogFile.Close()
		l.syslogFile = fh
		l.lastFileName = filename
	}

	l.setNextRotationTime()

	fmt.Println(filename)
}

func (l *Logger) getFileSize(f *os.File) int64 {
	file, err := f.Stat()
	if err != nil {
		return 0
	}
	return file.Size()
}

// several functions used by the client
func (l *Logger) generateLogLinePrefix() string {
	if len(l.logLinePrefix) == 0 {
		return ""
	}

	var buf string

	for _, p := range strings.Split(l.logLinePrefix, " ") {
		// format error, ignore it
		if !strings.HasPrefix(p, "%") || len(p) != 2 {
			continue
		}
		switch p {
		case "%f":
			_, file, _, ok := runtime.Caller(3)
			if ok {
				buf += filepath.Base(file) + SEP
			}
		case "%p":
			_, file, _, ok := runtime.Caller(3)
			if ok {
				buf += file + SEP
			}
		case "%l":
			_, _, line, ok := runtime.Caller(3)
			if ok {
				buf += strconv.Itoa(line) + SEP
			}
		case "%t":
			buf += l.getLcNowTime().Format("2006-01-02 15:04:05") + SEP
		case "%m":
			buf += l.getLcNowTime().Format("2006-01-02 15:04:05.000") + SEP
		case "%n":
			buf += strconv.FormatInt(l.getLcNowTime().Unix(), 10) + SEP
		}
	}

	return buf
}

func (l *Logger) sendMessageToServerLog(estr string, level string) {

	elem := LoggerElem{
		prefix: sysLogger.generateLogLinePrefix(),
		level:  level + ":" + SEP,
		edata:  estr,
	}
	sysLogger.loggerChan <- &elem
}

func Debug(estr string) {
	if sysLogger == nil {
		fmt.Println("Syslogger:" + SEP + estr)
		return
	}
	sysLogger.sendMessageToServerLog(estr, DEBUG)
}

func Log(estr string) {
	sysLogger.sendMessageToServerLog(estr, LOG)
}

func Warning(estr string) {
	sysLogger.sendMessageToServerLog(estr, WARNING)
}

func Error(estr string) {
	sysLogger.sendMessageToServerLog(estr, ERROR)
}
