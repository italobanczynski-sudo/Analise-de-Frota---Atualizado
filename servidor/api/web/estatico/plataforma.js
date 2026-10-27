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
  $('aba-whatsapp').hidden = eu.papel === 'leitor';
  $('botao-enviar').hidden = eu.papel !== 'admin';
  abrirAba(abaDoEndereco());
})();

$('sair').addEventListener('click', async () => {
  try { await api('POST', '/api/v1/auth/sair'); } catch (e) { /* sai mesmo assim */ }
  location.replace('/');
});

// ---------- abas ----------
const ABAS = ['clientes', 'whatsapp', 'usuarios'];

// a aba do endereço (#whatsapp, #usuarios), se o perfil puder vê-la
function abaDoEndereco() {
  const nome = location.hash.slice(1);
  return ABAS.includes(nome) && !$('aba-' + nome).hidden ? nome : 'clientes';
}

function abrirAba(nome) {
  ABAS.forEach((n) => {
    $('aba-' + n).setAttribute('aria-selected', String(n === nome));
    $('sec-' + n).hidden = n !== nome;
  });
  history.replaceState(null, '', nome === 'clientes' ? location.pathname : '#' + nome);
  if (nome === 'clientes') carregarClientes();
  else if (nome === 'whatsapp') iniciarWhatsApp();
  else carregarUsuarios();
}
window.addEventListener('hashchange', () => { if (eu) abrirAba(abaDoEndereco()); });
ABAS.forEach((n) => $('aba-' + n).addEventListener('click', () => abrirAba(n)));

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

// ---------- WhatsApp dos motoristas (colaborador e administrador) ----------
const SITUACAO_ENVIO = {
  enviada: ['Enviada', 'selo-ok'],
  enviando: ['Enviando…', 'selo-atencao'],
  erro: ['Com erro', 'selo-erro'],
};

function ontemISO() {
  const d = new Date(Date.now() - 3 * 3600 * 1000 - 24 * 3600 * 1000); // ontem em Brasília (UTC−3)
  return d.toISOString().slice(0, 10);
}

async function iniciarWhatsApp() {
  const sel = $('wa-cliente');
  if (!sel.options.length) {
    let clientes;
    try {
      clientes = await api('GET', '/api/v1/clientes');
    } catch (e) {
      toast(e.message, true);
      return;
    }
    sel.replaceChildren(...clientes.map((c) => el('option', { value: c.id }, c.nome)));
    $('wa-dia').value = ontemISO();
    $('wa-dia').max = ontemISO();
  }
  carregarWhatsApp();
}

function clienteWA() { return $('wa-cliente').value; }

async function carregarWhatsApp() {
  $('previas').replaceChildren();
  if (!clienteWA()) {
    $('lista-wa').replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: 'Nenhum cliente cadastrado ainda.' })));
    return;
  }
  let p;
  try {
    p = await api('GET', '/api/v1/clientes/' + clienteWA() + '/whatsapp');
  } catch (e) {
    $('lista-wa').replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: e.message })));
    return;
  }
  const aviso = $('wa-aviso');
  aviso.className = 'alerta ' + (p.configurado && p.automatico ? 'alerta-ok' : 'alerta-erro');
  aviso.textContent = !p.configurado
    ? 'Modo de teste: a conta Twilio ainda não está configurada no servidor. Nada é enviado; as mensagens aparecem só no log.'
    : p.automatico ? 'Envio automático ligado: todo dia às ' + p.hora + 'h (Brasília), com o resultado do dia anterior.'
      : 'Twilio configurada, mas o envio automático está desligado (WHATSAPP_AUTOMATICO=false).';
  aviso.hidden = false;

  motoristasWA = p.motoristas;
  $('wa-motoristas').replaceChildren(...p.motoristas.map((m) => el('option', { value: m })));
  const corpo = $('lista-wa');
  if (!p.contatos.length) {
    corpo.replaceChildren(el('tr', {}, el('td', { colspan: 5, class: 'vazio', text: 'Nenhum motorista com WhatsApp cadastrado.' })));
  } else {
    corpo.replaceChildren(...p.contatos.map((c) => el('tr', {},
      el('td', {}, el('div', { class: 'pessoa' }, el('span', { class: 'avatar', 'aria-hidden': 'true' }, iniciais(c.motorista)), el('b', { text: c.motorista }))),
      el('td', { class: 'num', text: '•••• ' + c.telefone_final }),
      el('td', { class: 'num', text: dataHora(c.autorizado_em) }),
      el('td', {}, c.ativo ? el('span', { class: 'selo selo-ok' }, 'Recebendo') : el('span', { class: 'selo selo-neutro' }, 'Pausado')),
      el('td', {}, el('div', { class: 'acoes' },
        el('button', { class: 'btn-link btn', type: 'button', onclick: () => ativarWA(c, !c.ativo) }, c.ativo ? 'Pausar' : 'Retomar'),
        el('button', { class: 'btn-link btn btn-perigo', type: 'button', onclick: () => removerWA(c) }, 'Remover'))))));
  }

  const envios = $('lista-envios');
  envios.replaceChildren(...(p.envios.length ? p.envios.map((e) => {
    const [rot, cls] = SITUACAO_ENVIO[e.situacao] || ['—', 'selo-neutro'];
    return el('tr', {},
      el('td', { class: 'num', text: diaBR(e.dia) }),
      el('td', { text: e.motorista }),
      el('td', {}, el('span', { class: 'selo ' + cls, title: e.erro || '' }, rot)),
      el('td', { class: 'num', text: dataHora(e.criado_em) }));
  }) : [el('tr', {}, el('td', { colspan: 4, class: 'vazio', text: 'Nenhuma mensagem enviada nos últimos 14 dias.' }))]));
}

$('wa-cliente').addEventListener('change', () => { fecharImportacao(); carregarWhatsApp(); });

$('form-wa').addEventListener('submit', async (ev) => {
  ev.preventDefault();
  const motorista = $('wa-motorista').value.trim();
  const telefone = $('wa-telefone').value.trim();
  const autorizado = $('wa-autorizado').checked;
  const erro = $('erro-wa');
  if (!motorista || !telefone) {
    mostrarErro(erro, 'Informe o motorista e o WhatsApp.');
    return;
  }
  if (!autorizado) {
    mostrarErro(erro, 'Só cadastre o motorista que autorizou receber as mensagens, e marque a confirmação.');
    return;
  }
  mostrarErro(erro, '');
  $('botao-wa').disabled = true;
  try {
    await api('PUT', '/api/v1/clientes/' + clienteWA() + '/whatsapp/contatos', { motorista, telefone, autorizado });
    toast('WhatsApp de ' + motorista + ' salvo.');
    $('form-wa').reset();
    carregarWhatsApp();
  } catch (e) {
    mostrarErro(erro, e.message);
  }
  $('botao-wa').disabled = false;
});

// ---------- importar planilha de WhatsApp (lida só no navegador; cada linha vai pela mesma rota do cadastro) ----------
let importacao = [];
let motoristasWA = [];

// normaliza o nome da coluna: "WhatsApp " e "whatsapp" valem igual, sem acento
function semAcento(t) { return String(t || '').normalize('NFD').replace(/[̀-ͯ]/g, '').trim().toLowerCase(); }
const COL_MOTORISTA = ['motorista', 'nome', 'condutor', 'nome do motorista'];
const COL_TELEFONE = ['whatsapp', 'telefone', 'celular', 'fone', 'numero', 'whatsapp do motorista'];

// mesma regra do servidor (modules/contato): só confere, quem decide é o servidor
function telefoneValido(v) {
  let d = String(v || '').replace(/\D/g, '');
  if (d.startsWith('0') && (d.length === 11 || d.length === 12)) d = d.slice(1);
  if (d.length === 10 || d.length === 11) d = '55' + d;
  if (d.length === 12 && d[4] >= '6') d = d.slice(0, 4) + '9' + d.slice(4); // celular antigo, sem o 9
  return /^55[1-9][1-9](9\d{8}|[2-5]\d{7})$/.test(d) ? d : '';
}

function carregarXLSX() {
  if (window.XLSX) return Promise.resolve();
  return new Promise((ok, falha) => {
    const s = document.createElement('script');
    s.src = '/painel/xlsx.full.min.js';
    s.onload = ok;
    s.onerror = () => falha(new Error('Não consegui carregar o leitor de Excel. Salve a planilha como .csv e tente de novo.'));
    document.head.appendChild(s);
  });
}

async function lerPlanilha(arquivo) {
  await carregarXLSX();
  const dados = await arquivo.arrayBuffer();
  // raw: o número da célula como está (formatado, 5511912345678 pode virar "5,5E+12")
  const wb = XLSX.read(dados, { type: 'array', raw: true });
  const linhas = XLSX.utils.sheet_to_json(wb.Sheets[wb.SheetNames[0]], { defval: '', raw: true });
  if (!linhas.length) throw new Error('A planilha está vazia.');
  const colunas = Object.keys(linhas[0]);
  const achar = (nomes) => colunas.find((c) => nomes.includes(semAcento(c)));
  const cm = achar(COL_MOTORISTA);
  const ct = achar(COL_TELEFONE);
  if (!cm || !ct) throw new Error('Não achei as colunas. A planilha precisa de uma coluna Motorista e uma coluna WhatsApp (ou Telefone/Celular).');
  const conhecidos = new Map(motoristasWA.map((m) => [semAcento(m).replace(/\s+/g, ' '), m]));
  return linhas.map((l) => {
    const bruto = String(l[cm] || '').trim();
    const nome = conhecidos.get(semAcento(bruto).replace(/\s+/g, ' ')) || bruto;
    const tel = telefoneValido(l[ct]);
    let problema = '';
    if (!bruto) problema = 'Sem nome';
    else if (!tel) problema = 'WhatsApp inválido';
    return { motorista: nome, telefone: String(l[ct] || '').trim(), final: tel ? tel.slice(-4) : '', problema, conhecido: conhecidos.has(semAcento(bruto).replace(/\s+/g, ' ')) };
  }).filter((r) => r.motorista || r.telefone);
}

function mostrarImportacao() {
  const validos = importacao.filter((r) => !r.problema).length;
  $('lista-importacao').replaceChildren(...importacao.map((r) => el('tr', {},
    el('td', { text: r.motorista || '—' }),
    el('td', { class: 'num', text: r.final ? '•••• ' + r.final : '—' }),
    el('td', {}, r.problema ? el('span', { class: 'selo selo-erro' }, r.problema)
      : r.conhecido ? el('span', { class: 'selo selo-ok' }, 'Ok')
        : el('span', { class: 'selo selo-atencao', title: 'Este nome não aparece nos dados da API dos últimos 30 dias: confira se está escrito igual' }, 'Nome não encontrado na API')))));
  $('botao-importar').textContent = 'Salvar ' + validos + ' válido(s)';
  $('botao-importar').disabled = validos === 0;
  $('wa-autorizados').checked = false;
  $('wa-importacao').hidden = false;
}

function fecharImportacao() {
  importacao = [];
  $('lista-importacao').replaceChildren();
  $('wa-importacao').hidden = true;
  $('wa-arquivo').value = '';
}

$('wa-arquivo').addEventListener('change', async (ev) => {
  const arquivo = ev.target.files[0];
  if (!arquivo) return;
  try {
    importacao = await lerPlanilha(arquivo);
    mostrarImportacao();
  } catch (e) {
    toast(e.message, true);
    fecharImportacao();
  }
});

$('botao-cancelar-importacao').addEventListener('click', fecharImportacao);

$('botao-importar').addEventListener('click', async () => {
  if (!$('wa-autorizados').checked) {
    toast('Confirme que os motoristas autorizaram receber as mensagens.', true);
    return;
  }
  const validos = importacao.filter((r) => !r.problema);
  $('botao-importar').disabled = true;
  let salvos = 0;
  const erros = [];
  for (const r of validos) {
    try {
      await api('PUT', '/api/v1/clientes/' + clienteWA() + '/whatsapp/contatos', { motorista: r.motorista, telefone: r.telefone, autorizado: true });
      salvos++;
    } catch (e) {
      erros.push(r.motorista + ': ' + e.message);
    }
  }
  fecharImportacao();
  toast(salvos + ' WhatsApp(s) salvo(s).' + (erros.length ? ' Com erro: ' + erros.join('; ') : ''), erros.length > 0);
  carregarWhatsApp();
});

async function ativarWA(c, ativo) {
  try {
    await api('PUT', '/api/v1/clientes/' + clienteWA() + '/whatsapp/contatos/ativo', { motorista: c.motorista, ativo });
    toast(ativo ? c.motorista + ' volta a receber.' : c.motorista + ' não recebe mais até ser retomado.');
  } catch (e) {
    toast(e.message, true);
  }
  carregarWhatsApp();
}

async function removerWA(c) {
  if (!confirm('Remover o WhatsApp de ' + c.motorista + '? O número é apagado e o motorista deixa de receber.')) return;
  try {
    await api('DELETE', '/api/v1/clientes/' + clienteWA() + '/whatsapp/contatos?motorista=' + encodeURIComponent(c.motorista));
    toast('WhatsApp de ' + c.motorista + ' removido.');
  } catch (e) {
    toast(e.message, true);
  }
  carregarWhatsApp();
}

$('botao-previa').addEventListener('click', async () => {
  const caixa = $('previas');
  let previas;
  try {
    previas = await api('GET', '/api/v1/clientes/' + clienteWA() + '/whatsapp/previa?dia=' + encodeURIComponent($('wa-dia').value));
  } catch (e) {
    toast(e.message, true);
    return;
  }
  if (!previas.length) {
    caixa.replaceChildren(el('p', { class: 'vazio', text: 'Nenhum motorista vinculado rodou neste dia.' }));
    return;
  }
  caixa.replaceChildren(...previas.map((p) => el('div', { class: 'previa' },
    el('div', { class: 'previa-cab' }, el('b', { text: p.motorista }),
      p.tem_whatsapp ? el('span', { class: 'selo selo-ok' }, 'Recebe') : el('span', { class: 'selo selo-neutro' }, 'Sem WhatsApp')),
    el('p', { class: 'previa-texto', text: p.texto }))));
});

$('botao-enviar').addEventListener('click', async () => {
  const dia = $('wa-dia').value;
  if (!confirm('Enviar agora o resultado de ' + diaBR(dia) + ' para os motoristas com WhatsApp? Quem já recebeu este dia não recebe de novo.')) return;
  $('botao-enviar').disabled = true;
  try {
    const r = await api('POST', '/api/v1/clientes/' + clienteWA() + '/whatsapp/enviar', { dia });
    const partes = [r.enviadas + ' enviada(s)'];
    if (r.ja_enviadas) partes.push(r.ja_enviadas + ' já tinham recebido');
    if (r.sem_whatsapp) partes.push(r.sem_whatsapp + ' sem WhatsApp');
    if (r.erros.length) partes.push(r.erros.length + ' com erro');
    toast((r.teste ? 'Modo de teste (nada saiu de verdade): ' : '') + partes.join(', ') + '.', r.erros.length > 0);
  } catch (e) {
    toast(e.message, true);
  }
  $('botao-enviar').disabled = false;
  carregarWhatsApp();
});

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
