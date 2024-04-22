package logger

import (
	"ReniaoCache/conf"
	"github.com/charmbracelet/log"
	"os"
	"time"
)

var Logger *log.Logger

func Init() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "ReniaoCache ",
	})
	if conf.AppMode == "debug" {
		Logger.SetLevel(log.DebugLevel)
	} else {
		Logger.SetLevel(log.InfoLevel)
	}
}
