// Package hubspotapi chama a API oficial do HubSpot (CRM v3/v4), só leitura. Não tem regra de negócio:
// monta a chamada, troca a chave de acesso pessoal por token quando preciso, respeita o limite (429) e devolve os objetos.
package hubspotapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/environment"
)

const baseHubspot = "https://api.hubapi.com"

var (
	ErrTokenInvalido = errors.New("o HubSpot não aceitou o token (HUBSPOT_TOKEN inválido ou expirado)")
	ErrSemEscopo     = errors.New("o token do HubSpot não tem permissão de leitura para este dado")
)

// ErroAPI_t é qualquer outra resposta de erro do HubSpot
type ErroAPI_t struct {
	Status  int
	Caminho string
}

func (e ErroAPI_t) Error() string {
	return fmt.Sprintf("o HubSpot respondeu %d em %s", e.Status, e.Caminho)
}

// Objeto_t é um registro do CRM: id e propriedades (o HubSpot manda tudo como texto)
type Objeto_t struct {
	ID         string            `json:"id"`
	Properties map[string]string `json:"properties"`
}

type Client_I interface {
	// BuscarEmpresas procura empresas pelo nome
	BuscarEmpresas(ctx context.Context, texto string) ([]Objeto_t, error)
	// Objeto lê um registro com as propriedades pedidas
	Objeto(ctx context.Context, tipo, id string, props []string) (Objeto_t, error)
	// Associados devolve os ids de um tipo ligados ao registro (todas as páginas)
	Associados(ctx context.Context, de, id, para string) ([]string, error)
	// LerLote lê vários registros de um tipo (de 100 em 100)
	LerLote(ctx context.Context, tipo string, ids, props []string) ([]Objeto_t, error)
}

type Client_t struct {
	http  *http.Client
	base  string
	chave string

	mu        sync.Mutex // guarda o token de acesso trocado a partir da chave pessoal
	acesso    string
	acessoAte time.Time
}

var (
	localClient Client_I
	clientOnce  sync.Once
)

// NewClient devolve o cliente único configurado pelo environment
func NewClient() Client_I {
	clientOnce.Do(func() {
		localClient = NewClientCom(baseHubspot, &http.Client{Timeout: 30 * time.Second}, environment.Cfg.Hubspot.Token)
	})
	return localClient
}

// NewClientCom monta um cliente com base, HTTP e chave próprios (usado nos testes)
func NewClientCom(base string, hc *http.Client, chave string) Client_I {
	return &Client_t{http: hc, base: base, chave: limparChave(chave)}
}

func limparChave(c string) string {
	c = strings.TrimSpace(c)
	c = strings.TrimPrefix(strings.TrimPrefix(c, "Bearer "), "bearer ")
	return strings.TrimSpace(c)
}

// token: o do app privado (pat-...) vai direto; a chave de acesso pessoal é trocada por um token de 30 min e renovada sozinha
func (c *Client_t) token(ctx context.Context) (string, error) {
	if strings.HasPrefix(c.chave, "pat-") {
		return c.chave, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.acesso != "" && time.Now().Before(c.acessoAte.Add(-2*time.Minute)) {
		return c.acesso, nil
	}
	corpo, _ := json.Marshal(map[string]string{"encodedOAuthRefreshToken": c.chave})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/localdevauth/v1/auth/refresh", bytes.NewReader(corpo))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("sem resposta do HubSpot: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", ErrTokenInvalido
	}
	var j struct {
		Token    string `json:"oauthAccessToken"`
		ExpiraMs int64  `json:"expiresAtMillis"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&j); err != nil || j.Token == "" {
		return "", ErrTokenInvalido
	}
	c.acesso = j.Token
	c.acessoAte = time.Now().Add(25 * time.Minute)
	if j.ExpiraMs > 0 {
		c.acessoAte = time.UnixMilli(j.ExpiraMs)
	}
	return c.acesso, nil
}

// chamar faz a requisição autenticada e decodifica a resposta; repete quando o HubSpot pede para esperar (429)
func (c *Client_t) chamar(ctx context.Context, metodo, caminho string, corpo any, destino any) error {
	var dados []byte
	if corpo != nil {
		dados, _ = json.Marshal(corpo)
	}
	for tentativa := 0; ; tentativa++ {
		tok, err := c.token(ctx)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, metodo, c.base+caminho, bytes.NewReader(dados))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("sem resposta do HubSpot: %w", err)
		}
		resposta, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests && tentativa < 8 {
			espera, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			if espera < 1 {
				espera = 1
			}
			logrus.Warnf("hubspotapi: limite de uso (429) em %s, esperando %ds", rota(caminho), espera)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(espera)*time.Second + 200*time.Millisecond):
			}
			continue
		}
		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			return ErrTokenInvalido
		case resp.StatusCode == http.StatusForbidden:
			return ErrSemEscopo
		case resp.StatusCode >= 300:
			return ErroAPI_t{Status: resp.StatusCode, Caminho: rota(caminho)}
		}
		if destino == nil || len(resposta) == 0 {
			return nil
		}
		if err := json.Unmarshal(resposta, destino); err != nil {
			return fmt.Errorf("resposta do HubSpot em formato inesperado: %w", err)
		}
		return nil
	}
}

// rota tira a query string (só para log e erro)
func rota(caminho string) string { return strings.SplitN(caminho, "?", 2)[0] }
