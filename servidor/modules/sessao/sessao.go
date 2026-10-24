// Package sessao cuida dos logins abertos: cria, confere e encerra.
package sessao

import (
	"context"
	"errors"
	"time"

	"analisefrota/common/cripto"
	"analisefrota/environment"
	"analisefrota/modules/sessao/sessao_repo"
)

type Service_I interface {
	Criar(ctx context.Context, usuarioID int64) (Sessao_t, error)
	// Validar devolve de quem é a sessão; vencida ou desconhecida dá ErrSessaoInvalida
	Validar(ctx context.Context, token string) (int64, error)
	Encerrar(ctx context.Context, token string) error
	// EncerrarDoUsuario derruba todos os logins da pessoa (ao desativar ou trocar o perfil)
	EncerrarDoUsuario(ctx context.Context, usuarioID int64) error
	LimparVencidas(ctx context.Context) (int64, error)
}

type sessao_t struct {
	repo    sessao_repo.Repository_I
	agora   func() time.Time
	duracao time.Duration
}

func NewSessao() Service_I {
	return &sessao_t{repo: sessao_repo.NewRepository(), agora: time.Now, duracao: time.Duration(environment.Cfg.Seguranca.SessaoHoras) * time.Hour}
}

// NewSessaoCom recebe as dependências (usado nos testes)
func NewSessaoCom(repo sessao_repo.Repository_I, agora func() time.Time, duracao time.Duration) Service_I {
	return &sessao_t{repo: repo, agora: agora, duracao: duracao}
}

func (s *sessao_t) Criar(ctx context.Context, usuarioID int64) (Sessao_t, error) {
	token, err := cripto.TokenAleatorio()
	if err != nil {
		return Sessao_t{}, err
	}
	ses := Sessao_t{Token: token, UsuarioID: usuarioID, ExpiraEm: s.agora().Add(s.duracao)}
	return ses, s.repo.Criar(ctx, cripto.HashToken(token), usuarioID, ses.ExpiraEm)
}

func (s *sessao_t) Validar(ctx context.Context, token string) (int64, error) {
	if !formatoValido(token) {
		return 0, ErrSessaoInvalida
	}
	id, err := s.repo.UsuarioDaSessao(ctx, cripto.HashToken(token), s.agora())
	if errors.Is(err, sessao_repo.ErrNaoEncontrada) {
		return 0, ErrSessaoInvalida
	}
	return id, err
}

func (s *sessao_t) Encerrar(ctx context.Context, token string) error {
	if !formatoValido(token) {
		return nil
	}
	return s.repo.Apagar(ctx, cripto.HashToken(token))
}

func (s *sessao_t) EncerrarDoUsuario(ctx context.Context, usuarioID int64) error {
	return s.repo.ApagarDoUsuario(ctx, usuarioID)
}

func (s *sessao_t) LimparVencidas(ctx context.Context) (int64, error) {
	return s.repo.ApagarVencidas(ctx, s.agora())
}
