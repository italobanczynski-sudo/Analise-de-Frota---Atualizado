package coleta_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"analisefrota/adapters/dbPostgres"
)

type Repository_I interface {
	// Buscar devolve a situação; cliente que nunca foi buscado volta vazio, sem erro
	Buscar(ctx context.Context, clienteID int64) (SituacaoRow_t, error)
	MarcarInicio(ctx context.Context, clienteID int64, em time.Time) error
	MarcarFim(ctx context.Context, clienteID int64, situacao, mensagem string, avisos []string, ultimoDia *string, em time.Time) error
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Buscar(ctx context.Context, clienteID int64) (SituacaoRow_t, error) {
	var row SituacaoRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscar, clienteID)
	if errors.Is(err, sql.ErrNoRows) {
		return SituacaoRow_t{ClienteID: clienteID}, nil
	}
	return row, err
}

func (r *Repository_t) MarcarInicio(ctx context.Context, clienteID int64, em time.Time) error {
	_, err := r.Db.ExecContext(ctx, sqlMarcarInicio, clienteID, em)
	return err
}

func (r *Repository_t) MarcarFim(ctx context.Context, clienteID int64, situacao, mensagem string, avisos []string, ultimoDia *string, em time.Time) error {
	if avisos == nil {
		avisos = []string{}
	}
	_, err := r.Db.ExecContext(ctx, sqlMarcarFim, clienteID, situacao, mensagem, pq.StringArray(avisos), ultimoDia, em)
	return err
}
