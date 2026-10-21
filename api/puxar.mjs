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
  const temPerf = perf.length > 0;
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
    const nota = numApi(sc && typeof sc === 'object' ? sc.score : sc);
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
  let cab = null, perfOk = true, estatOk = true, dias = 0;
  const avisos = [];
  for (let d = ini; d <= fim; d = somar(d, 1)) {
    process.stdout.write('  ' + br(d) + '\r');
    const jan = Object.assign(janelaUTC(d), { vehicleIdentification: '' });
    let estat = [], perf = [];
    if (estatOk) { try { estat = registros(await apiGet('v1/vehicle-statistics', jan, token)); } catch (e) { if (/token inválido/.test(e.message)) throw e; estatOk = false; avisos.push('estatísticas (km e litros) não vieram: ' + e.message); } }
    if (perfOk) { try { perf = registros(await apiGet('v1/vehicle-performance', jan, token)); } catch (e) { if (/token inválido/.test(e.message)) throw e; perfOk = false; avisos.push('performance não veio: ' + e.message); } }
    if (!estatOk && !perfOk) throw new Error(avisos.join('; '));
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
  return { dias, veiculos: veiculos.length, meses: Object.keys(porMes).length, perfOk, estatOk, avisos };
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
console.log('Arquivos em api/saida/.');
process.exit(falhas ? 1 : 0);
