/* Source pixels are reused with SVG clipping. L/R refer to screen sides.
 * The original PNG is never changed. Sampling fills only the newly exposed
 * eye/eyelid surfaces; it cannot recover hidden anatomy or new viewpoints. */
(() => {
  'use strict';
  const NS = 'http://www.w3.org/2000/svg';
  const sourceURL = 'assets/golden-male-3d-v1.png';
  const shapes = {
    head: 'M388 191 Q378 169 390 144 L371 108 Q375 56 424 32 Q498 -3 555 36 Q599 72 591 126 L571 159 Q595 147 596 178 Q596 213 572 229 L561 274 L547 316 L535 349 Q500 370 448 352 L420 306 L417 259 Q398 239 397 215 Q387 207 388 191Z',
    hairBack: 'M391 151 L377 125 L362 104 L373 96 L369 81 L383 78 L379 64 L397 52 Q416 18 467 17 Q515 9 553 34 L569 57 L593 71 L593 119 L578 159 L563 174 L558 130 Q535 72 499 93 Q435 112 413 154 L402 181Z',
    hairFront: 'M385 108 Q389 66 431 42 Q477 10 538 35 Q568 46 569 71 Q536 113 493 104 Q445 94 413 152 L407 166 L399 144Z',
    browL: 'M418 145 Q440 131 467 141 L468 154 Q441 144 419 157Z',
    browR: 'M501 140 Q520 127 547 140 L551 151 Q524 138 502 152Z',
    eyeL: 'M426 172 Q438 157 456 161 Q465 163 472 174 Q451 191 426 177Z',
    eyeR: 'M502 169 Q510 155 526 155 Q538 155 544 168 Q524 183 502 172Z',
    mouth: 'M444 231 Q481 216 532 227 L528 249 Q487 265 448 251Z',
    beard: 'M414 212 Q427 223 442 216 Q479 206 505 211 L531 212 L556 202 L568 215 Q563 271 537 286 Q490 317 453 287 Q426 274 419 245Z',
    chest: 'M431 272 Q453 344 492 350 Q540 353 552 273 L583 293 L604 326 L642 355 L663 390 L645 504 L636 621 L649 760 L655 795 Q621 811 548 816 L475 821 L354 805 L363 749 L375 686 L380 588 L378 498 L368 405 L373 360 L401 329Z',
    upperArmL: 'M371 370 Q391 387 385 457 L378 542 L366 594 L342 617 L287 609 L281 585 L298 519 L306 463 L322 415 L343 385Z',
    forearmL: 'M293 559 L366 567 L367 611 L354 665 L345 701 L334 751 L324 809 L281 817 L271 794 L276 743 L268 700 L266 669 L275 618Z',
    handL: 'M282 792 L322 797 L326 820 L343 854 L341 888 L334 900 L323 890 L316 864 L313 883 L328 910 L324 923 L306 918 L293 900 L279 912 L269 893 L259 870 L260 845 L271 811Z',
    upperArmR: 'M655 364 Q698 390 715 443 L734 493 L747 548 L768 585 L767 614 L702 629 L686 601 L675 551 L654 517 L646 463Z',
    forearmR: 'M697 558 L754 552 L771 596 L768 649 L767 692 L756 743 L744 797 L694 820 L683 801 L695 752 L690 698 L687 649Z',
    handR: 'M694 790 L746 788 L758 815 L765 844 L754 879 L743 901 L729 914 L713 917 L698 930 L683 926 L686 914 L704 888 L710 865 L694 886 L679 893 L670 882 L676 859 L684 837Z',
    thighL: 'M354 793 L478 817 L515 890 L494 956 L484 1005 L481 1072 L367 1084 L348 1046 L338 1007 L337 938 L345 875Z',
    thighR: 'M486 817 L644 796 L659 855 L662 923 L668 987 L680 1069 L560 1084 L538 1039 L523 981 L500 921Z',
    shinL: 'M348 1027 L482 1031 L476 1099 L474 1170 L466 1257 L462 1330 L456 1386 L358 1397 L341 1360 L336 1300 L337 1194 L333 1100Z',
    shinR: 'M553 1029 L677 1034 L680 1115 L682 1188 L681 1260 L679 1337 L683 1396 L582 1408 L569 1374 L563 1299 L561 1211 L548 1111Z',
    footL: 'M357 1368 L454 1383 L453 1418 L447 1456 L421 1470 L380 1488 L328 1493 L283 1484 L274 1464 L284 1437 L312 1415 L335 1384Z',
    footR: 'M585 1381 L675 1379 L687 1399 L711 1424 L726 1461 L727 1497 L712 1514 L661 1520 L614 1507 L597 1490 L594 1442Z'
  };
  const pivots = {
    head:[491,333], hairBack:[487,139], hairFront:[487,139],
    browL:[446,148], browR:[524,145], eyeL:[449,172], eyeR:[523,166],
    pupilL:[449,172], pupilR:[523,166], eyelidL:[449,157], eyelidR:[523,152],
    mouth:[488,243], beard:[490,262], chest:[497,765],
    upperArmL:[375,393], forearmL:[325,590], handL:[301,806],
    upperArmR:[659,394], forearmR:[726,592], handR:[717,804],
    thighL:[420,825], shinL:[411,1060], footL:[408,1392],
    thighR:[574,829], shinR:[615,1060], footR:[638,1400], documentAnchor:[301,845]
  };
  const anchors = {shoulderL:pivots.upperArmL,shoulderR:pivots.upperArmR,
    elbowL:pivots.forearmL,elbowR:pivots.forearmR,wristL:pivots.handL,wristR:pivots.handR,
    hipL:pivots.thighL,hipR:pivots.thighR,kneeL:pivots.shinL,kneeR:pivots.shinR,
    ground:[503,1520],uiTarget:[491,166]};
  const rigs = [];
  function el(tag, attrs={}, parent) {
    const n = document.createElementNS(NS,tag);
    for (const [k,v] of Object.entries(attrs)) n.setAttribute(k,String(v));
    if (parent) parent.append(n);
    return n;
  }
  function build(host,prefix,face=false) {
    const svg = el('svg',{viewBox:face?'355 0 260 335':'0 0 1024 1536',role:'img','aria-label':'قصاصات متحركة من الصورة المعتمدة'},host);
    const defs=el('defs',{},svg), groups={}, eyeWindows={};
    el('image',{id:`${prefix}-source`,href:sourceURL,width:1024,height:1536},defs);
    function cut(name,path,parent=svg, exclusions=[]) {
      const clip=el('clipPath',{id:`${prefix}-clip-${name}`,clipPathUnits:'userSpaceOnUse'},defs);
      el('path',{d:path},clip);
      const g=el('g',{id:`${prefix}-${name}`,'data-layer':name,class:'layer'},parent);
      let wrapper=g;
      if(exclusions.length){
        const mask=el('mask',{id:`${prefix}-mask-${name}`,maskUnits:'userSpaceOnUse',x:0,y:0,width:1024,height:1536},defs);
        el('rect',{width:1024,height:1536,fill:'white'},mask);
        exclusions.forEach(d=>el('path',{d,fill:'black',stroke:'white','stroke-width':1},mask));
        wrapper=el('g',{mask:`url(#${prefix}-mask-${name})`},g);
      }
      el('use',{href:`#${prefix}-source`,'clip-path':`url(#${prefix}-clip-${name})`},wrapper);
      groups[name]=g;
      return g;
    }
    // Child limbs inherit their parent's rotation, so wrists and held objects stay attached.
    for(const side of ['L','R']) {
      const thigh=cut('thigh'+side,shapes['thigh'+side]);
      const shin=cut('shin'+side,shapes['shin'+side],thigh);
      cut('foot'+side,shapes['foot'+side],shin);
    }
    cut('chest',shapes.chest);
    for(const side of ['L','R']) {
      const upper=cut('upperArm'+side,shapes['upperArm'+side]);
      const fore=cut('forearm'+side,shapes['forearm'+side],upper);
      cut('hand'+side,shapes['hand'+side],fore);
    }
    const facial=['hairBack','hairFront','browL','browR','eyeL','eyeR','beard'];
    const head=cut('head',shapes.head,svg,facial.map(n=>shapes[n]));
    cut('hairBack',shapes.hairBack,head,[shapes.hairFront]);
    cut('hairFront',shapes.hairFront,head);
    cut('beard',shapes.beard,head,[shapes.mouth]);
    cut('mouth',shapes.mouth,head);
    cut('browL',shapes.browL,head);cut('browR',shapes.browR,head);
    function sample(parent, sx,sy,sw,sh, x,y,w,h) {
      const s=el('svg',{x,y,width:w,height:h,viewBox:`${sx} ${sy} ${sw} ${sh}`,preserveAspectRatio:'none'},parent);
      el('use',{href:`#${prefix}-source`},s);
    }
    for(const side of ['L','R']) {
      const [cx,cy]=pivots['eye'+side], name='eye'+side;
      const aperture=el('clipPath',{id:`${prefix}-aperture-${side}`},defs);
      el('path',{d:shapes[name]},aperture);
      const eye=el('g',{id:`${prefix}-${name}`,'data-layer':name,class:'layer'},head);groups[name]=eye;
      const window=el('g',{'clip-path':`url(#${prefix}-aperture-${side})`},eye);eyeWindows[side]=window;
      // Replicate a small original white-of-eye sample beneath the moving original iris.
      sample(window, side==='L'?459:533,side==='L'?176:166,1,1,cx-24,cy-15,48,32);
      const pupil=cut('pupil'+side,`M${cx-10} ${cy}a10 11 0 1 0 20 0a10 11 0 1 0 -20 0`,window);
      const lid=el('g',{id:`${prefix}-eyelid${side}`,'data-layer':'eyelid'+side,class:'layer'},window);
      groups['eyelid'+side]=lid;
      sample(lid,cx-8,cy+24,16,5,cx-24,cy-16,48,34);
      // Compress the original eye edge into a lash crease. This retains the
      // source's eyelash colour and curve rather than drawing a new closed eye.
      const crease=el('g',{
        transform:`translate(0 ${cy+5}) scale(1 .065) translate(0 ${-cy})`,
        'data-eyelid-crease':side
      },lid);
      el('use',{href:`#${prefix}-source`,'clip-path':`url(#${prefix}-aperture-${side})`},crease);
      lid.style.transform='scaleY(0)';
    }
    const doc=el('g',{id:`${prefix}-documentAnchor`,'data-layer':'documentAnchor',class:'layer',visibility:'hidden'},groups.handL);
    groups.documentAnchor=doc;
    el('rect',{x:300,y:816,width:115,height:145,rx:6,fill:'#fffcf3',stroke:'#bad0e4','stroke-width':4},doc);
    const flap=el('g',{'data-paper-flap':''},doc);
    el('path',{d:'M302 818L357 868L413 818Z',fill:'#dfe9f4',stroke:'#a1b6cb','stroke-width':3},flap);
    el('path',{d:'M322 881h70m-70 18h59m-59 18h64',stroke:'#426b91','stroke-width':5},doc);
    const markers=el('g',{class:'joint-markers'},svg);
    for(const [name,[cx,cy]]of Object.entries(anchors)) {
      el('circle',{id:`${prefix}-${name}`,'data-anchor':name,cx,cy,r:10,fill:'#ffe251',stroke:'#001a31','stroke-width':3},markers);
    }
    const rig={svg,groups,flap,host,prefix};rigs.push(rig);
    for(const [name,g]of Object.entries(groups)) {
      const b=g.getBBox(),[x,y]=pivots[name];
      g.style.transformBox='fill-box';
      g.style.transformOrigin=`${b.width?(x-b.x)/b.width*100:50}% ${b.height?(y-b.y)/b.height*100:50}%`;
    }
    return rig;
  }
  const courier=build(document.querySelector('#courier'),'courier');
  const recipient=build(document.querySelector('#recipient'),'recipient');
  const lens=build(document.querySelector('#face'),'lens',true);
  const reduced=document.querySelector('#reduced'), media=matchMedia('(prefers-reduced-motion: reduce)');
  const label=document.querySelector('#state'),message=document.querySelector('#message'),ack=document.querySelector('#ack');
  const phaseNames=['RECEIVED','PICKED_UP','WALKING','OPENED','DELIVERED','ACK_WAIT'];
  const durations=[900,1000,2500,1400,900,Infinity];
  let phase=-1,elapsed=0,clock=0,raf=0,timer=0,last=0,mode='',testUntil=0,entries=[],wasReduced=false;
  const noMotion=()=>reduced.checked||media.matches;
  function state(name,text){
    label.textContent=name;message.textContent=text;
    entries.push(name);entries=entries.slice(-12);
    const li=document.createElement('li');li.textContent=name;
    document.querySelector('#log').append(li);
    while(document.querySelector('#log').children.length>12)document.querySelector('#log').firstChild.remove();
  }
  function controls(){
    document.querySelectorAll('#play,#gaze,#blink,#walk').forEach(b=>b.disabled=phase>=0);
    ack.disabled=phase!==5;
  }
  function advance(){
    phase++;elapsed=0;controls();
    const texts=['وصل أمر الخصم التجريبي','المراقب يحمل الورقة','المراقب في طريقه إلى الموظف','المراقب يفتح الورقة','استلم أفاتار الموظف الورقة','تنبيه تجريبي: خُصمت نقطة. اقرأ التفاصيل ثم أقرّ بالاطلاع.'];
    state(phaseNames[phase],texts[phase]);
  }
  function neutral(rig){
    Object.entries(rig.groups).forEach(([name,g])=>{g.style.transform=name.startsWith('eyelid')?'scaleY(0)':'none';});
    rig.groups.documentAnchor.setAttribute('visibility','hidden');
    rig.flap.style.transform='';rig.host.style.transform='';
  }
  function gaze(rig,x,blink){
    for(const s of ['L','R']){
      rig.groups['pupil'+s].style.transform=`translateX(${x}px)`;
      rig.groups['eyelid'+s].style.transform=`scaleY(${blink})`;
    }
  }
  function walking(rig,t,carry){
    const swing=Math.sin(t*10);
    for(const s of ['L','R']){
      const sign=s==='L'?1:-1;
      const hip=sign*swing*4,knee=Math.max(0,sign*swing)*7;
      rig.groups['thigh'+s].style.transform=`rotate(${hip}deg)`;
      rig.groups['shin'+s].style.transform=`rotate(${knee}deg)`;
      // Counter-rotate at the ankle to keep the sole level through the step.
      rig.groups['foot'+s].style.transform=`rotate(${-hip-knee}deg)`;
      if(!(carry&&s==='L'))rig.groups['upperArm'+s].style.transform=`rotate(${-sign*swing*5}deg)`;
    }
  }
  function carry(rig,opened=false){
    rig.groups.forearmL.style.transform='rotate(-12deg)';
    rig.groups.documentAnchor.setAttribute('visibility','visible');
    rig.flap.style.transformOrigin='300px 818px';rig.flap.style.transform=opened?'scaleY(-.65)':'none';
  }
  function draw(){
    if(noMotion()){
      if(!wasReduced)rigs.forEach(neutral);
      wasReduced=true;return;
    }
    wasReduced=false;rigs.forEach(neutral);
    const time=clock/1000,autoBlink=(clock%4400)>4200?Math.sin((clock%4400-4200)/200*Math.PI):0;
    const blink=mode==='blink'&&clock<testUntil?Math.sin(Math.min(1,(testUntil-clock)/450)*Math.PI):autoBlink;
    const look=(mode==='gaze'&&clock<testUntil?Math.sin(time*3)*5:Math.sin(time*.8)*2);
    rigs.forEach(r=>gaze(r,look,document.querySelector('#lidHold').checked?1:Math.max(0,blink)));
    if(mode==='walk'&&clock<testUntil)walking(courier,time,false);
    if(phase>=1&&phase<=3)carry(courier,phase===3);
    if(phase===2){
      walking(courier,time,true);
      const progress=Math.min(1,elapsed/durations[phase]);
      // Transfer the courier into the employee panel halfway through the journey.
      const target=recipient.host.parentElement;
      if(progress<.5){
        const distance=origin.clientWidth*.5+80;
        courier.host.style.transform=`translateX(${progress*2*distance}px)`;
      }
      else{
        if(courier.host.parentElement!==target)target.append(courier.host);
        courier.host.style.left='4%';
        const remaining=1-(progress-.5)*2;
        courier.host.style.transform=`translateX(${-remaining*180}px)`;
      }
    }
    if(phase>=3){
      const target=recipient.host.parentElement;if(courier.host.parentElement!==target)target.append(courier.host);
      courier.host.style.left='4%';
    }
    if(phase===3)courier.groups.head.style.transform='rotate(-3deg)';
    if(phase===4){
      carry(courier,true);
      carry(recipient,true);
      recipient.groups.documentAnchor.setAttribute('visibility','hidden');
      const progress=Math.min(1,elapsed/durations[phase]);
      const from=courier.groups.documentAnchor.parentElement.getScreenCTM();
      const to=recipient.groups.documentAnchor.parentElement.getScreenCTM();
      if(from&&to){
        const local=new DOMPoint(301,845).matrixTransform(to).matrixTransform(from.inverse());
        courier.groups.documentAnchor.style.transform=`translate(${(local.x-301)*progress}px,${(local.y-845)*progress}px)`;
      }
    }
    if(phase===5)carry(recipient,true);
  }
  function schedule(){
    if(document.hidden||raf||timer)return;
    if(noMotion()){
      if(phase>=0&&phase<5)timer=setTimeout(()=>tick(performance.now()),100);
    }else raf=requestAnimationFrame(tick);
  }
  function tick(now){
    raf=0;timer=0;if(document.hidden)return;
    const dt=last?Math.max(now-last,0):0;last=now;clock+=dt;
    if(phase>=0&&phase<5){elapsed+=dt;if(elapsed>=durations[phase])advance();}
    draw();
    schedule();
  }
  const origin=courier.host.parentElement;
  function reset(){phase=-1;elapsed=0;mode='';origin.append(courier.host);courier.host.style.left='calc(50% - 80px)';rigs.forEach(neutral);state('IDLE','بانتظار وصول الرسالة');controls();draw();}
  document.querySelector('#play').onclick=()=>{reset();advance();last=0;schedule();};
  document.querySelector('#reset').onclick=reset;
  ack.onclick=()=>{reset();message.textContent='تم الإقرار بالاطلاع في المعاينة فقط.';};
  ['gaze','blink','walk'].forEach(name=>document.querySelector('#'+name).onclick=()=>{mode=name;testUntil=clock+(name==='blink'?450:3500);});
  function preference(){cancelAnimationFrame(raf);clearTimeout(timer);raf=0;timer=0;last=0;document.querySelector('#pause').textContent=noMotion()?'الحركة متوقفة؛ الرسالة والإقرار يعملان.':'';draw();schedule();}
  reduced.onchange=preference;media.addEventListener('change',preference);
  document.addEventListener('visibilitychange',()=>{cancelAnimationFrame(raf);clearTimeout(timer);raf=0;timer=0;last=0;schedule();});
  document.querySelector('#debug').onchange=e=>document.body.classList.toggle('debug',e.target.checked);
  const select=document.querySelector('#layer');
  Object.keys(courier.groups).forEach(name=>{const o=document.createElement('option');o.value=name;o.textContent=name;select.append(o);});
  select.onchange=()=>{
    const chosen=lens.groups[select.value];
    for(const [name,g]of Object.entries(lens.groups)) {
      g.style.opacity=!chosen||name===select.value||g.contains(chosen)||chosen.contains(g)?'1':'.14';
    }
  };
  document.querySelector('#export').onclick=async()=>{
    const response=await fetch(sourceURL);if(!response.ok)return;
    const data=await response.blob();const reader=new FileReader();
    reader.onload=()=>{
      const clone=courier.svg.cloneNode(true);clone.querySelector('image').setAttribute('href',reader.result);
      clone.querySelectorAll('[data-layer]').forEach(g=>{g.id=g.dataset.layer;g.style.transform=g.dataset.layer.startsWith('eyelid')?'scaleY(0)':'none';});
      clone.querySelectorAll('[data-anchor]').forEach(g=>g.id=g.dataset.anchor);
      clone.querySelector('[data-layer="documentAnchor"]').setAttribute('visibility','hidden');
      clone.querySelector('.joint-markers').setAttribute('display','none');
      const blob=new Blob([new XMLSerializer().serializeToString(clone)],{type:'image/svg+xml'});
      const url=URL.createObjectURL(blob),a=document.createElement('a');a.href=url;a.download='golden-male-layered.svg';a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);
    };reader.readAsDataURL(data);
  };
  preference();
})();
