import bpy, math, sys
from mathutils import Vector
GLB = "amani-tech-v4.glb"
SHOTS = [("WALK_TO_TARGET",0.20),("WALK_TO_TARGET",0.60),("CELEBRATE",0.50),
         ("SHOW_WARNING",0.45),("POINT_AT_UI",0.60),("SPEAK",0.50),("SIT_SPEAK",0.93),("SIT_SPEAK",0.99),("SIT_SPEAK",0.02)]

bpy.ops.wm.read_factory_settings(use_empty=True)
bpy.ops.import_scene.gltf(filepath=GLB)
arm  = next(o for o in bpy.data.objects if o.type=='ARMATURE')
mesh = next(o for o in bpy.data.objects if o.type=='MESH')

pass  # المقياس مصحّح بالملف نفسه
bpy.context.view_layer.update()
print("عظام:", len(arm.data.bones), "| مقاطع:", len(bpy.data.actions))

sc = bpy.context.scene
sc.render.engine = 'CYCLES'
sc.render.resolution_x, sc.render.resolution_y = 520, 760
sc.render.film_transparent = False
w = bpy.data.worlds.new("W"); w.use_nodes=True
w.node_tree.nodes["Background"].inputs[0].default_value=(0.16,0.20,0.25,1)
w.node_tree.nodes["Background"].inputs[1].default_value=1.2
sc.world = w
sc.cycles.samples = 24
sc.cycles.use_denoising = True
sc.cycles.device = 'CPU'

# ضوءان
for loc,en in [((3,-3,4),900),((-3,-2,2),400)]:
    l = bpy.data.lights.new("L",'AREA'); l.energy=en; l.size=4
    o = bpy.data.objects.new("L",l); o.location=loc; sc.collection.objects.link(o)
    d = (Vector((0,0,1.0))-Vector(loc)).normalized().to_track_quat('-Z','Y')
    o.rotation_euler = d.to_euler()

cam_d = bpy.data.cameras.new("C"); cam = bpy.data.objects.new("C",cam_d)
sc.collection.objects.link(cam); sc.camera = cam

for clip, t in SHOTS:
    act = bpy.data.actions.get(clip)
    if not act: print("مفقود:", clip); continue
    if not arm.animation_data: arm.animation_data_create()
    arm.animation_data.action = act
    for tr in arm.animation_data.nla_tracks: tr.mute = True
    s,e = act.frame_range
    sc.frame_set(int(s + (e-s)*t))
    # حدود مقاسة بعد التقييم (لازم evaluated_get وإلا تطلع صفر)
    dg  = bpy.context.evaluated_depsgraph_get()
    ev  = mesh.evaluated_get(dg)
    me  = ev.to_mesh()
    ws  = [mesh.matrix_world @ v.co for v in me.vertices]
    ev.to_mesh_clear()
    lo  = Vector((min(v.x for v in ws), min(v.y for v in ws), min(v.z for v in ws)))
    hi  = Vector((max(v.x for v in ws), max(v.y for v in ws), max(v.z for v in ws)))
    ctr = (lo+hi)/2; size = max((hi-lo).x,(hi-lo).y,(hi-lo).z)
    dist = size*1.6
    cam_d.clip_start = 0.01
    cam.location = ctr + Vector((0.35,-1,0.18)).normalized()*dist
    cam.rotation_euler = (ctr-cam.location).to_track_quat('-Z','Y').to_euler()
    out = f"v4-{clip}-{t}.png"
    sc.render.filepath = "/tmp/claude-0/-home-user/ab9a6f4b-d0a0-53c6-8718-83ae10016613/scratchpad/"+out
    bpy.ops.render.render(write_still=True)
    print(f"✅ {clip} t={t} | طول={round((hi-lo).z,3)} | إطار={sc.frame_current}/{int(e)} → {out}")
