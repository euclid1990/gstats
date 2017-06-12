package utilities

import (
	"github.com/euclid1990/gstats/configs"
	"io"
	"log"
	"os"
	"time"
)

var (
	Info     *log.Logger
	Warning  *log.Logger
	Error    *log.Logger
	Critical *log.Logger
	Debug    *log.Logger
	Notice   *log.Logger
)

func InitLogger() {
	now := time.Now().Format(configs.FILE_LOG_FORMAT_DATE)

	logfile, err := os.OpenFile(configs.LOG_DIR+now+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("[Log] Open file log failed!")
	}

	Info = log.New(io.MultiWriter(logfile, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(io.MultiWriter(logfile, os.Stdout), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(logfile, os.Stdout), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	Critical = log.New(io.MultiWriter(logfile, os.Stdout), "Critical: ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(io.MultiWriter(logfile, os.Stdout), "Debug: ", log.Ldate|log.Ltime|log.Lshortfile)
	Notice = log.New(io.MultiWriter(logfile, os.Stdout), "Notice: ", log.Ldate|log.Ltime|log.Lshortfile)
}
