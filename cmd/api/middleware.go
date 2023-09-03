package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Alphasxd/greenlight/internal/data"
	"github.com/Alphasxd/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

// recoverPanic 是一个中间件，用来恢复 panic，并向客户端发送 500 Internal Server Error 响应。
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

// rateLimit 是一个中间件，用来实现基于令牌桶的请求速率限制。
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
		// 只有当速率限制器是启用的时候，才会执行速率限制
		if app.config.limiter.enabled {
			// 使用 SplitHostPort() 函数从请求的远程地址中提取host部分，赋值给ip变量
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			// 为 IP 地址对应的客户端创建一个新的速率限制器，存储到 clients map 中
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
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
		}

		next.ServeHTTP(w, r)
	})
}

// authenticate 是一个中间件，用来验证用户是否已经登录。
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser 是一个中间件，用来验证用户是否已经登录。
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		// 如果用户是匿名的，说明用户未登录，调用 authenticationRequiredResponse() 方法向客户端发送 401 Unauthorized 响应
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// requireActivatedUser 是一个中间件，用来验证已登录的用户是否已经激活。
func (app *application) requireActivateUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		// 如果用户未激活，调用 inactiveAccountResponse() 方法向客户端发送 403 Forbidden 响应
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
	// 将 requireAuthenticatedUser() 中间件包装在 requireActivatedUser() 中间件外面，这样就可以确保用户已经登录，然后再检查用户是否已经激活
	return app.requireAuthenticatedUser(fn)
}
