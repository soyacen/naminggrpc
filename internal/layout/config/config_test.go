package config

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/soyacen/grocer/grocer/dbx"
	"github.com/soyacen/grocer/grocer/redisx"
	"go.yaml.in/yaml/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConfig(t *testing.T) {
	conf := &Config{
		Db: &dbx.Config{
			Configs: map[string]*dbx.Options{
				"cronjob": {
					DriverName:      wrapperspb.String("pgx"),
					Dsn:             wrapperspb.String("host=localhost port=5432 user=postgres password=postgres dbname=vibeme sslmode=disable pool_max_conns=20"),
					PingTimeout:     durationpb.New(5 * time.Second),
					MaxIdleConns:    wrapperspb.Int32(5),
					MaxOpenConns:    wrapperspb.Int32(20),
					ConnMaxLifetime: durationpb.New(30 * time.Minute),
					ConnMaxIdleTime: durationpb.New(10 * time.Minute),
				},
			},
		},
		Redis: &redisx.Config{
			Configs: map[string]*redisx.Options{
				"cronjob": {
					Addrs:           []string{"localhost:6379"},
					Password:        wrapperspb.String("123456"),
					DialTimeout:     durationpb.New(4 * time.Second),
					ReadTimeout:     durationpb.New(2 * time.Second),
					WriteTimeout:    durationpb.New(2 * time.Second),
					PoolSize:        wrapperspb.Int32(int32(runtime.NumCPU() * 128)),
					MinIdleConns:    wrapperspb.Int32(16),
					MaxIdleConns:    wrapperspb.Int32(64),
					PoolTimeout:     durationpb.New(4 * time.Second),
					ConnMaxIdleTime: durationpb.New(32 * time.Minute),
					ConnMaxLifetime: durationpb.New(1 * time.Hour),
					MaxRetries:      wrapperspb.Int32(2),
					MinRetryBackoff: durationpb.New(128 * time.Millisecond),
					MaxRetryBackoff: durationpb.New(512 * time.Millisecond),
					ReadBufferSize:  wrapperspb.Int32(1024),
					WriteBufferSize: wrapperspb.Int32(1024),
					EnableTracing:   wrapperspb.Bool(true),
					EnableMetrics:   wrapperspb.Bool(true),
				},
			},
		},
	}
	jsonData, err := protojson.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]any{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		t.Fatal(err)
	}
	yamlData, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(yamlData))
}
