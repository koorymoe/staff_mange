import { useState, useCallback } from 'react'

// ═══ حارس الحفظ ═══
//
// النمط الي كان متكرر بعشر شاشات:
//
//     const doSomething = async () => {
//       await api.updateX(...)   // ← بلا try
//       reload()
//     }
//
// لمن يفشل الطلب — انقطاع، صلاحية ناقصة، سيرفر يعيد التشغيل — الوعد
// يترفض بلا ما يمسكه أحد. ما تطلع رسالة، وما ينتغيّر شي بالشاشة،
// والموظف يظن الشغلة انحفظت ويكمّل.
//
// **الحفظ الي يفشل بصمت أخطر من الي يفشل بصوت**: الي يفشل بصوت
// يعيده الموظف، والصامت يمشي وياه لآخر اليوم ويتراكم غلط بالبيانات.
//
// بدل ما نكتب try/catch بكل دالة (وننسى بوحدة)، هذا الخطّاف يوحّد
// السلوك: يمسك الخطأ، يعرضه، ويرجّع نجح/فشل حتى المستدعي يقرر.

export type SaveGuard = {
  /** آخر خطأ — اعرضه بالشاشة. */
  error: string | null
  /** true وقت تنفيذ العملية — لتعطيل الأزرار. */
  busy: boolean
  /** امسح الرسالة (زر «إخفاء»). */
  clear: () => void
  /**
   * نفّذ العملية بحماية.
   *
   * @param what وصف العملية بالعربي — يطلع بالرسالة: «تعذر {what}»
   * @returns القيمة عند النجاح، أو undefined عند الفشل
   */
  run: <T>(what: string, fn: () => Promise<T>) => Promise<T | undefined>
}

export function useSaveGuard(): SaveGuard {
  const [error, setError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const run = useCallback(async <T,>(what: string, fn: () => Promise<T>): Promise<T | undefined> => {
    setBusy(true)
    setError(null)
    try {
      return await fn()
    } catch (e) {
      // ⚠️ نضم الوصف للرسالة: «تعذر الاتصال» وحدها ما تگول للموظف
      // أي عملية فشلت لما تكون الشاشة بيها عشر أزرار.
      setError(`تعذر ${what}: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
      return undefined
    } finally {
      setBusy(false)
    }
  }, [])

  const clear = useCallback(() => setError(null), [])

  return { error, busy, clear, run }
}
