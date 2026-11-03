package hubspotapi

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// objetoBruto: as propriedades vêm como texto ou null
type objetoBruto struct {
	ID         string             `json:"id"`
	Properties map[string]*string `json:"properties"`
}

func (o objetoBruto) limpo() Objeto_t {
	p := make(map[string]string, len(o.Properties))
	for k, v := range o.Properties {
		if v != nil {
			p[k] = *v
		}
	}
	return Objeto_t{ID: o.ID, Properties: p}
}

func (c *Client_t) BuscarEmpresas(ctx context.Context, texto string) ([]Objeto_t, error) {
	var r struct {
		Results []objetoBruto `json:"results"`
	}
	corpo := map[string]any{"query": texto, "properties": []string{"name"}, "limit": 10}
	if err := c.chamar(ctx, "POST", "/crm/v3/objects/companies/search", corpo, &r); err != nil {
		return nil, err
	}
	out := make([]Objeto_t, len(r.Results))
	for i, o := range r.Results {
		out[i] = o.limpo()
	}
	return out, nil
}

func (c *Client_t) Objeto(ctx context.Context, tipo, id string, props []string) (Objeto_t, error) {
	var o objetoBruto
	caminho := fmt.Sprintf("/crm/v3/objects/%s/%s?properties=%s", tipo, url.PathEscape(id), url.QueryEscape(strings.Join(props, ",")))
	if err := c.chamar(ctx, "GET", caminho, nil, &o); err != nil {
		return Objeto_t{}, err
	}
	return o.limpo(), nil
}

func (c *Client_t) Associados(ctx context.Context, de, id, para string) ([]string, error) {
	ids := []string{}
	after := ""
	for {
		caminho := fmt.Sprintf("/crm/v4/objects/%s/%s/associations/%s?limit=500", de, url.PathEscape(id), para)
		if after != "" {
			caminho += "&after=" + url.QueryEscape(after)
		}
		var r struct {
			Results []struct {
				ToObjectID any `json:"toObjectId"`
			} `json:"results"`
			Paging struct {
				Next struct {
					After string `json:"after"`
				} `json:"next"`
			} `json:"paging"`
		}
		if err := c.chamar(ctx, "GET", caminho, nil, &r); err != nil {
			return nil, err
		}
		for _, x := range r.Results {
			ids = append(ids, fmt.Sprint(x.ToObjectID))
		}
		if r.Paging.Next.After == "" {
			return ids, nil
		}
		after = r.Paging.Next.After
	}
}

func (c *Client_t) LerLote(ctx context.Context, tipo string, ids, props []string) ([]Objeto_t, error) {
	out := []Objeto_t{}
	for i := 0; i < len(ids); i += 100 {
		fim := i + 100
		if fim > len(ids) {
			fim = len(ids)
		}
		entradas := make([]map[string]string, 0, fim-i)
		for _, id := range ids[i:fim] {
			entradas = append(entradas, map[string]string{"id": id})
		}
		var r struct {
			Results []objetoBruto `json:"results"`
		}
		corpo := map[string]any{"properties": props, "inputs": entradas}
		if err := c.chamar(ctx, "POST", "/crm/v3/objects/"+tipo+"/batch/read", corpo, &r); err != nil {
			return nil, err
		}
		for _, o := range r.Results {
			out = append(out, o.limpo())
		}
	}
	return out, nil
}
