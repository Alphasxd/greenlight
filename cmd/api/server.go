package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动一个goroutine来监听操作系统的中断信号
	go func() {
		quit := make(chan os.Signal, 1)

		// 监听 syscall.SIGINT 和 syscall.SIGTERM 信号
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// 接收到信号后，执行关闭服务器的操作, 在收到信号之前会一直阻塞
		s := <-quit

		// 记录信号的名称，使用 String() 方法将信号转换为字符串
		app.logger.PrintInfo("caught signal", map[string]string{
			"signal": s.String(),
		})

		os.Exit(0)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	return srv.ListenAndServe()
}
