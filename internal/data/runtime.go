package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Runtime 声明一个自定义类型，类型为int32
type Runtime int32

// ErrInvalidRuntimeFormat 定义一个错误，当 UnmarshalJSON 方法无法解码 JSON 数据时，会返回该错误
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// MarshalJSON 实现对自定义类型 Runtime 的 MarshalJSON 方法
func (r Runtime) MarshalJSON() ([]byte, error) {

	// 自定义打印格式
	jsonValue := fmt.Sprintf("%d mins", r)

	// 使用 strconv.Quote 函数对 jsonValue 进行转义
	quotedJSONValue := strconv.Quote(jsonValue)

	// 将转义后的字符串转换成 []byte 类型
	return []byte(quotedJSONValue), nil

}

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {

	// 移除两侧的双引号
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 将字符串(<runtime> mins)以空格分割，解析出数字部分
	parts := strings.Split(unquotedJSONValue, " ")

	// 判断分割后的字符串是否为两部分并且第二部分是否为 mins
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// 将 parts[0] 也就是数字部分解析为 10 进制的 int32 类型
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// 将 int32 转换为 Runtime 类型后赋值给 r
	*r = Runtime(i)

	return nil
}
