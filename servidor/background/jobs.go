// Package background roda a busca automática: logo que o servidor sobe e depois a cada COLETA_INTERVALO_MIN.
package background

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"

	"analisefrota/environment"
	"analisefrota/modules/coleta"
)

func Init() {
	cfg := environment.Cfg.Coleta
	if !cfg.Automatica {
		logrus.Info("background: busca automática desligada (COLETA_AUTOMATICA=false)")
		return
	}
	s, err := gocron.NewScheduler()
	if err != nil {
		logrus.Fatalf("background: agendador: %v", err)
	}
	_, err = s.NewJob(
		gocron.DurationJob(time.Duration(cfg.IntervaloMin)*time.Minute),
		gocron.NewTask(buscarTodos),
		gocron.WithStartAt(gocron.WithStartImmediately()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule), // uma rodada de cada vez
	)
	if err != nil {
		logrus.Fatalf("background: tarefa de busca: %v", err)
	}
	s.Start()
	logrus.Infof("background: busca automática a cada %d min", cfg.IntervaloMin)
}

func buscarTodos() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	coleta.NewColeta().ExecutarTodos(ctx)
}
