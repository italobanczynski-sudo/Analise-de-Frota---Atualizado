// Injeta no HTML gerado pelo harness um script que percorre todas as rotas e abas e grava um relatório no DOM
const fs = require('fs');
const [arq] = process.argv.slice(2);
const rotas = ['geral', 'carteira', 'estudo', 'dashboards', 'visao', 'visao?aba=graficos', 'visao?aba=indicadores', 'plano?aba=plano', 'individual/veiculo', 'individual/motorista', 'motoristas', 'motoristas/Teste%20A', 'premiacao', 'mensagens', 'churn', 'plano', 'comparativo', 'api', 'config', 'clientes', 'cliente', 'veiculos', 'individual/motorista/Teste%20B', 'individual/veiculo/AAA1'];
function varredura(rotas) {
  var rel = [], suspeito = /\bNaN\b|\bundefined\b|Infinity|\[object|\bnull\b/g, errosRota = [];
  window.addEventListener('error', function (e) { errosRota.push(e.message + ' @' + e.lineno); });
  function olhar(rotulo) {
    var app = document.getElementById('app'), t = app ? app.innerText : '', m = t.match(suspeito) || [], ctx = [], re = new RegExp(suspeito.source, 'g'), x;
    while ((x = re.exec(t)) && ctx.length < 4) ctx.push(t.slice(Math.max(0, x.index - 50), x.index + 20).replace(/\s+/g, ' '));
    var attr = [];
    document.querySelectorAll('#app *').forEach(function (el) {
      for (var i = 0; i < el.attributes.length; i++) {
        var a = el.attributes[i];
        if (/^(d|points|x|y|width|height|cx|cy|r|x1|x2|y1|y2|style|transform)$/.test(a.name) && /NaN|undefined|Infinity/.test(a.value) && attr.length < 3) attr.push(el.tagName + '.' + a.name + '=' + a.value.slice(0, 60));
      }
    });
    rel.push({ r: rotulo, len: t.length, erros: errosRota.splice(0), sus: m.length, ctx: ctx, attr: attr });
  }
  var SEL = '#app [role=tab],#app [data-acao=churn-vista]', fila = [];
  rotas.forEach(function (r) {
    fila.push(function () { location.hash = '#/' + r; });
    fila.push(function () {
      olhar(r);
      var abas = [].slice.call(document.querySelectorAll(SEL)), novos = [];
      abas.forEach(function (b, i) {
        var nome = b.dataset.aba || b.dataset.v || b.innerText.trim().slice(0, 20);
        novos.push(function () { var bb = document.querySelectorAll(SEL)[i]; if (bb) bb.click(); }, function () { olhar(r + ' :: ' + nome); });
      });
      fila.unshift.apply(fila, novos);
    });
  });
  function passo() {
    if (!fila.length) { var p = document.createElement('pre'); p.id = 'relatorio'; p.textContent = JSON.stringify(rel); document.body.appendChild(p); return; }
    try { fila.shift()(); } catch (e) { errosRota.push('HARNESS ' + e.message); }
    setTimeout(passo, 150);
  }
  setTimeout(passo, 800);
}
const js = '<script>(' + varredura.toString() + ')(' + JSON.stringify(rotas) + ');</script>';
let html = fs.readFileSync(arq, 'utf8');
const i = html.lastIndexOf('</body>');
html = i >= 0 ? html.slice(0, i) + js + html.slice(i) : html + js;
fs.writeFileSync(arq, html);
