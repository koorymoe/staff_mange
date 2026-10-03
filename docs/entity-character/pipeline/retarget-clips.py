"""نقل مقاطع الحركة من هيكل ميكسامو لهيكل Rigify.

═══════════════════════════════════════════════════════════════════
ليش موجود
═══════════════════════════════════════════════════════════════════

عندنا **نصفان من حل واحد**:
- **أماني v5** (شغل م، ١٢ أيلول): **١١ مقطع حركة** بأسمائنا بالضبط،
  بس **صفر عظمة عين وصفر تعبير** — فمستحيل ترمش أو تتبع المؤشر.
- **شخصية (م) الجديدة** (١٤ أيلول): **عظام عيون + ٣٠ تعبيراً**
  وشكل أحلى، بس **ماكو ولا مقطع** من مقاطع نظامنا (بس تلويحة).

ومالك النظام يريد **الاثنين**: شكل جديد **ويمشي ويحچي ويحذّر**.
فهذا السكربت ينقل المقاطع الـ١١ من الأول للثاني.

═══════════════════════════════════════════════════════════════════
لماذا النقل بـ«فرق وضعية الراحة» مو بنسخ الزوايا
═══════════════════════════════════════════════════════════════════

🔴 **نسخ الدوران الخام يطلّع وضعيات مكسورة**، والسبب مقاس: وضعية
راحة كارتون بوي أذرعها مفتوحة **٥٩.٢°** عن العمود، وميكسامو **وقفة
T** (٩٠°). فلو نسخنا دوران عظمة الذراع كما هو، تنضاف الفرق ٣٠° على
كل إطار والذراع تطلع برّا الجسم.

**الصح**: نحسب فرقاً **ثابتاً** بين وضعيتَي الراحة لكل عظمة:

    offset = restMixamo⁻¹ · restRigify

ثم لكل إطار:

    poseRigify(عالمي) = poseMixamo(عالمي) · offset

يعني ننقل **التغيّر** عن وضعية الراحة، مو الزاوية المطلقة. وهيچ
الشخصية تتحرّك نفس حركة المصدر وتبقى بوضعيتها الطبيعية.

⚠️ **والكتابة على عظام التحكّم (FK) مو على عظام التشويه**: عظام
`DEF-` بـRigify **مقيَّدة** (تتبع عظام أخرى)، فالكتابة عليها تنمحي.
والتصدير يخبز النتيجة على `DEF-` تلقائياً (انظر `blender-to-web.py`).

⚠️ **ومفاتيح FK/IK لازم تنقلب لـFK**: Rigify يجي افتراضياً على IK
للأطراف، فأي مفتاح FH نكتبه **ما يبيّن** لأن الـIK يفوز.

الاستخدام:
    python3.11 retarget-clips.py -- الهدف.blend المصدر.glb المخرج.blend
"""

import math
import sys

import bpy
from mathutils import Matrix

# ═══ الخريطة: ميكسامو ← عظمة تحكّم Rigify ═══
#
# ⚠️ **بعظام التحكّم بالاسم الصريح** — فحصت وجودها كلها بالهيكل
# (٧٢٤ عظمة، منها ٢٤ FK). وأي اسم ناقص يُبلَّغ بالسجل بدل ما يُتجاهل
# بهدوء: الخطأ الصامت هو الي خلّى أول محاولة ترجّع صفر أصابع وصفر
# تعابير بمجسّم فيه ٣٠ من كل واحد.
BONE_MAP = {
    'mixamorig:Hips': 'torso',
    'mixamorig:Spine': 'spine_fk.001',
    'mixamorig:Spine1': 'spine_fk.002',
    'mixamorig:Spine2': 'spine_fk.003',
    'mixamorig:Neck': 'neck',
    'mixamorig:Head': 'head',
}
for side, sfx in (('Left', 'L'), ('Right', 'R')):
    BONE_MAP.update({
        f'mixamorig:{side}Shoulder': f'shoulder.{sfx}',
        f'mixamorig:{side}Arm': f'upper_arm_fk.{sfx}',
        f'mixamorig:{side}ForeArm': f'forearm_fk.{sfx}',
        f'mixamorig:{side}Hand': f'hand_fk.{sfx}',
        f'mixamorig:{side}UpLeg': f'thigh_fk.{sfx}',
        f'mixamorig:{side}Leg': f'shin_fk.{sfx}',
        f'mixamorig:{side}Foot': f'foot_fk.{sfx}',
        f'mixamorig:{side}ToeBase': f'toe_fk.{sfx}',
    })

# المقاطع الي نظامنا يطلبها بالاسم — أي اسم ثاني ما ينفع.
WANTED = [
    'WALK_TO_TARGET', 'WALK_ALT', 'RUN_TO_EDGE', 'SPEAK', 'SPEAK_ALT',
    'CELEBRATE', 'SIT_IDLE', 'SIT_SPEAK', 'LAYING_IDLE', 'SHOW_WARNING',
    'POINT_AT_UI',
]

# الحوض وحده يحمل إزاحة — وباقي العظام دوران بس.
ROOT_SRC = 'mixamorig:Hips'

# ═══ طريقة النقل **لكل عظمة** ═══
#
# 🔴 **وليش مو طريقة واحدة للكل**: قِست الاثنتين وكل واحدة تنجح
# بمكان وتفشل بمكان:
#
# · **المطلق** (ننسخ الاتجاه العالمي كما هو) — صح للأطراف: فرق
#   **٠.٠°** بالذراع والساعد والفخذ. بس على `torso` (جذر الجسم كله
#   بـRigify) **يرفع الجسم**: قمة الرأس نزلت من `y = +٦٤` بكسل إلى
#   **`−٢٣`** — يعني برّا الكانفس، والرأس ينقص من فوق.
#
# · **الفرق** (`restSrc⁻¹ · restRig`) — تحافظ على اتجاه راحة الهدف
#   فما تنقل الجسم. بس على الأطراف تخلّي **فرق وضعيتَي الراحة
#   مضافاً على كل إطار**: الذراع **٣٧.٤°** والساعد **٥٣.٤°**،
#   فالأذرع تبقى مفتوحة مثل تمثال.
#
# فالخلاصة: **الحوض بالفرق، والباقي بالمطلق**. والحوض ماكو عنده
# «فتحة راحة» مثل الذراع فالـ٣٧° ما تظهر عليه.
DELTA_BONES = {'mixamorig:Hips'}


def log(*a):
    print('[retarget]', *a)


def curve_count(act):
    """عدد مسارات الحركة — **يشتغل على بنيتَي بلندر**.

    ⚠️ بلندر ٥ شال `action.fcurves` وحوّلها لطبقات/شرائح/حزم قنوات
    (`layers → strips → channelbags`). والاعتماد على القديم يطيّح
    السكربت بعد ما يخلّص كل الشغل الثقيل — صارت."""
    try:
        return len(act.fcurves)
    except AttributeError:
        n = 0
        for layer in getattr(act, 'layers', []):
            for strip in getattr(layer, 'strips', []):
                for cb in getattr(strip, 'channelbags', []) or []:
                    n += len(cb.fcurves)
        return n


def world_rot(obj_matrix, bone_matrix):
    """دوران **عالمي مطبَّع** لعظمة.

    🔴 **والتطبيع والفضاء العالمي كلاهما إلزامي**، والسبب مقاس:
    هيكل المصدر يُستورَد **بمقياس ٠.٠١ وبدوران ٩٠° حول X**
    (تحويل Y-up إلى Z-up مال glTF). فحساب الفرق بـ«فضاء الهيكل»
    يعني حساباً بفضاء **مقلوب ومصغَّر** — والنتيجة كانت: الحوض
    **يطير** (رأس الشخصية طلع فوق الكانفس: y = −١٥٦ ثم −٣١٠ ثم
    −٤٤٠)، وبعد ما شِلت الإزاحة بقت الأذرع **مفتوحة مثل وضعية
    الراحة** وحركة اليد **٥.٩ سم بثانية كاملة** بدل مشية حقيقية.
    """
    m = (obj_matrix @ bone_matrix).to_3x3()
    for c in range(3):
        m.col[c].normalize()
    return m


def pick_armatures():
    arms = [o for o in bpy.data.objects if o.type == 'ARMATURE']
    rig = next((o for o in arms if 'Rigify' in o.name), None)
    src = next((o for o in arms if o is not rig), None)
    if not rig or not src:
        raise SystemExit('ما لگيت هيكلين — تأكد إن الاستيراد نجح')
    return rig, src


def force_fk(rig):
    """يقلب مفاتيح Rigify لـFK.

    ⚠️ بلا هذا، مفاتيح FK **ما تبيّن**: الـIK يفوز على السلسلة،
    فالذراع تبقى واقفة مع إن البيانات مكتوبة."""
    n = 0
    for pb in rig.pose.bones:
        for key in ('IK_FK', 'IK/FK'):
            try:
                if key in pb:
                    pb[key] = 1.0      # ١ = FK كامل
                    n += 1
            except Exception:
                pass
    log(f'مفاتيح FK/IK المقلوبة: {n}')


def rest_offsets(rig, src):
    """طريقة النقل لكل زوج: فرقٌ للحوض، و`None` (مطلق) للباقي.

    انظر `DELTA_BONES` للأرقام الي قادت هالتقسيم."""
    out, missing = {}, []
    for sname, rname in BONE_MAP.items():
        sb = src.data.bones.get(sname)
        rb = rig.data.bones.get(rname)
        if not sb or not rb:
            missing.append(f'{sname}→{rname}')
            continue
        # الفرق يُحسب للعظام الي تحتاجه (انظر `DELTA_BONES`)،
        # و`None` تعني «انسخ الاتجاه المطلق».
        if sname in DELTA_BONES:
            rest_src = world_rot(src.matrix_world, sb.matrix_local)
            rest_rig = world_rot(rig.matrix_world, rb.matrix_local)
            out[sname] = (rname, rest_src.inverted() @ rest_rig)
        else:
            out[sname] = (rname, None)
    if missing:
        log('⚠️ أزواج ناقصة (تُتجاهل بوضوح مو بهدوء):', missing)
    log(f'أزواج جاهزة: {len(out)} من {len(BONE_MAP)}')
    return out


def hip_scale(rig, src):
    """نسبة طول الساق — الإزاحة لازم تتقاس بمقياس الهدف.

    ⚠️ بلاها المشي يطلع **قفزات** أو **زحف**: المصدر والهدف مو بنفس
    الطول، والإزاحة الخام تعني خطوة أطول أو أقصر من الساق."""
    # ⚠️ **بالإحداثيات العالمية مو المحلية**: `matrix_local` للعظمة
    # بفضاء الهيكل، وهياكل الـglTF المستوردة تجي **بمقياس كائن**
    # مختلف (وقياسي أول مرة طلع **−١٥٩** بسببها — رقم بلا معنى،
    # وسالب يعني مقلوب). فنقيس **ارتفاع الحوض عن الأرض بالعالم**.
    s = src.data.bones.get(ROOT_SRC)
    r = rig.data.bones.get('torso')
    if not s or not r:
        return 1.0
    sz = (src.matrix_world @ s.matrix_local.translation).z
    rz = (rig.matrix_world @ r.matrix_local.translation).z
    log(f'ارتفاع الحوض: المصدر {sz:.3f} م · الهدف {rz:.3f} م')
    if abs(sz) < 1e-4:
        return 1.0
    k = rz / sz
    # حدّ عقلاني: الشخصيتان بشريتان بطول ~١.٧ م، فأي نسبة برّا
    # النطاق تعني قياساً غلط مو فرق طول — فنرجع ١ بدل ما نخرّب.
    if not (0.4 < k < 2.5):
        log(f'⚠️ نسبة غير معقولة ({k:.3f}) — نستخدم ١ بلا تحجيم')
        return 1.0
    log(f'مقياس الإزاحة: {k:.4f}')
    return k


def retarget(rig, src, offsets, scale, action, name):
    """يبني حركة جديدة على الهيكل الهدف باسم `name`."""
    src.animation_data_create()
    src.animation_data.action = action
    f0, f1 = (int(round(x)) for x in action.frame_range)
    rig.animation_data_create()
    new = bpy.data.actions.new(name)
    rig.animation_data.action = new
    if hasattr(rig.animation_data, 'action_slot'):
        # بلندر ٤.٤+ يحتاج «خانة» للحركة، وبلاها ما تنكتب المفاتيح
        try:
            rig.animation_data.action_slot = new.slots.new(id_type='OBJECT', name='Object')
        except Exception:
            pass
    root_rest = src.data.bones[ROOT_SRC].matrix_local.translation if ROOT_SRC in src.data.bones else None

    for f in range(f0, f1 + 1):
        bpy.context.scene.frame_set(f)
        for sname, (rname, off) in offsets.items():
            spb = src.pose.bones.get(sname)
            rpb = rig.pose.bones.get(rname)
            if not spb or not rpb:
                continue
            # الاتجاه المطلق للمصدر هو نفسه للهدف (انظر `rest_offsets`)
            target_world = world_rot(src.matrix_world, spb.matrix)
            if off is not None:
                target_world = target_world @ off
            # ثم نرجعه لفضاء الهيكل الهدف (بابل/بلندر يبني المحلي منه)
            rig_rot = rig.matrix_world.to_3x3().inverted() @ target_world
            for c in range(3):
                rig_rot.col[c].normalize()
            m = rig_rot.to_4x4()
            # 🔴 **ولا إزاحة للحوض إطلاقاً — دوران بس.**
            #
            # جرّبت نقل الإزاحة أول مرة، والنتيجة مقاسة: **رأس
            # الشخصية طلع فوق الكانفس** (y بالبكسل: −١٥٦ ثم −٣١٠ ثم
            # −٤٤٠) يعني الجسم **يطير ويرتجف عمودياً** بدل يمشي.
            # والسبب إن هيكل المصدر المستورد بمقياس كائن مختلف،
            # فالإزاحة الخام تُقاس بوحدات ما تناسب الهدف.
            #
            # **ومو خسارة**: محرّكنا **يحيّد إزاحة الجذر أصلاً**
            # (`neutralizeRootMotion`) ويمشّي الشخصية برمجياً على
            # عقدة أب — يعني إزاحة المقطع كانت تُلغى بأي حال،
            # وحركة الأرجل (الي تبيّن المشي) محفوظة كاملة.
            m.translation = rpb.matrix.translation
            rpb.matrix = m
            rpb.keyframe_insert('rotation_quaternion', frame=f)
        # ⚠️ **تحديث المشهد كل إطار إلزامي**: مصفوفات الوضع تُحسب
        # متأخرة، وبلا التحديث نقرأ إطاراً قديماً — نفس الدرس الي
        # كلّفنا تذبذب ٨°–٢٠° بقياس الإشارة.
        bpy.context.view_layer.update()
    log(f'  {name}: إطارات {f0}–{f1}  مسارات {curve_count(new)}')
    # 🔴 **الحفظ بمسار NLA إلزامي، وبلاه يُصدَّر مقطع واحد بس.**
    #
    # قِستها: بنيت الـ١١ بنجاح (٩١ مساراً لكل واحد)، وبالملف الناتج
    # طلع **مقطع واحد** — الأخير، لأنه الي بقى «الحركة النشطة».
    # ومصدّر glTF يلگى الحركات من **مسارات NLA** أو من الحركة
    # النشطة، مو من مجرد وجودها بالملف.
    track = rig.animation_data.nla_tracks.new()
    track.name = name
    strip = track.strips.new(name, int(f0), new)
    strip.action_frame_start = f0
    strip.action_frame_end = f1
    # 🔴 **ولا كتم** — والسبب مقاس: كتمتها «حتى ما تتراكم بالمعاينة»،
    # والنتيجة إن التصدير **يخبز وضعية الراحة** بكل المقاطع: بالبلندر
    # الذراع تنزل صح (اليد ٠.٢٤ م من الجسم وتتحرّك ٣٠ سم)، وبالملف
    # الناتج الأذرع **مفتوحة** ومجموع بكسلات المشهد يتغيّر **١.٥٪
    # بس**. الخبز يقرأ حالة المشهد، والمكتوم = ماكو حركة.
    track.mute = False
    return new


def main():
    a = sys.argv[sys.argv.index('--') + 1:]
    if len(a) < 3:
        raise SystemExit(__doc__)
    target_blend, source_glb, out_blend = a[0], a[1], a[2]

    bpy.ops.wm.open_mainfile(filepath=target_blend)
    before = {o.name for o in bpy.data.objects}
    bpy.ops.import_scene.gltf(filepath=source_glb)
    imported = [o for o in bpy.data.objects if o.name not in before]

    rig, src = pick_armatures()
    log(f'الهدف: {rig.name} ({len(rig.data.bones)} عظمة) · المصدر: {src.name} ({len(src.data.bones)})')
    force_fk(rig)
    for pb in rig.pose.bones:
        pb.rotation_mode = 'QUATERNION'

    offsets = rest_offsets(rig, src)
    scale = hip_scale(rig, src)

    src_actions = {act.name: act for act in bpy.data.actions}
    made = []
    for name in WANTED:
        act = src_actions.get(name)
        if not act:
            log(f'⚠️ المقطع {name} مو موجود بالمصدر — يُتجاهل بوضوح')
            continue
        made.append(retarget(rig, src, offsets, scale, act, f'RT_{name}').name)

    # ⚠️ نشيل المستورد كله: مجسّم المصدر وهيكله ما لهم شغل بالمخرج،
    # وبقاؤهم يعني **شخصيتين** بالملف وحجماً مضاعفاً.
    for o in imported:
        bpy.data.objects.remove(o, do_unlink=True)

    # ⚠️ **وحركات المصدر تُحذف قبل إعادة التسمية**: حذف الكائن ما
    # يحذف حركاته، فإعادة التسمية تصطدم بالاسم الموجود ويطلع
    # `WALK_ALT.001` — واسم ما يطابق عقدنا **ما يشتغل** بالنظام.
    for nm in WANTED:
        act = bpy.data.actions.get(nm)
        if act:
            bpy.data.actions.remove(act)

    # نسمّي الحركات بأسمائها النهائية بعد ما انحذف المصدر (حتى ما
    # يتعارض الاسم مع حركة المصدر نفسها)
    for n in made:
        act = bpy.data.actions.get(n)
        if act:
            act.name = n[3:]
    # ⚠️ **نشيل الحركات الي ما يطلبها عقدنا**: كل حركة بهذا الهيكل
    # **٥١٣ قناة**، والزائدة (وقفات ثابتة + TPose + حركة جسم) تدخل
    # بالتصدير وتكبّر الملف بلا ما يستخدمها النظام ولا مرة.
    keep = set(n[3:] for n in made)
    dropped = []
    for act in list(bpy.data.actions):
        if act.name not in keep:
            dropped.append(act.name)
            bpy.data.actions.remove(act)
    if dropped:
        log(f'حركات محذوفة ({len(dropped)}): {dropped}')
    # ⚠️ ونشيل مسارات NLA المكتومة الي صارت بلا حركة
    for ob in bpy.data.objects:
        ad = ob.animation_data
        if not ad:
            continue
        for tr in list(ad.nla_tracks):
            if not any(st.action for st in tr.strips):
                ad.nla_tracks.remove(tr)
    log('الحركات المنقولة:', [n[3:] for n in made])
    bpy.ops.wm.save_as_mainfile(filepath=out_blend)
    log('انكتب:', out_blend)


if __name__ == '__main__':
    main()
