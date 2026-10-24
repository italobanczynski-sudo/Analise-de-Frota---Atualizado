package odometro_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/odometro"
	"analisefrota/modules/odometro/odometro_repo"
)

type RepositoryMock_t struct{ mock.Mock }

func (m *RepositoryMock_t) Substituir(ctx context.Context, clienteID int64, rows []odometro_repo.OdometroRow_t) error {
	return m.Called(ctx, clienteID, rows).Error(0)
}
func (m *RepositoryMock_t) Listar(ctx context.Context, clienteID int64) ([]odometro_repo.OdometroRow_t, error) {
	a := m.Called(ctx, clienteID)
	return a.Get(0).([]odometro_repo.OdometroRow_t), a.Error(1)
}

type OdometroSuite struct {
	suite.Suite
	repo *RepositoryMock_t
	svc  odometro.Service_I
}

func (s *OdometroSuite) SetupTest() {
	s.repo = new(RepositoryMock_t)
	s.svc = odometro.NewOdometroCom(s.repo)
}

func (s *OdometroSuite) TestSubstituirIgnoraSemPlacaERepetidos() {
	var recs []gobraxapi.OdometroRec_t
	s.Require().NoError(json.Unmarshal([]byte(`[
		{"vehicleIdentification":"abc1d23","odometer":"291862,2","lastUpdated":"2026-10-23 21:06:10"},
		{"vehicleIdentification":"ABC-1D23","odometer":1},
		{"vehicleIdentification":"","odometer":5}]`), &recs))
	var gravados []odometro_repo.OdometroRow_t
	s.repo.On("Substituir", mock.Anything, int64(4), mock.Anything).Run(func(a mock.Arguments) { gravados = a.Get(2).([]odometro_repo.OdometroRow_t) }).Return(nil)

	n, err := s.svc.Substituir(context.Background(), 4, recs)

	s.Require().NoError(err)
	s.Equal(1, n)
	s.Require().Len(gravados, 1)
	s.Equal("ABC1D23", gravados[0].Placa)
	s.Equal(291862.2, *gravados[0].Odometro)
	s.Equal("2026-10-23 21:06:10", gravados[0].Coletado)
}

func TestOdometroSuite(t *testing.T) { suite.Run(t, new(OdometroSuite)) }
