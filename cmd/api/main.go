package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// 定义版本号
const version = "1.0.0"

// 定义配置结构体
type config struct {
	port int
	env  string
}

// 应用结构体，用于存储应用程序的依赖项，handler，helper，middleware，logger等
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// 声明一个config实例
	var cfg config

	// 在命令行中设置配置
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// 初始化一个logger实例
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// 初始化一个application实例
	app := &application{
		config: cfg,
		logger: logger,
	}

	// 定义一个路由器
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// 定义一个http.Server实例
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动 http server
	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
