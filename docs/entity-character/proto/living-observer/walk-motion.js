/* Small planar solver for the approved front-view cutout, not a 3D rig. */
(function(root){
  'use strict';
  const clamp=(v,a,b)=>Math.max(a,Math.min(b,v));
  const angle=(a,b)=>Math.atan2(b[1]-a[1],b[0]-a[0]);
  const distance=(a,b)=>Math.hypot(b[0]-a[0],b[1]-a[1]);
  const degrees=r=>r*180/Math.PI;
  function leg(hip,knee,ankle,target){
    const a=distance(hip,knee),b=distance(knee,ankle);
    const restHip=angle(hip,knee),restShin=angle(knee,ankle);
    const d=clamp(distance(hip,target),Math.abs(a-b)+1e-6,a+b-1e-6);
    // Preserve the source knee's bend direction; never flip the joint branch.
    const sign=restShin>=restHip?1:-1;
    const bend=sign*Math.acos(clamp((d*d-a*a-b*b)/(2*a*b),-1,1));
    const upper=angle(hip,target)-Math.atan2(b*Math.sin(bend),a+b*Math.cos(bend));
    const hipRotation=degrees(upper-restHip);
    const kneeRotation=degrees(bend-(restShin-restHip));
    return {hip:hipRotation,knee:kneeRotation,foot:-hipRotation-kneeRotation};
  }
  function pose(pivots,cycle,weight=1){
    const result={};
    for(const side of ['L','R']){
      const phase=((cycle+(side==='R'?.5:0))%1+1)%1;
      const swing=phase<.5?phase*2:0;
      // Half a cycle rests at the exact original ankle position. The other
      // half lifts smoothly, with zero velocity at both contact boundaries.
      const lift=Math.sin(Math.PI*swing)**2;
      const stride=Math.sin(2*Math.PI*swing)**3;
      const ankle=pivots['foot'+side];
      const target=[ankle[0]+14*stride*weight,ankle[1]-18*lift*weight];
      result[side]={...leg(pivots['thigh'+side],pivots['shin'+side],ankle,target),
        target,stance:phase>=.5};
    }
    return result;
  }
  const api={leg,pose};
  if(typeof module==='object'&&module.exports)module.exports=api;
  else root.CutoutWalk=api;
})(typeof window==='object'?window:globalThis);
