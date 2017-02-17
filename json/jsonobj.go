package jiang

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// JSONObject json对象处理结构
type JSONObject struct {
	JSON interface{}
}

// JSONFromFile 从文件获取json
func JSONFromFile(f string) (*JSONObject, error) {
	if f == "" {
		return nil, errors.New("JSONFromFile's f is nil")
	}
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return JSONFromByte(b)
}

// JSONFromByte 将数据解析为json
func JSONFromByte(b []byte) (*JSONObject, error) {
	if b == nil {
		return nil, errors.New("JSONFromFile's b is nil")
	}
	var data interface{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}
	return &JSONObject{data}, nil
}

// InstanceByIndex 通过索引从对象内获取数据
func InstanceByIndex(data interface{}, i int) (interface{}, error) {
	if data == nil {
		return nil, errors.New("Root Object is nil")
	}
	if i < 0 {
		return nil, fmt.Errorf("Root Object's index Can not be negative, %v", i)
	}
	a, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Root Object is not array, %s", data)
	}
	if i >= len(a)-1 {
		return nil, fmt.Errorf("Root Object's index is out or range, %s", data)
	}
	return a[i], nil
}

// InstanceBykey 通过key从对象内获取数据
func InstanceBykey(data interface{}, k string) (interface{}, error) {
	if data == nil {
		return nil, errors.New("Root Object is nil")
	}
	if k == "" {
		return nil, errors.New("Root Object's key is nil, ")
	}
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Root Object is not map, %s", data)
	}
	v, ok := m[k]
	if !ok {
		return nil, fmt.Errorf("Root Object is not this key, %s, %s", m, k)
	}
	return v, nil
}

// GetInterface 从json集合获取当前key的值
func (j *JSONObject) GetInterface(args ...interface{}) (interface{}, error) {
	// JSONObject is nil
	if j == nil {
		return nil, errors.New("The JSONObject is nil")
	}

	// 无key
	if args == nil || len(args) == 0 {
		return j, nil
	}

	// 不能解析为数组或字符串
	data := j.JSON
	var err error
	for _, v := range args {
		// 处理array
		if i, ok := v.(int); ok {
			data, err = InstanceByIndex(data, i)
			if err != nil {
				return nil, err
			}
			continue
		}
		if f, ok := v.(float64); ok {
			i := int(f)
			data, err = InstanceByIndex(data, i)
			if err != nil {
				return nil, err
			}
			continue
		}
		// 处理map
		if k, ok := v.(string); ok {
			data, err = InstanceBykey(data, k)
			if err != nil {
				return nil, err
			}
			continue
		}
		if b, ok := v.([]byte); ok {
			k := string(b)
			data, err = InstanceBykey(data, k)
			if err != nil {
				return nil, err
			}
			continue
		}
		return nil, errors.New("Root JSONObject is invalid")
	}
	return data, nil
}

// GetBool 从json集合获取当前key的bool值
func (j *JSONObject) GetBool(args ...interface{}) (bool, error) {
	data, err := j.GetInterface(args...)
	if err != nil {
		return false, err
	}
	if b, ok := data.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("Result is not bool, %s", data)
}

// GetInt 从json集合获取当前key的int值
func (j *JSONObject) GetInt(args ...interface{}) (int, error) {
	data, err := j.GetInterface(args...)
	if err != nil {
		return 0, err
	}
	if i, ok := data.(int); ok {
		return i, nil
	}
	if f, ok := data.(float64); ok {
		i := int(f)
		return i, nil
	}
	return 0, fmt.Errorf("Result is not int, %s", data)
}

// GetString 从json集合获取当前key的string值
func (j *JSONObject) GetString(args ...interface{}) (string, error) {
	data, err := j.GetInterface(args...)
	if err != nil {
		return "", err
	}
	if s, ok := data.(string); ok {
		return s, nil
	}
	if b, ok := data.([]byte); ok {
		s := string(b)
		return s, nil
	}
	return "", fmt.Errorf("Result is not string, %s", data)
}
