package session

/**
* Session类，Session集合类
* 2016.09.30， 添加同一Session多站点应用的支持
 */

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SessionData 保存Seeeion数据
type SessionData map[string]interface{}

// Session 网页Session类
type Session struct {
	ProductInstance string                 //从当前URL获取product, instance
	ID              string                 //当前Session id
	Values          map[string]SessionData //Session 数据集
	Time            time.Time              //最新时间
	Element         *list.Element          // 在集合链的元素
	sync            *sync.RWMutex          //多线程操作锁
}

// SessionSet 页面Session类集合
type SessionSet struct {
	Values   map[string]*Session // Session数据集合
	List     *list.List          // Session集合的有效顺序链，按有效时效最早到最晚
	sync     *sync.RWMutex       // 多线程操作锁int
	Size     int64               // 用户session占用内容的大小
	Duration int                 // 用户session的内存寿命
	Time     time.Time           // 最近更新时间
	index    int64               // 当前Session建立数
	count    int64               //用户session实现的大小
}

//启用一个全局变量Session集合
func New(s int64, d int) *SessionSet {
	ss := &SessionSet{
		Values:   make(map[string]*Session),
		List:     new(list.List),
		Size:     s,
		Duration: d,
		index:    0,
		sync:     new(sync.RWMutex),
		Time:     time.Now(),
	}
	return ss
}

// Start 启用Session
func (ss *SessionSet) Start(w http.ResponseWriter, r *http.Request) *Session {
	name := "JCMS"
	key := ""
	//从请求获取cookie key
	cookie, err := r.Cookie(name)
	if err == nil {
		key = cookie.Value
	} else {
		//新生成cookie key
		key = ss.ID()
		c := &http.Cookie{
			Name:     name,
			Value:    key,
			Path:     "/",
			MaxAge:   ss.Duration,
			HttpOnly: true,
		}
		http.SetCookie(w, c)
	}
	s := ss.Get(key)

	// 设置定前机构信息
	a := strings.Split(r.URL.Path, "/")
	arr := []string{"web", "bookan"}
	if len(a) > 1 {
		arr[0] = a[1]
	}
	if len(a) > 2 {
		arr[1] = a[2]
	}
	s.ProductInstance = fmt.Sprintf("%s,%s", arr[0], arr[1])
	return s
}

// Get 从SessionSet集合中获取session
func (ss *SessionSet) Get(key string) *Session {
	//已有
	if ss.Values == nil {
		return ss.Set(key)
	}
	ss.sync.RLock()
	s, ok := ss.Values[key]
	ss.sync.RUnlock()
	if ok {
		s.Time = time.Now()
		ss.List.MoveToBack(s.Element)
		return s
	}
	//没有全新设置
	return ss.Set(key)
}

// Set 向SessionSet集合添加新的Session
func (ss *SessionSet) Set(key string) *Session {
	if ss.Values == nil {
		ss.Values = make(map[string]*Session)
	}
	e := ss.List.PushBack(key)
	s := &Session{"", key, make(map[string]SessionData), time.Now(), e, new(sync.RWMutex)}
	if ss.Values == nil {
		ss.Values = make(map[string]*Session)
	}
	ss.sync.Lock()
	ss.Values[key] = s
	ss.sync.Unlock()
	ss.index += int64(1)
	return s
}

// ID 生成SessionSet全局id
func (ss *SessionSet) ID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		t := time.Now()
		nano := t.UnixNano()
		s := fmt.Sprintf("%d%v", nano, ss.index)
		b = []byte(s)
	}
	ss.index++
	return base64.URLEncoding.EncodeToString(b)
}

// Update 更新集合中的有效性，有效更新
func (ss *SessionSet) Update() {
	if !ss.Time.IsZero() && int(time.Now().Sub(ss.Time)) < ss.Duration {
		return
	}

	var n = new(list.Element)
	for e := ss.List.Front(); e != nil; e = n {
		n = e.Next()
		k, ok := e.Value.(string)
		if !ok || k == "" {
			ss.List.Remove(e)
		}
		val, ok := ss.Values[k]
		if !ok {
			ss.List.Remove(e)
			continue
		}
		if !val.Time.IsZero() && int(time.Now().Sub(val.Time).Seconds()) < ss.Duration {
			return
		}
		ss.sync.Lock()
		delete(ss.Values, k)
		ss.sync.Unlock()
		ss.List.Remove(e)
	}
}

// UpdateAll 更新集合中的有效性，全部更新
func (ss *SessionSet) UpdateAll() {
	list := new(list.List)
	for key, val := range ss.Values {
		if !val.Time.IsZero() && int(time.Now().Sub(val.Time).Seconds()) < ss.Duration {
			list.PushBack(val.Element)
			continue
		}
		ss.sync.Lock()
		delete(ss.Values, key)
		ss.sync.Unlock()
	}
	ss.List = list
}

// Get 获取保存在Session的内容
func (s *Session) Get(key string) interface{} {
	if s.Values == nil {
		return nil
	}
	// log.Println("0", s.Values)
	s.sync.RLock()
	data, ok := s.Values[s.ProductInstance]
	s.sync.RUnlock()
	if !ok {
		return nil
	}
	if v, ok := data[key]; ok {
		return v
	}
	return nil
}

// Set 保存内容在Session
func (s *Session) Set(key string, value interface{}) {
	if s.Values == nil {
		s.Values = make(map[string]SessionData)
	}
	data, ok := s.Values[s.ProductInstance]
	if !ok {
		data = make(SessionData)
		data[key] = value
		s.Values[s.ProductInstance] = data
		return
	}
	s.sync.Lock()
	data[key] = value
	s.sync.Unlock()
	s.Values[s.ProductInstance] = data
	// log.Println("1", s.Values)
}
