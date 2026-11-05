package mapa

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"analisefrota/adapters/mapaosm"
)

// Imagem: GET /mapa/:z/:x/:y.png (logado) — fundo do mapa do painel, buscado pelo servidor no OpenStreetMap
func Imagem(c *gin.Context) {
	z, e1 := strconv.Atoi(c.Param("z"))
	x, e2 := strconv.Atoi(c.Param("x"))
	y, e3 := strconv.Atoi(strings.TrimSuffix(c.Param("y"), ".png"))
	if e1 != nil || e2 != nil || e3 != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	dados, err := mapaosm.NewMapa().Imagem(c.Request.Context(), z, x, y)
	if errors.Is(err, mapaosm.ErrPedido) {
		c.Status(http.StatusBadRequest)
		return
	}
	if err != nil {
		logrus.Warnf("mapa %d/%d/%d: %v", z, x, y, err)
		c.Status(http.StatusBadGateway)
		return
	}
	c.Header("Cache-Control", "private, max-age=604800")
	c.Data(http.StatusOK, "image/png", dados)
}
