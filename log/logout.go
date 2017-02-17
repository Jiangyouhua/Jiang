package logout

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

type Logout struct {
	Path    string
	File    string
	Class   int //按哪种时间类型输出：年、月、日天
	current string
	logger  *log.Logger
	logfile *os.File
}

//写入前初始工作
func (e *Logout) Start() {

	//创建目录
	err := os.MkdirAll(e.Path, os.ModePerm)
	if err != nil {
		fmt.Println(0, "os.MkdirAll is err", err, e.Path)
		os.Exit(1)
	}
	fmt.Println("os.MkdirAll is ok", e.Path)

	//创建文件
	f := fmt.Sprintf("%s//%s_%s.log", e.Path, e.File, e.current)
	e.logfile, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(0, "initLogger os.OpenFile is err", err)
		os.Exit(1)
	}

	//日志绑定文件
	e.logger = log.New(e.logfile, "\r\n", log.Ldate|log.Ltime)
	e.Out("initLogger is OK")
}

//输出异常
func (e *Logout) Out(a ...interface{}) {
	t := time.Now()
	s := "2006"
	if e.Class == 1 {
		s = "2006_01"
	}
	if e.Class == 2 {
		s = "2006_01_02"
	}
	if e.Class == 3 {
		s = "2006_01_02.15"
	}
	if e.Class == 4 {
		s = "2006_01_02.15.04"
	}

	s = t.Format(s)
	//判断日志文件是否准备好
	if e.logger == nil || e.current != s {
		e.current = s
		e.Start()
	}

	//获取调用者、文件名、行信息
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	m := fmt.Sprintf("%s %v %s", file, line, f.Name())
	e.logger.Println(m, a)
}
