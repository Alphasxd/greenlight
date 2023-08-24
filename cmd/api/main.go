package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Alphasxd/greenlight/internal/data"
	"github.com/Alphasxd/greenlight/internal/jsonlog"

	_ "github.com/lib/pq"
)

// 定义版本号
const version = "1.0.0"

// 定义配置结构体
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// 应用结构体，用于存储应用程序的依赖项，handler，helper，middleware，logger等
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	// 声明一个config实例
	var cfg config

	// 在命令行中设置配置
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// 解析命令行参数
	flag.Parse()

	// 初始化一个logger实例
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// 延迟关闭数据库连接
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.PrintFatal(err, nil)
		}
	}(db)

	logger.PrintInfo("database connection pool established", nil)

	// 初始化一个application实例
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
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
	logger.PrintInfo("Starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	// 使用 dsn 字符串创建一个数据库连接池
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// 设置数据池的最大 open 连接数（in-use+idle），如果 maxOpenConns <= 0，则表示没有限制
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	// 设置数据池的最大空闲连接数（idle），如果 maxIdleConns <= 0，则表示没有限制
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// 使用 time.ParseDuration() 函数将字符串转换为 time.Duration 类型
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	// 设置空闲连接的最大可空闲时间，超过该时间的空闲连接将会被关闭
	db.SetConnMaxIdleTime(duration)

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
