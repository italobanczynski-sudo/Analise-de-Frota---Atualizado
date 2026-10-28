package mensagem_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/modules/diario"
	"analisefrota/modules/mensagem"
	"analisefrota/modules/mensagem/mensagem_repo"
	"analisefrota/modules/premiacao"
)

// ---------- mocks ----------

type DiarioMock_t struct {
	mock.Mock
	diario.Service_I
}

func (m *DiarioMock_t) Listar(ctx context.Context, clienteID int64, ini, fim string) ([]diario.Registro_t, error) {
	a := m.Called(ctx, clienteID, ini, fim)
	return a.Get(0).([]diario.Registro_t), a.Error(1)
}

type PremiacaoMock_t struct {
	mock.Mock
	premiacao.Service_I
}

func (m *PremiacaoMock_t) Listar(ctx context.Context, clienteID int64, mes string) ([]premiacao.Premiacao_t, error) {
	a := m.Called(ctx, clienteID, mes)
	return a.Get(0).([]premiacao.Premiacao_t), a.Error(1)
}

type RepoMock_t struct{ mock.Mock }

func (m *RepoMock_t) Marcar(ctx context.Context, row mensagem_repo.EnviadaRow_t) error {
	return m.Called(ctx, row).Error(0)
}
func (m *RepoMock_t) Desmarcar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error) {
	a := m.Called(ctx, clienteID, motorista, dia)
	return a.Bool(0), a.Error(1)
}
func (m *RepoMock_t) ListarDia(ctx context.Context, clienteID int64, dia string) ([]mensagem_repo.EnviadaRow_t, error) {
	a := m.Called(ctx, clienteID, dia)
	return a.Get(0).([]mensagem_repo.EnviadaRow_t), a.Error(1)
}
func (m *RepoMock_t) ListarMotorista(ctx context.Context, clienteID int64, motorista string) ([]mensagem_repo.EnviadaRow_t, error) {
	a := m.Called(ctx, clienteID, motorista)
	return a.Get(0).([]mensagem_repo.EnviadaRow_t), a.Error(1)
}

// ---------- dados ----------

func f(v float64) *float64 { return &v }

// reg é um veículo num dia: 10 h, 500 km, 200 L (2,5 km/l), embalo 20%, piloto 30%, pedal crítico 5%,
// e o início da faixa verde e o motor ligado parado dados
func reg(mot, dia string, inicioFV, ocioso float64) diario.Registro_t {
	return diario.Registro_t{Motorista: mot, Dia: dia, Rodou: true, Km: f(500), Litros: f(200), Horas: f(10),
		InicioFaixaVerde: f(inicioFV), MotorOcioso: f(ocioso), Embalo: f(20), PilotoAutomatico: f(30), PressaoAcelerador: f(5)}
}

type MensagemSuite struct {
	suite.Suite
	repo      *RepoMock_t
	diario    *DiarioMock_t
	premiacao *PremiacaoMock_t
	svc       mensagem.Service_I
	ctx       context.Context
}

func (s *MensagemSuite) SetupTest() {
	s.repo, s.diario, s.premiacao = new(RepoMock_t), new(DiarioMock_t), new(PremiacaoMock_t)
	s.ctx = context.Background()
	agora := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC) // 27/10 às 7h de Brasília
	s.svc = mensagem.NewMensagemCom(mensagem.Deps_t{Repo: s.repo, Diario: s.diario, Premiacao: s.premiacao, Agora: func() time.Time { return agora }})
}

func (s *MensagemSuite) TestAnaliseComparaCadaIndicadorComAMedia() {
	ontem := reg("JOAO DA SILVA", "2026-10-26", 75, 20) // início da faixa verde subiu, motor ligado parado dobrou
	ontem.Litros = f(180)                               // 2,78 km/l
	ontem.PilotoAutomatico = f(30.5)                    // estável
	regs := []diario.Registro_t{reg("JOAO DA SILVA", "2026-10-24", 70, 8), reg("JOAO DA SILVA", "2026-10-25", 72, 10), ontem}

	r := mensagem.MontarResumos("2026-10-26", regs, nil)

	s.Require().Len(r, 1)
	v := r[0].Variaveis
	s.Equal("Joao", v["1"])
	s.Equal("26/10", v["2"])
	s.Equal("500 km", v["3"])
	s.Equal("2,78 km/l ✅ melhorou 0,28 km/l", v["4"])
	s.Equal("75,0% ✅ melhorou 4,0 pts", v["5"])
	s.Equal("20,0% ➖ estável", v["6"])
	s.Equal("20,0% ⚠️ piorou 11,0 pts", v["7"], "motor ligado parado: subir é piorar")
	s.Equal("30,5% ➖ estável", v["8"])
	s.Equal("5,0% ➖ estável", v["9"])
	s.Equal("ainda não calculada", v["10"])
	s.Equal("Você evoluiu! Contra a sua média dos últimos 7 dias, você melhorou em 2 e piorou em 1 dos 6 indicadores.", v["11"])
	s.Contains(v["12"], "Motor ligado parado ficou em 20,0% (sua média: 9,0%)")
	s.Contains(v["12"], "desligue o motor")
}

func (s *MensagemSuite) TestEvoluiuOuRegrediu() {
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 60, 10), reg("Ana", "2026-10-26", 70, 5)}, nil)
	s.True(strings.HasPrefix(r[0].Evolucao, "Você evoluiu!"), r[0].Evolucao)
	s.Contains(r[0].Evolucao, "melhorou em 2 e piorou em 0")

	r = mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 70, 5), reg("Ana", "2026-10-26", 60, 10)}, nil)
	s.True(strings.HasPrefix(r[0].Evolucao, "Você regrediu um pouco."), r[0].Evolucao)
	s.Contains(r[0].Dica, "Início da faixa verde ficou em 60,0%")
}

func (s *MensagemSuite) TestNotaDoMesContraOMesAnterior() {
	regs := []diario.Registro_t{reg("ANA  LIMA", "2026-10-26", 70, 5)}
	notas := map[string]mensagem.Nota_t{mensagem.ChaveMotorista("Ana Lima"): {Mes: "2026-10", Atual: f(87.5), Anterior: f(84), Parcial: true}}
	r := mensagem.MontarResumos("2026-10-26", regs, notas)
	s.Equal("87,5 em outubro (parcial) ✅ subiu 3,5 pontos em relação a setembro", r[0].Variaveis["10"])
	s.Equal(87.5, *r[0].Nota)

	notas[mensagem.ChaveMotorista("Ana Lima")] = mensagem.Nota_t{Mes: "2026-01", Atual: f(80), Anterior: f(90)}
	r = mensagem.MontarResumos("2026-10-26", regs, notas)
	s.Equal("80,0 em janeiro ⚠️ caiu 10,0 pontos em relação a dezembro", r[0].Variaveis["10"])

	notas[mensagem.ChaveMotorista("Ana Lima")] = mensagem.Nota_t{Mes: "2026-10", Atual: f(80)}
	r = mensagem.MontarResumos("2026-10-26", regs, notas)
	s.Equal("80,0 em outubro", r[0].Variaveis["10"])
}

func (s *MensagemSuite) TestJuntaOsVeiculosDoMesmoDiaPeloTempo() {
	a := reg("Ana", "2026-10-26", 80, 5)
	b := reg("Ana", "2026-10-26", 50, 5)
	b.Horas, b.Km, b.Litros = f(30), f(1000), f(500) // dirigiu 3x mais tempo no segundo veículo
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{a, b}, nil)

	s.Require().Len(r, 1)
	s.Equal("57,5%", r[0].Variaveis["5"])
	s.Equal("1.500 km", r[0].Variaveis["3"])
	s.Equal(2.14, *r[0].Kml, "1.500 km / 700 L")
}

func (s *MensagemSuite) TestSemHistoricoESemPiora() {
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-26", 80, 5)}, nil)
	s.Contains(r[0].Evolucao, "Ainda não temos dias anteriores")
	s.Equal("80,0%", r[0].Variaveis["5"], "sem histórico, só o valor")

	r = mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 80, 5), reg("Ana", "2026-10-26", 80.5, 5.5)}, nil)
	s.Contains(r[0].Evolucao, "Você manteve o nível")
	s.Contains(r[0].Dica, "nenhum indicador piorou")
}

func (s *MensagemSuite) TestIndicadorSemDado() {
	ontem := reg("Ana", "2026-10-26", 80, 5)
	ontem.PilotoAutomatico, ontem.Litros = nil, nil
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 80, 5), ontem}, nil)
	s.Equal("sem dado ontem", r[0].Variaveis["8"])
	s.Equal("sem dado ontem", r[0].Variaveis["4"])
	s.Contains(r[0].Evolucao, "dos 4 indicadores")
}

func (s *MensagemSuite) TestIgnoraQuemNaoRodouOuSemMotorista() {
	parado := reg("Beto", "2026-10-26", 80, 5)
	parado.Rodou = false
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{parado, reg("", "2026-10-26", 80, 5), reg("Caio", "2026-10-25", 80, 5)}, nil)
	s.Empty(r)
}

func (s *MensagemSuite) TestVariaveisSemQuebraDeLinha() {
	notas := map[string]mensagem.Nota_t{"ANA": {Mes: "2026-10", Atual: f(90), Anterior: f(88)}}
	r := mensagem.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-20", 60, 5), reg("Ana", "2026-10-26", 40, 30)}, notas)
	s.Len(r[0].Variaveis, 12)
	for k, v := range r[0].Variaveis {
		s.NotEmpty(v, k)
		s.NotContains(v, "\n", k)
	}
	txt := mensagem.Texto(r[0].Variaveis)
	s.True(strings.HasPrefix(txt, "Bom dia, Ana!"))
	s.Contains(txt, "⭐ Nota do mês: 90,0 em outubro")
	s.NotContains(txt, "{{")
}

// ---------- envio ----------

// ---------- mensagens do dia ----------

func (s *MensagemSuite) TestListarTrazOTextoProntoEANotaDaPremiacao() {
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-19", "2026-10-26").Return([]diario.Registro_t{reg("ANA LIMA", "2026-10-26", 80, 5)}, nil)
	s.premiacao.On("Listar", s.ctx, int64(1), "2026-10").Return([]premiacao.Premiacao_t{{Motorista: "Ana Lima", Nota: f(87.5)}}, nil)
	s.premiacao.On("Listar", s.ctx, int64(1), "2026-09").Return([]premiacao.Premiacao_t{{Motorista: "ANA LIMA", Nota: f(84)}}, nil)
	s.repo.On("ListarDia", s.ctx, int64(1), "2026-10-26").Return([]mensagem_repo.EnviadaRow_t{}, nil)

	m, err := s.svc.Listar(s.ctx, 1, "2026-10-26")

	s.Require().NoError(err)
	s.Require().Len(m, 1)
	s.True(strings.HasPrefix(m[0].Texto, "Bom dia, Ana!"))
	s.Contains(m[0].Texto, "87,5 em outubro (parcial) ✅ subiu 3,5 pontos em relação a setembro")
	s.True(strings.HasSuffix(m[0].Texto, "Boa viagem!"))
	s.Nil(m[0].Enviada)
}

func (s *MensagemSuite) semNotas() {
	s.premiacao.On("Listar", s.ctx, mock.Anything, mock.Anything).Return([]premiacao.Premiacao_t{}, nil)
}

func (s *MensagemSuite) TestMarcarGuardaORetratoDaAnalise() {
	s.semNotas()
	ontem := reg("ANA", "2026-10-26", 75, 20) // motor ligado parado foi de 8 para 20: é o ponto apontado
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-19", "2026-10-26").Return([]diario.Registro_t{reg("ANA", "2026-10-25", 70, 8), ontem}, nil)
	var gravado mensagem_repo.EnviadaRow_t
	s.repo.On("Marcar", s.ctx, mock.Anything).Run(func(a mock.Arguments) { gravado = a.Get(1).(mensagem_repo.EnviadaRow_t) }).Return(nil)

	m, err := s.svc.Marcar(s.ctx, 1, "2026-10-26", "ana", 9)

	s.Require().NoError(err)
	s.NotNil(m.Enviada)
	s.Equal("ANA", gravado.Motorista, "grava o nome como vem da API")
	s.Equal("motor_ocioso", gravado.Foco)
	s.Equal(20.0, *gravado.FocoValor)
	s.Equal(8.0, *gravado.FocoMedia)
	s.Equal(int64(9), *gravado.MarcadoPor)

	_, err = s.svc.Marcar(s.ctx, 1, "2026-10-26", "Beto", 9)
	s.ErrorIs(err, mensagem.ErrMotoristaSemDia)
}

func (s *MensagemSuite) TestListarMostraQuemJaFoiEnviado() {
	s.semNotas()
	s.diario.On("Listar", s.ctx, int64(1), mock.Anything, mock.Anything).Return([]diario.Registro_t{reg("ANA", "2026-10-26", 75, 5), reg("BETO", "2026-10-26", 75, 5)}, nil)
	s.repo.On("ListarDia", s.ctx, int64(1), "2026-10-26").Return([]mensagem_repo.EnviadaRow_t{{Motorista: "ANA", MarcadoEm: time.Now()}}, nil)

	m, err := s.svc.Listar(s.ctx, 1, "2026-10-26")

	s.Require().NoError(err)
	s.NotNil(m[0].Enviada)
	s.Nil(m[1].Enviada)
}

func (s *MensagemSuite) TestAcompanhamentoMedeOPontoApontadoNosDiasSeguintes() {
	s.repo.On("ListarMotorista", s.ctx, int64(1), "Ana").Return([]mensagem_repo.EnviadaRow_t{
		{Motorista: "ANA", Dia: "2026-10-10", Foco: "motor_ocioso", FocoValor: f(20), FocoMedia: f(8)},
		{Motorista: "ANA", Dia: "2026-10-20", Foco: "inicio_faixa_verde", FocoValor: f(60), FocoMedia: f(75)},
		{Motorista: "ANA", Dia: "2026-10-26", Foco: "motor_ocioso", FocoValor: f(15), FocoMedia: f(9)},
		{Motorista: "ANA", Dia: "2026-10-05", Foco: ""},
	}, nil)
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-05", "2026-10-26").Return([]diario.Registro_t{
		reg("ANA", "2026-10-11", 70, 9), reg("ANA", "2026-10-12", 70, 11), // depois de 10/10: motor parado caiu para 10 → melhorou
		reg("ANA", "2026-10-21", 59.5, 5), reg("BETO", "2026-10-21", 90, 5), // depois de 20/10: faixa verde ficou igual → não melhorou
	}, nil)

	a, err := s.svc.Acompanhamento(s.ctx, 1, "Ana")

	s.Require().NoError(err)
	s.Equal(4, a.DiasEnviada)
	s.Equal("2026-10-26", a.Envios[0].Dia, "o mais recente primeiro")
	por := map[string]mensagem.Envio_t{}
	for _, e := range a.Envios {
		por[e.Dia] = e
	}
	s.Equal(mensagem.ResultadoMelhorou, por["2026-10-10"].Resultado)
	s.Equal(10.0, *por["2026-10-10"].Depois)
	s.Equal(2, por["2026-10-10"].DiasDepois)
	s.Equal(mensagem.ResultadoNaoMelhorou, por["2026-10-20"].Resultado)
	s.Equal(mensagem.ResultadoAguardando, por["2026-10-26"].Resultado, "ainda não rodou depois")
	s.Equal(mensagem.ResultadoSemPonto, por["2026-10-05"].Resultado)
	s.Equal("Motor ligado parado", a.Focos[0].Rotulo)
	s.Equal(2, a.Focos[0].Vezes)
	s.Equal(1, a.Focos[0].Melhorou)
	s.Equal(1, a.Focos[0].Aguardando)
}

func (s *MensagemSuite) TestNaoMontaDiaQueNaoTerminou() {
	_, err := s.svc.Listar(s.ctx, 1, "2026-10-27")
	s.ErrorIs(err, mensagem.ErrDiaFuturo)
}

func TestMensagemSuite(t *testing.T) { suite.Run(t, new(MensagemSuite)) }
