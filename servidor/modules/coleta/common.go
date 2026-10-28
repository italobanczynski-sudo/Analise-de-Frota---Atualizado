package coleta

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"analisefrota/adapters/gobraxapi"
)

// emAndamento impede duas buscas do mesmo cliente ao mesmo tempo (a automática e a pedida na tela)
var emAndamento sync.Map

func trancar(clienteID int64) bool {
	_, ja := emAndamento.LoadOrStore(clienteID, true)
	return !ja
}

func liberar(clienteID int64) { emAndamento.Delete(clienteID) }

// contador de falhas seguidas de um tipo de busca; 3 seguidas desligam o tipo nesta rodada
type falhas_t struct {
	nome      string
	seguidas  int
	desligado bool
	semAcesso bool // o token não tem acesso a este dado: os dias entram sem ele (não adianta tentar de novo)
	limite    bool // a API pediu para esperar (429 longo): o resto da rodada fica para a próxima
	perdidos  []string
	ultimoErr error
}

func (f *falhas_t) acertou() { f.seguidas = 0 }

func (f *falhas_t) errou(dia string, err error) {
	f.seguidas++
	f.ultimoErr = err
	switch {
	case errors.Is(err, gobraxapi.ErrSemAcesso):
		f.semAcesso = true
	case errors.Is(err, gobraxapi.ErrLimite):
		f.limite, f.desligado = true, true
	}
	if !f.semAcesso {
		f.perdidos = append(f.perdidos, dia)
	}
	if f.seguidas >= 3 {
		f.desligado = true
	}
}

// faltou: o dado deste dia não veio por falha ou limite (não por falta de acesso): o dia não é gravado
// pela metade e volta na próxima busca. Dia não pedido porque o tipo já estava desligado também conta.
func (f *falhas_t) faltou(dia string, pediu bool) bool {
	if f.semAcesso {
		return false
	}
	if !pediu {
		f.perdidos = append(f.perdidos, dia)
	}
	return true
}

func (f *falhas_t) avisos() []string {
	var a []string
	if f.limite {
		a = append(a, fmt.Sprintf("%s: %v; %d dia(s) ficam para a próxima busca", f.nome, f.ultimoErr, len(f.perdidos)))
	} else if f.desligado {
		a = append(a, fmt.Sprintf("%s parou depois de 3 falhas seguidas: %v", f.nome, f.ultimoErr))
	} else if len(f.perdidos) > 0 {
		a = append(a, fmt.Sprintf("%s falhou em %d dia(s): %s (a próxima busca tenta de novo)", f.nome, len(f.perdidos), strings.Join(f.perdidos, ", ")))
	}
	return a
}

func resumo(r Resultado_t) string {
	partes := []string{fmt.Sprintf("%d dia(s) com dados, até %d veículos", r.DiasComDados, r.Veiculos)}
	if !r.Performance {
		partes = append(partes, "sem performance")
	}
	partes = append(partes, fmt.Sprintf("odômetro de %d veículos", r.Odometros), fmt.Sprintf("premiação de %d mês(es)", len(r.MesesPremio)))
	return strings.Join(partes, ", ")
}
