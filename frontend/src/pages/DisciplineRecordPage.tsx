import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type DisciplineEntry, type DisciplineRecord } from '../api'
import MatrixNote from '../components/MatrixNote'

// ═══ سجل الانضباط الوظيفي — طلب (ع) 10-05 ═══
// كل الموظفين ويا كل خصم من أول يوم بالنظام: كم مرة، شكد، شنو السبب، متى،
// منو خصمه، وانرجع لو لا. للمدير والمالك بس.

const SRC: Record<DisciplineEntry['source'], string> = { KPI: '💸 خصم', POINTS: '📉 نقاط انضباط', LEDGER: '🧾 تسوية عهدة' }
const money = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`
const date = (s: string) => new Date(s).toLocaleDateString('en-GB')

function csv(rows: DisciplineEntry[]) {
  const head = ['الموظف', 'النوع', 'التاريخ', 'المبلغ', 'النقاط', 'السبب', 'منو', 'الحجز', 'انرجع']
  const esc = (v: unknown) => `"${String(v ?? '').replace(/"/g, '""')}"`
  const lines = rows.map((r) => [r.employeeName, SRC[r.source], date(r.at), r.amount, r.points, r.reason, r.byName, r.bookingCode, r.returned ? 'نعم' : 'لا'].map(esc).join(','))
  const blob = new Blob(['﻿' + [head.join(','), ...lines].join('\n')], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `سجل-الانضباط-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}

export default function DisciplineRecordPage() {
  const [rec, setRec] = useState<DisciplineRecord | null>(null)
  const [err, setErr] = useState('')
  const [emp, setEmp] = useState<string | null>(null)
  const [src, setSrc] = useState<string>('ALL')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  useEffect(() => { void api.getDisciplineRecord().then(setRec).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])

  const entries = useMemo(() => (rec?.entries ?? []).filter((e) =>
    (!emp || e.employeeId === emp) && (src === 'ALL' || e.source === src)
    && (!from || e.at.slice(0, 10) >= from) && (!to || e.at.slice(0, 10) <= to)
  ), [rec, emp, src, from, to])

  if (err) return <p className="text-red-600">{err}</p>
  if (!rec) return <p className="text-slate-400">جاري التحميل…</p>
  const sel = rec.employees.find((e) => e.employeeId === emp)

  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📒 سجل الانضباط الوظيفي</h2>
        <p className="text-sm text-slate-500">كل خصم من أول يوم بالنظام: شكد، ليش، متى، منو خصمه، وإذا انرجع. للمدير والمالك بس.</p>
      </div>
      <MatrixNote>{rec.insights.join(' ')}</MatrixNote>

      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {rec.employees.map((e) => (
          <button key={e.employeeId} type="button" onClick={() => setEmp(emp === e.employeeId ? null : e.employeeId)}
            className={`rounded-2xl border p-3 text-right transition ${emp === e.employeeId ? 'border-[#0f2040] bg-white shadow ring-2 ring-sky-200' : 'border-slate-200 bg-white hover:shadow'}`}>
            <div className="flex items-baseline justify-between gap-2">
              <b className="truncate text-sm">{e.employeeName}</b>
              <span className="shrink-0 rounded-full bg-red-50 px-2 py-0.5 text-xs font-black text-red-700">{e.count} مرة</span>
            </div>
            <p className="mt-1 text-xs text-slate-600">الصافي <b>{money(e.net)}</b>{e.returned > 0 && <> · رجع {money(e.returned)}</>}</p>
            {e.pointsLost > 0 && <p className="text-xs text-slate-600">نقاط انضباط: −{e.pointsLost}{e.pointsBack ? ` · رجعت +${e.pointsBack}` : ''}</p>}
            <p className="text-[11px] text-slate-500">هالشهر {e.thisMonth} · الشهر الفات {e.lastMonth}{e.last && <> · آخر خصم {date(e.last)}</>}</p>
            {e.topReason && <p className="mt-1 truncate text-[11px] text-slate-500">أكثر سبب: {e.topReason}</p>}
          </button>
        ))}
        {rec.employees.length === 0 && <p className="rounded-xl bg-white p-6 text-center text-sm text-slate-400">ماكو خصومات مسجّلة.</p>}
      </div>

      {sel && sel.insights.length > 0 && <MatrixNote>{sel.employeeName}: {sel.insights.join(' ')}</MatrixNote>}

      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <div className="mb-3 flex flex-wrap items-center gap-2 text-xs">
          <b className="text-sm">{sel ? `سجل ${sel.employeeName}` : 'كل السجل'} ({entries.length})</b>
          <select value={src} onChange={(e) => setSrc(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1">
            <option value="ALL">كل الأنواع</option><option value="KPI">💸 خصم</option><option value="POINTS">📉 نقاط انضباط</option><option value="LEDGER">🧾 تسوية عهدة</option>
          </select>
          <label>من <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1" /></label>
          <label>لحد <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1" /></label>
          {emp && <button type="button" onClick={() => setEmp(null)} className="text-sky-700 underline">كل الموظفين</button>}
          <button type="button" onClick={() => csv(entries)} className="ms-auto rounded-lg border border-slate-200 px-3 py-1 font-bold hover:bg-slate-50">⬇️ تصدير Excel</button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[760px] text-right text-xs">
            <thead className="bg-slate-50 text-slate-600">
              <tr><th className="p-2">التاريخ</th><th className="p-2">الموظف</th><th className="p-2">النوع</th><th className="p-2">المبلغ</th><th className="p-2">النقاط</th><th className="p-2">السبب</th><th className="p-2">منو</th><th className="p-2">الحجز</th><th className="p-2">الحالة</th></tr>
            </thead>
            <tbody>
              {entries.map((e) => (
                <tr key={e.source + e.id} className={`border-t border-slate-100 ${e.returned ? 'text-slate-400 line-through decoration-slate-300' : ''}`}>
                  <td className="p-2 whitespace-nowrap">{date(e.at)}</td>
                  <td className="p-2 font-bold">{e.employeeName}</td>
                  <td className="p-2 whitespace-nowrap">{SRC[e.source]}</td>
                  <td className="p-2 whitespace-nowrap">{e.amount ? money(e.amount) : '—'}</td>
                  <td className="p-2">{e.points ? (e.points > 0 && e.source === 'POINTS' ? `+${e.points}` : e.source === 'POINTS' ? e.points : `−${e.points}`) : '—'}</td>
                  <td className="max-w-xs p-2">{e.reason || '—'}</td>
                  <td className="p-2">{e.byName ?? '—'}</td>
                  <td className="p-2">{e.bookingId ? <Link to={`/bookings?focus=${e.bookingId}`} className="text-sky-700 underline">{e.bookingCode}</Link> : '—'}</td>
                  <td className="p-2 whitespace-nowrap">{e.returned ? `↩️ انرجع${e.returnedAt ? ' ' + date(e.returnedAt) : ''}` : e.source === 'POINTS' && e.points > 0 ? '↩️ استرجاع' : 'قائم'}</td>
                </tr>
              ))}
              {entries.length === 0 && <tr><td colSpan={9} className="p-6 text-center text-slate-400">ماكو شي بهالفلترة.</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
