import { useEffect, useMemo, useState } from 'react'
import { apiRequest } from '../api'
import { askJustification } from '../utils/justify'

// ═══ رقابة مشرفي المشاريع على عرض السعر — (ع) 10-16 ═══
// «شوكت استلم المشروع، شوكت تواصل وي الزبون، شوكت سواله عرض سعر».
interface Row {
  projectId: string; code: string; name: string; stage: string
  supervisorId: string | null; supervisorName: string | null
  receivedAt: string; contactedAt: string | null; contactNote: string | null; quotedAt: string | null
  contactHours: number | null; quoteHours: number | null; waitingHours: number; late: boolean; lateWhy: string | null
  offSystem: boolean
}
interface Data { rows: Row[]; contactHours: number; quoteHours: number }

// الأوقات ترجع UTC — نعرضها بتوقيت بغداد
const fmt = (s: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' }) : '—')
const hrs = (h: number | null) => (h == null ? '—' : h < 24 ? `${Math.round(h)} ساعة` : `${(h / 24).toFixed(1)} يوم`)

export default function ProjectQuoteWatchPanel() {
  const [d, setD] = useState<Data | null>(null)
  const [err, setErr] = useState('')
  const [onlyLate, setOnlyLate] = useState(true)
  useEffect(() => { apiRequest<Data>('/project-quote-watch').then(setD).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب')) }, [])

  const bySup = useMemo(() => {
    const m = new Map<string, { name: string; n: number; late: number; c: number[]; q: number[] }>()
    for (const r of d?.rows ?? []) {
      const k = r.supervisorId || '—'
      const e = m.get(k) ?? { name: r.supervisorName || 'بلا مشرف', n: 0, late: 0, c: [], q: [] }
      e.n++; if (r.late) e.late++
      if (r.contactHours != null) e.c.push(r.contactHours)
      if (r.quoteHours != null) e.q.push(r.quoteHours)
      m.set(k, e)
    }
    const avg = (a: number[]) => (a.length ? a.reduce((x, y) => x + y, 0) / a.length : null)
    return [...m.values()].map((e) => ({ ...e, avgC: avg(e.c), avgQ: avg(e.q) })).sort((a, b) => b.late - a.late)
  }, [d])

  if (err) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const rows = d.rows.filter((r) => !onlyLate || r.late)

  return (
    <div className="space-y-4" dir="rtl">
      <p className="rounded-xl bg-sky-50 px-4 py-2 text-[12.5px] text-sky-900">
        لكل مشروع: <b>شوكت استلمه المشرف</b> ← <b>شوكت تواصل ويا الزبون</b> (لازم خلال {d.contactHours} ساعة) ← <b>شوكت سوّى عرض السعر</b> (لازم خلال {d.quoteHours / 24} أيام من الاستلام).
        التواصل يسجّله المشرف بزر «📞 تواصلت ويا الزبون» بشاشة المشاريع. المشروع الي عبر المرحلة ما ينحسب متأخر عليها، والمكتمل ما يطلع. للمتابعة — بلا غرامات.
      </p>

      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        {bySup.map((s) => (
          <div key={s.name} className={`rounded-xl border p-3 ${s.late ? 'border-red-200 bg-red-50/60' : 'border-emerald-200 bg-emerald-50/50'}`}>
            <p className="font-extrabold text-[#0f2040]">{s.name}</p>
            <p className="mt-1 text-xs text-slate-600">{s.n} مشروع · <b className={s.late ? 'text-red-700' : 'text-emerald-700'}>{s.late} متأخر</b></p>
            <p className="text-xs text-slate-600">معدل التواصل: <b>{hrs(s.avgC)}</b> · معدل عرض السعر: <b>{hrs(s.avgQ)}</b></p>
          </div>
        ))}
      </div>

      <label className="flex items-center gap-2 text-sm font-bold text-slate-600">
        <input type="checkbox" checked={onlyLate} onChange={(e) => setOnlyLate(e.target.checked)} /> المتأخرة بس ({d.rows.filter((r) => r.late).length} من {d.rows.length})
      </label>

      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full text-right text-[13px]">
          <thead className="bg-slate-50 text-[11px] text-slate-500">
            <tr><th className="p-2">المشروع</th><th className="p-2">المشرف</th><th className="p-2">① استلم</th><th className="p-2">② تواصل</th><th className="p-2">③ عرض السعر</th><th className="p-2">الحالة</th></tr>
          </thead>
          <tbody>
            {rows.map((r) => (
              <tr key={r.projectId} className={`border-t border-slate-100 align-top ${r.late ? 'bg-red-50/40' : ''}`}>
                <td className="p-2"><b>{r.code}</b><span className="block text-[11px] text-slate-500">{r.name} · {r.stage}</span></td>
                <td className="p-2 font-bold">{r.supervisorName ?? 'بلا مشرف'}</td>
                <td className="p-2 tabular-nums">{fmt(r.receivedAt)}</td>
                <td className="p-2">{r.contactedAt ? <>{fmt(r.contactedAt)}<span className="block text-[11px] text-slate-500">بعد {hrs(r.contactHours)}{r.contactNote ? ` — ${r.contactNote}` : ''}</span></> : (r.stage.startsWith('1.') ? <span className="text-red-700">لهسه ما تواصل</span> : <span className="text-slate-500">عبر المرحلة (ما مسجّل بالنظام)</span>)}</td>
                <td className="p-2">{r.quotedAt ? <>{fmt(r.quotedAt)}<span className="block text-[11px] text-slate-500">بعد {hrs(r.quoteHours)} من الاستلام</span></> : (/^[45]\.|الإعلام/.test(r.stage) ? <span className="text-slate-500">عبر المرحلة (برّا النظام)</span> : <span className="text-amber-700">ماكو عرض سعر · صارله {hrs(r.waitingHours)}</span>)}</td>
                <td className="p-2">{r.late && r.supervisorId && (
                  <button onClick={() => askJustification(r.supervisorId!, `مشروع ${r.code} (${r.name}): ${r.lateWhy} — صارله ${hrs(r.waitingHours)} من استلمته. ليش؟`).catch((e) => alert(e instanceof Error ? e.message : 'تعذر'))}
                    className="mb-1 block rounded-lg bg-red-600 px-2 py-1 text-[11px] font-bold text-white">📝 اطلب تفسير</button>
                )}{r.late ? <span className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-800">🔴 {r.lateWhy}</span> : <span className="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-bold text-emerald-800">🟢 بالوقت</span>}
                  {r.offSystem && <span className="mt-1 block rounded-full bg-violet-100 px-2 py-0.5 text-[11px] font-bold text-violet-800">🤖 ماتركس: مشى بالمشروع بلا ما يسجّل تواصل — ما ماشي بالنظام</span>}</td>
              </tr>
            ))}
            {rows.length === 0 && <tr><td colSpan={6} className="p-8 text-center text-slate-400">{onlyLate ? 'ماكو مشروع متأخر ✓' : 'ماكو مشاريع'}</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}
