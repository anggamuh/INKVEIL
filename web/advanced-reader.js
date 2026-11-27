(()=>{"use strict";
const main=document.querySelector(".dex-reading-main"),bar=document.getElementById("iv-reader-progress");if(!main||!bar)return;
const get=(k,d)=>{try{return localStorage.getItem(k)||d}catch{return d}},set=(k,v)=>{try{localStorage.setItem(k,v)}catch{}};
const speed=document.getElementById("iv-scroll-speed"),auto=document.getElementById("iv-auto-scroll"),next=document.getElementById("iv-auto-next"),focus=document.getElementById("iv-focus-mode"),full=document.getElementById("iv-fullscreen"),pct=document.getElementById("iv-reader-percent");
const pages=[...main.querySelectorAll(".iv-suwayomi-pages img,.dex-reading-pages img")];let running=false,last=0,frame=0,transitioning=false;
speed.value=get("iv-scroll-speed","65");next.checked=get("iv-auto-next","no")==="yes";speed.onchange=()=>set("iv-scroll-speed",speed.value);next.onchange=()=>set("iv-auto-next",next.checked?"yes":"no");
const width=document.querySelector("[data-iv-reader-width]");if(width){const saved=get("inkveil-dex-width","normal");if(["normal","wide","compact"].includes(saved)){width.value=saved;main.dataset.width=saved}width.addEventListener("change",()=>set("inkveil-dex-width",width.value));}
function stop(){running=false;cancelAnimationFrame(frame);auto.textContent="Auto-scroll";auto.setAttribute("aria-pressed","false");last=0}
function update(){const max=Math.max(1,document.documentElement.scrollHeight-innerHeight),n=Math.max(0,Math.min(100,Math.round(scrollY/max*100)));bar.value=n;pct.textContent=n+"%";if(running&&scrollY>=max-3){stop();const a=main.querySelector("[data-iv-next]");if(next.checked&&a&&pages.length&&pages.every(p=>p.complete&&p.naturalWidth>0)&&!transitioning){transitioning=true;location.assign(a.href)}}}
let carry=0;
function tick(t){if(!running)return;if(last){carry+=Number(speed.value||65)*Math.min(t-last,80)/1000;const step=Math.floor(carry);if(step){scrollBy(0,step);carry-=step}}last=t;update();if(running)frame=requestAnimationFrame(tick)}
auto.onclick=()=>{if(running){stop();return}running=true;carry=0;auto.textContent="Hentikan auto-scroll";auto.setAttribute("aria-pressed","true");frame=requestAnimationFrame(tick)};
focus.onclick=()=>{const active=document.body.classList.toggle("iv-focus-reading");focus.setAttribute("aria-pressed",String(active));focus.textContent=active?"Keluar mode fokus":"Mode fokus"};
full.onclick=async()=>{try{if(document.fullscreenElement)await document.exitFullscreen();else if(document.documentElement.requestFullscreen)await document.documentElement.requestFullscreen();else full.textContent="Fullscreen tidak didukung"}catch{full.textContent="Fullscreen tidak tersedia"}};
document.addEventListener("visibilitychange",()=>{if(document.hidden)stop()});window.addEventListener("pagehide",stop);window.addEventListener("scroll",update,{passive:true});window.addEventListener("resize",update);pages.forEach(p=>p.addEventListener("load",update));update();
})();