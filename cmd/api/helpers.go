package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// 定义一个名为 envelope 的 map 类型，它的键是 string 类型，值是任意类型
type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("Invalid id parameter")
	}

	return id, nil
}

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
	w.Write(js)

	return nil
}

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
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
