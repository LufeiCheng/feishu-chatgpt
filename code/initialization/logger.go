package initialization

import (
	"io"
	"log"
	"os"
	"start-feishubot/utils/logger"
)

var Logger *log.Logger

// InitLogger initialize the logger
func InitLogger(config Config) {
	if config.HttpLoggerEnable {
		httpLogger := logger.NewHttpLogger(config.HttpLoggerUrl, config.HttpLoggerMethod, config.HttpLoggerInterval, config.HttpLoggerThreshold)
		Logger = log.New(io.MultiWriter(os.Stdout, httpLogger), "HTTP_LOGGER", log.LstdFlags)
	} else {
		Logger = &log.Logger{}
	}
}
