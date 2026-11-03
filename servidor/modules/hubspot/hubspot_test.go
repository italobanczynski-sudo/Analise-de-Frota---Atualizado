package hubspot_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/hubspotapi"
	"analisefrota/modules/hubspot"
)

type APIMock_t struct{ mock.Mock }

func (m *APIMock_t) BuscarEmpresas(ctx context.Context, texto string) ([]hubspotapi.Objeto_t, error) {
	a := m.Called(ctx, texto)
	return a.Get(0).([]hubspotapi.Objeto_t), a.Error(1)
}
func (m *APIMock_t) Objeto(ctx context.Context, tipo, id string, props []string) (hubspotapi.Objeto_t, error) {
	a := m.Called(ctx, tipo, id, props)
	return a.Get(0).(hubspotapi.Objeto_t), a.Error(1)
}
func (m *APIMock_t) Associados(ctx context.Context, de, id, para string) ([]string, error) {
	a := m.Called(ctx, de, id, para)
	return a.Get(0).([]string), a.Error(1)
}
func (m *APIMock_t) LerLote(ctx context.Context, tipo string, ids, props []string) ([]hubspotapi.Objeto_t, error) {
	a := m.Called(ctx, tipo, ids, props)
	return a.Get(0).([]hubspotapi.Objeto_t), a.Error(1)
}

type HubspotSuite struct {
	suite.Suite
	api   *APIMock_t
	agora time.Time
	svc   hubspot.Service_I
}

func (s *HubspotSuite) SetupTest() {
	s.api = new(APIMock_t)
	s.agora = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	s.svc = hubspot.NewHubspotCom(s.api, 10*time.Minute, func() time.Time { return s.agora })
}

func obj(id string, p map[string]string) hubspotapi.Objeto_t {
	return hubspotapi.Objeto_t{ID: id, Properties: p}
}

// empresa 10 com 2 chamados (1 aberto), 1 nota com dados pessoais, 1 contato e nenhuma ligação/e-mail/reunião
func (s *HubspotSuite) prepararEmpresa() {
	s.api.On("Objeto", mock.Anything, "companies", "10", mock.Anything).Return(obj("10", map[string]string{"name": "Teste Transportes", "nota_do_chs": "72"}), nil)
	vazio := []string{}
	for _, t := range []string{"calls", "emails", "meetings"} {
		s.api.On("Associados", mock.Anything, "companies", "10", t).Return(vazio, nil)
		s.api.On("LerLote", mock.Anything, t, vazio, mock.Anything).Return([]hubspotapi.Objeto_t{}, nil)
	}
	s.api.On("Associados", mock.Anything, "companies", "10", "notes").Return([]string{"n1"}, nil)
	s.api.On("LerLote", mock.Anything, "notes", []string{"n1"}, mock.Anything).Return([]hubspotapi.Objeto_t{
		obj("n1", map[string]string{"hs_timestamp": "2026-10-01", "hs_note_body": "<p>Ligar para fulano@exemplo.com</p><p>CPF 123.456.789-09</p>"}),
	}, nil)
	s.api.On("Associados", mock.Anything, "companies", "10", "tickets").Return([]string{"t1", "t2"}, nil)
	s.api.On("LerLote", mock.Anything, "tickets", []string{"t1", "t2"}, mock.Anything).Return([]hubspotapi.Objeto_t{
		obj("t1", map[string]string{"subject": "Equipamento offline", "createdate": "2026-09-01", "closed_date": "2026-09-03"}),
		obj("t2", map[string]string{"subject": "Dúvida no app", "createdate": "2026-10-20"}),
	}, nil)
	s.api.On("Associados", mock.Anything, "companies", "10", "contacts").Return([]string{"c1"}, nil)
	s.api.On("LerLote", mock.Anything, "contacts", []string{"c1"}, mock.Anything).Return([]hubspotapi.Objeto_t{
		obj("c1", map[string]string{"firstname": "Contato", "email": "nao-deve-sair@exemplo.com"}),
	}, nil)
	s.api.On("Associados", mock.Anything, "contacts", "c1", "calls").Return([]string{"x", "y"}, nil)
	s.api.On("Associados", mock.Anything, "contacts", "c1", "emails").Return([]string{"z"}, nil)
}

func (s *HubspotSuite) TestDesligadoSemToken() {
	svc := hubspot.NewHubspotCom(nil, time.Minute, time.Now)
	_, err := svc.Resumo(context.Background(), "Teste", "", false)
	s.ErrorIs(err, hubspot.ErrDesligado)
}

func (s *HubspotSuite) TestResumoMontaTotaisEMascaraDadosPessoais() {
	s.prepararEmpresa()
	r, err := s.svc.Resumo(context.Background(), "Teste Transportes", "10", false)
	s.Require().NoError(err)
	s.Equal("gobrax-hubspot-1", r.Formato)
	s.Equal("72", r.Comp["nota_do_chs"])
	s.Equal(2, r.Tickets.Total)
	s.Equal("t2", r.Tickets.Results[0].ID, "chamado mais recente primeiro")
	s.Require().NotNil(r.Abertos)
	s.Equal(1, r.Abertos.Total)
	nota := r.Notas.Results[0].Properties["hs_note_body"]
	s.NotContains(nota, "@exemplo.com")
	s.NotContains(nota, "123.456.789-09")
	s.NotContains(nota, "<p>")
	_, temEmail := r.Contatos.Results[0].Properties["email"]
	s.False(temEmail, "e-mail do contato não sai do HubSpot")
	s.Equal([]hubspot.PorContato_t{{ID: "c1", L: 2, E: 1}}, r.PorContato)
}

func (s *HubspotSuite) TestLeituraGuardadaAteVencer() {
	s.prepararEmpresa()
	_, err := s.svc.Resumo(context.Background(), "Teste Transportes", "10", false)
	s.Require().NoError(err)
	_, err = s.svc.Resumo(context.Background(), "Teste Transportes", "10", false)
	s.Require().NoError(err)
	s.api.AssertNumberOfCalls(s.T(), "Objeto", 1)
	s.agora = s.agora.Add(11 * time.Minute)
	_, err = s.svc.Resumo(context.Background(), "Teste Transportes", "10", false)
	s.Require().NoError(err)
	s.api.AssertNumberOfCalls(s.T(), "Objeto", 2)
	_, err = s.svc.Resumo(context.Background(), "Teste Transportes", "10", true)
	s.Require().NoError(err)
	s.api.AssertNumberOfCalls(s.T(), "Objeto", 3)
}

func (s *HubspotSuite) TestProcuraPeloNomeSemID() {
	s.api.On("BuscarEmpresas", mock.Anything, "Teste Transportes").Return([]hubspotapi.Objeto_t{
		obj("99", map[string]string{"name": "Outra Empresa"}), obj("10", map[string]string{"name": "TESTE TRANSPORTES LTDA"}),
	}, nil)
	s.prepararEmpresa()
	r, err := s.svc.Resumo(context.Background(), "Teste Transportes Ltda", "", false)
	s.Require().NoError(err)
	s.Equal("10", r.ID)
}

func (s *HubspotSuite) TestEmpresaNaoEncontrada() {
	s.api.On("BuscarEmpresas", mock.Anything, mock.Anything).Return([]hubspotapi.Objeto_t{}, nil)
	_, err := s.svc.Resumo(context.Background(), "Inexistente", "", false)
	s.ErrorIs(err, hubspot.ErrNaoEncontrado)
}

func (s *HubspotSuite) TestSemEscopoDeChamadosFicaSoOTotal() {
	s.api.On("Objeto", mock.Anything, "companies", "10", mock.Anything).Return(obj("10", map[string]string{"name": "X"}), nil)
	for _, t := range []string{"calls", "emails", "meetings", "notes", "contacts"} {
		s.api.On("Associados", mock.Anything, "companies", "10", t).Return([]string{}, nil)
		s.api.On("LerLote", mock.Anything, t, []string{}, mock.Anything).Return([]hubspotapi.Objeto_t{}, nil)
	}
	s.api.On("Associados", mock.Anything, "companies", "10", "tickets").Return([]string{"t1", "t2", "t3"}, nil)
	s.api.On("LerLote", mock.Anything, "tickets", mock.Anything, mock.Anything).Return([]hubspotapi.Objeto_t(nil), hubspotapi.ErrSemEscopo)
	r, err := s.svc.Resumo(context.Background(), "X", "10", false)
	s.Require().NoError(err)
	s.Equal(3, r.Tickets.Total)
	s.True(r.Tickets.SemDetalhe)
	s.Nil(r.Abertos, "sem permissão: abertos desconhecido, não zero")
	s.Equal([]string{"tickets"}, r.SemDetalhe)
}

func TestHubspotSuite(t *testing.T) { suite.Run(t, new(HubspotSuite)) }
