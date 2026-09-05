import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { handleSessionExpired, isTokenExpired } from './api'

// كل نص ساعة
const REFRESH_INTERVAL_MS = 30 * 60 * 1000

// نتتبع إذا المستخدم كتب شي فعلاً بهذي الصفحة. ما ننفع نعتمد على قيم الحقول
// نفسها: React يملي value برمجياً (تواريخ، بحث، قيم افتراضية) بينما
// defaultValue يضل فاضي — فتنحسب كلها "كتابة ما انحفظت" وتمنع أي تحديث.
// الحل: نسمع أحداث الكتابة الحقيقية (isTrusted) الجاية من المستخدم نفسه.
let userTyped = false

/**
 * useAutoRefresh يحدّث الصفحة تلقائياً كل نص ساعة، ويعالج مشكلة "النافذة
 * تضل مفتوحة ويرجعلها الموظف فيلكاها معلّقة وتطلعله الجلسة منتهية".
 *
 * ثلاث حالات:
 *  1. مرّت نص ساعة والصفحة شغالة → تحديث (إلا إذا المستخدم كاتب شي بيدّه).
 *  2. الموظف رجع للتبويب بعد غياب أكثر من نص ساعة → تحديث فوري، لأن هذي
 *     بالضبط الحالة الي كانت تطلع بيها المشكلة.
 *  3. التوكن منتهي فعلاً (نعرفها من داخل التوكن بدون أي طلب) → نروح لشاشة
 *     الدخول مباشرة برسالة واضحة، بدل ما ينطلق سيل طلبات فاشلة.
 */
export function useAutoRefresh(enabled: boolean) {
  const location = useLocation()

  // كل ما يتغير المسار نعتبر الشغل السابق انتهى (حفظ أو ترك الصفحة)
  useEffect(() => {
    userTyped = false
  }, [location.pathname])

  useEffect(() => {
    const onInput = (e: Event) => {
      if (!e.isTrusted) return
      const t = e.target as HTMLElement | null
      if (!t) return
      const tag = t.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') userTyped = true
    }
    document.addEventListener('input', onInput, true)
    document.addEventListener('change', onInput, true)
    return () => {
      document.removeEventListener('input', onInput, true)
      document.removeEventListener('change', onInput, true)
    }
  }, [])

  useEffect(() => {
    // ما نشتغل بشاشة الدخول: التحديث هناك يمحي رسالة "انتهت الجلسة" قبل ما
    // يقراها الموظف، ويقاطعه وهو يكتب بياناته.
    if (!enabled) return
    let hiddenSince: number | null = null

    const refreshIfPossible = () => {
      if (isTokenExpired()) {
        handleSessionExpired('انتهت صلاحية الجلسة، الرجاء تسجيل الدخول مجدداً')
        return
      }
      if (userTyped) return // نأجل — ما نضيّع شغل المستخدم
      window.location.reload()
    }

    const timer = setInterval(refreshIfPossible, REFRESH_INTERVAL_MS)

    const onVisibility = () => {
      if (document.hidden) {
        hiddenSince = Date.now()
        return
      }
      if (hiddenSince !== null && Date.now() - hiddenSince >= REFRESH_INTERVAL_MS) {
        // رجع بعد غياب طويل — نحدّث حتى لو كاتب شي، لأن الصفحة بهذي الحالة
        // بايتة أصلاً وبياناتها قديمة، والبديل إنه يلكاها معلّقة.
        if (isTokenExpired()) {
          handleSessionExpired('انتهت صلاحية الجلسة، الرجاء تسجيل الدخول مجدداً')
        } else {
          window.location.reload()
        }
      }
      hiddenSince = null
    }

    document.addEventListener('visibilitychange', onVisibility)
    return () => {
      clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [enabled])
}
