import bpy
bpy.ops.wm.read_factory_settings(use_empty=True)
bpy.ops.import_scene.gltf(filepath="amani-tech-v4.glb")
arm=next(o for o in bpy.data.objects if o.type=='ARMATURE')
names=sorted(b.name.replace('mixamorig:','') for b in arm.data.bones)
fingers={}
for n in names:
    for f in ('Thumb','Index','Middle','Ring','Pinky'):
        if f in n:
            side = 'يسرى' if n.startswith('Left') else 'يمنى'
            fingers.setdefault(f,{}).setdefault(side,[]).append(n)
print("عظام الأصابع بالهيكل:")
for f in ('Thumb','Index','Middle','Ring','Pinky'):
    d=fingers.get(f)
    if not d: print(f"   {f:7s} 🔴 غير موجود"); continue
    for side,ns in d.items():
        print(f"   {f:7s} {side}: {len(ns)} عظمة → {', '.join(x.replace('Left','').replace('Right','') for x in sorted(ns))}")
print("\nمجموع عظام الأصابع:", sum(len(v) for d in fingers.values() for v in d.values()), "من", len(names))
