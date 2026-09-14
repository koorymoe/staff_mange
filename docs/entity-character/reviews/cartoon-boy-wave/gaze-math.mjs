export function gazeInput(x,y,rect){
 if(!rect.width||!rect.height)return {x:0,y:0};
 const clamp=v=>Math.max(-1,Math.min(1,v));
 return {x:clamp((x-rect.left)/rect.width*2-1),y:clamp(1-(y-rect.top)/rect.height*2)};
}
export function smoothing(dt){return 1-Math.exp(-12*Math.max(0,Math.min(dt,.05)));}
export function blinkWeight(elapsed){
 if(elapsed<0||elapsed>=.32)return 0;
 return elapsed<.10?elapsed/.10:elapsed<.14?1:1-(elapsed-.14)/.18;
}
