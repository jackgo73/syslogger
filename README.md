
# syslogger

Syslogger is a logger for Go, which supports **rolling log files by time and space**.

When the syslogger is started, a goroutine is automatically created to receive data from the pipe serial and write to the right log file.

User can concurrently call the log write API(`Debug(), Log(), Warning(), Error()`), and concurrently write log data to pipe, ensure concurrent write performance.

When the pipe is full, the log write API will be blocked. You can alleviate this problem by increasing the size of the `LogBufferLength` parameter.

The implementation comes from the postgresql log system.

## Configuration

Configuration uses the `syslogger.Config` struce , so don't have to worry about possible mismatches. The config can be roughly guessed and filled in. You can usually use the configuration in the example and then modify it according to your needs.

- LogDirectory(string)

  This parameter determines the directory in which log files will be created(absolute path).

  - Example:`"/tmp/logger"`

- LogFilename(string)

  This parameter sets the file names of the created log files. The value is treated as a `strftime` pattern.

  - Example:`syslogger-2006-01-02_150405.log`

- LogBufferLength

  This parameter sets the size of the log write buffer, if the log generation rate is fast, you can increase the parameter appropriately.

  - Example:`256`

- LogRotationMin

  This parameter determines the maximum lifetime of an individual log file. After this many minutes have elapsed, a new log file will be created. Set to zero to disable time-based creation of new log files.

  - Example:`1`

- LogRotationMb

  This parameter determines the maximum size of an individual log file. After this many kilobytes have been emitted into a log file, a new log file will be created.

  - Example:`10`

- LogTruncateOnRotation

  This parameter will cause sys logger to truncate (overwrite), rather than append to, any existing log file of the same name.

  - Example:`true`

- LogLinePrefix

  This is a `printf-style` string that is output at the beginning of each log line. `%` characters begin “escape sequences” that are replaced with status information as outlined below.

  | **Escape** | **Effect**                                                  |
  | ---------- | ----------------------------------------------------------- |
  | %f         | The file name where the current code is located             |
  | %p         | The full path to the file where the current code is located |
  | %l         | Line number of the current code                             |
  | %t         | Time like "2006-01-02 15:04:05"                             |
  | %m         | Time in milliseconds like "2006-01-02 15:04:05.000"         |
  | %n         | Unix timestamp like 1563610920                              |

  - Example:`%f %l %t %n`
  - Output:`star.go 27 2019-07-20 20:27:48 1563625668 DEBUG: this is debug!`

- LogOutputMode

  This is a `printf-style` string which determines where the log is output.

  | Escape | Effect           |
  | ------ | ---------------- |
  | %f     | Output to file   |
  | %o     | Output to STDOUT |
  | %e     | Output to STDERR |

  - Example:`%f %o`
  - Output to file and STDOUT

## Example

[star](example/star.go)

```go
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
```

output in STDOUT

```sh
do something
star.go 27 2019-07-20 21:02:20 1563627740 DEBUG: this is debug!
star.go 28 2019-07-20 21:02:20 1563627740 LOG: step1!
star.go 29 2019-07-20 21:02:20 1563627740 LOG: step2!
star.go 30 2019-07-20 21:02:20 1563627740 LOG: step3!
star.go 31 2019-07-20 21:02:20 1563627740 WARNING: warning!
star.go 32 2019-07-20 21:02:20 1563627740 ERROR: something wrong!
```

Output in File

```sh
$ ls
star-2019-07-20_210220.log

$ cat star-2019-07-20_210220.log
star.go 27 2019-07-20 21:02:20 1563627740 DEBUG: this is debug!
star.go 28 2019-07-20 21:02:20 1563627740 LOG: step1!
star.go 29 2019-07-20 21:02:20 1563627740 LOG: step2!
star.go 30 2019-07-20 21:02:20 1563627740 LOG: step3!
star.go 31 2019-07-20 21:02:20 1563627740 WARNING: warning!
star.go 32 2019-07-20 21:02:20 1563627740 ERROR: something wrong!
```

## Development status

`developing`



## Author

`GaoMingjie`

`jackgo73@outlook.com`