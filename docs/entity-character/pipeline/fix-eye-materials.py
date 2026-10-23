#!/usr/bin/env python3
"""إصلاح خامات العين بملف GLB جاي من Character Creator.

═══════════════════════════════════════════════════════════════════
ليش موجود
═══════════════════════════════════════════════════════════════════

مالك النظام شاف الشخصية وكال: **«أكو مشكلة بالعيون»**. وفحصت الملف
وطلعت عيبان **حقيقيان**، وكلاهما **من سكربت تحويلنا** مو من المجسّم:

① `Std_Eye_Occlusion_R/L` — جسم «ظل المحجر» بـCharacter Creator.
   طلع `alphaMode=BLEND` بلون **أسود** و**شفافية = ١** وبلا أي صورة.
   يعني **قبّة سوداء صلبة** فوق العين. وهذا الي يشوفه المستخدم:
   عيون مطفية أو حلقة سودة. وسببه إن سكربتنا يبني الخامات بلاحقة
   اسم الصورة، وهالخامة **ما عدها صورة** فبقى اللون الأسود الافتراضي
   بشفافية كاملة.

② `Std_Cornea_R/L` — القرنية **غشاء شفاف** فوق القزحية، وطلعت
   `OPAQUE`. يعني **قرص مطفي يغطي القزحية** فالنظرة تموت.

⚠️ **وشي ظنّيته عيباً وما هو**: `Std_Eye_L` يستخدم
`Std_Eye_R_Diffuse`. فحصت صور الملف كلها (٣٧ صورة) و**ماكو صورة عين
يسرى أصلاً** — Character Creator يشارك نفس الصورة للعينين. فهذا
**سليم**، وشِلت الادعاء.

═══════════════════════════════════════════════════════════════════
ليش رقعة على الـGLB ومو إعادة تصدير من بلندر
═══════════════════════════════════════════════════════════════════

إعادة التصدير تعني إعادة كل خط التحويل (٤٧.٥ ← ٥.٥ م.ب، دقائق
طويلة) وإعادة كل قرار خامات — **ومخاطرة رجوع عيوب انصلحت** (اللوح
الوردي، الجسم المبقّع). والعيبان هنا **بالبيانات الوصفية بس**: خانتا
شفافية وخانة نمط. فالرقعة **أدق وأصغر أثراً**، والهندسة والصور ما
تُلمس ولا بايت.

⚠️ **ونخفي بالشفافية مو بحذف الجسم**: حذف جسم من GLB يعني تنظيف
`accessors` و`bufferViews` و`nodes` والمؤشرات كلها — وأي مؤشر يبقى
غلط يطيّح التحميل **بخطأ غامض** على شاشة موظف. والشفافية صفر تنطي
**نفس النتيجة البصرية** بلا هالمخاطرة.

الاستخدام:
    python3 fix-eye-materials.py المدخل.glb المخرج.glb
"""

import json
import struct
import sys

JSON_CHUNK = 0x4E4F534A
BIN_CHUNK = 0x004E4942


def read_glb(path):
    """يرجّع (وصف JSON، مقطع البيانات) — بلا أي مساس بالبايتات."""
    data = open(path, "rb").read()
    if data[:4] != b"glTF":
        raise SystemExit("مو ملف GLB — البصمة غلط")
    off, meta, binary = 12, None, b""
    while off < len(data):
        length, kind = struct.unpack("<II", data[off : off + 8])
        chunk = data[off + 8 : off + 8 + length]
        if kind == JSON_CHUNK:
            meta = json.loads(chunk)
        elif kind == BIN_CHUNK:
            binary = chunk
        off += 8 + length
    if meta is None:
        raise SystemExit("ماكو مقطع JSON بالملف")
    return meta, binary


def write_glb(path, meta, binary):
    """يكتب GLB سليم الحشو — الحشو **إلزامي** بالمواصفة: أي مقطع
    لازم يكون طوله من مضاعفات ٤، وبلاه محمّلات تصارم ترفض الملف."""
    js = json.dumps(meta, separators=(",", ":"), ensure_ascii=False).encode()
    js += b" " * ((4 - len(js) % 4) % 4)
    bn = binary + b"\0" * ((4 - len(binary) % 4) % 4)
    total = 12 + 8 + len(js) + (8 + len(bn) if bn else 0)
    out = bytearray()
    out += b"glTF" + struct.pack("<II", 2, total)
    out += struct.pack("<II", len(js), JSON_CHUNK) + js
    if bn:
        out += struct.pack("<II", len(bn), BIN_CHUNK) + bn
    open(path, "wb").write(out)


def fix(meta):
    """يصلح الخامتين، ويرجّع سجل الي تغيّر حتى يُقرأ بالمراجعة."""
    log = []
    for mat in meta.get("materials", []):
        name = mat.get("name", "")
        pbr = mat.setdefault("pbrMetallicRoughness", {})

        # ① ظل المحجر: نخفيه بالكامل (شفافية صفر).
        if "Eye_Occlusion" in name:
            before = pbr.get("baseColorFactor")
            mat["alphaMode"] = "BLEND"
            pbr["baseColorFactor"] = [0, 0, 0, 0]
            log.append(f"{name}: اللون {before} ← [0,0,0,0] (كان أسود صلب)")

        # ② القرنية: غشاء شفاف لمّاع، مو قرص مطفي.
        #    ⚠️ الشفافية **٠.١٢ مو صفر**: القرنية تنطي اللمعة الي
        #    تخلي العين «حيّة». صفر يشيل اللمعة كلها.
        elif "Cornea" in name:
            before = mat.get("alphaMode", "OPAQUE")
            mat["alphaMode"] = "BLEND"
            pbr["baseColorFactor"] = [1, 1, 1, 0.12]
            pbr["roughnessFactor"] = 0.05
            pbr["metallicFactor"] = 0
            log.append(f"{name}: {before} ← BLEND بشفافية ٠.١٢ ولمعان عالي")

        # ③ خط الدمع: شفافيته ٠.٠٥ أصلاً — سليم، ما نلمسه.
    return log


def main():
    if len(sys.argv) != 3:
        raise SystemExit(__doc__)
    meta, binary = read_glb(sys.argv[1])
    log = fix(meta)
    if not log:
        raise SystemExit("ولا خامة عين انلگت — تأكد إن الملف من Character Creator")
    write_glb(sys.argv[2], meta, binary)
    print("\n".join(log))
    print(f"\nانكتب: {sys.argv[2]}")


if __name__ == "__main__":
    main()
