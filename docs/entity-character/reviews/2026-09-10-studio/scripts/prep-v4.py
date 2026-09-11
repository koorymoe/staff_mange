# -*- coding: utf-8 -*-
"""
تجهيز amani-tech-v4.glb من ملف مراجعة (م) — سكربت قابل للتكرار.

المبدأ: **الوضعيات المصحّحة مال (م) ما تُلمس إطلاقاً**. هذا السكربت
ينظّف ويقيس بس: مقياس · مقاطع زائدة · قصّ بوصلة ناعمة · تصدير.

يُشغَّل: python3 prep-v4.py   (يحتاج `pip install bpy`)
⚠️ نسخة bpy عندي 5.0.1 · ملف (م) من Blender 5.2.1 LTS — بلندر يحذّر
«ملف من نسخة أحدث» فأي قياس على الـblend ناقص بطبيعته. التبادل
بالـGLB يشيل هذا القيد.
"""
import bpy, math
from mathutils import Vector

SRC = "amani-studio-review.blend"
OUT = "/tmp/claude-0/-home-user/ab9a6f4b-d0a0-53c6-8718-83ae10016613/scratchpad/amani-tech-v4.glb"
TARGET_HEIGHT = 1.70          # متر — ميكسامو يصدّر ×0.01 فالطول يطلع 0.0169
KEEP = ["WALK_TO_TARGET", "WALK_ALT", "RUN_TO_EDGE", "SPEAK", "SPEAK_ALT",
        "CELEBRATE", "SIT_IDLE", "SIT_SPEAK", "LAYING_IDLE", "SHOW_WARNING",
        "POINT_AT_UI"]
# القصّ: الإطار مختار **بالقياس** — مسحت المرشّحات ولگيت 301 يعطي وصلة
# 4.06° مقابل 43.33° عند 241 (قصّي القديم بـv3، وهو الي كسر الحلقة).
TRIM = {"SIT_SPEAK": 301}
BLEND_TAIL = 15               # إطارات تُمزج نحو وضعية البداية لإغلاق الوصلة


def fcurves_of(act):
    """بلندر 5 نقل الـfcurves جوّا layers/strips/channelbags."""
    if hasattr(act, "fcurves"):
        return list(act.fcurves)
    out = []
    for lay in act.layers:
        for st in lay.strips:
            for cb in getattr(st, "channelbags", []):
                out += list(cb.fcurves)
    return out


bpy.ops.wm.open_mainfile(filepath=SRC)
arm = next(o for o in bpy.data.objects if o.type == 'ARMATURE')
mesh = next(o for o in bpy.data.objects if o.type == 'MESH')
print("قبل: مقاطع =", len(bpy.data.actions), "| مقياس =", tuple(round(v, 4) for v in arm.scale))

# ═══ ① حذف المقاطع الزائدة: SOURCE_* والمرجع ═══
for a in list(bpy.data.actions):
    if a.name not in KEEP:
        if arm.animation_data:
            for tr in list(arm.animation_data.nla_tracks):
                if tr.strips and tr.strips[0].action is a:
                    arm.animation_data.nla_tracks.remove(tr)
        bpy.data.actions.remove(a, do_unlink=True)
print("بعد التنظيف: مقاطع =", len(bpy.data.actions))

# ═══ ② القصّ + مزج الذيل لإغلاق الوصلة ═══
for name, cut in TRIM.items():
    act = bpy.data.actions.get(name)
    if not act:
        continue
    s, e = act.frame_range
    if e - s <= cut:
        print(f"  {name}: {int(e-s)} إطاراً — ما يحتاج قصّاً")
        continue
    end = s + cut
    for fc in fcurves_of(act):
        # القيمة الي لازم نرجعلها بآخر الذيل = قيمة الإطار الأول
        start_val = fc.evaluate(s)
        pts = list(fc.keyframe_points)
        # احذف كل شي بعد حد القصّ
        for i in reversed([i for i, k in enumerate(pts) if k.co[0] > end]):
            fc.keyframe_points.remove(fc.keyframe_points[i])
        # وامزج آخر BLEND_TAIL إطاراً نحو قيمة البداية بوزن تدريجي
        for k in fc.keyframe_points:
            t = (k.co[0] - (end - BLEND_TAIL)) / BLEND_TAIL
            if t > 0:
                w = min(1.0, t) ** 2          # تربيعي: يبدأ ناعماً وينتهي كاملاً
                k.co[1] = k.co[1] * (1 - w) + start_val * w
                k.handle_left[1] = k.co[1]
                k.handle_right[1] = k.co[1]
        fc.update()
    print(f"  قصّيت {name} عند {cut} + مزجت {BLEND_TAIL} إطاراً → {int(act.frame_range[1]-act.frame_range[0])} إطاراً")

# ═══ ③ المقياس — محسوب بالقياس مو بثابت ═══
def height_now():
    bpy.context.view_layer.update()
    dg = bpy.context.evaluated_depsgraph_get()
    ev = mesh.evaluated_get(dg)
    me = ev.to_mesh()
    zs = [(mesh.matrix_world @ v.co).z for v in me.vertices]
    h = max(zs) - min(zs)
    ev.to_mesh_clear()
    return h

if arm.animation_data is None:
    arm.animation_data_create()
ref = bpy.data.actions["WALK_TO_TARGET"]
arm.animation_data.action = ref
bpy.context.scene.frame_set(int(ref.frame_range[0]))
h0 = height_now()
factor = TARGET_HEIGHT / h0
arm.scale = tuple(v * factor for v in arm.scale)
print(f"  المقياس: الطول كان {h0:.4f} → معامل ×{factor:.2f} → {height_now():.3f} م")

# ═══ ④ ⚠️ ماكو قياس وصلة هنا — بقصد ═══
# حاولت أقيسها جوّا هذا السكربت فطلعت ٠.٠٠° بكل المقاطع — رقم مكسور:
# بلندر ٥ يحتاج ربط الحركة بـ`action_slot` وبدونه إسناد
# `animation_data.action` ما يقود الوضعية إطلاقاً. وصفر بكل شي مو
# نتيجة، **ورقم غلط أسوأ من ماكو رقم** — فشلته من هنا.
# القياس المعتمد يُسوّى على **الـGLB المصدَّر** بتحميل مستقل:
#     python3 loop-check.py amani-tech-v4.glb
# وهي مثبتة: أعطت SIT_SPEAK ٣٨.٥° قبل الإصلاح و٠.٤° بعده.

# ═══ ⑤ مسارات NLA مكتومة — بدونها بلندر يرمي المقاطع بالتصدير ═══
for a in bpy.data.actions:
    a.use_fake_user = True
for tr in list(arm.animation_data.nla_tracks):
    arm.animation_data.nla_tracks.remove(tr)
for name in KEEP:
    act = bpy.data.actions.get(name)
    if not act:
        print("  ⚠️ مفقود:", name)
        continue
    tr = arm.animation_data.nla_tracks.new()
    tr.name = name
    tr.strips.new(name, int(act.frame_range[0]), act)
    tr.mute = True

# ═══ ⑥ التصدير ═══
# export_all_influences=False (الافتراضي) يقصّ على 4 تأثيرات ويعيد
# التطبيع. قسته: 195 رأساً فقط (0.39٪) فوق الأربعة، ومتوسط الوزن
# المفقود بهالرؤوس 0.31٪ وأسوأ رأس 1.99٪ — مهمَل بالقياس.
bpy.ops.export_scene.gltf(
    filepath=OUT, export_format='GLB',
    export_animation_mode='ACTIONS', export_nla_strips=True,
    export_skins=True, export_morph=False, export_apply=False,
)
print("\nصدّرت:", OUT)
