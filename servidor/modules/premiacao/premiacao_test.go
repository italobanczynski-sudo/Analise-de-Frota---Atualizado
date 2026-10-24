package premiacao_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/premiacao"
	"analisefrota/modules/premiacao/premiacao_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) SubstituirMes(ctx context.Context, clienteID int64, mes string, rows []premiacao_repo.PremiacaoRow_t) error {
	return m.Called(ctx, clienteID, mes, rows).Error(0)
}
func (m *RepositoryMock_t) Listar(ctx context.Context, clienteID int64, mes string) ([]premiacao_repo.PremiacaoRow_t, error) {
	a := m.Called(ctx, clienteID, mes)
	return a.Get(0).([]premiacao_repo.PremiacaoRow_t), a.Error(1)
}

type PremiacaoSuite struct {
	suite.Suite
	repo *RepositoryMock_t
	svc  premiacao.Service_I
}

func (s *PremiacaoSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.svc = premiacao.NewPremiacaoCom(s.repo)
}

func (s *PremiacaoSuite) TestSalvarAgrupaPorMesEDescartaErros() {
	var recs []gobraxapi.PremiacaoRec_t
	s.Require().NoError(json.Unmarshal([]byte(`[
		{"ID":1,"Name":"Ana","DocumentNumber":"<cpf>","Overview":[{"Date":"09-2026","Score":90,"TotalKM":1000,"Reward":100},{"Date":"10-2026","Score":80,"TotalKm":500,"Reward":50}]},
		{"ID":2,"Name":"Beto","Overview":[{"Date":"09-2026","Score":70,"Err":"sem dados"},{"Date":"2026/09","Score":1}]},
		{"ID":3,"Name":"","Overview":[{"Date":"09-2026","Score":60}]}]`), &recs))
	gravado := map[string][]premiacao_repo.PremiacaoRow_t{}
	s.repo.On("SubstituirMes", mock.Anything, int64(2), mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		gravado[a.String(2)] = a.Get(3).([]premiacao_repo.PremiacaoRow_t)
	}).Return(nil)

	meses, err := s.svc.Salvar(context.Background(), 2, recs)

	s.Require().NoError(err)
	s.Equal([]string{"2026-09", "2026-10"}, meses)
	s.Require().Len(gravado["2026-09"], 1, "Beto teve erro e o sem nome é descartado")
	s.Equal("Ana", gravado["2026-09"][0].Motorista)
	s.Equal(500.0, *gravado["2026-10"][0].Km, "aceita TotalKm")
}

func (s *PremiacaoSuite) TestListarValidaMes() {
	_, err := s.svc.Listar(context.Background(), 1, "10-2026")
	s.ErrorIs(err, premiacao.ErrMesInvalido)
}

func TestPremiacaoSuite(t *testing.T) { suite.Run(t, new(PremiacaoSuite)) }
