package orm

import "github.com/bysir-zl/bygo/log"

var logger *log.Logger

func warn(a ...interface{}) {
	if Debug {
		logger.Warn("ORM", a...)
	}
}

func info(a ...interface{}) {
	if Debug {
		logger.Info("ORM", a...)
	}
}

func init() {
	logger = log.NewLogger()
	logger.SetCallDepth(4)
}
