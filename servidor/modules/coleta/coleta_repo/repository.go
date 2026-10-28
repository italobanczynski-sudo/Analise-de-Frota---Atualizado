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
	// MarcarInterrompidas troca "rodando" por erro nas buscas que o servidor não terminou (reiniciou no meio)
	MarcarInterrompidas(ctx context.Context, em time.Time) (int64, error)
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

const sqlMarcarInterrompidas = `UPDATE coletas SET situacao = 'erro', terminada_em = $1,
	mensagem = 'interrompida: o servidor reiniciou no meio da busca (a próxima busca refaz os dias que faltaram)' WHERE situacao = 'rodando'`

func (r *Repository_t) MarcarInterrompidas(ctx context.Context, em time.Time) (int64, error) {
	res, err := r.Db.ExecContext(ctx, sqlMarcarInterrompidas, em)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository_t) MarcarFim(ctx context.Context, clienteID int64, situacao, mensagem string, avisos []string, ultimoDia *string, em time.Time) error {
	if avisos == nil {
		avisos = []string{}
	}
	_, err := r.Db.ExecContext(ctx, sqlMarcarFim, clienteID, situacao, mensagem, pq.StringArray(avisos), ultimoDia, em)
	return err
}
