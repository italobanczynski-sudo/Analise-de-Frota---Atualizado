package cliente_repo

const sqlColunas = `id, nome, hubspot_id, token_cifrado IS NOT NULL AS tem_token, criado_em`

const sqlListar = `SELECT ` + sqlColunas + ` FROM clientes ORDER BY nome`

const sqlBuscarPorID = `SELECT ` + sqlColunas + ` FROM clientes WHERE id = $1`

const sqlBuscarPorNome = `SELECT ` + sqlColunas + ` FROM clientes WHERE lower(nome) = lower($1)`

const sqlCriar = `INSERT INTO clientes (nome, hubspot_id) VALUES ($1, $2) RETURNING id`

const sqlSalvarToken = `UPDATE clientes SET token_cifrado = $2, atualizado_em = now() WHERE id = $1`

const sqlTokenCifrado = `SELECT token_cifrado FROM clientes WHERE id = $1`
