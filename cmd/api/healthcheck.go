package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// 创建 map 用于存储当前应用的状态
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	// 将 map 转换为 JSON
	// func json.Marshal(v any) ([]byte, error)
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
		return
	}

	// 添加换行，便于在终端中查看
	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")

	// 将包含 JSON 数据的字节切片写入响应体 
	w.Write(js)
}
