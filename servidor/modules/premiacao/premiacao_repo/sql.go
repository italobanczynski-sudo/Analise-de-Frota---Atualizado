package premiacao_repo

const sqlApagarMes = `DELETE FROM premiacoes WHERE cliente_id = $1 AND mes = $2`

// ON CONFLICT: a API pode repetir o mesmo nome no mês; fica o último
const sqlInserir = `INSERT INTO premiacoes (cliente_id, mes, motorista, nota, km, premio)
VALUES (:cliente_id, :mes, :motorista, :nota, :km, :premio)
ON CONFLICT (cliente_id, mes, motorista) DO UPDATE SET nota = EXCLUDED.nota, km = EXCLUDED.km, premio = EXCLUDED.premio`

const sqlListar = `SELECT cliente_id, mes, motorista, nota, km, premio FROM premiacoes
WHERE cliente_id = $1 AND ($2::text = '' OR mes = $2::text)
ORDER BY mes, motorista`
