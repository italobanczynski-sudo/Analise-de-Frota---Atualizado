package cripto_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"analisefrota/common/cripto"
)

type CriptoSuite struct {
	suite.Suite
	chave []byte
}

func (s *CriptoSuite) SetupTest() { s.chave = bytes.Repeat([]byte{7}, 32) }

func (s *CriptoSuite) TestIdaEVolta() {
	cif, err := cripto.Cifrar(s.chave, "token-de-exemplo")
	s.Require().NoError(err)
	s.NotContains(string(cif), "token-de-exemplo")
	txt, err := cripto.Decifrar(s.chave, cif)
	s.Require().NoError(err)
	s.Equal("token-de-exemplo", txt)
}

func (s *CriptoSuite) TestCadaCifraEDiferente() {
	a, _ := cripto.Cifrar(s.chave, "mesmo")
	b, _ := cripto.Cifrar(s.chave, "mesmo")
	s.NotEqual(a, b)
}

func (s *CriptoSuite) TestChaveErradaFalha() {
	cif, _ := cripto.Cifrar(s.chave, "segredo")
	_, err := cripto.Decifrar(bytes.Repeat([]byte{8}, 32), cif)
	s.Error(err)
}

func (s *CriptoSuite) TestChaveCurtaRecusada() {
	_, err := cripto.Cifrar([]byte("curta"), "x")
	s.ErrorIs(err, cripto.ErrChave)
}

func TestCriptoSuite(t *testing.T) { suite.Run(t, new(CriptoSuite)) }
