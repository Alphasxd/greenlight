package data

import (
	"fmt"
	"strconv"
)

// 声明一个自定义类型，类型为int32
type Runtime int32

// 实现对自定义类型 Runtime 的 MarshalJSON 方法
func (r Runtime) MarshalJSON() ([]byte, error) {
	
	// 自定义打印格式
	jsonValue := fmt.Sprintf("%d mins", r)

	// 使用 strconv.Quote 函数对 jsonValue 进行转义
	quotedJSONValue := strconv.Quote(jsonValue)

	// 将转义后的字符串转换成 []byte 类型
	return []byte(quotedJSONValue), nil

}