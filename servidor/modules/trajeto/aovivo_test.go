package trajeto_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/trajeto"
)

type AoVivoSuite struct {
	suite.Suite
	api   *PosicoesMock_t
	agora time.Time
	svc   trajeto.Vivo_I
}

func (s *AoVivoSuite) SetupTest() {
	s.api = new(PosicoesMock_t)
	s.agora = time.Date(2026, 11, 5, 15, 0, 0, 0, time.UTC)
	s.svc = trajeto.NewAoVivoCom(s.api, func() time.Time { return s.agora })
}

// última posição das últimas 24 h e, quando recente, a última até 4 min antes dela
func (s *AoVivoSuite) posicoes(placa string, atual, anterior *gobraxapi.Posicao_t) {
	s.api.On("UltimaPosicao", mock.Anything, "tk", placa, s.agora.Add(-24*time.Hour), s.agora).Return(atual, nil).Once()
	if atual != nil {
		t1, _ := time.Parse("2006-01-02 15:04:05", atual.Data)
		s.api.On("UltimaPosicao", mock.Anything, "tk", placa, s.agora.Add(-24*time.Hour), t1.Add(-4*time.Minute)).Return(anterior, nil).Maybe()
	}
}

func (s *AoVivoSuite) TestEmMovimentoPelaDistanciaEntreAsDuasPosicoes() {
	// ~0,09° de latitude ≈ 10 km em 5 min = ~120 km/h
	s.posicoes("ABC1D23", pos("2026-11-05 14:59:00", -25.40, -49.2), pos("2026-11-05 14:54:00", -25.49, -49.2))

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"abc-1d23"})

	s.Require().NoError(err)
	s.True(r.Completo)
	s.Require().Len(r.Veiculos, 1)
	v := r.Veiculos[0]
	s.Equal(trajeto.StatusMovendo, v.Status)
	s.Equal(1, *v.MinAtras)
	s.InDelta(120, *v.Velocidade, 2)
}

func (s *AoVivoSuite) TestParadoQuandoNaoSaiDoLugar() {
	s.posicoes("ABC1D23", pos("2026-11-05 14:58:00", -25.4, -49.2), pos("2026-11-05 14:53:30", -25.4, -49.2))

	r, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"ABC1D23"})

	s.Require().NoError(err)
	s.Equal(trajeto.StatusParado, r.Veiculos[0].Status)
	s.Equal(0.0, *r.Veiculos[0].Velocidade)
}

func (s *AoVivoSuite) TestSemSinalComPosicaoAntigaNaoPedeAAnterior() {
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

func (s *AoVivoSuite) TestSemPlacaValida() {
	_, err := s.svc.Situacao(context.Background(), 1, "tk", []string{"", "!!"})
	s.ErrorIs(err, trajeto.ErrPlacas)
}

func TestAoVivoSuite(t *testing.T) { suite.Run(t, new(AoVivoSuite)) }
