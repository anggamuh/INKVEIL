(() => {
 const pages = Array.from(document.querySelectorAll('[data-dex-page]'));
 const main=document.getElementById('dex-reading-main');
 const status=document.getElementById('dex-status');
 const failedPages=new Set();
 const refreshSource=()=>{const u=new URL(location.href);location.href=u.href;};
 document.querySelectorAll(".dex-refresh").forEach(b=>b.addEventListener("click",refreshSource));
 document.getElementById("dex-quality")?.addEventListener("change",e=>{const u=new URL(location.href);u.searchParams.set("quality",e.target.value);location.href=u.href;});
 const updateStatus=()=>{if(status)status.textContent=failedPages.size?failedPages.size+" gambar gagal. Refresh sumber atau pilih kualitas Original.":"";};
 const load=img=>{
  if(img.dataset.started==='1')return;
  img.dataset.started='1';
  const error=img.parentElement.querySelector('.dex-page-error');
  img.addEventListener('load',()=>{img.classList.add('loaded');img.classList.remove("failed");failedPages.delete(img);updateStatus();if(error)error.hidden=true;},{once:true});
  img.addEventListener('error',()=>{img.classList.add('failed');if(error)error.hidden=false;failedPages.add(img);updateStatus();},{once:true});
  img.src=img.dataset.dexPage;
 };
 if('IntersectionObserver' in window){
  const observer=new IntersectionObserver(entries=>{for(const e of entries)if(e.isIntersecting){observer.unobserve(e.target);load(e.target);}},{rootMargin:'600px'});
  pages.forEach(img=>observer.observe(img));
 }else pages.forEach(load);
 for(const img of pages){const retry=img.parentElement.querySelector('.dex-retry');if(retry)retry.addEventListener('click',()=>{img.classList.remove('failed');img.removeAttribute('src');img.dataset.started='0';img.parentElement.querySelector('.dex-page-error').hidden=true;load(img);});}
 const width=document.getElementById('dex-width');
 if(width&&main){try{const saved=localStorage.getItem('inkveil-dex-width');if(['normal','wide','compact'].includes(saved))width.value=saved;}catch{}main.dataset.width=width.value;width.addEventListener('change',()=>{main.dataset.width=width.value;try{localStorage.setItem('inkveil-dex-width',width.value);}catch{}});}
 document.getElementById('dex-top')?.addEventListener('click',()=>window.scrollTo({top:0,behavior:'smooth'}));
 document.getElementById('dex-bottom')?.addEventListener('click',()=>window.scrollTo({top:document.body.scrollHeight,behavior:'smooth'}));
 document.getElementById('dex-fullscreen')?.addEventListener('click',async()=>{try{if(document.fullscreenElement)await document.exitFullscreen();else await document.documentElement.requestFullscreen();}catch{}});
 document.addEventListener('keydown',e=>{if(e.altKey||e.ctrlKey||e.metaKey||e.shiftKey||['INPUT','SELECT','TEXTAREA','BUTTON'].includes(e.target.tagName))return;const a=e.key==='ArrowLeft'?document.querySelector('[data-dex-prev]'):e.key==='ArrowRight'?document.querySelector('[data-dex-next]'):null;if(a){e.preventDefault();location.href=a.href;}});
})();