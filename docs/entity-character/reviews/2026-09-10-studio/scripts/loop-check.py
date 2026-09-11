import bpy, math
bpy.ops.wm.read_factory_settings(use_empty=True)
bpy.ops.import_scene.gltf(filepath="amani-tech-v3.glb")
arm=next(o for o in bpy.data.objects if o.type=='ARMATURE')
if not arm.animation_data: arm.animation_data_create()
sc=bpy.context.scene
def pose(act,fr):
    arm.animation_data.action=bpy.data.actions[act]; sc.frame_set(fr)
    bpy.context.view_layer.update()
    return {b.name:b.matrix.to_quaternion() for b in arm.pose.bones}
def ang(q):
    a=math.degrees(q.angle)%360
    return min(a,360-a)
print("فرق الوضعية بين آخر إطار وأول إطار (٠° = حلقة ناعمة):")
for name in ["SIT_SPEAK","SIT_IDLE","SPEAK","SPEAK_ALT","WALK_TO_TARGET","WALK_ALT","LAYING_IDLE","CELEBRATE","RUN_TO_EDGE","POINT_AT_UI","SHOW_WARNING"]:
    a=bpy.data.actions.get(name)
    if not a: continue
    s,e=a.frame_range
    A=pose(name,int(s)); B=pose(name,int(e))
    d=[ang(A[k].rotation_difference(B[k])) for k in A]
    avg=sum(d)/len(d); mx=max(d)
    flag="✅ حلقة ناعمة" if avg<3 else ("⚠️ قطع محسوس" if avg<12 else "🔴 قطع واضح")
    print(f"  {name:16s} متوسط={avg:5.1f}° أقصى={mx:5.1f}°  {flag}")
