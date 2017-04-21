// Internal logger system. Uses to change output format and use nanosec when logging time
package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const (
	// Enables debug output. You can setup environment variable "DEBUG_SATORI_SDK=true" to activate debug mode
	DEBUG_SATORI_SDK = false

	WARN  = "[warn] "
	ERR   = "[erro] "
	FATAL = "[fatl] "
	DEBUG = "[debg] "
	INFO  = "[info] "
)

// Logs message with [info] header
func Info(v ...interface{}) {
	fmt.Print(getHeader(INFO))
	fmt.Println(v...)
}

// Logs message with [warn] header
func Warn(v ...interface{}) {
	fmt.Print(getHeader(WARN))
	fmt.Println(v...)
}

// Logs message with [erro] header
func Error(e error) {
	fmt.Fprintf(os.Stderr, getHeader(ERR))
	printError(e)
}

// Logs message with [fatl] header
// Fatal is equivalent to logger.Error followed by a call to os.Exit(1).
func Fatal(e error) {
	fmt.Fprintf(os.Stderr, getHeader(FATAL))
	printError(e)
	os.Exit(1)
}

func printError(e error) {
	fmt.Fprintf(os.Stderr, "%s%+v\n", e)
	if os.Getenv("DEBUG_SATORI_SDK") == "true" || DEBUG_SATORI_SDK {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, false)
		fmt.Fprintf(os.Stderr, "%s\n", buf)
	}
}

// Logs message with [debg] header
// Will be displayed only when the ENABLE_DEBUG flag is true
func Debug(v ...interface{}) {
	if os.Getenv("DEBUG_SATORI_SDK") == "true" || DEBUG_SATORI_SDK {
		fmt.Print(getHeader(DEBUG))
		fmt.Println(v...)
	}
}

// Gets message header by type
func getHeader(prefix string) string {
	t := time.Now()
	return prefix + t.Format("2006/01/02 15:04:05.0000 ")
}
