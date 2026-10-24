// Package dbPostgres guarda a conexão única (Singleton) com o Postgres.
package dbPostgres

import (
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"analisefrota/environment"
)

var (
	conn     *sqlx.DB
	connOnce sync.Once
)

// GetPgConn abre a conexão na primeira chamada; o banco pode demorar a subir no Docker, então tenta por até 30 s
func GetPgConn() *sqlx.DB {
	connOnce.Do(func() {
		var err error
		for tentativa := 1; tentativa <= 15; tentativa++ {
			conn, err = sqlx.Connect("postgres", environment.Cfg.Database.URL)
			if err == nil {
				break
			}
			logrus.Warnf("postgres: tentativa %d sem conexão: %v", tentativa, err)
			time.Sleep(2 * time.Second)
		}
		if err != nil {
			logrus.Fatalf("postgres: não conectou: %v", err)
		}
		conn.SetMaxOpenConns(10)
		conn.SetMaxIdleConns(5)
		conn.SetConnMaxLifetime(30 * time.Minute)
	})
	return conn
}
