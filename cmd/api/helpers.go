package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("Invalid id parameter")
	}

	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
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