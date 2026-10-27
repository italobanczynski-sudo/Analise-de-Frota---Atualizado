package whatsapp_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/twilio"
	"analisefrota/environment"
	"analisefrota/modules/cliente"
	"analisefrota/modules/contato"
	"analisefrota/modules/diario"
	"analisefrota/modules/premiacao"
	"analisefrota/modules/whatsapp"
	"analisefrota/modules/whatsapp/whatsapp_repo"
)

// ---------- mocks ----------

type RepoMock_t struct{ mock.Mock }

func (m *RepoMock_t) Reservar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error) {
	a := m.Called(ctx, clienteID, motorista, dia)
	return a.Bool(0), a.Error(1)
}
func (m *RepoMock_t) Concluir(ctx context.Context, clienteID int64, motorista, dia, situacao, sid, erro string) error {
	return m.Called(ctx, clienteID, motorista, dia, situacao, sid, erro).Error(0)
}
func (m *RepoMock_t) Listar(ctx context.Context, clienteID int64, desde string) ([]whatsapp_repo.EnvioRow_t, error) {
	a := m.Called(ctx, clienteID, desde)
	return a.Get(0).([]whatsapp_repo.EnvioRow_t), a.Error(1)
}
func (m *RepoMock_t) BuscarConfig(ctx context.Context) (whatsapp_repo.ConfigRow_t, bool, error) {
	a := m.Called(ctx)
	return a.Get(0).(whatsapp_repo.ConfigRow_t), a.Bool(1), a.Error(2)
}
func (m *RepoMock_t) SalvarConfig(ctx context.Context, row whatsapp_repo.ConfigRow_t) error {
	return m.Called(ctx, row).Error(0)
}

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

type ContatosMock_t struct {
	mock.Mock
	contato.Service_I
}

func (m *ContatosMock_t) Destinos(ctx context.Context, clienteID int64) ([]contato.Destino_t, error) {
	a := m.Called(ctx, clienteID)
	return a.Get(0).([]contato.Destino_t), a.Error(1)
}

type ClientesMock_t struct {
	mock.Mock
	cliente.Service_I
}

func (m *ClientesMock_t) Listar(ctx context.Context) ([]cliente.Cliente_t, error) {
	a := m.Called(ctx)
	return a.Get(0).([]cliente.Cliente_t), a.Error(1)
}

type TwilioMock_t struct{ mock.Mock }

func (m *TwilioMock_t) Enviar(ctx context.Context, telefone string, vars map[string]string, texto string) (string, error) {
	a := m.Called(ctx, telefone, vars, texto)
	return a.String(0), a.Error(1)
}
func (m *TwilioMock_t) Teste() bool { return false }

// ---------- dados ----------

func f(v float64) *float64 { return &v }

// reg é um veículo num dia: 10 h, 500 km, 200 L (2,5 km/l), embalo 20%, piloto 30%, pedal crítico 5%,
// e o início da faixa verde e o motor ligado parado dados
func reg(mot, dia string, inicioFV, ocioso float64) diario.Registro_t {
	return diario.Registro_t{Motorista: mot, Dia: dia, Rodou: true, Km: f(500), Litros: f(200), Horas: f(10),
		InicioFaixaVerde: f(inicioFV), MotorOcioso: f(ocioso), Embalo: f(20), PilotoAutomatico: f(30), PressaoAcelerador: f(5)}
}

type WhatsAppSuite struct {
	suite.Suite
	repo      *RepoMock_t
	diario    *DiarioMock_t
	premiacao *PremiacaoMock_t
	contatos  *ContatosMock_t
	clientes  *ClientesMock_t
	twilio    *TwilioMock_t
	svc       whatsapp.Service_I
	ctx       context.Context
	chave     []byte
	cfgUsada  environment.WhatsAppConfig_t // a configuração com que o envio foi montado
}

func (s *WhatsAppSuite) SetupTest() {
	s.repo, s.diario, s.premiacao = new(RepoMock_t), new(DiarioMock_t), new(PremiacaoMock_t)
	s.contatos, s.clientes, s.twilio = new(ContatosMock_t), new(ClientesMock_t), new(TwilioMock_t)
	s.ctx = context.Background()
	s.chave = bytes.Repeat([]byte{7}, 32)
	agora := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC) // 27/10 às 7h de Brasília
	s.svc = whatsapp.NewWhatsAppCom(whatsapp.Deps_t{Repo: s.repo, Clientes: s.clientes, Contatos: s.contatos, Diario: s.diario,
		Premiacao: s.premiacao, Chave: s.chave, Agora: func() time.Time { return agora },
		Env:    environment.WhatsAppConfig_t{Hora: 7},
		Twilio: func(cfg environment.WhatsAppConfig_t) twilio.WhatsApp_I { s.cfgUsada = cfg; return s.twilio }})
}

// semConfigNaTela: o administrador ainda não salvou a conexão (vale a do ambiente)
func (s *WhatsAppSuite) semConfigNaTela() {
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{}, false, nil)
}

// semNotas: a premiação não traz ninguém (os testes de envio não olham a nota)
func (s *WhatsAppSuite) semNotas() {
	s.premiacao.On("Listar", s.ctx, mock.Anything, mock.Anything).Return([]premiacao.Premiacao_t{}, nil)
}

// ---------- análise individual ----------

func (s *WhatsAppSuite) TestAnaliseComparaCadaIndicadorComAMedia() {
	ontem := reg("JOAO DA SILVA", "2026-10-26", 75, 20) // início da faixa verde subiu, motor ligado parado dobrou
	ontem.Litros = f(180)                               // 2,78 km/l
	ontem.PilotoAutomatico = f(30.5)                    // estável
	regs := []diario.Registro_t{reg("JOAO DA SILVA", "2026-10-24", 70, 8), reg("JOAO DA SILVA", "2026-10-25", 72, 10), ontem}

	r := whatsapp.MontarResumos("2026-10-26", regs, nil)

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

func (s *WhatsAppSuite) TestEvoluiuOuRegrediu() {
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 60, 10), reg("Ana", "2026-10-26", 70, 5)}, nil)
	s.True(strings.HasPrefix(r[0].Evolucao, "Você evoluiu!"), r[0].Evolucao)
	s.Contains(r[0].Evolucao, "melhorou em 2 e piorou em 0")

	r = whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 70, 5), reg("Ana", "2026-10-26", 60, 10)}, nil)
	s.True(strings.HasPrefix(r[0].Evolucao, "Você regrediu um pouco."), r[0].Evolucao)
	s.Contains(r[0].Dica, "Início da faixa verde ficou em 60,0%")
}

func (s *WhatsAppSuite) TestNotaDoMesContraOMesAnterior() {
	regs := []diario.Registro_t{reg("ANA  LIMA", "2026-10-26", 70, 5)}
	notas := map[string]whatsapp.Nota_t{whatsapp.ChaveMotorista("Ana Lima"): {Mes: "2026-10", Atual: f(87.5), Anterior: f(84), Parcial: true}}
	r := whatsapp.MontarResumos("2026-10-26", regs, notas)
	s.Equal("87,5 em outubro (parcial) ✅ subiu 3,5 pontos em relação a setembro", r[0].Variaveis["10"])
	s.Equal(87.5, *r[0].Nota)

	notas[whatsapp.ChaveMotorista("Ana Lima")] = whatsapp.Nota_t{Mes: "2026-01", Atual: f(80), Anterior: f(90)}
	r = whatsapp.MontarResumos("2026-10-26", regs, notas)
	s.Equal("80,0 em janeiro ⚠️ caiu 10,0 pontos em relação a dezembro", r[0].Variaveis["10"])

	notas[whatsapp.ChaveMotorista("Ana Lima")] = whatsapp.Nota_t{Mes: "2026-10", Atual: f(80)}
	r = whatsapp.MontarResumos("2026-10-26", regs, notas)
	s.Equal("80,0 em outubro", r[0].Variaveis["10"])
}

func (s *WhatsAppSuite) TestJuntaOsVeiculosDoMesmoDiaPeloTempo() {
	a := reg("Ana", "2026-10-26", 80, 5)
	b := reg("Ana", "2026-10-26", 50, 5)
	b.Horas, b.Km, b.Litros = f(30), f(1000), f(500) // dirigiu 3x mais tempo no segundo veículo
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{a, b}, nil)

	s.Require().Len(r, 1)
	s.Equal("57,5%", r[0].Variaveis["5"])
	s.Equal("1.500 km", r[0].Variaveis["3"])
	s.Equal(2.14, *r[0].Kml, "1.500 km / 700 L")
}

func (s *WhatsAppSuite) TestSemHistoricoESemPiora() {
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-26", 80, 5)}, nil)
	s.Contains(r[0].Evolucao, "Ainda não temos dias anteriores")
	s.Equal("80,0%", r[0].Variaveis["5"], "sem histórico, só o valor")

	r = whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 80, 5), reg("Ana", "2026-10-26", 80.5, 5.5)}, nil)
	s.Contains(r[0].Evolucao, "Você manteve o nível")
	s.Contains(r[0].Dica, "nenhum indicador piorou")
}

func (s *WhatsAppSuite) TestIndicadorSemDado() {
	ontem := reg("Ana", "2026-10-26", 80, 5)
	ontem.PilotoAutomatico, ontem.Litros = nil, nil
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-25", 80, 5), ontem}, nil)
	s.Equal("sem dado ontem", r[0].Variaveis["8"])
	s.Equal("sem dado ontem", r[0].Variaveis["4"])
	s.Contains(r[0].Evolucao, "dos 4 indicadores")
}

func (s *WhatsAppSuite) TestIgnoraQuemNaoRodouOuSemMotorista() {
	parado := reg("Beto", "2026-10-26", 80, 5)
	parado.Rodou = false
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{parado, reg("", "2026-10-26", 80, 5), reg("Caio", "2026-10-25", 80, 5)}, nil)
	s.Empty(r)
}

func (s *WhatsAppSuite) TestVariaveisSemQuebraDeLinha() {
	notas := map[string]whatsapp.Nota_t{"ANA": {Mes: "2026-10", Atual: f(90), Anterior: f(88)}}
	r := whatsapp.MontarResumos("2026-10-26", []diario.Registro_t{reg("Ana", "2026-10-20", 60, 5), reg("Ana", "2026-10-26", 40, 30)}, notas)
	s.Len(r[0].Variaveis, 12)
	for k, v := range r[0].Variaveis {
		s.NotEmpty(v, k)
		s.NotContains(v, "\n", k)
	}
	txt := whatsapp.Texto(r[0].Variaveis)
	s.True(strings.HasPrefix(txt, "Bom dia, Ana!"))
	s.Contains(txt, "⭐ Nota do mês: 90,0 em outubro")
	s.NotContains(txt, "{{")
}

// ---------- envio ----------

func (s *WhatsAppSuite) TestNotaVemDaPremiacaoDoMesEDoAnterior() {
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-19", "2026-10-26").Return([]diario.Registro_t{reg("ANA LIMA", "2026-10-26", 80, 5)}, nil)
	s.premiacao.On("Listar", s.ctx, int64(1), "2026-10").Return([]premiacao.Premiacao_t{{Motorista: "Ana Lima", Nota: f(87.5)}}, nil)
	s.premiacao.On("Listar", s.ctx, int64(1), "2026-09").Return([]premiacao.Premiacao_t{{Motorista: "ANA LIMA", Nota: f(84)}}, nil)
	s.contatos.On("Destinos", s.ctx, int64(1)).Return([]contato.Destino_t{}, nil)

	p, err := s.svc.Previa(s.ctx, 1, "2026-10-26")

	s.Require().NoError(err)
	s.Require().Len(p, 1)
	s.Contains(p[0].Texto, "87,5 em outubro (parcial) ✅ subiu 3,5 pontos em relação a setembro")
}

func (s *WhatsAppSuite) TestEnviaSoParaQuemTemWhatsAppEUmaVez() {
	s.semNotas()
	s.semConfigNaTela()
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-19", "2026-10-26").Return([]diario.Registro_t{
		reg("Ana", "2026-10-26", 80, 5), reg("Beto", "2026-10-26", 70, 5), reg("Caio", "2026-10-26", 60, 5),
	}, nil)
	s.contatos.On("Destinos", s.ctx, int64(1)).Return([]contato.Destino_t{{Motorista: "Ana", Telefone: "+5511900000001"}, {Motorista: "Beto", Telefone: "+5511900000002"}}, nil)
	s.repo.On("Reservar", s.ctx, int64(1), "Ana", "2026-10-26").Return(true, nil)
	s.repo.On("Reservar", s.ctx, int64(1), "Beto", "2026-10-26").Return(false, nil) // já recebeu
	s.twilio.On("Enviar", s.ctx, "+5511900000001", mock.Anything, mock.Anything).Return("SM1", nil)
	s.repo.On("Concluir", s.ctx, int64(1), "Ana", "2026-10-26", whatsapp.SituacaoEnviada, "SM1", "").Return(nil)

	r, err := s.svc.Enviar(s.ctx, 1, "2026-10-26")

	s.Require().NoError(err)
	s.Equal(1, r.Enviadas)
	s.Equal(1, r.JaEnviadas)
	s.Equal(1, r.SemWhatsApp)
	s.twilio.AssertNumberOfCalls(s.T(), "Enviar", 1)
}

func (s *WhatsAppSuite) TestCredencialRecusadaParaTudo() {
	s.semNotas()
	s.semConfigNaTela()
	s.diario.On("Listar", s.ctx, int64(1), mock.Anything, mock.Anything).Return([]diario.Registro_t{
		reg("Ana", "2026-10-26", 80, 5), reg("Beto", "2026-10-26", 70, 5),
	}, nil)
	s.contatos.On("Destinos", s.ctx, int64(1)).Return([]contato.Destino_t{{Motorista: "Ana", Telefone: "+5511900000001"}, {Motorista: "Beto", Telefone: "+5511900000002"}}, nil)
	s.repo.On("Reservar", s.ctx, int64(1), mock.Anything, mock.Anything).Return(true, nil)
	s.twilio.On("Enviar", s.ctx, mock.Anything, mock.Anything, mock.Anything).Return("", twilio.ErrCredenciais)
	s.repo.On("Concluir", s.ctx, int64(1), "Ana", "2026-10-26", whatsapp.SituacaoErro, "", mock.Anything).Return(nil)

	_, err := s.svc.Enviar(s.ctx, 1, "2026-10-26")

	s.ErrorIs(err, twilio.ErrCredenciais)
	s.twilio.AssertNumberOfCalls(s.T(), "Enviar", 1)
}

func (s *WhatsAppSuite) TestNaoMandaDiaQueNaoTerminou() {
	s.semConfigNaTela()
	_, err := s.svc.Enviar(s.ctx, 1, "2026-10-27")
	s.ErrorIs(err, whatsapp.ErrDiaFuturo)
}

func (s *WhatsAppSuite) TestEnviarTodosUsaOntemDeBrasilia() {
	s.semNotas()
	s.semConfigNaTela()
	s.clientes.On("Listar", s.ctx).Return([]cliente.Cliente_t{{ID: 1, Nome: "Cliente X"}}, nil)
	s.diario.On("Listar", s.ctx, int64(1), "2026-10-19", "2026-10-26").Return([]diario.Registro_t{}, nil)
	s.contatos.On("Destinos", s.ctx, int64(1)).Return([]contato.Destino_t{}, nil)

	s.Empty(s.svc.EnviarTodos(s.ctx))
	s.diario.AssertExpectations(s.T())
}

func TestWhatsAppSuite(t *testing.T) { suite.Run(t, new(WhatsAppSuite)) }
