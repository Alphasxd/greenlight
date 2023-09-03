package main

import (
	"fmt"
	"net/http"
)

// 记录错误日志
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

// 向客户端发送 JSON 格式的错误信息和给定的状态码
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 向客户端发送 500 错误响应和 JSON 格式 Response, 服务器内部错误
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	msg := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

// 向客户端发送 404 错误响应和 JSON 格式 Response, 资源未找到
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

// 向客户端发送 405 错误响应和 JSON 格式 Response, 方法不允许
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}

// 向客户端发送 400 错误响应和 JSON 格式 Response, 请求无效
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// 向客户端发送 422 错误响应和 JSON 格式 Response, 校验失败
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// 向客户端发送 409 错误响应和 JSON 格式 Response, 编辑冲突
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, msg)
}

// 向客户端发送 429 错误响应和 JSON 格式 Response，超出速率限制
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	msg := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, msg)
}

// 向客户端发送 401 错误响应和 JSON 格式 Response, 无效的凭证
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// 向客户端发送 401 错误响应和 JSON 格式 Response, 无效的 Token
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	msg := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// 向客户端发送 401 错误响应和 JSON 格式 Response, 用户未通过身份验证，需要登录
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// 向客户端发送 403 错误响应和 JSON 格式 Response, 用户未激活
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	msg := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}

// 向客户端发送 403 错误响应和 JSON 格式 Response, 用户无权限
func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}
