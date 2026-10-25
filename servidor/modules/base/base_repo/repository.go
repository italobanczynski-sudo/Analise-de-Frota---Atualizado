package base_repo

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

var ErrNaoEncontrado = errors.New("base_repo: não encontrado")

type DocumentoRow_t struct {
	ID           string    `db:"id"`
	Dados        []byte    `db:"dados"`
	AtualizadoEm time.Time `db:"atualizado_em"`
}

type Repository_I interface {
	Listar(ctx context.Context, colecao string) ([]DocumentoRow_t, error)
	Gravar(ctx context.Context, colecao, id string, dados []byte, usuarioID int64) error
	Apagar(ctx context.Context, colecao, id string) error
}

type Repository_t struct{ Db *sqlx.DB }

func NewRepository() Repository_I { return &Repository_t{Db: dbPostgres.GetPgConn()} }

const sqlListar = `SELECT id, dados, atualizado_em FROM base_documentos WHERE colecao = $1 ORDER BY id`

const sqlGravar = `INSERT INTO base_documentos (colecao, id, dados, atualizado_por) VALUES ($1, $2, $3, NULLIF($4, 0))
ON CONFLICT (colecao, id) DO UPDATE SET dados = EXCLUDED.dados, atualizado_em = now(), atualizado_por = EXCLUDED.atualizado_por`

const sqlApagar = `DELETE FROM base_documentos WHERE colecao = $1 AND id = $2`

func (r *Repository_t) Listar(ctx context.Context, colecao string) ([]DocumentoRow_t, error) {
	rows := []DocumentoRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, colecao)
	return rows, err
}

func (r *Repository_t) Gravar(ctx context.Context, colecao, id string, dados []byte, usuarioID int64) error {
	_, err := r.Db.ExecContext(ctx, sqlGravar, colecao, id, string(dados), usuarioID)
	return err
}

func (r *Repository_t) Apagar(ctx context.Context, colecao, id string) error {
	res, err := r.Db.ExecContext(ctx, sqlApagar, colecao, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNaoEncontrado
	}
	return nil
}
