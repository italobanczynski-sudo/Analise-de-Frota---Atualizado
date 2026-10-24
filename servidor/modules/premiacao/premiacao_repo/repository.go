package premiacao_repo

import (
	"context"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

type Repository_I interface {
	SubstituirMes(ctx context.Context, clienteID int64, mes string, rows []PremiacaoRow_t) error
	Listar(ctx context.Context, clienteID int64, mes string) ([]PremiacaoRow_t, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) SubstituirMes(ctx context.Context, clienteID int64, mes string, rows []PremiacaoRow_t) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, sqlApagarMes, clienteID, mes); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := tx.NamedExecContext(ctx, sqlInserir, row); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64, mes string) ([]PremiacaoRow_t, error) {
	rows := []PremiacaoRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID, mes)
	return rows, err
}
