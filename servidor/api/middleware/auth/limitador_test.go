package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type LimitadorSuite struct {
	suite.Suite
	agora time.Time
}

func (s *LimitadorSuite) SetupTest() {
	travaMu.Lock()
	trava = map[string]*tentativas_t{}
	travaMu.Unlock()
	s.agora = time.Date(2026, 10, 25, 10, 0, 0, 0, time.UTC)
}

func (s *LimitadorSuite) TestCincoErrosDoMesmoEmailTravam() {
	for i := 0; i < 4; i++ {
		registrarErro("ana@gobrax.com", "1.1.1.1", s.agora)
	}
	s.False(bloqueado("ana@gobrax.com", "1.1.1.1", s.agora))
	registrarErro(" ANA@gobrax.com ", "2.2.2.2", s.agora) // mesmo e-mail, de outro IP e escrito diferente
	s.True(bloqueado("ana@gobrax.com", "3.3.3.3", s.agora))
	s.False(bloqueado("bia@gobrax.com", "3.3.3.3", s.agora), "outra pessoa não é afetada")
}

func (s *LimitadorSuite) TestTravaVenceEm15Minutos() {
	for i := 0; i < 5; i++ {
		registrarErro("ana@gobrax.com", "1.1.1.1", s.agora)
	}
	s.True(bloqueado("ana@gobrax.com", "1.1.1.1", s.agora.Add(14*time.Minute)))
	s.False(bloqueado("ana@gobrax.com", "1.1.1.1", s.agora.Add(16*time.Minute)))
}

func (s *LimitadorSuite) TestVinteErrosDoMesmoIPTravamOIP() {
	for i := 0; i < 20; i++ {
		registrarErro(fmt.Sprintf("p%d@gobrax.com", i), "9.9.9.9", s.agora)
	}
	s.True(bloqueado("nova@gobrax.com", "9.9.9.9", s.agora))
	s.False(bloqueado("nova@gobrax.com", "8.8.8.8", s.agora), "outro IP continua entrando")
}

func (s *LimitadorSuite) TestAcertoLimpaSoOEmail() {
	for i := 0; i < 4; i++ {
		registrarErro("ana@gobrax.com", "9.9.9.9", s.agora)
	}
	limparErros("ana@gobrax.com", "9.9.9.9")
	travaMu.Lock()
	_, temEmail := trava["e:ana@gobrax.com"]
	ip := trava["o:9.9.9.9"].erros
	travaMu.Unlock()
	s.False(temEmail)
	s.Equal(4, ip, "o contador do IP continua")
}

func (s *LimitadorSuite) TestJanelasVencidasSaemDaMemoria() {
	for i := 0; i < 1200; i++ {
		registrarErro(fmt.Sprintf("p%d@gobrax.com", i), fmt.Sprintf("10.0.%d.%d", i/250, i%250), s.agora)
	}
	registrarErro("depois@gobrax.com", "1.1.1.1", s.agora.Add(time.Hour))
	travaMu.Lock()
	n := len(trava)
	travaMu.Unlock()
	s.Equal(2, n, "só ficam as chaves da tentativa nova")
}

func TestLimitadorSuite(t *testing.T) { suite.Run(t, new(LimitadorSuite)) }
