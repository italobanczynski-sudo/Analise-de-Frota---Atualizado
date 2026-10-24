// Package migrations aplica, em ordem, os arquivos .sql desta pasta que ainda não rodaram no banco.
package migrations

import (
	"embed"
	"fmt"
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

//go:embed *.sql
var arquivos embed.FS

const sqlControle = `CREATE TABLE IF NOT EXISTS schema_migrations (
	versao     TEXT PRIMARY KEY,
	aplicada_em TIMESTAMPTZ NOT NULL DEFAULT now()
)`

func Executar(db *sqlx.DB) error {
	if _, err := db.Exec(sqlControle); err != nil {
		return fmt.Errorf("migrations: tabela de controle: %w", err)
	}
	entradas, err := arquivos.ReadDir(".")
	if err != nil {
		return err
	}
	nomes := make([]string, 0, len(entradas))
	for _, e := range entradas {
		nomes = append(nomes, e.Name())
	}
	sort.Strings(nomes)
	for _, nome := range nomes {
		var ja bool
		if err := db.Get(&ja, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE versao = $1)`, nome); err != nil {
			return err
		}
		if ja {
			continue
		}
		conteudo, err := arquivos.ReadFile(nome)
		if err != nil {
			return err
		}
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(conteudo)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migrations: %s: %w", nome, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (versao) VALUES ($1)`, nome); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		logrus.Infof("migrations: aplicada %s", nome)
	}
	return nil
}
