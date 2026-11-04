package coleta

import (
	"context"
	"errors"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/diario"
)

// motoristasDoDia grava as linhas dos motoristas de um dia já salvo: quem ficou o dia todo no veículo usa a linha do
// veículo; cada vínculo parcial é consultado na API só nos trechos daquele motorista (estatísticas e, se houver,
// performance) e os trechos são juntados pelo módulo diario.
// salvou = false quando alguma consulta falhou: as linhas antigas do dia ficam como estão e o dia volta na próxima
// busca (a falha conta em fe/fp como as do dia inteiro; limite de uso para a rodada). err só para token inválido ou banco.
func (c *coleta_t) motoristasDoDia(ctx context.Context, token string, clienteID int64, dia string, plano diario.PlanoMotoristas_t, fe, fp *falhas_t) (bool, error) {
	regs := append([]diario.Registro_t{}, plano.Inteiros...)
	for _, p := range plano.Pendentes {
		partes := make([]diario.Parte_t, 0, len(p.Janelas))
		for _, j := range p.Janelas {
			estat, err := c.d.API.EstatisticasVeiculo(ctx, token, p.Placa, j.Ini, j.Fim)
			switch {
			case err == nil || errors.Is(err, gobraxapi.ErrSemDados):
				fe.acertou()
			case errors.Is(err, gobraxapi.ErrTokenInvalido):
				return false, err
			default:
				fe.errou(dia, err)
				return false, nil
			}
			var perf []gobraxapi.PerformanceRec_t
			if !fp.desligado {
				perf, err = c.d.API.PerformanceVeiculo(ctx, token, p.Placa, j.Ini, j.Fim)
				switch {
				case err == nil || errors.Is(err, gobraxapi.ErrSemDados):
					fp.acertou()
				case errors.Is(err, gobraxapi.ErrTokenInvalido):
					return false, err
				case errors.Is(err, gobraxapi.ErrSemAcesso):
					fp.errou(dia, err) // sem acesso à performance: o motorista entra só com km e litros
					perf = nil
				default:
					fp.errou(dia, err)
					return false, nil
				}
			}
			partes = append(partes, diario.Parte_t{Estat: estat, Perf: perf})
		}
		if r, ok := c.d.Diario.MontarMotorista(clienteID, dia, p, partes); ok {
			regs = append(regs, r)
		}
	}
	if err := c.d.Diario.SalvarMotoristasDia(ctx, clienteID, dia, regs); err != nil {
		return false, err
	}
	return true, nil
}
