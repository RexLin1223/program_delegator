package logger

import (
	"log"
	"os"
	"sync"
)

var instance *log.Logger
var once sync.Once

func getInstance() *log.Logger {
	if instance != nil {
		return instance
	}

	once.Do(func() {
		log_path := "ds_scp.log"
		f, err := os.OpenFile(log_path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal("Can't creae file" + log_path)
		}
		instance = log.New(f, "", log.LstdFlags|log.Llongfile|log.Lmsgprefix)
	})

	return instance
}

func LogFatal(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Fatal] ")
	getInstance().Printf(fmt, arg...)
	println("Fatal Instance %p", getInstance())
}
func LogError(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Error] ")
	getInstance().Printf(fmt, arg...)
	println("Error Instance %p", getInstance())
}
func LogInfo(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Info] ")
	getInstance().Printf(fmt, arg...)
}
func LogDebug(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Debug] ")
	getInstance().Printf(fmt, arg...)
}
func LogTrace(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Trace] ")
	getInstance().Printf(fmt, arg...)
}
