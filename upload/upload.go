package upload

import (
	"fmt"
	"io/ioutil"
	"jiang/logout"
	"net/http"
	"os"
	"strings"
	"time"
)

type Upload struct {
	R    *http.Request
	Log  *logout.Logout
	Path string
}

func (u *Upload) Up() string {
	file, header, err := u.R.FormFile("file")
	if err != nil {
		//u.Log.Out("r.FormFile is err", err, w)
		return ""
	}
	//u.Log.Out("r.FormFile is ok", w)
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		//u.Log.Out("ioutil.ReadAll is err", err, &file)
		return ""
	}
	//u.Log.Out("ioutil.ReadAll is ok", &file)

	name := u.create(header.Filename)
	b := u.write(name, &bytes)
	if !b {
		//u.Log.Out("file write is err", name)
		return ""
	}
	// u.Log.Out("file write is ok", name)
	return name
}

//生成文件名称
func (u *Upload) create(file string) string {
	t := time.Now()
	folder := fmt.Sprintf("%v%v%v", t.Year(), int(t.Month()), t.Day())

	//判断目录是否存在
	p := u.Path + folder
	if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
		os.MkdirAll(p, os.ModePerm)
	}

	n := fmt.Sprintf("%v%v%v%v", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
	s := strings.Split(file, ".")
	name := fmt.Sprintf("%s/%s.%s", p, n, s[len(s)-1])
	return name
}

//写入上传文件
func (u *Upload) write(name string, b *[]byte) bool {
	fi, err := os.Create(name)
	if err != nil {
		return false
	}
	_, err = fi.Write(*b)
	return err == nil
}
