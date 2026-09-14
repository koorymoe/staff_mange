import * as THREE from 'three';
import {GLTFLoader} from 'three/addons/loaders/GLTFLoader.js';
import {gazeInput,smoothing,blinkWeight} from './gaze-math.mjs';
const canvas=document.querySelector('#view'),status=document.querySelector('#status');
const renderer=new THREE.WebGLRenderer({canvas,antialias:true});renderer.setPixelRatio(Math.min(devicePixelRatio,1.5));renderer.setClearColor(0x192331);renderer.outputColorSpace=THREE.SRGBColorSpace;renderer.toneMapping=THREE.ACESFilmicToneMapping;
const scene=new THREE.Scene(),camera=new THREE.PerspectiveCamera(30,1,.01,100);camera.position.set(0,1,4.2);camera.lookAt(0,.95,0);
renderer.shadowMap.enabled=true;renderer.shadowMap.type=THREE.PCFSoftShadowMap;
const floor=new THREE.Mesh(new THREE.PlaneGeometry(200,200),new THREE.MeshStandardMaterial({color:0x263346,roughness:1}));
floor.rotation.x=-Math.PI/2;floor.receiveShadow=true;scene.add(floor);
const keyLight=new THREE.DirectionalLight(0xfff2df,2);
keyLight.position.set(-2,4,3);keyLight.castShadow=true;
keyLight.shadow.mapSize.set(1024,1024);keyLight.shadow.camera.left=-2;keyLight.shadow.camera.right=2;
keyLight.shadow.camera.top=3;keyLight.shadow.camera.bottom=-2;keyLight.shadow.normalBias=.015;
scene.add(keyLight);
scene.add(new THREE.HemisphereLight(0xe7f0ff,0x6b5861,2));for(const [p,power] of [[[2,4,3],3],[[-3,2,1],1.5]]){const l=new THREE.DirectionalLight(0xffffff,power);l.position.set(...p);scene.add(l);}
let mixer,actions=[],eyes=[],meshes=[],blinkAt=-10,play=false,close=false;const pointer=new THREE.Vector2();
const motionButton=document.querySelector('#motion');
const timeline=document.querySelector('#timeline');
motionButton.disabled=true;
const reduced=matchMedia('(prefers-reduced-motion: reduce)'),smoothPointer=new THREE.Vector2();
let nextBlink=performance.now()/1000+4;
const worldDelta=new THREE.Quaternion(),parentWorld=new THREE.Quaternion(),targetRotation=new THREE.Quaternion();
window.addEventListener('pointermove',e=>{const p=gazeInput(e.clientX,e.clientY,canvas.getBoundingClientRect());pointer.set(p.x,p.y);});
window.addEventListener('blur',()=>pointer.set(0,0));
document.documentElement.addEventListener('pointerleave',()=>pointer.set(0,0));
new GLTFLoader().load('./amani-wave.glb',g=>{
 scene.add(g.scene);g.scene.traverse(o=>{if(o.isMesh){o.castShadow=true;o.receiveShadow=true;for(const m of (Array.isArray(o.material)?o.material:[o.material])){m.side=THREE.DoubleSide;if(m.transparent){m.transparent=false;m.alphaTest=.35;m.depthWrite=true;m.needsUpdate=true;}}}});g.scene.updateMatrixWorld(true);mixer=new THREE.AnimationMixer(g.scene);
 g.scene.traverse(o=>{if(o.isMesh){if(/EyeOcclusion|TearLine/.test(o.name))o.visible=false;for(const m of (Array.isArray(o.material)?o.material:[o.material]))if(/Cornea/i.test(m.name)){m.transparent=true;m.opacity=.08;m.alphaTest=0;m.depthWrite=false;}}});
 actions=g.animations.map(c=>mixer.clipAction(c));
 actions.forEach(a=>{a.setLoop(THREE.LoopOnce,1);a.clampWhenFinished=true;a.play();});mixer.setTime(0);
 g.scene.updateMatrixWorld(true);floor.position.y=new THREE.Box3().setFromObject(g.scene).min.y-.005;
 motionButton.disabled=actions.length===0;
 timeline.max=Math.max(0,...g.animations.map(c=>c.duration));timeline.disabled=actions.length===0;
 g.scene.traverse(o=>{if(o.name==='DEF-eyeL'||o.name==='DEF-eyeR'||/^DEF-eye[._][LR]$/.test(o.name))eyes.push({bone:o,rest:o.quaternion.clone()});if(o.isMesh&&o.morphTargetDictionary)meshes.push(o);});
 status.textContent=`تم التحميل — ${eyes.length} عظمتا عين، ${g.animations.length} مسارات حركة. تتبّع العيون تجريبي.`;
 console.info('AVATAR_LOADED',{eyes:eyes.map(e=>e.bone.name),clips:g.animations.length,meshes:meshes.length});
},undefined,e=>{status.textContent='تعذّر التحميل؛ افتح سجل الأخطاء.';console.error(e);});
document.querySelector('#blink').onclick=()=>blinkAt=performance.now()/1000;
timeline.oninput=()=>{
 if(!mixer)return;play=false;actions.forEach(a=>a.reset().play());mixer.setTime(Number(timeline.value));
 motionButton.disabled=false;motionButton.textContent='إعادة التلويح';
};
motionButton.onclick=()=>{
 if(!mixer||play)return;
 actions.forEach(a=>a.reset().play());
 play=true;motionButton.disabled=true;motionButton.textContent='جاري التلويح…';
};
document.querySelector('#face').onclick=()=>{close=!close;camera.position.set(0,close?1.66:1,close?1.25:4.2);camera.lookAt(0,close?1.66:.95,0);};
function resize(){const top=document.querySelector('header').getBoundingClientRect().bottom+12;const h=Math.max(280,innerHeight-top-48);canvas.style.marginTop=top+'px';canvas.style.height=h+'px';renderer.setSize(innerWidth,h,false);camera.aspect=innerWidth/h;camera.updateProjectionMatrix();}window.addEventListener('resize',resize);new ResizeObserver(resize).observe(document.querySelector('header'));resize();
let last=performance.now();function frame(now){requestAnimationFrame(frame);const dt=Math.min((now-last)/1000,.05);last=now;if(document.hidden)return;if(play&&mixer){
 mixer.update(dt);
 timeline.value=actions[0]?.time??0;
 if(actions.every(a=>!a.isRunning())){
  play=false;motionButton.disabled=false;motionButton.textContent='إعادة التلويح';
 }
}
 if(!reduced.matches&&now/1000>=nextBlink){blinkAt=now/1000;nextBlink=now/1000+4+Math.random()*2;}
 const v=blinkWeight(now/1000-blinkAt)*1.35;
 for(const m of meshes){const i=m.morphTargetDictionary.Eye_Blink;if(i!==undefined)m.morphTargetInfluences[i]=v;}
 smoothPointer.lerp(reduced.matches?new THREE.Vector2():pointer,smoothing(dt));
 worldDelta.setFromEuler(new THREE.Euler(-smoothPointer.y*.12,smoothPointer.x*.18,0));
 for(const e of eyes){e.bone.parent.getWorldQuaternion(parentWorld);targetRotation.copy(parentWorld).invert().multiply(worldDelta).multiply(parentWorld).multiply(e.rest);e.bone.quaternion.copy(targetRotation);}
 renderer.render(scene,camera);
}requestAnimationFrame(frame);
