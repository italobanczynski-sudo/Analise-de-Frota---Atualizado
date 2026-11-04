package trajeto_repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

type Repository_I interface {
	// Listar devolve as janelas já consultadas da placa que começam em [ini, fim)
	Listar(ctx context.Context, clienteID int64, placa string, ini, fim time.Time) ([]PosicaoRow_t, error)
	// Salvar grava (ou substitui) uma janela consultada
	Salvar(ctx context.Context, row PosicaoRow_t) error
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64, placa string, ini, fim time.Time) ([]PosicaoRow_t, error) {
	rows := []PosicaoRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID, placa, ini, fim)
	return rows, err
}

func (r *Repository_t) Salvar(ctx context.Context, row PosicaoRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlSalvar, row)
	return err
}
