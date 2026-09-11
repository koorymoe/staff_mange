import bpy, math
bpy.ops.wm.open_mainfile(filepath="amani-studio-review.blend")
arm=next(o for o in bpy.data.objects if o.type=='ARMATURE')
if not arm.animation_data: arm.animation_data_create()
sc=bpy.context.scene
act=bpy.data.actions["SIT_SPEAK"]
s,e=act.frame_range; s=int(s); e=int(e)
arm.animation_data.action=act
def pose(fr):
    sc.frame_set(fr); bpy.context.view_layer.update()
    return {b.name:b.matrix.to_quaternion() for b in arm.pose.bones}
def ang(q):
    a=math.degrees(q.angle)%360
    return min(a,360-a)
A=pose(s)
print(f"المقطع الأصلي: {s} → {e} ({e-s} إطاراً) · هدفنا ~١٨٠-٤٢٠ إطاراً بأنعم وصلة")
best=[]
for fr in range(s+150, min(e, s+520), 10):
    B=pose(fr)
    d=[ang(A[k].rotation_difference(B[k])) for k in A]
    best.append((sum(d)/len(d), max(d), fr))
best.sort()
print("أفضل خمسة إطارات قصّ (متوسط · أقصى · الإطار · الطول):")
for avg,mx,fr in best[:5]:
    print(f"   متوسط={avg:5.2f}° أقصى={mx:6.2f}° عند الإطار {fr} → طول {fr-s} إطاراً ({(fr-s)/30:.1f} ثا)")
print(f"\nللمقارنة — قصّي الحالي عند {s+240}: ", end="")
B=pose(s+240); d=[ang(A[k].rotation_difference(B[k])) for k in A]
print(f"متوسط={sum(d)/len(d):.2f}° أقصى={max(d):.2f}°")
