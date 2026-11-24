import { useState } from 'react'
import { api, type TechContactReport } from '../api'
import { useApi } from '../utils/useApi'
import { fmtDateTime } from '../utils/date'
import { LineIcon } from '../components/LineIcon'

// ═══ ☎️ تقارير اتصال التقنيين (ك 11-23، طلب زينا) ═══
// كل شي انحوّل لتقني (حجز أو استفسار طاقة شمسية) آخر ٣٠ يوم: متى انحوّل،
// متى رفع تقرير الاتصال وشنو كتب، وإذا عبرت ٢٤ ساعة وانخصمت عليه نقطة كي بي آي.
// تطلع للمراقب، ومدير العمليات ونائبه، ومدير الإدارة ونائبه.

// (ك 11-24) أرقام إنگليزية وبلا فارزة — «22/11 10:57 م» مثل صورة زينا (لهالشاشة بس)
const fmt = (s?: string | null) => fmtDateTime(s, { empty: '—', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', numberingSystem: 'latn' }).replace('،', '')

type Filter = 'ALL' | 'WAIT' | 'DONE' | 'KPI'

function state(r: TechContactReport): { t: string; cls: string; ico: 'check' | 'warn' | 'xCircle' | 'clock' } {
  if (r.contactedAt) return { t: 'رفع التقرير', cls: 'tr-ok', ico: 'check' }
  if (r.kpiAt) return { t: 'انخصمت نقطة', cls: 'tr-bad', ico: 'warn' }
  if (r.closed) return { t: 'انسكّر بلا تقرير', cls: 'tr-off', ico: 'xCircle' }
  const hours = (Date.now() - new Date(r.since).getTime()) / 36e5
  return { t: `ينتظر · باقي ${Math.max(0, Math.ceil(24 - hours))} س`, cls: 'tr-wait', ico: 'clock' }
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
  // (ك 11-24) حسب صورتي زينا — الألوان مقاسة منهن (.tr-*)
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h1 className="tr-title"><span className="tr-title-ico"><LineIcon n="phone" className="h-5 w-5" /></span>تقارير اتصال التقنيين</h1>
        <p className="tr-sub">كل حجز أو استفسار انحوّل لتقني آخر 30 يوم. التقني لازم يرفع تقرير الاتصال خلال 24 ساعة، وإلا تنخصم نقطة «تنفيذ المهام الموكلة إليه».</p>
      </div>
      <div className="flex flex-wrap gap-3">
        {([['ALL', 'الكل', 'layers'], ['WAIT', 'بانتظار تقرير', 'clock'], ['DONE', 'رفع التقرير', 'checkCircle'], ['KPI', 'انخصمت نقطة', 'warn']] as const).map(([k, t, i]) => (
          <button key={k} type="button" onClick={() => setFilter(k)} className={`tr-tab ${filter === k ? 'tr-tab-on' : ''}`}>
            <LineIcon n={i} className="h-5 w-5" />{t} ({count(k)})
          </button>
        ))}
      </div>
      {error && <p className="text-sm text-red-600">{error}</p>}
      {!rows && !error && <p className="tr-sub">جاري التحميل…</p>}
      {rows && list.length === 0 && <p className="tr-wrap tr-sub p-4">ماكو شي هنا.</p>}
      {list.length > 0 && (
        <div className="tr-wrap overflow-x-auto">
          <table className="w-full min-w-[900px] text-sm">
            <thead><tr className="tr-head">
              {([['gear', 'النوع'], ['#', 'الرقم'], ['user', 'الزبون'], ['wrench', 'التقني'], ['swap', 'حوّله'], ['clock', 'وقت التحويل'], ['cal', 'وقت التقرير'], ['file', 'تقرير الاتصال'], ['pulse', 'الحالة']] as const).map(([i, h]) => (
                <th key={h} className="tr-th"><span className="inline-flex items-center gap-2">{i === '#' ? <b className="tr-th-ico text-base">#</b> : i === 'swap' ? <span className="tr-th-ico text-base">⇄</span> : <LineIcon n={i} className="tr-th-ico h-5 w-5" />}{h}</span></th>
              ))}
            </tr></thead>
            <tbody>
              {list.map((r) => {
                const st = state(r)
                return (
                  <tr key={r.kind + r.ref} className="tr-row">
                    <td className="tr-td whitespace-nowrap">{r.kind === 'solar' ? <span className="inline-flex items-center gap-2"><span className="tr-sun">☀</span>استفسار</span> : <span className="inline-flex items-center gap-2"><LineIcon n="file" className="tr-doc h-5 w-5" />حجز</span>}</td>
                    <td className="tr-td font-bold tabular-nums">{r.ref}</td>
                    <td className="tr-td">{r.customer ?? '—'}</td>
                    <td className="tr-td font-bold">{r.tech}</td>
                    <td className="tr-td tr-muted text-xs">{r.by ?? '—'}</td>
                    <td className="tr-td whitespace-nowrap text-xs tabular-nums">{fmt(r.since)}</td>
                    <td className="tr-td whitespace-nowrap text-xs tabular-nums">{fmt(r.contactedAt)}</td>
                    <td className="tr-td tr-note text-xs">{r.note ?? '—'}</td>
                    <td className="tr-td"><span className={`tr-pill ${st.cls}`}><LineIcon n={st.ico} className="h-4 w-4" />{st.t}</span></td>
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
