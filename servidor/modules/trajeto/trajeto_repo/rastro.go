package trajeto_repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"analisefrota/adapters/dbPostgres"
)

type RastroRepo_I interface {
	// SalvarRastro grava a posição (a mesma placa e data não repete)
	SalvarRastro(ctx context.Context, row RastroRow_t) error
	// ListarRastro devolve os pontos de ignição ligada das placas desde a data, em ordem de placa e data
	ListarRastro(ctx context.Context, clienteID int64, placas []string, desde time.Time) ([]RastroRow_t, error)
	// LimparRastro apaga o que for anterior à data e diz quantos apagou
	LimparRastro(ctx context.Context, antes time.Time) (int64, error)
}

type RastroRepo_t struct {
	Db *sqlx.DB
}

func NewRastroRepo() RastroRepo_I {
	return &RastroRepo_t{Db: dbPostgres.GetPgConn()}
}

func (r *RastroRepo_t) SalvarRastro(ctx context.Context, row RastroRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlSalvarRastro, row)
	return err
}

func (r *RastroRepo_t) ListarRastro(ctx context.Context, clienteID int64, placas []string, desde time.Time) ([]RastroRow_t, error) {
	rows := []RastroRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListarRastro, clienteID, pq.Array(placas), desde)
	return rows, err
}

func (r *RastroRepo_t) LimparRastro(ctx context.Context, antes time.Time) (int64, error) {
	res, err := r.Db.ExecContext(ctx, sqlLimparRastro, antes)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
