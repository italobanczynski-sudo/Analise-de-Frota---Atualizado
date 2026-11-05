package frota

type PeriodoRequest struct {
	Inicio string `form:"inicio" validate:"required,datetime=2006-01-02"`
	Fim    string `form:"fim" validate:"required,datetime=2006-01-02"`
}

// TrajetoRequest: as horas são de Brasília; sem de/ate, o dia inteiro
type TrajetoRequest struct {
	Placa string `form:"placa" validate:"required,max=12"`
	Dia   string `form:"dia" validate:"required,datetime=2006-01-02"`
	De    int    `form:"de,default=0" validate:"min=0,max=23"`
	Ate   int    `form:"ate,default=24" validate:"min=1,max=24"`
}

type MesRequest struct {
	Mes string `form:"mes" validate:"omitempty,datetime=2006-01"`
}

// AoVivoRequest: placas separadas por vírgula; desde em UTC (AAAA-MM-DD HH:MM:SS)
type AoVivoRequest struct {
	Placas string `form:"placas" validate:"required,max=2000"`
	Desde  string `form:"desde" validate:"omitempty,max=19"`
}
