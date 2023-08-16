package main

import (
    "net/http"

    "github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
    // 初始化一个新的 httprouter 实例
    router := httprouter.New()

    // 因为 notFoundResponse 和 methodNotAllowedResponse 的签名符合 http.Handler 接口，
    // 所以我们可以将它们直接传递给 NotFound 和 MethodNotAllowed 字段
    router.NotFound = http.HandlerFunc(app.notFoundResponse)
    router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

    // http.Method* 都是字符串常量，分别对应标准的 HTTP 方法
    router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
    router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
    router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

    return router
}