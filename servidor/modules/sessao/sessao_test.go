package sessao_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/common/cripto"
	"analisefrota/modules/sessao"
	"analisefrota/modules/sessao/sessao_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Criar(ctx context.Context, tokenHash []byte, usuarioID int64, expira time.Time) error {
	return m.Called(ctx, tokenHash, usuarioID, expira).Error(0)
}
func (m *RepositoryMock_t) UsuarioDaSessao(ctx context.Context, tokenHash []byte, agora time.Time) (int64, error) {
	a := m.Called(ctx, tokenHash, agora)
	return a.Get(0).(int64), a.Error(1)
}
func (m *RepositoryMock_t) Apagar(ctx context.Context, tokenHash []byte) error {
	return m.Called(ctx, tokenHash).Error(0)
}
func (m *RepositoryMock_t) ApagarDoUsuario(ctx context.Context, usuarioID int64) error {
	return m.Called(ctx, usuarioID).Error(0)
}
func (m *RepositoryMock_t) ApagarVencidas(ctx context.Context, agora time.Time) (int64, error) {
	a := m.Called(ctx, agora)
	return a.Get(0).(int64), a.Error(1)
}

type SessaoSuite struct {
	suite.Suite
	repo  *RepositoryMock_t
	svc   sessao.Service_I
	agora time.Time
	ctx   context.Context
}

func (s *SessaoSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.agora = time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)
	s.svc = sessao.NewSessaoCom(s.repo, func() time.Time { return s.agora }, 12*time.Hour)
	s.ctx = context.Background()
}

func (s *SessaoSuite) TestCriarGuardaSoOHashEVence12hDepois() {
	var hash []byte
	s.repo.On("Criar", s.ctx, mock.Anything, int64(7), s.agora.Add(12*time.Hour)).Run(func(a mock.Arguments) { hash = a.Get(1).([]byte) }).Return(nil)

	ses, err := s.svc.Criar(s.ctx, 7)

	s.Require().NoError(err)
	s.Equal(cripto.HashToken(ses.Token), hash)
	s.NotEqual([]byte(ses.Token), hash)
}

func (s *SessaoSuite) TestValidar() {
	tk, _ := cripto.TokenAleatorio()
	s.repo.On("UsuarioDaSessao", s.ctx, cripto.HashToken(tk), s.agora).Return(int64(7), nil)
	id, err := s.svc.Validar(s.ctx, tk)
	s.Require().NoError(err)
	s.Equal(int64(7), id)
}

func (s *SessaoSuite) TestSessaoVencidaOuDesconhecida() {
	tk, _ := cripto.TokenAleatorio()
	s.repo.On("UsuarioDaSessao", s.ctx, cripto.HashToken(tk), s.agora).Return(int64(0), sessao_repo.ErrNaoEncontrada)
	_, err := s.svc.Validar(s.ctx, tk)
	s.ErrorIs(err, sessao.ErrSessaoInvalida)
}

func (s *SessaoSuite) TestTokenForaDoFormatoNemVaiAoBanco() {
	_, err := s.svc.Validar(s.ctx, "abc")
	s.ErrorIs(err, sessao.ErrSessaoInvalida)
	s.repo.AssertNotCalled(s.T(), "UsuarioDaSessao", mock.Anything, mock.Anything, mock.Anything)
}

func TestSessaoSuite(t *testing.T) { suite.Run(t, new(SessaoSuite)) }
