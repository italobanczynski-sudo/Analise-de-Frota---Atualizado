package cripto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// TokenAleatorio gera um token de 32 bytes aleatórios em base64 para URL (link de convite, cookie de sessão)
func TokenAleatorio() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken é o que vai para o banco no lugar do token: quem lê o banco não consegue usar o link nem a sessão
func HashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
