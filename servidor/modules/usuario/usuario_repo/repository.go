package usuario_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"analisefrota/adapters/dbPostgres"
)

var (
	ErrSemCodigo       = errors.New("usuario_repo: sem código ativo")
	ErrNaoEncontrado   = errors.New("usuario_repo: não encontrado")
	ErrRepetido        = errors.New("usuario_repo: e-mail repetido")
	ErrConviteInvalido = errors.New("usuario_repo: convite inválido")
)

type Repository_I interface {
	Listar(ctx context.Context) ([]UsuarioRow_t, error)
	BuscarPorID(ctx context.Context, id int64) (UsuarioRow_t, error)
	BuscarPorEmail(ctx context.Context, email string) (UsuarioRow_t, error)
	Criar(ctx context.Context, nome, email, papel string) (int64, error)
	Alterar(ctx context.Context, id int64, papel string, ativo bool) error
	ContarAdminsAtivos(ctx context.Context) (int, error)
	MarcarAcesso(ctx context.Context, id int64, em time.Time) error
	// CriarConvite invalida os convites pendentes da pessoa e grava o novo, numa transação
	CriarConvite(ctx context.Context, usuarioID int64, tokenHash []byte, expira, agora time.Time) error
	// UsarConvite marca o convite como usado e grava a senha, numa transação; devolve o id do usuário
	UsarConvite(ctx context.Context, tokenHash []byte, senhaHash string, agora time.Time) (int64, error)
	// UltimoCodigoEm diz quando saiu o último código da pessoa (zero se nunca)
	UltimoCodigoEm(ctx context.Context, usuarioID int64) (time.Time, error)
	// CriarCodigo invalida os códigos pendentes da pessoa e grava o novo, numa transação
	CriarCodigo(ctx context.Context, usuarioID int64, codigoHash []byte, expira, agora time.Time) error
	CodigoAtivo(ctx context.Context, usuarioID int64, agora time.Time) (CodigoRow_t, error)
	SomarTentativa(ctx context.Context, codigoID int64) error
	// UsarCodigo marca o código como usado, grava a senha e invalida convites pendentes, numa transação
	UsarCodigo(ctx context.Context, codigoID, usuarioID int64, senhaHash string, agora time.Time) error
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Listar(ctx context.Context) ([]UsuarioRow_t, error) {
	rows := []UsuarioRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar)
	return rows, err
}

func (r *Repository_t) BuscarPorID(ctx context.Context, id int64) (UsuarioRow_t, error) {
	var row UsuarioRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscarPorID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNaoEncontrado
	}
	return row, err
}

func (r *Repository_t) BuscarPorEmail(ctx context.Context, email string) (UsuarioRow_t, error) {
	var row UsuarioRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscarPorEmail, email)
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrNaoEncontrado
	}
	return row, err
}

func (r *Repository_t) Criar(ctx context.Context, nome, email, papel string) (int64, error) {
	var id int64
	err := r.Db.GetContext(ctx, &id, sqlCriar, nome, email, papel)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return 0, ErrRepetido
	}
	return id, err
}

func (r *Repository_t) Alterar(ctx context.Context, id int64, papel string, ativo bool) error {
	res, err := r.Db.ExecContext(ctx, sqlAlterar, id, papel, ativo)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func (r *Repository_t) ContarAdminsAtivos(ctx context.Context) (int, error) {
	var n int
	err := r.Db.GetContext(ctx, &n, sqlContarAdminsAtivos)
	return n, err
}

func (r *Repository_t) MarcarAcesso(ctx context.Context, id int64, em time.Time) error {
	_, err := r.Db.ExecContext(ctx, sqlMarcarAcesso, id, em)
	return err
}

func (r *Repository_t) CriarConvite(ctx context.Context, usuarioID int64, tokenHash []byte, expira, agora time.Time) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, sqlInvalidarConvites, usuarioID, agora); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sqlCriarConvite, tokenHash, usuarioID, expira); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository_t) UsarConvite(ctx context.Context, tokenHash []byte, senhaHash string, agora time.Time) (int64, error) {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	var id int64
	if err := tx.GetContext(ctx, &id, sqlUsarConvite, tokenHash, agora); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrConviteInvalido
		}
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, sqlDefinirSenha, id, senhaHash); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (r *Repository_t) UltimoCodigoEm(ctx context.Context, usuarioID int64) (time.Time, error) {
	var t sql.NullTime
	err := r.Db.GetContext(ctx, &t, sqlUltimoCodigoEm, usuarioID)
	return t.Time, err
}

func (r *Repository_t) CriarCodigo(ctx context.Context, usuarioID int64, codigoHash []byte, expira, agora time.Time) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, sqlInvalidarCodigos, usuarioID, agora); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sqlCriarCodigo, usuarioID, codigoHash, expira, agora); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository_t) CodigoAtivo(ctx context.Context, usuarioID int64, agora time.Time) (CodigoRow_t, error) {
	var row CodigoRow_t
	err := r.Db.GetContext(ctx, &row, sqlCodigoAtivo, usuarioID, agora)
	if errors.Is(err, sql.ErrNoRows) {
		return row, ErrSemCodigo
	}
	return row, err
}

func (r *Repository_t) SomarTentativa(ctx context.Context, codigoID int64) error {
	_, err := r.Db.ExecContext(ctx, sqlSomarTentativa, codigoID)
	return err
}

func (r *Repository_t) UsarCodigo(ctx context.Context, codigoID, usuarioID int64, senhaHash string, agora time.Time) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	res, err := tx.ExecContext(ctx, sqlUsarCodigo, codigoID, usuarioID, agora)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrSemCodigo // outro pedido usou o código ao mesmo tempo
	}
	if _, err := tx.ExecContext(ctx, sqlDefinirSenha, usuarioID, senhaHash); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, sqlInvalidarConvites, usuarioID, agora); err != nil {
		return err
	}
	return tx.Commit()
}
