package contato_repo

const sqlColunas = `cliente_id, motorista, telefone_cifrado, telefone_final, ativo, autorizado_em, autorizado_por`

const sqlListar = `SELECT ` + sqlColunas + ` FROM contatos_motoristas WHERE cliente_id = $1 ORDER BY motorista`

const sqlListarAtivos = `SELECT ` + sqlColunas + ` FROM contatos_motoristas WHERE cliente_id = $1 AND ativo ORDER BY motorista`

// trocar o telefone renova a autorização: quem salva confirma que o motorista aceitou receber
const sqlSalvar = `INSERT INTO contatos_motoristas (cliente_id, motorista, telefone_cifrado, telefone_final, ativo, autorizado_em, autorizado_por)
VALUES (:cliente_id, :motorista, :telefone_cifrado, :telefone_final, true, :autorizado_em, :autorizado_por)
ON CONFLICT (cliente_id, motorista) DO UPDATE SET telefone_cifrado = EXCLUDED.telefone_cifrado, telefone_final = EXCLUDED.telefone_final,
	ativo = true, autorizado_em = EXCLUDED.autorizado_em, autorizado_por = EXCLUDED.autorizado_por, atualizado_em = now()`

const sqlAtivar = `UPDATE contatos_motoristas SET ativo = $3, atualizado_em = now() WHERE cliente_id = $1 AND motorista = $2`

const sqlRemover = `DELETE FROM contatos_motoristas WHERE cliente_id = $1 AND motorista = $2`
