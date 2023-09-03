package main

import (
	"context"
	"net/http"

	"github.com/Alphasxd/greenlight/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

// contextSetUser 将给定的 User 对象添加到请求的上下文中
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser 从请求的上下文中返回 User 对象
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
