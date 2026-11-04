// Teste do painel (index.html): sintaxe dos scripts e varredura de todas as telas e abas num Chrome sem janela,
// com dados sintéticos (com e sem Nota Gobrax). Falha se alguma tela der erro de JavaScript ou mostrar NaN/undefined.
// uso: node testes/painel/rodar.js   (CHROME=/caminho/do/chrome para escolher o navegador)
'use strict';
const fs = require('fs');
const os = require('os');
const path = require('path');
const { execFileSync } = require('child_process');

const RAIZ = path.resolve(__dirname, '..', '..');
const AQUI = __dirname;
const chrome = process.env.CHROME || ['google-chrome', 'google-chrome-stable', 'chromium', 'chromium-browser'].find((c) => {
  try { execFileSync(c, ['--version'], { stdio: 'ignore' }); return true; } catch (e) { return false; }
});

// 1) sintaxe de cada <script> do index.html
execFileSync(process.execPath, [path.join(AQUI, 'sintaxe.js'), path.join(RAIZ, 'index.html')], { stdio: 'inherit' });
if (!chrome) { console.error('Chrome não encontrado: defina CHROME com o caminho do navegador.'); process.exit(1); }

// 2) varredura das telas
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'painel-'));
fs.cpSync(path.join(RAIZ, 'bibliotecas'), path.join(tmp, 'bibliotecas'), { recursive: true });
const decod = (t) => t.replace(/&quot;/g, '"').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
let falhas = 0;
for (const [semNota, comCliente] of [['0', '1'], ['1', '1'], ['1', '0']]) {
  const pagina = path.join(tmp, `p-${semNota}${comCliente}.html`);
  execFileSync(process.execPath, [path.join(AQUI, 'dados.js'), path.join(RAIZ, 'index.html'), pagina, semNota, comCliente]);
  execFileSync(process.execPath, [path.join(AQUI, 'varredura.js'), pagina]);
  const dom = execFileSync(chrome, ['--headless=new', '--disable-gpu', '--no-sandbox', '--virtual-time-budget=120000', '--dump-dom', 'file://' + pagina + '#/geral'],
    { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024, stdio: ['ignore', 'pipe', 'ignore'] });   // avisos do Chrome (dbus, GPU) não interessam
  const m = dom.match(/<pre id="relatorio">([\s\S]*?)<\/pre>/);
  const cenario = `nota ${semNota === '1' ? 'vazia' : 'presente'}, ${comCliente === '1' ? 'com' : 'sem'} cliente cadastrado`;
  if (!m) { console.log(`XX  ${cenario}: a varredura não terminou (${(dom.match(/data-erros="([^"]*)"/) || [])[1] || 'sem detalhe'})`); falhas++; continue; }
  const rel = JSON.parse(decod(m[1]));
  const ruins = rel.filter((x) => x.erros.length || x.sus || x.attr.length);
  console.log(`${ruins.length ? 'XX' : 'ok'}  ${cenario}: ${rel.length} telas e abas`);
  ruins.forEach((x) => console.log('      ' + x.r + ' ' + JSON.stringify({ erros: x.erros, texto: x.ctx, atributos: x.attr })));
  falhas += ruins.length;
}
// 3) fluxos de retenção usados como uma pessoa: combinado, próximos passos, ficha, dado suspeito, resumo, fases, regras
{
  const pagina = path.join(tmp, 'fluxos.html');
  execFileSync(process.execPath, [path.join(AQUI, 'dados.js'), path.join(RAIZ, 'index.html'), pagina, '0', '1']);
  execFileSync(process.execPath, [path.join(AQUI, 'fluxos.js'), pagina]);
  const dom = execFileSync(chrome, ['--headless=new', '--disable-gpu', '--no-sandbox', '--virtual-time-budget=60000', '--dump-dom', 'file://' + pagina + '#/geral'],
    { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024, stdio: ['ignore', 'pipe', 'ignore'] });
  const m = dom.match(/<pre id="fluxos">([\s\S]*?)<\/pre>/);
  if (!m) { console.log('XX  fluxos: o roteiro não terminou'); falhas++; }
  else {
    const rel = JSON.parse(decod(m[1]));
    const ruins = rel.filter((x) => !x.ok || x.erros.length);
    console.log(`${ruins.length ? 'XX' : 'ok'}  fluxos de retenção: ${rel.length} passos`);
    ruins.forEach((x) => console.log('      ' + x.passo + ' ' + JSON.stringify({ msg: x.msg, conferencia: x.conf, erros: x.erros, mensagens: x.toasts })));
    falhas += ruins.length;
  }
}
fs.rmSync(tmp, { recursive: true, force: true });
process.exit(falhas ? 1 : 0);
