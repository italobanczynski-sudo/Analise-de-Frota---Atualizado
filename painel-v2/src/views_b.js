/* =====================================================================
   TELAS · Frota · Motoristas
   ===================================================================== */
const initials=n=>String(n||'?').split(' ').filter(Boolean).map(x=>x[0]).slice(0,2).join('').toUpperCase();
const sigText=v=>{const m=v.days;return m+' dia'+(m>1?'s':'')+' sem sinal'};
const scoreBadge=s=>s==null||!isFinite(s)?'':`<span class="badge ${scoreBand(Math.round(s)).cls}">NOTA ${Math.round(s)}</span>`;

/* ---------- FROTA ---------- */
function filialChips(C){const fs=[...new Set(C.fleet.vehicles.map(v=>v.filial))].sort();if(fs.length<2)return'';return`<div class="chips" role="group" aria-label="Filial"><button class="chip ${!S.u.fleetFilial?'on':''}" data-act="fleet-filial" data-v="">Todas</button>${fs.map(f=>`<button class="chip ${S.u.fleetFilial===f?'on':''}" data-act="fleet-filial" data-v="${esc(f)}">${esc(f)}</button>`).join('')}</div>`}
function fleetVehicles(C){const f=S.u.fleetFilial;return C.fleet.vehicles.filter(v=>!f||v.filial===f)}
function vFleet(C){
  const u=S.u;const susp=suspCount(C);
  const tabs=tabsHtml([['visao','Visão da frota'],['indicadores','Todos os indicadores'],['qualidade','Qualidade do dado',susp||null],['mapa','Mapa']],u.fleetTab,'fleet-tab');
  const head=pageH('Frota',`${esc(C.def.name)} · ${C.P.label} · ${C.fleet.vehicles.length} veículos`,`<button class="btn ghost" data-act="csv" data-w="fleet">${ico('download',16)} Exportar CSV</button>`);
  const vs=fleetVehicles(C);
  let body='';
  if(u.fleetTab==='visao'){
    const rows=vs.map(v=>({v,c:C.cur.byVeh[v.id],p:C.prev.byVeh[v.id],st:vehStatus(C.cur,C.prev,v.id)}));
    const rk={veh:(a,b)=>(b.c&&b.c.kmpl||-1)-(a.c&&a.c.kmpl||-1),down:(a,b)=>(a.st.ch??99)-(b.st.ch??99),km:(a,b)=>(b.c?b.c.km:-1)-(a.c?a.c.km:-1)}[u.fleetRank];
    rows.sort(rk);
    const cnt={up:0,down:0,flat:0,base:0};rows.forEach(r=>cnt[r.st.k]++);
    const ns=new Map(C.cur.noSignal.map(x=>[x.id,x.days]));
    body=`<div class="grid g4 sec">${[['Evoluíram',cnt.up,'ok'],['Estáveis',cnt.flat,'n'],['Regrediram',cnt.down,'bad'],['Sem base de comparação',cnt.base,'n']].map(([l,n,c])=>`<div class="card lift rise" style="padding:16px 20px"><div class="tag-sm">${l}</div><div class="num" style="font-size:34px;font-weight:800">${count(n,0)}</div><span class="badge ${c}">${l==='Sem base de comparação'?'inclui quem não rodou no período':'vs. '+C.P.cmpLabel.split(' (')[0]}</span></div>`).join('')}</div>
    ${pendBlock(C)}
    <div class="card rise sec">${cardH('Veículos',`${rows.length} no filtro atual`,filialChips(C)+segHtml([['veh','Melhor consumo'],['down','Maior queda'],['km','Mais km']],u.fleetRank,'fleet-rank'))}<div class="scroll-x"><table class="tbl"><thead><tr><th>Veículo</th><th>Condutor</th><th class="r">Km</th><th class="r">Km/l</th><th class="r">Variação</th><th class="r">Nota</th><th>Situação</th></tr></thead><tbody>${rows.map(({v,c,st})=>`<tr class="click" data-act="veh-open" data-id="${esc(v.id)}"><td><b>${esc(v.id)}</b></td><td>${esc(v.driver)}</td><td class="r num">${c?fmt.n(c.km,0):'—'}</td><td class="r num b">${c?fmt.kmpl(c.kmpl):'—'}</td><td class="r">${dl(st.ch,true)}</td><td class="r num">${c?fmt.n(c.score,0):'—'}</td><td>${ns.has(v.id)?`<span class="badge bad">${sigText({days:ns.get(v.id)})}</span>`:`<span class="badge ${st.cls}">${st.label}</span>`}</td></tr>`).join('')}</tbody></table></div></div>`;
  }else if(u.fleetTab==='indicadores'){
    const rows=vs.map(v=>({v,c:C.cur.byVeh[v.id]})).filter(r=>r.c).sort((a,b)=>(b.c.kmpl||0)-(a.c.kmpl||0));
    const avg=C.cur;const pct=v=>v==null?'—':fmt.n(v,1)+'%';
    body=`<div class="card rise sec">${cardH('Todos os indicadores por veículo','Uma única tabela, calculada com os mesmos dados da Visão geral',filialChips(C))}<div class="scroll-x"><table class="tbl"><thead><tr><th>Veículo</th><th class="r">Km</th><th class="r">Litros</th><th class="r">Km/l</th><th class="r">Freadas /100 km</th><th class="r">Freadas em alta vel. /100 km</th><th class="r">Vel. média</th><th class="r">Faixa verde</th><th class="r">Motor ocioso</th><th class="r">Nota</th></tr></thead><tbody>
    <tr style="background:var(--brand-soft)"><td><b>Média da frota</b></td><td class="r num">${fmt.n(avg.km,0)}</td><td class="r num">${fmt.n(avg.l,0)}</td><td class="r num b">${fmt.kmpl(avg.kmpl)}</td><td class="r num">${fmt.n(avg.br100,1)}</td><td class="r num">${fmt.n(avg.brh100,2)}</td><td class="r num">${fmt.n(avg.speed,1)}</td><td class="r num">${pct(avg.green)}</td><td class="r num">${pct(avg.idle)}</td><td class="r num b">${fmt.n(avg.score,0)}</td></tr>
    ${rows.map(({v,c})=>`<tr class="click" data-act="veh-open" data-id="${esc(v.id)}"><td><b>${esc(v.id)}</b></td><td class="r num">${fmt.n(c.km,0)}</td><td class="r num">${fmt.n(c.l,0)}</td><td class="r num b">${fmt.kmpl(c.kmpl)}</td><td class="r num">${fmt.n(c.br100,1)}</td><td class="r num">${fmt.n(c.brh100,2)}</td><td class="r num">${fmt.n(c.speed,1)}</td><td class="r num">${pct(c.green)}</td><td class="r num">${pct(c.idle)}</td><td class="r num b">${fmt.n(c.score,0)}</td></tr>`).join('')}</tbody></table></div></div>`;
  }else if(u.fleetTab==='qualidade'){
    const rs=suspRecs(C.cid,ST).filter(r=>r.t>=C.P.from&&r.t<=C.P.to).sort((a,b)=>b.t-a.t);const R=ST.rules;
    const why=r=>{const k=r.km/r.l;return k<R.minKmpl?`consumo de ${fmt.n(k,2)} km/l, abaixo do mínimo de ${fmt.n(R.minKmpl,1)}`:k>R.maxKmpl?`consumo de ${fmt.n(k,2)} km/l, acima do máximo de ${fmt.n(R.maxKmpl,1)}`:`${fmt.n(r.km,0)} km num só dia, acima do máximo de ${fmt.n(R.maxKmDay,0)}`};
    body=`<div class="grid g-main sec"><div class="card rise">${cardH('Registros suspeitos',`${rs.length} registros ficam fora das médias até serem revisados no sistema de origem.`)}${rs.length?`<div class="stack">${rs.map(r=>`<div class="msgcard"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><div><b>${esc(r.veh)}</b> <span class="muted">${esc(C.fleet.byId[r.veh].driver)} · ${fmt.dmy(r.date)}</span></div><span class="badge warn">SUSPEITO</span></div><div class="sm">${fmt.n(r.km,0)} km · ${fmt.n(r.l,1)} L · ${why(r)}.</div></div>`).join('')}</div>`:`<div class="empty"><b>Nada para revisar</b>Todos os registros do período estão dentro dos limites.</div>`}</div>
    <div class="card brand rise stack"><div class="swoosh"></div>${cardH('Como sinalizamos','')}<p class="sm" style="margin:0">Registros fora destes limites são sinalizados e ficam de fora das médias, do estudo e do índice de saúde. Assim, todas as telas mostram o mesmo número.</p><p class="sm" style="margin:0">A correção é feita no sistema de origem; aqui a plataforma só avisa onde olhar.</p><div class="kv"><span>Consumo mínimo</span><span>${fmt.n(R.minKmpl,1)} km/l</span></div><div class="kv"><span>Consumo máximo</span><span>${fmt.n(R.maxKmpl,1)} km/l</span></div><div class="kv"><span>Máximo de km por dia</span><span>${fmt.n(R.maxKmDay,0)} km</span></div></div></div>`;
  }else{
    body=fleetMapBlock(C)+`<div class="card rise sec">${cardH('Sem transmitir',`${C.cur.noSignal.length} veículos sem nenhum dia de dado nos últimos 3 dias`,eqInfo())}${C.cur.noSignal.length?`<div class="grid g3" style="gap:8px">${C.cur.noSignal.map(x=>`<div class="li" data-act="veh-open" data-id="${esc(x.id)}"><div class="ava">${ico('alert',16)}</div><div style="flex:1"><b>${esc(x.id)}</b><div class="xs muted">${esc((C.fleet.byId[x.id]||{}).driver||'')}</div></div><span class="badge bad">${x.days} dias</span></div>`).join('')}</div>`:`<div class="empty"><b>Todos transmitindo</b></div>`}</div>`;
  }
  return head+tabs+body;
}
function openVehicle(C,id){
  const v=C.fleet.byId[id];if(!v)return;const c=C.cur.byVeh[id],p=C.prev.byVeh[id];const st=vehStatus(C.cur,C.prev,id);
  const vf=x=>x===id;const ds=dailySeries(C.cid,ST,C.P.from,C.P.to,vf);const vals=ds.map(d=>d.a?d.a.kmpl:null);
  const chart=lineChart({labels:ds.map(d=>fmt.dm(d.date)),series:[{vals,color:'var(--laranja)',area:true,dots:ds.length<=31}],base:p&&p.kmpl?{v:p.kmpl,label:'comparação '+fmt.n(p.kmpl,2)}:null,h:210,tip:i=>ds[i].a?`<b>${fmt.dm(ds[i].date)}</b><br>${fmt.n(ds[i].a.km,0)} km · ${fmt.n(ds[i].a.kmpl,2)} km/l`:'',aria:'Consumo diário do veículo'});
  const row=(l,a,b,good,d=1,suf='')=>`<tr><td>${l}</td><td class="r num b">${a!=null&&isFinite(a)?fmt.n(a,d)+suf:'—'}</td><td class="r num muted">${b!=null&&isFinite(b)?fmt.n(b,d)+suf:'—'}</td><td class="r">${a!=null&&b?dl((a-b)/b*100,good):'—'}</td></tr>`;
  openDrawer(`<div class="row" style="justify-content:space-between;flex-wrap:nowrap"><div class="row" style="flex-wrap:nowrap"><div class="ava lg">${initials(v.driver)}</div><div><h2 style="font-size:24px;font-weight:900">${esc(id)}</h2><div class="muted">${esc(v.driver)}</div></div></div><button class="btn ghost sm" data-act="ov-close" aria-label="Fechar">${ico('x',14)}</button></div>
  <div class="row" style="margin:14px 0"><span class="badge ${st.cls}">${st.label}</span>${st.ch!=null?dl(st.ch,true):''}</div>
  <div class="card" style="margin-bottom:16px">${cardH('Consumo dia a dia',C.P.label)}${chart}</div>
  <div class="card" style="margin-bottom:16px">${cardH('Trajeto de hoje',REAL.on?'Rastro de ignição ligada desde a meia-noite, vindo da API Gobrax.':'Posição simulada sobre a rodovia. Na integração vem da API de rastreamento.')}${routeMapBlock(C,id,230)}</div>
  <div class="card">${cardH('Sinais do veículo','Período atual contra o de comparação')}${sinaisVeiculo(C,id)}</div>
  <div class="row" style="margin-top:16px"><button class="btn primary" data-act="drv-open" data-id="${esc(id)}">Abrir condutor</button><button class="btn ghost" data-act="ov-close">Fechar</button></div>`);
}

/* ---------- MOTORISTAS ---------- */
function drvRows(C){
  const gl=Object.fromEntries(gainsLosses(C.cur,C.prev,C.price).map(g=>[g.id,g]));
  return C.fleet.vehicles.map(v=>({v,c:C.cur.byVeh[v.id],p:C.prev.byVeh[v.id],st:vehStatus(C.cur,C.prev,v.id),g:gl[v.id]})).filter(r=>r.c);
}
function vDrivers(C){
  const u=S.u;
  const tabs=tabsHtml([['resumo','Resumo'],['comportamento','Comportamento'],['comparativo','Comparativo'],['semanal','Resultado semanal'],['lista','Condutores',null],['mensagens','Mensagens'],['premiacao','Nota e premiação']],u.drvTab,'drv-tab');
  const head=pageH('Motoristas',`${esc(C.def.name)} · ${C.P.label} · comparado ao ${C.P.cmpLabel}`,`<button class="btn ghost" data-act="csv" data-w="drivers">${ico('download',16)} Exportar CSV</button>`);
  const note=REAL.on?`<div class="xs muted" style="margin:-12px 0 16px">Cada linha é um veículo com o motorista que mais o dirigiu no período (vínculo da API Gobrax).</div>`:'';
  return head+tabs+note+({resumo:drvResumo,comportamento:drvBehavior,lista:drvList,comparativo:drvCompare2,semanal:drvWeekly2,mensagens:drvMessages,premiacao:drvPrize}[u.drvTab]||drvResumo)(C);
}
function drvList(C){
  const u=S.u;let rows=drvRows(C);
  const counts={all:rows.length,up:rows.filter(r=>r.st.k==='up').length,down:rows.filter(r=>r.st.k==='down').length,base:rows.filter(r=>r.st.k==='base').length};
  if(u.drvFilter!=='all')rows=rows.filter(r=>r.st.k===u.drvFilter);
  if(u.drvQ){const q=u.drvQ.toLowerCase();rows=rows.filter(r=>r.v.id.toLowerCase().includes(q)||r.v.driver.toLowerCase().includes(q))}
  const so={money:(a,b)=>(a.g?a.g.money:0)-(b.g?b.g.money:0),kmpl:(a,b)=>(b.c.kmpl||0)-(a.c.kmpl||0),score:(a,b)=>(b.c.score??-1)-(a.c.score??-1),name:(a,b)=>a.v.driver.localeCompare(b.v.driver)}[u.drvSort];rows.sort(so);
  if(!u.drvSel||!C.fleet.byId[u.drvSel])u.drvSel=rows[0]?rows[0].v.id:null;
  const chip=(k,l)=>`<button class="chip ${u.drvFilter===k?'on':''}" data-act="drv-filter" data-v="${k}">${l} · ${counts[k]}</button>`;
  const list=rows.length?rows.map(r=>`<div class="li ${r.v.id===u.drvSel?'on':''}" data-act="drv-sel" data-id="${esc(r.v.id)}"><div class="ava">${initials(r.v.driver)}</div><div style="flex:1;min-width:0"><b style="display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">${esc(r.v.driver)}</b><span class="xs muted">${esc(r.v.id)} · ${fmt.kmpl(r.c.kmpl)} km/l</span></div><div style="text-align:right"><span class="badge ${r.st.cls}">${r.st.label}</span>${r.g?`<div class="xs num b" style="color:${r.g.money>=0?'var(--ok-t)':'var(--bad-t)'}">${r.g.money>=0?'+':'−'}${fmt.brl(Math.abs(r.g.money))}</div>`:''}</div></div>`).join(''):`<div class="empty"><b>Nenhum condutor</b>Ajuste o filtro ou a busca.</div>`;
  const left=`<div class="card listc rise" style="padding:18px"><div class="stack" style="gap:10px;margin-bottom:12px"><div class="chips">${chip('all','Todos')}${chip('up','Evoluíram')}${chip('down','Regrediram')}${chip('base','Sem base')}</div><div class="row" style="flex-wrap:nowrap"><input class="inp" data-chg="drv-q" placeholder="Buscar por nome ou placa" value="${esc(u.drvQ)}" style="flex:1" aria-label="Buscar condutor"><select class="inp" data-chg="drv-sort" style="width:auto" aria-label="Ordenar"><option value="money" ${u.drvSort==='money'?'selected':''}>Maior perda</option><option value="kmpl" ${u.drvSort==='kmpl'?'selected':''}>Melhor consumo</option><option value="score" ${u.drvSort==='score'?'selected':''}>Melhor nota</option><option value="name" ${u.drvSort==='name'?'selected':''}>Nome</option></select></div></div><div class="scrolly stack" style="gap:2px">${list}</div></div>`;
  return`<div class="md sec">${left}<div id="drv-detail">${u.drvSel?drvDetail(C,u.drvSel):'<div class="card empty"><b>Selecione um condutor</b></div>'}</div></div>`;
}
function drvDetail(C,id){
  const u=S.u;const v=C.fleet.byId[id];const c=C.cur.byVeh[id],p=C.prev.byVeh[id];const st=vehStatus(C.cur,C.prev,id);
  const dt=tabsHtml([['resumo','Painel'],['indicadores','Todos os indicadores'],['dia','Dia a dia'],['msg','Mensagens']],u.drvDet,'drv-det');
  const head=`<div class="card rise"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><div class="row" style="flex-wrap:nowrap"><div class="ava lg">${initials(v.driver)}</div><div><h2 style="font-size:24px;font-weight:900">${esc(v.driver)}</h2><div class="muted">${esc(id)}</div></div></div><div class="row"><span class="badge ${st.cls}">${st.label}</span>${c?scoreBadge(c.score):''}</div></div></div>`;
  let inner='';
  if(!c)return head+'<div class="card empty rise sec"><b>Sem rodagem no período</b></div>';
  if(u.drvDet==='resumo'){
    inner=drvDash(C,id);
  }else if(u.drvDet==='indicadores'){
    inner=drvIndTable(C,id);
  }else if(u.drvDet==='dia'){
    const vf=x=>x===id;const ds=dailySeries(C.cid,ST,C.P.from,C.P.to,vf);const ref=p&&p.kmpl?p.kmpl:c.kmpl;const lead=(wdOf(C.P.from)+6)%7;
    const cells=['seg','ter','qua','qui','sex','sáb','dom'].map(d=>`<div class="hd">${d}</div>`).join('')+'<div></div>'.repeat(lead)+ds.map((d,i)=>{const a=d.a;const k=a?a.kmpl:null;const cl=k==null?'e':k>=ref*1.02?'g':k<=ref*0.98?'r':'y';return`<div class="${cl}" style="--i:${i}" ${a?`data-tt="<b>${fmt.dmy(d.date)}</b><br>${fmt.n(a.km,0)} km · ${fmt.n(k,2)} km/l"`:`data-tt="${fmt.dmy(d.date)} · sem rodagem"`}>${+d.date.slice(8)}</div>`}).join('');
    inner=`<div class="card sec">${cardH('Dia a dia','Cor pelo consumo contra a média de comparação')}<div class="hm">${cells}</div><div class="leg" style="margin-top:14px"><span><i style="background:var(--ok-s)"></i>Melhor que a comparação</span><span><i style="background:var(--warn-s)"></i>Igual</span><span><i style="background:var(--bad-s)"></i>Pior</span><span><i style="background:var(--soft)"></i>Sem rodagem</span></div></div>`;
  }else{
    const recs=validRecs(C.cid,ST).filter(r=>r.veh===id&&r.t>=T0-6).sort((a,b)=>b.t-a.t);
    inner=`<div class="card sec">${cardH('Mensagens dos últimos dias','Texto pronto para enviar ao condutor')}<div class="stack">${recs.length?recs.map(r=>{const m=dayMessage(C.cid,ST,id,r.t);return m?msgCardHtml(C,m,r.t):''}).join(''):'<div class="empty"><b>Sem rodagem nos últimos 7 dias</b></div>'}</div></div>`;
  }
  return head+dt+inner;
}
function msgCardHtml(C,m,t){
  return`<div class="msgcard rise"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><div><b>${esc(m.name)}</b> <span class="xs muted">${esc(m.veh)} · ${fmt.dm(dIso(t))}</span></div>${scoreBadge(m.score)}</div><div class="bubble">${esc(m.text)}</div><div class="row"><button class="btn sm ghost" data-act="msg-copy" data-v="${esc(m.veh)}" data-t="${t}">${ico('copy',14)} Copiar mensagem</button>${m.phone?`<button class="btn sm primary" data-act="msg-wa" data-v="${esc(m.veh)}" data-t="${t}">${ico('msg',14)} WhatsApp</button>`:''}</div></div>`;
}
function drvMessages(C){
  const u=S.u;if(!u.msgDay)u.msgDay=dIso(T0-1);const t=dIdx(u.msgDay);
  const msgs=C.fleet.vehicles.map(v=>dayMessage(C.cid,ST,v.id,t)).filter(Boolean).sort((a,b)=>(a.score??999)-(b.score??999));
  const hist=msgHist(C);
  const evo=hist.length?msgEvolution(C,hist):REAL.on?`<div class="card rise"><div class="empty"><b>${REAL.busy['msg'+C.cid]?'Carregando as mensagens enviadas…':'Nenhuma mensagem marcada como enviada'}</b>Quando o CS ou o robô marcar uma mensagem como enviada, a evolução da nota aparece aqui.</div></div>`:'';
  return evo+`<div class="card rise sec">${cardH('Mensagem do dia',`${msgs.length} condutores rodaram em ${fmt.dmy(u.msgDay)}`,`<div class="row"><button class="btn sm ghost" data-act="msg-day" data-v="-1" aria-label="Dia anterior" style="transform:scaleX(-1)">${ico('chev',14)}</button><input class="inp" type="date" style="width:160px;height:34px" max="${TODAY}" value="${u.msgDay}" data-chg="msg-day" aria-label="Dia"><button class="btn sm ghost" data-act="msg-day" data-v="1" aria-label="Próximo dia">${ico('chev',14)}</button></div>`)}<div class="grid g2">${msgs.length?msgs.map(m=>msgCardHtml(C,m,t)).join(''):'<div class="empty" style="grid-column:1/-1"><b>Ninguém rodou nesse dia</b>Escolha outra data.</div>'}</div></div>`;
}
function drvPrize(C){
  const u=S.u;const R=ST.rules;const months=premMonths(C.cid,ST);if(!u.premMonth||!months.includes(u.premMonth))u.premMonth=months.includes(TODAY.slice(0,7))?TODAY.slice(0,7):months[months.length-1]||TODAY.slice(0,7);
  const rows=premiacao(C.cid,ST,u.premMonth);const tot=rows.reduce((a,r)=>a+r.bonus,0);const elig=rows.filter(r=>r.bonus>0).length;const official=rows.some(r=>r.official);
  const dist={g:0,y:0,r:0};rows.forEach(r=>{const k=scoreBand(r.score).k;if(k in dist)dist[k]++});
  return`<div class="grid g-main sec"><div class="card rise">${cardH('Nota e premiação',`Mês de ${MESES[+u.premMonth.slice(5)-1]} de ${u.premMonth.slice(0,4)}${official?' · resultado oficial da premiação Gobrax':''}`,segHtml(months.slice(-8).map(m=>[m,MESES[+m.slice(5)-1]+'/'+m.slice(2,4)]),u.premMonth,'prem-month'))}${rows.length?`<div class="scroll-x"><table class="tbl"><thead><tr><th>#</th><th>Condutor</th><th class="r">Nota</th><th>Faixa</th><th class="r">Mês anterior</th><th class="r">Km</th><th class="r">Bônus</th></tr></thead><tbody>${rows.map((r,i)=>{const b=scoreBand(r.score);return`<tr class="click" data-act="drv-open" data-id="${esc(r.id)}"><td class="muted">${i+1}</td><td><b>${esc(r.name)}</b> <span class="xs muted">${r.id!==r.name?esc(r.id):''}</span></td><td class="r num b">${r.score}</td><td><span class="badge ${b.cls}">${b.label}</span></td><td class="r">${r.prev!=null?dl(r.score-r.prev,true,0,' pts'):'—'}</td><td class="r num">${fmt.n(r.km,0)}</td><td class="r num b" style="color:${r.bonus>0?'var(--ok-t)':'var(--muted)'}">${r.bonus>0?fmt.brl(r.bonus,2):'—'}</td></tr>`}).join('')}</tbody></table></div>`:'<div class="empty"><b>Sem nota neste mês</b>A API não trouxe nota para os motoristas deste cliente.</div>'}</div>
  <div class="stack"><div class="card brand rise"><div class="swoosh"></div>${cardH('Total de bônus',`${elig} condutores receberam`)}<div class="num" style="font-size:36px;font-weight:800">${count(tot,0,'R$ ')}</div><div class="row" style="margin-top:12px"><span class="badge ok">${dist.g} bom</span><span class="badge warn">${dist.y} atenção</span><span class="badge bad">${dist.r} crítico</span></div></div>
  <div class="card rise stack">${cardH('Regra de premiação',official?'Valores calculados pela Gobrax (v2/driversOverview)':'Estimativa: valores da política do cliente')}${official?'<div class="sm muted">Nota, km e prêmio de cada motorista vêm prontos da API Gobrax. A tela só organiza e compara com o mês anterior.</div>':`<div class="kv"><span>Nota mínima para receber</span><b class="num">${R.bonusMin}</b></div><div class="kv"><span>Bônus por km rodado</span><b class="num">${fmt.brl(R.bonusRate,2)}</b></div>`}</div></div></div>`;
}

/* pendências de implantação (item 4): mesmos critérios e contagens de todas as telas */
function pendBlock(C){
  const e=C.eq;const li=(id,motivo,desde)=>`<tr class="click" data-act="veh-open" data-id="${esc(id)}"><td><b>${esc(id)}</b></td><td>${motivo}</td><td class="r">${desde}</td></tr>`;
  const ultimo=id=>{const l=C.fleet.lastComm&&C.fleet.lastComm[id];return l!=null?fmt.dmy(dIso(l))+' ('+(T0-l)+' dias)':'sem dado'};
  const rows=[...e.semTransmitir.map(x=>li(x.id,'Sem transmitir: nenhum dado nos últimos 3 dias',x.days+' dias')),
    ...e.semMotorista.map(id=>li(id,'Sem motorista identificado: rodou sem vínculo de motorista na API',e.desde[id]!=null?'desde '+fmt.dmy(dIso(e.desde[id]))+' ('+(T0-e.desde[id])+' dias)':'sem dado')),
    ...e.semBase.map(id=>li(id,'Sem base de comparação: rodou no período, mas não no de comparação','último dado '+ultimo(id)))].join('');
  const cob=e.longos.length?`<div class="sm" style="margin-top:10px;background:var(--bad-s,#fde8e6);border:1px solid var(--bad);border-radius:10px;padding:10px 12px">⚠ <b>${e.longos.length} equipamento${e.longos.length>1?'s':''} sem transmitir há mais de ${ST.rules.diasCobranca} dias: risco de cobrança indevida.</b> ${e.longos.map(x=>esc(x.id)+' ('+x.days+' dias)').join(', ')} · ${e.cobrado!=null?'cobrado por mês: <b>'+fmt.brl(e.cobrado,2)+'</b> ('+fmt.brl(e.valorUnit,2)+' por unidade)':'valor cobrado: <b>sem dado</b> (contrato não cadastrado no Estudo)'}</div>`:'';
  return`<div class="card rise sec">${cardH('Pendências de implantação',`${e.semTransmitir.length} sem transmitir · ${e.semSinalAgora.length} sem sinal agora · ${e.semBase.length} sem base de comparação · ${e.semMotorista.length} sem motorista identificado`,eqInfo())}${rows?`<div class="scroll-x"><table class="tbl"><thead><tr><th>Veículo</th><th>Motivo</th><th class="r">Há quanto tempo</th></tr></thead><tbody>${rows}</tbody></table></div>`:'<div class="empty"><b>Nenhuma pendência</b></div>'}${cob}</div>`;
}
