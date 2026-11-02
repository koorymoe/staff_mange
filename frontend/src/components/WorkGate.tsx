import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { useSession, isAdminOrOwner } from '../session'

// ═══ 🧠 بوابة الالتزام (دازاي 11-02 — تعديلات ٩.١) ═══
// ١. استمارة معلومات الموظف — مرة وحدة.
// ٢. تقرير الإنجاز اليومي — الي داوم يوم وما كتب تقريره، تنقفل عليه الشاشات
//    لحد ما يكتبه (يتأشّر «متأخر»). ماتركس يقرا التقارير ويطابقها بالشغل.
// ما تنسكّر: ماكو زر إغلاق — القرار قرار دازاي.

type DutyItem = { stage: string; bookingId: string; code: string; rateeId: string; rateeName: string; hint: string }
type Duty = { day: string; items: DutyItem[]; labels: Record<string, string> }
type Gate = { survey: boolean; missingDay: string | null; todayDone: boolean; jobTitle: string; ratingDuty: Duty | null }
const inp = 'w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-[#2c5aad]'
const lbl = 'mb-1 block text-xs font-bold text-slate-600'

export default function WorkGate() {
  const { employee } = useSession()
  const exempt = isAdminOrOwner(employee)
  const [gate, setGate] = useState<Gate | null>(null)
  const [tick, setTick] = useState(0)
  useEffect(() => {
    if (exempt || !employee) return
    let alive = true
    apiRequest<Gate>('/me/work-gate').then((g) => { if (alive) setGate(g) }).catch(() => {})
    return () => { alive = false }
  }, [exempt, employee, tick])
  if (exempt || !gate) return null
  if (gate.survey) return <Shell><Survey jobTitle={gate.jobTitle} onDone={() => setTick((n) => n + 1)} /></Shell>
  if (gate.missingDay) return <Shell><Daily day={gate.missingDay} onDone={() => setTick((n) => n + 1)} /></Shell>
  if (gate.ratingDuty) return <Shell><DutyForm duty={gate.ratingDuty} late onDone={() => setTick((n) => n + 1)} /></Shell>
  return null
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <div dir="rtl" className="fixed inset-0 z-[90] flex items-center justify-center overflow-y-auto bg-[#0f2040]/80 p-4 backdrop-blur-sm">
      <div className="my-6 w-full max-w-lg rounded-3xl bg-white p-6 shadow-2xl">{children}</div>
    </div>
  )
}

function Survey({ jobTitle, onDone }: { jobTitle: string; onDone: () => void }) {
  const [f, setF] = useState({ major: '', degree: '', selfSalaryIQD: '', monthlyHours: '', currentJob: jobTitle, companyRating: 0, companyNote: '', contractEnd: '' })
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const set = (k: keyof typeof f) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => setF({ ...f, [k]: e.target.value })
  const save = async () => {
    setBusy(true); setErr('')
    try {
      const n = (v: string) => (v.trim() ? Number(v.replace(/[,،\s]/g, '')) : null)
      await apiRequest('/me/profile-survey', { method: 'POST', body: JSON.stringify({ ...f, selfSalaryIQD: n(f.selfSalaryIQD), monthlyHours: n(f.monthlyHours) }) })
      onDone()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  return (
    <div className="space-y-3">
      <h2 className="text-xl font-extrabold text-[#0f2040]">🧾 استمارة معلوماتك</h2>
      <p className="text-sm text-slate-500">مرة وحدة بس. ماتركس يعتمد عليها حتى يقيّمك صح ويعرف وين قوتك. الراتب يشوفه بس إدارة الكوادر والمالك.</p>
      <div className="grid gap-3 sm:grid-cols-2">
        <label><span className={lbl}>تخصصك الدراسي</span><input className={inp} value={f.major} onChange={set('major')} placeholder="مثلاً: هندسة كهرباء" /></label>
        <label><span className={lbl}>شهادتك</span><input className={inp} value={f.degree} onChange={set('degree')} placeholder="مثلاً: بكالوريوس، دبلوم، إعدادية" /></label>
        <label><span className={lbl}>وظيفتك الحالية</span><input className={inp} value={f.currentJob} onChange={set('currentJob')} /></label>
        <label><span className={lbl}>راتبك الشهري (دينار)</span><input className={inp} dir="ltr" inputMode="numeric" value={f.selfSalaryIQD} onChange={set('selfSalaryIQD')} /></label>
        <label><span className={lbl}>الساعات المطلوبة منك بالشهر</span><input className={inp} dir="ltr" inputMode="numeric" value={f.monthlyHours} onChange={set('monthlyHours')} /></label>
        <label><span className={lbl}>تاريخ نهاية عقدك</span><input type="date" className={inp} value={f.contractEnd} onChange={set('contractEnd')} /></label>
      </div>
      <div>
        <span className={lbl}>تقييمك للعمل ويا شركة الأماني</span>
        <div className="flex gap-1">
          {[1, 2, 3, 4, 5].map((v) => (
            <button key={v} type="button" onClick={() => setF({ ...f, companyRating: v })}
              className={`h-10 w-10 rounded-xl text-xl ${f.companyRating >= v ? 'bg-amber-100' : 'bg-slate-100 grayscale'}`}>⭐</button>
          ))}
        </div>
        <textarea className={`${inp} mt-2`} rows={2} value={f.companyNote} onChange={set('companyNote')} placeholder="شنو يعجبك وشنو تحب يتحسن؟ (اختياري)" />
      </div>
      {err && <p className="text-sm font-bold text-red-600">{err}</p>}
      <button disabled={busy} onClick={() => void save()} className="w-full rounded-xl bg-[#0f2040] py-3 font-bold text-white disabled:opacity-50">{busy ? '...' : 'حفظ'}</button>
    </div>
  )
}

function Daily({ day, onDone }: { day: string; onDone: () => void }) {
  const [body, setBody] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const label = new Date(day + 'T12:00:00').toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })
  const save = async () => {
    setBusy(true); setErr('')
    try { await apiRequest('/me/daily-report', { method: 'POST', body: JSON.stringify({ day, body }) }); onDone() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  return (
    <div className="space-y-3">
      <h2 className="text-xl font-extrabold text-[#0f2040]">📝 تقرير إنجاز يوم {label}</h2>
      <p className="text-sm text-slate-500">داومت ذاك اليوم وما كتبت تقريرك. اكتب شنو سويت بالتفصيل حتى تكمل شغلك — ماتركس يطابقه ويا الشغل المسجّل بالنظام. التقرير ينحسب <b>متأخر</b>.</p>
      <textarea className={inp} rows={7} value={body} onChange={(e) => setBody(e.target.value)}
        placeholder="مثلاً: تابعت حجز B964 ويا الزبون، كملت فاتورة B958، رتبت مخزن الكاميرات…" />
      <p className="text-left text-[11px] text-slate-400" dir="ltr">{body.trim().length}/20+</p>
      {err && <p className="text-sm font-bold text-red-600">{err}</p>}
      <button disabled={busy || body.trim().length < 20} onClick={() => void save()} className="w-full rounded-xl bg-[#0f2040] py-3 font-bold text-white disabled:opacity-50">{busy ? '...' : 'إرسال التقرير'}</button>
    </div>
  )
}

// ٩.٢ تقييمات فاتته — ما تنفتح الشاشات إلا بعد ما يعوّض الكل ويكتب السبب.
// late=false: نفس النموذج للتقييم بنفس اليوم (بلا سبب ولا سجل إهمال).
export function DutyForm({ duty, late, onDone }: { duty: Duty; late: boolean; onDone: () => void }) {
  const key = (i: DutyItem) => `${i.stage}|${i.bookingId}|${i.rateeId}`
  const [scores, setScores] = useState<Record<string, number>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [reason, setReason] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const rated = duty.items.filter((i) => scores[key(i)])
  const ready = late ? rated.length === duty.items.length && reason.trim().length >= 10 : rated.length > 0
  const label = new Date(duty.day + 'T12:00:00').toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })
  const save = async () => {
    setBusy(true); setErr('')
    try {
      await apiRequest('/me/rating-duty', { method: 'POST', body: JSON.stringify({ day: duty.day, reason, ratings: rated.map((i) => ({ stage: i.stage, bookingId: i.bookingId, rateeId: i.rateeId, score: scores[key(i)], note: notes[key(i)] ?? '' })) }) })
      setScores({}); setNotes({}); onDone()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  return (
    <div className="space-y-3">
      <h2 className="text-xl font-extrabold text-[#0f2040]">{late ? `🧾 فاتك تقييم يوم ${label}` : '⭐ تقييمات اليوم'}</h2>
      {late && <p className="text-sm text-slate-500">ما قيّمت الي عليك بوقتها. <b>عوّض الكل واكتب السبب</b> حتى تكمل شغلك — ينكتب بسجل الإهمال وينزل تقييمك كمقيّم.</p>}
      <div className="max-h-[50vh] space-y-2 overflow-y-auto">
        {duty.items.map((i) => (
          <div key={key(i)} className="rounded-xl border border-slate-200 p-2.5">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p className="text-sm font-bold text-[#0f2040]">{i.rateeName}</p>
                <p className="text-[11px] text-slate-500">{duty.labels[i.stage] ?? i.stage}{i.code ? ` · ${i.code}` : ''}{i.stage === 'OPS_DAILY' && i.hint ? ` · ${i.hint}` : ''}</p>
              </div>
              <div className="flex gap-0.5" dir="ltr">
                {[1, 2, 3, 4, 5].map((v) => (
                  <button key={v} type="button" onClick={() => setScores({ ...scores, [key(i)]: v })}
                    className={`h-8 w-8 rounded-lg text-base ${(scores[key(i)] ?? 0) >= v ? 'bg-amber-100' : 'bg-slate-100 grayscale'}`}>⭐</button>
                ))}
              </div>
            </div>
            {i.stage === 'MONITOR_DAILY' && i.hint && <p className="mt-1 rounded-lg bg-slate-50 p-2 text-xs text-slate-600">📝 {i.hint}</p>}
            <input value={notes[key(i)] ?? ''} onChange={(e) => setNotes({ ...notes, [key(i)]: e.target.value })} placeholder="ملاحظة (اختياري)"
              className="mt-1.5 w-full rounded-lg border border-slate-200 px-2 py-1 text-xs" />
          </div>
        ))}
      </div>
      {late && <label className="block"><span className={lbl}>شنو السبب؟ وضّح ليش ما قيّمت بوقتها</span>
        <textarea className={inp} rows={3} value={reason} onChange={(e) => setReason(e.target.value)} /></label>}
      {err && <p className="text-sm font-bold text-red-600">{err}</p>}
      <button disabled={busy || !ready} onClick={() => void save()} className="w-full rounded-xl bg-[#0f2040] py-3 font-bold text-white disabled:opacity-50">
        {busy ? '...' : late ? `عوّض (${rated.length}/${duty.items.length})` : `حفظ (${rated.length})`}
      </button>
    </div>
  )
}
