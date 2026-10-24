package gobraxapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"analisefrota/adapters/gobraxapi"
)

type GobraxAPISuite struct {
	suite.Suite
	srv     *httptest.Server
	handler http.HandlerFunc
	api     gobraxapi.Client_I
}

func (s *GobraxAPISuite) SetupTest() {
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { s.handler(w, r) }))
	s.api = gobraxapi.NewClientCom(s.srv.URL+"/api/", s.srv.Client(), 0)
}

func (s *GobraxAPISuite) TearDownTest() { s.srv.Close() }

func (s *GobraxAPISuite) TestEstatisticasLeRecordsENumerosComoTexto() {
	s.handler = func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/api/v1/vehicle-statistics", r.URL.Path)
		s.Equal("Bearer tk", r.Header.Get("Authorization"))
		s.Equal("2026-10-23 03:00:00", r.URL.Query().Get("startDate"))
		_, _ = w.Write([]byte(`{"records":[{"vehicleIdentification":"ABC1D23","totalMileage":"12,5","totalConsumption":null,"odometer":1000}]}`))
	}
	ini := time.Date(2026, 10, 23, 3, 0, 0, 0, time.UTC)
	recs, err := s.api.Estatisticas(context.Background(), "tk", ini, ini.Add(time.Hour))
	s.Require().NoError(err)
	s.Require().Len(recs, 1)
	s.Equal(12.5, *recs[0].TotalMileage.Valor)
	s.Nil(recs[0].TotalConsumption.Valor)
	s.Equal(1000.0, *recs[0].Odometer.Valor)
}

func (s *GobraxAPISuite) TestPremiacaoLeData() {
	s.handler = func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":[{"ID":1,"Name":"Motorista Exemplo","DocumentNumber":"<cpf>","Overview":[{"Date":"09-2026","Score":88,"TotalKM":1200,"Reward":150}]}]}`))
	}
	recs, err := s.api.Premiacao(context.Background(), "tk", "09-2026", "10-2026")
	s.Require().NoError(err)
	s.Require().Len(recs, 1)
	s.Equal("Motorista Exemplo", recs[0].Name)
	s.Equal(88.0, *recs[0].Overview[0].Score.Valor)
}

func (s *GobraxAPISuite) TestTokenInvalido() {
	s.handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"data":{"code":61,"description":"Token inválido"}}`))
	}
	_, err := s.api.Odometro(context.Background(), "tk", time.Now(), time.Now())
	s.ErrorIs(err, gobraxapi.ErrTokenInvalido)
}

func (s *GobraxAPISuite) TestSemAcessoESemDados() {
	s.handler = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }
	_, err := s.api.Performance(context.Background(), "tk", time.Now(), time.Now())
	s.ErrorIs(err, gobraxapi.ErrSemAcesso)
	s.handler = func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) }
	_, err = s.api.Performance(context.Background(), "tk", time.Now(), time.Now())
	s.ErrorIs(err, gobraxapi.ErrSemDados)
}

func (s *GobraxAPISuite) TestOutroErroTrazADescricao() {
	s.handler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"data":{"code":3,"description":"Datas fornecidas inválidas"}}`))
	}
	_, err := s.api.Premiacao(context.Background(), "tk", "10-2026", "10-2026")
	s.EqualError(err, "a API respondeu 400 (Datas fornecidas inválidas)")
}

func (s *GobraxAPISuite) TestEsperaQuandoRecebe429() {
	chamadas := 0
	s.handler = func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		if chamadas == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}
	recs, err := s.api.Estatisticas(context.Background(), "tk", time.Now(), time.Now())
	s.NoError(err)
	s.Empty(recs)
	s.Equal(2, chamadas)
}

func TestGobraxAPISuite(t *testing.T) { suite.Run(t, new(GobraxAPISuite)) }
