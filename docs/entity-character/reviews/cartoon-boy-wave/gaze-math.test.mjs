import {test} from 'node:test';import assert from 'node:assert/strict';
import {gazeInput,smoothing,blinkWeight} from './gaze-math.mjs';
test('blink closes quickly, holds, then reopens',()=>{assert.equal(blinkWeight(-1),0);assert.equal(blinkWeight(.12),1);assert.equal(blinkWeight(.32),0);assert.ok(blinkWeight(.20)>0&&blinkWeight(.20)<1);});
test('canvas center maps to forward gaze',()=>assert.deepEqual(gazeInput(300,400,{left:100,top:200,width:400,height:400}),{x:0,y:0}));
test('outside pointer is clamped',()=>assert.deepEqual(gazeInput(-100,999,{left:0,top:200,width:400,height:400}),{x:-1,y:-1}));
test('smoothing is bounded and time based',()=>{assert.equal(smoothing(0),0);assert.ok(smoothing(1/60)>0);assert.ok(smoothing(1)<1);assert.equal(smoothing(-1),0);});
