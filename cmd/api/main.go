package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// 定义版本号
const version = "1.0.0"

// 定义配置结构体
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
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
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// 解析命令行参数
	flag.Parse()

	// 初始化一个logger实例
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// 延迟关闭数据库连接
	defer db.Close()

	logger.Printf("database connection pool established")

	// 初始化一个application实例
	app := &application{
		config: cfg,
		logger: logger,
	}

	// 定义一个http.Server实例
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动 http server
	logger.Printf("Starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	// 使用 dsn 字符串创建一个数据库连接池
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// 创建一个上下文对象，设置超时时间为5秒
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 使用上下文对象的 PingContext() 方法检测数据库连接是否正常
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	// 返回数据库连接池
	return db, nil
}
