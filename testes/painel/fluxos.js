// Injeta no HTML gerado por dados.js um roteiro que usa os fluxos de retenção como uma pessoa (preenche, clica, salva,
// cancela) e grava no DOM o que aconteceu em cada passo: mensagens (toasts), erros de JavaScript e verificações.
// uso: node fluxos.js <pagina.html>   (depois abrir no Chrome sem janela e ler <pre id="fluxos">)
'use strict';
const fs = require('fs');
const [arq] = process.argv.slice(2);
function roteiro() {
  var rel = [], erros = [];
  window.addEventListener('error', function (e) { erros.push(e.message + ' @' + e.lineno); });
  window.addEventListener('unhandledrejection', function (e) { erros.push('promessa: ' + ((e.reason && e.reason.message) || e.reason)); });
  var $ = function (s) { return document.querySelector(s); };
  function encher(sel, v) { var el = typeof sel === 'string' ? $(sel) : sel; if (!el) return false; el.value = v; el.dispatchEvent(new Event('input', { bubbles: true })); el.dispatchEvent(new Event('change', { bubbles: true })); return true; }
  function clicar(sel) { var el = typeof sel === 'string' ? $(sel) : sel; if (!el) return false; el.click(); return true; }
  function toasts() { return [].filter.call(document.querySelectorAll('.toast'), function (x) { if (x.dataset.visto) return false; x.dataset.visto = '1'; return true; }).map(function (x) { return x.textContent.trim(); }); }
  function col(n) { try { return JSON.parse(localStorage.getItem('gbx-col-' + n) || '[]'); } catch (e) { return []; } }
  function passo(nome, fazer, conferir, esperaErro) {
    return function () {
      var ok = true, msg = '';
      try { if (fazer() === false) { ok = false; msg = 'elemento não encontrado'; } } catch (e) { ok = false; msg = 'exceção: ' + e.message; }
      return new Promise(function (res) {
        setTimeout(function () {
          var t = toasts(), c = '';
          try { c = conferir ? conferir(t) : ''; } catch (e) { c = 'exceção na conferência: ' + e.message; }
          if (c) ok = false;
          var temErroToast = t.some(function (x) { return x.charAt(0) === '!'; });
          if (esperaErro !== undefined && esperaErro !== temErroToast) { ok = false; c = (c ? c + '; ' : '') + (esperaErro ? 'esperava mensagem de erro' : 'mensagem de erro inesperada'); }
          rel.push({ passo: nome, ok: ok, msg: msg, conf: c, toasts: t, erros: erros.splice(0) });
          res();
        }, 450);
      });
    };
  }
  var ir = function (h) { return function () { location.hash = h; }; };
  var n0 = {};
  var fila = [
    passo('abrir Combinados', ir('#/combinados'), function () { n0.c = col('combinados').length; n0.p = col('passos').length; n0.m = col('marcos').length; return $('[data-acao=comb-novo]') ? '' : 'sem botão Registrar combinado'; }),
    passo('abrir a gaveta', function () { return clicar('[data-acao=comb-novo]'); }, function () { return $('.gaveta') ? '' : 'gaveta não abriu'; }),
    passo('salvar vazio é recusado', function () { return clicar('[data-acao=comb-salvar]'); }, function () { return col('combinados').length === n0.c ? '' : 'salvou combinado vazio'; }, true),
    passo('escolher o problema', function () { return clicar('[data-acao=cf-problema][data-k=faixa_verde]'); }, function () { return $('.evid') && $('#cf-meta').value ? '' : 'sem evidência ou sem meta sugerida'; }),
    passo('preencher e salvar', function () { encher('#cf-comQuem', 'Gestor de frota'); encher('#cf-combinado', 'Treinar a troca de marcha na faixa verde'); encher('#cf-executa', 'Gestor repassa'); clicar('input[data-campo=cf-quemAvisa][value=cs]'); return clicar('[data-acao=comb-salvar]'); },
      function () { var c = col('combinados'), p = col('passos'), m = col('marcos'); return c.length === n0.c + 1 && p.length > n0.p && m.length === n0.m + 1 && !$('.gaveta') ? '' : 'combinados ' + c.length + ', passos ' + p.length + ', marcos ' + m.length; }, false),
    passo('evidência congelada no salvo', function () {}, function () { var c = col('combinados').filter(function (x) { return x.alvo === 'faixa_verde'; })[0]; return c && c.evidencia && c.evidencia.texto && c.base && c.base.p ? '' : 'sem evidência/linha de base gravada'; }),
    passo('segundo combinado ativo do mesmo indicador é recusado', function () { clicar('[data-acao=comb-novo]'); clicar('[data-acao=cf-problema][data-k=faixa_verde]'); encher('#cf-combinado', 'Outro'); encher('#cf-executa', 'X'); clicar('input[data-campo=cf-quemAvisa][value=cliente]'); return clicar('[data-acao=comb-salvar]'); }, function () { return col('combinados').length === n0.c + 1 ? '' : 'salvou duplicado'; }, true),
    passo('fechar a gaveta', function () { return clicar('[data-acao=comb-fechar]'); }, function () { return $('.gaveta') ? 'gaveta continua aberta' : ''; }),
    passo('cancelar sem motivo é recusado', function () { clicar('[data-acao=comb-cancelar]'); return clicar('[data-acao=comb-cancelar-ok]'); }, null, true),
    passo('cancelar com motivo', function () { encher('input[id^=cancel-motivo-]', 'Cliente pediu para adiar'); return clicar('[data-acao=comb-cancelar-ok]'); }, function () { return col('combinados').some(function (c) { return c.cancelado && c.motivoCancel; }) ? '' : 'não cancelou'; }, false),
    passo('registrar envio do WhatsApp', function () { var e = $('input[id^=envio-]'); if (!e) return false; encher(e, '2026-10-26'); return clicar('[data-acao=comb-envio]'); }, function () { return col('marcos').some(function (m) { return m.tipo === 'whatsapp' && m.data === '2026-10-26'; }) ? '' : 'marco do envio não criado'; }, false),
    passo('item sem verbo é recusado', function () { clicar('[data-acao=passo-novo]'); encher('#pf-acao', 'Acompanhamento do cliente'); return clicar('[data-acao=passo-salvar]'); }, null, true),
    passo('item válido é criado', function () { encher('#pf-acao', 'Revisar os veículos parados com o gestor'); encher('#pf-dono', 'Pessoa Teste'); return clicar('[data-acao=passo-salvar]'); }, function () { return col('passos').some(function (p) { return /^Revisar os ve/.test(p.acao); }) ? '' : 'item não criado'; }, false),
    passo('criar item de aviso a condutor', function () { clicar('[data-acao=passo-novo]'); encher('#pf-acao', 'Enviar o aviso aos condutores'); encher('#pf-dono', 'Pessoa Teste'); var c = document.querySelector('input[data-campo=pf-aviso]'); c.checked = true; c.dispatchEvent(new Event('input', { bubbles: true })); return clicar('[data-acao=passo-salvar]'); }, function () { return col('passos').some(function (p) { return p.aviso && /^Enviar o aviso/.test(p.acao); }) ? '' : 'item de aviso não criado'; }, false),
    passo('aviso a condutor não fecha sem data de envio', function () { var b = [].filter.call(document.querySelectorAll('.passo'), function (x) { return /aviso a condutor/.test(x.textContent) && x.querySelector('[data-acao=passo-acao][data-tipo=feito]'); })[0]; if (!b) return false; clicar(b.querySelector('[data-acao=passo-acao][data-tipo=feito]')); return clicar('[data-acao=passo-concluir][data-tipo=feito]'); }, null, true),
    passo('registrar marco', function () { clicar('[data-acao=passo-fechar]'); encher('#marco-data', '2026-10-28'); return clicar('[data-acao=marco-salvar]'); }, function () { return col('marcos').some(function (m) { return m.data === '2026-10-28'; }) ? '' : 'marco não salvo'; }, false),
    passo('abrir Saúde do cliente', ir('#/churn'), function () { return $('#ficha') && $('#passos') && $('#combinados') ? '' : 'faltam blocos (ficha/passos/combinados)'; }),
    passo('editar e salvar a ficha', function () { clicar('[data-acao=ficha-editar]'); encher('#fi-proximo', '2026-11-10'); return clicar('[data-acao=ficha-salvar]'); }, function () { var f = col('fichas')[0]; return f && f.proximo === '2026-11-10' && f.historico.length > 1 ? '' : 'ficha não atualizada'; }, false),
    passo('ficha com concorrente sem alternativa é recusada', function () { clicar('[data-acao=ficha-editar]'); encher('#fi-alternativa', ''); return clicar('[data-acao=ficha-salvar]'); }, null, true),
    passo('fechar edição da ficha', function () { return clicar('[data-acao=ficha-cancelar]'); }),
    passo('qualidade: confirmar sem motivo é recusado', function () { location.hash = '#/visao?aba=qualidade'; }, null),
    passo('abrir tratamento do suspeito', function () { return clicar('[data-acao=susp-abrir][data-tipo=ignorar]'); }),
    passo('salvar sem motivo', function () { return clicar('[data-acao=susp-salvar]'); }, null, true),
    passo('salvar com motivo', function () { encher('#susp-motivo', 'Litros zerados pela unidade'); return clicar('[data-acao=susp-salvar]'); }, function () { return col('suspeitos_trat').length === 1 ? '' : 'tratamento não gravado'; }, false),
    passo('abrir Resumo da reunião', ir('#/resumo'), function () { return $('.res-topo') || $('[data-acao=lib-salvar]') ? '' : 'resumo vazio'; }),
    passo('salvar premissas', function () { encher('#pr-preco', '7,10'); return clicar('[data-acao=premissas-salvar]'); }, function () { var p = col('premissas')[0]; return p && Math.abs(p.preco - 7.1) < 1e-9 ? '' : 'premissa não salva'; }, false),
    passo('abrir Comparação por fase', ir('#/fases'), function () { return $('.fase-card') ? '' : 'sem cartões de fase'; }),
    passo('trocar indicador', function () { return clicar('[data-acao=fase-ind][data-k=kml]'); }, function () { return $('.fase-card') ? '' : 'sem cartões'; }),
    passo('excluir dia com motivo', function () { encher('#excl-dia', '2026-10-12'); encher('#excl-motivo', 'Feriado'); return clicar('[data-acao=excl-salvar]'); }, function () { return col('dias_excluidos').length === 1 ? '' : 'dia não excluído'; }, false),
    passo('Plano de ação: registrar combinado abre a gaveta', function () { location.hash = '#/plano?aba=plano'; }),
    passo('clicar Registrar combinado do plano', function () { return clicar('[data-acao=plano-comb]'); }),
    passo('gaveta aberta em Combinados', function () {}, function () { return location.hash === '#/combinados' && $('.gaveta') ? '' : 'não abriu (hash ' + location.hash + ')'; }),
    passo('fechar', function () { return clicar('[data-acao=comb-fechar]'); }),
    passo('Configurações: regras sem mudança', ir('#/config'), function () { return $('#regras') ? '' : 'sem cartão de regras'; }),
    passo('salvar regras sem mudança', function () { return clicar('[data-acao=regras-salvar]'); }, function (t) { return t.some(function (x) { return /Nada mudou/.test(x); }) ? '' : 'mensagem inesperada'; }),
    passo('Carteira mostra retenção', ir('#/carteira'), function () { return /pendência/.test(document.body.textContent) ? '' : 'sem coluna de retenção'; })
  ];
  var i = 0;
  (function prox() {
    if (i >= fila.length) { var p = document.createElement('pre'); p.id = 'fluxos'; p.textContent = JSON.stringify(rel); document.body.appendChild(p); return; }
    fila[i++]().then(prox);
  })();
}
let html = fs.readFileSync(arq, 'utf8');
const js = '<script>setTimeout(function(){(' + roteiro.toString() + ')();},1500);</script>';
const i = html.lastIndexOf('</body>');
fs.writeFileSync(arq, html.slice(0, i) + js + html.slice(i));
