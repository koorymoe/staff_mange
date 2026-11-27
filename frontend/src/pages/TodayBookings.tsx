// (ك) 11-27 زينا: تصميم «حجوزات اليوم» (من المتابعة بمكتب المدير) — نهاري وليلي.
// الأرقام كلها من نفس قائمة الخادم (drill today_scheduled). المبلغ والدفع ماكو بالقائمة، فيطلعن «—».
import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import type { DrillResult, DrillRow } from '../apiTypes'
import { BackLink } from '../ui/BackLink'
import { LineIcon } from '../components/LineIcon'
import './TodayBookings.css'

type Tab = 'all' | 'done' | 'run' | 'not' | 'pay' | 'rev'
// «بانتظار الدفع» = عليه آجل مفتوح (CreditSale ما تسدّد) — قرار زينا 11-27
const kind = (r: DrillRow): Tab | null => (r.creditDue ?? 0) > 0 ? 'pay' : r.status === 'COMPLETED' ? 'done' : r.station === 'بالتنفيذ' ? 'run' : null
const DAYS = ['الأحد', 'الاثنين', 'الثلاثاء', 'الأربعاء', 'الخميس', 'الجمعة', 'السبت']

export default function TodayBookings({ d }: { d: DrillResult }) {
  const [q, setQ] = useState('')
  const [tab, setTab] = useState<Tab>('all')
  const [station, setStation] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [size, setSize] = useState(10)
  const [page, setPage] = useState(1)

  const all = d.rows
  const done = all.filter((r) => kind(r) === 'done').length
  const run = all.filter((r) => kind(r) === 'run').length
  const pay = all.filter((r) => kind(r) === 'pay')
  const payDue = pay.reduce((t, r) => t + (r.creditDue ?? 0), 0)
  const revRows = all.filter((r) => r.status === 'COMPLETED' && (r.collected ?? 0) > 0)
  const revenue = revRows.reduce((t, r) => t + (r.collected ?? 0), 0)
  const notDone = all.filter((r) => r.status !== 'COMPLETED').length
  const iqd = (n: number) => `${n.toLocaleString('en-US')} د.ع`
  const pick = (t: Tab) => { setTab(t); setPage(1); setStation(''); document.querySelector('.tb-panel')?.scrollIntoView({ behavior: 'smooth' }) }
  const stations = useMemo(() => [...new Set(all.map((r) => r.station))], [all])
  const today = useMemo(() => {
    const s = new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
    return `${DAYS[new Date(s + 'T12:00:00').getDay()]} ${s}`
  }, [])

  const rows = all.filter((r) => (tab === 'all' || (tab === 'not' ? r.status !== 'COMPLETED' : tab === 'rev' ? revRows.includes(r) : kind(r) === tab)) && (!station || r.station === station)
    && (!q.trim() || [r.code, r.customer, r.phone, r.service, r.crew, r.responsible].some((v) => v?.includes(q.trim()))))
  const pages = Math.max(1, Math.ceil(rows.length / size))
  const cur = Math.min(page, pages)
  const shown = rows.slice((cur - 1) * size, cur * size)

  const exportCsv = () => {
    const head = ['الكود', 'الزبون', 'الهاتف', 'الخدمة', 'الموعد', 'الحالة', 'المبلغ', 'الآجل', 'الكادر', 'المسؤول']
    const esc = (v: unknown) => `"${String(v ?? '').replace(/"/g, '""')}"`
    const body = rows.map((r) => [r.code, r.customer, r.phone, r.service, r.scheduledAt, r.station, r.amount, r.creditDue, r.crew, r.responsible].map(esc).join(','))
    const blob = new Blob(['﻿' + [head.map(esc).join(','), ...body].join('\n')], { type: 'text/csv;charset=utf-8' })
    const a = document.createElement('a'); a.href = URL.createObjectURL(blob); a.download = `حجوزات-اليوم.csv`; a.click()
    URL.revokeObjectURL(a.href)
  }

  const cards = [
    { c: 'blue', ic: 'cal', l: 'إجمالي حجوزات اليوم', v: `${d.count}`, u: 'حجز', t: 'all' as Tab },
    { c: 'green', ic: 'checkCircle', l: 'حجوزات مكتملة', v: `${done}`, u: 'تم تنفيذها وتحصيلها', t: 'done' as Tab },
    { c: 'slate', ic: 'xCircle', l: 'حجوزات غير مكتملة', v: `${notDone}`, u: 'ما خلصت بعد', t: 'not' as Tab },
    { c: 'amber', ic: 'hourglass', l: 'حجوزات بانتظار الدفع', v: `${pay.length}`, u: payDue > 0 ? `آجل: ${iqd(payDue)}` : 'لم يتم تأكيد الدفع', t: 'pay' as Tab },
    { c: 'rose', ic: 'clock', l: 'حجوزات قيد التنفيذ', v: `${run}`, u: 'جارية حالياً', t: 'run' as Tab },
    { c: 'violet', ic: 'coins', l: 'الإيرادات المحققة', v: iqd(revenue), u: `من ${revRows.length} حجوزات مكتملة`, t: 'rev' as Tab },
  ]
  const tabs: { k: Tab; l: string; n: number | null; ic: string }[] = [
    { k: 'all', l: 'الكل', n: all.length, ic: 'cal' }, { k: 'done', l: 'مكتملة', n: done, ic: 'check' },
    { k: 'run', l: 'قيد التنفيذ', n: run, ic: 'play' }, { k: 'not', l: 'غير مكتملة', n: notDone, ic: 'xCircle' }, { k: 'pay', l: 'بانتظار الدفع', n: pay.length, ic: 'hourglass' },
  ]
  const pill = (r: DrillRow) => {
    const k = kind(r)
    if (k === 'done') return <span className="tb-pill ok"><LineIcon n="check" className="h-3.5 w-3.5" />مكتمل</span>
    if (k === 'pay') return <span className="tb-pill pay"><LineIcon n="hourglass" className="h-3.5 w-3.5" />بانتظار الدفع</span>
    if (k === 'run') return <span className="tb-pill run"><LineIcon n="play" className="h-3.5 w-3.5" />قيد التنفيذ</span>
    return <span className="tb-pill">{r.station}</span>
  }

  return (
    <div className="tb-page space-y-4" dir="rtl">
      <div className="tb-hero flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="tb-logo"><LineIcon n="cal" className="h-6 w-6" /></span>
          <div>
            <h1 className="tb-title text-[18px] font-black">حجوزات اليوم</h1>
            <p className="tb-sub text-[12px]">{d.explain}</p>
          </div>
        </div>
        <BackLink to="/" label="رجوع ←" className="tb-back" />
      </div>

      <div className="tb-cards">
        {cards.map((k) => (
          <button type="button" key={k.l} onClick={() => pick(k.t)} className={`tb-card ${k.c} ${tab === k.t ? 'on' : ''} text-right`}>
            <span className="tb-ic"><LineIcon n={k.ic} className="h-5 w-5" /></span>
            <div className="min-w-0">
              <p className="tb-cl">{k.l}</p>
              <p className="tb-cv">{k.v}</p>
              <p className="tb-cu">{k.u}</p>
            </div>
          </button>
        ))}
      </div>

      <div className="tb-bar">
        <div className="tb-tabs">
          {tabs.map((t) => (
            <button key={t.k} type="button" className={tab === t.k ? 'on' : ''} onClick={() => { setTab(t.k); setPage(1) }}>
              <LineIcon n={t.ic} className="h-5 w-5 ic" />{t.l} <span>({t.n ?? '—'})</span>
            </button>
          ))}
        </div>
        <div className="relative flex gap-2">
          <button type="button" className="tb-btn sq" onClick={() => setFilterOpen((o) => !o)} aria-label="فلتر"><LineIcon n="filter" className="h-5 w-5" /></button>
          {filterOpen && (
            <div className="tb-pop">
              <p className="mb-1 text-[12px] font-bold">وين واكف</p>
              <select value={station} onChange={(e) => { setStation(e.target.value); setPage(1) }} className="tb-input w-full">
                <option value="">الكل</option>
                {stations.map((s) => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
          )}
          <button type="button" className="tb-btn" onClick={() => window.print()}><LineIcon n="printer" className="h-5 w-5" />طباعة</button>
          <button type="button" className="tb-btn" onClick={exportCsv}><LineIcon n="file" className="h-5 w-5" />تصدير التقرير</button>
        </div>
      </div>

      <div className="tb-panel">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <label className="tb-search">
            <LineIcon n="search" className="h-5 w-5" />
            <input value={q} onChange={(e) => { setQ(e.target.value); setPage(1) }} placeholder="بحث بالكود، الزبون، الخدمة..." />
          </label>
          <span className="tb-chip"><LineIcon n="cal" className="h-5 w-5" />اليوم: {today}</span>
        </div>
        {shown.length === 0 ? (
          <p className="py-10 text-center tb-sub">{d.count === 0 ? 'ماكو ولا حجز اليوم.' : 'ماكو نتيجة للبحث.'}</p>
        ) : (
          <table className="tb-table w-full text-right">
            <thead>
              <tr><th>#</th>{([['user','الزبون'],['layers','الخدمة'],['cal','الموعد'],['pulse','الحالة'],['coins','المبلغ'],['star','الكادر'],['user','المسؤول'],['','صادره'],['gear','اجراءات']] as const).map(([ic, l]) => <th key={l}><span className="inline-flex items-center gap-1.5">{ic && <LineIcon n={ic} />}{l}</span></th>)}</tr>
            </thead>
            <tbody>
              {shown.map((r) => (
                <tr key={r.id}>
                  <td><span className="tb-code">{r.code}</span></td>
                  <td>{r.customer ?? '—'}{r.phone && <span className="tb-phone" dir="ltr">{r.phone}</span>}</td>
                  <td>{r.service ?? '—'}</td>
                  <td className="tabular-nums">{r.scheduledAt ? r.scheduledAt.slice(-5) : '—'}</td>
                  <td>{pill(r)}</td>
                  <td className="tabular-nums whitespace-nowrap">{r.amount ? iqd(r.amount) : '-'}</td>
                  <td>{r.crew ? <><span className="tb-star">⭐</span> {r.crew}</> : '—'}</td>
                  <td>{r.responsible ?? '—'}</td>
                  <td>{r.ageDays == null ? '—' : r.ageDays === 0 ? 'اليوم' : `${r.ageDays} يوم`}</td>
                  <td><Link to={`/bookings?focus=${r.id}`} className="tb-dots" aria-label="فتح الحجز">•••</Link></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        <div className="tb-foot">
          <div className="flex items-center gap-3">
            <select value={size} onChange={(e) => { setSize(Number(e.target.value)); setPage(1) }} className="tb-input">
              {[10, 20, 50].map((n) => <option key={n} value={n}>{n}</option>)}
            </select>
            <span>عرض <b>{shown.length}</b> من أصل <b>{rows.length}</b> حجز</span>
          </div>
          <div className="tb-pager">
            <button type="button" className="tb-pg" disabled={cur === 1} onClick={() => setPage(1)}>»</button>
            <button type="button" className="tb-pg" disabled={cur === 1} onClick={() => setPage(cur - 1)}><LineIcon n="right" /></button>
            {Array.from({ length: pages }, (_, i) => i + 1).map((p) => (
              <button key={p} type="button" className={`tb-pg ${p === cur ? 'on' : ''}`} onClick={() => setPage(p)}>{p}</button>
            ))}
            <button type="button" className="tb-pg" disabled={cur === pages} onClick={() => setPage(cur + 1)}><LineIcon n="left" /></button>
            <button type="button" className="tb-pg" disabled={cur === pages} onClick={() => setPage(pages)}>«</button>
          </div>
        </div>
        {d.rows.length < d.count && <p className="mt-2 text-[12px] tb-sub">معروض {d.rows.length} من {d.count}.</p>}
      </div>
    </div>
  )
}
