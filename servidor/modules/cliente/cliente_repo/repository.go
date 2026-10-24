package cliente_repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"analisefrota/adapters/dbPostgres"
)

// ErrNaoEncontrado e ErrRepetido são traduzidos pelo módulo em mensagens para o usuário
var (
	ErrNaoEncontrado = errors.New("cliente_repo: não encontrado")
	ErrRepetido      = errors.New("cliente_repo: nome repetido")
)

type Repository_I interface {
	Listar(ctx context.Context) ([]ClienteRow_t, error)
	BuscarPorID(ctx context.Context, id int64) (ClienteRow_t, error)
	BuscarPorNome(ctx context.Context, nome string) (ClienteRow_t, error)
	Criar(ctx context.Context, nome, hubspotID string) (int64, error)
	SalvarTokenCifrado(ctx context.Context, id int64, cifrado []byte) error
	TokenCifrado(ctx context.Context, id int64) ([]byte, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Listar(ctx context.Context) ([]ClienteRow_t, error) {
	rows := []ClienteRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar)
	return rows, err
}

func (r *Repository_t) BuscarPorID(ctx context.Context, id int64) (ClienteRow_t, error) {
	var row ClienteRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscarPorID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNaoEncontrado
	}
	return row, err
}

func (r *Repository_t) BuscarPorNome(ctx context.Context, nome string) (ClienteRow_t, error) {
	var row ClienteRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscarPorNome, nome)
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNaoEncontrado
	}
	return row, err
}

func (r *Repository_t) Criar(ctx context.Context, nome, hubspotID string) (int64, error) {
	var id int64
	err := r.Db.GetContext(ctx, &id, sqlCriar, nome, hubspotID)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return 0, ErrRepetido
	}
	return id, err
}

func (r *Repository_t) SalvarTokenCifrado(ctx context.Context, id int64, cifrado []byte) error {
	res, err := r.Db.ExecContext(ctx, sqlSalvarToken, id, cifrado)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Repository_t) TokenCifrado(ctx context.Context, id int64) ([]byte, error) {
	var b []byte
	err := r.Db.GetContext(ctx, &b, sqlTokenCifrado, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNaoEncontrado
	}
	return b, err
}
