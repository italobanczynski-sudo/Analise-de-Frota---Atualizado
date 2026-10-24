package utils

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validador     *validator.Validate
	validadorOnce sync.Once
)

// ValidateStruct aplica as tags `validate` do struct (usado nos controllers, na entrada da API)
func ValidateStruct(v any) error {
	validadorOnce.Do(func() { validador = validator.New() })
	return validador.Struct(v)
}
