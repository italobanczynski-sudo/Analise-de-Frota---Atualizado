package usuario_repo

const sqlColunas = `id, nome, email, papel, senha_hash, ativo, criado_em, ultimo_acesso`

const sqlListar = `SELECT ` + sqlColunas + ` FROM usuarios ORDER BY ativo DESC, nome`

const sqlBuscarPorID = `SELECT ` + sqlColunas + ` FROM usuarios WHERE id = $1`

const sqlBuscarPorEmail = `SELECT ` + sqlColunas + ` FROM usuarios WHERE lower(email) = lower($1)`

const sqlCriar = `INSERT INTO usuarios (nome, email, papel) VALUES ($1, $2, $3) RETURNING id`

const sqlAlterar = `UPDATE usuarios SET papel = $2, ativo = $3, atualizado_em = now() WHERE id = $1`

const sqlContarAdminsAtivos = `SELECT count(*) FROM usuarios WHERE papel = 'admin' AND ativo`

const sqlMarcarAcesso = `UPDATE usuarios SET ultimo_acesso = $2 WHERE id = $1`

// um convite novo invalida os anteriores ainda não usados da mesma pessoa
const sqlInvalidarConvites = `UPDATE convites SET usado_em = $2 WHERE usuario_id = $1 AND usado_em IS NULL`

const sqlCriarConvite = `INSERT INTO convites (token_hash, usuario_id, expira_em) VALUES ($1, $2, $3)`

// usa o convite (uma vez só, dentro do prazo, só para usuário ativo) e devolve de quem ele era
const sqlUsarConvite = `UPDATE convites c SET usado_em = $2
FROM usuarios u
WHERE c.token_hash = $1 AND c.usado_em IS NULL AND c.expira_em > $2 AND u.id = c.usuario_id AND u.ativo
RETURNING c.usuario_id`

const sqlDefinirSenha = `UPDATE usuarios SET senha_hash = $2, atualizado_em = now() WHERE id = $1`

// ---------- códigos de acesso por e-mail ----------

const sqlUltimoCodigoEm = `SELECT max(criado_em) FROM codigos_acesso WHERE usuario_id = $1`

const sqlInvalidarCodigos = `UPDATE codigos_acesso SET usado_em = $2 WHERE usuario_id = $1 AND usado_em IS NULL`

const sqlCriarCodigo = `INSERT INTO codigos_acesso (usuario_id, codigo_hash, expira_em, criado_em) VALUES ($1, $2, $3, $4)`

const sqlCodigoAtivo = `SELECT id, codigo_hash, tentativas FROM codigos_acesso
WHERE usuario_id = $1 AND usado_em IS NULL AND expira_em > $2
ORDER BY criado_em DESC LIMIT 1`

const sqlSomarTentativa = `UPDATE codigos_acesso SET tentativas = tentativas + 1 WHERE id = $1`

const sqlUsarCodigo = `UPDATE codigos_acesso SET usado_em = $3 WHERE id = $1 AND usuario_id = $2 AND usado_em IS NULL`
