package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// 创建 json 响应
	js := `{"status": "available", "environment": %q, "version": %q}`
	js = fmt.Sprintf(js, app.config.env, version)

	// 将 json 响应写入到响应体中，并设置 Content-Type 头
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(js))
}
