package coleta

import (
	"fmt"
	"strings"
	"sync"
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
	perdidos  []string
	ultimoErr error
}

func (f *falhas_t) acertou() { f.seguidas = 0 }

func (f *falhas_t) errou(dia string, err error) {
	f.seguidas++
	f.ultimoErr = err
	f.perdidos = append(f.perdidos, dia)
	if f.seguidas >= 3 {
		f.desligado = true
	}
}

func (f *falhas_t) avisos() []string {
	var a []string
	if f.desligado {
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
