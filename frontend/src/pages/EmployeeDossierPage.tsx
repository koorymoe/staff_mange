import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, apiRequest } from '../api'
import { useApi } from '../utils/useApi'
import PctWhy from '../ui/PctWhy'
import { card } from '../ui/classes'

// ═══ 📂 ملف الموظف الكامل (دازاي 11-02 — تعديلات ٩.٣) ═══
// نفس الملف الي يقراه ماتركس قبل ما يحكم على موظف: دراسته ومهاراته وتدريبه،
// منحناه ٦ أشهر (يتقدم لو يتراجع)، أقوى وأضعف شي، سرعته، والتزامه ومشاكله.
// بلا راتب.

type Month = { month: string; matrixPct: number | null; humanAvg: number | null; ratings: number }
type Rule = { rule: string; title: string; pct: number; n: number }
type Dossier = {
  name: string; role: string; effectiveRole: string; position: string; jobTitle: string; hireDate: string; contractEnd: string; isLeader: boolean
  major: string; degree: string; certificate: string; monthlyHours: number | null; companyRating: number | null; companyNote: string
  skills: string[]; training: string[]; leaderSkills: string[]
  months: Month[]; trend: 'UP' | 'DOWN' | 'STEADY' | 'UNKNOWN'; trendNote: string; strong: Rule[]; weak: Rule[]
  speedRatio: number | null; speedN: number
  attendance30: number; reports30: number; reportsLate30: number; negligence90: number; discipline90: number
  justifications90: number; justificationsRejected90: number; complaints90: number; workplaceIssues90: number; peerNotesAbout90: number
  flags: string[]
}
const TREND = { UP: ['📈 يتقدم', 'text-emerald-700 bg-emerald-50'], DOWN: ['📉 يتراجع', 'text-red-700 bg-red-50'], STEADY: ['➖ ثابت', 'text-slate-700 bg-slate-100'], UNKNOWN: ['… بعده ما يبين', 'text-slate-500 bg-slate-50'] } as const

export default function EmployeeDossierPage() {
  const [sp, setSp] = useSearchParams()
  const id = sp.get('id') ?? ''
  const [q, setQ] = useState('')
  const { data: emps } = useApi(() => api.getEmployees(), [])
  const list = useMemo(() => (emps ?? []).filter((e) => e.status === 'ACTIVE' && (!q.trim() || e.name.includes(q.trim()))).slice(0, 12), [emps, q])
  const { data: d, error } = useApi(() => apiRequest<Dossier>(`/employees/${encodeURIComponent(id)}/dossier`), [id], { enabled: !!id })
  return (
    <div dir="rtl" className="space-y-4">
      <div className={card}>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">📂 ملف الموظف</h1>
        <p className="text-sm text-slate-500">نفس الملف الي يقراه ماتركس قبل ما يحكم على أي موظف. بلا راتب.</p>
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="🔍 ابحث عن موظف" className="mt-2 w-full max-w-sm rounded-xl border border-slate-300 px-3 py-2 text-sm" />
        {q.trim() && <div className="mt-1 flex flex-wrap gap-1">{list.map((e) => (
          <button key={e.id} onClick={() => { setSp({ id: e.id }); setQ('') }} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold hover:bg-sky-100">{e.name}</button>
        ))}</div>}
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {d && (
        <>
          <div className={card}>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h2 className="text-xl font-extrabold text-[#0f2040]">{d.name} {d.isLeader && <span className="text-sm">⭐ ليدر</span>}</h2>
                <p className="text-sm text-slate-500">{d.position || d.jobTitle || d.effectiveRole}{d.hireDate && ` · من ${d.hireDate.slice(0, 10)}`}{d.contractEnd && ` · العقد لـ ${d.contractEnd}`}</p>
              </div>
              <span className={`rounded-full px-3 py-1 text-sm font-bold ${TREND[d.trend][1]}`}>{TREND[d.trend][0]}</span>
            </div>
            <p className="mt-1 text-xs text-slate-500">{d.trendNote}</p>
            {d.flags.length > 0 && <ul className="mt-3 space-y-1 rounded-xl bg-amber-50 p-3 text-sm text-amber-900">{d.flags.map((f) => <li key={f}>{f}</li>)}</ul>}
          </div>

          <div className="grid gap-4 lg:grid-cols-2">
            <div className={card}>
              <h3 className="mb-2 font-bold text-[#0f2040]">📊 المنحنى (٦ أشهر)</h3>
              <div className="flex h-40 items-end gap-2">
                {d.months.map((m) => (
                  <div key={m.month} className="flex flex-1 flex-col items-center gap-1">
                    <div className="flex h-32 w-full items-end justify-center gap-0.5">
                      <div title={`ماتركس ${m.matrixPct ?? '—'}٪`} className="w-1/2 rounded-t bg-[#2c5aad]" style={{ height: `${m.matrixPct ?? 0}%`, opacity: m.matrixPct === null ? 0.15 : 1 }} />
                      <div title={`البشر ${m.humanAvg ?? '—'}/5 (${m.ratings})`} className="w-1/2 rounded-t bg-amber-400" style={{ height: `${(m.humanAvg ?? 0) * 20}%`, opacity: m.humanAvg === null ? 0.15 : 1 }} />
                    </div>
                    <span className="text-[10px] text-slate-500" dir="ltr">{m.month.slice(2)}</span>
                  </div>
                ))}
              </div>
              <p className="mt-1 text-[11px] text-slate-500"><span className="text-[#2c5aad]">■</span> ماتركس ٪ · <span className="text-amber-500">■</span> تقييم البشر (من ٥)</p>
            </div>
            <div className={card}>
              <h3 className="mb-2 font-bold text-[#0f2040]">💪 أقوى / 🔻 أضعف (٦٠ يوم)</h3>
              {d.strong.map((r) => <p key={r.rule} className="text-sm text-emerald-700">✓ {r.title} — <PctWhy title={`قاعدة «${r.title}»`} suffix="٪" value={Math.round(r.pct)}
                formula={<>مجموع النقاط الي أخذها بهالقاعدة ÷ مجموع أعلى نقاط ممكنة إلها، على {r.n} مرة انحسبت عليه بآخر ٦٠ يوم = {r.pct.toFixed(0)}٪.</>}
                meaning={<>شكد يجيب من حق هالقاعدة. «أقوى» = ٨٠٪ أو أكثر (أعلى ٣)، «أضعف» = أقل من ٦٠٪ (أوطى ٣). القاعدة الي انحسبت أقل من ٣ مرات ما تطلع هنا.</>}
                source={<>نقاط ماتركس المسجلة على هالموظف (بلا الملغية، وبلا الحضور والجرد) بآخر ٦٠ يوم.</>} /> <span className="text-xs text-slate-400">({r.n})</span></p>)}
              {d.weak.map((r) => <p key={r.rule} className="text-sm text-red-700">✗ {r.title} — <PctWhy title={`قاعدة «${r.title}»`} suffix="٪" value={Math.round(r.pct)}
                formula={<>مجموع النقاط الي أخذها بهالقاعدة ÷ مجموع أعلى نقاط ممكنة إلها، على {r.n} مرة انحسبت عليه بآخر ٦٠ يوم = {r.pct.toFixed(0)}٪.</>}
                meaning={<>شكد يجيب من حق هالقاعدة. «أقوى» = ٨٠٪ أو أكثر (أعلى ٣)، «أضعف» = أقل من ٦٠٪ (أوطى ٣). القاعدة الي انحسبت أقل من ٣ مرات ما تطلع هنا.</>}
                source={<>نقاط ماتركس المسجلة على هالموظف (بلا الملغية، وبلا الحضور والجرد) بآخر ٦٠ يوم.</>} /> <span className="text-xs text-slate-400">({r.n})</span></p>)}
              {d.strong.length + d.weak.length === 0 && <p className="text-sm text-slate-400">بعد ماكو بيانات كافية.</p>}
              <p className="mt-2 text-sm"><b>السرعة:</b> {d.speedRatio === null ? 'ماكو حجوزات كافية' : `${d.speedRatio.toFixed(2)} من المعتاد على ${d.speedN} حجز ${d.speedRatio >= 1.4 ? '🐢' : d.speedRatio <= 0.7 ? '⚡' : ''}`}</p>
            </div>
          </div>

          <div className="grid gap-4 lg:grid-cols-3">
            <div className={card}>
              <h3 className="mb-2 font-bold text-[#0f2040]">🎓 الدراسة والمهارات</h3>
              <p className="text-sm"><b>التخصص:</b> {d.major || '—'} · <b>الشهادة:</b> {d.degree || d.certificate || '—'}</p>
              <p className="text-sm"><b>الساعات المطلوبة:</b> {d.monthlyHours ?? '—'}</p>
              {d.companyRating !== null && <p className="text-sm"><b>رأيه بالشركة:</b> {'⭐'.repeat(d.companyRating)} {d.companyNote}</p>}
              <p className="mt-2 text-sm"><b>المهارات:</b> {d.skills.join('، ') || '—'}</p>
              <p className="text-sm"><b>التدريب:</b> {d.training.join('، ') || '—'}</p>
              {d.leaderSkills.length > 0 && <p className="text-sm"><b>مهارات الليدرية:</b> {d.leaderSkills.join('، ')}</p>}
            </div>
            <div className={card}>
              <h3 className="mb-2 font-bold text-[#0f2040]">📝 الالتزام (٣٠ يوم)</h3>
              <p className="text-sm">دوام: <b>{d.attendance30}</b> يوم</p>
              <p className="text-sm">تقارير إنجاز: <b>{d.reports30}</b> (متأخرة <b className={d.reportsLate30 ? 'text-red-600' : ''}>{d.reportsLate30}</b>)</p>
            </div>
            <div className={card}>
              <h3 className="mb-2 font-bold text-[#0f2040]">⚠️ المشاكل (٩٠ يوم)</h3>
              <p className="text-sm">إهمال تقييم: <b>{d.negligence90}</b> · خصم انضباط: <b>{d.discipline90}</b></p>
              <p className="text-sm">تبريرات: <b>{d.justifications90}</b> (مرفوضة <b>{d.justificationsRejected90}</b>)</p>
              <p className="text-sm">شكاوي زبائن: <b>{d.complaints90}</b> · مشاكل وظيفية: <b>{d.workplaceIssues90}</b></p>
              <p className="text-sm">ملاحظات الزملاء عنه: <b>{d.peerNotesAbout90}</b></p>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
