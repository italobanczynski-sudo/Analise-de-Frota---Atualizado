// Package mapaosm busca as imagens do fundo do mapa (OpenStreetMap) para o painel. Pelo servidor, o pedido sai com
// User-Agent e Referer da plataforma, como pede a política de uso do OSM: no navegador, a plataforma manda
// Referrer-Policy: no-referrer, e o OSM recusa a imagem ("Access blocked"). As imagens ficam num cache em memória.
package mapaosm

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"analisefrota/environment"
)

const (
	base       = "https://tile.openstreetmap.org/%d/%d/%d.png"
	zoomMax    = 19
	cacheMax   = 2000 // imagens (~20 KB cada)
	tamanhoMax = 512 * 1024
)

var ErrPedido = errors.New("imagem do mapa inválida")

type Mapa_I interface {
	Imagem(ctx context.Context, z, x, y int) ([]byte, error)
}

type item_t struct {
	chave string
	dados []byte
}

type mapa_t struct {
	hc    *http.Client
	ua    string
	ref   string
	mu    sync.Mutex
	ordem *list.List
	idx   map[string]*list.Element
}

var (
	unico     *mapa_t
	unicoOnce sync.Once
)

// NewMapa devolve o cliente único (o cache é compartilhado)
func NewMapa() Mapa_I {
	unicoOnce.Do(func() {
		pub := environment.Cfg.Servidor.URLPublica
		unico = &mapa_t{hc: &http.Client{Timeout: 15 * time.Second}, ua: "AnaliseDeFrota-Gobrax/1.0 (+" + pub + ")", ref: pub + "/",
			ordem: list.New(), idx: map[string]*list.Element{}}
	})
	return unico
}

func (m *mapa_t) Imagem(ctx context.Context, z, x, y int) ([]byte, error) {
	if z < 0 || z > zoomMax || x < 0 || y < 0 || x >= 1<<z || y >= 1<<z {
		return nil, ErrPedido
	}
	chave := fmt.Sprintf("%d/%d/%d", z, x, y)
	m.mu.Lock()
	if el, ok := m.idx[chave]; ok {
		m.ordem.MoveToFront(el)
		d := el.Value.(*item_t).dados
		m.mu.Unlock()
		return d, nil
	}
	m.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(base, z, x, y), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", m.ua)
	req.Header.Set("Referer", m.ref)
	resp, err := m.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mapa: OpenStreetMap respondeu %d", resp.StatusCode)
	}
	dados, err := io.ReadAll(io.LimitReader(resp.Body, tamanhoMax))
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.idx[chave]; !ok {
		m.idx[chave] = m.ordem.PushFront(&item_t{chave: chave, dados: dados})
		for m.ordem.Len() > cacheMax {
			u := m.ordem.Back()
			m.ordem.Remove(u)
			delete(m.idx, u.Value.(*item_t).chave)
		}
	}
	return dados, nil
}
