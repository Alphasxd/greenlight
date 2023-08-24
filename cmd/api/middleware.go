package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// 创建一个速率限制器，设置每秒允许的请求数量为 2，设置桶的容量为 4
	limiter := rate.NewLimiter(2, 4)
	// 返回一个闭包函数，这个函数包装了 limiter 对象的 Allow 方法的调用
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果请求超过了速率限制器的限制，则向客户端发送一个带有 429 状态码的响应
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
