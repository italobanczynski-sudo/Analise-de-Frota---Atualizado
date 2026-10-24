package usuario_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"analisefrota/common/cripto"
	"analisefrota/modules/usuario"
	"analisefrota/modules/usuario/usuario_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Listar(ctx context.Context) ([]usuario_repo.UsuarioRow_t, error) {
	a := m.Called(ctx)
	return a.Get(0).([]usuario_repo.UsuarioRow_t), a.Error(1)
}
func (m *RepositoryMock_t) BuscarPorID(ctx context.Context, id int64) (usuario_repo.UsuarioRow_t, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(usuario_repo.UsuarioRow_t), a.Error(1)
}
func (m *RepositoryMock_t) BuscarPorEmail(ctx context.Context, email string) (usuario_repo.UsuarioRow_t, error) {
	a := m.Called(ctx, email)
	return a.Get(0).(usuario_repo.UsuarioRow_t), a.Error(1)
}
func (m *RepositoryMock_t) Criar(ctx context.Context, nome, email, papel string) (int64, error) {
	a := m.Called(ctx, nome, email, papel)
	return a.Get(0).(int64), a.Error(1)
}
func (m *RepositoryMock_t) Alterar(ctx context.Context, id int64, papel string, ativo bool) error {
	return m.Called(ctx, id, papel, ativo).Error(0)
}
func (m *RepositoryMock_t) ContarAdminsAtivos(ctx context.Context) (int, error) {
	a := m.Called(ctx)
	return a.Int(0), a.Error(1)
}
func (m *RepositoryMock_t) MarcarAcesso(ctx context.Context, id int64, em time.Time) error {
	return m.Called(ctx, id, em).Error(0)
}
func (m *RepositoryMock_t) CriarConvite(ctx context.Context, usuarioID int64, tokenHash []byte, expira, agora time.Time) error {
	return m.Called(ctx, usuarioID, tokenHash, expira, agora).Error(0)
}
func (m *RepositoryMock_t) UsarConvite(ctx context.Context, tokenHash []byte, senhaHash string, agora time.Time) (int64, error) {
	a := m.Called(ctx, tokenHash, senhaHash, agora)
	return a.Get(0).(int64), a.Error(1)
}

func (m *RepositoryMock_t) UltimoCodigoEm(ctx context.Context, usuarioID int64) (time.Time, error) {
	a := m.Called(ctx, usuarioID)
	return a.Get(0).(time.Time), a.Error(1)
}
func (m *RepositoryMock_t) CriarCodigo(ctx context.Context, usuarioID int64, codigoHash []byte, expira, agora time.Time) error {
	return m.Called(ctx, usuarioID, codigoHash, expira, agora).Error(0)
}
func (m *RepositoryMock_t) CodigoAtivo(ctx context.Context, usuarioID int64, agora time.Time) (usuario_repo.CodigoRow_t, error) {
	a := m.Called(ctx, usuarioID, agora)
	return a.Get(0).(usuario_repo.CodigoRow_t), a.Error(1)
}
func (m *RepositoryMock_t) SomarTentativa(ctx context.Context, codigoID int64) error {
	return m.Called(ctx, codigoID).Error(0)
}
func (m *RepositoryMock_t) UsarCodigo(ctx context.Context, codigoID, usuarioID int64, senhaHash string, agora time.Time) error {
	return m.Called(ctx, codigoID, usuarioID, senhaHash, agora).Error(0)
}

type UsuarioSuite struct {
	suite.Suite
	repo  *RepositoryMock_t
	svc   usuario.Service_I
	agora time.Time
	ctx   context.Context
}

func (s *UsuarioSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.agora = time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC)
	s.svc = usuario.NewUsuarioCom(s.repo, func() time.Time { return s.agora }, 72*time.Hour, chaveTeste)
	s.ctx = context.Background()
}

var chaveTeste = []byte("chave-de-teste-com-32-bytes-!!!!")

func linha(id int64, papel string, ativo bool, senha string) usuario_repo.UsuarioRow_t {
	r := usuario_repo.UsuarioRow_t{ID: id, Nome: "Pessoa", Email: "pessoa@gobrax.com", Papel: papel, Ativo: ativo}
	if senha != "" {
		h, _ := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.MinCost)
		r.SenhaHash = sql.NullString{String: string(h), Valid: true}
	}
	return r
}

func (s *UsuarioSuite) TestCriarGeraConviteComPrazoEGuardaSoOHash() {
	s.repo.On("Criar", s.ctx, "Ana Lima", "ana@gobrax.com", usuario.PapelLeitor).Return(int64(5), nil)
	s.repo.On("BuscarPorID", s.ctx, int64(5)).Return(linha(5, usuario.PapelLeitor, true, ""), nil)
	var hashGravado []byte
	s.repo.On("CriarConvite", s.ctx, int64(5), mock.Anything, s.agora.Add(72*time.Hour), s.agora).
		Run(func(a mock.Arguments) { hashGravado = a.Get(2).([]byte) }).Return(nil)

	u, conv, err := s.svc.Criar(s.ctx, " Ana Lima ", " ANA@gobrax.com ", usuario.PapelLeitor)

	s.Require().NoError(err)
	s.False(u.TemSenha)
	s.Len(conv.Token, 43)
	s.Equal(cripto.HashToken(conv.Token), hashGravado, "no banco vai só o hash do link")
	s.Equal(s.agora.Add(72*time.Hour), conv.ExpiraEm)
}

func (s *UsuarioSuite) TestCriarValidaCadastro() {
	_, _, err := s.svc.Criar(s.ctx, "", "a@gobrax.com", usuario.PapelLeitor)
	s.ErrorIs(err, usuario.ErrNome)
	_, _, err = s.svc.Criar(s.ctx, "Ana", "Ana <a@gobrax.com>", usuario.PapelLeitor)
	s.ErrorIs(err, usuario.ErrEmail)
	_, _, err = s.svc.Criar(s.ctx, "Ana", "a@gobrax", usuario.PapelLeitor)
	s.ErrorIs(err, usuario.ErrEmail)
	_, _, err = s.svc.Criar(s.ctx, "Ana", "a@gobrax.com", "dono")
	s.ErrorIs(err, usuario.ErrPapel)
}

func (s *UsuarioSuite) TestEmailRepetido() {
	s.repo.On("Criar", s.ctx, "Ana", "a@gobrax.com", usuario.PapelAdmin).Return(int64(0), usuario_repo.ErrRepetido)
	_, _, err := s.svc.Criar(s.ctx, "Ana", "a@gobrax.com", usuario.PapelAdmin)
	s.ErrorIs(err, usuario.ErrEmailRepetido)
}

func (s *UsuarioSuite) TestAceitarConviteGravaHashBcrypt() {
	var senhaHash string
	s.repo.On("UsarConvite", s.ctx, cripto.HashToken("tk"), mock.Anything, s.agora).
		Run(func(a mock.Arguments) { senhaHash = a.String(2) }).Return(int64(5), nil)
	s.repo.On("BuscarPorID", s.ctx, int64(5)).Return(linha(5, usuario.PapelLeitor, true, "x"), nil)

	_, err := s.svc.AceitarConvite(s.ctx, "tk", "senhaForte123")

	s.Require().NoError(err)
	s.NotContains(senhaHash, "senhaForte123")
	s.NoError(bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte("senhaForte123")))
}

func (s *UsuarioSuite) TestSenhaFracaNemChegaAoBanco() {
	for _, senha := range []string{"curta1", "semnumeroaqui", "1234567890"} {
		_, err := s.svc.AceitarConvite(s.ctx, "tk", senha)
		s.ErrorIs(err, usuario.ErrSenhaFraca, senha)
	}
	s.repo.AssertNotCalled(s.T(), "UsarConvite", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *UsuarioSuite) TestConviteUsadoOuVencido() {
	s.repo.On("UsarConvite", s.ctx, mock.Anything, mock.Anything, s.agora).Return(int64(0), usuario_repo.ErrConviteInvalido)
	_, err := s.svc.AceitarConvite(s.ctx, "tk", "senhaForte123")
	s.ErrorIs(err, usuario.ErrConviteInvalido)
}

func (s *UsuarioSuite) TestAutenticar() {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelColaborador, true, "senhaForte123"), nil)
	s.repo.On("MarcarAcesso", s.ctx, int64(3), s.agora).Return(nil)

	u, err := s.svc.Autenticar(s.ctx, " Pessoa@Gobrax.com ", "senhaForte123")
	s.Require().NoError(err)
	s.Equal(usuario.PapelColaborador, u.Papel)

	_, err = s.svc.Autenticar(s.ctx, "pessoa@gobrax.com", "outraSenha123")
	s.ErrorIs(err, usuario.ErrLogin)
}

func (s *UsuarioSuite) TestEmailInexistenteDaAMesmaMensagem() {
	s.repo.On("BuscarPorEmail", s.ctx, "ninguem@gobrax.com").Return(usuario_repo.UsuarioRow_t{}, usuario_repo.ErrNaoEncontrado)
	_, err := s.svc.Autenticar(s.ctx, "ninguem@gobrax.com", "qualquer123")
	s.ErrorIs(err, usuario.ErrLogin)
}

func (s *UsuarioSuite) TestConviteNaoAceitoNaoEntra() {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelLeitor, true, ""), nil)
	_, err := s.svc.Autenticar(s.ctx, "pessoa@gobrax.com", "qualquer123")
	s.ErrorIs(err, usuario.ErrLogin)
}

func (s *UsuarioSuite) TestDesativadoNaoEntra() {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelLeitor, false, "senhaForte123"), nil)
	_, err := s.svc.Autenticar(s.ctx, "pessoa@gobrax.com", "senhaForte123")
	s.ErrorIs(err, usuario.ErrInativo)
}

func (s *UsuarioSuite) TestNaoTiraOUltimoAdmin() {
	s.repo.On("BuscarPorID", s.ctx, int64(1)).Return(linha(1, usuario.PapelAdmin, true, "x"), nil)
	s.repo.On("ContarAdminsAtivos", s.ctx).Return(1, nil)

	_, err := s.svc.Alterar(s.ctx, 1, usuario.PapelLeitor, true)
	s.ErrorIs(err, usuario.ErrUltimoAdmin)
	_, err = s.svc.Alterar(s.ctx, 1, usuario.PapelAdmin, false)
	s.ErrorIs(err, usuario.ErrUltimoAdmin)
	s.repo.AssertNotCalled(s.T(), "Alterar", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *UsuarioSuite) TestTrocaPerfilComOutroAdmin() {
	s.repo.On("BuscarPorID", s.ctx, int64(1)).Return(linha(1, usuario.PapelAdmin, true, "x"), nil).Once()
	s.repo.On("ContarAdminsAtivos", s.ctx).Return(2, nil)
	s.repo.On("Alterar", s.ctx, int64(1), usuario.PapelColaborador, true).Return(nil)
	s.repo.On("BuscarPorID", s.ctx, int64(1)).Return(linha(1, usuario.PapelColaborador, true, "x"), nil)

	u, err := s.svc.Alterar(s.ctx, 1, usuario.PapelColaborador, true)
	s.Require().NoError(err)
	s.Equal(usuario.PapelColaborador, u.Papel)
}

func (s *UsuarioSuite) TestNovoConviteParaDesativadoERecusado() {
	s.repo.On("BuscarPorID", s.ctx, int64(9)).Return(linha(9, usuario.PapelLeitor, false, ""), nil)
	_, err := s.svc.NovoConvite(s.ctx, 9)
	s.ErrorIs(err, usuario.ErrInativo)
}

func TestUsuarioSuite(t *testing.T) { suite.Run(t, new(UsuarioSuite)) }

// ---------- código de acesso por e-mail ----------

func (s *UsuarioSuite) TestSolicitarCodigoGeraSeisDigitosEGuardaSoOHMAC() {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelLeitor, true, ""), nil)
	s.repo.On("UltimoCodigoEm", s.ctx, int64(3)).Return(time.Time{}, nil)
	var hash []byte
	s.repo.On("CriarCodigo", s.ctx, int64(3), mock.Anything, s.agora.Add(15*time.Minute), s.agora).
		Run(func(a mock.Arguments) { hash = a.Get(2).([]byte) }).Return(nil)

	c, err := s.svc.SolicitarCodigo(s.ctx, " Pessoa@gobrax.com ")

	s.Require().NoError(err)
	s.Regexp(`^\d{6}$`, c.Codigo)
	s.Len(hash, 32)
	s.NotContains(string(hash), c.Codigo)
}

func (s *UsuarioSuite) TestSolicitarCodigoDeNovoAntesDeUmMinuto() {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelLeitor, true, ""), nil)
	s.repo.On("UltimoCodigoEm", s.ctx, int64(3)).Return(s.agora.Add(-30*time.Second), nil)
	_, err := s.svc.SolicitarCodigo(s.ctx, "pessoa@gobrax.com")
	s.ErrorIs(err, usuario.ErrAguarde)
	s.repo.AssertNotCalled(s.T(), "CriarCodigo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *UsuarioSuite) TestSolicitarCodigoParaEmailDesconhecido() {
	s.repo.On("BuscarPorEmail", s.ctx, "ninguem@gobrax.com").Return(usuario_repo.UsuarioRow_t{}, usuario_repo.ErrNaoEncontrado)
	_, err := s.svc.SolicitarCodigo(s.ctx, "ninguem@gobrax.com")
	s.ErrorIs(err, usuario.ErrNaoEncontrado)
}

// codigoGravado pede um código e devolve o número e o HMAC que iria para o banco
func (s *UsuarioSuite) codigoGravado() (string, []byte) {
	s.repo.On("BuscarPorEmail", s.ctx, "pessoa@gobrax.com").Return(linha(3, usuario.PapelLeitor, true, ""), nil)
	s.repo.On("UltimoCodigoEm", s.ctx, int64(3)).Return(time.Time{}, nil)
	var hash []byte
	s.repo.On("CriarCodigo", s.ctx, int64(3), mock.Anything, mock.Anything, mock.Anything).Run(func(a mock.Arguments) { hash = a.Get(2).([]byte) }).Return(nil)
	c, err := s.svc.SolicitarCodigo(s.ctx, "pessoa@gobrax.com")
	s.Require().NoError(err)
	return c.Codigo, hash
}

func (s *UsuarioSuite) TestUsarCodigoCertoGravaASenha() {
	codigo, hash := s.codigoGravado()
	s.repo.On("CodigoAtivo", s.ctx, int64(3), s.agora).Return(usuario_repo.CodigoRow_t{ID: 9, CodigoHash: hash}, nil)
	var senhaHash string
	s.repo.On("UsarCodigo", s.ctx, int64(9), int64(3), mock.Anything, s.agora).Run(func(a mock.Arguments) { senhaHash = a.String(3) }).Return(nil)
	s.repo.On("MarcarAcesso", s.ctx, int64(3), s.agora).Return(nil)
	s.repo.On("BuscarPorID", s.ctx, int64(3)).Return(linha(3, usuario.PapelLeitor, true, "x"), nil)

	_, err := s.svc.UsarCodigo(s.ctx, "pessoa@gobrax.com", codigo[:3]+" "+codigo[3:], "senhaForte123")

	s.Require().NoError(err)
	s.NoError(bcrypt.CompareHashAndPassword([]byte(senhaHash), []byte("senhaForte123")))
}

func (s *UsuarioSuite) TestCodigoErradoSomaTentativa() {
	codigo, hash := s.codigoGravado()
	errado := "000000"
	if codigo == errado {
		errado = "111111"
	}
	s.repo.On("CodigoAtivo", s.ctx, int64(3), s.agora).Return(usuario_repo.CodigoRow_t{ID: 9, CodigoHash: hash, Tentativas: 1}, nil)
	s.repo.On("SomarTentativa", s.ctx, int64(9)).Return(nil)

	_, err := s.svc.UsarCodigo(s.ctx, "pessoa@gobrax.com", errado, "senhaForte123")

	s.ErrorIs(err, usuario.ErrCodigoInvalido)
	s.repo.AssertCalled(s.T(), "SomarTentativa", s.ctx, int64(9))
	s.repo.AssertNotCalled(s.T(), "UsarCodigo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *UsuarioSuite) TestCincoTentativasBloqueiamOCodigo() {
	codigo, hash := s.codigoGravado()
	s.repo.On("CodigoAtivo", s.ctx, int64(3), s.agora).Return(usuario_repo.CodigoRow_t{ID: 9, CodigoHash: hash, Tentativas: 5}, nil)
	_, err := s.svc.UsarCodigo(s.ctx, "pessoa@gobrax.com", codigo, "senhaForte123")
	s.ErrorIs(err, usuario.ErrCodigoInvalido, "nem o código certo vale depois de 5 erros")
}

func (s *UsuarioSuite) TestCodigoComEmailDesconhecidoDaAMesmaMensagem() {
	s.repo.On("BuscarPorEmail", s.ctx, "ninguem@gobrax.com").Return(usuario_repo.UsuarioRow_t{}, usuario_repo.ErrNaoEncontrado)
	_, err := s.svc.UsarCodigo(s.ctx, "ninguem@gobrax.com", "123456", "senhaForte123")
	s.ErrorIs(err, usuario.ErrCodigoInvalido)
}

func (s *UsuarioSuite) TestCodigoForaDoFormato() {
	_, err := s.svc.UsarCodigo(s.ctx, "pessoa@gobrax.com", "12a456", "senhaForte123")
	s.ErrorIs(err, usuario.ErrCodigoInvalido)
	s.repo.AssertNotCalled(s.T(), "BuscarPorEmail", mock.Anything, mock.Anything)
}
