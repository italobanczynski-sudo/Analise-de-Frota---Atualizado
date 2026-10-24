package odometro_repo

const sqlApagar = `DELETE FROM odometros WHERE cliente_id = $1`

const sqlInserir = `INSERT INTO odometros (cliente_id, placa, odometro, coletado) VALUES (:cliente_id, :placa, :odometro, :coletado)`

const sqlListar = `SELECT cliente_id, placa, odometro, coletado, buscado_em FROM odometros WHERE cliente_id = $1 ORDER BY placa`
