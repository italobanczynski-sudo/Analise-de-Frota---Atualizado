// Package gobraxapi chama a API da plataforma Gobrax (gateway-v3). Não tem regra de negócio:
// só monta a chamada, respeita o limite de uso e devolve os registros como vieram.
package gobraxapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/common/utils"
	"analisefrota/environment"
)

var (
	ErrTokenInvalido = errors.New("a API não aceitou o token (token inválido)")
	ErrSemAcesso     = errors.New("o token não tem acesso a este dado")
	ErrSemDados      = errors.New("a API respondeu 404 (sem dado para o período)")
	ErrLimite        = errors.New("a API atingiu o limite de uso deste token")
)

// esperaMaxima: espera pedida pela API (429) que ainda vale aguardar na hora; acima disso a chamada desiste
// e o caminho fica bloqueado até o prazo que a API deu (a busca tenta de novo na próxima rodada)
const esperaMaxima = 3 * time.Minute // a performance pede ~75 s entre chamadas: vale esperar

// ErroLimite_t diz até quando a API pediu para não chamar de novo
type ErroLimite_t struct {
	Caminho string
	Ate     time.Time
}

func (e ErroLimite_t) Error() string {
	return fmt.Sprintf("%v em %s: a API pediu para esperar até %s (Brasília)", ErrLimite, e.Caminho, utils.HoraBR(e.Ate))
}

func (e ErroLimite_t) Unwrap() error { return ErrLimite }

// ErroAPI é qualquer outra resposta de erro (400, 500...)
type ErroAPI_t struct {
	Status    int
	Descricao string
}

func (e ErroAPI_t) Error() string {
	if e.Descricao != "" {
		return fmt.Sprintf("a API respondeu %d (%s)", e.Status, e.Descricao)
	}
	return fmt.Sprintf("a API respondeu %d", e.Status)
}

type Client_I interface {
	Estatisticas_I
	Performance_I
	Odometro_I
	Premiacao_I
}

type Client_t struct {
	http      *http.Client
	base      string
	intervalo time.Duration
	mu        sync.Mutex           // uma chamada por vez: a API aceita rajada de 10 e depois 1 por segundo
	bloqueio  map[string]time.Time // caminho → até quando a API pediu para não chamar (429 com espera longa)
	agora     func() time.Time
}

var (
	localClient Client_I
	clientOnce  sync.Once
)

// NewClient devolve o cliente único configurado pelo environment
func NewClient() Client_I {
	clientOnce.Do(func() {
		localClient = NewClientCom(environment.Cfg.Gobrax.APIBase, &http.Client{Timeout: 60 * time.Second}, time.Duration(environment.Cfg.Gobrax.IntervaloMs)*time.Millisecond)
	})
	return localClient
}

// NewClientCom monta um cliente com base e HTTP próprios (usado nos testes)
func NewClientCom(base string, hc *http.Client, intervalo time.Duration) Client_I {
	return &Client_t{http: hc, base: base, intervalo: intervalo, bloqueio: map[string]time.Time{}, agora: time.Now}
}

// get faz o GET autenticado e decodifica a resposta em destino; repete quando a API pede para esperar (429)
func (c *Client_t) get(ctx context.Context, caminho string, params url.Values, token string, destino any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ate, ok := c.bloqueio[caminho]; ok && c.agora().Before(ate) {
		return ErroLimite_t{Caminho: caminho, Ate: ate}
	}
	for tentativa := 0; ; tentativa++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+caminho+"?"+params.Encode(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		inicio := time.Now()
		resp, err := c.http.Do(req)
		if err != nil {
			logrus.Warnf("gobraxapi: %s %s: sem resposta em %s: %v", caminho, params.Get("startDate"), time.Since(inicio).Round(time.Millisecond), err)
			return fmt.Errorf("sem resposta da API: %w", err)
		}
		corpo, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		// uma linha por chamada (sem o token): mostra onde a busca demora ou falha
		logrus.Infof("gobraxapi: %s %s → %d em %s (%d KB)", caminho, params.Get("startDate"), resp.StatusCode, time.Since(inicio).Round(time.Millisecond), len(corpo)/1024)
		if resp.StatusCode == http.StatusTooManyRequests && tentativa < 8 {
			espera, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			if espera < 1 {
				espera = 1
			}
			logrus.Warnf("gobraxapi: %s: limite de uso (429), a API pediu %ds de espera (tentativa %d de 8)", caminho, espera, tentativa+1)
			if time.Duration(espera)*time.Second > esperaMaxima {
				ate := c.agora().Add(time.Duration(espera) * time.Second)
				c.bloqueio[caminho] = ate
				return ErroLimite_t{Caminho: caminho, Ate: ate}
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(espera)*time.Second + 100*time.Millisecond):
			}
			continue
		}
		if resp.StatusCode >= 300 {
			return erroDaResposta(resp.StatusCode, corpo)
		}
		time.Sleep(c.intervalo)
		if err := json.Unmarshal(corpo, destino); err != nil {
			return fmt.Errorf("resposta da API em formato inesperado: %w", err)
		}
		return nil
	}
}

// Os erros vêm como {"success": false, "data": {"code", "description"}} ou {"errorCode", "errorDescription"}
func erroDaResposta(status int, corpo []byte) error {
	var env struct {
		Data             json.RawMessage `json:"data"`
		Code             json.Number     `json:"code"`
		Description      string          `json:"description"`
		ErrorDescription string          `json:"errorDescription"`
	}
	_ = json.Unmarshal(corpo, &env)
	var dados struct {
		Code        json.Number `json:"code"`
		Description string      `json:"description"`
	}
	_ = json.Unmarshal(env.Data, &dados)
	cod := dados.Code.String()
	if cod == "" {
		cod = env.Code.String()
	}
	desc := dados.Description
	if desc == "" {
		desc = env.Description
	}
	if desc == "" {
		desc = env.ErrorDescription
	}
	switch {
	case cod == "61" || status == http.StatusUnauthorized:
		return ErrTokenInvalido
	case cod == "71" || status == http.StatusForbidden:
		return ErrSemAcesso
	case status == http.StatusNotFound:
		return ErrSemDados
	}
	return ErroAPI_t{Status: status, Descricao: desc}
}

// janela monta os parâmetros startDate/endDate (UTC) com todos os veículos do token
func janela(ini, fim time.Time) url.Values {
	return url.Values{
		"startDate":             {ini.UTC().Format("2006-01-02 15:04:05")},
		"endDate":               {fim.UTC().Format("2006-01-02 15:04:05")},
		"vehicleIdentification": {""},
	}
}
