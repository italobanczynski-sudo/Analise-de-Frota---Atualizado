package gobraxapi

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// Numero_t aceita número, texto ("12,5"), vazio ou null; sem valor válido fica Valor == nil
type Numero_t struct{ Valor *float64 }

func (n *Numero_t) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), `"`)
	n.Valor = nil
	if s == "" || s == "null" {
		return nil
	}
	if f, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64); err == nil {
		n.Valor = &f
	}
	return nil
}

// registros_t aceita {"records": [...]}, {"data": [...]} ou a lista direto
type registros_t[T any] struct{ Lista []T }

func (r *registros_t[T]) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '[' {
		return json.Unmarshal(b, &r.Lista)
	}
	var env struct {
		Records []T             `json:"records"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	r.Lista = env.Records
	if r.Lista == nil && len(env.Data) > 0 && env.Data[0] == '[' {
		return json.Unmarshal(env.Data, &r.Lista)
	}
	return nil
}

// GET v1/vehicle-statistics
type EstatisticaRec_t struct {
	VehicleIdentification    string   `json:"vehicleIdentification"`
	TotalMileage             Numero_t `json:"totalMileage"`
	TotalConsumption         Numero_t `json:"totalConsumption"`
	TotalBreaking            Numero_t `json:"totalBreaking"`
	TotalBreakingOnHighSpeed Numero_t `json:"totalBreakingOnHighSpeed"`
	AverageSpeed             Numero_t `json:"averageSpeed"`
	Odometer                 Numero_t `json:"odometer"`
}

// GET v1/vehicle-performance. O CPF do motorista não é declarado aqui de propósito: nunca é lido.
type PerformanceRec_t struct {
	VehicleIdentification string                 `json:"vehicleIdentification"`
	Drivers               []Motorista_t          `json:"drivers"`
	Indicators            map[string]Indicador_t `json:"indicators"`
	Score                 json.RawMessage        `json:"score"`
}

type Motorista_t struct {
	DriverName string `json:"driverName"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

type Indicador_t struct {
	Duration   Numero_t `json:"duration"` // minutos
	Percentage Numero_t `json:"percentage"`
	Score      Numero_t `json:"score"`
}

// GET v2/vehicle-odometer
type OdometroRec_t struct {
	VehicleIdentification string   `json:"vehicleIdentification"`
	Odometer              Numero_t `json:"odometer"`
	LastUpdated           string   `json:"lastUpdated"`
}

// GET v2/driversOverview. ID e DocumentNumber (CPF) não são declarados: nunca são lidos.
type PremiacaoRec_t struct {
	Name     string           `json:"Name"`
	Overview []PremiacaoMes_t `json:"Overview"`
}

type PremiacaoMes_t struct {
	Date    string          `json:"Date"` // MM-AAAA
	Reward  Numero_t        `json:"Reward"`
	TotalKM Numero_t        `json:"TotalKM"`
	TotalKm Numero_t        `json:"TotalKm"`
	Score   Numero_t        `json:"Score"`
	Err     json.RawMessage `json:"Err"`
}
