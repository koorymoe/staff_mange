import { useEffect, useMemo, useState } from 'react'
import BookingChainView from './BookingChain'
import { Link } from 'react-router-dom'
import { api, type ChainEmployeeReport, type ChainStationStat, type RoleChainReport } from '../api'
import MatrixNote from './MatrixNote'
import { CHAIN_BAR, CHAIN_TONE, chainDuration } from './chainTone'

// ═══ ماتركس ٢٠٥٠ — تقارير الأدوار من «سلسلة الحجز» ═══
// طلب (ع) 10-05: تصميم أوضح من الجدول، وكلمات عراقية دقيقة.
// فوگ: نسبة الدور واتجاهها. بعدين خطوات الدور بأشرطة ملوّنة. بعدين كل موظف
// ببطاقة (الأضعف أول)، وتفاصيله تنفتح بالضغط: ملاحظات ماتركس، مقارنته
// بزملائه، والحجوزات الي تعثّر بيها مجمّعة حسب الخطوة.

const tone = (p: number) => (p >= 80 ? '#059669' : p >= 60 ? '#d97706' : '#dc2626')

function Ring({ pct, size = 64, label }: { pct: number | null; size?: number; label?: string }) {
  const r = size / 2 - 6, c = 2 * Math.PI * r
  const v = pct ?? 0
  return (
    <div className="relative shrink-0" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="rgba(148,163,184,0.28)" strokeWidth={6} />
        {pct != null && <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke={tone(v)} strokeWidth={6} strokeLinecap="round"
          strokeDasharray={`${(v / 100) * c} ${c}`} />}
      </svg>
      <div className="absolute inset-0 grid place-items-center text-center leading-none">
        <span className="font-black" style={{ fontSize: size / 4.2, color: pct == null ? '#94a3b8' : tone(v) }}>{pct == null ? '—' : `${v}%`}</span>
        {label && <span className="mt-0.5 text-[9px] text-slate-400">{label}</span>}
      </div>
    </div>
  )
}

// شريط مقسّم: بوقتها / بيها نقص / متأخرة / ما صارت.
function SplitBar({ ok, issue, late, missed }: { ok: number; issue: number; late: number; missed: number }) {
  const total = ok + late + missed
  if (!total) return <div className="h-2 rounded-full bg-slate-100" />
  const okClean = Math.max(ok - issue, 0)
  const parts = [
    { n: okClean, c: CHAIN_BAR.ok, t: 'بوقتها' }, { n: issue, c: CHAIN_BAR.issue, t: 'بوقتها بس بيها نقص' },
    { n: late, c: CHAIN_BAR.late, t: 'متأخرة' }, { n: missed, c: CHAIN_BAR.missed, t: 'ما صارت' },
  ]
  return (
    <div className="flex h-2 overflow-hidden rounded-full bg-slate-100">
      {parts.map((p) => p.n > 0 && <div key={p.t} title={`${p.t}: ${p.n}`} style={{ width: `${(p.n / total) * 100}%`, background: p.c }} />)}
    </div>
  )
}

function Legend({ s }: { s: { ok: number; issue: number; late: number; missed: number } }) {
  return (
    <div className="mt-1.5 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-slate-600">
      <span><i className="me-1 inline-block h-2 w-2 rounded-full" style={{ background: CHAIN_BAR.ok }} />بوقتها {Math.max(s.ok - s.issue, 0)}</span>
      {s.issue > 0 && <span><i className="me-1 inline-block h-2 w-2 rounded-full" style={{ background: CHAIN_BAR.issue }} />بيها نقص {s.issue}</span>}
      <span><i className="me-1 inline-block h-2 w-2 rounded-full" style={{ background: CHAIN_BAR.late }} />متأخرة {s.late}</span>
      <span><i className="me-1 inline-block h-2 w-2 rounded-full" style={{ background: CHAIN_BAR.missed }} />ما صارت {s.missed}</span>
    </div>
  )
}

function StationCard({ s }: { s: ChainStationStat }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <div className="flex items-baseline justify-between gap-2">
        <b className="text-sm text-slate-800">{s.title}</b>
        <span className="text-[11px] text-slate-500">{s.count} مرة</span>
      </div>
      <p className="mb-2 text-[11px] text-slate-500">المعتاد: {chainDuration(s.medianMin)}</p>
      <SplitBar ok={s.ok} issue={s.issue} late={s.late} missed={s.missed} />
      <Legend s={s} />
    </div>
  )
}

// مقارنة الموظف بزملائه بكل خطوة — شريطين: هو والفريق.
function Compare({ s }: { s: ChainStationStat }) {
  const me = s.medianMin, team = s.teamMedianMin
  const max = Math.max(me ?? 0, team ?? 0, 1)
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-2.5 text-xs">
      <div className="mb-1.5 flex items-center justify-between">
        <b>{s.title}</b>
        <span className="text-slate-500">{s.count} مرة</span>
      </div>
      <div className="space-y-1">
        <div className="flex items-center gap-2"><span className="w-12 shrink-0 text-slate-500">هو</span>
          <div className="h-2 flex-1 rounded-full bg-slate-100"><div className="h-2 rounded-full bg-sky-500" style={{ width: `${((me ?? 0) / max) * 100}%` }} /></div>
          <span className="w-28 shrink-0 text-left">{chainDuration(me)}</span></div>
        <div className="flex items-center gap-2"><span className="w-12 shrink-0 text-slate-500">زملاؤه</span>
          <div className="h-2 flex-1 rounded-full bg-slate-100"><div className="h-2 rounded-full bg-slate-400" style={{ width: `${((team ?? 0) / max) * 100}%` }} /></div>
          <span className="w-28 shrink-0 text-left">{chainDuration(team)}</span></div>
      </div>
      <div className="mt-2"><SplitBar ok={s.ok} issue={s.issue} late={s.late} missed={s.missed} /></div>
    </div>
  )
}

// الحجوزات الي تعثّر بيها — مجمّعة حسب الخطوة، رقائق أكواد قابلة للضغط.
function Broken({ bad }: { bad: NonNullable<ChainEmployeeReport['bad']> }) {
  const [all, setAll] = useState<Record<string, boolean>>({})
  // الضغط على الكود يفتح الدليل تحته: ليش، ومنو سجّل وتواصل وثبّت ومتى (طلب (ع) 10-06).
  const [openKey, setOpenKey] = useState<string | null>(null)
  const groups = useMemo(() => {
    const m = new Map<string, typeof bad>()
    for (const b of bad) m.set(b.station, [...(m.get(b.station) ?? []), b])
    return [...m.entries()].sort((a, b) => b[1].length - a[1].length)
  }, [bad])
  return (
    <div className="space-y-2">
      {groups.map(([station, rows]) => {
        const shown = all[station] ? rows : rows.slice(0, 6)
        const opened = rows.find((b, i) => openKey === station + i + b.id)
        return (
          <div key={station} className="rounded-lg border border-slate-200 bg-white p-2.5">
            <p className="mb-1.5 text-xs font-bold text-slate-700">{station} <span className="font-normal text-slate-500">— {rows.length} حجز</span></p>
            <div className="flex flex-wrap gap-1.5">
              {shown.map((b, i) => (
                <button type="button" key={i} onClick={() => setOpenKey(openKey === station + i + b.id ? null : station + i + b.id)} title={b.note}
                  className={`rounded-full border px-2 py-0.5 text-[11px] font-bold hover:underline ${CHAIN_TONE[b.status].cls} ${openKey === station + i + b.id ? 'ring-2 ring-sky-400' : ''}`}>
                  {CHAIN_TONE[b.status].icon} {b.code}
                </button>
              ))}
              {rows.length > 6 && !all[station] && (
                <button type="button" onClick={() => setAll((a) => ({ ...a, [station]: true }))} className="rounded-full px-2 py-0.5 text-[11px] text-sky-700 hover:underline">
                  + {rows.length - 6} ثانية
                </button>
              )}
            </div>
            {opened && (
              <div className="mt-2 space-y-1.5 rounded-lg border border-sky-200 bg-sky-50/60 p-2.5 text-xs text-slate-700">
                <p className="font-bold">🔎 حجز {opened.code} · {station}: <span className="font-normal">{opened.note || '—'}</span></p>
                {opened.trail && opened.trail.length > 0 && (
                  <ul className="space-y-0.5">{opened.trail.map((t, k) => <li key={k}>{t}</li>)}</ul>
                )}
                <div className="flex flex-wrap gap-3">
                  <Link to={`/bookings?focus=${opened.id}`} className="font-bold text-sky-700 underline">افتح الحجز ←</Link>
                </div>
                <details><summary className="cursor-pointer font-bold text-violet-800">🔗 سلسلة الحجز كاملة (كل الخطوات ومنو سواها)</summary><div className="mt-2"><BookingChainView bookingId={opened.id} /></div></details>
              </div>
            )}
          </div>
        )
      })}
      <p className="text-[11px] text-slate-400">اضغط على أي كود حتى يطلع الدليل: ليش، ومنو سجّل الحجز ومنو تواصل ومنو ثبّت ومتى.</p>
    </div>
  )
}

function EmployeeCard({ e, open, onToggle }: { e: ChainEmployeeReport; open: boolean; onToggle: () => void }) {
  const unknown = !e.id
  const sum = e.stations.reduce((a, s) => ({ ok: a.ok + s.ok, issue: a.issue + s.issue, late: a.late + s.late, missed: a.missed + s.missed }), { ok: 0, issue: 0, late: 0, missed: 0 })
  const diff = e.prevPct != null && e.total ? e.onTimePct - e.prevPct : null
  return (
    <div className={`rounded-2xl border ${unknown ? 'border-dashed border-slate-300 bg-slate-50' : 'border-slate-200 bg-white'} ${open ? 'ring-2 ring-sky-200' : ''}`}>
      <button type="button" onClick={onToggle} className="flex w-full items-center gap-3 p-3 text-right">
        <Ring pct={e.total ? e.onTimePct : null} size={56} label="بوقتها" />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-extrabold text-slate-800">{unknown ? 'خطوات ما معروف منو مسؤولها' : e.name}</p>
          <p className="text-[11px] text-slate-500">
            {e.total} خطوة
            {diff != null && <> · الفترة الفاتت {e.prevPct}% <b style={{ color: diff >= 0 ? '#059669' : '#dc2626' }}>{diff >= 0 ? `▲${diff}` : `▼${-diff}`}</b></>}
            {e.ratingAvg != null && <> · تقييم الليدرية {e.ratingAvg}/5</>}
          </p>
          {e.worst && <p className="text-[11px] text-slate-600">يتعثّر أكثر بـ: <b>{e.worst}</b></p>}
          <div className="mt-1.5"><SplitBar {...sum} /></div>
        </div>
        <span className="text-slate-400">{open ? '▴' : '▾'}</span>
      </button>
      {open && (
        <div className="space-y-3 border-t border-slate-100 p-3">
          {unknown && <MatrixNote>هاي خطوات بالحجوزات ما انسجّل منو سوّاها، لأن الخطوة نفسها ما انسجّلت بالنظام (مثلاً محد ضغط «تواصلت»). يعني المشكلة بالتسجيل قبل ما تكون بالشخص.</MatrixNote>}
          {e.insights.length > 0 && (
            <div className="matrix-note">
              🤖 <b className="mx-tag">ماتركس:</b>
              <ul className="mt-1 list-inside list-disc space-y-0.5">{e.insights.map((t, i) => <li key={i}>{t}</li>)}</ul>
            </div>
          )}
          {!unknown && e.stations.length > 0 && (
            <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">{e.stations.map((s) => <Compare key={s.key} s={s} />)}</div>
          )}
          {e.ratingNotes.length > 0 && (
            <div className="rounded-lg border border-slate-200 bg-white p-2.5 text-xs">
              <b>ملاحظات الليدرية عنه:</b>
              {e.ratingNotes.map((n, i) => <p key={i} className="text-slate-600">• {n}</p>)}
            </div>
          )}
          {e.bad && e.bad.length > 0 && (
            <div>
              <p className="mb-1.5 text-xs font-extrabold text-slate-700">وين تعثّر:</p>
              <Broken bad={e.bad} />
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default function MatrixRoleChains() {
  const [roles, setRoles] = useState<{ key: string; title: string }[]>([])
  const [role, setRole] = useState('COORDINATOR')
  const [days, setDays] = useState(30)
  const [rep, setRep] = useState<{ k: string; r: RoleChainReport | null } | null>(null)
  const [open, setOpen] = useState<string | null>(null)
  const key = `${role}:${days}`
  const loading = rep?.k !== key
  const r = loading ? null : rep.r

  useEffect(() => { void api.getRoleChainRoles().then(setRoles).catch(() => {}) }, [])
  useEffect(() => {
    let alive = true
    void api.getRoleChain(role, days).catch(() => null).then((d) => { if (alive) setRep({ k: key, r: d }) })
    return () => { alive = false }
  }, [role, days, key])

  const people = r ? [...r.employees.filter((e) => e.id), ...r.employees.filter((e) => !e.id)] : []
  const teamDiff = r && r.prevPct != null ? r.teamPct - r.prevPct : null

  return (
    <div dir="rtl" className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50/60 p-3 text-slate-800 sm:p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p className="text-base font-extrabold text-[#0f2040]">🔗 سلسلة الحجز — تقارير الأدوار</p>
          <p className="text-[11px] text-slate-500">كل دور يتحاسب على خطواته هو بس</p>
        </div>
        <select value={days} onChange={(e) => setDays(Number(e.target.value))} className="rounded-lg border border-slate-200 bg-white px-2 py-1 text-xs">
          {[7, 30, 60, 90].map((d) => <option key={d} value={d}>آخر {d} يوم</option>)}
        </select>
      </div>

      <div className="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-1">
        {roles.map((x) => (
          <button key={x.key} type="button" onClick={() => { setRole(x.key); setOpen(null) }}
            className={`shrink-0 rounded-full px-3.5 py-1.5 text-xs font-bold transition ${role === x.key ? 'bg-[#0f2040] text-white shadow' : 'bg-white text-slate-700 ring-1 ring-slate-200 hover:bg-slate-100'}`}>{x.title}</button>
        ))}
      </div>

      {loading ? <p className="text-xs text-slate-400">ماتركس يحلّل…</p> : !r ? <p className="text-xs text-red-600">ما گدرت أجيب التقرير</p> : (
        <>
          <div className="flex flex-wrap items-center gap-4 rounded-2xl border border-slate-200 bg-white p-4">
            <Ring pct={r.stations.length ? r.teamPct : null} size={88} label="بوقتها" />
            <div className="min-w-[14rem] flex-1">
              <p className="text-lg font-black text-slate-800">{r.title}</p>
              <p className="text-xs text-slate-500">{r.bookings} حجز بآخر {r.days} يوم
                {teamDiff != null && <> · الفترة الفاتت {r.prevPct}% <b style={{ color: teamDiff >= 0 ? '#059669' : '#dc2626' }}>{teamDiff >= 0 ? `▲${teamDiff}` : `▼${-teamDiff}`}</b></>}
              </p>
              <MatrixNote className="mt-2">{r.insights.join(' ')}</MatrixNote>
            </div>
          </div>

          {r.stations.length > 0 && (
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">{r.stations.map((s) => <StationCard key={s.key} s={s} />)}</div>
          )}

          <div className="space-y-2">
            <p className="text-xs font-extrabold text-slate-600">الموظفين — الأضعف أول (اضغط على أي واحد حتى تشوف تفاصيله)</p>
            {people.map((e) => (
              <EmployeeCard key={e.id || 'unknown'} e={e} open={open === (e.id || 'unknown')}
                onToggle={() => setOpen(open === (e.id || 'unknown') ? null : (e.id || 'unknown'))} />
            ))}
            {people.length === 0 && <p className="rounded-xl bg-white p-4 text-center text-xs text-slate-400">ماكو شغل لهذا الدور بهالفترة.</p>}
          </div>
        </>
      )}
    </div>
  )
}
