import { useEffect, useState, type DependencyList } from 'react'

// ═══ (DRY 11-02) جلب بيانات الشاشة ═══
// چان مكتوب بـ١١٠ ملف بنفس الشكل: alive + then + catch + setErr. هنا مرة وحدة.
// • data=null لحد ما يوصل الرد. • النجاح يمسح الخطأ. • الرد المتأخر بعد
//   ما الشاشة تتسكّر (أو deps تتغير) ينبلع — نفس حماية alive.
// • enabled=false: ما يجلب (مثلاً بلا صلاحية — حتى ما تنسجل مخالفة).
export function useApi<T>(fn: () => Promise<T>, deps: DependencyList, opts: { fallback?: string; enabled?: boolean } = {}) {
  const [data, setData] = useState<T | null>(null)
  const [error, setError] = useState('')
  const enabled = opts.enabled ?? true
  useEffect(() => {
    if (!enabled) return
    let alive = true
    fn().then((d) => { if (alive) { setData(d); setError('') } })
      .catch((e) => { if (alive) setError(e instanceof Error ? e.message : (opts.fallback ?? 'تعذر التحميل')) })
    return () => { alive = false }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, ...deps])
  return { data, error, setData }
}
