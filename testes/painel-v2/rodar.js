// Teste do painel novo (painel-v2/dist/analise-de-frota.html) num Chrome sem janela.
// 1) modo demonstração (arquivo aberto direto, dados simulados)
// 2) modo plataforma: um servidor falso responde as rotas do servidor Go com dados sintéticos no mesmo formato
//    (inclusive campos nulos, dias sem litros, motorista sem nota, veículo sem posição).
// Percorre todas as telas e abas e falha se alguma der erro de JavaScript ou mostrar NaN/undefined/Infinity.
// uso: node testes/painel-v2/rodar.js      (CHROME=/caminho/do/chrome para escolher o navegador)
'use strict';
const fs = require('fs');
const os = require('os');
const path = require('path');
const http = require('http');
const { execFile, execFileSync } = require('child_process');

const RAIZ = path.resolve(__dirname, '..', '..');
const HTML = path.join(RAIZ, 'painel-v2', 'dist', 'analise-de-frota.html');
const chrome = process.env.CHROME || ['google-chrome', 'google-chrome-stable', 'chromium', 'chromium-browser',
  'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe', 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe'].find((c) => {
  if (path.isAbsolute(c)) return fs.existsSync(c);
  try { execFileSync(c, ['--version'], { stdio: 'ignore' }); return true; } catch (e) { return false; }
});
if (!chrome) { console.error('Chrome não encontrado: defina CHROME com o caminho do navegador.'); process.exit(1); }

const ROTAS = ['geral', 'retencao', 'retencao?aba=carteira', 'plano', 'estudo', 'frota', 'frota?aba=indicadores', 'frota?aba=qualidade', 'frota?aba=mapa',
  'motoristas', 'motoristas?aba=comportamento', 'motoristas?aba=comparativo', 'motoristas?aba=semanal', 'motoristas?aba=lista', 'motoristas?aba=mensagens',
  'motoristas?aba=premiacao', 'dados', 'dados?aba=integracao', 'dados?aba=diagnostico'];

/* roda dentro da página: percorre rotas, abas, segmentos e o painel do condutor; grava o relatório no DOM */
function varredura(rotas) {
  var rel = [], erros = [], suspeito = /\bNaN\b|\bundefined\b|Infinity|\[object/g;
  window.addEventListener('error', function (e) { erros.push(e.message + ' @' + e.lineno); });
  window.addEventListener('unhandledrejection', function (e) { erros.push('promise: ' + (e.reason && e.reason.message || e.reason)); });
  function olhar(rotulo) {
    var v = document.getElementById('view'), t = v ? v.innerText : '', ctx = [], re = new RegExp(suspeito.source, 'g'), x;
    while ((x = re.exec(t)) && ctx.length < 4) ctx.push(t.slice(Math.max(0, x.index - 60), x.index + 20).replace(/\s+/g, ' '));
    var attr = [];
    document.querySelectorAll('#view *').forEach(function (el) {
      for (var i = 0; i < el.attributes.length; i++) {
        var a = el.attributes[i];
        if (/^(d|points|x|y|width|height|cx|cy|r|x1|x2|y1|y2|style|transform|data-tt)$/.test(a.name) && /NaN|undefined|Infinity/.test(a.value) && attr.length < 3) attr.push(el.tagName + '.' + a.name + '=' + a.value.slice(0, 80));
      }
    });
    var falhou = /Algo deu errado/.test(t);
    // campos sem valor ("—") e a Nota Gobrax da Visão geral, que precisa aparecer quando a API traz a premiação
    var vazios = (t.match(/—/g) || []).length;
    var nota = document.querySelector('#view .card.yel .gnota .gv');
    rel.push({ r: rotulo, len: t.length, erros: erros.splice(0), sus: ctx.length, ctx: ctx, attr: attr, falhou: falhou, vazios: vazios, nota: nota ? nota.innerText.trim() : null });
  }
  var SEL = '#view [role=tab],#view .seg button', fila = [];
  rotas.forEach(function (r) {
    fila.push(function () { location.hash = '#/' + r; });
    fila.push(function () {
      olhar(r);
      var n = document.querySelectorAll(SEL).length, novos = [];
      for (var i = 0; i < n && i < 24; i++) (function (i) {
        novos.push(function () { var b = document.querySelectorAll(SEL)[i]; if (b) b.click(); }, function () { olhar(r + ' :: botão ' + i); });
      })(i);
      fila.unshift.apply(fila, novos);
    });
  });
  // contrato editável no Estudo: grava 200 por unidade × 10 unidades e confere a mensalidade de R$ 2.000,00
  fila.push(function () { location.hash = '#/estudo'; }, function () {
    var v = document.getElementById('ct-valor'), q = document.getElementById('ct-qtd'), b = document.querySelector('[data-act=contrato-salvar]');
    if (v && q && b) { v.value = '200'; q.value = '10'; b.click(); }
  }, function () {}, function () {
    var t = (document.getElementById('view') || {}).innerText || '';
    rel.push({ r: 'estudo · salvar contrato', len: t.length, erros: erros.splice(0), sus: 0, ctx: [], attr: [], falhou: !/2\.000,00/.test(t) });
  });
  // gaveta de veículo e troca de cliente
  fila.push(function () { location.hash = '#/geral'; }, function () { location.hash = '#/frota?aba=visao'; }, function () { var t = document.querySelector('#view tr.click'); if (t) t.click(); }, function () {
    var d = document.querySelector('.drawer'); rel.push({ r: 'gaveta do veículo', len: d ? d.innerText.length : 0, erros: erros.splice(0), sus: d && /NaN|undefined/.test(d.innerText) ? 1 : 0, ctx: [], attr: [], falhou: !d });
  });
  fila.push(function () { var s = document.getElementById('clientSel'); if (s && s.options.length > 1) { s.selectedIndex = 1; s.dispatchEvent(new Event('change', { bubbles: true })); } }, function () {}, function () {}, function () {}, function () { olhar('outro cliente · geral'); });
  function passo() {
    if (!fila.length) { var p = document.createElement('pre'); p.id = 'relatorio'; p.textContent = JSON.stringify(rel); document.body.appendChild(p); return; }
    try { fila.shift()(); } catch (e) { erros.push('HARNESS ' + e.message); }
    setTimeout(passo, 350);
  }
  setTimeout(passo, 2500);
}
const injetar = (html) => html.replace('</body>', '<script>(' + varredura.toString() + ')(' + JSON.stringify(ROTAS) + ');</script></body>');

/* ---------- dados sintéticos no formato do servidor Go ---------- */
function dadosServidor() {
  const hoje = new Date(new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo' }).format(new Date()) + 'T12:00:00Z');
  const dia = (n) => new Date(hoje.getTime() - n * 864e5).toISOString().slice(0, 10);
  let s = 7; const rnd = () => ((s = (s * 16807) % 2147483647) / 2147483647);
  // semPerf: como um cliente real em que a performance não veio: sem motorista, faixa verde e motor ocioso nos registros,
  // mas com a premiação do mês (a Nota Gobrax tem de aparecer mesmo assim)
  const cli = (id, nome, nVeic, semNota, semPerf) => {
    const registros = [], motoristas = [], placas = [];
    for (let v = 0; v < nVeic; v++) placas.push('TST' + v + 'A' + (10 + v));
    placas.forEach((p, v) => {
      for (let n = 90; n >= 0; n--) {
        if (v === nVeic - 1 && n < 6) continue;                 // um veículo sem sinal
        if (rnd() < 0.25) continue;
        const km = Math.round(200 + rnd() * 300), rodou = rnd() > 0.05;
        const r = { cliente_id: id, dia: dia(n) + 'T00:00:00Z', placa: p, motorista: v === 1 ? '' : 'Motorista ' + p, rodou, km: rodou ? km : 0,
          litros: rnd() < 0.05 ? null : +(km / (1.4 + rnd() * 0.5)).toFixed(1), frenagens: rnd() < 0.1 ? null : Math.round(km * 0.2),
          frenagens_alta: Math.round(rnd() * 3), velocidade_media: 40 + rnd() * 10, odometro: 100000 + n, horas: 8,
          nota: null /* como no servidor real: a performance diária não traz nota */, faixa_verde: v === 2 ? null : 50 + rnd() * 30, motor_ocioso: 5 + rnd() * 10 };
        r.inicio_faixa_verde = v === 2 ? null : 40 + rnd() * 55;
        if (semPerf) { r.motorista = ''; r.faixa_verde = null; r.motor_ocioso = null; r.inicio_faixa_verde = null; }
        registros.push(r);
        if (r.motorista) motoristas.push(r);
      }
    });
    return { id, nome, registros, motoristas,
      odometros: placas.map((p) => ({ placa: p, odometro: 123456, coletado: dia(0) + ' 08:00:00', buscado_em: hoje.toISOString() })),
      // a Nota Gobrax real chega só aqui: por motorista e mês (v2/driversOverview)
      premiacoes: semNota ? [] : [0, 31, 62, 93].flatMap((n) => placas.filter((p, i) => i !== 1).map((p) => ({ mes: dia(n).slice(0, 7), motorista: 'Motorista ' + p,
        nota: 35 + Math.round(rnd() * 60), km: 5000, premio: rnd() < 0.5 ? 150 : null }))) };
  };
  return [cli(1, 'Cliente Teste', 8, false), cli(4, 'Cliente Sem Performance', 5, false, true), cli(2, 'Cliente Sem Nota', 4, true), { id: 3, nome: 'Cliente Novo', registros: [], motoristas: [], odometros: [], premiacoes: [] }];
}
function servidor(html) {
  const dados = dadosServidor();
  // premissas gravadas pelo painel clássico (mesmo formato): o Cliente Teste já tem diesel próprio
  const premissas = new Map([['pr-cliente-teste', { cliente: 'Cliente Teste', preco: 6.5, valorEquip: null, qtd: null, historico: [] }]]);
  const ok = (res, d) => { res.writeHead(200, { 'Content-Type': 'application/json' }); res.end(JSON.stringify({ ok: true, dados: d })); };
  return http.createServer((req, res) => {
    const u = new URL(req.url, 'http://x');
    if (u.pathname === '/painel') { res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' }); return res.end(html); }
    if (u.pathname === '/api/v1/auth/eu') return ok(res, { id: 1, nome: 'Pessoa Teste', papel: 'admin' });
    if (u.pathname === '/api/v1/painel/dados') return ok(res, dados);
    if (u.pathname === '/api/v1/base/premissas') return ok(res, [...premissas.entries()].map(([id, d]) => ({ id, dados: d })));
    const mp = u.pathname.match(/^\/api\/v1\/base\/premissas\/(.+)$/);
    if (mp && req.method === 'PUT') { let b = ''; req.on('data', (c) => { b += c; }); req.on('end', () => { premissas.set(decodeURIComponent(mp[1]), JSON.parse(b)); ok(res, null); }); return; }
    let m = u.pathname.match(/^\/api\/v1\/clientes\/(\d+)\/(ao-vivo|mensagens\/enviadas|condutores)$/);
    if (m) {
      const c = dados.find((x) => String(x.id) === m[1]);
      if (m[2] === 'ao-vivo') {
        const placas = (u.searchParams.get('placas') || '').split(',');
        return ok(res, { veiculos: placas.map((p, i) => i === 0 ? { placa: p, status: 'sem_sinal' } : { placa: p, status: ['movimento', 'ligado', 'desligado'][i % 3], lat: -23.5 - i * 0.1, lon: -46.6 - i * 0.1, min_atras: 3, velocidade: 60 }),
          rastro: { [placas[2]]: [{ data: '2026-01-01 10:00:00', lat: -23.5, lon: -46.6, status: 'movimento' }, { data: '2026-01-01 10:05:00', lat: -23.6, lon: -46.7, status: 'ligado' }] }, completo: true });
      }
      if (m[2] === 'condutores') return ok(res, [{ motorista: 'Motorista TST0A10', telefone: '5511999990000', dias: 3 }]);
      const mot = c.motoristas.filter((r) => r.motorista);
      const nomes = [...new Set(mot.map((r) => r.motorista))].slice(0, 3);
      return ok(res, nomes.map((n) => ({ motorista: n, dias_enviada: 3, indicadores: [], envios: mot.filter((r) => r.motorista === n).slice(-12, -1).filter((_, i) => i % 3 === 0)
        .map((r, i) => ({ dia: r.dia.slice(0, 10), nota: i === 0 ? null : 60 + i, foco: 'faixa_verde', foco_rotulo: 'Faixa verde', indicadores: ['Faixa verde'] })) })));
    }
    res.writeHead(404); res.end();
  });
}

function chromeDom(url) {
  const perfil = fs.mkdtempSync(path.join(os.tmpdir(), 'chrome-'));
  return new Promise((ok, falha) => execFile(chrome, ['--headless=new', '--disable-gpu', '--no-sandbox', '--user-data-dir=' + perfil, '--virtual-time-budget=180000', '--dump-dom', url],
    { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024, timeout: 240000 }, (e, out) => { try { fs.rmSync(perfil, { recursive: true, force: true }); } catch (x) {} e && !out ? falha(e) : ok(out); }));
}
const decod = (t) => t.replace(/&quot;/g, '"').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&#39;/g, "'").replace(/&amp;/g, '&');
function relatorio(cenario, dom) {
  const m = dom.match(/<pre id="relatorio">([\s\S]*?)<\/pre>/);
  if (!m) { console.log(`XX  ${cenario}: a varredura não terminou`); return 1; }
  const rel = JSON.parse(decod(m[1]));
  // Visão geral do primeiro cliente e do segundo (sem performance): a Nota Gobrax da frota tem de ter número
  const semNota = (x) => (x.r === 'geral' || x.r === 'outro cliente · geral') && x.nota !== null && !/\d/.test(x.nota);
  const ruins = rel.filter((x) => x.erros.length || x.sus || x.attr.length || x.falhou || x.len < 20 || semNota(x));
  console.log(`${ruins.length ? 'XX' : 'ok'}  ${cenario}: ${rel.length} telas, abas e botões`);
  ruins.forEach((x) => console.log('      ' + x.r + ' ' + JSON.stringify({ erros: x.erros, texto: x.ctx, atributos: x.attr, falhou: x.falhou, tamanho: x.len, nota: x.nota })));
  // onde há mais campos sem valor: para revisar se é falta de dado da API ou falha do painel
  const vaz = rel.filter((x) => !x.r.includes('::') && x.vazios > 0).sort((a, b) => b.vazios - a.vazios).slice(0, 6);
  if (vaz.length) console.log('      campos "—" por tela: ' + vaz.map((x) => x.r + '=' + x.vazios).join(', '));
  return ruins.length;
}

/* node testes/painel-v2/rodar.js --fotos <pasta>: capturas de tela do modo plataforma, sem a varredura */
async function fotos(dir) {
  fs.mkdirSync(dir, { recursive: true });
  const srv = servidor(fs.readFileSync(HTML, 'utf8'));
  await new Promise((r) => srv.listen(0, '127.0.0.1', r));
  for (const r of ['geral', 'dados?aba=diagnostico', 'frota?aba=mapa', 'motoristas', 'motoristas?aba=lista', 'motoristas?aba=mensagens', 'motoristas?aba=premiacao', 'estudo', 'plano']) {
    const perfil = fs.mkdtempSync(path.join(os.tmpdir(), 'chrome-'));
    const arq = path.join(dir, r.replace(/\W+/g, '-') + '.png');
    await new Promise((ok) => execFile(chrome, ['--headless=new', '--disable-gpu', '--no-sandbox', '--hide-scrollbars', '--user-data-dir=' + perfil, '--window-size=1440,2600',
      '--virtual-time-budget=15000', '--screenshot=' + arq, 'http://127.0.0.1:' + srv.address().port + '/painel#/' + r], { timeout: 120000 }, () => ok()));
    try { fs.rmSync(perfil, { recursive: true, force: true }); } catch (x) {}
    console.log(arq);
  }
  srv.close();
}

(async () => {
  if (!fs.existsSync(HTML)) { console.error('Rode antes: python painel-v2/build.py'); process.exit(1); }
  const iF = process.argv.indexOf('--fotos');
  if (iF > 0) { await fotos(process.argv[iF + 1] || path.join(os.tmpdir(), 'painel-v2-fotos')); return; }
  const html = injetar(fs.readFileSync(HTML, 'utf8'));
  let falhas = 0;
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'painel-v2-'));
  const arq = path.join(tmp, 'painel.html');
  fs.writeFileSync(arq, html);
  falhas += relatorio('modo demonstração (dados simulados)', await chromeDom('file:///' + arq.replace(/\\/g, '/') + '#/geral'));
  const srv = servidor(html);
  await new Promise((r) => srv.listen(0, '127.0.0.1', r));
  falhas += relatorio('modo plataforma (dados no formato do servidor)', await chromeDom('http://127.0.0.1:' + srv.address().port + '/painel#/geral'));
  srv.close();
  fs.rmSync(tmp, { recursive: true, force: true });
  process.exit(falhas ? 1 : 0);
})();
