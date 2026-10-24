'use strict';

let eu = null;
const $ = (id) => document.getElementById(id);

// ---------- início: quem está logado ----------
(async function iniciar() {
  try {
    eu = await api('GET', '/api/v1/auth/eu');
  } catch (e) {
    location.replace('/');
    return;
  }
  $('avatar').textContent = iniciais(eu.nome);
  $('quem-nome').textContent = eu.nome;
  $('quem-papel').textContent = PAPEIS[eu.papel] || eu.papel;
  $('aba-usuarios').hidden = eu.papel !== 'admin';
  abrirAba(location.hash === '#usuarios' && eu.papel === 'admin' ? 'usuarios' : 'clientes');
})();

$('sair').addEventListener('click', async () => {
  try { await api('POST', '/api/v1/auth/sair'); } catch (e) { /* sai mesmo assim */ }
  location.replace('/');
});

// ---------- abas ----------
function abrirAba(nome) {
  ['clientes', 'usuarios'].forEach((n) => {
    $('aba-' + n).setAttribute('aria-selected', String(n === nome));
    $('sec-' + n).hidden = n !== nome;
  });
  history.replaceState(null, '', nome === 'usuarios' ? '#usuarios' : location.pathname);
  if (nome === 'clientes') carregarClientes();
  else carregarUsuarios();
}
window.addEventListener('hashchange', () => { if (eu) abrirAba(location.hash === '#usuarios' && eu.papel === 'admin' ? 'usuarios' : 'clientes'); });
$('aba-clientes').addEventListener('click', () => abrirAba('clientes'));
$('aba-usuarios').addEventListener('click', () => abrirAba('usuarios'));

// ---------- clientes ----------
const SITUACAO = {
  ok: ['Em dia', 'selo-ok'],
  rodando: ['Buscando…', 'selo-atencao'],
  erro: ['Com erro', 'selo-erro'],
  nunca: ['Nunca buscado', 'selo-neutro'],
};
let espera = null;

async function carregarClientes() {
  const corpo = $('lista-clientes');
  let clientes;
  try {
    clientes = await api('GET', '/api/v1/clientes');
  } catch (e) {
    corpo.replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: e.message })));
    return;
  }
  if (!clientes.length) {
    corpo.replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: 'Nenhum cliente cadastrado ainda.' })));
    return;
  }
  const podeBuscar = eu.papel === 'admin' || eu.papel === 'colaborador';
  corpo.replaceChildren(...clientes.map((c) => {
    const col = c.coleta || {};
    const [rot, cls] = SITUACAO[col.situacao] || SITUACAO.nunca;
    const selo = el('span', { class: 'selo ' + cls, title: col.mensagem || '' }, rot);
    const acoes = el('div', { class: 'acoes' });
    if (podeBuscar && c.tem_token) {
      acoes.append(el('button', { class: 'btn btn-peq', type: 'button', disabled: col.situacao === 'rodando', onclick: () => buscarAgora(c) },
        col.situacao === 'rodando' ? 'Buscando…' : 'Buscar agora'));
    } else if (!c.tem_token) {
      acoes.append(el('span', { class: 'ajuda', text: 'Sem token da API' }));
    }
    return el('tr', {},
      el('td', {}, el('div', { class: 'pessoa' }, el('span', { class: 'avatar', 'aria-hidden': 'true' }, iniciais(c.nome)), el('div', {}, el('b', { text: c.nome }), col.situacao === 'erro' && col.mensagem ? el('span', { text: col.mensagem }) : null))),
      el('td', {}, selo),
      el('td', { class: 'num', text: diaBR(col.ultimo_dia) }),
      el('td', { class: 'num', text: dataHora(col.terminada_em) }),
      el('td', {}, acoes));
  }));
  // enquanto alguma busca roda, atualiza a lista a cada 5 s
  clearTimeout(espera);
  if (clientes.some((c) => c.coleta && c.coleta.situacao === 'rodando') && !$('sec-clientes').hidden) espera = setTimeout(carregarClientes, 5000);
}

async function buscarAgora(c) {
  try {
    await api('POST', '/api/v1/clientes/' + c.id + '/coleta');
    toast('Busca de ' + c.nome + ' iniciada.');
  } catch (e) {
    toast(e.message, true);
  }
  carregarClientes();
}

// ---------- usuários (só administrador) ----------
async function carregarUsuarios() {
  const corpo = $('lista-usuarios');
  let usuarios;
  try {
    usuarios = await api('GET', '/api/v1/usuarios');
  } catch (e) {
    corpo.replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: e.message })));
    return;
  }
  corpo.replaceChildren(...usuarios.map((u) => {
    const souEu = u.id === eu.id;
    const situacao = !u.ativo ? el('span', { class: 'selo selo-neutro' }, 'Desativado')
      : !u.tem_senha ? el('span', { class: 'selo selo-atencao' }, 'Convite pendente')
        : el('span', { class: 'selo selo-ok' }, 'Ativo');
    const papel = el('select', { class: 'inline', 'aria-label': 'Perfil de ' + u.nome, disabled: !u.ativo, onchange: (ev) => alterar(u, ev.target.value, u.ativo) },
      Object.entries(PAPEIS).map(([v, r]) => el('option', { value: v, selected: v === u.papel }, r)));
    const acoes = el('div', { class: 'acoes' });
    if (u.ativo) acoes.append(el('button', { class: 'btn-link btn', type: 'button', onclick: () => novoLink(u) }, u.tem_senha ? 'Novo link de senha' : 'Novo link de convite'));
    if (!souEu) {
      acoes.append(u.ativo
        ? el('button', { class: 'btn-link btn btn-perigo', type: 'button', onclick: () => alterar(u, u.papel, false) }, 'Desativar')
        : el('button', { class: 'btn-link btn', type: 'button', onclick: () => alterar(u, u.papel, true) }, 'Reativar'));
    }
    return el('tr', {},
      el('td', {}, el('div', { class: 'pessoa' }, el('span', { class: 'avatar', 'aria-hidden': 'true' }, iniciais(u.nome)),
        el('div', {}, el('b', { text: u.nome + (souEu ? ' (você)' : '') }), el('span', { text: u.email })))),
      el('td', {}, papel),
      el('td', {}, situacao),
      el('td', { class: 'num', text: dataHora(u.ultimo_acesso) }),
      el('td', {}, acoes));
  }));
}

function mostrarLink(dados, texto) {
  $('link-texto').textContent = texto;
  $('link-valor').textContent = dados.link_convite;
  $('link-convite').hidden = false;
  $('link-convite').scrollIntoView({ behavior: 'smooth', block: 'center' });
}

$('form-convidar').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const nome = $('novo-nome').value.trim();
  const email = $('novo-email').value.trim();
  const papel = $('novo-papel').value;
  const erro = $('erro-convite');
  if (!nome || !email) {
    mostrarErro(erro, 'Informe o nome e o e-mail da pessoa.');
    return;
  }
  mostrarErro(erro, '');
  $('botao-convidar').disabled = true;
  try {
    const d = await api('POST', '/api/v1/usuarios', { nome, email, papel });
    mostrarLink(d, 'Convite criado para ' + d.usuario.nome + ' (' + PAPEIS[d.usuario.papel] + '). Vale até ' + dataHora(d.expira_em) + '.');
    $('form-convidar').reset();
    carregarUsuarios();
  } catch (e) {
    mostrarErro(erro, e.message);
  }
  $('botao-convidar').disabled = false;
});

async function novoLink(u) {
  try {
    const d = await api('POST', '/api/v1/usuarios/' + u.id + '/convite');
    mostrarLink(d, (u.tem_senha ? 'Link para ' + u.nome + ' criar uma senha nova' : 'Novo convite para ' + u.nome) + '. Vale até ' + dataHora(d.expira_em) + '; o link anterior deixou de valer.');
  } catch (e) {
    toast(e.message, true);
  }
}

async function alterar(u, papel, ativo) {
  if (!ativo && !confirm('Desativar ' + u.nome + '? A pessoa sai da plataforma na hora e não consegue mais entrar.')) {
    return;
  }
  try {
    await api('PUT', '/api/v1/usuarios/' + u.id, { papel, ativo });
    toast(!ativo ? u.nome + ' desativado.' : ativo !== u.ativo ? u.nome + ' reativado.' : 'Perfil de ' + u.nome + ' agora é ' + PAPEIS[papel] + '.');
    if (u.id === eu.id && papel !== 'admin') {
      location.replace('/');
      return;
    }
  } catch (e) {
    toast(e.message, true);
  }
  carregarUsuarios();
}

$('copiar').addEventListener('click', async () => {
  const texto = $('link-valor').textContent;
  try {
    await navigator.clipboard.writeText(texto);
    toast('Link copiado.');
  } catch (e) {
    const r = document.createRange();
    r.selectNodeContents($('link-valor'));
    getSelection().removeAllRanges();
    getSelection().addRange(r);
    toast('Selecionei o link: use Ctrl+C para copiar.');
  }
});
