package trajeto_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/trajeto"
	"analisefrota/modules/trajeto/trajeto_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Listar(ctx context.Context, clienteID int64, placa string, ini, fim time.Time) ([]trajeto_repo.PosicaoRow_t, error) {
	a := m.Called(ctx, clienteID, placa, ini, fim)
	return a.Get(0).([]trajeto_repo.PosicaoRow_t), a.Error(1)
}
func (m *RepositoryMock_t) Salvar(ctx context.Context, row trajeto_repo.PosicaoRow_t) error {
	return m.Called(ctx, row).Error(0)
}

type PosicoesMock_t struct{ mock.Mock }

func (m *PosicoesMock_t) UltimaPosicao(ctx context.Context, token, placa string, ini, fim time.Time) (*gobraxapi.Posicao_t, error) {
	a := m.Called(ctx, token, placa, ini, fim)
	p, _ := a.Get(0).(*gobraxapi.Posicao_t)
	return p, a.Error(1)
}

type TrajetoSuite struct {
	suite.Suite
	repo  *RepositoryMock_t
	api   *PosicoesMock_t
	agora time.Time
	svc   trajeto.Service_I
}

// 20/10/2026 em Brasília começa às 03:00 UTC
var inicioDia = time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC)

func (s *TrajetoSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.api = new(PosicoesMock_t)
	s.agora = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	s.svc = trajeto.NewTrajetoCom(s.repo, s.api, func() time.Time { return s.agora })
}

func pos(data string, lat, lon float64) *gobraxapi.Posicao_t {
	return &gobraxapi.Posicao_t{Data: data, Lat: lat, Lon: lon}
}

func (s *TrajetoSuite) TestUmaHoraViraDozeJanelasEmUTCSemRepetirAMesmaColeta() {
	s.repo.On("Listar", mock.Anything, int64(3), "ABC1D23", inicioDia, inicioDia.Add(time.Hour)).Return([]trajeto_repo.PosicaoRow_t{}, nil)
	s.repo.On("Salvar", mock.Anything, mock.Anything).Return(nil)
	// a primeira janela pede de 03:00:00 a 03:04:59 UTC (meia-noite em Brasília)
	s.api.On("UltimaPosicao", mock.Anything, "tk", "ABC1D23", inicioDia, inicioDia.Add(5*time.Minute-time.Second)).Return(pos("2026-10-20 03:04:30", -25.4, -49.2), nil).Once()
	// parado na segunda janela: a API repete a mesma coleta, que não vira um ponto novo
	s.api.On("UltimaPosicao", mock.Anything, "tk", "ABC1D23", inicioDia.Add(5*time.Minute), mock.Anything).Return(pos("2026-10-20 03:04:30", -25.4, -49.2), nil).Once()
	s.api.On("UltimaPosicao", mock.Anything, "tk", "ABC1D23", mock.Anything, mock.Anything).Return(nil, nil)

	t, err := s.svc.Amostrar(context.Background(), 3, "tk", "abc-1d23", "2026-10-20", 0, 1)

	s.Require().NoError(err)
	s.Equal(12, t.Janelas)
	s.True(t.Completo)
	s.Require().Len(t.Pontos, 1)
	s.Equal("00:04:30", t.Pontos[0].Hora)
	s.Equal(-25.4, t.Pontos[0].Lat)
	s.repo.AssertNumberOfCalls(s.T(), "Salvar", 12)
}

func (s *TrajetoSuite) TestPontosEmOrdemDeData() {
	s.repo.On("Listar", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]trajeto_repo.PosicaoRow_t{}, nil)
	s.repo.On("Salvar", mock.Anything, mock.Anything).Return(nil)
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, inicioDia, mock.Anything).Return(pos("2026-10-20 03:04:00", 1, 1), nil).Once()
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, inicioDia.Add(5*time.Minute), mock.Anything).Return(pos("2026-10-20 03:01:00", 2, 2), nil).Once()
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	t, err := s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 0, 1)

	s.Require().NoError(err)
	s.Require().Len(t.Pontos, 2)
	s.Equal("2026-10-20 03:01:00", t.Pontos[0].Data)
	s.Equal("2026-10-20 03:04:00", t.Pontos[1].Data)
}

func (s *TrajetoSuite) TestJanelasNoFuturoFicamDeFora() {
	// hoje em Brasília, 00:12: só as janelas de 00:00, 00:05 e 00:10 já começaram
	s.agora = inicioDia.Add(12 * time.Minute)
	s.repo.On("Listar", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]trajeto_repo.PosicaoRow_t{}, nil)
	s.repo.On("Salvar", mock.Anything, mock.Anything).Return(nil)
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	t, err := s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 0, 24)

	s.Require().NoError(err)
	s.Equal(3, t.Janelas)
	s.api.AssertNumberOfCalls(s.T(), "UltimaPosicao", 3)
}

func (s *TrajetoSuite) TestDeEAteRecortamAsHoras() {
	// das 14h às 16h de Brasília = 17:00 a 19:00 UTC, 24 janelas
	ini := inicioDia.Add(14 * time.Hour)
	s.repo.On("Listar", mock.Anything, int64(3), "ABC1D23", ini, ini.Add(2*time.Hour)).Return([]trajeto_repo.PosicaoRow_t{}, nil)
	s.repo.On("Salvar", mock.Anything, mock.Anything).Return(nil)
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	t, err := s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 14, 16)

	s.Require().NoError(err)
	s.Equal(24, t.Janelas)
	s.api.AssertCalled(s.T(), "UltimaPosicao", mock.Anything, "tk", "ABC1D23", ini, ini.Add(5*time.Minute-time.Second))
}

func (s *TrajetoSuite) TestJanelaDefinitivaVemDoBancoEJanelaRecenteEPedidaDeNovo() {
	lat, lon := -23.5, -46.6
	salvas := []trajeto_repo.PosicaoRow_t{
		// buscada 1 h depois de acabar: definitiva, não chama a API
		{ClienteID: 3, Placa: "ABC1D23", JanelaIni: inicioDia, Data: "2026-10-20 03:02:00", Lat: &lat, Lon: &lon, BuscadoEm: inicioDia.Add(time.Hour)},
		// buscada 6 min depois de começar: ainda podia mudar, pede de novo
		{ClienteID: 3, Placa: "ABC1D23", JanelaIni: inicioDia.Add(5 * time.Minute), Vazio: true, BuscadoEm: inicioDia.Add(11 * time.Minute)},
	}
	s.repo.On("Listar", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(salvas, nil)
	s.repo.On("Salvar", mock.Anything, mock.Anything).Return(nil)
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	t, err := s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 0, 1)

	s.Require().NoError(err)
	s.Require().Len(t.Pontos, 1)
	s.Equal("2026-10-20 03:02:00", t.Pontos[0].Data)
	s.api.AssertNotCalled(s.T(), "UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, inicioDia, mock.Anything)
	s.api.AssertCalled(s.T(), "UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, inicioDia.Add(5*time.Minute), mock.Anything)
	s.api.AssertNumberOfCalls(s.T(), "UltimaPosicao", 11)
}

func (s *TrajetoSuite) TestVeiculoNaoEncontradoVolta() {
	s.repo.On("Listar", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]trajeto_repo.PosicaoRow_t{}, nil)
	s.api.On("UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, gobraxapi.ErrVeiculoNaoEncontrado)

	_, err := s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 0, 1)

	s.ErrorIs(err, gobraxapi.ErrVeiculoNaoEncontrado)
}

func (s *TrajetoSuite) TestValidacoes() {
	_, err := s.svc.Amostrar(context.Background(), 3, "tk", "AB-1", "2026-10-20", 0, 24)
	s.ErrorIs(err, trajeto.ErrPlaca)
	_, err = s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "20/10/2026", 0, 24)
	s.ErrorIs(err, trajeto.ErrDia)
	_, err = s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-23", 0, 24)
	s.ErrorIs(err, trajeto.ErrDiaFuturo)
	_, err = s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-07-01", 0, 24)
	s.ErrorIs(err, trajeto.ErrDiaAntigo)
	_, err = s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 10, 10)
	s.ErrorIs(err, trajeto.ErrHoras)
	_, err = s.svc.Amostrar(context.Background(), 3, "tk", "ABC1D23", "2026-10-20", 0, 25)
	s.ErrorIs(err, trajeto.ErrHoras)
	s.api.AssertNotCalled(s.T(), "UltimaPosicao", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTrajetoSuite(t *testing.T) { suite.Run(t, new(TrajetoSuite)) }
