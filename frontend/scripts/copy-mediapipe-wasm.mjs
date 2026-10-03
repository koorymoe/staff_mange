// ═══════════════════════════════════════════════════════════════════
// ينسخ WASM مال ميدياپايپ من node_modules لـpublic/mp/wasm
// ═══════════════════════════════════════════════════════════════════
//
// ⚠️ **ليش ننسخها بدل ما نجيبها من CDN گوگل؟**
//   قيد (ع) الصريح: **صور الموظفين ما تطلع لأي مزوّد خارجي**. ولو
//   حمّلنا الـWASM من `cdn.jsdelivr.net` وقت التشغيل، فكل موظف يفتح
//   الكاميرا **يعلن لطرف ثالث** إنه يستخدم تتبّع وجه — مو الصورة
//   بس، بل واقعة الاستخدام نفسها. فنستضيفها عندنا.
//
// ⚠️ **وليش نسخة مو گيت؟**
//   الملفات **٢٣ م.ب**. لو دخلت گيت تبقى بتاريخه للأبد وتثقّل كل
//   استنساخ. وهي أصلاً موجودة بـ`node_modules` بعد `npm ci` —
//   فالنسخ وقت البناء أنضف، ولا يحتاج إنترنت إضافي.
import { copyFileSync, mkdirSync, existsSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const from = join(here, '..', 'node_modules', '@mediapipe', 'tasks-vision', 'wasm')
const to = join(here, '..', 'public', 'mp', 'wasm')

// النسختان مقصودتان: `internal` تستخدم SIMD وهي الأسرع، و
// `nosimd` احتياط للأجهزة القديمة — وميدياپايپ يختار بنفسه أيهما
// يحمّل حسب قدرة المتصفح. حذف الاحتياط يكسر الأجهزة القديمة صامتاً.
const files = [
  'vision_wasm_internal.js',
  'vision_wasm_internal.wasm',
  'vision_wasm_nosimd_internal.js',
  'vision_wasm_nosimd_internal.wasm',
]

if (!existsSync(from)) {
  console.error(`[mediapipe] ماكو مجلد wasm بـ${from} — شغّل npm install أول.`)
  process.exit(1)
}
mkdirSync(to, { recursive: true })
for (const f of files) copyFileSync(join(from, f), join(to, f))
console.log(`[mediapipe] نُسخت ${files.length} ملفات لـpublic/mp/wasm`)
