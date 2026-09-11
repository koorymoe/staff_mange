import bpy, math
from mathutils import Vector
bpy.ops.wm.read_factory_settings(use_empty=True)
bpy.ops.import_scene.gltf(filepath="amani-tech-v2.glb")
mesh = next(o for o in bpy.data.objects if o.type=='MESH')
arm  = next(o for o in bpy.data.objects if o.type=='ARMATURE')
for tr in (arm.animation_data.nla_tracks if arm.animation_data else []): tr.mute=True
if arm.animation_data: arm.animation_data.action=None
bpy.context.view_layer.update()
V=[mesh.matrix_world @ v.co for v in mesh.data.vertices]
zs=[v.z for v in V]; foot_top=min(zs)+0.10
def stats(side):
    pts=[v for v in V if v.z<foot_top and (v.x>0.02 if side=='R' else v.x<-0.02)]
    if not pts: return None
    xs=[p.x for p in pts]; ys=[p.y for p in pts]
    w=max(xs)-min(xs); l=max(ys)-min(ys)
    # زاوية التواء: اتجاه المحور الطولي بالمستوي الأفقي
    cy=sum(ys)/len(ys); cx=sum(xs)/len(xs)
    num=sum((p.x-cx)*(p.y-cy) for p in pts); den=sum((p.y-cy)**2 for p in pts)
    ang=math.degrees(math.atan2(num,den))
    return len(pts), w, l, ang
for s in ('L','R'):
    n,w,l,a = stats(s)
    print(f"  قدم {s}: رؤوس={n} | عرض={w*100:.1f} سم | طول={l*100:.1f} سم | التواء={a:+.2f}°")
nL,wL,lL,aL = stats('L'); nR,wR,lR,aR = stats('R')
print(f"\n  فرق العرض  = {abs(wL-wR)*100:.1f} سم")
print(f"  فرق الالتواء = {abs(aL-aR):.2f}°  (المجموع {aL+aR:+.2f}° — الصفر يعني متناظرين)")
