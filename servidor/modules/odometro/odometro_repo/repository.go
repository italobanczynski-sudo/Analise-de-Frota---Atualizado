package odometro_repo

import (
	"context"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

type Repository_I interface {
	Substituir(ctx context.Context, clienteID int64, rows []OdometroRow_t) error
	Listar(ctx context.Context, clienteID int64) ([]OdometroRow_t, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Substituir(ctx context.Context, clienteID int64, rows []OdometroRow_t) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, sqlApagar, clienteID); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := tx.NamedExecContext(ctx, sqlInserir, row); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64) ([]OdometroRow_t, error) {
	rows := []OdometroRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID)
	return rows, err
}
