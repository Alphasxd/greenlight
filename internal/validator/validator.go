package validator

import "regexp"

// 声明一个检查电子邮件地址格式的正则表达式
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	// 用于存储验证错误的映射
	Errors map[string]string
}

// New() 函数返回一个新的 Validator 实例，其中包含一个空 Validator.Errors 映射
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid() 如果 Validator.Errors 映射中没有任何错误，则 Valid() 方法返回 true，否则返回 false
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError() 将指定的错误消息添加到 Validator.Errors 映射中
func (v *Validator) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

// Check() 当 ok 参数为 false 时，调用 AddError() 方法
func (v *Validator) Check(ok bool, key, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

// In() 检查 value 是否包含在 list 中
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Matches() 检查 value 是否与正则表达式 rx 匹配
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique() 检查 values 切片中的字符串(value)是否唯一
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
