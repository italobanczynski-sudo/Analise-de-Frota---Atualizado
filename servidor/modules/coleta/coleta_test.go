package coleta_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/cliente"
	"analisefrota/modules/coleta"
	"analisefrota/modules/coleta/coleta_repo"
	"analisefrota/modules/diario"
	"analisefrota/modules/odometro"
	"analisefrota/modules/premiacao"
)

// ---------- mocks ----------

type RepoMock_t struct{ mock.Mock }

func (m *RepoMock_t) Buscar(ctx context.Context, id int64) (coleta_repo.SituacaoRow_t, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(coleta_repo.SituacaoRow_t), a.Error(1)
}
func (m *RepoMock_t) MarcarInicio(ctx context.Context, id int64, em time.Time) error {
	return m.Called(ctx, id, em).Error(0)
}
func (m *RepoMock_t) MarcarFim(ctx context.Context, id int64, situacao, mensagem string, avisos []string, ultimoDia *string, em time.Time) error {
	return m.Called(ctx, id, situacao, mensagem, avisos, ultimoDia, em).Error(0)
}

type APIMock_t struct{ mock.Mock }

func (m *APIMock_t) Estatisticas(ctx context.Context, token string, ini, fim time.Time) ([]gobraxapi.EstatisticaRec_t, error) {
	a := m.Called(ctx, token, ini, fim)
	r, _ := a.Get(0).([]gobraxapi.EstatisticaRec_t)
	return r, a.Error(1)
}
func (m *APIMock_t) Performance(ctx context.Context, token string, ini, fim time.Time) ([]gobraxapi.PerformanceRec_t, error) {
	a := m.Called(ctx, token, ini, fim)
	r, _ := a.Get(0).([]gobraxapi.PerformanceRec_t)
	return r, a.Error(1)
}
func (m *APIMock_t) Odometro(ctx context.Context, token string, ini, fim time.Time) ([]gobraxapi.OdometroRec_t, error) {
	a := m.Called(ctx, token, ini, fim)
	r, _ := a.Get(0).([]gobraxapi.OdometroRec_t)
	return r, a.Error(1)
}
func (m *APIMock_t) Premiacao(ctx context.Context, token, mesIni, mesFim string) ([]gobraxapi.PremiacaoRec_t, error) {
	a := m.Called(ctx, token, mesIni, mesFim)
	r, _ := a.Get(0).([]gobraxapi.PremiacaoRec_t)
	return r, a.Error(1)
}

type ClientesMock_t struct{ mock.Mock }

func (m *ClientesMock_t) Listar(ctx context.Context) ([]cliente.Cliente_t, error) {
	a := m.Called(ctx)
	return a.Get(0).([]cliente.Cliente_t), a.Error(1)
}
func (m *ClientesMock_t) Buscar(ctx context.Context, id int64) (cliente.Cliente_t, error) {
	a := m.Called(ctx, id)
	return a.Get(0).(cliente.Cliente_t), a.Error(1)
}
func (m *ClientesMock_t) BuscarPorNome(ctx context.Context, nome string) (cliente.Cliente_t, error) {
	a := m.Called(ctx, nome)
	return a.Get(0).(cliente.Cliente_t), a.Error(1)
}
func (m *ClientesMock_t) Criar(ctx context.Context, nome, hubspotID string) (cliente.Cliente_t, error) {
	a := m.Called(ctx, nome, hubspotID)
	return a.Get(0).(cliente.Cliente_t), a.Error(1)
}
func (m *ClientesMock_t) SalvarToken(ctx context.Context, id int64, token string) error {
	return m.Called(ctx, id, token).Error(0)
}
func (m *ClientesMock_t) Token(ctx context.Context, id int64) (string, error) {
	a := m.Called(ctx, id)
	return a.String(0), a.Error(1)
}

type DiarioMock_t struct{ mock.Mock }

func (m *DiarioMock_t) MontarDia(clienteID int64, dia string, estat []gobraxapi.EstatisticaRec_t, perf []gobraxapi.PerformanceRec_t) []diario.Registro_t {
	return m.Called(clienteID, dia, estat, perf).Get(0).([]diario.Registro_t)
}
func (m *DiarioMock_t) SalvarDia(ctx context.Context, clienteID int64, dia string, regs []diario.Registro_t) error {
	return m.Called(ctx, clienteID, dia, regs).Error(0)
}
func (m *DiarioMock_t) Listar(ctx context.Context, clienteID int64, ini, fim string) ([]diario.Registro_t, error) {
	a := m.Called(ctx, clienteID, ini, fim)
	return a.Get(0).([]diario.Registro_t), a.Error(1)
}

type OdometroMock_t struct{ mock.Mock }

func (m *OdometroMock_t) Substituir(ctx context.Context, clienteID int64, recs []gobraxapi.OdometroRec_t) (int, error) {
	a := m.Called(ctx, clienteID, recs)
	return a.Int(0), a.Error(1)
}
func (m *OdometroMock_t) Listar(ctx context.Context, clienteID int64) ([]odometro.Odometro_t, error) {
	return nil, nil
}

type PremiacaoMock_t struct{ mock.Mock }

func (m *PremiacaoMock_t) Salvar(ctx context.Context, clienteID int64, recs []gobraxapi.PremiacaoRec_t) ([]string, error) {
	a := m.Called(ctx, clienteID, recs)
	return a.Get(0).([]string), a.Error(1)
}
func (m *PremiacaoMock_t) Listar(ctx context.Context, clienteID int64, mes string) ([]premiacao.Premiacao_t, error) {
	return nil, nil
}

// ---------- suite ----------

type ColetaSuite struct {
	suite.Suite
	repo   *RepoMock_t
	api    *APIMock_t
	cli    *ClientesMock_t
	diario *DiarioMock_t
	odo    *OdometroMock_t
	prem   *PremiacaoMock_t
	svc    coleta.Service_I
	agora  time.Time
	ctx    context.Context
}

func (s *ColetaSuite) SetupTest() {
	s.repo, s.api, s.cli = new(RepoMock_t), new(APIMock_t), new(ClientesMock_t)
	s.diario, s.odo, s.prem = new(DiarioMock_t), new(OdometroMock_t), new(PremiacaoMock_t)
	s.agora = time.Date(2026, 10, 24, 15, 0, 0, 0, time.UTC) // 12h de 24/10 em Brasília
	s.ctx = context.Background()
	s.svc = coleta.NewColetaCom(coleta.Deps_t{
		Repo: s.repo, API: s.api, Clientes: s.cli, Diario: s.diario, Odometro: s.odo, Premiacao: s.prem,
		Agora: func() time.Time { return s.agora }, Dias: 3,
	})
	s.cli.On("Token", mock.Anything, int64(1)).Return("tk", nil).Maybe()
	s.repo.On("MarcarInicio", mock.Anything, int64(1), mock.Anything).Return(nil).Maybe()
	s.diario.On("MontarDia", int64(1), mock.Anything, mock.Anything, mock.Anything).Return([]diario.Registro_t{{Placa: "ABC1D23"}}).Maybe()
	s.diario.On("SalvarDia", mock.Anything, int64(1), mock.Anything, mock.Anything).Return(nil).Maybe()
	s.api.On("Odometro", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.OdometroRec_t{}, nil).Maybe()
	s.odo.On("Substituir", mock.Anything, int64(1), mock.Anything).Return(15, nil).Maybe()
	s.api.On("Premiacao", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.PremiacaoRec_t{}, nil).Maybe()
	s.prem.On("Salvar", mock.Anything, int64(1), mock.Anything).Return([]string{"2026-10"}, nil).Maybe()
}

func (s *ColetaSuite) diaUTC(dia string) time.Time {
	t, _ := time.Parse("2006-01-02", dia)
	return t.Add(3 * time.Hour)
}

func (s *ColetaSuite) TestPrimeiraBuscaPegaOsUltimosDiasEGuardaHoje() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{}, nil)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.EstatisticaRec_t{}, nil)
	s.api.On("Performance", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.PerformanceRec_t{}, nil)
	var ultimo *string
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoOk, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(a mock.Arguments) { ultimo = a.Get(5).(*string) }).Return(nil)

	res, err := s.svc.Executar(s.ctx, 1)

	s.Require().NoError(err)
	s.Equal(3, res.DiasComDados)
	s.Equal(15, res.Odometros)
	s.api.AssertCalled(s.T(), "Estatisticas", mock.Anything, "tk", s.diaUTC("2026-10-22"), mock.Anything)
	s.api.AssertNotCalled(s.T(), "Estatisticas", mock.Anything, "tk", s.diaUTC("2026-10-21"), mock.Anything)
	s.Require().NotNil(ultimo)
	s.Equal("2026-10-24", *ultimo)
	s.api.AssertCalled(s.T(), "Premiacao", mock.Anything, "tk", "10-2026", "10-2026")
}

func (s *ColetaSuite) TestBuscaSeguinteRefazOUltimoDia() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{UltimoDia: sql.NullString{String: "2026-10-24", Valid: true}}, nil)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.EstatisticaRec_t{}, nil)
	s.api.On("Performance", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.PerformanceRec_t{}, nil)
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoOk, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, err := s.svc.Executar(s.ctx, 1)

	s.Require().NoError(err)
	s.Equal(2, res.DiasComDados, "refaz 23/10 e busca 24/10")
	s.api.AssertNotCalled(s.T(), "Estatisticas", mock.Anything, "tk", s.diaUTC("2026-10-22"), mock.Anything)
}

func (s *ColetaSuite) TestTokenInvalidoParaTudoERegistraErro() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{}, nil)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return(nil, gobraxapi.ErrTokenInvalido)
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoErro, gobraxapi.ErrTokenInvalido.Error(), mock.Anything, (*string)(nil), mock.Anything).Return(nil)

	_, err := s.svc.Executar(s.ctx, 1)

	s.ErrorIs(err, gobraxapi.ErrTokenInvalido)
	s.repo.AssertExpectations(s.T())
	s.diario.AssertNotCalled(s.T(), "SalvarDia", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *ColetaSuite) TestFalhaPontualNaoApagaODiaEVoltaNaProximaBusca() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{}, nil)
	falha := gobraxapi.ErroAPI_t{Status: 500}
	s.api.On("Estatisticas", mock.Anything, "tk", s.diaUTC("2026-10-23"), mock.Anything).Return(nil, falha)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.EstatisticaRec_t{}, nil)
	s.api.On("Performance", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.PerformanceRec_t{}, nil)
	var ultimo *string
	var avisos []string
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoOk, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(a mock.Arguments) { avisos = a.Get(4).([]string); ultimo = a.Get(5).(*string) }).Return(nil)

	res, err := s.svc.Executar(s.ctx, 1)

	s.Require().NoError(err)
	s.Equal(2, res.DiasComDados)
	s.diario.AssertNotCalled(s.T(), "SalvarDia", mock.Anything, int64(1), "2026-10-23", mock.Anything)
	s.Equal("2026-10-22", *ultimo, "o dia perdido volta na próxima busca")
	s.Len(avisos, 1)
}

func (s *ColetaSuite) TestTresFalhasDaPerformanceDesligamAPerformance() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{}, nil)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.EstatisticaRec_t{}, nil)
	s.api.On("Performance", mock.Anything, "tk", mock.Anything, mock.Anything).Return(nil, gobraxapi.ErrSemAcesso)
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoOk, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, err := s.svc.Executar(s.ctx, 1)

	s.Require().NoError(err)
	s.False(res.Performance)
	s.Equal(3, res.DiasComDados, "as estatísticas entram mesmo sem performance")
	s.Contains(res.Avisos[0], "performance parou depois de 3 falhas")
}

func (s *ColetaSuite) TestOdometroEPremiacaoQueFalhamViramAviso() {
	s.api = new(APIMock_t)
	s.svc = coleta.NewColetaCom(coleta.Deps_t{Repo: s.repo, API: s.api, Clientes: s.cli, Diario: s.diario, Odometro: s.odo, Premiacao: s.prem,
		Agora: func() time.Time { return s.agora }, Dias: 1})
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{}, nil)
	s.api.On("Estatisticas", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.EstatisticaRec_t{}, nil)
	s.api.On("Performance", mock.Anything, "tk", mock.Anything, mock.Anything).Return([]gobraxapi.PerformanceRec_t{}, nil)
	s.api.On("Odometro", mock.Anything, "tk", mock.Anything, mock.Anything).Return(nil, errors.New("a API respondeu 500"))
	s.api.On("Premiacao", mock.Anything, "tk", mock.Anything, mock.Anything).Return(nil, gobraxapi.ErroAPI_t{Status: 400, Descricao: "Datas fornecidas inválidas"})
	s.repo.On("MarcarFim", mock.Anything, int64(1), coleta.SituacaoOk, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	res, err := s.svc.Executar(s.ctx, 1)

	s.Require().NoError(err)
	s.Equal([]string{"odômetro não veio: a API respondeu 500", "premiação não veio: a API respondeu 400 (Datas fornecidas inválidas)"}, res.Avisos)
}

func (s *ColetaSuite) TestSituacaoDeQuemNuncaFoiBuscado() {
	s.repo.On("Buscar", mock.Anything, int64(1)).Return(coleta_repo.SituacaoRow_t{ClienteID: 1}, nil)
	sit, err := s.svc.Situacao(s.ctx, 1)
	s.Require().NoError(err)
	s.Equal(coleta.SituacaoNunca, sit.Situacao)
	s.Empty(sit.UltimoDia)
	s.NotNil(sit.Avisos)
}

func TestColetaSuite(t *testing.T) { suite.Run(t, new(ColetaSuite)) }
