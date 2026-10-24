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

	"analisefrota/environment"
)

var (
	ErrTokenInvalido = errors.New("a API não aceitou o token (token inválido)")
	ErrSemAcesso     = errors.New("o token não tem acesso a este dado")
	ErrSemDados      = errors.New("a API respondeu 404 (sem dado para o período)")
)

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
	mu        sync.Mutex // uma chamada por vez: a API aceita rajada de 10 e depois 1 por segundo
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
	return &Client_t{http: hc, base: base, intervalo: intervalo}
}

// get faz o GET autenticado e decodifica a resposta em destino; repete quando a API pede para esperar (429)
func (c *Client_t) get(ctx context.Context, caminho string, params url.Values, token string, destino any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for tentativa := 0; ; tentativa++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+caminho+"?"+params.Encode(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("sem resposta da API: %w", err)
		}
		corpo, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests && tentativa < 8 {
			espera, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			if espera < 1 {
				espera = 1
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
