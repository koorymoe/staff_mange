import bpy, os, sys

# ═══ المسارات وسائط، مو مكتوبة ثابتاً ═══
#
# 🔴 **السبب**: (م) رفع نسخته المعدّلة (`amani-wave.blend`) وفيها
# **حركة تلويح وعظام عيون**، وحجمها **١٩.٨٩ م.ب** — يعني **أكبر من
# حد الرفع بخادمنا (١٠ م.ب)** فما تُرفَع أصلاً. والسبب **٩٥١ تعبيراً**
# (كل تعبير نسخة كاملة من مواقع الرؤوس) — ونفس المشكلة الي هالسكربت
# حلّها سابقاً (٥٠٧ ← ٣١). فبدل ما نكتب سكربتاً ثانياً، نخلي المدخل
# والمخرج **وسيطين** ونشغّله على ملفه.
#
# ⚠️ و**نحفظ الحركات** (`--keep-anim`): ملف (م) فيه تلويحته، وحذفها
# يعني نخسر شغله. وبالملف الأصلي ماكو حركات فالوسيط ما يفرّق.
ARGS = sys.argv[sys.argv.index('--') + 1:] if '--' in sys.argv else []
SRC = ARGS[0] if ARGS else '/tmp/blendcheck/Cartoon boy_Blend/cartoon boy.blend'
OUT = ARGS[1] if len(ARGS) > 1 else '/tmp/blendcheck/cartoon_boy_web.glb'
KEEP_ANIM = '--keep-anim' in ARGS

bpy.ops.wm.open_mainfile(filepath=SRC)

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
#
# ⚠️ **السقف حسب دور الصورة مو حجماً واحداً للكل**، والسبب مقاس:
# نسخة (م) طلعت **١٩.٨٩ م.ب** وبعد تقليم التعابير بقت **١٠.٧٣** —
# لا تزال فوق **حد الرفع بخادمنا (١٠ م.ب)** فما تُرفَع. والباقي
# خامات: خريطة نتوء البنطرون وحدها **٠.٩٤ م.ب** وخريطة الحذاء
# **٠.٨٥**.
#
# 🔴 و**خرائط النتوء واللمعان تفاصيل ثانوية**: الشخصية تُعرض بصندوق
# **٩٦×١١٢** بودجة الموظف، فنتوء بدقة ١٠٢٤ **ما يبيّن ولا بكسل**
# منه. واللون (`Diffuse`) هو الي يبيّن فعلاً — فيبقى أعلى.
SMALL = ('Teeth', 'Nails', 'Tongue', 'Cornea', 'Eyelash', 'TearLine', 'Occlusion')
# الجلد والرأس يبقون ١٠٢٤: الوجه هو الي يتفرّس بيه المستخدم.
FACE = ('Skin_Head', 'Std_Eye', 'Eyebrow')
DETAIL = ('Normal', 'Specular', 'Roughness', 'Metallic', 'AO', 'Occlusion')
for im in bpy.data.images:
    if im.size[0] == 0:
        try:
            im.pixels[0]
        except Exception:
            pass
    w, h = im.size
    if w == 0:
        continue
    if any(k in im.name for k in SMALL):
        cap = 512
    elif any(k in im.name for k in DETAIL):
        # تفاصيل ثانوية — ٥١٢ للوجه و٢٥٦ لغيره
        cap = 512 if any(k in im.name for k in FACE) else 256
    elif any(k in im.name for k in FACE):
        cap = 1024
    else:
        # لون الملابس والشعر — ٧٦٨ كافية بحجم الودجة
        cap = 768
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

out = OUT
bpy.ops.export_scene.gltf(
    filepath=out,
    export_format='GLB',
    use_selection=True,
    export_def_bones=True,
    export_morph=True,
    export_morph_normal=False,
    export_morph_tangent=False,
    export_skins=True,
    export_animations=KEEP_ANIM,
    # 🔴 **الخبز إلزامي، وبلاه الحركة تطلع «موجودة» وما تتحرّك.**
    #
    # قِستها: صدّرت تلويحة (م) بلا خبز، والمقطع ظهر بقائمة المقاطع
    # بعارضنا، وضغطته، والنتيجة **صفر بالمية من البكسلات تتغيّر**.
    # والسبب: حركة Rigify مكتوبة على **عظام التحكّم** (`hand_fk.R`
    # و`upper_arm_fk.L` — المصدّر نفسه حذّر منها)، و`export_def_bones`
    # يصدّر **عظام التشويه وحدها**. فالمسارات تشير لعقد **غير
    # موجودة** بالملف — حركة ميتة.
    #
    # والخبز يحسب وضع عظام التشويه **كل إطار** بعد القيود، فتنكتب
    # عليها مباشرةً.
    # ⚠️ **تقليم المفاتيح الزائدة**: المقاطع المنقولة تنكتب **بمفتاح
    # لكل إطار لكل عظمة** (٥١٣ قناة × ١١ مقطعاً) — وهذا رفع الملف من
    # ٦.٧٨ إلى **١١.٨٢ م.ب**، فوق حد الرفع (١٠). والتقليم يشيل
    # المفاتيح الي ما تضيف شي (عظمة ساكنة بمقطع كامل تنخزن مرة وحدة).
    export_optimize_animation_size=True,
    # ⚠️ **مفتاح كل إطارين مو كل إطار**: المصدر ٣٠ إطاراً/ثانية،
    # والناتج ١٥ — وبابل **يُنعّم بين المفاتيح** فالحركة تبقى سلسة.
    # والسبب رقمي: بمفتاح لكل إطار الملف **١١.١٤ م.ب** وحد الرفع
    # **١٠**. والشخصية تُعرض بصندوق ٩٦×١١٢ بودجة الموظف — فرق
    # ٣٠ مقابل ١٥ إطاراً ما يبيّن بهذا الحجم، وحجم ما يُرفَع **يبيّن
    # فوراً** لأنه يمنع الميزة كلها.
    export_frame_step=2,
    export_force_sampling=True,
    export_bake_animation=True,
    export_animation_mode='ACTIONS',
    export_nla_strips=False,
    export_apply=False,
)
print('OUT SIZE MB: %.2f' % (os.path.getsize(out) / 1048576))
