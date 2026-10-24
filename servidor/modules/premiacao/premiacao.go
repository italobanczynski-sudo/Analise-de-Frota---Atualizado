// Package premiacao guarda nota, km e bonificação de cada motorista por mês.
// Só o nome do motorista é guardado: o ID e o CPF que a API devolve nem são lidos.
package premiacao

import (
	"context"
	"sort"
	"strings"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/premiacao/premiacao_repo"
)

type Service_I interface {
	// Salvar substitui os meses que vieram na resposta; devolve os meses gravados (AAAA-MM)
	Salvar(ctx context.Context, clienteID int64, recs []gobraxapi.PremiacaoRec_t) ([]string, error)
	// Listar devolve um mês (AAAA-MM) ou, com mes vazio, todos
	Listar(ctx context.Context, clienteID int64, mes string) ([]Premiacao_t, error)
}

type premiacao_t struct {
	repo premiacao_repo.Repository_I
}

func NewPremiacao() Service_I {
	return &premiacao_t{repo: premiacao_repo.NewRepository()}
}

// NewPremiacaoCom recebe o repositório (usado nos testes)
func NewPremiacaoCom(repo premiacao_repo.Repository_I) Service_I {
	return &premiacao_t{repo: repo}
}

func (p *premiacao_t) Salvar(ctx context.Context, clienteID int64, recs []gobraxapi.PremiacaoRec_t) ([]string, error) {
	porMes := map[string][]premiacao_repo.PremiacaoRow_t{}
	for _, r := range recs {
		nome := strings.TrimSpace(r.Name)
		if nome == "" {
			continue
		}
		for _, o := range r.Overview {
			mes := mesDaAPI(o.Date)
			if mes == "" || comErro(o.Err) {
				continue
			}
			km := o.TotalKM.Valor
			if km == nil {
				km = o.TotalKm.Valor
			}
			porMes[mes] = append(porMes[mes], premiacao_repo.PremiacaoRow_t{ClienteID: clienteID, Mes: mes, Motorista: nome, Nota: o.Score.Valor, Km: km, Premio: o.Reward.Valor})
		}
	}
	meses := make([]string, 0, len(porMes))
	for m := range porMes {
		meses = append(meses, m)
	}
	sort.Strings(meses)
	for _, m := range meses {
		if err := p.repo.SubstituirMes(ctx, clienteID, m, porMes[m]); err != nil {
			return nil, err
		}
	}
	return meses, nil
}

func (p *premiacao_t) Listar(ctx context.Context, clienteID int64, mes string) ([]Premiacao_t, error) {
	if err := validarMes(mes); err != nil {
		return nil, err
	}
	rows, err := p.repo.Listar(ctx, clienteID, mes)
	if err != nil {
		return nil, err
	}
	out := make([]Premiacao_t, len(rows))
	for i, r := range rows {
		out[i] = Premiacao_t{Mes: r.Mes, Motorista: r.Motorista, Nota: r.Nota, Km: r.Km, Premio: r.Premio}
	}
	return out, nil
}
