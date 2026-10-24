package usuario

import "time"

// Perfis de acesso
const (
	PapelAdmin       = "admin"       // cuida de usuários, clientes e tokens
	PapelColaborador = "colaborador" // pede buscas e vê tudo
	PapelLeitor      = "leitor"      // só vê
)

// Usuario_t é o que sai do módulo: a senha nunca aparece, só se ela já foi criada
type Usuario_t struct {
	ID           int64      `json:"id"`
	Nome         string     `json:"nome"`
	Email        string     `json:"email"`
	Papel        string     `json:"papel"`
	Ativo        bool       `json:"ativo"`
	TemSenha     bool       `json:"tem_senha"` // false = convite ainda não aceito
	CriadoEm     time.Time  `json:"criado_em"`
	UltimoAcesso *time.Time `json:"ultimo_acesso"`
}

// Codigo_t é o código de 6 dígitos enviado por e-mail; o número só existe aqui (no banco fica o HMAC)
type Codigo_t struct {
	Codigo   string
	Usuario  Usuario_t
	ExpiraEm time.Time
}

// Regras do código de acesso
const (
	CodigoValidade   = 15 * time.Minute
	CodigoIntervalo  = 60 * time.Second // espera mínima entre dois pedidos
	CodigoTentativas = 5
)

// Convite_t carrega o token do link; ele só existe aqui, na hora da criação (no banco fica o hash)
type Convite_t struct {
	Token    string    `json:"-"`
	ExpiraEm time.Time `json:"expira_em"`
}
