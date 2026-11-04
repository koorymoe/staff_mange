import { useEffect, useState } from 'react'
import { api, type PrivacyPolicyPoint } from '../api'

/**
 * PrivacyPolicyGate يعرض سياسة الخصوصية أول ما يسجل الموظف دخول، ويطلب منه
 * يقراها ويضغط "موافق". ما ينفتح مرة ثانية إلا إذا انضافت نقاط جديدة بعد
 * موافقته (السيرفر يقارن عدد النقاط وقت الموافقة بالعدد الحالي).
 *
 * ما يوقف النظام: لو فشل الطلب لأي سبب نعديه بصمت بدل ما نحجز الموظف بره.
 */
export default function PrivacyPolicyGate() {
  const [points, setPoints] = useState<PrivacyPolicyPoint[]>([])
  const [open, setOpen] = useState(false)
  const [accepting, setAccepting] = useState(false)
  const [readAll, setReadAll] = useState(false)
  const [hash, setHash] = useState('')
  const [changed, setChanged] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    api.getPrivacyPolicyStatus()
      .then((s) => {
        if (s.needsAcceptance && s.points.length > 0) {
          setPoints(s.points)
          setHash(s.hash ?? '')
          setChanged(Boolean(s.changed))
          setOpen(true)
        }
      })
      .catch(() => { /* ما نوقف الموظف لو تعذّر الجلب */ })
  }, [])

  const accept = async () => {
    setAccepting(true)
    setError('')
    try {
      await api.acceptPrivacyPolicy(hash)
      setOpen(false)
    } catch (e) {
      // النص تحدّث بالأثناء: نجيب النسخة الجديدة ويقراها من جديد
      const s = await api.getPrivacyPolicyStatus().catch(() => null)
      // وافق من تبويب ثاني أو انشالت النقاط — نسكّر بدل ما تعلگ
      if (s && (!s.needsAcceptance || s.points.length === 0)) {
        setOpen(false)
        return
      }
      if (s?.needsAcceptance) {
        setPoints(s.points)
        setHash(s.hash ?? '')
        setReadAll(false)
      }
      setError((e as Error).message || 'تعذر تسجيل الموافقة، حاول مرة ثانية')
    } finally {
      setAccepting(false)
    }
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-[3000] flex items-center justify-center bg-black/60 p-4" dir="rtl">
      <div className="flex max-h-[88vh] w-full max-w-2xl flex-col rounded-2xl bg-white shadow-2xl">
        <div className="border-b border-slate-100 p-6">
          <h2 className="text-xl font-bold text-brand-900">📜 مبادئ العمل وسياسة الخصوصية</h2>
          <p className="mt-1 text-sm text-slate-500">
            {changed
              ? 'تحدّثت المبادئ — اقراها من جديد ووافق عليها قبل ما تكمل.'
              : 'هاي مبادئ شركة الأماني الي نشتغل بيها كلنا. اقراها ووافق عليها قبل استخدام النظام.'}
          </p>
        </div>

        <div
          className="flex-1 overflow-auto p-6"
          onScroll={(e) => {
            // ما نفعّل زر الموافقة إلا بعد ما يوصل لآخر النص فعلاً
            const el = e.currentTarget
            if (el.scrollTop + el.clientHeight >= el.scrollHeight - 24) setReadAll(true)
          }}
          ref={(el) => {
            // لو النقاط قليلة وما اكو سكرول أصلاً، نعتبرها مقروءة
            if (el && el.scrollHeight <= el.clientHeight) setReadAll(true)
          }}
        >
          <ol className="space-y-3">
            {points.map((p, i) => (
              <li key={p.id} className="flex gap-3 rounded-xl bg-slate-50 p-4">
                <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand-100 text-sm font-bold text-brand-700">
                  {p.principleNo ?? i + 1}
                </span>
                <p className="text-slate-800">{p.principleNo ? p.content.replace(/^مبدأ \d+: /, '') : p.content}</p>
              </li>
            ))}
          </ol>
        </div>

        <div className="border-t border-slate-100 p-6">
          {!readAll && (
            <p className="mb-2 text-center text-xs font-bold text-amber-600">
              انزل لآخر النقاط حتى تكدر توافق
            </p>
          )}
          <button
            onClick={accept}
            disabled={accepting || !readAll}
            className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-3 font-bold text-white disabled:opacity-50"
          >
            {accepting ? 'جاري الحفظ...' : '✍️ قرأت وألتزم'}
          </button>
          {error && <p className="mt-2 text-center text-xs font-bold text-red-600">{error}</p>}
          <p className="mt-2 text-center text-xs text-slate-400">
            تكدر ترجع تقراها بأي وقت من الإعدادات.
          </p>
        </div>
      </div>
    </div>
  )
}
