package dbx

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/soyacen/gox/conc/lazyload"
)

func NewDBs(ctx context.Context, config *Config) *lazyload.Group[*sql.DB] {
	return &lazyload.Group[*sql.DB]{
		New: func(key string) (*sql.DB, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("database %s not found", key)
			}
			return NewDB(ctx, options)
		},
		Finalize: func(ctx context.Context, db *sql.DB) error {
			return db.Close()
		},
	}
}

func NewDB(ctx context.Context, options *Options) (*sql.DB, error) {
	db, err := sql.Open(options.GetDriverName().GetValue(), options.GetDsn().GetValue())
	if err != nil {
		return nil, err
	}
	if options.GetPingTimeout() != nil {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(ctx, options.GetPingTimeout().AsDuration())
		defer cancelFunc()
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if options.GetMaxIdleConns() != nil {
		db.SetMaxIdleConns(int(options.GetMaxIdleConns().GetValue()))
	}
	if options.GetMaxOpenConns() != nil {
		db.SetMaxOpenConns(int(options.GetMaxOpenConns().GetValue()))
	}
	if options.GetConnMaxLifetime() != nil {
		db.SetConnMaxLifetime(options.GetConnMaxLifetime().AsDuration())
	}
	if options.GetConnMaxIdleTime() != nil {
		db.SetConnMaxIdleTime(options.GetConnMaxIdleTime().AsDuration())
	}
	return db, nil
}

func NewSqlxDBs(ctx context.Context, config *Config) *lazyload.Group[*sqlx.DB] {
	return &lazyload.Group[*sqlx.DB]{
		New: func(key string) (*sqlx.DB, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("database %s not found", key)
			}
			return NewSqlxDB(ctx, options)
		},
		Finalize: func(ctx context.Context, db *sqlx.DB) error {
			return db.Close()
		},
	}
}

func NewSqlxDB(ctx context.Context, options *Options) (*sqlx.DB, error) {
	db, err := NewDB(ctx, options)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, options.GetDriverName().GetValue()), nil
}

