import bpy
from mathutils import Vector
U="/root/.claude/uploads/ab9a6f4b-d0a0-53c6-8718-83ae10016613/"
def gaps(path, is_glb, frames):
    bpy.ops.wm.read_factory_settings(use_empty=True)
    if is_glb: bpy.ops.import_scene.gltf(filepath=path)
    else:      bpy.ops.import_scene.fbx(filepath=path)
    arm=next(o for o in bpy.data.objects if o.type=='ARMATURE')
    if is_glb:
        act=bpy.data.actions["CELEBRATE"]
        if not arm.animation_data: arm.animation_data_create()
        for tr in list(arm.animation_data.nla_tracks): arm.animation_data.nla_tracks.remove(tr)
        arm.animation_data.action=act
    # طول الجذع مرجع مقياس: حوض ← رأس
    sc=bpy.context.scene; sc.frame_set(1); bpy.context.view_layer.update()
    pb=arm.pose.bones
    hips=arm.matrix_world @ pb['mixamorig:Hips'].head
    head=arm.matrix_world @ pb['mixamorig:Head'].head
    torso=(head-hips).length
    out=[]
    for fr in frames:
        sc.frame_set(fr); bpy.context.view_layer.update()
        L=arm.matrix_world @ pb['mixamorig:LeftHand'].head
        R=arm.matrix_world @ pb['mixamorig:RightHand'].head
        g=(L-R).length
        out.append((fr, g, g/torso))   # نسبة للجذع = مقياس-مستقل
    return torso, out
t1,g1 = gaps(U+"fb2cdae8-Clapping.fbx", False, [1,9,17])
t2,g2 = gaps("amani-tech-v4.glb", True, [0,7,13])
print(f"المصدر: طول الجذع={t1:.4f} وحدة")
for fr,g,r in g1: print(f"   إطار {fr:2d}: المسافة={g:.5f} وحدة = **{r*100:5.1f}٪ من الجذع**")
print(f"\nشخصية أماني v4: طول الجذع={t2:.4f} م")
for fr,g,r in g2: print(f"   إطار {fr:2d}: المسافة={g*100:5.2f} سم = **{r*100:5.1f}٪ من الجذع**")
