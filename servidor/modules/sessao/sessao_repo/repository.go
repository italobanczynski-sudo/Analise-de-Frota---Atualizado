package sessao_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

var ErrNaoEncontrada = errors.New("sessao_repo: não encontrada")

type Repository_I interface {
	Criar(ctx context.Context, tokenHash []byte, usuarioID int64, expira time.Time) error
	UsuarioDaSessao(ctx context.Context, tokenHash []byte, agora time.Time) (int64, error)
	Apagar(ctx context.Context, tokenHash []byte) error
	ApagarDoUsuario(ctx context.Context, usuarioID int64) error
	ApagarVencidas(ctx context.Context, agora time.Time) (int64, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Criar(ctx context.Context, tokenHash []byte, usuarioID int64, expira time.Time) error {
	_, err := r.Db.ExecContext(ctx, sqlCriar, tokenHash, usuarioID, expira)
	return err
}

func (r *Repository_t) UsuarioDaSessao(ctx context.Context, tokenHash []byte, agora time.Time) (int64, error) {
	var id int64
	err := r.Db.GetContext(ctx, &id, sqlUsuarioDaSessao, tokenHash, agora)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNaoEncontrada
	}
	return id, err
}

func (r *Repository_t) Apagar(ctx context.Context, tokenHash []byte) error {
	_, err := r.Db.ExecContext(ctx, sqlApagar, tokenHash)
	return err
}

func (r *Repository_t) ApagarDoUsuario(ctx context.Context, usuarioID int64) error {
	_, err := r.Db.ExecContext(ctx, sqlApagarDoUsuario, usuarioID)
	return err
}

func (r *Repository_t) ApagarVencidas(ctx context.Context, agora time.Time) (int64, error) {
	res, err := r.Db.ExecContext(ctx, sqlApagarVencidas, agora)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
