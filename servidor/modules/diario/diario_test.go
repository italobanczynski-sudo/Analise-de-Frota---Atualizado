package diario_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

func (m *RepositoryMock_t) SubstituirMotoristasDia(ctx context.Context, clienteID int64, dia string, rows []diario_repo.RegistroRow_t) error {
	return m.Called(ctx, clienteID, dia, rows).Error(0)
}
func (m *RepositoryMock_t) ListarMotoristas(ctx context.Context, clienteID int64, ini, fim string) ([]diario_repo.RegistroRow_t, error) {
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

func (s *DiarioSuite) TestGuardaABaseDeCadaIndicadorComoAPlataforma() {
	e := estat(s, `[{"vehicleIdentification":"AAA1111","totalMileage":100}]`)
	// 540 min em movimento = 90% da ignição (600 min); pedal: 50% = 120 min → 240 min com o pé no acelerador
	p := perf(s, `[{"vehicleIdentification":"AAA1111","drivers":[{"driverName":"Ana"}],
		"indicators":{"movement":{"duration":540,"percentage":90},"idle":{"duration":60,"percentage":10},
		"pedalPressureOnLow":{"duration":120,"percentage":50},"pedalPressureOnHig":{"duration":24,"percentage":10}}}]`)

	regs := s.svc.MontarDia(1, "2026-10-23", e, p)

	s.Require().Len(regs, 1)
	s.Equal(9.0, *regs[0].Horas)
	s.Equal(10.0, *regs[0].HorasMotor, "ignição = 540 ÷ 90% = 600 min")
	s.Equal(4.0, *regs[0].HorasPedal, "pé no acelerador = 120 ÷ 50% = 240 min")
}
func TestDiarioSuite(t *testing.T) { suite.Run(t, new(DiarioSuite)) }

// ---------- motoristas: o dado de cada um só no tempo do vínculo dele com o veículo ----------

func (s *DiarioSuite) TestPlanoSeparaQuemFicouODiaTodoDosVinculosParciais() {
	// 23/10 em Brasília = 03:00 de 23/10 até 02:59:59 de 24/10 (UTC)
	e := estat(s, `[{"vehicleIdentification":"AAA1111","totalMileage":300},{"vehicleIdentification":"BBB2222","totalMileage":200},
		{"vehicleIdentification":"CCC3333","totalMileage":0}]`)
	p := perf(s, `[{"vehicleIdentification":"AAA1111","drivers":[{"driverName":" Caio "}],"indicators":{"movement":{"duration":300}}},
		{"vehicleIdentification":"BBB2222","drivers":[
			{"driverName":"Ana","startDate":"2026-10-23T01:00:00+0000","endDate":"2026-10-23T05:00:00+0000"},
			{"driverName":"Beto","startDate":"2026-10-23T05:00:00+0000","endDate":"2026-10-23T15:00:00+0000"},
			{"driverName":"Ana","startDate":"2026-10-23T18:00:00+0000","endDate":"2026-10-25T00:00:00+0000"}],
			"indicators":{"movement":{"duration":400}}},
		{"vehicleIdentification":"CCC3333","drivers":[{"driverName":"Duda"}]}]`)
	regs := s.svc.MontarDia(1, "2026-10-23", e, p)

	plano := s.svc.PlanoMotoristas("2026-10-23", regs, p)

	s.Require().Len(plano.Inteiros, 1, "parado não entra")
	s.Equal("Caio", plano.Inteiros[0].Motorista)
	s.Equal(300.0, *plano.Inteiros[0].Km, "o dia todo no veículo = a linha do veículo")
	s.Require().Len(plano.Pendentes, 2)
	ana, beto := plano.Pendentes[0], plano.Pendentes[1]
	s.Equal("Ana", ana.Motorista)
	s.Equal("BBB2222", ana.Placa)
	s.Require().Len(ana.Janelas, 2, "os dois trechos da Ana")
	s.Equal(time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC), ana.Janelas[0].Ini.UTC(), "recortado no início do dia")
	s.Equal(time.Date(2026, 10, 24, 2, 59, 59, 0, time.UTC), ana.Janelas[1].Fim.UTC(), "recortado no fim do dia")
	s.Equal("Beto", beto.Motorista)
	s.Equal(time.Date(2026, 10, 23, 15, 0, 0, 0, time.UTC), beto.Janelas[0].Fim.UTC())
}

func (s *DiarioSuite) TestUmMotoristaSoComParteDoDiaTambemEConsultado() {
	e := estat(s, `[{"vehicleIdentification":"AAA1111","totalMileage":300}]`)
	p := perf(s, `[{"vehicleIdentification":"AAA1111","drivers":[{"driverName":"Ana","startDate":"2026-10-23T05:00:00+0000","endDate":"2026-10-23T15:00:00+0000"}],
		"indicators":{"movement":{"duration":300}}}]`)
	regs := s.svc.MontarDia(1, "2026-10-23", e, p)

	plano := s.svc.PlanoMotoristas("2026-10-23", regs, p)

	s.Empty(plano.Inteiros, "o veículo rodou também fora do vínculo dela")
	s.Require().Len(plano.Pendentes, 1)
	s.Equal("Ana", plano.Pendentes[0].Motorista)
}

func (s *DiarioSuite) TestMotoristaJuntaOsTrechosPonderandoCadaIndicadorPelaSuaBase() {
	pend := diario.Pendente_t{Placa: "ABC-1D23", Motorista: "Ana"}
	partes := []diario.Parte_t{
		{
			Estat: estat(s, `[{"vehicleIdentification":"ABC1D23","totalMileage":100,"totalConsumption":40,"totalBreaking":3,"averageSpeed":60,"odometer":1000},
				{"vehicleIdentification":"ZZZ0000","totalMileage":999}]`),
			// movimento 120 min = 80% da ignição (150 min); pé no acelerador 30 + 30 = 60 min
			Perf: perf(s, `[{"vehicleIdentification":"ABC1D23","indicators":{"movement":{"duration":120,"percentage":80},"greenRange":{"percentage":50},
				"idle":{"percentage":10},"pedalPressureOnLow":{"duration":30,"percentage":50},"pedalPressureOnHigh":{"duration":30,"percentage":50}}}]`),
		},
		{
			Estat: estat(s, `[{"vehicleIdentification":"ABC1D23","totalMileage":50,"totalConsumption":25,"totalBreaking":1,"averageSpeed":90,"odometer":1050}]`),
			// movimento 60 min = 50% da ignição (120 min); pé no acelerador 14 + 6 = 20 min
			Perf: perf(s, `[{"vehicleIdentification":"ABC1D23","indicators":{"movement":{"duration":60,"percentage":50},"greenRange":{"percentage":80},
				"idle":{"percentage":20},"pedalPressureOnLow":{"duration":14,"percentage":70},"pedalPressureOnHig":{"duration":6,"percentage":30}}}]`),
		},
	}

	r, ok := s.svc.MontarMotorista(1, "2026-10-23", pend, partes)

	s.Require().True(ok)
	s.Equal("Ana", r.Motorista)
	s.Equal("ABC-1D23", r.Placa)
	s.True(r.Rodou)
	s.Equal(150.0, *r.Km, "a outra placa não entra")
	s.Equal(65.0, *r.Litros)
	s.Equal(4.0, *r.Frenagens)
	s.Equal(70.0, *r.VelocidadeMedia, "ponderada pelos km")
	s.Equal(1050.0, *r.Odometro, "o maior odômetro")
	s.Equal(3.0, *r.Horas, "180 min em movimento")
	s.Equal(4.5, *r.HorasMotor, "150 + 120 min de ignição")
	s.Equal(1.33, *r.HorasPedal, "60 + 20 min de pé no acelerador")
	s.Equal(60.0, *r.FaixaVerde, "(50×120 + 80×60) ÷ 180")
	s.Equal(14.44, *r.MotorOcioso, "(10×150 + 20×120) ÷ 270: pela ignição")
	s.Equal(55.0, *r.PressaoIdeal, "(50×60 + 70×20) ÷ 80: pelo pé no acelerador")
	s.Equal(45.0, *r.PressaoAcelerador, "(50×60 + 30×20) ÷ 80, aceitando pedalPressureOnHig")
	s.Nil(r.Embalo, "a API não trouxe")
	s.Nil(r.Nota)
}

func (s *DiarioSuite) TestMotoristaQueNaoRodouNoTrechoNaoEntra() {
	pend := diario.Pendente_t{Placa: "ABC1D23", Motorista: "Ana"}
	partes := []diario.Parte_t{{Estat: estat(s, `[{"vehicleIdentification":"ABC1D23","totalMileage":0}]`)}}
	_, ok := s.svc.MontarMotorista(1, "2026-10-23", pend, partes)
	s.False(ok)
}

func (s *DiarioSuite) TestSalvarMotoristasRecusaLinhaSemNome() {
	err := s.svc.SalvarMotoristasDia(context.Background(), 1, "2026-10-23", []diario.Registro_t{{Placa: "ABC1D23"}})
	s.ErrorIs(err, diario.ErrSemMotorista)
	s.repo.AssertNotCalled(s.T(), "SubstituirMotoristasDia", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (s *DiarioSuite) TestSalvarMotoristasSubstituiODia() {
	s.repo.On("SubstituirMotoristasDia", mock.Anything, int64(1), "2026-10-23", mock.Anything).Return(nil)
	err := s.svc.SalvarMotoristasDia(context.Background(), 1, "2026-10-23", []diario.Registro_t{{Placa: "ABC1D23", Motorista: "Ana"}})
	s.Require().NoError(err)
	s.repo.AssertExpectations(s.T())
}
