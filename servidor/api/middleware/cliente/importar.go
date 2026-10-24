package cliente

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	modcliente "analisefrota/modules/cliente"
)

// Importar traz os clientes do script antigo: api/clientes.json (nome, variável do token, HubSpot)
// e os tokens do api/.env. Cria quem não existe e salva o token cifrado. Nenhum token vai para o log.
func Importar(ctx context.Context, arqClientes, arqTokens string) ([]string, error) {
	var lista []struct {
		Cliente   string `json:"cliente"`
		Token     string `json:"token"`
		HubspotID string `json:"hubspot_id"`
	}
	b, err := os.ReadFile(arqClientes)
	if err != nil {
		return nil, fmt.Errorf("não li %s: %w", arqClientes, err)
	}
	if err := json.Unmarshal(b, &lista); err != nil {
		return nil, fmt.Errorf("%s não é uma lista JSON válida: %w", arqClientes, err)
	}
	tokens, err := lerEnv(arqTokens)
	if err != nil {
		return nil, err
	}
	svc := modcliente.NewCliente()
	var feito []string
	for _, item := range lista {
		nome := strings.TrimSpace(item.Cliente)
		hub := item.HubspotID
		if !regexp.MustCompile(`^[0-9]+$`).MatchString(hub) {
			hub = "" // exemplo ou vazio no arquivo
		}
		c, err := svc.BuscarPorNome(ctx, nome)
		if errors.Is(err, modcliente.ErrNaoEncontrado) {
			c, err = svc.Criar(ctx, nome, hub)
		}
		if err != nil {
			feito = append(feito, fmt.Sprintf("%s: erro: %v", nome, err))
			continue
		}
		tk := tokens[item.Token]
		if tk == "" {
			feito = append(feito, fmt.Sprintf("%s: cadastrado, sem token em %s", nome, item.Token))
			continue
		}
		if err := svc.SalvarToken(ctx, c.ID, tk); err != nil {
			feito = append(feito, fmt.Sprintf("%s: token recusado: %v", nome, err))
			continue
		}
		feito = append(feito, fmt.Sprintf("%s: cadastrado com token", nome))
	}
	return feito, nil
}

// lerEnv lê um arquivo CHAVE=valor (o api/.env do script antigo); não mexe nas variáveis do processo
func lerEnv(arq string) (map[string]string, error) {
	out := map[string]string{}
	if arq == "" {
		return out, nil
	}
	f, err := os.Open(arq)
	if err != nil {
		return nil, fmt.Errorf("não li %s: %w", arq, err)
	}
	defer f.Close()
	linha := regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*?)\s*$`)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		t := sc.Text()
		if strings.HasPrefix(strings.TrimSpace(t), "#") {
			continue
		}
		if m := linha.FindStringSubmatch(t); m != nil {
			out[m[1]] = strings.Trim(m[2], `"'`)
		}
	}
	return out, sc.Err()
}
