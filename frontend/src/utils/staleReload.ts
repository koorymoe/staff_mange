import { lazy, type ComponentType } from 'react'

// ═══ (دازاي 11-02) «النظام يوقف — أي زر أضغطه ما يوديني إلا أحدّث» ═══
// كل شاشة ملف لحاله واسمه بيه بصمة. بعد كل رفع، الملفات القديمة تنمسح من
// السيرفر، والمتصفح المفتوح من قبل يطلب القديم → 404 → التنقل يعلگ.
// الحل: إذا ملف شاشة ما انلگه، نحدّث الصفحة **مرة وحدة** (تجيب النسخة
// الجديدة). العلامة تمنع حلقة تحديث لو المشكلة شي ثاني.

const KEY = 'stale-reload-at'

/** يحدّث الصفحة إذا ما حدّثناها بآخر ٢٠ ثانية. يرجّع false لو ما حدّث. */
export function reloadOnceForStale(): boolean {
  try {
    const last = Number(sessionStorage.getItem(KEY) || 0)
    if (Date.now() - last < 20_000) return false
    sessionStorage.setItem(KEY, String(Date.now()))
  } catch { /* بلا تخزين: نحدّث مرة وحدة بس، والجاية ترمي الخطأ */ }
  window.location.reload()
  return true
}

export function installStaleChunkReload() {
  window.addEventListener('vite:preloadError', (e) => {
    if (reloadOnceForStale()) e.preventDefault()
  })
}

/** نفس lazy، بس ملف شاشة قديم ينحل بتحديث بدل ما يعلگ التنقل. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function lazyRetry<T extends ComponentType<any>>(load: () => Promise<{ default: T }>) {
  return lazy(() => load().catch((err) => {
    if (reloadOnceForStale()) return new Promise<{ default: T }>(() => {}) // الصفحة راح تتحدث
    throw err // فشل بعد التحديث: يطلع للـErrorBoundary برسالة بدل التعليق
  }))
}
