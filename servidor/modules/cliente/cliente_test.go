package cliente_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/common/cripto"
	"analisefrota/modules/cliente"
	"analisefrota/modules/cliente/cliente_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Listar(ctx context.Context) ([]cliente_repo.ClienteRow_t, error) {
	a := m.Called(ctx)
	return a.Get(0).([]cliente_repo.ClienteRow_t), a.Error(1)
}
func (m *RepositoryMock_t) BuscarPorID(ctx context.Context, id int64) (cliente_repo.ClienteRow_t, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(cliente_repo.ClienteRow_t), a.Error(1)
}
func (m *RepositoryMock_t) BuscarPorNome(ctx context.Context, nome string) (cliente_repo.ClienteRow_t, error) {
	a := m.Called(ctx, nome)
	return a.Get(0).(cliente_repo.ClienteRow_t), a.Error(1)
}
func (m *RepositoryMock_t) Criar(ctx context.Context, nome, hubspotID string) (int64, error) {
	a := m.Called(ctx, nome, hubspotID)
	return a.Get(0).(int64), a.Error(1)
}
func (m *RepositoryMock_t) SalvarTokenCifrado(ctx context.Context, id int64, cifrado []byte) error {
	return m.Called(ctx, id, cifrado).Error(0)
}
func (m *RepositoryMock_t) TokenCifrado(ctx context.Context, id int64) ([]byte, error) {
	a := m.Called(ctx, id)
	b, _ := a.Get(0).([]byte)
	return b, a.Error(1)
}

type ClienteSuite struct {
	suite.Suite
	repo  *RepositoryMock_t
	chave []byte
	svc   cliente.Service_I
	ctx   context.Context
}

func (s *ClienteSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.chave = bytes.Repeat([]byte{3}, 32)
	s.svc = cliente.NewClienteCom(s.repo, s.chave)
	s.ctx = context.Background()
}

func (s *ClienteSuite) TestSalvarTokenLimpaECifra() {
	var gravado []byte
	s.repo.On("SalvarTokenCifrado", s.ctx, int64(7), mock.Anything).Run(func(a mock.Arguments) { gravado = a.Get(2).([]byte) }).Return(nil)

	s.Require().NoError(s.svc.SalvarToken(s.ctx, 7, "Bearer \"abc.def-123\"\n"))

	s.NotContains(string(gravado), "abc.def-123")
	txt, err := cripto.Decifrar(s.chave, gravado)
	s.Require().NoError(err)
	s.Equal("abc.def-123", txt)
}

func (s *ClienteSuite) TestTokenVoltaDecifrado() {
	cif, _ := cripto.Cifrar(s.chave, "tk-exemplo")
	s.repo.On("TokenCifrado", s.ctx, int64(2)).Return(cif, nil)
	tk, err := s.svc.Token(s.ctx, 2)
	s.Require().NoError(err)
	s.Equal("tk-exemplo", tk)
}

func (s *ClienteSuite) TestSemTokenSalvo() {
	s.repo.On("TokenCifrado", s.ctx, int64(2)).Return(nil, nil)
	_, err := s.svc.Token(s.ctx, 2)
	s.ErrorIs(err, cliente.ErrSemToken)
}

func (s *ClienteSuite) TestTokenComCaractereEstranhoERecusado() {
	err := s.svc.SalvarToken(s.ctx, 1, "abc<script>")
	s.ErrorIs(err, cliente.ErrTokenEstranho)
	s.repo.AssertNotCalled(s.T(), "SalvarTokenCifrado", mock.Anything, mock.Anything, mock.Anything)
}

func (s *ClienteSuite) TestCriarValidaNomeEHubspot() {
	_, err := s.svc.Criar(s.ctx, "  ", "")
	s.ErrorIs(err, cliente.ErrNomeVazio)
	_, err = s.svc.Criar(s.ctx, "Cliente X", "abc")
	s.ErrorIs(err, cliente.ErrHubspotID)
}

func (s *ClienteSuite) TestCriarNomeRepetido() {
	s.repo.On("Criar", s.ctx, "Cliente X", "").Return(int64(0), cliente_repo.ErrRepetido)
	_, err := s.svc.Criar(s.ctx, " Cliente X ", "")
	s.ErrorIs(err, cliente.ErrNomeRepetido)
}

func (s *ClienteSuite) TestBuscarNaoEncontrado() {
	s.repo.On("BuscarPorID", s.ctx, int64(9)).Return(cliente_repo.ClienteRow_t{}, cliente_repo.ErrNaoEncontrado)
	_, err := s.svc.Buscar(s.ctx, 9)
	s.ErrorIs(err, cliente.ErrNaoEncontrado)
}

func TestClienteSuite(t *testing.T) { suite.Run(t, new(ClienteSuite)) }
