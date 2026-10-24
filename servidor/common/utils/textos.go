package utils

import (
	"math"
	"regexp"
	"strings"
)

var naoAlfanum = regexp.MustCompile(`[^A-Z0-9]`)

// Placa devolve a placa em maiúsculas, sem espaços nas pontas
func Placa(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }

// ChavePlaca junta placas escritas de jeitos diferentes ("ABC-1234" e "abc1234")
func ChavePlaca(v string) string { return naoAlfanum.ReplaceAllString(Placa(v), "") }

var (
	prefixoAuth   = regexp.MustCompile(`(?i)^\s*authorization\s*:\s*`)
	prefixoBearer = regexp.MustCompile(`(?i)^\s*bearer\s+`)
	sujeiraToken  = regexp.MustCompile(`["'\s]`)
)

// LimparToken tira o que costuma vir colado no token: aspas, espaços, quebras de linha, "Bearer " e "Authorization:"
func LimparToken(t string) string {
	t = prefixoAuth.ReplaceAllString(t, "")
	t = prefixoBearer.ReplaceAllString(t, "")
	return sujeiraToken.ReplaceAllString(t, "")
}

// Arred2 arredonda para duas casas, como o painel mostra
func Arred2(v float64) float64 { return math.Round(v*100) / 100 }
