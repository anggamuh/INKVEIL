(() => {
 const pages = Array.from(document.querySelectorAll('[data-dex-page]'));
 // Load only near the viewport; report actual load results without fabricated byte counts.
 const load = img => {
  const url = img.dataset.dexPage;
  if (!url || img.dataset.started) return;
  img.dataset.started = '1';
  img.addEventListener('error', () => {
   const status=document.getElementById('dex-status');
   if(status) status.textContent='Sebagian halaman gagal dimuat. Coba refresh chapter.';
  }, {once:true});
  img.src=url;
 };
 if ('IntersectionObserver' in window) {
  const observer=new IntersectionObserver(entries => {
   for(const entry of entries) if(entry.isIntersecting){observer.unobserve(entry.target);load(entry.target);}
  },{rootMargin:'600px'});
  pages.forEach(img=>{img.width=800;img.height=1100;observer.observe(img);img.addEventListener('load',()=>{img.removeAttribute('height');img.removeAttribute('width');},{once:true});});
 } else pages.forEach(load);
})();