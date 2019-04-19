package log

import (
	"os"
	"github.com/go-kit/kit/log"
	"time"
	"github.com/go-stack/stack"
	"fmt"
)

// Caller returns a Valuer that returns a file and line from a specified depth
// in the callstack. Users will probably want to use DefaultCaller.
func Caller(depth int) log.Valuer {
	return func() interface{} { return fmt.Sprintf("%+v", stack.Caller(depth)) }
}

var logger = log.With(log.NewJSONLogger(os.Stdout), "ts", log.TimestampFormat(time.Now, "2006-01-02 15:04:05"), "caller", Caller(4))
var elogger = log.With(logger, "leave", "ERROR")

func Error(a ...interface{}) {
	elogger.Log(a...)
}
