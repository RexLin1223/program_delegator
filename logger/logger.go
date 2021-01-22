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

var LogFatal = func(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Fatal] ")
	getInstance().Printf(fmt, arg...)
	println("Fatal Instance %p", getInstance())
}
var LogError = func(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Error] ")
	getInstance().Printf(fmt, arg...)
	println("Error Instance %p", getInstance())
}
var LogInfo = func(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Info] ")
	getInstance().Printf(fmt, arg...)
}
var LogDebug = func(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Debug] ")
	getInstance().Printf(fmt, arg...)
}
var LogTrace = func(fmt string, arg ...interface{}) {
	getInstance().SetPrefix("[Trace] ")
	getInstance().Printf(fmt, arg...)
}

