package main

import (
	"fmt"
	log "github.com/mutex73/syslogger"
)

func main() {
	LogFileMode := "star-2006-01-02_150405.log"
	log.Init(
		log.Config{
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
		})
	log.Run()

	fmt.Println("do something")
	// start using and have fun!
	log.Debug("this is debug!")
	log.Log("step1!")
	log.Log("step2!")
	log.Log("step3!")
	log.Warning("warning!")
	log.Error("something wrong!")


	// false means waiting for all data to be processed before exiting
	log.Exit(false)
}
