package vpn

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron"
	"google.golang.org/grpc/credentials"
	"hub.ahiho.com/ahiho/squirrel-srv/internal/vpn/protocol/grpc"
	"hub.ahiho.com/ahiho/squirrel-srv/internal/vpn/protocol/restful"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/api/v1"
	"hub.ahiho.com/ahiho/squirrel-srv/pkg/logger"
	"os"
	"strconv"
)

var (

	kEnvPrivateKey = "PRIVATE_KEY"
	kEnvPublicKey  = "PUBLIC_KEY"

	kEnvGRPCPort = "GRPC_PORT"
	kEnvHTTPPort = "HTTP_PORT"

	kEnvDBDriver   = "DB_DRIVER"
	kEnvDBHost     = "DB_HOST"
	kEnvDBUser     = "DB_USER"
	kEnvDBPassword = "DB_PASSWORD"
	kEnvDBSchema   = "DB_SCHEMA"

	kEnvLogLevel      = "LOG_LEVEL"
	kEnvLogTimeFormat = "LOG_TIME_FORMAT"
)

// Config is configuration for Server
type Config struct {
	// TLS options
	TLS            bool
	TLSCertificate string
	TLSKey         string

	PrivateKey string
	PublicKey  string

	// gRPC server start parameters section
	// gRPC is TCP port to listen by gRPC server
	GRPCPort string

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string

	// DB type
	DBDriver string
	// DB parameters section
	// DBHost is host of database
	DBHost string
	// DBUser is username to connect to database
	DBUser string
	// DBPassword password to connect to database
	DBPassword string
	// DBSchema is schema of database
	DBSchema string

	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int
	// LogTimeFormat is print time format for logger e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	logLevelEnv, _ := strconv.Atoi(os.Getenv(kEnvLogLevel))

	// get configuration
	var cfg Config
	flag.BoolVar(&cfg.TLS, "tls", false, "gRPC TLS or plain TCP")
	flag.StringVar(&cfg.TLSCertificate, "tls-cert", "", "TLS certificate file")
	flag.StringVar(&cfg.TLSKey, "tls-key", "", "TLS key file")
	flag.StringVar(&cfg.PrivateKey, "private-key", os.Getenv(kEnvPrivateKey), "Private key value")
	flag.StringVar(&cfg.PublicKey, "public-key", os.Getenv(kEnvPublicKey), "Public key value")
	flag.StringVar(&cfg.GRPCPort, "grpc-port", os.Getenv(kEnvGRPCPort), "gRPC port to bind")
	flag.StringVar(&cfg.HTTPPort, "http-port", os.Getenv(kEnvHTTPPort), "HTTP port to bind")
	flag.StringVar(&cfg.DBDriver, "db-driver", os.Getenv(kEnvDBDriver), "Database driver")
	flag.StringVar(&cfg.DBHost, "db-host", os.Getenv(kEnvDBHost), "Database host")
	flag.StringVar(&cfg.DBUser, "db-user", os.Getenv(kEnvDBUser), "Database user")
	flag.StringVar(&cfg.DBPassword, "db-password", os.Getenv(kEnvDBPassword), "Database password")
	flag.StringVar(&cfg.DBSchema, "db-schema", os.Getenv(kEnvDBSchema), "Database schema")
	flag.IntVar(&cfg.LogLevel, "log-level", logLevelEnv, "Global log level")
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", os.Getenv(kEnvLogTimeFormat),
		"Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	if len(cfg.HTTPPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", cfg.HTTPPort)
	}

	// initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	// add MySQL driver specific parameter to parse date/time
	// Drop it for another database
	param := "parseTime=true&multiStatements=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBSchema,
		param)
	if cfg.DBDriver == "sqlite3" {
		dsn = cfg.DBSchema
	}

	db, err := sqlx.Connect(cfg.DBDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	dbMigrate, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect migrate database: %v", err)
	}

	driver, err := mysql.WithInstance(dbMigrate, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver migrate database: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:migrations",
		"mysql", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate database instance: %v", err)
	}
	err = m.Steps(1)
	if err != nil {
		logger.Log.Warn("migrate database err ->"+err.Error())
	}

	_ = dbMigrate.Close()


	// TLS
	var creds credentials.TransportCredentials
	if cfg.TLS == true {
		if cfg.TLSCertificate == "" {
			cfg.TLSCertificate = "../../certs/certificate.pem"
		}
		if cfg.TLSKey == "" {
			cfg.TLSKey = "../../certs/key.pem"
		}
		creds, err = credentials.NewServerTLSFromFile(cfg.TLSCertificate, cfg.TLSKey)
		if err != nil {
			return fmt.Errorf("failed to generate credentials %v", err)
		}
	}

	v1API := NewServiceServer(db)

	c := cron.New()
	defer c.Stop()
	_ = c.AddFunc("@every 1m", func() {
		crawled, err := v1API.VPNGateCrawler(ctx, &v1.VPNGateCrawlerRequest{Api: apiVersion})
		if err != nil {
			logger.Log.Warn("crawl error: "+err.Error())
		}
		logger.Log.Info("crawled success "+strconv.Itoa(len(crawled.Data))+ " items")
	})
	c.Start()

	// run HTTP gateway
	go func() {
		_ = restful.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort, creds)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort, creds)
}
