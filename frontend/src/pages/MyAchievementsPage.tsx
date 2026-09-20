import { useEffect, useState } from 'react'
import { useSession } from '../session'
import { api, type Achievement } from '../api'

// ═══ الإنجازات — تقرير يومي حر (ماتركس) ═══
//
// طلب صاحب النظام: كل موظف (مو الفني بس) يرفع تقريراً يومياً بشغله —
// الليدر شكد حجز طلّع، مهندس الجودة شكد زبون اتصل، المراقب شكد دقّق،
// المحاسب شكد طابق. الربط بحجز اختياري بالتصميم — الفني يكتب كود
// حجزه لو يريد، وغيره يرفع تقريراً عاماً بلا حجز.
//
// ⚠️ التقرير يوصل لمدير النظام والمالك حصراً للمراجعة — هذي الشاشة
// بس تسجّله وتعرض تاريخ الموظف نفسه.
export default function MyAchievementsPage() {
  const { employee } = useSession()
  const [reportText, setReportText] = useState('')
  const [bookingCode, setBookingCode] = useState('')
  const [saving, setSaving] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [mine, setMine] = useState<Achievement[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let alive = true
    if (!employee) return
    queueMicrotask(() => { if (alive) setLoading(true) })
    void (async () => {
      try {
        const rows = await api.getAchievements({ employeeId: employee.id, limit: 30 })
        if (alive) setMine(rows)
      } catch {
        // صامت — التاريخ إضافي، ما نعطّل الإرسال لأجله
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [employee])

  const submit = async () => {
    if (!reportText.trim()) {
      setErr('يرجى كتابة شنو سويت اليوم')
      return
    }
    setSaving(true)
    setErr(null)
    try {
      const created = await api.createAchievement(bookingCode.trim() || null, reportText.trim())
      setMine((prev) => [created, ...prev])
      setReportText('')
      setBookingCode('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الإرسال')
    } finally {
      setSaving(false)
    }
  }

  const statusBadge = (a: Achievement) => {
    if (a.reviewStatus === 'GOOD') return <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-bold text-emerald-800">✓ زين</span>
    if (a.reviewStatus === 'NEEDS_REVIEW') return <span className="rounded-full bg-amber-100 px-2 py-0.5 text-[11px] font-bold text-amber-800">⚠️ يحتاج مراجعة</span>
    return <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold text-slate-500">بانتظار المراجعة</span>
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📋 إنجازاتي اليوم</h2>
        <p className="mt-1 text-sm text-slate-500">
          اكتب شنو سويت اليوم بشغلك — يوصل مباشرة لمدير النظام والمالك. الربط بحجز اختياري.
        </p>
      </div>

      <div className="space-y-3 rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <textarea
          value={reportText}
          onChange={(e) => setReportText(e.target.value)}
          placeholder="مثلاً: طلعت لثلاث حجوزات وخلصتهن، وحلّيت مشكلة كامرة عند زبون..."
          rows={4}
          className="w-full rounded-xl border border-slate-200 p-3 text-sm focus:border-brand-500 focus:outline-none"
        />
        <input
          value={bookingCode}
          onChange={(e) => setBookingCode(e.target.value)}
          placeholder="كود الحجز المربوط (اختياري)"
          className="w-full rounded-xl border border-slate-200 p-2.5 text-sm focus:border-brand-500 focus:outline-none"
        />
        {err && <p className="text-sm text-red-600">{err}</p>}
        <button
          onClick={submit}
          disabled={saving}
          className="rounded-xl bg-brand-700 px-5 py-2.5 text-sm font-bold text-white transition hover:bg-brand-800 disabled:opacity-50"
        >
          {saving ? 'جاري الإرسال...' : '📤 إرسال التقرير'}
        </button>
      </div>

      <div>
        <h3 className="mb-2 text-sm font-bold text-slate-600">تقاريري السابقة</h3>
        {loading && <p className="text-slate-400">جاري التحميل...</p>}
        {!loading && mine.length === 0 && (
          <p className="rounded-xl border border-white bg-white p-6 text-center text-slate-400">ماكو تقارير سابقة.</p>
        )}
        <div className="space-y-2">
          {mine.map((a) => (
            <div key={a.id} className="rounded-xl border border-white bg-white p-3 shadow-[0_2px_10px_rgba(15,32,64,0.05)]">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="text-[11px] text-slate-400">{new Date(a.createdAt).toLocaleString('en-GB')}</span>
                {statusBadge(a)}
              </div>
              <p className="mt-1 text-sm text-slate-700">{a.reportText}</p>
              {a.bookingCode && <p className="mt-1 text-[11px] text-slate-400">مربوط بحجز: {a.bookingCode}</p>}
              {a.reviewNote && (
                <p className="mt-1 rounded-lg bg-amber-50 p-2 text-[12px] text-amber-800">ملاحظة {a.reviewedByName || ''}: {a.reviewNote}</p>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
