package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Alphasxd/greenlight/internal/data"
	"github.com/Alphasxd/greenlight/internal/jsonlog"
	"github.com/Alphasxd/greenlight/internal/mailer"

	_ "github.com/lib/pq"
)

// 定义版本号
const version = "1.0.0"

// 定义构建时间
var buildTime string

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
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// 应用结构体，用于存储应用程序的依赖项，handler，helper，middleware，logger等
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	// 声明一个config实例
	var cfg config

	// 在命令行中设置配置
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "d94086b9b52487", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "db56cefb32b838", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@github/Alphasxd>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// 定义一个命令行参数，用于显示版本号
	displayVersion := flag.Bool("version", false, "Display version and exit")

	// 解析命令行参数
	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

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

	// 在 expvar 中注册一个名为 version 的字符串变量，用于存储应用程序的版本号
	expvar.NewString("version").Set(version)

	// 在 expvar 中注册一个名为 goroutines 的变量，用于存储当前 goroutine 的数量
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// 在 expvar 中注册一个名为 database 的变量，用于存储数据库连接池的统计信息
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// 在 expvar 中注册一个名为 timestamp 的变量，用于存储当前时间戳
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	// 初始化一个application实例
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	// 调用serve方法启动服务器
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
