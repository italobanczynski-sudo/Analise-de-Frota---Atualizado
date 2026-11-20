package cliente_repo

const sqlColunas = `id, nome, token_cifrado IS NOT NULL AS tem_token, criado_em`

const sqlListar = `SELECT ` + sqlColunas + ` FROM clientes ORDER BY nome`

const sqlBuscarPorID = `SELECT ` + sqlColunas + ` FROM clientes WHERE id = $1`

const sqlBuscarPorNome = `SELECT ` + sqlColunas + ` FROM clientes WHERE lower(nome) = lower($1)`

const sqlCriar = `INSERT INTO clientes (nome) VALUES ($1) RETURNING id`

const sqlSalvarToken = `UPDATE clientes SET token_cifrado = $2, atualizado_em = now() WHERE id = $1`

const sqlTokenCifrado = `SELECT token_cifrado FROM clientes WHERE id = $1`

const sqlNomePorID = `SELECT nome FROM clientes WHERE id = $1`

// documentos do painel ligados ao cliente (planilhas importadas, ficha na coleção clientes, tratativas...)
const sqlApagarDocsDoCliente = `DELETE FROM base_documentos
WHERE lower(trim(dados->>'cliente')) = lower(trim($1))
   OR (colecao = 'clientes' AND lower(trim(dados->>'nome')) = lower(trim($1)))`

// registros, odômetros, premiações, coletas, mensagens, telefones e posições saem junto (ON DELETE CASCADE)
const sqlExcluir = `DELETE FROM clientes WHERE id = $1`
