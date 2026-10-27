package contato_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/common/cripto"
	"analisefrota/modules/contato"
	"analisefrota/modules/contato/contato_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Listar(ctx context.Context, clienteID int64) ([]contato_repo.ContatoRow_t, error) {
	a := m.Called(ctx, clienteID)
	return a.Get(0).([]contato_repo.ContatoRow_t), a.Error(1)
}
func (m *RepositoryMock_t) ListarAtivos(ctx context.Context, clienteID int64) ([]contato_repo.ContatoRow_t, error) {
	a := m.Called(ctx, clienteID)
	return a.Get(0).([]contato_repo.ContatoRow_t), a.Error(1)
}
func (m *RepositoryMock_t) Salvar(ctx context.Context, row contato_repo.ContatoRow_t) error {
	return m.Called(ctx, row).Error(0)
}
func (m *RepositoryMock_t) Ativar(ctx context.Context, clienteID int64, motorista string, ativo bool) error {
	return m.Called(ctx, clienteID, motorista, ativo).Error(0)
}
func (m *RepositoryMock_t) Remover(ctx context.Context, clienteID int64, motorista string) error {
	return m.Called(ctx, clienteID, motorista).Error(0)
}

type ContatoSuite struct {
	suite.Suite
	repo  *RepositoryMock_t
	chave []byte
	agora time.Time
	svc   contato.Service_I
	ctx   context.Context
}

func (s *ContatoSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.chave = bytes.Repeat([]byte{5}, 32)
	s.agora = time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	s.svc = contato.NewContatoCom(s.repo, s.chave, func() time.Time { return s.agora })
	s.ctx = context.Background()
}

func (s *ContatoSuite) TestNormalizaTelefone() {
	for entrada, esperado := range map[string]string{
		"(11) 91234-5678":   "+5511912345678",
		"+55 11 91234-5678": "+5511912345678",
		"011 91234 5678":    "+5511912345678",
		"4133334444":        "+554133334444",
		"55 41 3333-4444":   "+554133334444",
		"(51) 8123-4567":    "+5551981234567", // celular antigo, sem o 9
		"51 9123-4567":      "+5551991234567",
	} {
		tel, err := contato.NormalizarTelefone(entrada)
		s.Require().NoError(err, entrada)
		s.Equal(esperado, tel, entrada)
	}
	for _, ruim := range []string{"", "1234", "(01) 91234-5678", "+1 202 555 0100", "11 91234-56789"} {
		_, err := contato.NormalizarTelefone(ruim)
		s.ErrorIs(err, contato.ErrTelefone, ruim)
	}
}

func (s *ContatoSuite) TestSalvarCifraEGuardaSoOFinal() {
	var gravado contato_repo.ContatoRow_t
	s.repo.On("Salvar", s.ctx, mock.Anything).Run(func(a mock.Arguments) { gravado = a.Get(1).(contato_repo.ContatoRow_t) }).Return(nil)

	c, err := s.svc.Salvar(s.ctx, 3, "  Ana Souza ", "(11) 91234-5678", true, 9)

	s.Require().NoError(err)
	s.Equal("Ana Souza", c.Motorista)
	s.Equal("5678", c.TelefoneFinal)
	s.Equal(s.agora, c.AutorizadoEm)
	s.NotContains(string(gravado.TelefoneCifrado), "91234")
	tel, err := cripto.Decifrar(s.chave, gravado.TelefoneCifrado)
	s.Require().NoError(err)
	s.Equal("+5511912345678", tel)
	s.Equal(int64(9), *gravado.AutorizadoPor)
}

func (s *ContatoSuite) TestSalvarExigeAutorizacao() {
	_, err := s.svc.Salvar(s.ctx, 3, "Ana", "(11) 91234-5678", false, 9)
	s.ErrorIs(err, contato.ErrSemAutorizacao)
	s.repo.AssertNotCalled(s.T(), "Salvar", mock.Anything, mock.Anything)
}

func (s *ContatoSuite) TestSalvarExigeMotorista() {
	_, err := s.svc.Salvar(s.ctx, 3, "   ", "(11) 91234-5678", true, 9)
	s.ErrorIs(err, contato.ErrMotorista)
}

func (s *ContatoSuite) TestDestinosDecifraEPulaOQueNaoAbre() {
	bom, _ := cripto.Cifrar(s.chave, "+5511912345678")
	outraChave, _ := cripto.Cifrar(bytes.Repeat([]byte{6}, 32), "+5511900000000")
	s.repo.On("ListarAtivos", s.ctx, int64(3)).Return([]contato_repo.ContatoRow_t{
		{Motorista: "Ana", TelefoneCifrado: bom}, {Motorista: "Beto", TelefoneCifrado: outraChave},
	}, nil)

	d, err := s.svc.Destinos(s.ctx, 3)

	s.Require().NoError(err)
	s.Equal([]contato.Destino_t{{Motorista: "Ana", Telefone: "+5511912345678"}}, d)
}

func (s *ContatoSuite) TestRemoverQuemNaoExiste() {
	s.repo.On("Remover", s.ctx, int64(3), "Ana").Return(contato_repo.ErrNaoEncontrado)
	s.ErrorIs(s.svc.Remover(s.ctx, 3, "Ana"), contato.ErrNaoEncontrado)
}

func TestContatoSuite(t *testing.T) { suite.Run(t, new(ContatoSuite)) }
