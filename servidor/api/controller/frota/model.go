package frota

type PeriodoRequest struct {
	Inicio string `form:"inicio" validate:"required,datetime=2006-01-02"`
	Fim    string `form:"fim" validate:"required,datetime=2006-01-02"`
}

type MesRequest struct {
	Mes string `form:"mes" validate:"omitempty,datetime=2006-01"`
}
