package sessao_repo

const sqlCriar = `INSERT INTO sessoes (token_hash, usuario_id, expira_em) VALUES ($1, $2, $3)`

const sqlUsuarioDaSessao = `SELECT usuario_id FROM sessoes WHERE token_hash = $1 AND expira_em > $2`

const sqlApagar = `DELETE FROM sessoes WHERE token_hash = $1`

const sqlApagarDoUsuario = `DELETE FROM sessoes WHERE usuario_id = $1`

const sqlApagarVencidas = `DELETE FROM sessoes WHERE expira_em <= $1`
