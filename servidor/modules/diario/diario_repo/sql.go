package diario_repo

// colunas comuns de registros_diarios e registros_motoristas (mesmas medidas nas duas tabelas)
const colunas = `cliente_id, dia, placa, motorista, rodou, km, litros, frenagens, frenagens_alta, velocidade_media, odometro, horas, horas_motor, horas_pedal, nota,
	faixa_verde, inicio_faixa_verde, final_faixa_verde, eco_roll, pressao_ideal, pressao_atencao, embalo, motor_ocioso,
	pressao_acelerador, excesso_rotacao, tempo_excesso_velocidade, piloto_automatico, freio_motor`

const valores = `:cliente_id, :dia, :placa, :motorista, :rodou, :km, :litros, :frenagens, :frenagens_alta, :velocidade_media, :odometro, :horas, :horas_motor, :horas_pedal, :nota,
	:faixa_verde, :inicio_faixa_verde, :final_faixa_verde, :eco_roll, :pressao_ideal, :pressao_atencao, :embalo, :motor_ocioso,
	:pressao_acelerador, :excesso_rotacao, :tempo_excesso_velocidade, :piloto_automatico, :freio_motor`

const selecao = `cliente_id, to_char(dia, 'YYYY-MM-DD') AS dia, placa, motorista, rodou, km, litros, frenagens, frenagens_alta, velocidade_media,
	odometro, horas, horas_motor, horas_pedal, nota, faixa_verde, inicio_faixa_verde, final_faixa_verde, eco_roll, pressao_ideal, pressao_atencao, embalo,
	motor_ocioso, pressao_acelerador, excesso_rotacao, tempo_excesso_velocidade, piloto_automatico, freio_motor`

const sqlApagarDia = `DELETE FROM registros_diarios WHERE cliente_id = $1 AND dia = $2`

const sqlInserir = `INSERT INTO registros_diarios (` + colunas + `) VALUES (` + valores + `)`

const sqlListar = `SELECT ` + selecao + `
FROM registros_diarios
WHERE cliente_id = $1 AND dia BETWEEN $2 AND $3
ORDER BY dia, placa`

// Um dia está completo quando tem a performance gravada (ao menos um veículo com indicador de condução) e, se algum
// veículo rodou com motorista vinculado, as linhas dos motoristas já foram montadas. Dia gravado antes da tabela
// registros_motoristas existir volta a ser buscado quando cair na janela da busca.
const sqlDiasComIndicadores = `SELECT DISTINCT to_char(r.dia, 'YYYY-MM-DD') FROM registros_diarios r
WHERE r.cliente_id = $1 AND r.dia BETWEEN $2 AND $3
	AND (r.faixa_verde IS NOT NULL OR r.inicio_faixa_verde IS NOT NULL OR r.embalo IS NOT NULL OR r.motor_ocioso IS NOT NULL)
	AND (
		EXISTS (SELECT 1 FROM registros_motoristas m WHERE m.cliente_id = r.cliente_id AND m.dia = r.dia)
		OR NOT EXISTS (SELECT 1 FROM registros_diarios v WHERE v.cliente_id = r.cliente_id AND v.dia = r.dia AND v.rodou AND v.motorista <> '')
	)`

// ---------- registros_motoristas: o dado de cada motorista só no tempo do vínculo dele com o veículo ----------

const sqlApagarMotoristasDia = `DELETE FROM registros_motoristas WHERE cliente_id = $1 AND dia = $2`

const sqlInserirMotorista = `INSERT INTO registros_motoristas (` + colunas + `) VALUES (` + valores + `)`

const sqlListarMotoristas = `SELECT ` + selecao + `
FROM registros_motoristas
WHERE cliente_id = $1 AND dia BETWEEN $2 AND $3
ORDER BY dia, placa, motorista`
