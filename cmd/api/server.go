package main

import (
	"context"
	"errors"
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

	// shutdownError 通道用来接收服务器关闭时返回的错误
	shutdownError := make(chan error)

	// 启动一个goroutine来监听操作系统的中断信号
	go func() {
		quit := make(chan os.Signal, 1)

		// 监听 syscall.SIGINT 和 syscall.SIGTERM 信号
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// 接收到信号后，执行关闭服务器的操作, 在收到信号之前会一直阻塞
		s := <-quit

		// 记录信号的名称，使用 String() 方法将信号转换为字符串
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// 创建一个5秒超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 调用 Shutdown() 方法来关闭服务器，传入我们自己创建的上下文
		// 如果成功，Shutdown() 方法会返回 nil，如果失败，说明没有在 5 秒内完成所有活动的连接，
		shutdownError <- srv.Shutdown(ctx)
	}()

	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// 检查 err 是否为 http.ErrServerClosed，如果不是，则返回 err
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// 等待接受 shutdownError 通道返回的错误
	err = <-shutdownError
	if err != nil {
		return err
	}

	// 服务器已经优雅的关闭，记录日志并退出
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
