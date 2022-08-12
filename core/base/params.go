package base

import (
	"fmt"
	"strconv"

	"event/core/helper"
)

// Params 参数
type Params map[string]interface{}

// Exists 是否存在
func (p Params) Exists(key string) bool {
	_, exists := p[key]
	return exists
}

// String 获取字符串
func (p Params) String(key string) string {
	value, _ := p[key].(string)
	return value
}

func (p Params) GetString(key string) string {
	value, exists := p[key]
	if !exists {
		return ""
	}

	switch val := value.(type) {
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case int:
		return strconv.Itoa(val)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case int32:
		return strconv.Itoa(int(val))
	case uint32:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.Itoa(int(val))
	case uint16:
		return strconv.Itoa(int(val))
	case int8:
		return strconv.Itoa(int(val))
	case uint8:
		return strconv.Itoa(int(val))
	case bool:
		return strconv.FormatBool(val)
	case []byte:
		return helper.Bytes2String(val)
	case string:
		return val
	default:
		return fmt.Sprint(val)
	}
}

// Int64 获取int64值
func (p Params) Int64(key string) int64 {
	value, _ := p[key].(int64)
	return value
}

// GetInt64 获取int64值，如果不是int64会进行转换为int64，转换失败则返回0
func (p Params) GetInt64(key string) int64 {
	value, exists := p[key]
	if !exists {
		return 0
	}

	switch val := value.(type) {
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case int64:
		return val
	case uint64:
		return int64(val)
	case int:
		return int64(val)
	case uint:
		return int64(val)
	case int32:
		return int64(val)
	case uint32:
		return int64(val)
	case int16:
		return int64(val)
	case uint16:
		return int64(val)
	case int8:
		return int64(val)
	case uint8:
		return int64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case []byte:
		v, _ := strconv.ParseInt(helper.Bytes2String(val), 10, 64)
		return v
	case string:
		v, _ := strconv.ParseInt(val, 10, 64)
		return v
	default:
		return 0
	}
}

// Int 获取int值
func (p Params) Int(key string) int {
	value, _ := p[key].(int)
	return value
}

// GetInt 获取int值，如果不是int会进行转换为int，转换失败则返回0
func (p Params) GetInt(key string) int {
	return int(p.GetInt64(key))
}

// Float64 获取float64值
func (p Params) Float64(key string) float64 {
	value, _ := p[key].(float64)
	return value
}

// GetFloat64 获取float64值，如果不是float64会进行转换为float64，转换失败则返回0
func (p Params) GetFloat64(key string) float64 {
	value, exists := p[key]
	if !exists {
		return 0
	}

	switch val := value.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case uint64:
		return float64(val)
	case int:
		return float64(val)
	case uint:
		return float64(val)
	case int32:
		return float64(val)
	case uint32:
		return float64(val)
	case int16:
		return float64(val)
	case uint16:
		return float64(val)
	case int8:
		return float64(val)
	case uint8:
		return float64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case []byte:
		v, _ := strconv.ParseFloat(helper.Bytes2String(val), 64)
		return v
	case string:
		v, _ := strconv.ParseFloat(val, 64)
		return v
	default:
		return 0
	}
}
