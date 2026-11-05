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

type RastroMock_t struct{ mock.Mock }

func (m *RastroMock_t) SalvarRastro(ctx context.Context, row trajeto_repo.RastroRow_t) error {
	return m.Called(ctx, row).Error(0)
}
func (m *RastroMock_t) ListarRastro(ctx context.Context, clienteID int64, placas []string, desde time.Time) ([]trajeto_repo.RastroRow_t, error) {
	a := m.Called(ctx, clienteID, placas, desde)
	return a.Get(0).([]trajeto_repo.RastroRow_t), a.Error(1)
}
func (m *RastroMock_t) LimparRastro(ctx context.Context, antes time.Time) (int64, error) {
	a := m.Called(ctx, antes)
	return int64(a.Int(0)), a.Error(1)
}

type AoVivoSuite struct {
	suite.Suite
	api   *PosicoesMock_t
	repo  *RastroMock_t
	agora time.Time
	svc   trajeto.Vivo_I
}

func (s *AoVivoSuite) SetupTest() {
	s.api = new(PosicoesMock_t)
	s.repo = new(RastroMock_t)
	s.repo.On("SalvarRastro", mock.Anything, mock.Anything).Return(nil)
	s.agora = time.Date(2026, 11, 5, 15, 0, 0, 0, time.UTC)
	s.svc = trajeto.NewAoVivoCom(s.api, s.repo, func() time.Time { return s.agora })
}

// última posição das últimas 24 h e, quando recente, a última entre 5 min e 90 s antes dela
func (s *AoVivoSuite) posicoes(placa string, atual, anterior *gobraxapi.Posicao_t) {
	s.api.On("UltimaPosicao", mock.Anything, "tk", placa, s.agora.Add(-24*time.Hour), s.agora).Return(atual, nil).Once()
	if atual != nil {
		t1, _ := time.Parse("2006-01-02 15:04:05", atual.Data)
		s.api.On("UltimaPosicao", mock.Anything, "tk", placa, t1.Add(-5*time.Minute), t1.Add(-90*time.Second)).Return(anterior, nil).Maybe()
	}
}

func (s *AoVivoSuite) TestIgnicaoLigadaEmMovimento() {
	// posições a cada minuto; ~0,036° de latitude ≈ 4 km em 2 min = ~120 km/h
	s.posicoes("ABC1D23", pos("2026-11-05 14:59:30", -25.400, -49.2), pos("2026-11-05 14:57:30", -25.436, -49.2))

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"abc-1d23"})

	s.Require().NoError(err)
	s.True(r.Completo)
	s.Require().Len(r.Veiculos, 1)
	v := r.Veiculos[0]
	s.Equal(trajeto.StatusMovendo, v.Status)
	s.Equal(0, *v.MinAtras)
	s.InDelta(120, *v.Velocidade, 3)
	s.repo.AssertCalled(s.T(), "SalvarRastro", mock.Anything, mock.MatchedBy(func(r trajeto_repo.RastroRow_t) bool {
		return r.Placa == "ABC1D23" && r.Status == trajeto.StatusMovendo && r.Data.Equal(time.Date(2026, 11, 5, 14, 59, 30, 0, time.UTC))
	}))
}

func (s *AoVivoSuite) TestMotorLigadoParado() {
	s.posicoes("ABC1D23", pos("2026-11-05 14:59:00", -25.4, -49.2), pos("2026-11-05 14:57:00", -25.4, -49.2))

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})

	s.Require().NoError(err)
	s.Equal(trajeto.StatusLigado, r.Veiculos[0].Status)
	s.Equal(0.0, *r.Veiculos[0].Velocidade)
}

func (s *AoVivoSuite) TestDesligadoSoTemOSinalDeDezEmDezMinutos() {
	// a última posição é recente, mas a anterior foi 10 min antes: nada entre 5 min e 90 s antes dela
	s.posicoes("ABC1D23", pos("2026-11-05 14:58:00", -25.4, -49.2), nil)

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})

	s.Require().NoError(err)
	s.Equal(trajeto.StatusDesligado, r.Veiculos[0].Status)
	s.Nil(r.Veiculos[0].Velocidade)
}

func (s *AoVivoSuite) TestDesligadoComPosicaoDeAlgunsMinutosNaoPedeAAnterior() {
	s.posicoes("ABC1D23", pos("2026-11-05 14:52:00", -25.4, -49.2), nil)

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})

	s.Require().NoError(err)
	s.Equal(trajeto.StatusDesligado, r.Veiculos[0].Status)
	s.api.AssertNumberOfCalls(s.T(), "UltimaPosicao", 1)
}

func (s *AoVivoSuite) TestSemSinalComPosicaoAntiga() {
	s.posicoes("ABC1D23", pos("2026-11-05 13:00:00", -25.4, -49.2), nil)

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})

	s.Require().NoError(err)
	s.Equal(trajeto.StatusSemSinal, r.Veiculos[0].Status)
	s.Equal(120, *r.Veiculos[0].MinAtras)
	s.api.AssertNumberOfCalls(s.T(), "UltimaPosicao", 1)
}

func (s *AoVivoSuite) TestPlacaDesconhecidaNaoDerrubaAsOutras() {
	s.api.On("UltimaPosicao", mock.Anything, "tk", "XYZ9Z99", mock.Anything, mock.Anything).Return(nil, gobraxapi.ErrVeiculoNaoEncontrado).Once()
	s.posicoes("ABC1D23", nil, nil)

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"XYZ9Z99", "ABC1D23"})

	s.Require().NoError(err)
	s.Require().Len(r.Veiculos, 2)
	s.Equal(trajeto.StatusNaoAchou, r.Veiculos[0].Status)
	s.Equal(trajeto.StatusSemSinal, r.Veiculos[1].Status)
	s.Nil(r.Veiculos[1].Lat)
	s.repo.AssertNotCalled(s.T(), "SalvarRastro", mock.Anything, mock.Anything)
}

func (s *AoVivoSuite) TestCacheEvitaChamarDeNovo() {
	s.posicoes("ABC1D23", pos("2026-11-05 13:00:00", -25.4, -49.2), nil)

	_, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})
	s.Require().NoError(err)
	s.agora = s.agora.Add(60 * time.Second)
	_, err = s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23", "abc1d23"})

	s.Require().NoError(err)
	s.api.AssertNumberOfCalls(s.T(), "UltimaPosicao", 1)
}

func (s *AoVivoSuite) TestRastroAgrupaPorPlacaEmUTC() {
	desde := time.Date(2026, 11, 5, 3, 0, 0, 0, time.UTC)
	s.repo.On("ListarRastro", mock.Anything, int64(1), []string{"ABC1D23"}, desde).Return([]trajeto_repo.RastroRow_t{
		{Placa: "ABC1D23", Data: time.Date(2026, 11, 5, 14, 0, 0, 0, time.UTC), Lat: 1, Lon: 2, Status: trajeto.StatusMovendo},
		{Placa: "ABC1D23", Data: time.Date(2026, 11, 5, 14, 5, 0, 0, time.UTC), Lat: 3, Lon: 4, Status: trajeto.StatusLigado},
	}, nil)

	r, err := s.svc.Rastro(context.Background(), 1, []string{"abc-1d23"}, desde)

	s.Require().NoError(err)
	s.Require().Len(r["ABC1D23"], 2)
	s.Equal("2026-11-05 14:05:00", r["ABC1D23"][1].Data)
	s.Equal(trajeto.StatusLigado, r["ABC1D23"][1].Status)
}

func (s *AoVivoSuite) TestSemPlacaValida() {
	_, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"", "!!"})
	s.ErrorIs(err, trajeto.ErrPlacas)
}

func TestAoVivoSuite(t *testing.T) { suite.Run(t, new(AoVivoSuite)) }
