package twilio_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/twilio"
	"analisefrota/environment"
)

type TwilioSuite struct {
	suite.Suite
}

func (s *TwilioSuite) cliente(h http.HandlerFunc) *twilio.Twilio_t {
	srv := httptest.NewServer(h)
	s.T().Cleanup(srv.Close)
	return &twilio.Twilio_t{
		Cfg: environment.WhatsAppConfig_t{TwilioAccountSID: "<account-sid>", TwilioAPIKey: "<api-key>", TwilioAPISecret: "<api-secret>",
			Remetente: "whatsapp:+550000000000", TemplateSID: "<template-sid>"},
		Base: srv.URL + "/", HTTP: srv.Client(),
	}
}

func (s *TwilioSuite) TestEnviaTemplateComVariaveis() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/<account-sid>/Messages.json", r.URL.Path)
		u, p, ok := r.BasicAuth()
		s.True(ok)
		s.Equal("<api-key>", u)
		s.Equal("<api-secret>", p)
		s.Require().NoError(r.ParseForm())
		s.Equal("whatsapp:+5511900000000", r.PostForm.Get("To"))
		s.Equal("whatsapp:+550000000000", r.PostForm.Get("From"))
		s.Equal("<template-sid>", r.PostForm.Get("ContentSid"))
		var vars map[string]string
		s.Require().NoError(json.Unmarshal([]byte(r.PostForm.Get("ContentVariables")), &vars))
		s.Equal("Ana", vars["1"])
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sid":"SM123","status":"queued"}`))
	})

	sid, err := c.Enviar(context.Background(), "+5511900000000", map[string]string{"1": "Ana"}, "texto")

	s.Require().NoError(err)
	s.Equal("SM123", sid)
}

func (s *TwilioSuite) TestRecusaTrazAMensagemDaTwilio() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":63016,"message":"Failed to send freeform message"}`))
	})
	_, err := c.Enviar(context.Background(), "+5511900000000", nil, "")
	s.ErrorContains(err, "63016")
}

func (s *TwilioSuite) TestCredencialRecusada() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	_, err := c.Enviar(context.Background(), "+5511900000000", nil, "")
	s.True(errors.Is(err, twilio.ErrCredenciais))
}

func (s *TwilioSuite) TestSemTemplateMandaOTexto() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) {
		s.Require().NoError(r.ParseForm())
		s.Equal("Bom dia, Ana!", r.PostForm.Get("Body"))
		s.Empty(r.PostForm.Get("ContentSid"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sid":"SM9"}`))
	})
	c.Cfg.TemplateSID = ""
	sid, err := c.Enviar(context.Background(), "+5511900000000", map[string]string{"1": "Ana"}, "Bom dia, Ana!")
	s.Require().NoError(err)
	s.Equal("SM9", sid)
}

func (s *TwilioSuite) TestRecusaExplicaOCodigo() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":63015,"message":"Channel Sandbox can only send messages to phone numbers that have joined the Sandbox"}`))
	})
	_, err := c.Enviar(context.Background(), "+5511900000000", nil, "")
	s.ErrorIs(err, twilio.ErrRecusa)
	s.ErrorContains(err, "não entrou no Sandbox")
}

func (s *TwilioSuite) TestStatusDaMensagem() {
	c := s.cliente(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/<account-sid>/Messages/SM1.json", r.URL.Path)
		_, _ = w.Write([]byte(`{"sid":"SM1","status":"undelivered","error_code":63016}`))
	})
	st, codigo, err := c.Status(context.Background(), "SM1")
	s.Require().NoError(err)
	s.Equal("undelivered", st)
	s.Equal(63016, codigo)
}

func (s *TwilioSuite) TestMascarar() {
	s.Equal("••••0000", twilio.Mascarar("+5511900000000"))
	s.Equal("****", twilio.Mascarar("12"))
}

func TestTwilioSuite(t *testing.T) { suite.Run(t, new(TwilioSuite)) }
