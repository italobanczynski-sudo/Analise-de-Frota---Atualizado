package diario_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/diario"
	"analisefrota/modules/diario/diario_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) SubstituirDia(ctx context.Context, clienteID int64, dia string, rows []diario_repo.RegistroRow_t) error {
	return m.Called(ctx, clienteID, dia, rows).Error(0)
}
func (m *RepositoryMock_t) DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) ([]string, error) {
	a := m.Called(ctx, clienteID, ini, fim)
	return a.Get(0).([]string), a.Error(1)
}
func (m *RepositoryMock_t) Listar(ctx context.Context, clienteID int64, ini, fim string) ([]diario_repo.RegistroRow_t, error) {
	a := m.Called(ctx, clienteID, ini, fim)
	return a.Get(0).([]diario_repo.RegistroRow_t), a.Error(1)
}

type DiarioSuite struct {
	suite.Suite
	repo *RepositoryMock_t
	svc  diario.Service_I
}

func (s *DiarioSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.svc = diario.NewDiarioCom(s.repo)
}

// lê JSON no formato da API, como o adapter faria
func estat(s *DiarioSuite, j string) []gobraxapi.EstatisticaRec_t {
	var v []gobraxapi.EstatisticaRec_t
	s.Require().NoError(json.Unmarshal([]byte(j), &v))
	return v
}
func perf(s *DiarioSuite, j string) []gobraxapi.PerformanceRec_t {
	var v []gobraxapi.PerformanceRec_t
	s.Require().NoError(json.Unmarshal([]byte(j), &v))
	return v
}

func (s *DiarioSuite) TestJuntaEstatisticaEPerformancePelaPlaca() {
	e := estat(s, `[{"vehicleIdentification":"abc-1d23","totalMileage":412.345,"totalConsumption":150.5,"averageSpeed":61,"odometer":291862.2}]`)
	p := perf(s, `[{"vehicleIdentification":"ABC1D23",
		"drivers":[{"driverName":"Ana","cpf":"<cpf>","startDate":"2026-10-23T03:00:00+0000","endDate":"2026-10-23T05:00:00+0000"},
		           {"driverName":"Beto","cpf":"<cpf>","startDate":"2026-10-23T05:00:00+0000","endDate":"2026-10-23T15:00:00+0000"}],
		"indicators":{"movement":{"duration":540},"greenRange":{"percentage":71.234},"pedalPressureOnHig":{"percentage":4.5},"idle":{"percentage":9}}}]`)

	regs := s.svc.MontarDia(1, "2026-10-23", e, p)

	s.Require().Len(regs, 1)
	r := regs[0]
	s.Equal("ABC-1D23", r.Placa)
	s.True(r.Rodou)
	s.Equal(412.35, *r.Km)
	s.Equal(150.5, *r.Litros)
	s.Equal("Beto", r.Motorista, "fica quem dirigiu por mais tempo")
	s.Equal(9.0, *r.Horas, "a duração vem em minutos")
	s.Equal(71.23, *r.FaixaVerde)
	s.Equal(4.5, *r.PressaoAcelerador, "aceita pedalPressureOnHig")
	s.Equal(9.0, *r.MotorOcioso)
	s.Nil(r.Nota, "a performance não traz nota geral")
	s.Nil(r.Embalo)
}

func (s *DiarioSuite) TestParadoEntraComZeroKmESemIndicadores() {
	e := estat(s, `[{"vehicleIdentification":"XYZ9A99","totalMileage":0,"totalConsumption":3,"odometer":1000}]`)
	p := perf(s, `[{"vehicleIdentification":"XYZ9A99","drivers":[{"driverName":"Caio"}],"indicators":{"greenRange":{"percentage":50}}}]`)

	regs := s.svc.MontarDia(1, "2026-10-23", e, p)

	s.Require().Len(regs, 1)
	s.False(regs[0].Rodou)
	s.Equal(0.0, *regs[0].Km)
	s.Equal(1000.0, *regs[0].Odometro)
	s.Nil(regs[0].Litros)
	s.Nil(regs[0].FaixaVerde)
	s.Empty(regs[0].Motorista)
}

func (s *DiarioSuite) TestSemEstatisticasUsaSoAPerformance() {
	p := perf(s, `[{"vehicleIdentification":"QWE1R23","drivers":[{"driverName":"Duda"}],"indicators":{"movement":{"duration":120},"leverage":{"percentage":20}}},
	               {"vehicleIdentification":"QWE1R24","indicators":{"movement":{"duration":0}}}]`)

	regs := s.svc.MontarDia(1, "2026-10-23", nil, p)

	s.Require().Len(regs, 2)
	s.True(regs[0].Rodou, "rodou = teve tempo em movimento")
	s.Nil(regs[0].Km, "sem estatísticas não há km")
	s.Equal(20.0, *regs[0].Embalo)
	s.Equal("Duda", regs[0].Motorista)
	s.False(regs[1].Rodou)
}

func (s *DiarioSuite) TestPlacaRepetidaEntraUmaVez() {
	e := estat(s, `[{"vehicleIdentification":"AAA1111","totalMileage":10},{"vehicleIdentification":"aaa-1111","totalMileage":99},{"vehicleIdentification":""}]`)
	regs := s.svc.MontarDia(1, "2026-10-23", e, nil)
	s.Require().Len(regs, 1)
	s.Equal(10.0, *regs[0].Km)
}

func (s *DiarioSuite) TestNotaForaDaFaixaEDescartada() {
	e := estat(s, `[{"vehicleIdentification":"AAA1111","totalMileage":10},{"vehicleIdentification":"BBB2222","totalMileage":10}]`)
	p := perf(s, `[{"vehicleIdentification":"AAA1111","score":{"score":87.5}},{"vehicleIdentification":"BBB2222","score":140}]`)
	regs := s.svc.MontarDia(1, "2026-10-23", e, p)
	s.Equal(87.5, *regs[0].Nota)
	s.Nil(regs[1].Nota)
}

func (s *DiarioSuite) TestSalvarDiaRecusaDataInvalida() {
	err := s.svc.SalvarDia(context.Background(), 1, "23/10/2026", nil)
	s.ErrorIs(err, diario.ErrDiaInvalido)
	s.repo.AssertNotCalled(s.T(), "SubstituirDia", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *DiarioSuite) TestListarValidaPeriodo() {
	_, err := s.svc.Listar(context.Background(), 1, "2026-10-30", "2026-10-01")
	s.ErrorIs(err, diario.ErrPeriodoInvertido)
}

func TestDiarioSuite(t *testing.T) { suite.Run(t, new(DiarioSuite)) }
