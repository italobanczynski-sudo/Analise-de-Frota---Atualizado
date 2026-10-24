package usuario

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// gerarCodigo sorteia 6 dígitos (000000 a 999999) com o gerador criptográfico
func gerarCodigo() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashCodigo usa HMAC com a chave do servidor: com só 1 milhão de códigos possíveis, um hash simples
// seria quebrado na hora por quem lesse o banco; sem a chave, não é.
func hashCodigo(chave []byte, usuarioID int64, codigo string) []byte {
	m := hmac.New(sha256.New, chave)
	fmt.Fprintf(m, "codigo-acesso:%d:%s", usuarioID, codigo)
	return m.Sum(nil)
}
