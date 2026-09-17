import { useEffect, useMemo, useRef, useState } from 'react'
import type { AvatarHandle, ClipName } from './entityAvatarEngine'
import { BUILTIN_MODEL, forgetActiveModel, resolveModelUrl } from './entityModelSource'

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
  /**
   * تُستدعى بأسماء المقاطع الموجودة فعلاً بالملف.
   *
   * 🔴 **ليش المستدعي يحتاجها**: المشي بالصفحة **ما ينفع بلا مقطع
   * مشي** — الشخصية تنزلق مثل قطعة أثاث. وبعض مجسّماتنا فيها ١١
   * مقطعاً وبعضها **صفر** (مقاس من الملفات). فالمستدعي يسأل قبل ما
   * يشغّل المشي، بدل ما نوهم بحركة ما موجودة.
   */
  onClips?: (clips: readonly string[]) => void
}

/**
 * ⚠️ الفحص **قبل** المحاولة: بلا WebGL بابل يرمي بشكل غامض، وأوضح
 * إننا نقع على النصي فوراً. ويُحسب خارج الـeffect حتى ما نضبط حالة
 * بشكل متزامن داخله.
 */
/**
 * المجسّم المدمج — الافتراضي والاحتياطي الي **ما ينكسر أبداً**.
 *
 * 🔴 **v5 من (م)** (`docs/entity-character/reviews/2026-09-12-consolidated`):
 * **١١ مقطع حركة بأسمائنا بالضبط** وتسمية ميكسامو الي محرّكنا يقراها،
 * ٤١ مفصلاً بثمانية تأثيرات، ٥.٩ م.ب — مفحوصة من الملف نفسه.
 *
 * ⚠️ **وليش مدمج مو مرفوعاً**: هذا الملف **داخل المستودع**، فيوصل
 * السيرفر بـ«سحب وبناء» بلا رفع ولا نسخ يدوي. وهاي بالضبط الي انعقدت:
 * المجسّم الثاني (كارتون بوي) **مستثنى من گيت** برخصته، فما وصل سيرفر
 * المالك إطلاقاً — وهو يشوف القديمة ويحسب إن التحديث ما نزل.
 *
 * ⚠️ **وv4 تبقى بالمستودع**: نسخة مبنية قديمة بمتصفح موظف ممكن
 * تطلبها، وحذف الملف يعني ٤٠٤ على شاشته.
 */



function webglSupported(): boolean {
  try {
    return !!document.createElement('canvas').getContext('webgl2')
      || !!document.createElement('canvas').getContext('webgl')
  } catch {
    return false
  }
}

export default function EntityAvatar({
  clip, onUnavailable, onClips, className = 'h-52 w-full',
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
    //
    // 🔴 **والفشل على مجسّم مرفوع يرجع للمدمج مو يخفي الكيان**:
    // قبلها، أي فشل بتحميل المجسّم النشط كان يشيل الشخصية **من كل
    // الشاشات** — والسبب ممكن يكون بسيطاً وخارج إيدنا: الملف انمحى
    // من قرص السيرفر، أو الوسم انتهى، أو خادم أقدم من الواجهة.
    // وهاي انكشفت بالقياس فعلاً: صف نشط يشير لملف مو موجود ← ٤٠٤
    // ← الكيان **اختفى بالكامل** بلا سبب ظاهر للموظف ولا للمالك.
    //
    // ⚠️ والمحاولة الثانية **بالمدمج حصراً ومرة وحدة**: إعادة نفس
    // الرابط تعيد نفس الفشل، ودورة محاولات على شاشة موظف أسوأ من
    // الفشل نفسه.
    Promise.all([import('./entityAvatarEngine'), resolveModelUrl()])
      .then(([{ mountAvatar }, url]) =>
        mountAvatar(canvas, url, clip, framing).catch((err: unknown) => {
          if (cancelled || url === BUILTIN_MODEL) throw err
          // ⚠️ ننسى المخزون حتى المحاولة الجاية تسأل الخادم من جديد
          // (ممكن (ع) رجّع الملف أو بدّل النشط بينها وبين الآن).
          forgetActiveModel()
          console.warn('[entity] فشل المجسّم المرفوع — رجعنا للمدمج', err)
          return mountAvatar(canvas, BUILTIN_MODEL, clip, framing)
        }))
      .then((h) => {
        if (cancelled) { h.dispose(); return }
        handleRef.current = h
        setReady(true)
        onClips?.(h.stats.clips)
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
