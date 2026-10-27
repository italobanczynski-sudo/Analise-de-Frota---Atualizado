// Puxa os dados da API Gobrax de cada cliente listado em api/clientes.json, com o token lido de api/.env.
// Gera em api/saida/ os documentos no formato da base compartilhada do painel (coleção "planilhas")
// e as planilhas CSV de motoristas e veículos, que também podem ser importadas no painel.
// Uso: node api/puxar.mjs [--dias 30] [--cliente "Nome do cliente"]
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PASTA = path.dirname(fileURLToPath(import.meta.url));
const SAIDA = path.join(PASTA, 'saida');
const API_BASE = 'https://gateway-v3.gobrax.com.br:8889/api/';

const arg = (nome, padrao) => { const i = process.argv.indexOf('--' + nome); return i > 0 && process.argv[i + 1] ? process.argv[i + 1] : padrao; };
const DIAS = Math.min(93, Math.max(1, Number(arg('dias', 30)) || 30));
const SO_CLIENTE = arg('cliente', '');

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

// Motorista não é veículo: o dado do condutor é só o do tempo em que ele estava vinculado ao veículo (como na análise por
// motorista da plataforma). drivers[] traz o início e o fim de cada vínculo, recortados na janela consultada. Um único
// motorista vinculado o dia todo usa a linha do próprio veículo; nos demais casos o script consulta estatísticas e
// performance só nas janelas daquele motorista (vehicleIdentification + início e fim do vínculo).
const PERF_COLS = [['faixa_verde', 'greenRange'], ['inicio_faixa_verde', 'extraEconomicRange'], ['final_faixa_verde', 'economicRange'], ['eco_roll', 'ecoRoll'], ['pressao_ideal', 'pedalPressureOnLow'], ['pressao_atencao', 'pedalPressureOnMid'], ['embalo', 'leverage'], ['motor_ocioso', 'idle'], ['pressao_acelerador', 'pedalPressureOnHigh'],
  ['excesso_rotacao', 'powerRange'], ['tempo_excesso_velocidade', 'speeding'], ['piloto_automatico', 'cruiseControl'], ['freio_motor', 'engineBrake']];
const IND_DUR = ['movement'].concat(PERF_COLS.map(([, k]) => k));
const COLS = ['data', 'motorista', 'veiculo', 'km', 'litros', 'frenagens', 'frenagens_alta', 'velocidade_media', 'odometro', 'horas', 'nota'].concat(PERF_COLS.map(([k]) => k), ['cliente', 'dias_periodo']);
const msApi = (s) => { const t = Date.parse(String(s || '').replace(/([+-]\d{2})(\d{2})$/, '$1:$2')); return Number.isFinite(t) ? t : null; };
const msJanela = (s) => Date.parse(s.replace(' ', 'T') + 'Z');
function vinculos(p, jan) {
  const ji = msJanela(jan.startDate), jf = msJanela(jan.endDate), porNome = new Map();
  ((p && p.drivers) || []).forEach((d) => {
    const nome = String((d && d.driverName) || '').trim();
    if (!nome) return;
    const a = Math.max(ji, msApi(d.startDate) === null ? ji : msApi(d.startDate)), b = Math.min(jf, msApi(d.endDate) === null ? jf : msApi(d.endDate));
    if (b <= a) return;
    if (!porNome.has(nome)) porNome.set(nome, []);
    porNome.get(nome).push([a, b]);
  });
  const lista = [...porNome.entries()].map(([nome, janelas]) => ({ nome, janelas }));
  const diaTodo = lista.length === 1 && lista[0].janelas.some(([a, b]) => a - ji <= 60000 && jf - b <= 60000);
  return { lista, diaTodo };
}
// Soma as janelas de um motorista: km, litros e freadas somados; indicadores pela duração de cada um sobre o tempo em movimento
function juntarJanelas(res) {
  const est = { totalMileage: 0, totalConsumption: 0, totalBreaking: 0, totalBreakingOnHighSpeed: 0, averageSpeed: null, odometer: null }, dur = {};
  let kmVel = 0;
  res.forEach(({ e, p }) => {
    e.forEach((r) => {
      const km = numApi(r.totalMileage) || 0, v = numApi(r.averageSpeed), o = numApi(r.odometer);
      est.totalMileage += km; est.totalConsumption += numApi(r.totalConsumption) || 0;
      est.totalBreaking += numApi(r.totalBreaking) || 0; est.totalBreakingOnHighSpeed += numApi(r.totalBreakingOnHighSpeed) || 0;
      if (v !== null && km) { est.averageSpeed = ((est.averageSpeed || 0) * kmVel + v * km) / (kmVel + km); kmVel += km; }
      if (o !== null) est.odometer = Math.max(est.odometer || 0, o);
    });
    p.forEach((r) => {
      const ind = (r && r.indicators) || {};
      IND_DUR.forEach((k) => { const x = ind[k] || (k === 'pedalPressureOnHigh' && ind.pedalPressureOnHig); if (x) dur[k] = (dur[k] || 0) + (numApi(x.duration) || 0); });
    });
  });
  const mov = dur.movement || 0, ind = {};
  Object.keys(dur).forEach((k) => { ind[k] = { duration: dur[k], percentage: k === 'movement' ? 100 : mov ? (dur[k] / mov) * 100 : null }; });
  return { est, ind };
}
// Uma linha no layout do painel: estatísticas (r) + indicadores da performance (ind). O CPF dos motoristas nunca entra.
function linhaApi(dia, nome, placa, r, ind, cli, rodou, soPerf) {
  const n = (v) => { const x = numApi(v); return x === null ? '' : String(Math.round(x * 100) / 100).replace('.', ','); };
  const escC = (s) => (/[;"\n\r]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s);
  const pct = (k) => { const x = ind[k] || (k === 'pedalPressureOnHigh' && ind.pedalPressureOnHig); return x ? n(x.percentage) : ''; };   // a API escreve pedalPressureOnHig
  const mov = ind.movement ? numApi(ind.movement.duration) : null;
  return [br(dia), escC(nome || placa), escC(placa), soPerf ? '' : n(r.totalMileage || 0), rodou ? n(r.totalConsumption) : '', rodou ? n(r.totalBreaking) : '', rodou ? n(r.totalBreakingOnHighSpeed) : '',
    rodou ? n(r.averageSpeed) : '', n(r.odometer), mov === null ? '' : n(mov / 60), '']   // a performance não traz nota geral; a nota oficial vem mês a mês na premiação
    .concat(PERF_COLS.map(([, k]) => pct(k)), [escC(cli), '1']).join(';');
}
// Um dia: as linhas dos veículos (dia inteiro de cada um), as dos motoristas que ficaram o dia todo no veículo e,
// em "pendentes", os vínculos parciais que ainda precisam ser consultados na janela do motorista
function diaParaLinhas(dia, estat, perf, cli, jan) {
  const perfPor = new Map();
  perf.forEach((r) => { const k = chavePlaca(r && r.vehicleIdentification); if (k && !perfPor.has(k)) perfPor.set(k, r); });
  const vistos = new Set(), veic = [], mot = [], pendentes = [];
  const soPerf = !estat.length;
  (soPerf ? perf : estat).forEach((r) => {
    const placa = placaApi(r && r.vehicleIdentification), k = chavePlaca(placa);
    if (!k || vistos.has(k)) return;
    vistos.add(k);
    const pk = perfPor.get(k), movK = pk && pk.indicators && pk.indicators.movement ? numApi(pk.indicators.movement.duration) : null;
    const km = numApi(r.totalMileage) || 0, rodou = soPerf ? (movK || 0) > 0 : km > 0;
    const p = rodou ? pk : null, ind = (p && p.indicators) || {};
    const v = vinculos(p, jan);
    // na linha do veículo, o motorista é só o rótulo de quem mais o dirigiu no dia (os números são do veículo)
    const tempo = (m) => m.janelas.reduce((t, [a, b]) => t + (b - a), 0);
    const principal = v.lista.slice().sort((x, y) => tempo(y) - tempo(x))[0];
    veic.push(linhaApi(dia, principal ? principal.nome : '', placa, r, ind, cli, rodou, soPerf));
    if (!rodou) return;   // sem motorista identificado, o dia fica só na planilha de veículos
    if (v.diaTodo) mot.push(linhaApi(dia, v.lista[0].nome, placa, r, ind, cli, true, soPerf));
    else v.lista.forEach((m) => pendentes.push({ placa, nome: m.nome, janelas: m.janelas }));
  });
  return { cols: COLS, veic, mot, pendentes };
}
// Documento da base tem até 256 KiB: divide as linhas do dia em partes abaixo de 200 KB
function partes(cab, ls) {
  for (let n = 1; ; n++) {
    const tam = Math.ceil(ls.length / n), out = [];
    for (let i = 0; i < ls.length; i += tam) out.push([cab].concat(ls.slice(i, i + tam)).join('\n'));
    if (tam <= 1 || out.every((t) => bytes(t) < 200000)) return out;
  }
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
    const { cols, veic, mot, pendentes } = diaParaLinhas(d, estat, perf, cli, jan);
    if (!veic.length) continue;
    // vínculos só em parte do dia: estatísticas e performance consultadas na janela de cada motorista
    for (const v of pendentes) {
      const res = [];
      for (const [ja, jb] of v.janelas) {
        const j = { startDate: utc(ja), endDate: utc(jb), vehicleIdentification: v.placa };
        res.push({ e: estat.length ? await buscar('estat', 'v1/vehicle-statistics', j, d) : [], p: await buscar('perf', 'v1/vehicle-performance', j, d) });
      }
      const jj = juntarJanelas(res);
      const rodou = estat.length ? (jj.est.totalMileage || 0) > 0 : !!(jj.ind.movement && jj.ind.movement.duration > 0);
      if (rodou) mot.push(linhaApi(d, v.nome, v.placa, jj.est, jj.ind, cli, true, !estat.length));
    }
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
  console.log(c.cliente + ': últimos ' + DIAS + ' dias…');
  try {
    const r = await puxar(c);
    console.log('\r  ok: ' + r.dias + ' dias com dados' + (r.estatOk ? '' : ' (só performance, sem km e litros)') + (r.perfOk ? '' : ' (sem performance)') + ', odômetro de ' + r.veiculos + ' veículos, premiação de ' + r.meses + ' meses.');
    r.avisos.forEach((a) => console.warn('  aviso: ' + a));
  } catch (e) { falhas++; console.error('\r  erro: ' + e.message); }
}
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
