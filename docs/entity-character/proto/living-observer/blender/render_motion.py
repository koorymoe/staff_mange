"""Fast studio-shaded motion review; run after loading the generated blend."""
import bpy
import os
scene=bpy.context.scene
out=os.path.join(os.path.dirname(bpy.data.filepath),'motion-frames')
os.makedirs(out,exist_ok=True)
scene.render.engine='BLENDER_WORKBENCH'
scene.display.shading.light='STUDIO'
scene.display.shading.color_type='MATERIAL'
scene.display.shading.show_shadows=True
scene.display.shading.show_cavity=True
scene.display.shading.cavity_type='BOTH'
scene.display.shading.background_type='WORLD'
scene.world.color=(.035,.055,.085)
scene.render.resolution_x=480;scene.render.resolution_y=480
scene.render.resolution_percentage=100
scene.render.image_settings.file_format='PNG'
for frame in range(1,181,3):
    scene.frame_set(frame)
    scene.render.filepath=os.path.join(out,f'{frame:03}.png')
    bpy.ops.render.render(write_still=True)
print('MOTION_REVIEW_READY',out)
