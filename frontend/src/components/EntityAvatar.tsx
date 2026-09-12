import { useEffect, useMemo, useRef, useState } from 'react'
import type { AvatarHandle, ClipName } from './entityAvatarEngine'

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
  /** المقطع المطلوب — يتغيّر مع مرحلة القصة. */
  clip: ClipName
  /** تُستدعى لمّا يفشل العرض، حتى الأب يعرف إنه على النصي. */
  onUnavailable?: () => void
}

/**
 * ⚠️ الفحص **قبل** المحاولة: بلا WebGL بابل يرمي بشكل غامض، وأوضح
 * إننا نقع على النصي فوراً. ويُحسب خارج الـeffect حتى ما نضبط حالة
 * بشكل متزامن داخله.
 */
function webglSupported(): boolean {
  try {
    return !!document.createElement('canvas').getContext('webgl2')
      || !!document.createElement('canvas').getContext('webgl')
  } catch {
    return false
  }
}

export default function EntityAvatar({ clip, onUnavailable }: Props) {
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

    import('./entityAvatarEngine')
      .then(({ mountAvatar }) =>
        mountAvatar(canvas, `${import.meta.env.BASE_URL}amani-tech-v4.glb`, clip))
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
      className="relative h-52 w-full overflow-hidden"
      style={{ background: 'linear-gradient(180deg,#0f2a4a 0%,#16395f 100%)' }}
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
