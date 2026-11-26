/* =====================================================================
   APP · shell, rotas, ações
   ===================================================================== */
const ROUTES={
  geral:{t:'Visão geral',i:'home',f:vOverview,g:'Análise'},
  retencao:{t:'Saúde do cliente',i:'shield',f:vRetention},
  plano:{t:'Plano de ação',i:'flag',f:vPlan},
  estudo:{t:'Estudo do projeto',i:'coin',f:vStudy},
  frota:{t:'Frota',i:'truck',f:vFleet,g:'Operação'},
  motoristas:{t:'Motoristas',i:'users',f:vDrivers},
  dados:{t:'Regras e integração',i:'sliders',f:vSettings,g:'Sistema'}
};
const LEGACY={churn:'retencao',visao:'frota','individual/veiculo':'frota','individual/motorista':'motoristas',premiacao:'motoristas?aba=premiacao',mensagens:'motoristas?aba=mensagens',api:'dados?aba=integracao',config:'dados',individual:'motoristas'};
const TABKEY={retencao:['retTab',{cliente:1,carteira:1}],frota:['fleetTab',{visao:1,indicadores:1,qualidade:1,mapa:1,veiculos:'visao'}],motoristas:['drvTab',{resumo:1,comportamento:1,lista:1,comparativo:1,semanal:1,mensagens:1,premiacao:1}],dados:['dataTab',{regras:1,integracao:1,diagnostico:1,notaconsumo:1}]};
let LASTC=null,RTOK=0,CUR_ROUTE='geral';

function parseHash(){
  let h=location.hash.replace(/^#\/?/,'');let [path,qs]=h.split('?');path=path||'geral';
  if(LEGACY[path]){const m=LEGACY[path].split('?');path=m[0];if(m[1])qs=m[1]}
  if(!ROUTES[path])path='geral';
  const q=Object.fromEntries(new URLSearchParams(qs||''));return{path,q};
}
function nav(route){const t='#/'+route;if(location.hash===t)render();else location.hash=t}

async function buildCtx(){
  const cid=S.client;const fleet=await DB.load(cid);const P=resolvePeriod(S);
  const cur=metrics(cid,ST,P.from,P.to),prev=metrics(cid,ST,P.cf,P.ct);
  const def=CLIENT_DEFS.find(c=>c.id===cid);
  const dsl=dieselInfo(def);
  const C={cid,def,fleet,P,cur,prev,price:dsl.preco,dsl};
  /* um só número de economia (vs. linha de base do projeto) e uma só contagem de situação dos equipamentos, para todas as telas */
  C.st=study(cid,ST,C.price);C.ec=econOf(cur,C.price);C.gl=gainsLosses(cur,prev,C.price);C.conc=concOf(C.gl,ST.rules);C.conf=confOf(cid,cur,ST.rules);
  C.hl=health(cur,prev,ST.rules,C.st.noBase?null:{econ:C.st.econ,inv:C.st.inv});C.eq=eqSit(C);LASTC=C;loadExtras(cid);return C;
}
/* ---------- Situação dos equipamentos (item 4): critérios com nome, os mesmos em todas as telas ----------
   sem transmitir = nenhum dia de dado nos últimos 3 dias (dia a dia da API)
   sem sinal agora = a posição ao vivo não veio nesta consulta (pode ser momentâneo)
   sem base de comparação = rodou no período, mas não no período de comparação
   sem motorista identificado = rodou no período sem vínculo de motorista na API */
function eqSit(C){
  const R=ST.rules,ct=contratoDe(C.def);
  const semTransmitir=C.cur.noSignal;const longos=semTransmitir.filter(x=>x.days>=R.diasCobranca);
  const semSinalAgora=C.fleet.vehicles.filter(v=>vehStat(C,v)==='off').map(v=>v.id);
  const semBase=Object.keys(C.cur.byVeh).filter(id=>!(C.prev.byVeh[id]&&C.prev.byVeh[id].kmpl)&&C.fleet.byId[id]);
  const semMotorista=Object.keys(C.cur.byVeh).filter(id=>C.fleet.byId[id]&&C.fleet.byId[id].noDriver);
  /* desde quando: início da sequência mais recente de dias rodados sem motorista (só com os registros do servidor) */
  const desde={};if(REAL.on)for(const id of semMotorista){const rs=C.fleet.recs.filter(r=>r.veh===id&&r.km>0).sort((a,b)=>b.t-a.t);let t=null;for(const r of rs){if(r.drv)break;t=r.t}desde[id]=t}
  return{semTransmitir,longos,semSinalAgora,semBase,semMotorista,desde,valorUnit:ct.unitAuto?null:ct.unit,cobrado:ct.unitAuto?null:longos.length*ct.unit};
}
function eqInfo(){return infoDot('Critérios da situação dos equipamentos',`<ul><li><b>Sem transmitir</b>: nenhum dia de dado nos últimos 3 dias (dia a dia da API).</li><li><b>Sem sinal agora</b>: a posição ao vivo não veio nesta consulta; pode ser só momentâneo.</li><li><b>Sem base de comparação</b>: rodou no período, mas não no período de comparação.</li><li><b>Sem motorista identificado</b>: rodou sem vínculo de motorista na API.</li></ul><p>As mesmas contagens aparecem na Visão geral, na Frota, no mapa e nos avisos.</p>`,{left:true})}
function eqLine(C){const e=C.eq;return[['bad',e.semTransmitir.length,'sem transmitir'],['warn',e.semSinalAgora==null?null:e.semSinalAgora.length,'sem sinal agora'],['n',e.semBase.length,'sem base de comparação'],['n',e.semMotorista.length,'sem motorista identificado']].map(([c,n,l])=>`<span class="badge ${n?c:'n'}">${n==null?'sem dado':n} ${l}</span>`).join('')}
/* cabeçalho do PDF: cliente, período, diesel usado e a base da economia (aparece só na impressão) */
function printHead(C){return`<div class="print-head"><b>${esc(C.def.name)} · ${esc(ROUTES[CUR_ROUTE]?ROUTES[CUR_ROUTE].t:'')} · ${C.P.label}</b><br>Diesel: <b>${fmt.brl(C.price,2)}/L</b> (${esc(C.dsl.fonte)}) · Economia ${esc(C.ec.b.rotulo)} · gerado em ${new Date().toLocaleString('pt-BR')}</div>`;}
const skeleton=()=>`<div class="page-h"><div><h1 style="opacity:.25">Carregando</h1></div></div><div class="grid g4">${'<div class="skel"></div>'.repeat(4)}</div><div class="grid g-main sec"><div class="skel" style="min-height:300px"></div><div class="skel" style="min-height:300px"></div></div>`;

async function render(opt={}){
  const tok=++RTOK;const anim=opt.anim!==false;const {path}=parseHash();CUR_ROUTE=path;
  const view=$('#view');const y=opt.keepScroll?window.scrollY:0;
  const ae=document.activeElement;const fsel=ae&&ae.dataset&&ae.dataset.chg&&ae.tagName!=='SELECT'?`[data-chg="${ae.dataset.chg}"]`:null;const caret=fsel?ae.selectionStart:null;
  if(opt.skeleton){view.innerHTML=skeleton();view.classList.remove('noanim')}
  let html;
  if(!CLIENT_DEFS.length){html=`<div class="card" style="margin-top:24px"><div class="empty"><b>Nenhum cliente com dados ainda</b>${REAL.on?'Cadastre o cliente e o token em <a href="/plataforma">Clientes e coleta</a>. A busca na API Gobrax roda de hora em hora.':''}</div></div>`}
  else try{const C=await buildCtx();if(tok!==RTOK)return;html=printHead(C)+await ROUTES[path].f(C)}
  catch(e){console.error(e);html=`<div class="card" style="margin-top:24px"><h2>Algo deu errado ao montar esta tela</h2><p class="muted">${esc(e.message)}</p></div>`}
  if(tok!==RTOK)return;
  view.classList.toggle('noanim',!anim);view.innerHTML=html;enhance(view,anim);
  $$('.nav a').forEach(a=>a.classList.toggle('on',a.dataset.r===path));placeNav();renderTop();
  document.title=ROUTES[path].t+' · '+(LASTC?LASTC.def.name:'')+' · Análise de frota Gobrax';
  window.scrollTo(0,y);
  if(fsel){const el=$(fsel,view);if(el){el.focus();try{if(caret!=null)el.setSelectionRange(caret,caret)}catch(e){}}}
}
const go=(anim=false)=>render({anim,keepScroll:true});
/* redesenha sem atrapalhar: não mexe se houver painel lateral aberto ou alguém digitando */
function softRender(){if($('#ov').innerHTML||/INPUT|TEXTAREA|SELECT/.test((document.activeElement||{}).tagName||''))return;go()}

/* ---------- dados extras do servidor (posição, mensagens enviadas, telefones) ---------- */
function loadExtras(cid,force){
  if(!REAL.on)return;const f=DB.fleets[cid];if(!f)return;REAL.liveErr=REAL.liveErr||{};
  const stale=!REAL.liveAt[cid]||Date.now()-REAL.liveAt[cid]>120000;
  if((force||stale)&&!REAL.busy['live'+cid]&&f.vehicles.length){
    const first=!REAL.live[cid];REAL.busy['live'+cid]=1;
    Gobrax.api.getLive(cid).then(d=>{REAL.live[cid]=d;REAL.liveAt[cid]=Date.now();REAL.liveErr[cid]=null})
      .catch(e=>{REAL.liveErr[cid]=e.message;REAL.liveAt[cid]=Date.now()})
      .finally(()=>{REAL.busy['live'+cid]=0;if(S.client!==cid||!LASTC)return;
        if(first&&['geral','frota'].includes(CUR_ROUTE))softRender();else updateLiveMarkers()});
  }
  if(!REAL.msgs[cid]&&!REAL.busy['msg'+cid]){REAL.busy['msg'+cid]=1;
    Gobrax.api.getSent(cid).then(d=>{REAL.msgs[cid]=d||[]}).catch(()=>{REAL.msgs[cid]=[]})
      .finally(()=>{REAL.busy['msg'+cid]=0;if(S.client===cid&&CUR_ROUTE==='motoristas'&&S.u.drvTab==='mensagens')softRender()})}
  if(!REAL.phones[cid]&&!REAL.busy['ph'+cid]){REAL.busy['ph'+cid]=1;
    Gobrax.api.getPhones(cid).then(list=>{(list||[]).forEach(c=>{if(!c.telefone)return;const p=f.drvPlate[nkey(c.motorista)];if(p&&f.byId[p]&&!f.byId[p].phone)f.byId[p].phone=c.telefone})})
      .catch(()=>{}).finally(()=>{REAL.phones[cid]=true;REAL.busy['ph'+cid]=0})}
}
/* posição nova sem redesenhar a tela (mantém o zoom do mapa) */
function updateLiveMarkers(){
  const C=LASTC;if(!C)return;
  MAPS.forEach(inst=>{if(!inst.markers)return;let changed=false;
    Object.values(inst.markers).forEach(o=>{const g=vehGeo(C,o.v);if(!g)return;o.geo=g;o.m.setLatLng([g.lat,g.lng]);o.m.setIcon(pin(g.st,o.v.id.slice(-2),inst.mode==='mini'));changed=true});
    if(changed)applyMapFilter(inst)});
}
const waUrl=(ph,txt)=>{let d=String(ph||'').replace(/\D/g,'');if(d.length===10||d.length===11)d='55'+d;return d.length>=12?'https://wa.me/'+d+'?text='+encodeURIComponent(txt):null};

/* ---------- shell ---------- */
function shell(){
  const nav=Object.entries(ROUTES).map(([k,r])=>`${r.g?`<div class="nav-g">${r.g}</div>`:''}<a href="#/${k}" data-r="${k}">${ico(r.i,20)}<span class="lb">${r.t}</span>${k==='frota'?'<span class="cnt" id="cnt-frota" hidden></span>':''}</a>`).join('');
  const eu=REAL.eu||{nome:'Visitante'};
  const badge=REAL.on?`<span class="demo live" title="Dados buscados pelo servidor na API Gobrax">API GOBRAX</span>`:`<span class="demo" title="Fora da plataforma: dados gerados pelo adaptador de simulação">DADOS SIMULADOS</span>`;
  return`<div class="app" id="appRoot"><aside class="sidebar" id="sidebar"><div class="logo" aria-label="gobrax"><span class="g">go</span><span class="l">b</span><span class="l">r</span><span class="l">a</span><span class="l">x</span></div><nav class="nav" aria-label="Principal"><i class="nav-ind"></i>${nav}</nav><div class="side-foot"><button class="ctl icon" data-act="collapse" aria-label="Recolher menu" style="background:transparent;border-color:rgba(255,255,253,.2);color:#fff">${ico('menu',18)}</button></div></aside>
  <div class="main"><header class="topbar" id="topbar"><button class="ctl icon menu-btn" data-act="menu" aria-label="Abrir menu">${ico('menu',18)}</button>
    <div class="search">${ico('search',18)}<input id="gs" type="search" placeholder="Buscar placa ou condutor  ( / )" autocomplete="off" aria-label="Buscar placa ou condutor"><div class="sres" id="sres"></div></div>
    <div style="flex:1"></div>${badge}
    <div class="relative"><select class="ctl" id="clientSel" aria-label="Cliente"></select></div>
    <div class="relative"><button class="ctl" data-act="pop" data-p="period" id="periodBtn"></button><div class="pop" id="pop-period"></div></div>
    <div class="relative"><button class="ctl diesel" data-act="pop" data-p="diesel" id="dieselBtn" title="Preço do diesel usado nos cálculos"></button><div class="pop" id="pop-diesel" style="min-width:300px;max-height:70vh;overflow:auto"></div></div>
    <button class="ctl icon" data-act="theme" aria-label="Alternar tema" id="themeBtn"></button>
    <div class="relative"><button class="ctl icon" data-act="pop" data-p="bell" aria-label="Alertas" style="position:relative">${ico('bell',18)}<span class="dot" id="bellDot" hidden></span></button><div class="pop" id="pop-bell" style="min-width:400px;max-width:92vw;max-height:78vh;overflow:auto"></div></div>
    <div class="relative"><button class="avatar" data-act="pop" data-p="acct" title="${esc(eu.nome)}" aria-label="Sua conta: ${esc(eu.nome)}">${esc(initials(eu.nome))}</button><div class="pop acct" id="pop-acct"></div></div></header>
    <main class="view" id="view" tabindex="-1"></main></div></div><div id="ov"></div><div class="toasts" id="toasts" aria-live="polite"></div><div class="tt" id="tt" role="tooltip"></div>`;
}
function acctPop(){
  const eu=REAL.eu;const papel={admin:'Administrador',colaborador:'CS',leitor:'Leitura'};
  if(!eu)return`<div class="who"><b>Modo demonstração</b><span class="xs muted">Abra pelo servidor (/painel) para ver os dados reais.</span></div>`;
  return`<div class="who"><b>${esc(eu.nome)}</b><span class="xs muted">${esc(papel[eu.papel]||eu.papel||'')}</span></div><a href="/plataforma">${ico('users',16)} Clientes e coleta${eu.papel==='admin'?' · usuários':''}</a><a href="#/dados?aba=diagnostico" data-act="pop-close">${ico('alert',16)} Diagnóstico do dado</a><a href="/painel/classico">${ico('file',16)} Painel clássico</a><button data-act="logout">${ico('out',16)} Sair</button>`;
}
function renderTop(){
  const C=LASTC;
  $('#clientSel').innerHTML=CLIENT_DEFS.map(c=>`<option value="${esc(c.id)}" ${c.id===S.client?'selected':''}>${esc(c.name)}</option>`).join('');
  $('#dieselBtn').innerHTML=`${ico('coin',16)} Diesel ${fmt.brl(C?C.price:ST.rules.diesel,2)}/L`;if(C)$('#dieselBtn').title='Diesel usado nos cálculos: '+C.dsl.fonte;
  $('#themeBtn').innerHTML=ico(S.theme==='dark'?'sun':'moon',18);
  if(!C)return;
  $('#periodBtn').innerHTML=`${ico('refresh',16)} ${C.P.label}`;
  const alerts=buildAlerts(C);const d=$('#bellDot');d.hidden=!alerts.length||ST.notifRead;d.textContent=alerts.length;
  const fc=$('#cnt-frota');const n=suspCount(C)+C.eq.semTransmitir.length;fc.hidden=!n;fc.textContent=n;
  if($('#pop-period').classList.contains('open'))$('#pop-period').innerHTML=periodPop();
}
function buildAlerts(C){
  const a=[];const hl=C.hl;
  /* Onde agir primeiro: ordenado pelo dinheiro em jogo no período */
  try{const op=opportunities(C.cur,C.prev,C.price);
    if(op.ref)a.push({act:1,ic:'fuel',t:'Consumo acima da referência',v:fmt.brl(op.potential),s:op.worse.length?`${op.worse.length} de ${op.evaluated} veículos abaixo de ${fmt.n(op.ref,2)} km/l. Comece por ${op.worse.slice(0,3).map(w=>w.id).join(', ')} com o instrutor.`:'Nenhum veículo abaixo da referência.',f:'down'});
    a.push({act:1,ic:'trend',t:'Condutores que pioraram',v:String(op.fell.length),s:op.fell.length?`${op.fell[0].id} caiu de ${fmt.n(op.fell[0].from,2)} para ${fmt.n(op.fell[0].to,2)} km/l. Converse para entender rota, veículo ou carga.`:'Ninguém piorou mais que 2%.',f:'down'});
    if(op.kdRef)a.push({act:1,ic:'truck',t:'Pouco rendimento por dia',v:fmt.n(C.cur.kmDay,0)+' km',s:op.low.length?`${op.low.length} veículos rodam menos de 80% da referência (${fmt.n(op.kdRef,0)} km). Revise escala e tempo de espera.`:'Rendimento dentro da referência.',f:'all'})}catch(e){}
  if(hl.band!=='BAIXO')a.push({ic:'shield',t:`Risco ${hl.band.toLowerCase()} de cancelamento (${hl.idx}/100)`,s:hl.top?'Principal sinal: '+hl.top.label.toLowerCase():'',r:'retencao'});
  if(cmpShort(C))a.push({ic:'alert',t:'Período de comparação com poucos dados',s:`Só ${C.prev.dataDays} dias com dado (de ${C.P.clen}): as variações de km, dias e diesel ficam infladas`,r:'frota?aba=qualidade'});
  const sc=suspCount(C);if(sc)a.push({ic:'alert',t:`${sc} registros suspeitos fora das médias`,s:'Veja quais registros ficaram fora das médias',r:'frota?aba=qualidade'});
  const e=C.eq;
  if(e.longos.length)a.push({ic:'alert',t:`${e.longos.length} equipamento${e.longos.length>1?'s':''} sem transmitir há mais de ${ST.rules.diasCobranca} dias: risco de cobrança indevida`,s:e.longos.slice(0,4).map(x=>x.id+' ('+x.days+' dias)').join(', ')+' · '+(e.cobrado!=null?fmt.brl(e.cobrado,2)+'/mês cobrados':'valor cobrado: sem dado (contrato não cadastrado)'),r:'frota?aba=mapa'});
  if(e.semTransmitir.length)a.push({ic:'truck',t:`${e.semTransmitir.length} veículo${e.semTransmitir.length>1?'s':''} sem transmitir (nenhum dado nos últimos 3 dias)`,s:e.semTransmitir.slice(0,3).map(x=>x.id).join(', '),r:'frota?aba=mapa'});
  const st=C.st;if(!st.noBase&&!st.pays)a.push({ic:'coin',t:'O projeto ainda não se paga',s:`Economia acumulada ${fmt.brl(st.econ)} contra ${fmt.brl(st.inv)} investidos (cobertura ${fmt.n(st.econ/st.inv,2)}×, vs. linha de base do projeto). Faltam ${fmt.brl(st.miss)}.`,r:'estudo'});
  if(C.conc.alerta)a.push({ic:'user',t:`${C.conc.top.id} responde por ${fmt.n(C.conc.share,0)}% da perda de diesel`,s:`Saldo da frota ${fmt.brl(C.conc.saldo)}; sem esse veículo, ${fmt.brl(C.conc.semTop)} (vs. linha de base do projeto)`,r:'geral'});
  return a;
}
function bellPop(C){
  if(!C)return'<div class="empty"><b>Sem dados</b></div>';
  const a=buildAlerts(C);if(!a.length)return'<div class="empty"><b>Tudo em dia</b></div>';
  const acts=a.filter(x=>x.act),oth=a.filter(x=>!x.act);
  const li=x=>`<a class="li" href="#/${x.r}" data-act="pop-close" style="text-decoration:none;color:inherit"><div class="ava">${ico(x.ic,16)}</div><div><b class="sm">${x.t}</b><div class="xs muted">${esc(x.s)}</div></div></a>`;
  const ac=x=>`<div class="nact"><div class="row" style="justify-content:space-between;flex-wrap:nowrap;gap:10px"><b class="sm">${x.t}</b><span class="num b nv">${esc(x.v)}</span></div><div class="xs">${esc(x.s)}</div><div class="row" style="gap:6px"><button class="btn primary sm" data-act="go-drv" data-f="${x.f}">Ver condutores</button></div></div>`;
  return`<div class="stack" style="gap:6px"><div class="tag">Onde agir primeiro</div><div class="xs muted" style="margin-top:-4px">Ordenado pelo dinheiro em jogo no período</div>${acts.map(ac).join('')}${oth.length?`<div class="tag" style="margin-top:8px">Avisos</div>${oth.map(li).join('')}`:''}</div>`;
}
function dieselPop(){
  const C=LASTC;if(!C)return'';const d=C.dsl,pode=podeEditarContrato();
  const op=(uf,nome)=>{const v=DIESEL.uf?DIESEL.uf[uf]:null;return`<button class="li" ${pode?`data-act="diesel-uf" data-v="${uf}"`:'disabled'} style="width:100%;border:0;background:${d.uf===uf&&d.api?'var(--a200)':'transparent'};text-align:left;cursor:${pode?'pointer':'default'}"><span style="flex:1">${esc(nome)}</span><b class="num">${v?fmt.brl(v,2):'sem dado'}</b></button>`};
  return`<div class="stack" style="gap:4px"><div class="tag">Preço do diesel</div><div class="xs muted">Em uso: <b>${fmt.brl(d.preco,2)}/L</b> · ${esc(d.fonte)}</div>
  ${DIESEL.uf?'':'<div class="xs" style="color:var(--bad-t)">A API de preço (Petrobras, combustivelapi.com.br) não respondeu agora.</div>'}
  ${pode?'<div class="xs muted">Escolha a média Brasil ou o estado do cliente. Vale para todas as telas, o PDF e o CSV deste cliente.</div>':'<div class="xs muted">Só CS e administrador mudam o estado do diesel.</div>'}
  ${op('br','Média Brasil (geral)')}<div class="tag" style="margin-top:6px">Por estado</div>${UFS.map(([uf,n])=>op(uf,uf.toUpperCase()+' · '+n)).join('')}</div>`;
}
function periodPop(){
  const p=S.period;const pre=[['month','Mês atual'],['7d','7 dias'],['30d','30 dias'],['90d','90 dias']];
  const from=p.from||dIso(T0-29),to=p.to||TODAY;
  return`<div class="stack"><div class="tag">Período</div><div class="chips">${pre.map(([k,l])=>`<button class="chip ${p.preset===k?'on':''}" data-act="period" data-v="${k}">${l}</button>`).join('')}<button class="chip ${p.preset==='custom'?'on':''}" data-act="period" data-v="custom">Personalizado</button></div>
  ${p.preset==='custom'?`<div class="grid g2" style="gap:8px"><div class="field"><label>De</label><input class="inp" type="date" data-chg="pfrom" value="${from}" max="${TODAY}"></div><div class="field"><label>Até</label><input class="inp" type="date" data-chg="pto" value="${to}" max="${TODAY}"></div></div>`:''}
  <div class="tag" style="margin-top:6px">Comparar com</div>${segHtml([['mes','Mês anterior'],['ant','Período anterior']],S.compare,'compare')}</div>`;
}

/* ---------- utilidades de ação ---------- */
function download(name,text,type){const b=new Blob([text],{type});const a=document.createElement('a');a.href=URL.createObjectURL(b);a.download=name;document.body.appendChild(a);a.click();setTimeout(()=>{URL.revokeObjectURL(a.href);a.remove()},500)}
async function copyText(t){try{await navigator.clipboard.writeText(t);return true}catch(e){const ta=document.createElement('textarea');ta.value=t;ta.style.position='fixed';ta.style.opacity='0';document.body.appendChild(ta);ta.select();let ok=false;try{ok=document.execCommand('copy')}catch(_){}ta.remove();return ok}}
const csvQ=v=>'"'+String(v??'').replace(/"/g,'""')+'"';
function csvFor(C,what){
  const gl=Object.fromEntries(gainsLosses(C.cur,C.prev,C.price).map(g=>[g.id,g]));const n=(v,d)=>v==null||!isFinite(v)?'':fmt.n(v,d).replace('−','-');
  const head=['Placa','Condutor','Km','Litros','Km/l','Freadas /100 km','Faixa verde %','Motor ocioso %','Nota'];if(what==='drivers')head.push('Situação','Diesel economizado vs. linha de base (R$)');
  const rows=C.fleet.vehicles.map(v=>{const c=C.cur.byVeh[v.id];if(!c)return null;const r=[v.id,v.driver,n(c.km,0),n(c.l,0),n(c.kmpl,2),n(c.br100,1),n(c.green,1),n(c.idle,1),n(c.score,0)];if(what==='drivers'){const g=gl[v.id];r.push(vehStatus(C.cur,C.prev,v.id).label,g?n(g.money,0):'')}return r}).filter(Boolean);
  const pre=[['Cliente',C.def.name],['Período',fmt.dmy(dIso(C.P.from))+' a '+fmt.dmy(dIso(C.P.to))],['Diesel (R$/L)',fmt.n(C.price,2)+' ('+C.dsl.fonte+')'],['Economia',C.ec.b.rotulo],[]];
  return'﻿'+[...pre,head,...rows].map(r=>r.map(csvQ).join(';')).join('\n');
}
function refreshDetail(){const C=LASTC;const el=$('#drv-detail');if(!el||!S.u.drvSel)return;el.innerHTML=drvDetail(C,S.u.drvSel);enhance(el,true);$$('.md .li').forEach(l=>l.classList.toggle('on',l.dataset.id===S.u.drvSel))}
function closePops(){$$('.pop.open').forEach(p=>p.classList.remove('open'))}
function weekWa(el,open){const C=LASTC,wk=weeklySeries(C.cid,ST,C.P.from,C.P.to);const w=wk.find(x=>x.n===+el.dataset.w);if(!w)return;const pw=wk.find(x=>x.n===w.n-1);const id=el.dataset.id;const a=w.a.byVeh[id];if(!a)return;const pa=pw?pw.a.byVeh[id]:null;
  const good=[...Array(w.to-w.from+1).keys()].filter(k=>{const r=aggFor(C,w.from+k,w.from+k,id);return r.n&&r.kmpl>=w.a.kmpl}).length;
  const v=C.fleet.byId[id];const txt=weekMsg(v.driver,{a,ch:pa&&pa.kmpl?(a.kmpl-pa.kmpl)/pa.kmpl*100:null,good},w,v);const url=waUrl(v.phone,txt);
  if(open&&url){window.open(url,'_blank','noopener');toast('Abrindo o WhatsApp')}else copyText(txt).then(ok=>toast(ok?(open?'Sem telefone cadastrado. Mensagem copiada':'Mensagem copiada'):'Não foi possível copiar'))}

/* ---------- ações (data-act) ---------- */
const ACT={
  'ov-close':()=>closeOv(),
  'pop-close':()=>closePops(),
  pop:el=>{const id='pop-'+el.dataset.p;const p=$('#'+id);const was=p.classList.contains('open');closePops();if(!was){if(el.dataset.p==='period')p.innerHTML=periodPop();else if(el.dataset.p==='diesel')p.innerHTML=dieselPop();else if(el.dataset.p==='acct')p.innerHTML=acctPop();else if(el.dataset.p==='bell'){p.innerHTML=bellPop(LASTC);ST.notifRead=true;$('#bellDot').hidden=true}p.classList.add('open')}},
  logout:async()=>{try{await srv('/api/v1/auth/sair','POST')}catch(e){}location.replace('/')},
  theme:()=>{S.theme=S.theme==='dark'?'light':'dark';document.documentElement.dataset.theme=S.theme;savePrefs();renderTop();mapsTheme()},
  menu:()=>$('#sidebar').classList.toggle('open'),
  collapse:()=>{S.collapsed=!S.collapsed;$('#sidebar').classList.remove('peek');$('#appRoot').classList.toggle('collapsed',S.collapsed);savePrefs();setTimeout(placeNav,420)},
  period:el=>{S.period.preset=el.dataset.v;if(el.dataset.v==='custom'&&!S.period.from){S.period.from=dIso(T0-29);S.period.to=TODAY}render({anim:false,keepScroll:true})},
  compare:el=>{S.compare=el.dataset.v;render({anim:false,keepScroll:true})},
  'go-prem':()=>{S.u.dataTab='regras';nav('dados')},
  'diesel-uf':async el=>{const C=LASTC;if(!C)return;try{await salvarUfDiesel(C.def,el.dataset.v);closePops();toast('Diesel: '+nomeUF(el.dataset.v));DB.bump();go()}catch(e){toast('Não foi possível salvar: '+e.message)}},
  print:()=>window.print(),
  csv:el=>{download((el.dataset.w==='fleet'?'frota':'motoristas')+'-'+S.client+'-'+TODAY+'.csv',csvFor(LASTC,el.dataset.w),'text/csv;charset=utf-8');toast('Arquivo CSV gerado')},
  'ov-metric':el=>{S.u.ovMetric=el.dataset.v;go()},
  'trend-win':el=>{S.u.trendWin=el.dataset.v;go()},
  'ret-tab':el=>{S.u.retTab=el.dataset.v;go(true)},
  'client-open':el=>{setClient(el.dataset.id);S.u.retTab='cliente';render({skeleton:true})},
  'fleet-tab':el=>{S.u.fleetTab=el.dataset.v;if(el.dataset.v==='mapa'&&LASTC)loadExtras(LASTC.cid);go(true)},
  'fleet-filial':el=>{S.u.fleetFilial=el.dataset.v;go()},
  'fleet-rank':el=>{S.u.fleetRank=el.dataset.v;go()},
  'veh-open':el=>openVehicle(LASTC,el.dataset.id),
  'drv-tab':el=>{S.u.drvTab=el.dataset.v;go(true)},
  'drv-det':el=>{S.u.drvDet=el.dataset.v;refreshDetail()},
  'drv-sel':el=>{S.u.drvSel=el.dataset.id;refreshDetail()},
  'drv-filter':el=>{S.u.drvFilter=el.dataset.v;go()},
  'drv-open':el=>{closeOv();S.u.drvSel=el.dataset.id;S.u.drvTab='lista';S.u.drvDet='resumo';S.u.drvFilter='all';S.u.drvQ='';nav('motoristas')},
  'go-drv':el=>{closePops();S.u.drvTab='lista';S.u.drvFilter=el.dataset.f==='down'?'down':'all';S.u.drvSort=el.dataset.f==='down'?'money':'kmpl';nav('motoristas')},
  'msg-drv':el=>{S.u.msgDrv=el.dataset.id;go()},
  'msg-gran':el=>{S.u.msgGran=el.dataset.v;go()},
  'evo-gran':el=>{S.u.evoGran=el.dataset.v;go()},
  'day-view':el=>{S.u.dayView=el.dataset.v;refreshDetail()},
  'wk-wa':el=>weekWa(el,true),
  'wk-copy':el=>weekWa(el,false),
  'ind-csv':el=>{const C=LASTC,id=el.dataset.id,c=C.cur.byVeh[id],p=C.prev.byVeh[id];const rows=IND.map(i=>[i.l,fInd(i,i.get(c)),fInd(i,p?i.get(p):null),fInd(i,i.get(C.cur)),indRes(i,c,p).label]);download('indicadores-'+id+'-'+TODAY+'.csv','﻿'+[['Indicador','Período','Comparação','Média da frota','Resultado'],...rows].map(r=>r.map(csvQ).join(';')).join('\n'),'text/csv;charset=utf-8');toast('Arquivo CSV gerado')},
  'map-filter':el=>{S.u.mapFilter=el.dataset.v;applyAllFilters()},
  'map-focus':el=>mapFocus(el.dataset.id),
  'map-route':el=>mapRoute(el.dataset.id),
  'map-route-clear':()=>mapRouteClear(),
  'cmp-metric':el=>{S.u.cmpMetric=el.dataset.v;go()},
  'week-sel':el=>{S.u.weekSel=+el.dataset.v;go()},
  'prem-month':el=>{S.u.premMonth=el.dataset.v;go()},
  'msg-day':el=>{const t=Math.min(T0,dIdx(S.u.msgDay||dIso(T0-1))+ +el.dataset.v);S.u.msgDay=dIso(t);go()},
  'msg-copy':async el=>{const m=dayMessage(S.client,ST,el.dataset.v,+el.dataset.t);if(m)toast(await copyText(m.text)?'Mensagem copiada':'Não foi possível copiar')},
  'msg-wa':el=>{const m=dayMessage(S.client,ST,el.dataset.v,+el.dataset.t);const url=m&&waUrl(m.phone,m.text);if(url){window.open(url,'_blank','noopener');toast('Abrindo o WhatsApp')}},
  'data-tab':el=>{S.u.dataTab=el.dataset.v;go(true)},
  'est-gran':el=>{S.u.estGran=el.dataset.v;go()},
  'contrato-salvar':async el=>{
    const ler=id=>{const s=String(($('#'+id)||{}).value||'').trim().replace(/\s|R\$/g,'');if(!s)return null;return Number(s.includes(',')?s.replace(/\./g,'').replace(',','.'):/^\d{1,3}(\.\d{3})+$/.test(s)?s.replace(/\./g,''):s)};
    const unit=ler('ct-valor'),units=ler('ct-qtd');
    if((unit!==null&&!(unit>0))||(units!==null&&!(units>0&&Number.isInteger(units)))){toast('Confira os números: valor maior que zero e unidades inteiras.');return}
    el.disabled=true;
    try{const r=await salvarContrato(LASTC.def,unit,units);DB.bump();toast(r===false?'Nada mudou.':'Contrato salvo: vale em todas as telas.');go()}
    catch(e){el.disabled=false;toast(/403|permiss/i.test(e.message)?'Só CS e administrador mudam o contrato.':'Não foi possível salvar: '+e.message)}
  },
};
function setClient(id){S.client=id;S.u.drvSel=null;S.u.fleetFilial='';S.u.msgDay='';S.u.msgDrv=null;S.u.weekSel=null;S.u.premMonth='';savePrefs()}

/* ---------- eventos ---------- */
function bind(){
  document.addEventListener('click',async e=>{
    const el=e.target.closest('[data-act]');
    if(!e.target.closest('.relative'))closePops();
    if(!e.target.closest('.search'))$('#sres').classList.remove('open');
    if(!el)return;const a=ACT[el.dataset.act];if(!a)return;
    if(!(el.tagName==='A'&&el.getAttribute('href')))e.preventDefault();
    if(el.dataset.act==='pop-close'){closePops();return}
    await a(el,e);
  });
  document.addEventListener('change',e=>{
    const el=e.target.closest('[data-chg]');if(!el)return;const k=el.dataset.chg;
    if(k==='drv-sort'){S.u.drvSort=el.value;go()}
    else if(k==='msg-drv'){S.u.msgDrv=el.value;go()}
    else if(k==='evo-metric'){S.u.evoMetric=el.value;go()}
    else if(k==='day-metric'){S.u.dayMetric=el.value;refreshDetail()}
    else if(k==='msg-day'){if(el.value){S.u.msgDay=el.value>TODAY?TODAY:el.value;go()}}
    else if(k==='pfrom'||k==='pto'){S.period[k==='pfrom'?'from':'to']=el.value;if(S.period.from&&S.period.to&&S.period.from<=S.period.to)go()}
  });
  document.addEventListener('input',e=>{const el=e.target.closest('[data-chg="drv-q"]');if(el){S.u.drvQ=el.value;clearTimeout(bind.t);bind.t=setTimeout(go,180)}});
  $('#clientSel').addEventListener('change',e=>{setClient(e.target.value);render({skeleton:true})});
  window.addEventListener('resize',()=>{clearTimeout(bind.rz);bind.rz=setTimeout(()=>fitCharts(document,true),180)});
  window.addEventListener('hashchange',()=>{const {path,q}=parseHash();const t=TABKEY[path];if(t&&q.aba&&t[1][q.aba]){S.u[t[0]]=t[1][q.aba]===1?q.aba:t[1][q.aba]}closeOv();closePops();$('#sidebar').classList.remove('open');render({skeleton:false})});
  /* busca global */
  const gs=$('#gs'),sres=$('#sres');
  gs.addEventListener('input',()=>{const q=gs.value.trim().toLowerCase();if(q.length<2){sres.classList.remove('open');return}
    const rs=(LASTC?LASTC.fleet.vehicles:[]).filter(v=>v.id.toLowerCase().includes(q)||v.driver.toLowerCase().includes(q)).slice(0,8);
    sres.innerHTML=rs.length?rs.map(v=>`<button data-act="gs-pick" data-id="${esc(v.id)}"><div class="ava">${initials(v.driver)}</div><div><b>${esc(v.driver)}</b><div class="xs muted">${esc(v.id)}</div></div></button>`).join(''):'<div class="empty"><b>Nada encontrado</b></div>';sres.classList.add('open')});
  gs.addEventListener('keydown',e=>{if(e.key==='Enter'){const b=$('button',sres);if(b)b.click()}if(e.key==='Escape'){sres.classList.remove('open');gs.blur()}});
  ACT['gs-pick']=el=>{gs.value='';sres.classList.remove('open');ACT['drv-open'](el)};
  document.addEventListener('keydown',e=>{if(e.key==='/'&&!/INPUT|TEXTAREA|SELECT/.test(document.activeElement.tagName)){e.preventDefault();gs.focus()}if(e.key==='Escape'){closeOv();closePops()}});
  /* tooltip */
  const tt=$('#tt');
  document.addEventListener('mouseover',e=>{const el=e.target.closest('[data-tt]');if(!el){tt.classList.remove('on');return}tt.innerHTML=el.getAttribute('data-tt');tt.classList.add('on')});
  document.addEventListener('mousemove',e=>{if(!tt.classList.contains('on'))return;const w=tt.offsetWidth,h=tt.offsetHeight;let x=e.clientX+14,y=e.clientY+16;if(x+w>innerWidth-8)x=e.clientX-w-14;if(y+h>innerHeight-8)y=e.clientY-h-14;tt.style.left=x+'px';tt.style.top=y+'px'});
  document.addEventListener('mouseleave',()=>tt.classList.remove('on'));
  window.addEventListener('resize',()=>{placeNav();placeTabs()});
  window.addEventListener('scroll',()=>$('#topbar').classList.toggle('scrolled',scrollY>8),{passive:true});
  /* posição ao vivo: atualiza a cada 2 minutos com a aba visível */
  setInterval(()=>{if(REAL.on&&!document.hidden&&LASTC)loadExtras(LASTC.cid)},120000);
}

/* ---------- boot ---------- */
async function boot(){
  document.documentElement.dataset.theme=S.theme;
  $('#root').innerHTML=`<div class="view">${skeleton()}</div>`;
  await detectServer();
  let list=[];
  try{list=await Gobrax.api.getClients()}catch(e){$('#root').innerHTML=`<div class="view"><div class="card" style="margin-top:40px"><h2>Não foi possível carregar os dados</h2><p class="muted">${esc(e.message)}</p><button class="btn primary" onclick="location.reload()">Tentar de novo</button></div></div>`;return}
  if(REAL.on){CLIENT_DEFS.length=0;list.forEach(c=>CLIENT_DEFS.push(c));try{REAL.prem=await Gobrax.api.getPremissas()}catch(e){REAL.prem=[]}}
  else for(const c of list){const d=CLIENT_DEFS.find(x=>x.id===c.id);if(d)Object.assign(d,c)}
  if(!CLIENT_DEFS.some(c=>c.id===S.client))S.client=CLIENT_DEFS[0]?CLIENT_DEFS[0].id:'';
  $('#root').innerHTML=shell();
  if(S.collapsed)$('#appRoot').classList.add('collapsed');
  /* menu recolhido: clicar abre por cima; tirar o mouse recolhe de volta */
  {const sb=$('#sidebar');const peek=on=>{if(sb.classList.contains('peek')===on)return;sb.classList.toggle('peek',on);setTimeout(placeNav,320)};
    sb.addEventListener('click',e=>{if(!S.collapsed||innerWidth<=900||e.target.closest('[data-act="collapse"]'))return;peek(true)});
    sb.addEventListener('mouseleave',()=>peek(false));
    sb.addEventListener('focusout',e=>{if(!sb.contains(e.relatedTarget))peek(false)});}
  bind();
  const {path,q}=parseHash();const t=TABKEY[path];if(t&&q.aba&&t[1][q.aba])S.u[t[0]]=t[1][q.aba]===1?q.aba:t[1][q.aba];
  await render({skeleton:true});
  buscarDiesel().then(ok=>{if(ok){DB.bump();go()}});setInterval(()=>buscarDiesel().then(ok=>{if(ok){DB.bump();go()}}),6*3600000);
  document.fonts&&document.fonts.ready.then(()=>{placeNav();placeTabs()});
}
