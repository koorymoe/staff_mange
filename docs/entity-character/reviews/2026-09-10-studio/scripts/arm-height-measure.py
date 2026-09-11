import bpy, math
from mathutils import Vector
bpy.ops.wm.open_mainfile(filepath="amani-studio-review.blend")
arm=next(o for o in bpy.data.objects if o.type=='ARMATURE')
if not arm.animation_data: arm.animation_data_create()
sc=bpy.context.scene
def m(act, fr):
    arm.animation_data.action=bpy.data.actions[act]; sc.frame_set(fr)
    bpy.context.view_layer.update()
    pb=arm.pose.bones
    hips=pb['mixamorig:Hips'].head; head=pb['mixamorig:Head'].head
    up=(head-hips).normalized()          # محور «فوق» الحقيقي مأخوذ من الهيكل نفسه
    span=(head-hips).length
    r={}
    for s in ('Left','Right'):
        hd=pb['mixamorig:%sHand'%s].head
        r[s]=(hd-hips).dot(up)/span*100  # ارتفاع الكف عن الحوض كنسبة من طول الجذع
    return r
print("ارتفاع الكف عن الحوض (٠٪=مستوى الحوض · ١٠٠٪=مستوى الرأس):")
for name,fr in [("WALK_TO_TARGET",20),("SPEAK",60),("CELEBRATE",22)]:
    A=m(name,fr); B=m("SOURCE_"+name,fr)
    print(f"  {name:16s} نسختي: يسرى {B['Left']:+5.1f}٪ يمنى {B['Right']:+5.1f}٪  |  (م): يسرى {A['Left']:+5.1f}٪ يمنى {A['Right']:+5.1f}٪")
