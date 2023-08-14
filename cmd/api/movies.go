package main

import (
    "fmt"
    "net/http"
    "strconv" 

    "github.com/julienschmidt/httprouter" 
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
    // 使用 httprouter.ParamsFromContext() 函数来获取 request context 中的参数
    params := httprouter.ParamsFromContext(r.Context())

    // 将参数转换为 10 进制的 int64 类型
    id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

	// 将 id 插入到响应中
    fmt.Fprintf(w, "show the details of movie %d\n", id)
}