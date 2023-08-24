package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// 声明一个互斥锁和一个map，用来存储IP地址和相应的速率限制器
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// 启动一个后台goroutine，定期清理map中过期的速率限制器
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				// 如果速率限制器距离上一次使用超过了 3 分钟，则将其从 map 中删除
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 使用 SplitHostPort() 函数从请求的远程地址中提取host部分，赋值给ip变量
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()

		// 为 IP 地址对应的客户端创建一个新的速率限制器，存储到 clients map 中
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}

		// 更新客户端的最后活跃时间
		clients[ip].lastSeen = time.Now()

		// 检查与当前请求关联的速率限制器是否允许这个请求，如果不允许，则返回一个带有 429 状态码的响应
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
