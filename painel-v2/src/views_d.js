/* =====================================================================
   TELAS COM GRÁFICOS · Motoristas (resumo, comportamento, comparativo,
   semanal, painel do condutor, mensagens) e Plano de ação
   ===================================================================== */
const spanCmp=C=>`${C.P.label} · comparado a ${C.P.cmpLabel}`;

/* ---------- Evolução da frota/condutor por granularidade ---------- */
function evoBlock(C,id){
  const u=S.u,gran=u.evoGran||'semana',mk=u.evoMetric||'kmpl';const ind=INDK[mk];
  const bs=buckets(C,gran,id);const pv=bs.map(b=>b.a?ind.get(b.a):null);
  const dc=pv.map((v,i)=>{if(v==null)return'var(--n300)';let j=i-1;while(j>=0&&pv[j]==null)j--;if(j<0)return'var(--n400)';const g=indGood(ind,(v-pv[j])/Math.abs(pv[j]||1)*100);return g==null?'var(--n400)':g?'var(--ok)':'var(--bad)'});
  const chart=lineChart({labels:bs.map(b=>b.label),series:[{vals:pv,color:'var(--laranja)',area:true,dots:true,dotColors:dc}],yfmt:v=>fmt.n(v,ind.d),h:230,
    tip:i=>bs[i].a?`<b>${esc(bs[i].label)}</b><br>${esc(ind.l)}: ${fInd(ind,pv[i])}<br>${fmt.n(bs[i].a.km,0)} km`:'',aria:'Evolução de '+ind.l});
  const rows=IND.filter(i=>!(i.veh&&!id)&&i.k!=='n'||id&&i.k==='n').slice(0,10);
  const cell=(i,b,k)=>{if(!b.a)return'<td class="r muted">—</td>';const v=i.get(b.a);if(v==null)return'<td class="r muted">—</td>';let j=k-1;while(j>=0&&!bs[j].a)j--;let cls='',dv='';
    if(j>=0){const pvv=i.get(bs[j].a);if(pvv){const chg=(v-pvv)/Math.abs(pvv)*100,g=indGood(i,chg);if(Math.abs(chg)>=2&&g!=null)cls=g?'cg':'cb';dv=`<span class="evd">${chg>0?'▲':chg<0?'▼':''} ${fmt.n(Math.abs(chg),1)}%</span>`}}
    return`<td class="r ${cls}"><div class="num evv">${fmt.n(v,i.d)}</div>${dv}</td>`};
  const unit=i=>{const u=(i.u||'').trim();return i.k==='lph100'?'L por 100 km':u.startsWith('/')?'por '+u.slice(1).trim():u==='%'?'% do tempo':u==='pts'?'pontos (0–100)':u};
  const tbl=`<div class="scroll-x"><table class="tbl evo"><thead><tr><th>Indicador</th><th>Unidade</th>${bs.map(b=>`<th class="r">${esc(b.label)}</th>`).join('')}</tr></thead><tbody>${rows.map(i=>`<tr><td><b>${esc(i.l)}</b></td><td class="muted xs">${esc(unit(i))}</td>${bs.map((b,k)=>cell(i,b,k)).join('')}</tr>`).join('')}</tbody></table></div><div class="xs muted" style="margin-top:10px">Cada célula traz o valor do período e, abaixo, a variação em % contra a coluna anterior (▲ subiu, ▼ caiu). Verde melhorou e vermelho piorou (variações menores que 2% não são coloridas).</div>`;
  const opts=IND.filter(i=>!i.veh).map(i=>`<option value="${i.k}" ${i.k===mk?'selected':''}>${esc(i.l)}</option>`).join('');
  return`<div class="card sec rise">${cardH('Evolução '+({dia:'dia a dia',semana:'por semana',quinzena:'por quinzena',mes:'mês a mês'}[gran]),'Pontos verdes melhoraram contra o período anterior; vermelhos pioraram.',`<select class="inp" data-chg="evo-metric" style="width:auto" aria-label="Indicador">${opts}</select>${segHtml([['dia','Dia'],['semana','Semana'],['quinzena','Quinzena'],['mes','Mês']],gran,'evo-gran')}`)}${chart}<div class="leg" style="margin:8px 0 14px"><span><i style="background:var(--ok);border-radius:99px"></i>Melhorou</span><span><i style="background:var(--bad);border-radius:99px"></i>Piorou</span><span><i style="background:var(--n400);border-radius:99px"></i>Primeiro período / estável</span></div>${tbl}</div>`;
}
function improvedRows(C,rows,keys){
  return keys.map(k=>{const i=INDK[k];let b=0,s=0,w=0;rows.forEach(r=>{const x=indRes(i,r.c,r.p);if(x.k==='better')b++;else if(x.k==='worse')w++;else if(x.k==='same')s++});
    const fl=indRes(i,C.cur,C.prev);return{l:i.l,sub:`${fInd(i,i.get(C.cur))} na frota`,b,s,w,dl:fl.ch!=null?dl(fl.ch,i.up===true,1):'—'}});
}

/* ---------- MOTORISTAS · Resumo ---------- */
function drvResumo(C){
  const rows=drvRows(C);const withBase=rows.filter(r=>r.st.k!=='base');
  const up=rows.filter(r=>r.st.k==='up').length,down=rows.filter(r=>r.st.k==='down').length,flat=rows.filter(r=>r.st.k==='flat').length;
  const gl=gainsLosses(C.cur,C.prev,C.price);const net=gl.reduce((a,g)=>a+g.money,0);
  const kpi=(t,v,s,c='',y='')=>`<div class="card kpi lift rise ${y}"><div class="tag">${t}</div><div class="v" ${c?`style="color:${c}"`:''}>${v}</div><div class="xs muted">${s}</div></div>`;
  const kp=`<div class="grid g4">${kpi('Condutores no período',count(rows.length,0),`${withBase.length} com base de comparação`)}${kpi('Evoluíram',count(up,0),'consumo subiu mais de 2%','var(--ok-t)')}${kpi('Regrediram',count(down,0),`${flat} estáveis`,'var(--bad-t)')}${kpi(net>=0?'Economia no período':'Perda no período',count(Math.abs(net),0,'R$ '),(net>=0?'diesel a menos':'diesel a mais')+' nos mesmos km','','yel')}</div>`;
  const pts=withBase.map(r=>({id:r.v.id,x:r.c.kmpl,y:r.st.ch,tip:`<b>${esc(r.v.driver)}</b> · ${esc(r.v.id)}<br>${fmt.kmpl(r.c.kmpl)} km/l · ${fmt.sp(r.st.ch)} vs comparação`}));
  const sc=scatter(pts,{h:470,xMid:C.cur.kmpl||1,yMid:0,xl:'Média de consumo no período (km/l)',yl:'Variação contra a comparação',q:['Bom e evoluindo','Abaixo da média, mas melhorando','Bom, mas piorando','Abaixo da média e piorando'],aria:'Onde cada condutor está'});
  const top=[...gl.slice(0,6),...gl.slice(-6).filter(g=>g.money<0)].filter((g,i,a)=>a.findIndex(x=>x.id===g.id)===i);
  const dv=diverging(top.map(g=>({id:g.id,l:C.fleet.byId[g.id].driver,s:g.id,v:g.money})));
  const sumG=gl.filter(g=>g.money>0).reduce((a,g)=>a+g.money,0),sumL=gl.filter(g=>g.money<0).reduce((a,g)=>a+g.money,0);
  /* pódio */
  const pod=rows.filter(r=>r.c.score!=null).sort((a,b)=>b.c.score-a.c.score).slice(0,3);const md=['#e0b100','#a9a9a3','#b5763a'];
  const podium=pod.length?`<div class="podium">${[1,0,2].map(i=>{const r=pod[i];if(!r)return'<div></div>';return`<div class="pd p${i+1}" data-act="drv-open" data-id="${esc(r.v.id)}" style="--i:${i}"><svg class="medal" width="38" height="38" viewBox="0 0 40 40" aria-hidden="true"><path d="M12 2h6l4 12h-6zM28 2h-6l-4 12h6z" fill="${md[i]}" opacity=".5"/><circle cx="20" cy="25" r="11" fill="${md[i]}"/><text x="20" y="30" text-anchor="middle" style="fill:#fff;font:800 14px var(--num)">${i+1}</text></svg><b>${esc(r.v.driver)}</b><span class="xs muted">${esc(r.v.id)}</span><span class="num b" style="font-size:20px">${fmt.n(r.c.score,0)}</span><span class="tag">nota</span><div class="pdb"></div></div>`}).join('')}</div>`:'<div class="empty"><b>Sem nota no período</b></div>';
  const imp=improvedRows(C,rows,['kmpl','br100','green','idle','score']);
  return kp+`<div class="grid sec" style="grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr))"><div class="card rise">${cardH('Onde cada condutor está','Cada ponto é um condutor. Clique para abrir. As linhas tracejadas são a média da frota e a variação zero.')}${sc}</div>
  <div class="card rise">${cardH('Quem ganhou e quem perdeu',`Em diesel, nos mesmos km · ganhos ${fmt.brl(sumG)} · perdas ${fmt.brl(Math.abs(sumL))}`)}${dv}</div></div>
  <div class="grid g-main sec"><div class="card rise">${cardH('Em que os condutores melhoraram','Quantos melhoraram, ficaram estáveis ou pioraram em cada indicador.')}${improvedStack(imp)}<div class="leg" style="margin-top:10px"><span><i style="background:var(--ok)"></i>Melhoraram</span><span><i style="background:var(--n300)"></i>Estáveis</span><span><i style="background:var(--bad)"></i>Pioraram</span></div></div>
  <div class="card dark hero rise"><span class="wm" aria-hidden="true">gobrax</span>${cardH('Pódio da nota Gobrax',C.P.label)}${podium}</div></div>`+evoBlock(C,null);
}

/* ---------- MOTORISTAS · Comportamento ---------- */
function drvBehavior(C){
  const ks=['br100','brh100','speed','green','idle','score'];
  const ds=dailySeries(C.cid,ST,C.P.from,C.P.to);const pd=dailySeries(C.cid,ST,C.P.cf,C.P.ct);
  const cards=ks.map(k=>{const i=INDK[k];const cv=ds.map(d=>d.a?i.get(d.a):null),pv=ds.map((d,x)=>pd[x]&&pd[x].a?i.get(pd[x].a):null);const r=indRes(i,C.cur,C.prev);
    const ch=lineChart({labels:ds.map(d=>fmt.dm(d.date)),series:[{vals:pv,color:'var(--text)',dash:'4 4',w:1.6},{vals:cv,color:'var(--laranja)',area:true,w:2.6}],w:360,h:170,pl:36,yfmt:v=>fmt.n(v,i.d>1?1:i.d),
      tip:x=>cv[x]!=null?`<b>${fmt.dm(ds[x].date)}</b><br>${esc(i.l)}: ${fInd(i,cv[x])}${pv[x]!=null?'<br>Comparação: '+fInd(i,pv[x]):''}`:'',aria:i.l});
    return`<div class="card rise"><div class="row" style="justify-content:space-between;flex-wrap:nowrap;align-items:flex-start"><div><div class="tag">${esc(i.l)}</div><div class="num" style="font-size:26px;font-weight:800">${fInd(i,i.get(C.cur))}</div></div><span class="badge ${r.cls}">${r.label}</span></div><div class="xs muted" style="margin-bottom:4px">${r.ch!=null?dl(r.ch,i.up===true,1)+' contra a comparação':'sem base de comparação'}</div>${ch}</div>`}).join('');
  const ax=radar(radarAxes(C.cur,C.prev,C.cur));
  const cal=dotCal(C.P.from,C.P.to,t=>{const d=ds[t-C.P.from];return d&&d.a?{score:d.a.score,extra:fmt.n(d.a.km,0)+' km · '+d.a.n+' veículos'}:null});
  return`<div class="grid g3">${cards}</div><div class="grid g2 sec"><div class="card rise">${cardH('Perfil de condução da frota','Amarelo: agora. Azul: antes. Cinza: média. Quanto mais longe do centro, melhor.')}${ax}</div><div class="card rise">${cardH('Calendário de desempenho','Cada ponto é um dia, colorido pela nota da frota. Passe o mouse para ver o valor.')}${cal}</div></div>`;
}

/* ---------- MOTORISTAS · Comparativo ---------- */
function drvCompare2(C){
  const u=S.u;const M=['kmpl','score','br100','green','idle','kmDay'].includes(u.cmpMetric)?u.cmpMetric:'kmpl';const i=INDK[M];
  const rs=drvRows(C).filter(r=>i.get(r.c)!=null).map(r=>({...r,val:i.get(r.c),pv:r.p?i.get(r.p):null}));
  rs.sort((a,b)=>i.up?b.val-a.val:a.val-b.val);const avg=i.get(C.cur);
  const rows=rs.map(r=>({id:r.v.id,name:r.v.driver,val:r.val,prev:r.pv,ch:r.pv?(r.val-r.pv)/Math.abs(r.pv)*100:null}));
  const opts=[['kmpl','Consumo'],['score','Nota'],['br100','Freadas'],['green','Faixa verde'],['idle','Ociosidade'],['kmDay','Km por dia']];
  const imp=improvedRows(C,drvRows(C),['kmpl','br100','green','idle','score']);
  return`<div class="card rise sec">${cardH('Comparativo entre condutores',`Do melhor para o pior em ${i.l.toLowerCase()}. Barra verde: melhor que a média da frota; vermelha: pior. Marca azul: como estava antes. Linha tracejada: média da frota (${fInd(i,avg)}).`,segHtml(opts,M,'cmp-metric'))}${compRows(rows,{avg,up:i.up===true,d:i.d,u:i.u})}</div>
  <div class="card rise sec">${cardH('Em que os condutores melhoraram','Quantos melhoraram, ficaram estáveis ou pioraram em cada indicador. À direita, a variação da frota.')}${improvedStack(imp)}</div>`;
}

/* ---------- MOTORISTAS · Resultado semanal (cards) ---------- */
function weekMsg(name,r,w){return`Olá${/^Sem motorista/.test(name)?'':', '+name.split(' ')[0]}! Seu resultado da semana ${w.n}: média de ${fmt.kmpl(r.a.kmpl)} km/l${r.ch!=null?` (${fmt.sp(r.ch)} contra a semana anterior)`:''}, ${fmt.n(r.a.km,0)} km rodados e ${r.good} dia(s) bom(ns). Qualquer dúvida, é só chamar.`}
function drvWeekly2(C){
  const wk=weeklySeries(C.cid,ST,C.P.from,C.P.to);if(!wk.length)return'<div class="card empty"><b>Sem semanas no período</b></div>';
  if(S.u.weekSel==null||S.u.weekSel>wk.length)S.u.weekSel=wk.length;const w=wk[S.u.weekSel-1],pw=wk[S.u.weekSel-2];
  const fleetAvg=w.a.kmpl;
  const rows=Object.values(w.a.byVeh).filter(a=>a.km>0&&C.fleet.byId[a.id]).map(a=>{const pa=pw?pw.a.byVeh[a.id]:null;const ch=pa&&pa.kmpl?(a.kmpl-pa.kmpl)/pa.kmpl*100:null;
    const days=[];for(let t=w.from;t<=w.to;t++){const r=aggFor(C,t,t,a.id);days.push({t,a:r.n?r:null})}
    const good=days.filter(d=>d.a&&d.a.kmpl>=fleetAvg).length,att=days.filter(d=>d.a&&d.a.kmpl<fleetAvg).length;
    let evo=0,tot=0;['kmpl','br100','green','idle'].forEach(k=>{if(!pa)return;const r=indRes(INDK[k],a,pa);if(r.k!=='base'){tot++;if(r.k==='better')evo++}});
    return{a,v:C.fleet.byId[a.id],pa,ch,days,good,att,evo,tot}}).sort((x,y)=>(y.a.kmpl||0)-(x.a.kmpl||0));
  const n=rows.length;const up=rows.filter(r=>r.ch!=null&&r.ch>2).length,dn=rows.filter(r=>r.ch!=null&&r.ch<-2).length;const goodT=rows.reduce((s,r)=>s+r.good,0),attT=rows.reduce((s,r)=>s+r.att,0);
  const sum=`<div class="grid g4" style="margin-top:14px">${[['Média de consumo da frota',fmt.kmpl(fleetAvg)+' km/l',pw?`antes ${fmt.kmpl(pw.a.kmpl)} · ${dl(pc(fleetAvg,pw.a.kmpl),true)}`:'sem semana anterior'],['Condutores',String(n),`<span style="color:var(--ok-t)">${up} evoluíram</span> · <span style="color:var(--bad-t)">${dn} regrediram</span> · ${n-up-dn} estáveis`],['Dias bons',`<span style="color:var(--ok-t)">${goodT}</span>`,`${attT} dias de atenção · km/l do dia contra a média da frota`],['Km rodados',fmt.n(w.a.km,0)+' km',pw?`antes ${fmt.n(pw.a.km,0)} km · ${dl(pc(w.a.km,pw.a.km),true)}`:'']].map(([t,v,s])=>`<div class="kpi flat"><div class="tag">${t}</div><div class="v">${v}</div><div class="xs muted">${s}</div></div>`).join('')}</div>`;
  const wl=['seg','ter','qua','qui','sex','sáb','dom'];
  const cards=rows.map((r,i)=>{const st=r.ch==null?{l:'SEM BASE',c:'n'}:r.ch>2?{l:'EVOLUIU',c:'ok'}:r.ch<-2?{l:'REGREDIU',c:'bad'}:{l:'ESTÁVEL',c:'n'};
    const circles=wl.map((l,k)=>{const t=w.from+(k-((wdOf(w.from)+6)%7));const d=r.days.find(x=>x.t===t);const cl=!d||!d.a?'e':d.a.kmpl>=fleetAvg?'g':'y';return`<div class="wdc"><span>${l}</span><i class="${cl}" ${d&&d.a?`data-tt="<b>${fmt.dm(dIso(t))}</b><br>${fmt.kmpl(d.a.kmpl)} km/l · ${fmt.n(d.a.km,0)} km"`:''}>${d&&d.a?fmt.kmpl(d.a.kmpl):''}</i></div>`}).join('');
    const chips=wk.map(x=>{const q=x.a.byVeh[r.a.id];return`<div class="wch ${x.n===w.n?'cur':''} ${pw&&x.n===pw.n?'cmp':''}"><span>S${x.n}</span><b class="num">${q&&q.kmpl?fmt.kmpl(q.kmpl):'–'}</b></div>`}).join('');
    const kmCh=r.pa?pc(r.a.km,r.pa.km):null;
    return`<div class="card wcard rise" style="--i:${Math.min(i,10)}"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><div class="row" style="flex-wrap:nowrap;gap:8px"><span class="ava xs">${esc(r.v.driver[0])}</span><div><b style="display:block;line-height:1.2">${esc(r.v.id)}</b><span class="xs muted">${esc(r.v.driver)}</span></div></div><span class="badge ${st.c}">${st.l}</span></div>
    <div class="wbig"><div class="tag">Média de consumo atual</div><div class="row" style="justify-content:space-between;flex-wrap:nowrap;align-items:flex-end"><span class="num" style="font-size:30px;font-weight:800">${fmt.kmpl(r.a.kmpl)} <small style="font-size:14px">km/l</small></span><span class="xs b">${i+1}º de ${n}</span></div><div class="xs">${r.ch!=null?dl(r.ch,true,1)+` <span class="muted">antes ${fmt.kmpl(r.pa.kmpl)} km/l</span>`:'<span class="muted">sem semana anterior</span>'}</div></div>
    <div class="xs muted" style="margin:8px 0">${r.tot?`Evoluiu em <b style="color:var(--text)">${r.evo}</b> de ${r.tot} indicadores`:'Sem indicadores para comparar'}</div>
    <div class="xs"><span class="muted">Km rodados</span><br><span class="muted">${r.pa?fmt.n(r.pa.km,0)+' km → ':''}</span><b>${fmt.n(r.a.km,0)} km</b> ${kmCh!=null?dl(kmCh,true,1):''}</div>
    <div class="row" style="justify-content:space-between;margin-top:10px"><span class="tag">Dias da semana</span><span class="xs b"><span style="color:var(--ok-t)">${r.good} bons</span> · <span style="color:var(--warn-t)">${r.att} de atenção</span></span></div><div class="wdays">${circles}</div>
    <div class="tag" style="margin:10px 0 6px">Média de consumo nas semanas</div><div class="wchips">${chips}</div>
    <div class="row" style="margin-top:12px;flex-wrap:nowrap"><button class="btn primary sm" style="flex:1" data-act="wk-wa" data-id="${esc(r.v.id)}" data-w="${w.n}">${ico('msg',14)} WhatsApp</button><button class="btn ghost sm icon" data-act="drv-open" data-id="${esc(r.v.id)}" aria-label="Abrir condutor">${ico('user',14)}</button><button class="btn ghost sm icon" data-act="wk-copy" data-id="${esc(r.v.id)}" data-w="${w.n}" aria-label="Copiar mensagem">${ico('copy',14)}</button></div></div>`}).join('');
  return`<div class="card rise sec">${cardH(`Resumo da frota na semana ${w.n}`,`${fmt.dm(dIso(w.from))} a ${fmt.dm(dIso(w.to))}${pw?` · comparada à semana ${pw.n}`:''}${w.partial?' (semana parcial)':''}`,segHtml(wk.map(x=>[String(x.n),'Sem '+x.n]),String(S.u.weekSel),'week-sel')+`<button class="btn ghost sm" data-act="csv" data-w="drivers">${ico('download',14)} Planilha</button>`)}${sum}</div><div class="wgrid sec">${cards}</div>`;
}

/* ---------- Painel do condutor ---------- */
function drvDash(C,id){
  const u=S.u;const c=C.cur.byVeh[id],p=C.prev.byVeh[id];const gl=gainsLosses(C.cur,C.prev,C.price).find(x=>x.id===id);
  const kp=(t,val,cur,prev,up,d)=>`<div class="kpl"><div class="tag">${t}</div><div class="num kv">${val}</div><div class="xs">${prev!=null&&prev?dl(pc(cur,prev),up,d>1?1:d)+` <span class="muted">vs. ${fmt.dm(dIso(C.P.cf))} a ${fmt.dm(dIso(C.P.ct))}</span>`:'<span class="muted">sem comparação</span>'}</div></div>`;
  const strip=`<div class="card rise kpstrip sec"><div class="gblock"><span class="wm" aria-hidden="true">go</span>${gaugeNota(c.score,'NOTA GOBRAX',true)}</div>${kp('Km rodados',fmt.n(c.km,0)+' km',c.km,p&&p.km,true,0)}${kp('Média de consumo',fmt.kmpl(c.kmpl)+' km/l',c.kmpl,p&&p.kmpl,true,2)}${kp('Diesel consumido',fmt.n(c.l,0)+' L',c.l,p&&p.l,null,0)}${kp('Dias trabalhados',c.n+' dias',c.n,p&&p.n,true,0)}${kp('Km por dia',fmt.n(c.kmDay,0)+' km',c.kmDay,p&&p.kmDay,true,0)}</div>`;
  /* leitura automática */
  const reads=[];
  if(gl)reads.push({ok:gl.money>=0,t:gl.money>=0?`Gastou <b>${fmt.brl(gl.money)} a menos</b> em diesel (${fmt.n(gl.liters,0)} L) que na comparação, para os mesmos km.`:`Gastou <b>${fmt.brl(Math.abs(gl.money))} a mais</b> em diesel (${fmt.n(Math.abs(gl.liters),0)} L) que na comparação, para os mesmos km.`});
  const sc=IND.filter(i=>['br100','green','idle','score','kmpl'].includes(i.k)).map(i=>({i,v:i.get(c),f:i.get(C.cur)})).filter(x=>x.v!=null&&x.f);
  let best=null,worst=null;sc.forEach(x=>{const q=x.i.up?x.v/x.f:x.f/(x.v||1e-9);if(!best||q>best.q)best={...x,q};if(!worst||q<worst.q)worst={...x,q}});
  if(best&&best.q>1.03)reads.push({ok:true,t:`Ponto forte: <b>${esc(best.i.l)}</b> em ${fInd(best.i,best.v)}, melhor que ${fInd(best.i,best.f)} da frota.`});
  if(worst&&worst.q<.97)reads.push({ok:false,t:`Ponto de atenção: <b>${esc(worst.i.l)}</b> em ${fInd(worst.i,worst.v)}, pior que ${fInd(worst.i,worst.f)} da frota.`});
  const lido=`<div class="card rise sec"><div class="row" style="margin-bottom:10px"><i class="legend-dot" style="background:var(--laranja)"></i><b style="font-size:16px">Leitura automática do comportamento</b><span class="tag">${c.n} registros no período</span></div><div class="grid g2" style="gap:10px">${reads.map(r=>`<div class="read ${r.ok?'ok':'bad'}"><span class="ri">${ico(r.ok?'check':'alert',14)}</span><span>${r.t}</span></div>`).join('')||'<div class="muted">Sem leitura disponível.</div>'}</div></div>`;
  /* radar + antes/agora/frota */
  const bars=['br100','kmpl','green','idle','score'].map(k=>{const i=INDK[k];const a=i.get(c),b=p?i.get(p):null,f=i.get(C.cur);if(a==null)return'';const mx=Math.max(a||0,b||0,f||0)*1.15||1;const good=f==null?true:i.up?a>=f:a<=f;
    return`<div class="abf"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><b>${esc(i.l)}</b><span class="xs"><b class="num">${fInd(i,a)}</b> <span class="muted">antes ${fInd(i,b)} · frota ${fInd(i,f)}</span></span></div><div class="cmt hatch"><div class="bfill" style="left:0;width:${(a||0)/mx*100}%;background:${good?'var(--ok)':'var(--bad)'}"></div>${b!=null?`<i class="before" style="left:${b/mx*100}%"></i>`:''}${f!=null?`<i class="avgl" style="left:${f/mx*100}%"></i>`:''}</div><div class="xs">${b?dl(pc(a,b),i.up===true,1)+' <span class="muted">vs. comparação</span>':''}</div></div>`}).join('');
  const rd=`<div class="grid ga sec"><div class="card rise">${cardH('Perfil de condução','Amarelo: este condutor agora. Azul: como estava. Cinza: média da frota. Quanto mais longe do centro, melhor.')}${radar(radarAxes(c,p,C.cur))}</div><div class="card rise">${cardH('Antes, agora e frota','Barra: agora; verde se melhor que a frota, vermelha se pior. Marca azul: antes. Linha tracejada: média da frota.')}<div class="stack" style="gap:14px">${bars}</div></div></div>`;
  /* calendário */
  const ds=dailySeries(C.cid,ST,C.P.from,C.P.to,x=>x===id);
  const cal=`<div class="card rise sec">${cardH('Calendário de desempenho','Cada ponto é um dia do período, colorido pela nota. Passe o mouse para ver o valor.')}${dotCal(C.P.from,C.P.to,t=>{const d=ds[t-C.P.from];return d&&d.a?{score:d.a.score,extra:fmt.n(d.a.km,0)+' km · '+fmt.n(d.a.kmpl,2)+' km/l'}:null})}</div>`;
  /* evolução dia a dia + km por dia */
  const mk=u.dayMetric||'br100',ind=INDK[mk];const pds=dailySeries(C.cid,ST,C.P.cf,C.P.ct,x=>x===id);
  const cv=ds.map(d=>d.a?ind.get(d.a):null),pv=ds.map((d,k)=>pds[k]&&pds[k].a?ind.get(pds[k].a):null);
  const view=u.dayView==='tabela'?`<div class="scroll-x" style="max-height:320px"><table class="tbl"><thead><tr><th>Dia</th><th class="r">${esc(ind.l)}</th><th class="r">Km</th></tr></thead><tbody>${ds.filter(d=>d.a).map(d=>`<tr><td>${fmt.dmy(d.date)}</td><td class="r num b">${fInd(ind,ind.get(d.a))}</td><td class="r num">${fmt.n(d.a.km,0)}</td></tr>`).join('')}</tbody></table></div>`
    :lineChart({labels:ds.map(d=>fmt.dm(d.date)),series:[{vals:pv,color:'var(--text)',dash:'4 4',w:1.6},{vals:cv,color:'var(--laranja)',area:true,dots:ds.length<=14,w:2.8}],yfmt:x=>fmt.n(x,ind.d),h:240,tip:k=>cv[k]!=null?`<b>${fmt.dm(ds[k].date)}</b><br>${esc(ind.l)}: ${fInd(ind,cv[k])}${pv[k]!=null?'<br>Comparação: '+fInd(ind,pv[k]):''}`:'',aria:'Evolução dia a dia'});
  const opts=IND.filter(i=>!i.veh).map(i=>`<option value="${i.k}" ${i.k===mk?'selected':''}>${esc(i.l)}</option>`).join('');
  const kmv=ds.map(d=>d.a?d.a.km:null);const kmAvg=c.n?c.km/c.n:null;
  const evo=`<div class="grid ga2 sec"><div class="card rise">${cardH('Evolução dia a dia',`${esc(ind.s)}. Passe o mouse para comparar dia a dia.`,`<select class="inp" data-chg="day-metric" style="width:auto" aria-label="Indicador">${opts}</select>${segHtml([['grafico','Gráfico'],['tabela','Tabela']],u.dayView||'grafico','day-view')}`)}<div class="leg" style="margin-bottom:6px"><span><i style="background:var(--laranja)"></i>Período analisado</span><span><i style="background:var(--n400)"></i>Comparação</span></div>${view}</div>
  <div class="card rise">${cardH('Km rodados por dia','Dias sem barra não têm registro.')}${dailyBars(kmv,ds.map(d=>fmt.dm(d.date)),{avg:kmAvg,u:' km',aria:'Km rodados por dia'})}</div></div>`;
  return strip+lido+rd+cal+evo;
}
function drvIndTable(C,id){
  const c=C.cur.byVeh[id],p=C.prev.byVeh[id];let h='',g='';
  IND.forEach(i=>{if(i.g!==g){g=i.g;h+=`<tr class="grp"><td colspan="6">${esc(g)}</td></tr>`}
    const a=i.get(c),b=p?i.get(p):null,f=i.get(C.cur);const r=indRes(i,c,p);
    h+=`<tr><td><b>${esc(i.l)}</b><div class="xs muted">${esc(i.s)}</div></td><td class="r num b">${fInd(i,a)}</td><td class="r num muted">${fInd(i,b)}</td><td class="r">${r.ch!=null?dl(r.ch,i.up===true,1):'<span class="muted">—</span>'}</td><td class="r"><span class="badge ${r.cls}">${r.label}</span></td><td class="r num muted">${fInd(i,f)}</td></tr>`});
  return`<div class="card sec rise">${cardH('Todos os indicadores',`${C.P.label} comparado a ${C.P.cmpLabel}`,`<button class="btn ghost sm" data-act="ind-csv" data-id="${esc(id)}">${ico('download',14)} Exportar CSV</button>`)}<div class="scroll-x"><table class="tbl"><thead><tr><th>Indicador</th><th class="r">Período analisado</th><th class="r">Comparação</th><th class="r">Variação</th><th class="r">Resultado</th><th class="r">Média da frota</th></tr></thead><tbody>${h}</tbody></table></div></div>`;
}

/* ---------- PLANO DE AÇÃO ---------- */
function vPlan(C){
  const {cur,prev,price}=C;const op=opportunities(cur,prev,price);const gl=gainsLosses(cur,prev,price);
  const losers=gl.filter(g=>g.money<0).slice(-5).reverse();
  /* gargalos: meta x atual */
  const g=[
    {l:'Média de consumo',v:cur.kmpl,m:op.ref,up:true,d:2,u:' km/l',n:'meta: 75º percentil da frota'},
    {l:'Freadas a cada 100 km',v:cur.br100,m:op.brRef,up:false,d:1,u:'',n:'meta: melhor quartil da frota'},
    {l:'Faixa verde',v:cur.green,m:60,up:true,d:1,u:'%',n:'meta: 60% do tempo'},
    {l:'Motor ocioso',v:cur.idle,m:10,up:false,d:1,u:'%',n:'meta: até 10%'},
    {l:'Nota Gobrax',v:cur.score,m:70,up:true,d:0,u:' pts',n:'meta: faixa Bom (70+)'},
    {l:'Km por dia trabalhado',v:cur.kmDay,m:op.kdRef?op.kdRef*.8:null,up:true,d:0,u:' km',n:'meta: 80% da referência'}
  ].filter(x=>x.v!=null&&x.m);
  const gar=g.map((x,i)=>{const prog=clamp(x.up?x.v/x.m:x.m/(x.v||1e-9),0,1.3);const ok=prog>=1;const w=Math.min(prog,1.3)/1.3*100;
    return`<div class="gar"><div class="row" style="justify-content:space-between;flex-wrap:nowrap;gap:8px"><b>${x.l}</b><span class="xs" style="white-space:nowrap"><b class="num">${fmt.n(x.v,x.d)}${x.u}</b> <span class="muted">/ meta ${fmt.n(x.m,x.d)}${x.u}</span></span></div><div class="cmt"><div class="bfill" style="--i:${i};left:0;width:${w}%;background:${ok?'var(--ok)':prog>.85?'var(--warn)':'var(--bad)'}"></div><i class="avgl" style="left:${100/1.3}%"></i></div><div class="xs muted"><span style="color:${ok?'var(--ok-t)':'inherit'};font-weight:600">${ok?'Meta atingida':`Faltam ${fmt.n(Math.abs(x.m-x.v),x.d)}${x.u}`}</span> · ${x.n}</div></div>`}).join('')||'<div class="empty"><b>Sem dados para as metas</b></div>';
  /* onde trabalhar primeiro */
  const pr=[];
  if(op.worse[0])pr.push({t:'Consumo acima da referência',n:op.worse.slice(0,3).map(w=>w.id).join(', '),v:fmt.brl(op.potential),s:`${op.worse.length} veículos abaixo de ${fmt.n(op.ref,2)} km/l`,tone:'bad',f:'down'});
  if(op.fell[0])pr.push({t:'Condutores que pioraram',n:op.fell.slice(0,3).map(w=>w.id).join(', '),v:`${op.fell.length}`,s:`${op.fell[0].id} caiu de ${fmt.n(op.fell[0].from,2)} para ${fmt.n(op.fell[0].to,2)} km/l`,tone:'bad',f:'down'});
  if(op.low[0])pr.push({t:'Pouco rendimento por dia',n:op.low.slice(0,3).map(w=>w.id).join(', '),v:`${op.low.length}`,s:`Menos de 80% de ${fmt.n(op.kdRef,0)} km por dia`,tone:'warn',f:'all'});
  if(op.brW)pr.push({t:'Freadas acima do melhor quartil',n:'',v:`${op.brW}`,s:`veículos acima de ${fmt.n(op.brRef,1)} /100 km`,tone:'warn',f:'all'});
  const prio=pr.map((x,i)=>`<div class="prio ${x.tone}" style="--i:${i}"><span class="pn num">${i+1}</span><div style="flex:1;min-width:0"><b>${x.t}</b><div class="xs muted">${x.s}${x.n?' · '+esc(x.n):''}</div></div><div class="num b" style="font-size:20px;white-space:nowrap">${x.v}</div><button class="btn primary sm" data-act="go-drv" data-f="${x.f||'all'}">Ver condutores</button></div>`).join('')||'<div class="empty"><b>Nada crítico agora</b>Continue acompanhando.</div>';
  const li=(g,pos)=>`<div class="li" data-act="drv-open" data-id="${esc(g.id)}"><div class="ava">${esc(C.fleet.byId[g.id].driver[0])}</div><div style="flex:1;min-width:0"><b>${esc(C.fleet.byId[g.id].driver)}</b><div class="xs muted">${esc(g.id)} · ${fmt.kmpl(g.from)} → ${fmt.kmpl(g.to)} km/l</div></div><span class="num b" style="color:${pos?'var(--ok-t)':'var(--bad-t)'}">${pos?'+':''}${fmt.brl(g.money)}</span></div>`;
  const win=gl.filter(g=>g.money>0).slice(0,5).map(g=>li(g,true)).join('')||'<div class="empty"><b>Ninguém com ganho relevante</b></div>';
  const lose=losers.map(g=>li(g,false)).join('')||'<div class="empty"><b>Ninguém com perda relevante</b></div>';
  return`${pageH('Plano de ação',`${esc(C.def.name)} · ${spanCmp(C)}`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  <div class="grid g-main" style="align-items:start"><div class="card rise">${cardH('Onde trabalhar primeiro','Ordenado pelo dinheiro e pelo número de condutores envolvidos.')}<div class="stack" style="gap:10px">${prio}</div></div><div class="card rise">${cardH('Gargalos operacionais','Atual contra a meta. A linha marca a meta.')}<div class="stack" style="gap:14px">${gar}</div></div></div>
  <div class="grid g2 sec"><div class="card rise">${cardH('Maiores perdas em diesel','Condutores que mais gastaram a mais nos mesmos km.')}<div class="stack" style="gap:2px">${lose}</div></div><div class="card rise">${cardH('Maiores ganhos em diesel','Condutores que mais economizaram nos mesmos km.')}<div class="stack" style="gap:2px">${win}</div></div></div>`;
}

/* ---------- Visão geral · blocos extras (nota, perfil, calendário, evolução) ---------- */
function ovExtra(C){
  const ds=dailySeries(C.cid,ST,C.P.from,C.P.to);
  const cal=dotCal(C.P.from,C.P.to,t=>{const d=ds[t-C.P.from];return d&&d.a?{score:d.a.score,extra:fmt.n(d.a.km,0)+' km'}:null});
  const rs=indRes(INDK.score,C.cur,C.prev);const sb=scoreBand(C.cur.score);
  return`<div class="grid g3 sec"><div class="card yel rise" style="text-align:center"><span class="wm" aria-hidden="true">gobrax</span>${cardH('Nota Gobrax da frota',C.P.label)}${gaugeNota(C.cur.score,'NOTA GOBRAX',true)}<div style="margin-top:6px">${rs.ch!=null?dl(rs.ch,true,1)+' <span class="xs muted">contra '+esc(C.P.cmpLabel)+'</span>':'<span class="xs muted">sem comparação</span>'}</div><div class="row" style="justify-content:center;margin-top:10px"><span class="badge onyel">${sb.label}</span></div></div>
  <div class="card rise">${cardH('Perfil de condução','Agora contra o período anterior')}${radar(radarAxes(C.cur,C.prev,C.cur))}</div>
  <div class="card rise">${cardH('Calendário de desempenho','Nota da frota em cada dia')}${cal}</div></div>${evoBlock(C,null)}`;
}

/* ---------- MOTORISTAS · Mensagens (evolução da nota depois do aviso) ---------- */
const MSG_POINT={'Faixa verde':{get:r=>r.gr,up:true},'Motor ocioso':{get:r=>r.id,up:false},'Freadas':{get:r=>r.br!=null&&r.km?r.br/r.km*100:null,up:false}};
/* depois de um aviso: o ponto apontado melhorou nos 7 dias seguintes? */
function pointImproved(rs,t,point){
  const m=MSG_POINT[point];if(!m)return null;const avg=a=>{const v=a.map(m.get).filter(x=>x!=null&&isFinite(x));return v.length?v.reduce((s,x)=>s+x,0)/v.length:null};
  const a=avg(rs.filter(x=>x.t>t&&x.t<=t+7)),b=avg(rs.filter(x=>x.t>t-7&&x.t<=t));if(a==null||b==null)return null;return m.up?a>b:a<b;
}
/* nota do mês acumulada até cada dia (por veículo) e nota de cada mês */
function monthScores(rs){let mk='',km=0,scw=0;const mtd={},mon={};
  rs.forEach(r=>{const m=r.date.slice(0,7);if(m!==mk){mk=m;km=0;scw=0}if(r.sc!=null){km+=r.km;scw+=r.sc*r.km}const v=km?scw/km:null;mtd[r.t]=v;if(v!=null)mon[m]=v});return{mtd,mon}}
function msgHist(C){
  const by={};for(const r of validRecs(C.cid,ST))(by[r.veh]=by[r.veh]||[]).push(r);
  if(REAL.on){
    /* mensagens que o CS ou o robô marcaram como enviadas (/mensagens/enviadas) */
    const sent=REAL.msgs[C.cid];if(!sent)return[];const out=[];
    for(const s of sent){
      const plate=C.fleet.drvPlate[nkey(s.motorista)];const rs=(by[plate]||[]).slice().sort((a,b)=>a.t-b.t);const ms=monthScores(rs);
      const msgs=(s.envios||[]).map(e=>{const t=dIdx(String(e.dia).slice(0,10));let k=null;for(const r of rs){if(r.t<=t)k=r.t;else break}
        const point=e.foco_rotulo||null;return{t,score:nv(e.nota)??(k!=null?ms.mtd[k]:null),point,imp:pointImproved(rs,t,point)}}).sort((a,b)=>a.t-b.t);
      if(!msgs.length)continue;out.push({id:plate||s.motorista,name:s.motorista,msgs,first:msgs[0],last:msgs[msgs.length-1],mon:ms.mon});
    }
    return out.sort((a,b)=>b.msgs.length-a.msgs.length||a.name.localeCompare(b.name));
  }
  const key='mh|'+C.cid+'|'+DB.ver;if(DB.cache.has(key))return DB.cache.get(key);
  const f=C.fleet;const start=Math.min(...f.vehicles.map(v=>v.from));
  const FIRST=Math.max(T0-29,start+2),LAST=T0-3;
  const out=[];
  for(const v of f.vehicles){
    const rs=(by[v.id]||[]).slice().sort((a,b)=>a.t-b.t);if(rs.length<3)continue;
    const rnd=mulberry(hs(C.cid+v.id+'msg'));const p=.4+rnd()*.45;const ms=monthScores(rs);
    const msgs=[];
    rs.forEach((r,i)=>{
      if(r.t<FIRST||r.t>LAST)return;const roll=rnd();if(roll>p&&!(msgs.length===0&&r.t===FIRST))return;
      const pr=rs[i-1];let point=null;
      if(pr){const cand=[['Faixa verde',r.gr-pr.gr],['Motor ocioso',-(r.id-pr.id)],['Freadas',-(r.br/r.km*100-pr.br/pr.km*100)]].sort((a,b)=>a[1]-b[1]);point=cand[0][0]}
      msgs.push({t:r.t,score:ms.mtd[r.t],point,imp:pointImproved(rs,r.t,point)});
    });
    if(msgs.length<2)continue;
    out.push({id:v.id,name:v.driver,msgs,first:msgs[0],last:msgs[msgs.length-1],mon:ms.mon});
  }
  out.sort((a,b)=>b.msgs.length-a.msgs.length||a.name.localeCompare(b.name));
  DB.cache.set(key,out);return out;
}
function msgEvolution(C,hist){
  const u=S.u;const gran=u.msgGran||'dia';
  if(!hist.some(h=>h.id===u.msgDrv))u.msgDrv=hist[0].id;
  const h=hist.find(x=>x.id===u.msgDrv);
  const imps=h.msgs.filter(m=>m.imp!=null);const okn=imps.filter(m=>m.imp).length;
  const d=h.last.score!=null&&h.first.score!=null?h.last.score-h.first.score:null;
  const kpi=(c,t,v,s)=>`<div class="card kpi lift rise" style="border-top:3px solid ${c}"><div><div class="xs muted">${t}</div><div class="v">${v}</div><div class="xs muted" style="margin-top:4px">${s}</div></div></div>`;
  const kp=`<div class="grid g4">
    ${kpi('var(--info)','Mensagens recebidas',count(h.msgs.length,0),`de ${fmt.dmy(dIso(h.first.t))} a ${fmt.dmy(dIso(h.last.t))}`)}
    ${kpi('var(--n400)','Nota no 1º envio',fmt.n(h.first.score,1),'nota do mês naquele dia')}
    ${kpi('var(--amarelo)','Nota no último envio',fmt.n(h.last.score,1),`${dl(d,true,0,' pts')} vs 1º envio`)}
    ${kpi('var(--ok)','Melhorou no ponto apontado',imps.length?`${okn} <small>de ${imps.length}</small>`:'—','nos 7 dias depois de cada aviso')}</div>`;
  let chart,head,sub;
  if(gran==='dia'){
    const ls=h.msgs.map(m=>fmt.dm(dIso(m.t)));
    chart=lineChart({labels:ls,series:[{vals:h.msgs.map(m=>m.score),color:'var(--laranja)',area:true,dots:true}],w:520,h:290,pl:34,pr:26,vline:0,sub:h.msgs.map((m,i)=>i===0?'1º envio':''),vlabels:v=>fmt.n(v,1),yfmt:v=>fmt.n(v,0),
      tip:i=>{const m=h.msgs[i];return`<b>${fmt.dmy(dIso(m.t))}</b><br>Nota do mês neste dia: ${fmt.n(m.score,1)}<br>Ponto apontado: ${esc(m.point||'—')}<br>7 dias depois: ${m.imp==null?'sem dados':m.imp?'melhorou':'não melhorou'}`},aria:'Evolução da nota por dia'});
    sub='Nota do condutor em cada dia em que ele recebeu a análise (a nota do mês da Gobrax naquele dia). Os dias sem mensagem não aparecem. Passe o mouse num ponto para ver o dia e o ponto apontado.';
    head=`${fmt.dm(dIso(h.first.t))} ${fmt.n(h.first.score,1)} → <b class="big">${fmt.n(h.last.score,1)}</b> ${dl(d,true,1,' pts')} <span class="muted">em ${fmt.dmy(dIso(h.last.t))}</span>`;
  }else{
    const ms=Object.keys(h.mon).sort();const fm=dIso(h.first.t).slice(0,7);const cur=TODAY.slice(0,7);
    const lab=ms.map(m=>MESES[+m.slice(5)-1].replace(/^./,c=>c.toUpperCase())+'/'+m.slice(2,4));
    chart=lineChart({labels:lab,series:[{vals:ms.map(m=>h.mon[m]),color:'var(--laranja)',area:true,dots:true}],w:520,h:290,pl:34,pr:26,vline:ms.indexOf(fm),sub:ms.map(m=>m===fm?'1º envio':m===cur?'(parcial)':''),vlabels:v=>fmt.n(v,1),yfmt:v=>fmt.n(v,0),
      tip:i=>`<b>${lab[i]}</b><br>Nota do mês: ${fmt.n(h.mon[ms[i]],1)}${ms[i]===fm?'<br>1º envio neste mês':''}${ms[i]===cur?'<br>mês em andamento':''}`,aria:'Evolução da nota por mês'});
    const bi=Math.max(0,ms.indexOf(fm)-1);const bv=h.mon[ms[bi]],nv2=h.mon[cur]??h.last.score,dd=bv!=null&&nv2!=null?nv2-bv:null;
    sub='Nota do mês. Antes = mês anterior ao 1º envio (quando existe); agora = mês atual. A linha tracejada marca o mês da 1ª mensagem.';
    head=ms.length?`${lab[bi]} ${fmt.n(bv,1)} → <b class="big">${fmt.n(nv2,1)}</b> ${dl(dd,true,1,' pts')} <span class="muted">neste mês (parcial)</span>`:'';
  }
  const opts=hist.map(x=>`<option value="${esc(x.id)}" ${x.id===h.id?'selected':''}>${esc(x.name)}</option>`).join('');
  const rows=hist.map(x=>{const dd=x.last.score!=null&&x.first.score!=null?x.last.score-x.first.score:null;const hue=hs(x.id)%360;
    return`<tr class="click msgrow ${x.id===h.id?'sel':''}" data-act="msg-drv" data-id="${esc(x.id)}"><td><div class="row" style="flex-wrap:nowrap;gap:10px"><span class="ava" style="flex-basis:26px;width:26px;height:26px;font-size:10px;background:hsl(${hue} 70% 88%);color:#2b2b2b">${esc(initials(x.name))}</span><b class="nw">${esc(x.name)}</b></div></td><td class="r num">${x.msgs.length}</td><td class="r num">${fmt.n(x.first.score,1)}</td><td class="r num b">${fmt.n(x.last.score,1)}</td><td class="r">${dl(dd,true,1,' pts')}</td></tr>`}).join('');
  return kp+`<div class="grid sec" style="grid-template-columns:repeat(auto-fit,minmax(min(100%,460px),1fr))">
    <div class="card rise">${cardH('Evolução da nota de '+esc(h.name),'',`<select class="inp" data-chg="msg-drv" style="width:auto;max-width:260px" aria-label="Condutor">${opts}</select>${segHtml([['dia','Por dia'],['mes','Por mês']],gran,'msg-gran')}`)}<div class="xs muted" style="margin:-4px 0 10px">${sub}</div><div class="msghead">${head}</div>${chart}</div>
    <div class="card rise">${cardH('Condutores com mensagem','Nota do mês no 1º e no último envio. Clique em um condutor para ver a evolução dele.')}<div class="msgscroll"><table class="tbl msgtbl"><thead><tr><th>Condutor</th><th class="r">Mensagens</th><th class="r">Nota no 1º envio</th><th class="r">Nota no último</th><th class="r">Evolução</th></tr></thead><tbody>${rows}</tbody></table></div></div></div>`;
}
