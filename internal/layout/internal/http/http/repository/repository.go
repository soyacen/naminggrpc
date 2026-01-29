package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/soyacen/gox/conc/lazyload"
)

type Repository interface{}

type repositoryImpl struct {
	db *sqlx.DB
	rd redis.UniversalClient
}

func NewRepository(dbs *lazyload.Group[*sqlx.DB], rds *lazyload.Group[redis.UniversalClient]) (Repository, error) {
	db, err, _ := dbs.Load("http")
	if err != nil {
		return nil, err
	}
	rd, err, _ := rds.Load("http")
	if err != nil {
		return nil, err
	}
	return &repositoryImpl{db: db, rd: rd}, nil
}
