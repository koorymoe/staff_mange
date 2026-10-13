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

const COLORS = ['bg-sky-100 text-sky-700', 'bg-rose-100 text-rose-700', 'bg-violet-100 text-violet-700', 'bg-emerald-100 text-emerald-700', 'bg-amber-100 text-amber-700']
const initial = (n: string) => (n.trim()[0] ?? '؟')
const card = 'rounded-2xl border border-slate-200 bg-white shadow-sm'

function Kpi({ label, value, hint, icon, tone }: { label: string; value: string | number; hint: string; icon: string; tone: string }) {
  return (
    <div className={`${card} flex items-start justify-between gap-2 p-4 ${tone}`}>
      <div className="min-w-0">
        <p className="text-sm font-extrabold">{label}</p>
        <p className="mt-1 text-2xl font-black text-[#0f2040] tabular-nums">{value}</p>
        <p className="text-[11px] text-slate-500">{hint}</p>
      </div>
      <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-white/80 text-xl shadow-sm">{icon}</span>
    </div>
  )
}

// ═══ تصميم (ع) 10-13: رأس، بطاقات، الأكثر تكراراً، فلاتر وجدول — نفس البيانات ═══
export default function DisciplineRecordPage() {
  const [rec, setRec] = useState<DisciplineRecord | null>(null)
  const [err, setErr] = useState('')
  const [emp, setEmp] = useState<string | null>(null)
  const [src, setSrc] = useState<string>('ALL')
  const [state, setState] = useState<'ALL' | 'OPEN' | 'RETURNED'>('ALL')
  const [q, setQ] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [allEmps, setAllEmps] = useState(false)
  useEffect(() => { void api.getDisciplineRecord().then(setRec).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])

  const entries = useMemo(() => (rec?.entries ?? []).filter((e) =>
    (!emp || e.employeeId === emp) && (src === 'ALL' || e.source === src)
    && (state === 'ALL' || (state === 'RETURNED' ? e.returned : !e.returned))
    && (!q || `${e.employeeName} ${e.reason} ${e.byName ?? ''} ${e.bookingCode ?? ''}`.includes(q))
    && (!from || e.at.slice(0, 10) >= from) && (!to || e.at.slice(0, 10) <= to)
  ), [rec, emp, src, state, q, from, to])

  if (err) return <p className="text-red-600">{err}</p>
  if (!rec) return <p className="text-slate-400">جاري التحميل…</p>
  const sel = rec.employees.find((e) => e.employeeId === emp)
  const today = new Date().toISOString().slice(0, 10)
  const month = today.slice(0, 7)
  const all = rec.entries
  const todayAmount = all.filter((e) => e.at.slice(0, 10) === today && !e.returned).reduce((n, e) => n + (e.amount || 0), 0)
  const pointsLost = all.filter((e) => e.source === 'POINTS' && e.points < 0).reduce((n, e) => n - e.points, 0)
    + all.filter((e) => e.source === 'KPI' && !e.returned).reduce((n, e) => n + (e.points || 0), 0)
  const returnedThisMonth = all.filter((e) => (e.returned && (e.returnedAt ?? e.at).slice(0, 7) === month) || (e.source === 'POINTS' && e.points > 0 && e.at.slice(0, 7) === month)).length
  const top = [...rec.employees].sort((a, b) => b.count - a.count)
  const shown = allEmps ? top : top.slice(0, 5)

  return (
    <div dir="rtl" className="w-full space-y-4">
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-sky-50 to-indigo-50 p-6">
        <span aria-hidden className="pointer-events-none absolute left-6 top-1/2 -translate-y-1/2 text-[90px] opacity-20">📋</span>
        <h2 className="text-2xl font-black text-[#0f2040]">📒 سجل الانضباط الوظيفي</h2>
        <p className="mt-1 text-sm text-slate-600">كل خصم من أول يوم بالنظام يُسجل هنا، مع توضيح السبب والمبلغ ونقاط الانضباط — للمدير والمالك.</p>
      </div>

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-5">
        <Kpi label="إجمالي السجل" value={all.length} hint="جميع سجلات الانضباط الوظيفي" icon="📄" tone="text-sky-700" />
        <Kpi label="الخصومات اليوم" value={money(todayAmount)} hint={`إجمالي الخصومات في ${date(new Date().toISOString())}`} icon="💵" tone="bg-rose-50/50 text-rose-700" />
        <Kpi label="نقاط الانضباط" value={pointsLost} hint="إجمالي نقاط الخصم في النظام" icon="⭐" tone="bg-amber-50/50 text-amber-700" />
        <Kpi label="الموظفون بالسجل" value={rec.employees.length} hint="موظف عنده حركة انضباط" icon="👥" tone="text-violet-700" />
        <Kpi label="الاسترجاعات" value={returnedThisMonth} hint="سجل مسترجع هذا الشهر" icon="↩️" tone="bg-emerald-50/50 text-emerald-700" />
      </div>

      {rec.insights.length > 0 && <MatrixNote>{rec.insights.join(' ')}</MatrixNote>}

      <div className={`${card} p-4`}>
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <div>
            <h3 className="font-extrabold text-[#0f2040]">👥 الأكثر تكراراً في السجل</h3>
            <p className="text-[11px] text-slate-500">الموظفين الي عدهم أكبر عدد من تسجيلات الانضباط — اضغط على وحد حتى تشوف سجله</p>
          </div>
          {top.length > 5 && <button onClick={() => setAllEmps(!allEmps)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-[#2c5aad]">{allEmps ? 'أعلى ٥ بس' : `عرض جميع الموظفين (${top.length}) ‹`}</button>}
        </div>
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
          {shown.map((e, i) => (
            <button key={e.employeeId} type="button" onClick={() => setEmp(emp === e.employeeId ? null : e.employeeId)}
              className={`rounded-2xl border p-3 text-right transition ${emp === e.employeeId ? 'border-[#0f2040] ring-2 ring-sky-200' : 'border-slate-200 hover:shadow-md'}`}>
              <div className="flex items-start justify-between gap-2">
                <div className="flex min-w-0 items-center gap-2">
                  <span className={`grid h-10 w-10 shrink-0 place-items-center rounded-full font-black ${COLORS[i % COLORS.length]}`}>{initial(e.employeeName)}</span>
                  <div className="min-w-0"><b className="block truncate text-sm text-[#0f2040]">{e.employeeName}</b><span className="text-[11px] text-slate-500">الصافي: {money(e.net)}</span></div>
                </div>
                <span className="shrink-0 rounded-full bg-red-50 px-2 py-0.5 text-xs font-black text-red-700">{e.count} مرة</span>
              </div>
              <div className="mt-2 space-y-1 text-[11px] text-slate-600">
                <p>💵 إجمالي الخصم <b>{money(e.amount)}</b>{e.pointsLost > 0 && <> · نقاط −{e.pointsLost}</>}</p>
                <p>📅 آخر خصم {e.last ? date(e.last) : '—'}</p>
                {e.topReason && <p className="truncate">📝 السبب الأكثر: <b>{e.topReason}</b></p>}
              </div>
            </button>
          ))}
          {top.length === 0 && <p className="text-sm text-slate-400">ماكو خصومات مسجّلة.</p>}
        </div>
      </div>

      {sel && sel.insights.length > 0 && <MatrixNote>{sel.employeeName}: {sel.insights.join(' ')}</MatrixNote>}

      <div className={`${card} p-4`}>
        <div className="mb-3 flex flex-wrap items-center gap-2 text-sm">
          <b className="text-[#0f2040]">📒 {sel ? `سجل ${sel.employeeName}` : 'كل السجل'} ({entries.length})</b>
          <label className="flex items-center gap-1 text-xs">من <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5" /></label>
          <label className="flex items-center gap-1 text-xs">إلى <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5" /></label>
          <select value={src} onChange={(e) => setSrc(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs">
            <option value="ALL">كل الأنواع</option><option value="KPI">💸 خصم</option><option value="POINTS">📉 نقاط انضباط</option><option value="LEDGER">🧾 تسوية عهدة</option>
          </select>
          <select value={state} onChange={(e) => setState(e.target.value as typeof state)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs">
            <option value="ALL">كل الحالات</option><option value="OPEN">قائم</option><option value="RETURNED">مسترجع</option>
          </select>
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="🔍 البحث عن موظف، سبب، أو أي كلمة…" className="min-w-[200px] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-xs" />
          {emp && <button type="button" onClick={() => setEmp(null)} className="text-xs text-sky-700 underline">كل الموظفين</button>}
          <button type="button" onClick={() => csv(entries)} className="rounded-lg bg-[#0f2040] px-4 py-2 text-xs font-bold text-white">⬇️ تصدير Excel</button>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[900px] text-right text-xs">
            <thead className="bg-slate-50 text-slate-600">
              <tr>{['التاريخ', 'الموظف', 'النوع', 'المبلغ', 'النقاط', 'السبب', 'منو', 'الحجز', 'الحالة', ''].map((h) => <th key={h} className="p-2.5 font-bold">{h}</th>)}</tr>
            </thead>
            <tbody>
              {entries.map((e, i) => {
                const back = e.returned || (e.source === 'POINTS' && e.points > 0)
                return (
                  <tr key={e.source + e.id} className="border-t border-slate-100 hover:bg-slate-50/60">
                    <td className="whitespace-nowrap p-2.5">{date(e.at)}</td>
                    <td className="p-2.5"><span className="flex items-center gap-2 font-bold"><span className={`grid h-7 w-7 place-items-center rounded-full text-[11px] ${COLORS[i % COLORS.length]}`}>{initial(e.employeeName)}</span>{e.employeeName}</span></td>
                    <td className="whitespace-nowrap p-2.5"><span className="rounded-full bg-sky-50 px-2 py-0.5 text-[11px] font-bold text-sky-700">{SRC[e.source]}</span></td>
                    <td className="whitespace-nowrap p-2.5">{e.amount ? money(e.amount) : '—'}</td>
                    <td className="p-2.5">{e.points ? <span className={`rounded-full px-2 py-0.5 font-black ${e.source === 'POINTS' && e.points > 0 ? 'bg-emerald-50 text-emerald-700' : 'bg-rose-50 text-rose-700'}`}>{e.source === 'POINTS' ? (e.points > 0 ? `+${e.points}` : e.points) : `−${e.points}`}</span> : '—'}</td>
                    <td className="max-w-xs p-2.5">{e.reason || '—'}</td>
                    <td className="p-2.5">{e.byName ?? '—'}</td>
                    <td className="p-2.5">{e.bookingId ? <Link to={`/bookings?focus=${e.bookingId}`} className="text-sky-700 underline">{e.bookingCode}</Link> : '—'}</td>
                    <td className="whitespace-nowrap p-2.5"><span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-bold ${back ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}><i className={`h-1.5 w-1.5 rounded-full ${back ? 'bg-emerald-500' : 'bg-slate-400'}`} />{back ? `مسترجع${e.returnedAt ? ' ' + date(e.returnedAt) : ''}` : 'قائم'}</span></td>
                    <td className="p-2.5">{e.bookingId ? <Link to={`/bookings?focus=${e.bookingId}`} className="rounded-lg border border-slate-200 px-2 py-0.5 text-slate-500">⋯</Link> : null}</td>
                  </tr>
                )
              })}
              {entries.length === 0 && <tr><td colSpan={10} className="p-6 text-center text-slate-400">ماكو شي بهالفلترة.</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
