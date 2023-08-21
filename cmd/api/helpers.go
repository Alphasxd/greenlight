package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alphasxd/greenlight/internal/validator"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// 定义一个名为 envelope 的 map 类型，它的键是 string 类型，值是任意类型
type envelope map[string]any

// readIDParam() 读取 URL 中的 id 参数
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// readString() 返回 URL 查询字符串参数中读取字符串参数，如果参数不存在，则返回默认值
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV() 从 URL 查询字符串参数中读取 CSV（逗号分隔值）参数，将其解析为字符串切片，并返回，否则返回默认值
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readInt() 从 URL 查询字符串参数中读取整数值，如果参数不存在或者无法解析为整数，则返回默认值
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// writeJSON() 写入 JSON 响应
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// 将 data 封装成 JSON 格式
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 添加换行，便于在终端中查看
	js = append(js, '\n')

	for key, value := range headers {
		// w.Header() 返回一个 map[string][]string 类型的 map
		w.Header()[key] = value
	}

	// 设置响应头的 Content-Type 字段，并将状态码写入响应体和 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

// readJSON() 读取 JSON 请求
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	// 使用 http.MaxBytesReader() 限制读取请求体的大小为 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// 初始化 json.Decoder，在解码之前调用 DisallowUnknownFields() 方法
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// 将请求体中的 JSON 数据解码到目标变量 dst 中
	err := dec.Decode(dst)
	if err != nil {
		// 定义可能出现的错误类型
		var syntaxError *json.SyntaxError                     // 语法错误
		var unmarshalTypeError *json.UnmarshalTypeError       // 类型错误，JSON 与目标的 Go 类型不匹配
		var invalidUnmarshalError *json.InvalidUnmarshalError // 解码目标错误，通常是目标代码的问题

		// 根据错误类型，返回自定义的错误消息
		switch {

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unkown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// 再次调用 Decode() 方法，确保请求体只包含单个 JSON 值
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
