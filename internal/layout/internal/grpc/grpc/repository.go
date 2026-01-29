package grpc

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/soyacen/gox/conc/lazyload"
)

type Repository struct {
	db *sqlx.DB
	rd redis.UniversalClient
}

func NewRepository(dbs *lazyload.Group[*sqlx.DB], rds *lazyload.Group[redis.UniversalClient]) (*Repository, error) {
	db, err, _ := dbs.Load("grpc")
	if err != nil {
		return nil, err
	}
	rd, err, _ := rds.Load("grpc")
	if err != nil {
		return nil, err
	}
	return &Repository{db: db, rd: rd}, nil
}
