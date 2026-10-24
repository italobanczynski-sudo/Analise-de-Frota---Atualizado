package coleta_repo

const sqlBuscar = `SELECT cliente_id, to_char(ultimo_dia, 'YYYY-MM-DD') AS ultimo_dia, iniciada_em, terminada_em, situacao, mensagem, avisos
FROM coletas WHERE cliente_id = $1`

const sqlMarcarInicio = `INSERT INTO coletas (cliente_id, iniciada_em, situacao, mensagem, avisos) VALUES ($1, $2, 'rodando', '', '{}')
ON CONFLICT (cliente_id) DO UPDATE SET iniciada_em = EXCLUDED.iniciada_em, situacao = 'rodando', mensagem = '', avisos = '{}'`

// ultimo_dia só muda quando a busca termina bem ($5 não nulo)
const sqlMarcarFim = `UPDATE coletas SET situacao = $2, mensagem = $3, avisos = $4, ultimo_dia = COALESCE($5::date, ultimo_dia), terminada_em = $6
WHERE cliente_id = $1`
