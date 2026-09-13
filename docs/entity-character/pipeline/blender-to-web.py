import bpy, os

bpy.ops.wm.open_mainfile(filepath='/tmp/blendcheck/Cartoon boy_Blend/cartoon boy.blend')

for o in bpy.data.objects:
    try:
        o.hide_set(False)
    except Exception:
        pass
    o.hide_viewport = False
    o.hide_render = False

# ═══ ① التعابير ═══
WANT = {
    'Basis',
    'Eye_Blink_L', 'Eye_Blink_R', 'Eye_Squint_L', 'Eye_Squint_R',
    'Eye_Wide_L', 'Eye_Wide_R',
    'Brow_Raise_L', 'Brow_Raise_R', 'Brow_Drop_L', 'Brow_Drop_R',
    'Brow_Raise_Inner_L', 'Brow_Raise_Inner_R',
    'Mouth_Smile_L', 'Mouth_Smile_R', 'Mouth_Frown_L', 'Mouth_Frown_R',
    'Mouth_Open', 'Mouth_Pucker', 'Mouth_Widen', 'Mouth_Lips_Part',
    'Cheek_Raise_L', 'Cheek_Raise_R', 'Nose_Scrunch',
    'Explosive', 'Dental_Lip', 'Affricate', 'Open', 'Tight', 'Tight-O', 'Wide',
}
for o in bpy.data.objects:
    if o.type != 'MESH' or not o.data.shape_keys:
        continue
    if o.name != 'CC_Base_Body':
        o.shape_key_clear()
        continue
    for nm in reversed([kb.name for kb in o.data.shape_keys.key_blocks]):
        if nm not in WANT:
            kb = o.data.shape_keys.key_blocks.get(nm)
            if kb:
                o.shape_key_remove(kb)

# ═══ ② الخامات: نبنيها من جديد نظيفة ═══
# ⚠️ **السبب مقاس بالصورة**: خامات Character Creator شجرة عقد كبيرة
# (شيدر جلد بعشرات العقد). ومصدّر glTF **ما يگدر يخزّن شجرة عقد**،
# فيأخذ «أقرب صورة» ويحطها باللون الأساسي — والنتيجة إن خريطة
# النتوء أو اللمعان تطلع **كأنها لون الجلد**، فالجسم يبين **مبقّعاً
# مثل رقعة**. شفتها بعيني باللقطة.
#
# فالحل: نبني لكل خامة شجرة **بسيطة ومعروفة**: صورة `Diffuse` للون،
# و`Normal` للنتوء، و`Opacity` للشفافية — ونختارها **بالاسم** مو
# بالتخمين. هاي تخسر لمعان الشيدر الأصلي، بس تنطي شكلاً **صحيحاً**،
# والصحيح أهم من الفخم.
def pick(mat, *keys):
    """يلگى صورة داخل الخامة اسمها يحتوي أحد المفاتيح."""
    for n in mat.node_tree.nodes:
        if n.type == 'TEX_IMAGE' and n.image:
            nm = n.image.name
            for k in keys:
                if k.lower() in nm.lower():
                    return n.image
    return None

NEEDS_ALPHA = ('Hair', 'Scalp', 'Eyelash', 'Beard', 'Mustache', 'Eyebrow',
               'TearLine', 'Occlusion')

rebuilt = 0
for mat in bpy.data.materials:
    if not mat.use_nodes or not mat.node_tree:
        continue
    diff = pick(mat, '_diffuse', 'basecolor', '_base_color')
    norm = pick(mat, '_normal')
    opac = pick(mat, '_opacity', '_alpha')
    if not diff:
        continue
    nt = mat.node_tree
    nt.nodes.clear()
    out = nt.nodes.new('ShaderNodeOutputMaterial')
    out.location = (400, 0)
    bsdf = nt.nodes.new('ShaderNodeBsdfPrincipled')
    bsdf.location = (100, 0)
    nt.links.new(bsdf.outputs['BSDF'], out.inputs['Surface'])
    if 'Specular IOR Level' in bsdf.inputs:
        bsdf.inputs['Specular IOR Level'].default_value = 0.25
    bsdf.inputs['Roughness'].default_value = 0.55

    tx = nt.nodes.new('ShaderNodeTexImage')
    tx.image = diff
    tx.location = (-350, 100)
    nt.links.new(tx.outputs['Color'], bsdf.inputs['Base Color'])

    alpha_needed = any(k in mat.name for k in NEEDS_ALPHA)
    if opac and alpha_needed:
        to = nt.nodes.new('ShaderNodeTexImage')
        to.image = opac
        to.location = (-350, -200)
        to.image.colorspace_settings.name = 'Non-Color'
        nt.links.new(to.outputs['Color'], bsdf.inputs['Alpha'])
    elif alpha_needed and diff.channels == 4:
        nt.links.new(tx.outputs['Alpha'], bsdf.inputs['Alpha'])

    if norm:
        tn = nt.nodes.new('ShaderNodeTexImage')
        tn.image = norm
        tn.location = (-350, -450)
        tn.image.colorspace_settings.name = 'Non-Color'
        nm = nt.nodes.new('ShaderNodeNormalMap')
        nm.location = (-100, -450)
        nt.links.new(tn.outputs['Color'], nm.inputs['Color'])
        nt.links.new(nm.outputs['Normal'], bsdf.inputs['Normal'])

    for attr, val in (('blend_method', 'BLEND' if alpha_needed else 'OPAQUE'),
                      ('surface_render_method', 'DITHERED' if alpha_needed else 'DITHERED')):
        if hasattr(mat, attr):
            try:
                setattr(mat, attr, val)
            except Exception:
                pass
    rebuilt += 1
print('materials rebuilt:', rebuilt)

# ═══ ③ الخامات: تصغير ═══
SMALL = ('Teeth', 'Nails', 'Tongue', 'Cornea', 'Eyelash', 'TearLine', 'Occlusion')
for im in bpy.data.images:
    if im.size[0] == 0:
        try:
            im.pixels[0]
        except Exception:
            pass
    w, h = im.size
    if w == 0:
        continue
    cap = 512 if any(k in im.name for k in SMALL) else 1024
    if max(w, h) > cap:
        s = cap / max(w, h)
        im.scale(max(1, int(w * s)), max(1, int(h * s)))

bpy.ops.object.select_all(action='DESELECT')
SKIP = {'Plane', 'Camera', 'Point'}
keep = [o for o in bpy.data.objects
        if o.type in ('MESH', 'ARMATURE')
        and not o.name.startswith('WGT-')
        and not o.name.startswith('Sun')
        and o.name not in SKIP]
for o in keep:
    o.select_set(True)
bpy.context.view_layer.objects.active = keep[0]

out = '/tmp/blendcheck/cartoon_boy_web.glb'
bpy.ops.export_scene.gltf(
    filepath=out,
    export_format='GLB',
    use_selection=True,
    export_def_bones=True,
    export_morph=True,
    export_morph_normal=False,
    export_morph_tangent=False,
    export_skins=True,
    export_animations=False,
    export_apply=False,
)
print('OUT SIZE MB: %.2f' % (os.path.getsize(out) / 1048576))
