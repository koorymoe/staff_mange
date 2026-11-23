import { useState } from 'react'
import { api, type TechContactReport } from '../api'
import { useApi } from '../utils/useApi'
import { fmtDateTime } from '../utils/date'

// ═══ ☎️ تقارير اتصال التقنيين (ك 11-23، طلب زينا) ═══
// كل شي انحوّل لتقني (حجز أو استفسار طاقة شمسية) آخر ٣٠ يوم: متى انحوّل،
// متى رفع تقرير الاتصال وشنو كتب، وإذا عبرت ٢٤ ساعة وانخصمت عليه نقطة كي بي آي.
// تطلع للمراقب، ومدير العمليات ونائبه، ومدير الإدارة ونائبه.

const fmt = (s?: string | null) => fmtDateTime(s, { empty: '—', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })

type Filter = 'ALL' | 'WAIT' | 'DONE' | 'KPI'

function state(r: TechContactReport): { t: string; cls: string } {
  if (r.contactedAt) return { t: 'رفع التقرير', cls: 'bg-emerald-100 text-emerald-800' }
  if (r.kpiAt) return { t: 'انخصمت نقطة', cls: 'bg-red-100 text-red-800' }
  if (r.closed) return { t: 'انسكّر بلا تقرير', cls: 'bg-slate-100 text-slate-600' }
  const hours = (Date.now() - new Date(r.since).getTime()) / 36e5
  return { t: `ينتظر · باقي ${Math.max(0, Math.ceil(24 - hours))} س`, cls: 'bg-amber-100 text-amber-800' }
}

export default function TechContactReportsPage() {
  const { data: rows, error } = useApi(() => api.techContactReports(), [], { fallback: 'تعذر جلب التقارير' })
  const [filter, setFilter] = useState<Filter>('ALL')
  const list = (rows ?? []).filter((r) =>
    filter === 'ALL' ? true
      : filter === 'DONE' ? !!r.contactedAt
        : filter === 'KPI' ? !!r.kpiAt
          : !r.contactedAt && !r.kpiAt && !r.closed)
  const count = (f: Filter) => (rows ?? []).filter((r) =>
    f === 'DONE' ? !!r.contactedAt : f === 'KPI' ? !!r.kpiAt : f === 'WAIT' ? !r.contactedAt && !r.kpiAt && !r.closed : true).length
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-brand-900">☎️ تقارير اتصال التقنيين</h1>
        <p className="text-sm text-slate-500">كل حجز أو استفسار انحوّل لتقني آخر ٣٠ يوم. التقني لازم يرفع تقرير الاتصال خلال ٢٤ ساعة، وإلا تنخصم نقطة «تنفيذ المهام الموكلة إليه».</p>
      </div>
      <div className="flex flex-wrap gap-2">
        {([['ALL', 'الكل'], ['WAIT', 'ينتظر تقرير'], ['DONE', 'رفع التقرير'], ['KPI', 'انخصمت نقطة']] as const).map(([k, t]) => (
          <button key={k} type="button" onClick={() => setFilter(k)}
            className={`rounded-xl px-3.5 py-2 text-xs font-bold ${filter === k ? 'bg-[#0f2040] text-white' : 'border border-slate-200 bg-white text-slate-600'}`}>
            {t} ({count(k)})
          </button>
        ))}
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {!rows && !error && <p className="text-sm text-slate-400">جاري التحميل…</p>}
      {rows && list.length === 0 && <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">ماكو شي هنا.</p>}
      {list.length > 0 && (
        <div className="overflow-x-auto rounded-xl border border-slate-200 bg-white">
          <table className="w-full min-w-[900px] text-sm">
            <thead><tr className="border-b bg-slate-50 text-[12px] text-slate-500">
              {['النوع', 'الرقم', 'الزبون', 'التقني', 'حوّله', 'وقت التحويل', 'وقت التقرير', 'تقرير الاتصال', 'الحالة'].map((h) => <th key={h} className="px-3 py-2 text-right font-bold">{h}</th>)}
            </tr></thead>
            <tbody>
              {list.map((r) => {
                const st = state(r)
                return (
                  <tr key={r.kind + r.ref} className="border-b border-slate-100">
                    <td className="px-3 py-2 text-xs">{r.kind === 'solar' ? '☀️ استفسار' : '📋 حجز'}</td>
                    <td className="px-3 py-2 font-bold tabular-nums">{r.ref}</td>
                    <td className="px-3 py-2">{r.customer ?? '—'}</td>
                    <td className="px-3 py-2 font-bold">{r.tech}</td>
                    <td className="px-3 py-2 text-xs text-slate-500">{r.by ?? '—'}</td>
                    <td className="whitespace-nowrap px-3 py-2 text-xs tabular-nums">{fmt(r.since)}</td>
                    <td className="whitespace-nowrap px-3 py-2 text-xs tabular-nums">{fmt(r.contactedAt)}</td>
                    <td className="px-3 py-2 text-xs">{r.note ?? '—'}</td>
                    <td className="px-3 py-2"><span className={`whitespace-nowrap rounded-full px-2 py-0.5 text-[11px] font-bold ${st.cls}`}>{st.t}</span></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
