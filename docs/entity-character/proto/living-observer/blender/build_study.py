"""Local, original procedural 3D study. No uploaded images or external assets.
Run: blender --background --factory-startup --python build_study.py
This is a jointed mesh study, not a production skin/face rig.
"""
import bpy
import math
import os
from mathutils import Vector

OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'output')
os.makedirs(OUT, exist_ok=True)
bpy.ops.object.select_all(action='SELECT')
bpy.ops.object.delete(use_global=False)

def material(name, color, rough=.45, metallic=0):
    m = bpy.data.materials.new(name)
    m.diffuse_color = (*color, 1)
    m.use_nodes = True
    bs = m.node_tree.nodes.get('Principled BSDF')
    bs.inputs['Base Color'].default_value = (*color, 1)
    bs.inputs['Roughness'].default_value = rough
    bs.inputs['Metallic'].default_value = metallic
    return m

skin = material('Warm skin', (.64, .34, .18))
blue = material('Amani blue', (.008, .07, .50))
blue_trim = material('Vest seam', (.02, .085, .28))
black = material('Black fabric', (.016, .019, .025), .8)
shoe = material('Black shoes', (.012, .015, .021), .36)
hair = material('Dark brown hair', (.037, .017, .009), .6)
white = material('Eye whites', (.91, .91, .86), .23)
iris = material('Brown iris', (.14, .052, .014), .25)
pupil = material('Pupil', (.004, .003, .003), .2)
silver = material('Zipper metal', (.43, .52, .64), .3, .65)

def empty(name, parent=None, location=(0,0,0)):
    o = bpy.data.objects.new(name, None)
    bpy.context.collection.objects.link(o)
    o.parent = parent
    o.location = location
    return o

root = empty('Character • direction and travel')
head = empty('Head joint', root, (0,0,2.44))

def sphere(name, loc, scale, mat, parent=root):
    bpy.ops.mesh.primitive_uv_sphere_add(segments=32, ring_count=20)
    o = bpy.context.object
    o.name = name
    o.parent = parent
    o.location = loc
    o.scale = scale
    o.data.materials.append(mat)
    for p in o.data.polygons: p.use_smooth = True
    return o

def box(name, loc, scale, mat, parent=root, bevel=.035):
    bpy.ops.mesh.primitive_cube_add(size=1)
    o = bpy.context.object
    o.name = name
    o.scale = scale
    bpy.ops.object.transform_apply(location=False, rotation=False, scale=True)
    o.parent = parent
    o.location = loc
    o.data.materials.append(mat)
    mod = o.modifiers.new('Rounded tailoring', 'BEVEL')
    mod.width = bevel
    mod.segments = 3
    o.modifiers.new('Weighted normals', 'WEIGHTED_NORMAL')
    return o

def tube(name, points, radius, mat, parent=root):
    c = bpy.data.curves.new(name, 'CURVE')
    c.dimensions = '3D'
    c.bevel_depth = radius
    c.bevel_resolution = 3
    s = c.splines.new('POLY')
    s.points.add(len(points)-1)
    for p, co in zip(s.points, points): p.co = (*co, 1)
    o = bpy.data.objects.new(name, c)
    bpy.context.collection.objects.link(o)
    o.parent = parent
    o.data.materials.append(mat)
    return o

# The head has actual volume, including a back surface, ears and nose.
sphere('Face volume', (0,0,0), (.46,.365,.52), skin, head)
sphere('Jaw beard', (0,-.018,-.24), (.374,.35,.265), hair, head)
for sign in [-1,1]:
    sphere('Ear', (sign*.452,0,-.015), (.09,.07,.14), skin, head)
    sphere('Cheek', (sign*.205,-.27,-.125), (.18,.105,.17), skin, head)
    sphere('Eye white', (sign*.175,-.337,.055), (.124,.066,.143), white, head)
    sphere('Iris', (sign*.175,-.397,.052), (.066,.022,.083), iris, head)
    sphere('Pupil', (sign*.175,-.417,.052), (.037,.013,.055), pupil, head)
    sphere('Eye catchlight', (sign*.175-.018,-.431,.083), (.016,.008,.019), white, head)
    tube('Eyebrow', [(sign*.175+v,-.355,.24+.03*math.cos(v*12)) for v in [-.115,-.06,0,.06,.105]], .027, hair, head)
sphere('Nose bridge', (0,-.351,-.045), (.07,.075,.15), skin, head)
sphere('Nose tip', (0,-.423,-.115), (.097,.085,.069), skin, head)
sphere('Smile opening', (0,-.343,-.27), (.17,.026,.065), hair, head)
sphere('Smile teeth', (0,-.365,-.251), (.128,.013,.024), white, head)
for sign in [-1,1]:
    tuft = sphere('Moustache', (sign*.073,-.353,-.205), (.10,.042,.036), hair, head)
    tuft.rotation_euler.y = sign*.18
sphere('Hair back', (0,.035,.30), (.464,.363,.28), hair, head)
for i in range(8):
    tuft = sphere('Swept hair lock', (-.34+i*.092,-.15,.40+.055*math.sin(i*.4)), (.14,.25,.12), hair, head)
    tuft.rotation_euler.y = -.35
    tuft.rotation_euler.z = -.3
sphere('Neck', (0,0,1.97), (.14,.14,.22), skin)
sphere('Black shirt', (0,0,1.60), (.305,.205,.43), black)

# A closed vest mesh, not a picture on a plane.
rings = [(1.18,.31,.235),(1.24,.35,.25),(1.6,.345,.255),(1.84,.37,.23),(1.94,.24,.17)]
verts, faces = [], []
for z, rx, ry in rings:
    for j in range(48):
        a = 2*math.pi*j/48
        verts.append((rx*math.cos(a), ry*math.sin(a), z))
for i in range(len(rings)-1):
    for j in range(48):
        a=i*48+j; b=i*48+(j+1)%48
        faces.append((a,b,b+48,a+48))
faces.extend([tuple(reversed(range(48))),tuple((len(rings)-1)*48+j for j in range(48))])
mesh = bpy.data.meshes.new('Tailored vest mesh')
mesh.from_pydata(verts, [], faces)
vest = bpy.data.objects.new('Blue zipped vest', mesh)
bpy.context.collection.objects.link(vest)
vest.parent = root
vest.data.materials.append(blue)
for p in mesh.polygons: p.use_smooth=True
box('Zip black backing', (0,-.254,1.58), (.026,.021,.67), black, bevel=.006)
tube('Zip teeth', [(0,-.269,1.25),(0,-.27,1.87)], .008, silver)
box('Zipper pull', (0,-.287,1.8), (.036,.019,.058), silver, bevel=.008)
for sign in [-1,1]:
    box('Lower pocket', (sign*.20,-.218,1.38), (.205,.036,.20), blue, bevel=.022)
    tube('Pocket edge', [(sign*.20-.09,-.244,1.48),(sign*.20+.09,-.244,1.48)], .007, blue_trim)
    collar=box('Collar', (sign*.11,-.14,1.92), (.16,.07,.16), blue, bevel=.025)
    collar.rotation_euler.y=sign*.25
# Placeholder badge is labelled as such, not the official company logo.
badge=sphere('Placeholder badge', (-.19,-.229,1.74), (.06,.012,.06), white)
sphere('Badge blue centre', (-.19,-.243,1.74), (.043,.006,.043), blue)

def trouser(name):
    vertices, quads = [], []
    for z,r in [(-.5,.8),(-.40,1),(.36,1),(.5,.82)]:
        for i in range(24):
            a=i*math.tau/24
            vertices.append((r*math.cos(a),r*math.sin(a),z))
    for ring in range(3):
        for i in range(24):
            a=ring*24+i;b=ring*24+(i+1)%24
            quads.append((a,b,b+24,a+24))
    quads.extend([tuple(reversed(range(24))),tuple(72+i for i in range(24))])
    mesh=bpy.data.meshes.new(name)
    mesh.from_pydata(vertices,[],quads)
    obj=bpy.data.objects.new(name,mesh)
    bpy.context.collection.objects.link(obj);obj.parent=root
    obj.data.materials.append(black)
    for poly in mesh.polygons:poly.use_smooth=True
    return obj

arms, legs = {}, {}
for sign, side in [(-1,'L'),(1,'R')]:
    shoulder=empty('Shoulder '+side,root,(sign*.375,0,1.82))
    sphere('Sleeve upper '+side,(sign*.025,0,-.19),(.115,.12,.26),black,shoulder)
    elbow=empty('Elbow '+side,shoulder,(sign*.035,0,-.39))
    sphere('Sleeve forearm '+side,(0,-.012,-.14),(.087,.094,.19),black,elbow)
    sphere('Hand '+side,(0,-.016,-.35),(.075,.065,.12),skin,elbow)
    sphere('Thumb '+side,(-sign*.061,-.043,-.32),(.029,.035,.07),skin,elbow)
    arms[side]=shoulder
    legs[side]={
        'upper':trouser('Trouser thigh '+side),
        'lower':trouser('Trouser shin '+side),
        'shoe':box('Shoe '+side,(sign*.19,-.08,.10),(.23,.39,.20),shoe,bevel=.065)
    }

def segment(o,a,b,width,depth):
    a,b=Vector(a),Vector(b)
    o.location=(a+b)/2
    o.rotation_mode='QUATERNION'
    o.rotation_quaternion=(b-a).to_track_quat('Z','Y')
    o.scale=(width,depth,(b-a).length*1.10)

def set_leg(side,sign,forward,lift):
    hip=Vector((sign*.19,0,1.19))
    ankle=Vector((sign*.19,-forward,.20+lift))
    delta=ankle-hip
    d=delta.length
    # Equal-length two-link IK bends forward in the sagittal plane.
    bend=Vector((0,-delta.z,delta.y)).normalized()
    if bend.y>0: bend=-bend
    knee=(hip+ankle)/2+bend*math.sqrt(max(0,.55**2-(d/2)**2))
    segment(legs[side]['upper'],hip,knee,.143,.15)
    segment(legs[side]['lower'],knee,ankle,.105,.115)
    legs[side]['shoe'].location=(sign*.19,-forward-.08,.10+lift)

def smooth(t):
    t=max(0,min(1,t));return t*t*(3-2*t)

def travel(t,offset):
    # Stance holds a world-space landing position; swing moves to next landing.
    period=40; half=20; speed=.018
    u=t+offset; cycle=math.floor(u/period); p=u%period
    land=(cycle*period-offset)*speed
    if p<half: return land,0
    v=(p-half)/half
    return land+period*speed*smooth(v), .13*math.sin(math.pi*v)**2

scene=bpy.context.scene
scene.frame_start=1;scene.frame_end=180;scene.render.fps=24
for frame in range(1,181):
    scene.frame_set(frame)
    # Face viewer -> turn right -> walk right -> stop -> face viewer.
    if frame<=30: yaw=0; t=0
    elif frame<=50: yaw=math.pi/2*smooth((frame-30)/20);t=0
    elif frame<=130: yaw=math.pi/2;t=frame-50
    else: yaw=math.pi/2*(1-smooth((frame-140)/24));t=80
    root.location.x=-.72+t*.018
    root.rotation_euler.z=yaw
    root.keyframe_insert('location',frame=frame)
    root.keyframe_insert('rotation_euler',frame=frame)
    active=50<frame<=130
    for sign,side,offset in [(-1,'L',0),(1,'R',20)]:
        if active:
            ground,lift=travel(t,offset)
            forward=ground-t*.018
            # Gentle starts/ends blend the planted cycle to the neutral pose.
            weight=min(smooth(t/8),smooth((80-t)/8))
            forward*=weight;lift*=weight
        else: forward=lift=0
        set_leg(side,sign,forward,lift)
        for o in legs[side].values():
            o.keyframe_insert('location',frame=frame)
            o.keyframe_insert('scale',frame=frame)
            if o.rotation_mode=='QUATERNION':o.keyframe_insert('rotation_quaternion',frame=frame)
        arms[side].rotation_euler.x=sign*.30*math.sin(t/40*2*math.pi)* (weight if active else 0)
        arms[side].keyframe_insert('rotation_euler',frame=frame)

floor_mat=material('Studio navy',(.025,.045,.08),.75)
box('Ground',(0,0,-.10),(200,200,.15),floor_mat,parent=None,bevel=.02)
world=bpy.data.worlds.new('Studio world')
scene.world=world;world.use_nodes=True
world.node_tree.nodes['Background'].inputs[0].default_value=(.12,.16,.24,1)
world.node_tree.nodes['Background'].inputs[1].default_value=.4
def area(name,loc,power,size):
    bpy.ops.object.light_add(type='AREA',location=loc)
    o=bpy.context.object;o.name=name;o.data.energy=power;o.data.shape='DISK';o.data.size=size
    o.rotation_euler=(Vector((0,0,1.5))-o.location).to_track_quat('-Z','Y').to_euler()
area('Soft key',(-3,-4,6),650,5)
area('Cool fill',(4,-2,3),350,4)
area('Rim',(1,3,5),850,3)
bpy.ops.object.camera_add(location=(.1,-8,3.65))
camera=bpy.context.object
camera.rotation_euler=(Vector((0,0,1.48))-camera.location).to_track_quat('-Z','Y').to_euler()
camera.data.type='ORTHO';camera.data.ortho_scale=4.3
scene.camera=camera
scene.render.engine='CYCLES'
scene.cycles.samples=24
scene.cycles.use_denoising=True
scene.render.resolution_x=720;scene.render.resolution_y=720;scene.render.resolution_percentage=100
scene.render.image_settings.file_format='PNG'
scene.view_settings.view_transform='AgX'
scene.frame_set(1)
for screen in bpy.data.screens:
    for area_ui in screen.areas:
        if area_ui.type=='VIEW_3D':
            area_ui.spaces.active.region_3d.view_perspective='CAMERA'
            area_ui.spaces.active.shading.type='MATERIAL'
bpy.ops.wm.save_as_mainfile(filepath=os.path.join(OUT,'amani-3d-study.blend'))
scene.render.filepath=os.path.join(OUT,'front.png')
bpy.ops.render.render(write_still=True)
scene.frame_set(90)
scene.render.filepath=os.path.join(OUT,'side-walk.png')
bpy.ops.render.render(write_still=True)
print('AMANI_STUDY_READY',OUT)
