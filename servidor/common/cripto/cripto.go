// Package cripto cifra os tokens dos clientes antes de irem para o banco (AES-256-GCM).
package cripto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var ErrChave = errors.New("cripto: a chave precisa ter 32 bytes")

// Cifrar devolve nonce + texto cifrado
func Cifrar(chave []byte, texto string) ([]byte, error) {
	gcm, err := novoGCM(chave)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(texto), nil), nil
}

// Decifrar desfaz Cifrar; falha se a chave for outra ou o conteúdo tiver sido alterado
func Decifrar(chave, dado []byte) (string, error) {
	gcm, err := novoGCM(chave)
	if err != nil {
		return "", err
	}
	if len(dado) < gcm.NonceSize() {
		return "", errors.New("cripto: conteúdo cifrado incompleto")
	}
	texto, err := gcm.Open(nil, dado[:gcm.NonceSize()], dado[gcm.NonceSize():], nil)
	if err != nil {
		return "", errors.New("cripto: não foi possível decifrar (chave diferente ou conteúdo alterado)")
	}
	return string(texto), nil
}

func novoGCM(chave []byte) (cipher.AEAD, error) {
	if len(chave) != 32 {
		return nil, ErrChave
	}
	bloco, err := aes.NewCipher(chave)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(bloco)
}
