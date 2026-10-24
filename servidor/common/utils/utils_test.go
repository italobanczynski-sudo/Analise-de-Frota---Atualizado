package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"analisefrota/common/utils"
)

type UtilsSuite struct{ suite.Suite }

func (s *UtilsSuite) TestJanelaDiaBR() {
	ini, fim, err := utils.JanelaDiaBR("2026-10-23")
	s.Require().NoError(err)
	s.Equal("2026-10-23 03:00:00", utils.FormatoAPI(ini))
	s.Equal("2026-10-24 02:59:59", utils.FormatoAPI(fim))
}

func (s *UtilsSuite) TestDiaBRViraAsVinteEUmaUTC() {
	s.Equal("2026-10-23", utils.DiaBR(time.Date(2026, 10, 24, 2, 59, 0, 0, time.UTC)))
	s.Equal("2026-10-24", utils.DiaBR(time.Date(2026, 10, 24, 3, 0, 0, 0, time.UTC)))
}

func (s *UtilsSuite) TestDiasEntre() {
	s.Equal([]string{"2026-10-30", "2026-10-31", "2026-11-01"}, utils.DiasEntre("2026-10-30", "2026-11-01"))
	s.Empty(utils.DiasEntre("2026-11-02", "2026-11-01"))
}

func (s *UtilsSuite) TestLimparToken() {
	s.Equal("abc.def", utils.LimparToken("Authorization: Bearer \"abc.def\"\n"))
	s.Equal("abc", utils.LimparToken("  bearer abc "))
}

func (s *UtilsSuite) TestChavePlaca() {
	s.Equal("ABC1234", utils.ChavePlaca(" abc-1234 "))
}

func (s *UtilsSuite) TestInstanteAPI() {
	t, ok := utils.InstanteAPI("2026-10-22T20:01:00+0000")
	s.True(ok)
	s.Equal(20, t.UTC().Hour())
}

func TestUtilsSuite(t *testing.T) { suite.Run(t, new(UtilsSuite)) }
