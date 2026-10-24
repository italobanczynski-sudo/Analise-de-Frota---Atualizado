// Puxa os dados da API Gobrax de cada cliente listado em api/clientes.json, com o token lido de api/.env.
// Gera em api/saida/ os documentos no formato da base compartilhada do painel (coleção "planilhas")
// e as planilhas CSV de motoristas e veículos, que também podem ser importadas no painel.
// Uso: node api/puxar.mjs [--dias 30] [--cliente "Nome do cliente"] [--mapa]
// --mapa: só a posição e o percurso do dia de cada veículo, em api/saida/mapa/ (o resto não é buscado)
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PASTA = path.dirname(fileURLToPath(import.meta.url));
const SAIDA = path.join(PASTA, 'saida');
const API_BASE = 'https://gateway-v3.gobrax.com.br:8889/api/';
const API_POSICAO = 'v1/vehicle-last-position';   // resposta: records[].lastPosition { date (hora de Brasília), lat, lon }

const arg = (nome, padrao) => { const i = process.argv.indexOf('--' + nome); return i > 0 && process.argv[i + 1] ? process.argv[i + 1] : padrao; };
const DIAS = Math.min(93, Math.max(1, Number(arg('dias', 30)) || 30));
const SO_CLIENTE = arg('cliente', '');
const SO_MAPA = process.argv.includes('--mapa');   // só posição e percurso do dia (rápido, para rodar a cada minuto)

// ---------- Configuração: api/.env (tokens) e api/clientes.json (quem buscar) ----------
function lerEnv() {
  const f = path.join(PASTA, '.env');
  if (!fs.existsSync(f)) return {};
  const env = {};
  fs.readFileSync(f, 'utf8').split(/\r?\n/).forEach((l) => {
    const m = l.match(/^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)\s*$/);
    if (m && !l.trim().startsWith('#')) env[m[1]] = m[2].replace(/^(['"])(.*)\1$/, '$2');
  });
  return env;
}
const ENV = Object.assign(lerEnv(), process.env);
const fClientes = path.join(PASTA, 'clientes.json');
if (!fs.existsSync(fClientes)) {
  console.error('Falta api/clientes.json. Copie api/clientes.example.json para api/clientes.json e preencha.');
  process.exit(1);
}
const CLIENTES = JSON.parse(fs.readFileSync(fClientes, 'utf8'));

// ---------- Mesmas regras do painel (index.html, seção API Gobrax) ----------
const nrm = (v) => String(v || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/\s+/g, ' ').trim();
const ckey = (v) => nrm(v).replace(/[.,/\-]/g, ' ').replace(/\b(ltda|eireli|epp|me|s a|sa)\b/g, ' ').replace(/\s+/g, ' ').trim();
const slugCli = (n) => ckey(n).replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'sem-cliente';
const idPlan = (alvo, cli, data, parte) => [alvo, slugCli(cli), data, 'p' + parte].join('~');
const iso = (d) => d.toISOString().slice(0, 10);
const somar = (s, n) => { const d = new Date(s + 'T00:00:00Z'); d.setUTCDate(d.getUTCDate() + n); return iso(d); };
const hoje = () => iso(new Date(Date.now() - 3 * 3600e3));   // dia de Brasília
const br = (s) => { const [a, m, d] = s.split('-'); return d + '/' + m + '/' + a; };
const limparToken = (t) => String(t || '').replace(/^\s*(authorization\s*:\s*)?/i, '').replace(/^\s*bearer\s+/i, '').replace(/["'\s]/g, '');
const esperar = (ms) => new Promise((r) => setTimeout(r, ms));
const numApi = (v) => { if (v === null || v === undefined || v === '') return null; const n = typeof v === 'number' ? v : Number(String(v).replace(',', '.')); return Number.isFinite(n) ? n : null; };
const placaApi = (v) => String(v || '').trim().toUpperCase();
const chavePlaca = (p) => placaApi(p).replace(/[^A-Z0-9]/g, '');
const utc = (ms) => new Date(ms).toISOString().slice(0, 19).replace('T', ' ');
const janelaUTC = (dia) => ({ startDate: dia + ' 03:00:00', endDate: somar(dia, 1) + ' 02:59:59' });
const registros = (j) => (Array.isArray(j) ? j : (j && (j.records || j.data)) || []);
// mesma regra do index.html: só coordenada válida (fora de 0,0) entra no mapa; ignição e velocidade aceitam os nomes mais comuns
const ligadoApi = (v) => {
  if (v === true || v === 1) return true;
  if (v === false || v === 0) return false;
  const s = String(v === null || v === undefined ? '' : v).trim().toLowerCase();
  return ['true', '1', 'on', 'ligada', 'ligado'].includes(s) ? true : ['false', '0', 'off', 'desligada', 'desligado'].includes(s) ? false : null;
};
const campoPos = (r, p, nomes) => { for (const k of nomes) { if (p[k] !== undefined) return p[k]; if (r[k] !== undefined) return r[k]; } return undefined; };
const posicaoDosRegistros = (recs) => recs.filter((r) => r && r.vehicleIdentification && r.lastPosition).map((r) => {
  const p = r.lastPosition;
  return { placa: placaApi(r.vehicleIdentification), lat: numApi(p.lat !== undefined ? p.lat : p.latitude), lon: numApi(p.lon !== undefined ? p.lon : p.lng !== undefined ? p.lng : p.longitude), data: String(p.date || ''),
    ign: ligadoApi(campoPos(r, p, ['ignition', 'ignitionOn', 'ignitionStatus', 'ignicao', 'engineOn', 'engineStatus'])), vel: numApi(campoPos(r, p, ['speed', 'velocity', 'velocidade'])) };
}).filter((v) => v.lat !== null && v.lon !== null && Math.abs(v.lat) <= 90 && Math.abs(v.lon) <= 180 && !(v.lat === 0 && v.lon === 0));
// percurso do dia (mesma regra do index.html): de 5 em 5 minutos desde a meia-noite, depois o ponto novo de cada rodada.
// Zera à meia-noite (Brasília) e quando o motorista desvincula ou vincula; só entra ponto se andou 200 m; até 600 pontos.
const PASSO_MIN = 5, TRILHA_MAX = 600, PERF_CICLO = 5 * 60000;
const msPos = (s) => { const x = String(s || '').trim().replace(' ', 'T'); const t = Date.parse(/(Z|[+-]\d{2}:?\d{2})$/.test(x) ? x.replace(/([+-]\d{2})(\d{2})$/, '$1:$2') : x + '-03:00'); return Number.isFinite(t) ? t : null; };   // a posição vem na hora de Brasília, sem fuso
const msApi = (s) => { const t = Date.parse(String(s || '').replace(/([+-]\d{2})(\d{2})$/, '$1:$2')); return Number.isFinite(t) ? t : null; };   // datas da performance: ISO com +0000
const diaBR = (ms) => new Date(ms - 3 * 3600000).toISOString().slice(0, 10);
const inicioDiaBR = (dia) => Date.parse(dia + 'T00:00:00-03:00');
const distKm = (a, b) => { const r = Math.PI / 180, dLa = (b[0] - a[0]) * r, dLo = (b[1] - a[1]) * r; const h = Math.sin(dLa / 2) ** 2 + Math.cos(a[0] * r) * Math.cos(b[0] * r) * Math.sin(dLo / 2) ** 2; return 12742 * Math.asin(Math.sqrt(h)); };
const pontoPos = (v) => { const t = msPos(v.data); return t === null ? null : [Math.round(v.lat * 1e5) / 1e5, Math.round(v.lon * 1e5) / 1e5, Math.round(t / 60000)]; };
function juntarPontos(tr, pts, desdeMin) {
  pts.filter((p) => p && p[2] >= desdeMin).sort((a, b) => a[2] - b[2]).forEach((p) => {
    const ult = tr[tr.length - 1];
    if (!ult || (p[2] > ult[2] && distKm(ult, p) >= 0.2)) tr.push(p);
  });
  while (tr.length > TRILHA_MAX) tr = tr.filter((p, i) => i % 2 === 0 || i === tr.length - 1);
  return tr;
}
function sessoesDoDia(perf, dia, agora) {
  const ini = inicioDiaBR(dia), m = {};
  perf.forEach((r) => {
    const k = chavePlaca(r && r.vehicleIdentification);
    if (!k) return;
    let desde = ini, mot = '', fimUlt = 0;
    ((r && r.drivers) || []).forEach((d) => {
      const a = msApi(d.startDate), b = msApi(d.endDate);
      if (a !== null && a > ini + 60000 && a <= agora) desde = Math.max(desde, a);
      if (b !== null && b < agora - 15 * 60000) desde = Math.max(desde, b);
      if (b !== null && b >= fimUlt) { fimUlt = b; mot = String(d.driverName || '').trim(); }   // só o nome; o CPF não é guardado
    });
    if (fimUlt && fimUlt < agora - 15 * 60000) mot = '';
    m[k] = { desde, mot };
  });
  return m;
}
const bytes = (t) => Buffer.byteLength(t, 'utf8');

async function apiGet(caminho, params, token) {
  const qs = Object.entries(params).map(([k, v]) => k + '=' + encodeURIComponent(v)).join('&');
  for (let tentativa = 0; ; tentativa++) {
    const r = await fetch(API_BASE + caminho + '?' + qs, { headers: { Authorization: 'Bearer ' + token } });
    if (r.status === 429 && tentativa < 8) { await esperar(Math.max(1, Number(r.headers.get('Retry-After')) || 1) * 1000 + 100); continue; }
    if (!r.ok) {
      let corpo = null; try { corpo = await r.json(); } catch (e) { /* resposta sem JSON */ }
      const info = (corpo && corpo.data && typeof corpo.data === 'object' && !Array.isArray(corpo.data)) ? corpo.data : (corpo || {});
      const cod = Number(info.code), desc = String(info.description || info.errorDescription || '');
      if (cod === 61 || r.status === 401) throw new Error('token inválido (a API não aceitou o token)');
      if (cod === 71 || r.status === 403) throw new Error('o token não tem acesso a este dado');
      throw new Error('a API respondeu ' + r.status + (desc ? ' (' + desc + ')' : ''));
    }
    await esperar(120);
    return r.json();
  }
}

// Mais de um motorista no dia: fica quem dirigiu por mais tempo (somar a linha para todos contaria os km e as horas em dobro)
function motoristaPrincipal(p) {
  const t = new Map();
  ((p && p.drivers) || []).forEach((d) => {
    const n = String((d && d.driverName) || '').trim();
    if (!n) return;
    const ms = Date.parse(String(d.endDate || '').replace(/([+-]\d{2})(\d{2})$/, '$1:$2')) - Date.parse(String(d.startDate || '').replace(/([+-]\d{2})(\d{2})$/, '$1:$2'));
    t.set(n, (t.get(n) || 0) + (Number.isFinite(ms) && ms > 0 ? ms : 1));
  });
  return [...t.entries()].sort((a, b) => b[1] - a[1]).map(([n]) => n)[0] || '';
}
const PERF_COLS = [['faixa_verde', 'greenRange'], ['inicio_faixa_verde', 'extraEconomicRange'], ['final_faixa_verde', 'economicRange'], ['eco_roll', 'ecoRoll'], ['pressao_ideal', 'pedalPressureOnLow'], ['pressao_atencao', 'pedalPressureOnMid'], ['embalo', 'leverage'], ['motor_ocioso', 'idle'], ['pressao_acelerador', 'pedalPressureOnHigh'],
  ['excesso_rotacao', 'powerRange'], ['tempo_excesso_velocidade', 'speeding'], ['piloto_automatico', 'cruiseControl'], ['freio_motor', 'engineBrake']];
// Um dia de estatísticas + performance → linhas no layout do painel. O CPF dos motoristas é descartado aqui.
function diaParaLinhas(dia, estat, perf, cli) {
  const temPerf = true;   // colunas sempre iguais: dia sem performance fica com os indicadores vazios
  const cols = ['data', 'motorista', 'veiculo', 'km', 'litros', 'frenagens', 'frenagens_alta', 'velocidade_media', 'odometro']
    .concat(temPerf ? ['horas', 'nota'].concat(PERF_COLS.map(([k]) => k)) : [], ['cliente', 'dias_periodo']);
  const n = (v) => { const x = numApi(v); return x === null ? '' : String(Math.round(x * 100) / 100).replace('.', ','); };
  const escC = (s) => (/[;"\n\r]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s);
  const perfPor = new Map();
  perf.forEach((r) => { const k = chavePlaca(r && r.vehicleIdentification); if (k && !perfPor.has(k)) perfPor.set(k, r); });
  const vistos = new Set(), veic = [], mot = [];
  const soPerf = !estat.length;
  (soPerf ? perf : estat).forEach((r) => {
    const placa = placaApi(r && r.vehicleIdentification), k = chavePlaca(placa);
    if (!k || vistos.has(k)) return;
    vistos.add(k);
    const pk = perfPor.get(k), movK = pk && pk.indicators && pk.indicators.movement ? numApi(pk.indicators.movement.duration) : null;
    const km = numApi(r.totalMileage) || 0, rodou = soPerf ? (movK || 0) > 0 : km > 0;
    const p = rodou ? pk : null, ind = (p && p.indicators) || {};
    const nome = motoristaPrincipal(p);
    const pct = (nome) => { const x = ind[nome] || (nome === 'pedalPressureOnHigh' && ind.pedalPressureOnHig); return x ? n(x.percentage) : ''; };   // a API escreve pedalPressureOnHig
    const sc = p ? (p.score !== undefined ? p.score : ind.score) : null;
    const nota = numApi(sc && typeof sc === 'object' ? sc.score : sc);   // a performance não traz nota geral; a nota oficial vem mês a mês na premiação
    const mov = ind.movement ? numApi(ind.movement.duration) : null;
    const lin = [br(dia), escC(nome || placa), escC(placa), soPerf ? '' : n(km), rodou ? n(r.totalConsumption) : '', rodou ? n(r.totalBreaking) : '', rodou ? n(r.totalBreakingOnHighSpeed) : '',
      rodou ? n(r.averageSpeed) : '', n(r.odometer)]
      .concat(temPerf ? [mov === null ? '' : n(mov / 60), nota !== null && nota >= 0 && nota <= 100 ? n(nota) : ''].concat(PERF_COLS.map(([, nome]) => pct(nome))) : [], [escC(cli), '1']).join(';');
    veic.push(lin);
    if (rodou && nome) mot.push(lin);   // sem motorista identificado, a linha fica só na planilha de veículos
  });
  return { cols, veic, mot };
}
// Documento da base tem até 256 KiB: divide as linhas do dia em partes abaixo de 200 KB
function partes(cab, ls) {
  for (let n = 1; ; n++) {
    const tam = Math.ceil(ls.length / n), out = [];
    for (let i = 0; i < ls.length; i += tam) out.push([cab].concat(ls.slice(i, i + tam)).join('\n'));
    if (tam <= 1 || out.every((t) => bytes(t) < 200000)) return out;
  }
}

// Mapa ao vivo: posição atual e percurso do dia de cada veículo, em api/saida/mapa/<cliente>/ (a pasta fica entre as rodadas,
// porque o percurso continua o da rodada anterior). Um arquivo por veículo ("posicao~cliente~placa~p1") e um de controle ("posmeta").
// Devolve os arquivos que mudaram nesta rodada, para carregar na base do painel.
async function puxarMapa(c) {
  const cli = c.cliente, token = limparToken(ENV[c.token]);
  if (!token) throw new Error('sem token: preencha ' + c.token + ' em api/.env');
  const dir = path.join(SAIDA, 'mapa', slugCli(cli));
  fs.mkdirSync(dir, { recursive: true });
  const ler = (id) => { try { return JSON.parse(fs.readFileSync(path.join(dir, id + '.json'), 'utf8')); } catch (e) { return null; } };
  const agora = Date.now(), dia = diaBR(agora), ini = inicioDiaBR(dia);
  const idMeta = idPlan('posmeta', cli, 'atual', 1);
  let meta = ler(idMeta) || {};
  if (meta.dia !== dia) meta = { dia, ate: ini };
  const antes = new Map(fs.readdirSync(dir).filter((f) => f.startsWith('posicao~')).map((f) => ler(f.replace(/\.json$/, ''))).filter(Boolean)
    .flatMap((d) => d.veiculos || []).map((v) => [chavePlaca(v.placa), v]));
  if (!meta.sess || agora - (meta.perfEm || 0) > PERF_CICLO) {
    try { meta.sess = sessoesDoDia(registros(await apiGet('v1/vehicle-performance', { startDate: utc(ini), endDate: utc(agora), vehicleIdentification: '' }, token)), dia, agora); meta.perfEm = agora; }
    catch (e) { meta.sess = meta.sess || {}; }
  }
  const novos = new Map();
  const somarPts = (recs) => posicaoDosRegistros(recs).forEach((v) => { const k = chavePlaca(v.placa), p = pontoPos(v); if (p) (novos.get(k) || novos.set(k, []).get(k)).push(p); });
  let n = 0;
  for (let t = Math.max(meta.ate || ini, ini); t < agora - 60000; t += PASSO_MIN * 60000) {
    const fimJ = Math.min(t + PASSO_MIN * 60000, agora);
    if (++n % 12 === 0) process.stdout.write('  percurso: ' + new Date(fimJ - 3 * 3600000).toISOString().slice(11, 16) + '\r');
    try { somarPts(registros(await apiGet(API_POSICAO, { startDate: utc(t), endDate: utc(fimJ), vehicleIdentification: '' }, token))); }
    catch (e) { if (!/respondeu 404/.test(e.message)) throw e; }
    meta.ate = fimJ;
  }
  const recs = registros(await apiGet(API_POSICAO, { startDate: utc(agora - 30 * 86400000), endDate: utc(agora), vehicleIdentification: '' }, token));
  somarPts(recs);
  const em = new Date().toISOString(), mudaram = [];
  posicaoDosRegistros(recs).forEach((v) => {
    const k = chavePlaca(v.placa), s = meta.sess[k] || { desde: ini, mot: '' }, desdeMin = Math.floor(s.desde / 60000), ant = antes.get(k);
    const trAnt = ant && ant.dia === dia && Array.isArray(ant.tr) ? ant.tr.filter((p) => Array.isArray(p) && p[2] >= desdeMin) : [];
    v.tr = juntarPontos(trAnt, novos.get(k) || [], desdeMin);
    v.desde = s.desde; v.mot = s.mot; v.dia = dia;
    const id = idPlan('posicao', cli, k || 'sem-placa', 1);
    // só regrava o veículo quando a posição, o percurso ou o vínculo mudaram
    if (!ant || ant.data !== v.data || (ant.tr || []).length !== v.tr.length || ant.desde !== v.desde || ant.mot !== v.mot || ant.dia !== dia) {
      fs.writeFileSync(path.join(dir, id + '.json'), JSON.stringify({ alvo: 'posicao', cliente: cli, data: dia, veiculos: [v], em }));
      mudaram.push(id);
    }
  });
  fs.writeFileSync(path.join(dir, idMeta + '.json'), JSON.stringify({ alvo: 'posmeta', cliente: cli, dia, ate: meta.ate, perfEm: meta.perfEm || 0, sess: meta.sess, em }));
  mudaram.push(idMeta);
  fs.writeFileSync(path.join(dir, 'mudaram.json'), JSON.stringify(mudaram));
  return { veiculos: recs.length, mudaram: mudaram.length - 1, comPercurso: [...novos.keys()].length };
}

async function puxar(c) {
  const cli = c.cliente, varToken = c.token;
  const token = limparToken(ENV[varToken]);
  if (!token) throw new Error('sem token: preencha ' + varToken + ' em api/.env');
  const dir = path.join(SAIDA, slugCli(cli));
  fs.rmSync(dir, { recursive: true, force: true });
  fs.mkdirSync(dir, { recursive: true });
  const em = new Date().toISOString(), arquivo = 'API Gobrax';
  const gravar = (id, doc) => fs.writeFileSync(path.join(dir, id + '.json'), JSON.stringify(doc));
  const fim = hoje(), ini = somar(fim, -(DIAS - 1));
  const csv = { veiculos: [], motoristas: [] };
  let cab = null, dias = 0;
  const avisos = [], falhas = { estat: 0, perf: 0 }, veio = { estat: false, perf: false }, perdidos = { estat: [], perf: [] };
  // uma falha pontual pula só aquele dia; 404 ("Veículo não identificado") é dia sem dado; 3 falhas seguidas desligam o tipo
  const buscar = async (tipo, caminho, jan, d) => {
    if (falhas[tipo] >= 3) return [];
    try { const r = registros(await apiGet(caminho, jan, token)); falhas[tipo] = 0; if (r.length) veio[tipo] = true; return r; }
    catch (e) {
      if (/token inválido/.test(e.message)) throw e;
      if (/respondeu 404/.test(e.message)) return [];
      falhas[tipo]++; perdidos[tipo].push(br(d));
      if (falhas[tipo] >= 3) avisos.push((tipo === 'estat' ? 'estatísticas (km e litros)' : 'performance') + ' parou depois de 3 falhas seguidas: ' + e.message);
      return [];
    }
  };
  for (let d = ini; d <= fim; d = somar(d, 1)) {
    process.stdout.write('  ' + br(d) + '\r');
    const jan = Object.assign(janelaUTC(d), { vehicleIdentification: '' });
    const estat = await buscar('estat', 'v1/vehicle-statistics', jan, d);
    const perf = await buscar('perf', 'v1/vehicle-performance', jan, d);
    if (falhas.estat >= 3 && falhas.perf >= 3) throw new Error(avisos.join('; '));
    const { cols, veic, mot } = diaParaLinhas(d, estat, perf, cli);
    if (!veic.length) continue;
    dias++;
    const c1 = cols.join(';');
    cab = cab || c1;
    [['veiculos', veic], ['motoristas', mot]].forEach(([alvo, ls]) => {
      if (!ls.length) return;
      const ps = partes(c1, ls);
      ps.forEach((texto, i) => gravar(idPlan(alvo, cli, d, i + 1), { alvo, cliente: cli, data: d, colunas: cols, texto, parte: i + 1, partes: ps.length, arquivo, em }));
      if (c1 === cab) csv[alvo].push(...ls);
    });
  }
  const agora = Date.now();
  let odo = [];
  try { odo = registros(await apiGet('v2/vehicle-odometer', { startDate: utc(agora - 30 * 86400000), endDate: utc(agora), vehicleIdentification: '' }, token)); } catch (e) { avisos.push('odômetro não veio: ' + e.message); }
  const veiculos = odo.filter((r) => r && r.vehicleIdentification).map((r) => ({ placa: placaApi(r.vehicleIdentification), odometro: numApi(r.odometer), coletado: String(r.lastUpdated || '') }));
  if (veiculos.length) gravar(idPlan('odometro', cli, 'atual', 1), { alvo: 'odometro', cliente: cli, data: fim, veiculos, em });
  const mes = (s) => s.slice(5, 7) + '-' + s.slice(0, 4);
  let j = [];
  try { j = await apiGet('v2/driversOverview', { documentNumbers: '', startDate: mes(ini), endDate: mes(fim) }, token); } catch (e) { avisos.push('premiação não veio: ' + e.message); }
  // só nome, nota, km e bonificação: CPF e id do motorista são descartados antes de gravar
  const porMes = {};
  registros(j).forEach((m) => {
    const nome = String((m && (m.Name || m.name)) || '').trim();
    if (!nome) return;
    ((m.Overview || m.overview) || []).forEach((o) => {
      const mm = String(o.Date || o.date || '').match(/^(\d{2})-(\d{4})$/);
      if (!mm || o.Err) return;
      const k = mm[2] + '-' + mm[1];
      (porMes[k] = porMes[k] || []).push({ nome, nota: numApi(o.Score), km: numApi(o.TotalKM !== undefined ? o.TotalKM : o.TotalKm), premio: numApi(o.Reward) });
    });
  });
  Object.keys(porMes).forEach((k) => gravar(idPlan('premiacao', cli, k, 1), { alvo: 'premiacao', cliente: cli, data: k, motoristas: porMes[k], em }));
  // planilhas para importar no painel (GitHub Pages ou Configurações → perfil do cliente)
  ['veiculos', 'motoristas'].forEach((alvo) => { if (csv[alvo].length) fs.writeFileSync(path.join(SAIDA, slugCli(cli) + '-' + alvo + '.csv'), '﻿' + [cab].concat(csv[alvo]).join('\n')); });
  ['estat', 'perf'].forEach((t) => { if (perdidos[t].length && falhas[t] < 3) avisos.push((t === 'estat' ? 'estatísticas' : 'performance') + ' falharam em ' + perdidos[t].length + ' dia(s): ' + perdidos[t].join(', ') + ' (rode de novo para completar)'); });
  return { dias, veiculos: veiculos.length, meses: Object.keys(porMes).length, perfOk: veio.perf, estatOk: veio.estat, avisos };
}

const alvos = CLIENTES.filter((c) => !SO_CLIENTE || ckey(c.cliente) === ckey(SO_CLIENTE));
if (!alvos.length) { console.error('Nenhum cliente em api/clientes.json' + (SO_CLIENTE ? ' com o nome ' + SO_CLIENTE : '') + '.'); process.exit(1); }
let falhas = 0;
for (const c of alvos) {
  if (!SO_MAPA) {
    console.log(c.cliente + ': últimos ' + DIAS + ' dias…');
    try {
      const r = await puxar(c);
      console.log('\r  ok: ' + r.dias + ' dias com dados' + (r.estatOk ? '' : ' (só performance, sem km e litros)') + (r.perfOk ? '' : ' (sem performance)') + ', odômetro de ' + r.veiculos + ' veículos, premiação de ' + r.meses + ' meses.');
      r.avisos.forEach((a) => console.warn('  aviso: ' + a));
    } catch (e) { falhas++; console.error('\r  erro: ' + e.message); }
  }
  try {
    const m = await puxarMapa(c);
    console.log('\r  ' + c.cliente + ' · mapa: ' + m.veiculos + ' veículos, ' + m.mudaram + ' atualizados (api/saida/mapa/' + slugCli(c.cliente) + '/).');
  } catch (e) { falhas++; console.error('\r  ' + c.cliente + ' · mapa: erro: ' + e.message); }
}
if (SO_MAPA) process.exit(falhas ? 1 : 0);
// Preço do diesel por estado (Petrobras, via combustivelapi.com.br): vai para config/diesel_uf na base do painel
try {
  const r = await fetch('https://combustivelapi.com.br/api/precos/');
  const j = await r.json(), d = j && j.precos && j.precos.diesel;
  if (d) {
    const diesel = {};
    Object.entries(d).forEach(([uf, v]) => { const n = Number(String(v).replace(',', '.')); if (n > 0) diesel[uf.toLowerCase()] = n; });
    fs.mkdirSync(SAIDA, { recursive: true });
    fs.writeFileSync(path.join(SAIDA, 'diesel_uf.json'), JSON.stringify({ diesel, coleta: String(j.data_coleta || ''), fonte: String(j.fonte || 'Petrobras') }));
    console.log('Diesel por estado: ' + Object.keys(diesel).length + ' preços (coleta ' + j.data_coleta + ').');
  }
} catch (e) { console.warn('Diesel por estado não veio: ' + e.message); }
console.log('Arquivos em api/saida/.');
process.exit(falhas ? 1 : 0);
