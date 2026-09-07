import bpy
import math
scene=bpy.context.scene
root=bpy.data.objects['Character • direction and travel']
scene.frame_set(30);start=root.location.x
scene.frame_set(50)
assert abs(root.rotation_euler.z-math.pi/2)<1e-5
assert abs(root.location.x-start)<1e-5, 'turn must precede travel'
for name,frames in [('Shoe L',range(59,70)),('Shoe R',range(72,89))]:
    positions=[]
    for f in frames:
        scene.frame_set(f)
        positions.append(bpy.data.objects[name].matrix_world.translation.copy())
    drift=max((p-positions[0]).length for p in positions)
    assert drift<1e-5,(name,drift)
    print('STANCE_WORLD_DRIFT',name,drift)
scene.frame_set(130)
assert abs(root.location.x-start-1.44)<1e-5
scene.frame_set(180)
assert abs(root.rotation_euler.z)<1e-5
assert len([o for o in bpy.data.objects if o.type=='MESH'])>40
print('PASS: volumetric meshes, turn before travel, planted mid-cycle shoes, final front orientation')
