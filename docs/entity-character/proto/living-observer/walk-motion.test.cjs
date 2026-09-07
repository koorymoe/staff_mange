'use strict';
const assert=require('node:assert/strict');
const {pose}=require('./walk-motion.js');
const pivots={thighL:[420,825],shinL:[411,1060],footL:[408,1392],thighR:[574,829],shinR:[615,1060],footR:[638,1400]};
const rotate=([x,y],degrees)=>{const r=degrees*Math.PI/180;return [x*Math.cos(r)-y*Math.sin(r),x*Math.sin(r)+y*Math.cos(r)];};
let maxError=0;
for(let frame=0;frame<=1000;frame++){
  const p=pose(pivots,frame/1000);
  for(const s of ['L','R']){
    const h=pivots['thigh'+s],k=pivots['shin'+s],a=pivots['foot'+s],v=p[s];
    const upper=rotate([k[0]-h[0],k[1]-h[1]],v.hip);
    const lower=rotate([a[0]-k[0],a[1]-k[1]],v.hip+v.knee);
    const error=Math.hypot(h[0]+upper[0]+lower[0]-v.target[0],h[1]+upper[1]+lower[1]-v.target[1]);
    maxError=Math.max(maxError,error);
    assert.ok(error<1e-7,'ankle reaches target');
    assert.ok(Math.abs(v.hip+v.knee+v.foot)<1e-9,'sole stays level');
    assert.ok(v.target[1]<=a[1]+1e-9,'no penetration below source ground');
    if(v.stance)assert.deepEqual(v.target,a,'support ankle stays fixed in rig space');
    assert.ok(Math.abs(v.hip)<20&&Math.abs(v.knee)<35,'bounded cutout rotations');
  }
  assert.ok(p.L.stance||p.R.stance,'at least one support leg');
}
for(const boundary of [0,.5,1]){
  const before=pose(pivots,boundary-1e-6),after=pose(pivots,boundary+1e-6);
  for(const s of ['L','R'])for(const joint of ['hip','knee','foot'])assert.ok(Math.abs(before[s][joint]-after[s][joint])<.001,'continuous contact boundary');
}
for(const v of Object.values(pose(pivots,.25,0)))assert.ok(Math.abs(v.hip)+Math.abs(v.knee)+Math.abs(v.foot)<1e-8,'zero weight restores source pose');
console.log(`PASS: 1001 cycle samples, support, contact continuity, neutral blend; maximum ankle error ${maxError.toExponential(2)} source px`);
