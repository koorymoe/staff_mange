import { useEffect, useMemo, useRef, useState } from 'react'
import type { AvatarHandle, ClipName } from './entityAvatarEngine'
import { api, ensureFileToken, entityModelUrl } from '../api'

// ═══ شخصية الكيان فوق ورقة القصة ═══
//
// ⚠️ **الفن يركب فوق النسخة النصية ما يبدّلها**: لو فشل التحميل، أو
// كان المتصفح على `prefers-reduced-motion`، أو ما كان يدعم WebGL —
// هذا المكوّن **ما يعرض شي** والقصة تبقى مقروءة كاملة. الفشل هنا
// ما يمنع الموظف من قراءة تحذيره.
//
// ⚠️ والمحرّك يُجاب بـ`import()` — الحزمة ١.٢ م.ب مضغوطة، فما
// تنحمّل إلا لمّا تُفتح قصة فعلاً.

interface Props {
  /** المقطع المطلوب — يتغيّر مع مرحلة القصة أو مزاج الكيان. */
  clip: ClipName
  /** تُستدعى لمّا يفشل العرض، حتى الأب يعرف إنه على النصي. */
  onUnavailable?: () => void
  /** أصناف الإطار — للتحكم بالحجم من المستدعي. */
  className?: string
  /** خلفية الإطار. `transparent` للبوت العائم حتى يبين بلا مربّع. */
  background?: string
  /**
   * تأطير الكاميرا: `bust` للصدر-فوق (ورقة القصة)، و`full` لكل
   * الجسم (البوت العائم الصغير — بلاه تبين الرقبة بس).
   */
  framing?: 'bust' | 'full'
}

/**
 * ⚠️ الفحص **قبل** المحاولة: بلا WebGL بابل يرمي بشكل غامض، وأوضح
 * إننا نقع على النصي فوراً. ويُحسب خارج الـeffect حتى ما نضبط حالة
 * بشكل متزامن داخله.
 */
/** المجسّم المدمج — الاحتياطي الي **ما ينكسر أبداً**. */
const BUILTIN_MODEL = `${import.meta.env.BASE_URL}amani-tech-v4.glb`

/**
 * رابط المجسّم الي يُعرض لهذا الموظف.
 *
 * 🔴 **ليش هالدالة موجودة أصلاً**: قبلها كان اسم الملف مكتوباً
 * **ثابتاً** هنا، ومعناها إن أي مجسّم يُرفَع من داخل النظام **ما يوصل
 * لولا موظف** — يبين بشاشة المختبر وحدها. ومالك النظام شكى إنه
 * «ما شاف الشخصية الجديدة أبداً بالنظام»، وكان محقاً: هذا المكوّن هو
 * **كل** ما يشوفه الموظف (الودجة العائمة وورقة القصة)، وهو ما كان
 * يعرف إن اكو شخصية جديدة أصلاً.
 *
 * ⚠️ **الترتيب إلزامي**: الوسم (`ensureFileToken`) **قبل** بناء
 * الرابط — `entityModelUrl` يقرأ الوسم من متغيّر بالوحدة، فلو انبنى
 * قبله يطلع رابط بلا وسم و`GET /api/files` يرد ٤٠١.
 *
 * 🔴 **وكل مسار فشل يرجّع المدمج مو استثناءً**: هاي شاشة كل موظف.
 * ماكو نشط · وسم فشل · خادم أقدم من الواجهة (٤٠٤) · شبكة مقطوعة —
 * كلها تعني «اعرض القديم»، مو «لا تعرض شي».
 */
let modelUrlPromise: Promise<string> | null = null
function resolveModelUrl(): Promise<string> {
  // ⚠️ **الوعد يُخزَّن على مستوى الوحدة** لأن الودجة العائمة وورقة
  // القصة ممكن تتركّبان مع بعض: بلا خزن يصير نداءان لكل موظف بكل
  // تنقّل، وكلهم يرجّعون نفس الجواب.
  if (!modelUrlPromise) {
    modelUrlPromise = ensureFileToken()
      .then(() => api.getActiveEntityModel())
      .then((m) => (m ? entityModelUrl(m.fileKey) : BUILTIN_MODEL))
      .catch(() => BUILTIN_MODEL)
  }
  return modelUrlPromise
}

/** يُنسى المخزون لمّا يبدّل (ع) الشخصية، حتى التحميل الجاي يجيب الجديد. */
export function forgetActiveModel(): void {
  modelUrlPromise = null
}

function webglSupported(): boolean {
  try {
    return !!document.createElement('canvas').getContext('webgl2')
      || !!document.createElement('canvas').getContext('webgl')
  } catch {
    return false
  }
}

export default function EntityAvatar({
  clip, onUnavailable, className = 'h-52 w-full',
  background = 'linear-gradient(180deg,#0f2a4a 0%,#16395f 100%)',
  framing = 'bust',
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const handleRef = useRef<AvatarHandle | null>(null)
  const hasWebGL = useMemo(() => webglSupported(), [])
  const [ready, setReady] = useState(false)
  const [failed, setFailed] = useState(!hasWebGL)

  useEffect(() => {
    if (!hasWebGL) { onUnavailable?.(); return }
    let cancelled = false
    const canvas = canvasRef.current
    if (!canvas) return

    // المحرّك ورابط المجسّم بالتوازي — الاثنان مستقلان، والتسلسل
    // يضيف رحلةَ شبكة على وقت ظهور الشخصية بلا فائدة.
    Promise.all([import('./entityAvatarEngine'), resolveModelUrl()])
      .then(([{ mountAvatar }, url]) => mountAvatar(canvas, url, clip, framing))
      .then((h) => {
        if (cancelled) { h.dispose(); return }
        handleRef.current = h
        setReady(true)
      })
      .catch(() => {
        if (cancelled) return
        setFailed(true)
        onUnavailable?.()
      })

    return () => {
      cancelled = true
      handleRef.current?.dispose()
      handleRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // تغيير المقطع بلا إعادة بناء المشهد — إعادة البناء تعني تحميل
  // ٤ م.ب من جديد كل مرحلة.
  useEffect(() => {
    handleRef.current?.play(clip)
  }, [clip])

  if (failed) return null

  return (
    <div
      className={`relative overflow-hidden ${className}`}
      style={{ background }}
      data-avatar-state={ready ? 'ready' : 'loading'}
    >
      <canvas
        ref={canvasRef}
        className={`h-full w-full transition-opacity duration-500 ${ready ? 'opacity-100' : 'opacity-0'}`}
        aria-hidden
      />
    </div>
  )
}
