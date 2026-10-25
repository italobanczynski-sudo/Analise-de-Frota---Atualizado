package base_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/modules/base"
	"analisefrota/modules/base/base_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Listar(ctx context.Context, colecao string) ([]base_repo.DocumentoRow_t, error) {
	a := m.Called(ctx, colecao)
	return a.Get(0).([]base_repo.DocumentoRow_t), a.Error(1)
}
func (m *RepositoryMock_t) Gravar(ctx context.Context, colecao, id string, dados []byte, usuarioID int64) error {
	return m.Called(ctx, colecao, id, dados, usuarioID).Error(0)
}
func (m *RepositoryMock_t) Apagar(ctx context.Context, colecao, id string) error {
	return m.Called(ctx, colecao, id).Error(0)
}

type BaseSuite struct {
	suite.Suite
	repo *RepositoryMock_t
	svc  base.Service_I
	ctx  context.Context
}

func (s *BaseSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.svc = base.NewBaseCom(s.repo)
	s.ctx = context.Background()
}

func (s *BaseSuite) TestGravaDocumentoDoPainel() {
	d := []byte(`{"alvo":"veiculos","cliente":"Caju","texto":"a;b"}`)
	s.repo.On("Gravar", s.ctx, "planilhas", "veiculos~caju~2026-10-01~p1", d, int64(3)).Return(nil)
	s.NoError(s.svc.Gravar(s.ctx, "planilhas", "veiculos~caju~2026-10-01~p1", d, 3))
}

func (s *BaseSuite) TestRecusaEntradaEstranha() {
	s.ErrorIs(s.svc.Gravar(s.ctx, "usuarios", "x", []byte(`{}`), 1), base.ErrColecao, "coleção fora da lista")
	s.ErrorIs(s.svc.Gravar(s.ctx, "planilhas", "../x", []byte(`{}`), 1), base.ErrID)
	s.ErrorIs(s.svc.Gravar(s.ctx, "planilhas", "x", []byte(`[1]`), 1), base.ErrDados)
	s.ErrorIs(s.svc.Gravar(s.ctx, "planilhas", "x", []byte(`{"a":`), 1), base.ErrDados)
	grande := []byte(`{"t":"` + strings.Repeat("a", base.TamanhoMaximo) + `"}`)
	s.ErrorIs(s.svc.Gravar(s.ctx, "planilhas", "x", grande, 1), base.ErrGrande)
	s.repo.AssertNotCalled(s.T(), "Gravar", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *BaseSuite) TestCriarGeraIdNovo() {
	s.repo.On("Gravar", s.ctx, "tratativas", mock.Anything, mock.Anything, int64(1)).Return(nil)
	a, err := s.svc.Criar(s.ctx, "tratativas", []byte(`{"x":1}`), 1)
	s.Require().NoError(err)
	b, _ := s.svc.Criar(s.ctx, "tratativas", []byte(`{"x":1}`), 1)
	s.Len(a, 20)
	s.NotEqual(a, b)
}

func (s *BaseSuite) TestApagarOQueNaoExisteNaoEErro() {
	s.repo.On("Apagar", s.ctx, "planilhas", "x").Return(base_repo.ErrNaoEncontrado)
	s.NoError(s.svc.Apagar(s.ctx, "planilhas", "x"))
}

func TestBaseSuite(t *testing.T) { suite.Run(t, new(BaseSuite)) }
