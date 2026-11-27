(() => {
 const pages=Array.from(document.querySelectorAll('[data-dex-page]'));
 const main=document.getElementById('dex-reading-main'),status=document.getElementById('dex-status');
 const failed=new Set(),loaded=new Set(),queued=new Set(),queue=[];
 let active=0,refreshing=null,lastRefresh=0;
 const update=()=>{if(status)status.textContent=loaded.size+'/'+pages.length+' halaman dimuat'+(failed.size?' · '+failed.size+' gagal':'');};
 const refresh=async()=>{
  if(refreshing)return refreshing;
  if(Date.now()-lastRefresh<3000)return;
  lastRefresh=Date.now();
  refreshing=(async()=>{const u=new URL(location.href);u.searchParams.set('manifest','1');const r=await fetch(u,{cache:'no-store'});if(!r.ok)throw Error('Manifest '+r.status);const m=await r.json();if(!Array.isArray(m.images)||m.images.length!==pages.length)throw Error('Page count changed');pages.forEach((img,i)=>img.dataset.dexPage=m.images[i]);})();
  try{await refreshing;}finally{refreshing=null;}
 };
 // Render independently of telemetry. If CORS fetch fails, use normal image loading.
 const report=(url,success,cached,bytes,duration)=>{fetch('https://api.mangadex.network/report',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url,success,cached,bytes,duration}),keepalive:true,signal:AbortSignal.timeout(5000)}).catch(()=>{});};
 const display=(img,src)=>new Promise((resolve,reject)=>{
  const timer=setTimeout(()=>finish(false),18000);
  const finish=ok=>{clearTimeout(timer);img.onload=null;img.onerror=null;if(ok)resolve();else reject(Error('Image load failed'));};
  img.onload=()=>finish(true);img.onerror=()=>finish(false);img.src=src;
 });
 const getImage=async(img)=>{
  const url=img.dataset.dexPage,start=performance.now();
  let response,blob;
  try{response=await fetch(url,{credentials:'omit',signal:AbortSignal.timeout(8000)});}
  catch{await display(img,url);return;}
  if(!response.ok){report(url,false,false,0,Math.round(performance.now()-start));throw Error('HTTP '+response.status);}
  try{blob=await response.blob();if(!blob.type.startsWith('image/'))throw Error('Not an image');const objectURL=URL.createObjectURL(blob);try{await display(img,objectURL);}finally{URL.revokeObjectURL(objectURL);}}
  catch(e){report(url,false,false,0,Math.round(performance.now()-start));throw e;}
  report(url,true,(response.headers.get('X-Cache')||'').startsWith('HIT'),blob.size,Math.round(performance.now()-start));
 };
 const run=async(img)=>{
  const error=img.parentElement.querySelector('.dex-page-error');
  try{try{await getImage(img);}catch{await refresh();await new Promise(r=>setTimeout(r,700));await getImage(img);}
   img.classList.add('loaded');img.classList.remove('failed');loaded.add(img);failed.delete(img);if(error)error.hidden=true;
  }catch(e){console.warn('INKVEIL page load:',e);img.classList.add('failed');failed.add(img);if(error)error.hidden=false;}
  finally{queued.delete(img);active--;update();pump();}
 };
 const pump=()=>{while(active<3&&queue.length){const img=queue.shift();active++;run(img);}};
 const enqueue=img=>{if(loaded.has(img)||queued.has(img))return;queued.add(img);queue.push(img);pump();};
 if('IntersectionObserver' in window){const observer=new IntersectionObserver(entries=>{for(const e of entries)if(e.isIntersecting){observer.unobserve(e.target);enqueue(e.target);}},{rootMargin:'600px'});pages.forEach(img=>observer.observe(img));}else pages.forEach(enqueue);
 for(const img of pages){const error=img.parentElement.querySelector('.dex-page-error');error?.querySelector('.dex-retry')?.addEventListener('click',()=>enqueue(img));error?.querySelector('.dex-refresh')?.addEventListener('click',async()=>{try{await refresh();failed.forEach(enqueue);}catch{update();}});}
 document.getElementById('dex-quality')?.addEventListener('change',e=>{const u=new URL(location.href);u.searchParams.set('quality',e.target.value);location.href=u.href;});
 update();
 const width=document.getElementById('dex-width');
 if(width&&main){try{const saved=localStorage.getItem('inkveil-dex-width');if(['normal','wide','compact'].includes(saved))width.value=saved;}catch{}main.dataset.width=width.value;width.addEventListener('change',()=>{main.dataset.width=width.value;try{localStorage.setItem('inkveil-dex-width',width.value);}catch{}});}
 document.getElementById('dex-top')?.addEventListener('click',()=>window.scrollTo({top:0,behavior:'smooth'}));
 document.getElementById('dex-bottom')?.addEventListener('click',()=>window.scrollTo({top:document.body.scrollHeight,behavior:'smooth'}));
 document.getElementById('dex-fullscreen')?.addEventListener('click',async()=>{try{if(document.fullscreenElement)await document.exitFullscreen();else await document.documentElement.requestFullscreen();}catch{}});
 document.addEventListener('keydown',e=>{if(e.altKey||e.ctrlKey||e.metaKey||e.shiftKey||['INPUT','SELECT','TEXTAREA','BUTTON'].includes(e.target.tagName))return;const a=e.key==='ArrowLeft'?document.querySelector('[data-dex-prev]'):e.key==='ArrowRight'?document.querySelector('[data-dex-next]'):null;if(a){e.preventDefault();location.href=a.href;}});
})();