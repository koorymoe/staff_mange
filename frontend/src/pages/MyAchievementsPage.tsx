import { useEffect, useState } from 'react'
import { useSession } from '../session'
import { api, apiRequest, type Achievement } from '../api'
import { dayBaghdad } from '../utils/format'
import { VoicePlayer, VoiceRecorder } from '../components/VoiceNote'

// ═══ الإنجازات — تقرير يومي حر (ماتركس) ═══
//
// طلب صاحب النظام: كل موظف (مو الفني بس) يرفع تقريراً يومياً بشغله —
// الليدر شكد حجز طلّع، مهندس الجودة شكد زبون اتصل، المراقب شكد دقّق،
// المحاسب شكد طابق. الربط بحجز اختياري بالتصميم — الفني يكتب كود
// حجزه لو يريد، وغيره يرفع تقريراً عاماً بلا حجز.
//
// ⚠️ التقرير يوصل لمدير النظام والمالك حصراً للمراجعة — هذي الشاشة
// بس تسجّله وتعرض تاريخ الموظف نفسه.
// (دازاي 11-05) تقرير اليوم الإجباري من السيرفر — ينكتب من هالشاشة (أو من قفل اليوم الفائت)
type DailyRow = { id: string; day: string; body: string; late: boolean; createdAt: string; matchStatus: string | null; matchVerdict: string | null; matchReview: string | null }
const MATCH: Record<string, [string, string]> = {
  MATCH: ['✅ ماتركس: يطابق شغلك المسجّل', 'bg-emerald-50 text-emerald-700'],
  PARTIAL: ['🟡 ماتركس: يطابق جزئياً', 'bg-amber-50 text-amber-700'],
  NO_TRACE: ['⏳ ماتركس: ما لگى أثر — ينتظر حكم المراقب', 'bg-slate-100 text-slate-600'],
}
const VERDICT: Record<string, string> = { JUSTIFIED: '✅ المراقب: مبرّر', UNJUSTIFIED: '❌ المراقب: غير مبرّر' }
const timeAr = (iso: string) => new Date(iso).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Baghdad' })
const dayTitle = (day: string) => new Date(day + 'T12:00:00').toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })

export default function MyAchievementsPage() {
  const { employee } = useSession()
  const [reportText, setReportText] = useState('')
  const [bookingCode, setBookingCode] = useState('')
  const [voice, setVoice] = useState<{ blob: Blob; seconds: number } | null>(null)
  const [saving, setSaving] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [mine, setMine] = useState<Achievement[]>([])
  const [loading, setLoading] = useState(true)
  const [daily, setDaily] = useState<DailyRow[]>([])

  useEffect(() => {
    let alive = true
    if (!employee) return
    queueMicrotask(() => { if (alive) setLoading(true) })
    void (async () => {
      try {
        const [rows, d] = await Promise.all([api.getMyAchievements(60), apiRequest<DailyRow[]>('/me/daily-reports').catch(() => [] as DailyRow[])])
        if (alive) { setMine(rows); setDaily(d) }
      } catch {
        // صامت — التاريخ إضافي، ما نعطّل الإرسال لأجله
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [employee])

  const submit = async () => {
    if (!reportText.trim() && !voice) {
      setErr('اكتب أو سجّل فويس بشنو سويت اليوم')
      return
    }
    setSaving(true)
    setErr(null)
    try {
      const created = await api.createAchievement(bookingCode.trim() || null, reportText.trim() || '🎙️ فويس مسج')
      if (voice) {
        try {
          const r = await api.uploadAchievementVoice(created.id, voice.blob, voice.seconds)
          created.voiceSeconds = r.seconds
        } catch (e) {
          setErr('انرسل التقرير بس الفويس ما انرفع: ' + (e instanceof Error ? e.message : ''))
        }
      }
      setMine((prev) => [created, ...prev])
      apiRequest<DailyRow[]>('/me/daily-reports').then(setDaily).catch(() => {})
      setReportText('')
      setVoice(null)
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

  const byDay = new Map<string, Achievement[]>()
  for (const a of mine) {
    const d = dayBaghdad(a.createdAt)
    byDay.set(d, [...(byDay.get(d) ?? []), a])
  }
  const dailyBy = new Map(daily.map((r) => [r.day, r]))
  const days = [...new Set([...byDay.keys(), ...dailyBy.keys()])].sort().reverse()

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📋 تقرير إنجاز اليوم</h2>
        <p className="mt-1 text-sm text-slate-500">
          <b>إجباري كل يوم.</b> اكتب أو سجّل فويس بشنو سويت — يوصل لمديرك والمالك، وماتركس يطابقه ويا شغلك المسجّل. تگدر تضيف أكثر من مرة باليوم. الربط بحجز اختياري.
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
        <VoiceRecorder value={voice} onChange={setVoice} />
        {err && <p className="text-sm text-red-600">{err}</p>}
        <button
          onClick={submit}
          disabled={saving}
          className="rounded-xl bg-brand-700 px-5 py-2.5 text-sm font-bold text-white transition hover:bg-brand-800 disabled:opacity-50"
        >
          {saving ? 'جاري الإرسال...' : '📤 إرسال التقرير'}
        </button>
      </div>

      {/* (دازاي 11-05) يوم بيوم كتقرير مرتب: اليوم، وساعة الرفع، والتأخير، وحكم ماتركس، وكل الي انكتب */}
      <div>
        <h3 className="mb-2 text-sm font-bold text-slate-600">تقاريري السابقة — يوم بيوم</h3>
        {loading && <p className="text-slate-400">جاري التحميل...</p>}
        {!loading && days.length === 0 && (
          <p className="rounded-xl border border-white bg-white p-6 text-center text-slate-400">ماكو تقارير سابقة.</p>
        )}
        <div className="space-y-3">
          {days.map((day) => {
            const rep = dailyBy.get(day)
            const items = byDay.get(day) ?? []
            const first = rep?.createdAt ?? items[items.length - 1]?.createdAt
            const match = rep?.matchStatus ? MATCH[rep.matchStatus] : undefined
            return (
              <div key={day} className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_2px_10px_rgba(15,32,64,0.05)]">
                <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 bg-slate-50 px-4 py-2.5">
                  <b className="text-sm text-slate-800">📅 {dayTitle(day)}</b>
                  <div className="flex flex-wrap items-center gap-1.5 text-[11px] font-bold">
                    {first && <span className="rounded-full bg-white px-2 py-0.5 text-slate-600 ring-1 ring-slate-200">🕐 رُفع الساعة {timeAr(first)}</span>}
                    {rep?.late && <span className="rounded-full bg-red-50 px-2 py-0.5 text-red-700">⏰ متأخر</span>}
                    {!rep && <span className="rounded-full bg-slate-100 px-2 py-0.5 text-slate-500">قبل التقرير الإجباري</span>}
                    {match && <span className={`rounded-full px-2 py-0.5 ${match[1]}`}>{match[0]}</span>}
                    {rep?.matchVerdict && <span className="rounded-full bg-slate-100 px-2 py-0.5 text-slate-700">{VERDICT[rep.matchVerdict] ?? rep.matchVerdict}</span>}
                  </div>
                </div>
                <div className="divide-y divide-slate-100 px-4">
                  {items.length === 0 && rep && <p className="whitespace-pre-wrap py-3 text-sm text-slate-700">{rep.body}</p>}
                  {[...items].reverse().map((a) => (
                    <div key={a.id} className="py-3">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className="text-[11px] text-slate-400">🕐 {timeAr(a.createdAt)}{a.bookingCode && ` · حجز ${a.bookingCode}`}</span>
                        {statusBadge(a)}
                      </div>
                      <p className="mt-1 whitespace-pre-wrap text-sm text-slate-700">{a.reportText}</p>
                      {a.voiceSeconds != null && <VoicePlayer achievementId={a.id} seconds={a.voiceSeconds} />}
                      {a.reviewNote && (
                        <p className="mt-1 rounded-lg bg-amber-50 p-2 text-[12px] text-amber-800">ملاحظة {a.reviewedByName || ''}: {a.reviewNote}</p>
                      )}
                    </div>
                  ))}
                  {rep?.matchReview && <p className="py-2 text-[12px] text-slate-600">ملاحظة المراقب: {rep.matchReview}</p>}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
