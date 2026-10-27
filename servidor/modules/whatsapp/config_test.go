package whatsapp_test

import (
	"time"

	"github.com/stretchr/testify/mock"

	"analisefrota/common/cripto"
	"analisefrota/modules/cliente"
	"analisefrota/modules/contato"
	"analisefrota/modules/diario"
	"analisefrota/modules/whatsapp"
	"analisefrota/modules/whatsapp/whatsapp_repo"
)

// valores de mentira no formato da Twilio
const (
	sidTeste      = "AC00000000000000000000000000000000"
	tokenTeste    = "00000000000000000000000000000000"
	templateTeste = "HX00000000000000000000000000000000"
)

func (s *WhatsAppSuite) novaConfig() whatsapp.NovaConfig_t {
	return whatsapp.NovaConfig_t{AccountSID: sidTeste, AuthToken: tokenTeste, Remetente: "+1 415 523 8886", TemplateSID: templateTeste, Hora: 7}
}

func (s *WhatsAppSuite) TestSalvarConfigCifraOTokenENormalizaORemetente() {
	var gravado whatsapp_repo.ConfigRow_t
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{}, false, nil).Once()
	s.repo.On("SalvarConfig", s.ctx, mock.Anything).Run(func(a mock.Arguments) { gravado = a.Get(1).(whatsapp_repo.ConfigRow_t) }).Return(nil)
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{AccountSID: sidTeste, AuthTokenCifrado: []byte{1}, AtualizadoEm: time.Now()}, true, nil)

	cfg, err := s.svc.SalvarConfig(s.ctx, s.novaConfig(), 3)

	s.Require().NoError(err)
	s.Equal("whatsapp:+14155238886", gravado.Remetente)
	s.NotContains(string(gravado.AuthTokenCifrado), tokenTeste)
	tk, err := cripto.Decifrar(s.chave, gravado.AuthTokenCifrado)
	s.Require().NoError(err)
	s.Equal(tokenTeste, tk)
	s.Equal(int64(3), *gravado.AtualizadoPor)
	s.True(cfg.TemToken)
	s.Equal("tela", cfg.Origem)
}

func (s *WhatsAppSuite) TestSalvarConfigSemTokenMantemOSalvo() {
	salvo, _ := cripto.Cifrar(s.chave, tokenTeste)
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{AccountSID: sidTeste, AuthTokenCifrado: salvo}, true, nil)
	s.repo.On("SalvarConfig", s.ctx, mock.MatchedBy(func(r whatsapp_repo.ConfigRow_t) bool { return string(r.AuthTokenCifrado) == string(salvo) })).Return(nil)

	n := s.novaConfig()
	n.AuthToken = ""
	_, err := s.svc.SalvarConfig(s.ctx, n, 3)
	s.Require().NoError(err)
}

func (s *WhatsAppSuite) TestSalvarConfigValida() {
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{}, false, nil)
	casos := map[error]func(n *whatsapp.NovaConfig_t){
		whatsapp.ErrAccountSID:  func(n *whatsapp.NovaConfig_t) { n.AccountSID = "SK123" },
		whatsapp.ErrTemplateSID: func(n *whatsapp.NovaConfig_t) { n.TemplateSID = "abc" },
		whatsapp.ErrRemetente:   func(n *whatsapp.NovaConfig_t) { n.Remetente = "123" },
		whatsapp.ErrHora:        func(n *whatsapp.NovaConfig_t) { n.Hora = 24 },
		whatsapp.ErrAuthToken:   func(n *whatsapp.NovaConfig_t) { n.AuthToken = "curto" },
		whatsapp.ErrSemToken:    func(n *whatsapp.NovaConfig_t) { n.AuthToken = "" },
	}
	for esperado, mudar := range casos {
		n := s.novaConfig()
		mudar(&n)
		_, err := s.svc.SalvarConfig(s.ctx, n, 3)
		s.ErrorIs(err, esperado)
	}
	s.repo.AssertNotCalled(s.T(), "SalvarConfig", mock.Anything, mock.Anything)
}

func (s *WhatsAppSuite) TestEnvioUsaAConfigDaTelaComOToken() {
	salvo, _ := cripto.Cifrar(s.chave, tokenTeste)
	s.repo.On("BuscarConfig", s.ctx).Return(whatsapp_repo.ConfigRow_t{AccountSID: sidTeste, AuthTokenCifrado: salvo,
		Remetente: "whatsapp:+14155238886", TemplateSID: templateTeste, Hora: 7}, true, nil)
	s.semNotas()
	s.diario.On("Listar", s.ctx, int64(1), mock.Anything, mock.Anything).Return([]diario.Registro_t{}, nil)
	s.contatos.On("Destinos", s.ctx, int64(1)).Return([]contato.Destino_t{}, nil).Maybe()

	_, err := s.svc.Enviar(s.ctx, 1, "2026-10-26")

	s.Require().NoError(err)
	s.Equal(sidTeste, s.cfgUsada.TwilioAccountSID)
	s.Equal(sidTeste, s.cfgUsada.TwilioAPIKey, "com Auth Token, o usuário da API é o próprio Account SID")
	s.Equal(tokenTeste, s.cfgUsada.TwilioAPISecret)
	s.True(s.cfgUsada.Configurado())
}

func (s *WhatsAppSuite) TestAgendadoSoNaHoraMarcadaEComAutomaticoLigado() {
	salvo, _ := cripto.Cifrar(s.chave, tokenTeste)
	cfg := whatsapp_repo.ConfigRow_t{AccountSID: sidTeste, AuthTokenCifrado: salvo, Remetente: "whatsapp:+14155238886", TemplateSID: templateTeste, Hora: 7}

	s.repo.On("BuscarConfig", s.ctx).Return(cfg, true, nil).Once() // desligado
	s.Nil(s.svc.EnviarAgendado(s.ctx))

	cfg.Automatico, cfg.Hora = true, 8
	s.repo.On("BuscarConfig", s.ctx).Return(cfg, true, nil).Once() // ligado, mas são 7h
	s.Nil(s.svc.EnviarAgendado(s.ctx))
	s.clientes.AssertNotCalled(s.T(), "Listar", mock.Anything)

	cfg.Hora = 7
	s.repo.On("BuscarConfig", s.ctx).Return(cfg, true, nil)
	s.clientes.On("Listar", s.ctx).Return([]cliente.Cliente_t{}, nil)
	s.svc.EnviarAgendado(s.ctx)
	s.clientes.AssertCalled(s.T(), "Listar", s.ctx)
}
