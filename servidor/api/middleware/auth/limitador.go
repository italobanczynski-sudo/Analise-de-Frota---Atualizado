package auth

import (
	"strings"
	"sync"
	"time"
)

// Trava de tentativas, por 15 minutos: 5 erros seguidos do mesmo e-mail, ou 20 do mesmo endereço IP
// (o limite por IP é maior porque várias pessoas podem chegar pelo mesmo IP, como dentro do Docker ou num escritório).
const janelaTrava = 15 * time.Minute

type tentativas_t struct {
	erros    int
	primeira time.Time
}

type chave_t struct {
	k      string
	limite int
}

var (
	travaMu sync.Mutex
	trava   = map[string]*tentativas_t{}
)

func chavesTrava(email, origem string) []chave_t {
	return []chave_t{{"e:" + strings.ToLower(strings.TrimSpace(email)), 5}, {"o:" + origem, 20}}
}

func bloqueado(email, origem string, agora time.Time) bool {
	travaMu.Lock()
	defer travaMu.Unlock()
	for _, c := range chavesTrava(email, origem) {
		if t, ok := trava[c.k]; ok {
			if agora.Sub(t.primeira) > janelaTrava {
				delete(trava, c.k)
			} else if t.erros >= c.limite {
				return true
			}
		}
	}
	return false
}

func registrarErro(email, origem string, agora time.Time) {
	travaMu.Lock()
	defer travaMu.Unlock()
	// tira as janelas já vencidas antes de crescer (muitos e-mails diferentes não enchem a memória)
	if len(trava) >= 1000 {
		for k, t := range trava {
			if agora.Sub(t.primeira) > janelaTrava {
				delete(trava, k)
			}
		}
	}
	for _, c := range chavesTrava(email, origem) {
		t, ok := trava[c.k]
		if !ok || agora.Sub(t.primeira) > janelaTrava {
			t = &tentativas_t{primeira: agora}
			trava[c.k] = t
		}
		t.erros++
	}
}

// limparErros zera só o e-mail que acertou; o contador do IP continua
func limparErros(email, origem string) {
	travaMu.Lock()
	defer travaMu.Unlock()
	delete(trava, chavesTrava(email, origem)[0].k)
}
