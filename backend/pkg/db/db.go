package db

import (
	"database/sql"
	"fmt"
	"go-fiber-starter/pkg/config"
	"go-fiber-starter/pkg/logger"
	"go-fiber-starter/pkg/util"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"moul.io/zapgorm2"
)

var DB *gorm.DB

func Init() error {
	gormLogger := zapgorm2.New(logger.Logger.Desugar())
	gormLogger.IgnoreRecordNotFoundError = true

	if err := ensureDatabaseReady(); err != nil {
		return err
	}

	dialector, err := buildDialector()
	if err != nil {
		return err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:         gormLogger,
		NamingStrategy: schema.NamingStrategy{},
	})

	if err != nil {
		return err
	}
	DB = db
	configureConnectionPool(db)
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	return err
}

func ensureDatabaseReady() error {
	if !isPostgresDriver() {
		return nil
	}
	return ensurePostgresDatabase()
}

func buildDialector() (gorm.Dialector, error) {
	switch currentDriver() {
	case "sqlite":
		path := config.Current.Database.Path
		if path == "" {
			return nil, fmt.Errorf("sqlite 数据库路径不能为空")
		}
		if err := util.EnsureDir(path); err != nil {
			logger.Error("创建数据库目录失败: %w", err)
			return nil, err
		}
		return sqlite.Open(path), nil
	case "postgres", "postgresql":
		dsn := buildPostgresDSN()
		if dsn == "" {
			return nil, fmt.Errorf("postgres 数据库配置不完整")
		}
		return postgres.Open(dsn), nil
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", currentDriver())
	}
}

func ensurePostgresDatabase() error {
	adminDSN, databaseName, err := buildPostgresAdminDSN()
	if err != nil {
		return err
	}

	rawDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("连接 PostgreSQL 管理库失败: %w", err)
	}
	defer rawDB.Close()

	if err := rawDB.Ping(); err != nil {
		return fmt.Errorf("连接 PostgreSQL 管理库失败: %w", err)
	}

	var exists bool
	if err := rawDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", databaseName).Scan(&exists); err != nil {
		return fmt.Errorf("检查 PostgreSQL 数据库是否存在失败: %w", err)
	}
	if exists {
		return nil
	}

	quotedName := pgx.Identifier{databaseName}.Sanitize()
	if _, err := rawDB.Exec("CREATE DATABASE " + quotedName); err != nil {
		return fmt.Errorf("创建 PostgreSQL 数据库失败: %w", err)
	}

	return nil
}

func buildPostgresDSN() string {
	if strings.TrimSpace(config.Current.Database.DSN) != "" {
		return config.Current.Database.DSN
	}

	if config.Current.Database.Host == "" || config.Current.Database.User == "" || config.Current.Database.Name == "" {
		return ""
	}

	parts := []string{
		fmt.Sprintf("host=%s", config.Current.Database.Host),
		fmt.Sprintf("port=%d", maxInt(config.Current.Database.Port, 5432)),
		fmt.Sprintf("user=%s", config.Current.Database.User),
		fmt.Sprintf("password=%s", config.Current.Database.Password),
		fmt.Sprintf("dbname=%s", config.Current.Database.Name),
		fmt.Sprintf("sslmode=%s", resolveValue(config.Current.Database.SSLMode, "disable")),
		fmt.Sprintf("TimeZone=%s", resolveValue(config.Current.Database.TimeZone, "Asia/Shanghai")),
	}
	return strings.Join(parts, " ")
}

func buildPostgresAdminDSN() (string, string, error) {
	dsn := buildPostgresDSN()
	if dsn == "" {
		return "", "", fmt.Errorf("postgres 数据库配置不完整")
	}

	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		return "", "", fmt.Errorf("解析 PostgreSQL DSN 失败: %w", err)
	}

	databaseName := strings.TrimSpace(parsed.Database)
	if databaseName == "" {
		return "", "", fmt.Errorf("postgres 数据库名不能为空")
	}

	parsed.Database = "postgres"
	return stdlibDsn(parsed), databaseName, nil
}

func configureConnectionPool(db *gorm.DB) {
	rawDB, err := db.DB()
	if err != nil {
		return
	}

	applyPoolSetting(rawDB, config.Current.Database.MaxIdleConns, (*sql.DB).SetMaxIdleConns)
	applyPoolSetting(rawDB, config.Current.Database.MaxOpenConns, (*sql.DB).SetMaxOpenConns)
}

func applyPoolSetting(database *sql.DB, value int, setter func(*sql.DB, int)) {
	if value > 0 {
		setter(database, value)
	}
}

func resolveValue(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func currentDriver() string {
	driver := strings.TrimSpace(strings.ToLower(config.Current.Database.Driver))
	if driver == "" {
		return "sqlite"
	}
	return driver
}

func isPostgresDriver() bool {
	driver := currentDriver()
	return driver == "postgres" || driver == "postgresql"
}

func stdlibDsn(cfg *pgx.ConnConfig) string {
	return stdlib.RegisterConnConfig(cfg)
}
