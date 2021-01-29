package logger

import (
	"log"
	"os"
	"path/filepath"
	"scp_delegator/constant"
	"scp_delegator/system"
	"sync"
)

type LoggerWrapper struct {
	instance *log.Logger
	file     *os.File
	isOpen   bool
}

var Wrapper LoggerWrapper

func (l *LoggerWrapper) getInstance() *log.Logger {
	if l.instance != nil {
		return l.instance
	}

	l.Open()
	return l.instance
}

func (l *LoggerWrapper) Open(){
	var once sync.Once
	once.Do(func() {
		logPath := GetOutputPath()
		var err error = nil
		l.file, err = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0744)
		if err != nil {
			log.Fatal("Can't open log file" + logPath)
		}
		l.instance = log.New(l.file, "", log.LstdFlags|log.Llongfile|log.Lmsgprefix)

	})
}

func (l *LoggerWrapper) Close() {
	err := l.file.Close()
	if err != nil {
		log.Fatal("Can't close log ile" + GetOutputPath())
	}
	l.instance = nil
}

func GetOutputDir() string {
	return filepath.Dir(GetOutputPath())
}

func GetOutputPath() string {
	p, err := os.Getwd()
	if err != nil {
		log.Printf("Can't get current path for compose log output path, error=%s", err)
		return getXBCLogPath()
	}

	p = filepath.Join(p, constant.OutputDirectory)
	err = os.MkdirAll(p, 0744)
	if err != nil {
		log.Printf("Can't get create directory for log output path, error=%s", err)
		return getXBCLogPath()
	}

	return filepath.Join(p, constant.LogFileName)
}

func getXBCLogPath() string {
	orArch := system.GetOSandArch()
	switch orArch {
	case "windows/386":
		return filepath.Join(constant.WindowsXBCLogDir32, constant.LogFileName)
	case "windows/amd64":
		return filepath.Join(constant.WindowsXBCLogDir64, constant.LogFileName)
	case "linux/386":
		return filepath.Join(constant.LinuxXBCLogDir32, constant.LogFileName)
	case "linux/amd64":
		return filepath.Join(constant.LinuxXBCLogDir64, constant.LogFileName)
	}
	return ""
}

func (l *LoggerWrapper) LogFatal(fmt string, arg ...interface{}) {
	l.getInstance().SetPrefix("[Fatal] ")
	l.getInstance().Printf(fmt, arg...)
}
func (l *LoggerWrapper) LogError(fmt string, arg ...interface{}) {
	l.getInstance().SetPrefix("[Error] ")
	l.getInstance().Printf(fmt, arg...)
}
func (l *LoggerWrapper) LogInfo(fmt string, arg ...interface{}) {
	l.getInstance().SetPrefix("[Info] ")
	l.getInstance().Printf(fmt, arg...)
}
func (l *LoggerWrapper) LogDebug(fmt string, arg ...interface{}) {
	l.getInstance().SetPrefix("[Debug] ")
	l.getInstance().Printf(fmt, arg...)
}
func (l *LoggerWrapper) LogTrace(fmt string, arg ...interface{}) {
	l.getInstance().SetPrefix("[Trace] ")
	l.getInstance().Printf(fmt, arg...)
}
