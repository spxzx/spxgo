package orm

import (
	"database/sql"
	spxLog "gitbuh.com/spxzx/spxgo/log"
	"strings"
	"time"
)

type SpxDb struct {
	db     *sql.DB
	logger *spxLog.Logger
}

// SpxSession 使每个操作都在一个会话内完成 相互独立
type SpxSession struct {
	db              *SpxDb
	tx              *sql.Tx
	beginTx         bool
	tableName       string
	fieldName       []string
	placeHolder     []string
	values          []any
	updateParam     strings.Builder
	conditionParam  strings.Builder
	conditionValues []any
}

func Open(driverName string, source string) *SpxDb {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	spxDb := &SpxDb{
		db:     db,
		logger: spxLog.Default(),
	}
	return spxDb
}

func (db *SpxDb) Close() error {
	err := db.db.Close()
	if err != nil {
		return err
	}
	return nil
}

// SetMaxIdleConns 设置最大空闲连接数，默认不配置，2两个最大空闲连接
func (db *SpxDb) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

// SetMaxOpenConns 设置最大连接数，默认不设置，不限制最大连接数
func (db *SpxDb) SetMaxOpenConns(n int) {
	db.db.SetMaxOpenConns(n)
}

// SetConnMaxLifetime 设置连接最大存活时间
func (db *SpxDb) SetConnMaxLifetime(duration time.Duration) {
	db.db.SetConnMaxLifetime(duration)
}

// SetConnMaxIdleTime 设置空闲连接最大存活时间
func (db *SpxDb) SetConnMaxIdleTime(duration time.Duration) {
	db.db.SetConnMaxIdleTime(duration)
}

func (db *SpxDb) New(table string) *SpxSession {
	return &SpxSession{
		db:        db,
		tableName: table,
	}
}

// Table Deprecated
func (s *SpxSession) Table(name string) *SpxSession {
	s.tableName = name
	return s
}
