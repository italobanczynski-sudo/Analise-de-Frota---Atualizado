/* =====================================================================
   MAPA (Leaflet) · frota agora e trajeto do dia
   - No servidor: posição e rastro de hoje vêm de /api/v1/clientes/:id/ao-vivo
     e o mapa-base passa pelo próprio servidor (/mapa/z/x/y).
   - Fora do servidor: posições simuladas sobre rodovias reais.
   ===================================================================== */
const GEO={
  via:{routes:[
    [[-23.55,-46.63],[-24.50,-47.84],[-25.43,-49.27],[-26.30,-48.85],[-26.91,-48.66],[-27.59,-48.55]],
    [[-23.55,-46.63],[-22.90,-47.06],[-21.18,-47.81],[-19.75,-47.93]],
    [[-25.43,-49.27],[-25.09,-50.16],[-24.96,-53.46]],
    [[-23.55,-46.63],[-22.73,-45.12],[-22.91,-43.17]]],
    hubs:[['Matriz · São Paulo',-23.55,-46.63],['Filial Sul · Curitiba',-25.43,-49.27],['Filial Norte · Ribeirão Preto',-21.18,-47.81]]},
  caju:{routes:[
    [[-16.68,-49.25],[-17.80,-50.00],[-18.92,-48.28]],
    [[-16.68,-49.25],[-15.79,-47.88]],
    [[-16.68,-49.25],[-18.00,-45.50],[-19.92,-43.94]],
    [[-18.92,-48.28],[-19.75,-47.93],[-21.18,-47.81]]],
    hubs:[['Matriz · Goiânia',-16.68,-49.25],['Filial · Uberlândia',-18.92,-48.28],['Filial · Brasília',-15.79,-47.88]]},
  xtsul:{routes:[
    [[-30.03,-51.23],[-29.17,-51.18],[-28.26,-52.41]],
    [[-30.03,-51.23],[-29.68,-53.80]],
    [[-30.03,-51.23],[-31.77,-52.34]],
    [[-29.17,-51.18],[-27.10,-52.62]]],
    hubs:[['Matriz · Porto Alegre',-30.03,-51.23],['Filial · Caxias do Sul',-29.17,-51.18],['Filial · Passo Fundo',-28.26,-52.41]]}
};
const geoFor=cid=>REAL.on?{routes:[],hubs:[]}:(GEO[cid]||GEO.via);
const ST_LBL={mov:'Em movimento',idle:'Motor ligado parado',stop:'Desligado',off:'Sem sinal agora',unk:'Sem posição'};
const ST_COL={mov:'var(--ok)',idle:'var(--warn)',stop:'var(--n400)',off:'var(--bad)',unk:'var(--n300)'};
const ST_ORDER=['mov','idle','stop','off','unk'];
const LIVE_ST={movimento:'mov',ligado:'idle',desligado:'stop',sem_sinal:'off',nao_encontrado:'off'};
function liveOf(C,id){const l=REAL.live[C.cid];return l&&l.veiculos?l.veiculos[id]:null}
function vehStat(C,v){
  if(REAL.on){const lv=liveOf(C,v.id);if(lv)return LIVE_ST[lv.status]||'off';return C.cur.noSignal.some(x=>x.id===v.id)?'off':'unk'}
  if(C.cur.noSignal.some(x=>x.id===v.id))return'off';const a=hs(v.id)%10;return a<7?'mov':a<9?'idle':'stop';
}
function pathAt(route,t){
  const seg=[];let tot=0;for(let i=1;i<route.length;i++){const d=Math.hypot(route[i][0]-route[i-1][0],route[i][1]-route[i-1][1]);seg.push(d);tot+=d}
  let x=clamp(t,0,1)*tot;for(let i=0;i<seg.length;i++){if(x<=seg[i]||i===seg.length-1){const k=seg[i]?x/seg[i]:0;return[route[i][0]+(route[i+1][0]-route[i][0])*k,route[i][1]+(route[i+1][1]-route[i][1])*k]}x-=seg[i]}
  return route[0];
}
function vehGeo(C,v){
  if(REAL.on){const lv=liveOf(C,v.id);if(!lv||lv.lat==null||lv.lon==null)return null;return{lat:+lv.lat,lng:+lv.lon,st:vehStat(C,v),live:lv}}
  const g=geoFor(C.cid);const a=hs(v.id),b=hs(v.id+'r'),c=hs(v.id+'j');
  const ri=a%g.routes.length,t=((b%1000)/1000)*.9+.05;const p=pathAt(g.routes[ri],t);
  const jx=((c%1000)/1000-.5)*.12,jy=(((c>>10)%1000)/1000-.5)*.12;
  return{lat:p[0]+jx,lng:p[1]+jy,ri,t,st:vehStat(C,v)};
}
const hhBR=utc=>{try{return new Date(utc.replace(' ','T')+'Z').toLocaleTimeString('pt-BR',{timeZone:'America/Sao_Paulo',hour:'2-digit',minute:'2-digit'})}catch(e){return''}};
const kmBetween=(a,b)=>{const R=6371,rad=Math.PI/180,dLa=(b[0]-a[0])*rad,dLo=(b[1]-a[1])*rad;const h=Math.sin(dLa/2)**2+Math.cos(a[0]*rad)*Math.cos(b[0]*rad)*Math.sin(dLo/2)**2;return 2*R*Math.asin(Math.sqrt(h))};
/* trajeto do dia. Servidor: o rastro de ignição ligada desde a meia-noite. Simulado: 16 pontos ao longo da rota */
function dayRoute(C,v){
  if(REAL.on){
    const l=REAL.live[C.cid];const rs=(l&&l.rastro&&l.rastro[v.id]||[]).slice().sort((a,b)=>a.data<b.data?-1:1);
    if(rs.length<2)return null;const pts=rs.map(p=>[+p.lat,+p.lon]);let km=0;for(let i=1;i<pts.length;i++)km+=kmBetween(pts[i-1],pts[i]);
    let paradas=0;for(let i=1;i<rs.length;i++)if(rs[i].status!=='movimento'&&rs[i-1].status==='movimento')paradas++;
    return{pts,km:Math.round(km),paradas,stops:[{i:0,l:'Primeiro sinal do dia',h:hhBR(rs[0].data)},{i:pts.length-1,l:'Último sinal',h:hhBR(rs[rs.length-1].data)}],geo:vehGeo(C,v)};
  }
  const g=geoFor(C.cid);const geo=vehGeo(C,v);const route=g.routes[geo.ri];const pts=[];const n=16;const span=.28;
  const rnd=mulberry(hs(v.id+'day'));
  for(let i=0;i<n;i++){const t=geo.t-span+span*(i/(n-1));const p=pathAt(route,t);pts.push([p[0]+(rnd()-.5)*.03,p[1]+(rnd()-.5)*.03])}
  pts[n-1]=[geo.lat,geo.lng];
  const c=C.cur.byVeh[v.id];const km=c?Math.round(c.kmDay):0;
  const start=5*60+40+Math.floor(rnd()*30);const hh=m=>String(Math.floor(m/60)%24).padStart(2,'0')+':'+String(m%60).padStart(2,'0');
  const stops=[{i:0,l:'Saída',h:hh(start)},{i:5,l:'Parada · 18 min',h:hh(start+95)},{i:11,l:'Parada · 32 min',h:hh(start+230)},{i:n-1,l:'Agora',h:hh(start+330)}];
  return{pts,stops,km,paradas:2,geo};
}

/* ---------- instâncias ---------- */
let MAPS=[];
/* Mapa-base: OpenStreetMap. No servidor, os quadros passam por /mapa (com cache e sem expor quem pede). */
const TILE_URL='https://tile.openstreetmap.org/{z}/{x}/{y}.png';
const TILE_ATTR='© OpenStreetMap';
const tileUrl=()=>REAL.on?'/mapa/{z}/{x}/{y}':TILE_URL;
const tilesDark=el=>el.classList.toggle('dk',S.theme==='dark');
/* a tela pode ser redesenhada antes do mapa terminar de abrir: só mexe nele se ainda estiver na página */
const alive=inst=>MAPS.includes(inst)&&document.body.contains(inst.el);
const later=(inst,fn,ms)=>setTimeout(()=>{if(alive(inst))try{fn()}catch(e){}},ms);
function pruneMaps(){MAPS=MAPS.filter(m=>{if(document.body.contains(m.el))return true;try{m.map.stop();m.map.remove()}catch(e){}return false})}
function mapEl(id,o={}){return`<div class="lmap ${o.cls||''}" data-map="${o.mode||'full'}" ${id?`data-id="${esc(id)}"`:''} style="height:${o.h||520}px" role="application" aria-label="Mapa da frota"></div>`}

function pin(st,label,mini){const z=mini?16:30;return L.divIcon({className:'mkw',html:`<div class="mk ${st==='unk'?'stop':st}"><b>${mini?'':esc(label)}</b></div>`,iconSize:[z,z],iconAnchor:[z/2,z/2],popupAnchor:[0,-z/2]})}
function popupHtml(C,v,geo){
  const c=C.cur.byVeh[v.id];const lv=geo.live;
  const when=lv&&lv.min_atras!=null?`<div class="xs muted">Posição de ${lv.min_atras<1?'agora':lv.min_atras+' min atrás'}${lv.velocidade!=null&&geo.st==='mov'?' · '+fmt.n(lv.velocidade,0)+' km/h':''}</div>`:'';
  const hasRoute=!REAL.on||((REAL.live[C.cid]||{}).rastro||{})[v.id]?.length>1;
  return`<div class="mpop"><div class="tag">${ST_LBL[geo.st]}</div><b style="font-size:17px">${esc(v.id)}</b><div class="xs muted">${esc(v.driver)}</div>${when}${c?`<div class="mrow"><span><span class="tag">Consumo</span><br><b class="num">${fmt.kmpl(c.kmpl)}</b> km/l</span><span><span class="tag">Nota</span><br><b class="num">${fmt.n(c.score,0)}</b></span><span><span class="tag">Km/dia</span><br><b class="num">${fmt.n(c.kmDay,0)}</b></span></div>`:'<div class="xs muted" style="margin:8px 0">Sem rodagem no período</div>'}<div class="row" style="gap:6px;margin-top:8px;flex-wrap:nowrap">${hasRoute?`<button class="btn primary sm" data-act="map-route" data-id="${esc(v.id)}">Trajeto do dia</button>`:''}<button class="btn ghost sm" data-act="veh-open" data-id="${esc(v.id)}">Abrir</button></div></div>`;
}
function initMaps(root=document){
  pruneMaps();
  $$('.lmap',root).forEach(el=>{
    if(MAPS.some(m=>m.el===el))return;
    const C=LASTC;if(!C)return;
    if(typeof L==='undefined'){el.innerHTML=`<div class="map" style="height:100%">${mapSvg(C,el.clientHeight||300)}</div>`;return}
    const mode=el.dataset.mode||el.dataset.map;const mini=mode==='mini',one=mode==='route';
    const map=L.map(el,{zoomControl:false,attributionControl:true,scrollWheelZoom:!mini&&!one,dragging:!mini,doubleClickZoom:!mini,boxZoom:false,keyboard:!mini,tap:false,zoomAnimation:!mini,fadeAnimation:!mini,markerZoomAnimation:!mini});
    if(!mini)L.control.zoom({position:'bottomright'}).addTo(map);map.attributionControl.setPrefix(false);
    tilesDark(el);
    const tl=L.tileLayer(tileUrl(),{maxZoom:18,attribution:TILE_ATTR}).addTo(map);
    const inst={el,map,tl,layers:{},markers:{},mode,route:null};MAPS.push(inst);
    const g=geoFor(C.cid);const all=[];
    /* rodovias de referência (sob os veículos, para o mapa fazer sentido mesmo sem o mapa-base) */
    const ref=L.layerGroup(g.routes.map(r=>L.polyline(r,{color:'#6e6e6a',weight:3,opacity:.7,dashArray:'2 7',interactive:false})));inst.ref=ref;
    let bad=0;tl.on('tileerror',()=>{if(++bad===3){el.classList.add('notiles');ref.addTo(map)}});tl.on('tileload',()=>{if(el.classList.contains('notiles')){el.classList.remove('notiles');map.removeLayer(ref)}});
    g.hubs.forEach(h=>{L.marker([h[1],h[2]],{icon:L.divIcon({className:'mkw',html:`<div class="hub" title="${esc(h[0])}"><span>${esc(h[0].split('·')[1].trim())}</span></div>`,iconSize:[10,10],iconAnchor:[5,5]}),interactive:false,zIndexOffset:-500}).addTo(map)});
    if(one){
      const v=C.fleet.byId[el.dataset.id];const dr=v&&dayRoute(C,v);
      if(dr){drawRoute(inst,C,v,dr);later(inst,()=>{map.invalidateSize();map.fitBounds(L.latLngBounds(dr.pts),{padding:[40,40],maxZoom:11,animate:false})},450)}
      else{map.setView([-15.8,-47.9],4);const geo=v&&vehGeo(C,v);if(geo){L.marker([geo.lat,geo.lng],{icon:pin(geo.st,v.id.slice(-2))}).addTo(map);map.setView([geo.lat,geo.lng],11)}
        const msg=document.createElement('div');msg.className='notiles-msg';msg.style.display='block';msg.textContent=REAL.on?'Sem rastro de ignição ligada hoje para este veículo.':'Sem trajeto';el.parentNode.appendChild(msg)}
      return;
    }
    inst.cluster=!mini;
    C.fleet.vehicles.forEach(v=>{const geo=vehGeo(C,v);if(!geo)return;const grp=inst.layers[geo.st]||(inst.layers[geo.st]=L.layerGroup().addTo(map));
      const m=L.marker([geo.lat,geo.lng],{icon:pin(geo.st,v.id.slice(-2),mini),riseOnHover:true});if(mini)m.addTo(grp);
      if(!mini)m.bindPopup(()=>popupHtml(C,v,geo),{closeButton:false,offset:[0,0],maxWidth:260});else m.bindTooltip(`<b>${esc(v.id)}</b> · ${ST_LBL[geo.st]}`);
      m.on('click',()=>{inst.sel=v.id;$$('.mside .li').forEach(l=>l.classList.toggle('on',l.dataset.id===v.id))});
      inst.markers[v.id]={m,geo,v};all.push([geo.lat,geo.lng])});
    g.hubs.forEach(h=>all.push([h[1],h[2]]));
    if(all.length)map.fitBounds(all,{padding:[36,36],maxZoom:9});else map.setView([-15.8,-47.9],4);
    later(inst,()=>{map.invalidateSize();if(all.length)map.fitBounds(all,{padding:[36,36],maxZoom:9,animate:false})},250);
    if(inst.cluster)map.on('zoomend',()=>renderClusters(inst));
    applyMapFilter(inst);
  });
}
const CL_COL={mov:'#1e9e5a',idle:'#fcb709',stop:'#9b9b96',off:'#d93a2b',unk:'#c7c7c2'};
/* agrupamento por proximidade: junta veículos que se sobrepõem na tela e separa quando o zoom aproxima */
function renderClusters(inst){
  const map=inst.map,f=S.u.mapFilter||'all';
  if(inst.cl)inst.cl.clearLayers();else inst.cl=L.layerGroup().addTo(map);
  const z=map.getZoom(),cells={};
  Object.values(inst.markers).forEach(o=>{if(!(f==='all'||f===o.geo.st))return;const p=map.project([o.geo.lat,o.geo.lng],z);const k=Math.floor(p.x/46)+'_'+Math.floor(p.y/46);(cells[k]=cells[k]||[]).push(o)});
  Object.values(cells).forEach(arr=>{
    if(arr.length===1){arr[0].m.setLatLng([arr[0].geo.lat,arr[0].geo.lng]).addTo(inst.cl);return}
    if(z>=12){const c=map.project([arr[0].geo.lat,arr[0].geo.lng],z);arr.forEach((o,i)=>{const a=i*2*Math.PI/arr.length,r=22;o.m.setLatLng(map.unproject([c.x+Math.cos(a)*r,c.y+Math.sin(a)*r],z)).addTo(inst.cl)});return}
    const n=arr.length,cnt={};arr.forEach(o=>cnt[o.geo.st]=(cnt[o.geo.st]||0)+1);let acc=0;
    const ring=Object.entries(cnt).map(([k,v])=>{const a=acc/n*360;acc+=v;return`${CL_COL[k]} ${a}deg ${acc/n*360}deg`}).join(',');
    const lat=arr.reduce((a,o)=>a+o.geo.lat,0)/n,lng=arr.reduce((a,o)=>a+o.geo.lng,0)/n;
    L.marker([lat,lng],{icon:L.divIcon({className:'mkw',html:`<div class="mkc" style="background:conic-gradient(${ring})"><b>${n}</b></div>`,iconSize:[44,44],iconAnchor:[22,22]}),zIndexOffset:300})
      .on('click',()=>map.fitBounds(L.latLngBounds(arr.map(o=>[o.geo.lat,o.geo.lng])),{padding:[70,70],maxZoom:Math.min(13,z+4)})).addTo(inst.cl)});
}
function applyMapFilter(inst){
  if(inst.cluster){renderClusters(inst);return}
  const f=S.u.mapFilter||'all';Object.entries(inst.layers).forEach(([st,lg])=>{const on=f==='all'||f===st;if(on&&!inst.map.hasLayer(lg))inst.map.addLayer(lg);if(!on&&inst.map.hasLayer(lg))inst.map.removeLayer(lg)});
}
function drawRoute(inst,C,v,dr){
  const map=inst.map;if(inst.route)inst.route.forEach(l=>map.removeLayer(l));inst.route=[];
  const base=L.polyline(dr.pts,{color:'#2b2b2b',weight:7,opacity:.9,lineCap:'round',lineJoin:'round'}).addTo(map);
  const dash=L.polyline(dr.pts,{color:'#fedb00',weight:3,opacity:1,className:'route-anim',lineCap:'round'}).addTo(map);
  inst.route.push(base,dash);
  dr.stops.forEach((s,k)=>{const last=k===dr.stops.length-1;
    const mk=L.marker(dr.pts[s.i],{icon:L.divIcon({className:'mkw',html:`<div class="stp-pin ${last?'now':''}"><b>${last?'●':k===0?'S':k}</b></div>`,iconSize:[26,26],iconAnchor:[13,13]})}).addTo(map);
    mk.bindTooltip(`<b>${s.l}</b> · ${s.h}`,{direction:'top',offset:[0,-10]});inst.route.push(mk)});
  map.fitBounds(L.latLngBounds(dr.pts),{padding:[50,50],maxZoom:11});
  return dr;
}
function mapFocus(id){
  const inst=MAPS.find(m=>m.markers&&m.markers[id]);if(!inst){toast('Este veículo ainda não tem posição conhecida');return}const mk=inst.markers[id];
  const fl=S.u.mapFilter||'all';if(fl!=='all'&&fl!==mk.geo.st){S.u.mapFilter='all';applyAllFilters()}
  inst.map.flyTo([mk.geo.lat,mk.geo.lng],Math.max(inst.map.getZoom(),13),{duration:.8});later(inst,()=>{if(!inst.map.hasLayer(mk.m)){mk.m.addTo(inst.cl)}mk.m.openPopup()},1050);
  $$('.mside .li').forEach(l=>l.classList.toggle('on',l.dataset.id===id));
}
function mapRoute(id){
  const C=LASTC,inst=MAPS.find(m=>m.markers&&m.markers[id]);if(!inst)return;const v=C.fleet.byId[id];const dr=dayRoute(C,v);
  if(!dr){toast('Sem rastro de hoje para '+id);return}
  inst.map.closePopup();drawRoute(inst,C,v,dr);
  const bar=$('#routeBar');if(bar){bar.hidden=false;bar.innerHTML=`<div><div class="tag">Trajeto de hoje</div><b style="font-size:18px">${esc(v.id)}</b> <span class="muted xs">${esc(v.driver)}</span></div><div class="mrow" style="margin:0"><span><span class="tag">Distância</span><br><b class="num">${dr.km} km</b></span><span><span class="tag">Paradas</span><br><b class="num">${dr.paradas}</b></span><span><span class="tag">Início</span><br><b class="num">${dr.stops[0].h}</b></span></div><button class="btn ghost sm" data-act="map-route-clear" aria-label="Limpar trajeto">${ico('x',14)}</button>`}
}
function mapRouteClear(){MAPS.forEach(m=>{if(m.route){m.route.forEach(l=>m.map.removeLayer(l));m.route=null}});const bar=$('#routeBar');if(bar)bar.hidden=true}
function applyAllFilters(){MAPS.forEach(m=>m.layers&&applyMapFilter(m));$$('[data-act=map-filter]').forEach(b=>b.classList.toggle('on',b.dataset.v===(S.u.mapFilter||'all')));$$('.mside .li').forEach(l=>{const f=S.u.mapFilter||'all';l.hidden=!(f==='all'||l.dataset.st===f)})}
function mapsTheme(){MAPS.forEach(m=>tilesDark(m.el))}

/* ---------- blocos de tela ---------- */
function liveNote(C){
  if(!REAL.on)return'';const at=REAL.liveAt[C.cid];
  if(REAL.busy['live'+C.cid]&&!at)return`<div class="xs muted" style="margin:-4px 0 10px">Buscando a posição de cada veículo na API Gobrax…</div>`;
  if(REAL.liveErr&&REAL.liveErr[C.cid])return`<div class="xs muted" style="margin:-4px 0 10px">Posição indisponível agora: ${esc(REAL.liveErr[C.cid])}</div>`;
  return at?`<div class="xs muted" style="margin:-4px 0 10px">Posições de ${new Date(at).toLocaleTimeString('pt-BR',{hour:'2-digit',minute:'2-digit'})}. Atualiza sozinho a cada 2 minutos.</div>`:'';
}
function fleetMapBlock(C){
  const sc=statusCounts(C);const f=S.u.mapFilter||'all';
  const chip=(k,l,n,c)=>n||k==='all'?`<button class="chip ${f===k?'on':''}" data-act="map-filter" data-v="${k}">${c?`<i style="background:${c}"></i>`:''}${l} · ${n}</button>`:'';
  const list=C.fleet.vehicles.map(v=>({v,st:vehStat(C,v)})).sort((a,b)=>ST_ORDER.indexOf(a.st)-ST_ORDER.indexOf(b.st)).map(({v,st})=>`<div class="li" data-act="map-focus" data-id="${esc(v.id)}" data-st="${st}"><span class="dotst" style="background:${ST_COL[st]}"></span><div style="flex:1;min-width:0"><b style="display:block">${esc(v.id)}</b><span class="xs muted">${esc(v.driver)}</span></div><span class="xs muted">${ST_LBL[st]}</span></div>`).join('');
  return`<div class="mapwrap sec rise"><aside class="mside card"><div class="tag" style="margin-bottom:8px">Frota agora</div>${liveNote(C)}<div class="chips" style="margin-bottom:12px">${chip('all','Todos',C.fleet.vehicles.length)}${chip('mov','Movimento',sc.mov,'var(--ok)')}${chip('idle','Parado ligado',sc.idle,'var(--warn)')}${chip('stop','Desligado',sc.stop,'var(--n400)')}${chip('off','Sem sinal agora',sc.off,'var(--bad)')}${chip('unk','Sem posição',sc.unk,'var(--n300)')}</div><div class="scrolly stack" style="gap:2px">${list}</div></aside>
  <div class="mapcol"><div id="routeBar" class="routebar" hidden></div>${mapEl('',{mode:'full',h:600})}<div class="notiles-msg">Mapa-base indisponível agora. Os veículos continuam posicionados.</div></div></div>`;
}
function miniMapBlock(C,h=240){return`<div class="map" style="position:relative">${mapEl('',{mode:'mini',h})}<div class="notiles-msg">Sem internet para o mapa-base</div></div>`}
function routeMapBlock(C,id,h=240){return`<div class="map" style="position:relative">${mapEl(id,{mode:'route',h})}<div class="notiles-msg">Sem internet para o mapa-base</div></div>`}
