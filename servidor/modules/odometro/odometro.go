// Package odometro guarda o último odômetro de cada veículo (substituído a cada busca).
package odometro

import (
	"context"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
	"analisefrota/modules/odometro/odometro_repo"
)

type Service_I interface {
	// Substituir troca os odômetros do cliente pelos da última busca; devolve quantos veículos entraram
	Substituir(ctx context.Context, clienteID int64, recs []gobraxapi.OdometroRec_t) (int, error)
	Listar(ctx context.Context, clienteID int64) ([]Odometro_t, error)
}

type odometro_t struct {
	repo odometro_repo.Repository_I
}

func NewOdometro() Service_I {
	return &odometro_t{repo: odometro_repo.NewRepository()}
}

// NewOdometroCom recebe o repositório (usado nos testes)
func NewOdometroCom(repo odometro_repo.Repository_I) Service_I {
	return &odometro_t{repo: repo}
}

func (o *odometro_t) Substituir(ctx context.Context, clienteID int64, recs []gobraxapi.OdometroRec_t) (int, error) {
	vistos := map[string]bool{}
	rows := []odometro_repo.OdometroRow_t{}
	for _, r := range recs {
		placa := utils.Placa(r.VehicleIdentification)
		k := utils.ChavePlaca(placa)
		if !valido(placa) || vistos[k] {
			continue
		}
		vistos[k] = true
		rows = append(rows, odometro_repo.OdometroRow_t{ClienteID: clienteID, Placa: placa, Odometro: r.Odometer.Valor, Coletado: r.LastUpdated})
	}
	return len(rows), o.repo.Substituir(ctx, clienteID, rows)
}

func (o *odometro_t) Listar(ctx context.Context, clienteID int64) ([]Odometro_t, error) {
	rows, err := o.repo.Listar(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	out := make([]Odometro_t, len(rows))
	for i, r := range rows {
		out[i] = Odometro_t{Placa: r.Placa, Odometro: r.Odometro, Coletado: r.Coletado, BuscadoEm: r.BuscadoEm}
	}
	return out, nil
}
