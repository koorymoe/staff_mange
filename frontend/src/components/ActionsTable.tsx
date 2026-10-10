import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'

// ═══ لوحة «الإجراءات» — بتصميم (ع) 10-04 ═══
// كل شي ينتظر قرار أو متابعة بجدول واحد: مهام إضافية، إجازات، طلبات كادر،
// حذف حجوزات، شكاوى، مشتريات، قرارات ماتركس، حجوزات بلا كادر، تكرارات،
// إنجازات بلا تقييم. كل صف من بيانات النظام الحقيقية، ويودّيك لشاشته.
// المصدر الي ما عندك صلاحيته يرجع فاضي بهدوء (كلها قراءة GET).

type Status = 'LATE' | 'FOLLOW' | 'PROGRESS' | 'DONE'
interface Row {
  key: string
  type: string
  icon: string
  desc: string
  who: string
  dept: string
  due: string | null
  status: Status
  to: string
  act: string
}

const STATUS: Record<Status, { label: string; cls: string; icon: string }> = {
  LATE: { label: 'متأخر', cls: 'bg-red-50 text-red-700', icon: '⚠️' },
  FOLLOW: { label: 'يحتاج متابعة', cls: 'bg-amber-50 text-amber-700', icon: '🕐' },
  PROGRESS: { label: 'قيد التنفيذ', cls: 'bg-sky-50 text-sky-700', icon: '⚙️' },
  DONE: { label: 'خالص', cls: 'bg-emerald-50 text-emerald-700', icon: '✅' },
}

const DAY = 24 * 3600 * 1000
const olderThan = (iso: string, days: number, now: number) => now - new Date(iso).getTime() > days * DAY
const fmtDate = (s: string | null) => (s ? new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) : '—')
const clip = (s: string | null | undefined, n = 90) => { const t = (s ?? '').trim(); return t.length > n ? t.slice(0, n) + '…' : t }

async function load(now: number): Promise<Row[]> {
  const rows: Row[] = []
  const safe = async <T,>(p: Promise<T>): Promise<T | null> => { try { return await p } catch { return null } }
  const [tasks, leaves, staff, deletes, complaints, procurement, matrix, dups, achievements] = await Promise.all([
    safe(api.getExtraTasks({})), safe(api.getLeaveInbox('PENDING')), safe(api.getStaffRequests()),
    safe(api.getBookingDeleteRequests('PENDING')), safe(api.getComplaints()), safe(api.getProcurementRequests()),
    safe(api.getMatrixDecisions()), safe(api.getDuplicateCandidates(undefined, 'PENDING')), safe(api.getAchievements({ limit: 200 })),
  ])

  for (const t of tasks ?? []) {
    if (t.status === 'CANCELLED') continue
    // الخالصة: بس آخر ٧ أيام — حتى «خالص» يبيّن شغل هالأسبوع مو كل التاريخ.
    if (t.status === 'DONE' && (!t.doneAt || olderThan(t.doneAt, 7, now))) continue
    const status: Status = t.status === 'DONE' ? 'DONE' : t.overdue ? 'LATE' : t.status === 'IN_PROGRESS' ? 'PROGRESS' : 'FOLLOW'
    rows.push({ key: 'task' + t.id, type: 'مهمة إضافية', icon: '📋', desc: clip(t.title + (t.description ? ` — ${t.description}` : '')),
      who: t.assignedToName ?? '—', dept: 'المهام', due: t.dueAt ?? null, status, to: '/extra-tasks', act: status === 'DONE' ? 'شوف' : 'تابع' })
  }
  for (const l of leaves ?? []) {
    rows.push({ key: 'leave' + l.id, type: 'طلب إجازة', icon: '🌴', desc: clip(`${l.kind === 'URGENT' ? 'عاجلة · ' : ''}${fmtDate(l.startDate)} ← ${fmtDate(l.endDate)}${l.reason ? ` — ${l.reason}` : ''}`),
      who: l.employeeName, dept: 'الموارد البشرية', due: l.startDate, status: new Date(l.startDate).getTime() < now ? 'LATE' : 'FOLLOW', to: '/leaves', act: 'قرّر' })
  }
  for (const s of staff ?? []) {
    if (s.status !== 'PENDING') continue
    rows.push({ key: 'staff' + s.id, type: 'طلب كادر', icon: '👷', desc: clip(`${s.projectName ?? 'مشروع'}${s.notes ? ` — ${s.notes}` : ''}`),
      who: '—', dept: 'المشاريع', due: s.createdAt, status: olderThan(s.createdAt, 2, now) ? 'LATE' : 'FOLLOW', to: '/staff-requests', act: 'وفّر' })
  }
  for (const d of deletes ?? []) {
    rows.push({ key: 'del' + d.id, type: 'طلب حذف حجز', icon: '🗑️', desc: clip(`${d.bookingCode} — ${d.reason}`),
      who: d.requestedByName, dept: 'الحجوزات', due: d.createdAt, status: olderThan(d.createdAt, 2, now) ? 'LATE' : 'FOLLOW', to: '/booking-delete-requests', act: 'قرّر' })
  }
  for (const c of complaints ?? []) {
    if (c.status !== 'NEW' && c.status !== 'IN_PROGRESS') continue
    const status: Status = olderThan(c.createdAt, 3, now) ? 'LATE' : c.status === 'IN_PROGRESS' ? 'PROGRESS' : 'FOLLOW'
    rows.push({ key: 'cmp' + c.id, type: 'شكوى زبون', icon: '🎧', desc: clip(`${c.customer?.name ?? ''} — ${c.description}`),
      who: '—', dept: 'خدمة الزبائن', due: c.createdAt, status, to: '/complaints', act: 'تابع' })
  }
  for (const p of procurement ?? []) {
    if (p.status !== 'PENDING' && p.status !== 'IN_PROGRESS') continue
    const items = (p.items ?? []).map((i) => i.productName).filter(Boolean).slice(0, 3).join('، ')
    const status: Status = olderThan(p.createdAt, 3, now) ? 'LATE' : p.status === 'IN_PROGRESS' ? 'PROGRESS' : 'FOLLOW'
    rows.push({ key: 'pr' + p.id, type: 'شراء مواد', icon: '🛒', desc: clip(`${p.code}${items ? ` — ${items}` : ''}`),
      who: p.requestedBy?.name ?? '—', dept: 'المشتريات', due: p.createdAt, status, to: '/procurement', act: 'جهّز' })
  }
  for (const m of matrix?.pending ?? []) {
    rows.push({ key: 'mx' + m.id, type: 'قرار ماتركس', icon: '🤖', desc: clip(`${m.title}${m.bookingCode ? ` (${m.bookingCode})` : ''}`),
      who: m.employeeName ?? '—', dept: 'ماتركس', due: m.createdAt, status: olderThan(m.createdAt, 3, now) ? 'LATE' : 'FOLLOW', to: '/matrix/decisions', act: 'قرّر' })
  }
  for (const u of matrix?.unstaffed ?? []) {
    rows.push({ key: 'un' + u.id, type: 'حجز بلا كادر', icon: '📅', desc: `${u.code} — موعده قريب وما عليه كادر`,
      who: '—', dept: 'الحجوزات', due: u.scheduledAt, status: new Date(u.scheduledAt).getTime() < now ? 'LATE' : 'FOLLOW', to: '/coordinator', act: 'كلّف' })
  }
  for (const d of dups ?? []) {
    const a = d.bookingA?.code ?? d.customerA?.name ?? '', b = d.bookingB?.code ?? d.customerB?.name ?? ''
    rows.push({ key: 'dup' + d.id, type: 'تكرار مشتبه', icon: '👯', desc: clip(`${a} و${b} — ${d.matchReason}`),
      who: '—', dept: 'ماتركس', due: d.detectedAt, status: 'FOLLOW', to: '/duplicate-review', act: 'دقّق' })
  }
  for (const a of achievements ?? []) {
    if (a.reviewStatus !== 'PENDING') continue
    rows.push({ key: 'ach' + a.id, type: 'إنجاز بلا تقييم', icon: '📝', desc: clip(a.reportText),
      who: a.employeeName ?? '—', dept: 'الموظفين', due: a.createdAt, status: olderThan(a.createdAt, 2, now) ? 'LATE' : 'FOLLOW', to: '/achievements', act: 'قيّم' })
  }
  const order: Record<Status, number> = { LATE: 0, FOLLOW: 1, PROGRESS: 2, DONE: 3 }
  return rows.sort((x, y) => order[x.status] - order[y.status] || (x.due ?? '').localeCompare(y.due ?? ''))
}

const PAGE = 10

export default function ActionsTable() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [q, setQ] = useState('')
  const [dept, setDept] = useState('')
  const [type, setType] = useState('')
  const [status, setStatus] = useState<Status | ''>('')
  const [page, setPage] = useState(1)

  useEffect(() => {
    let alive = true
    load(Date.now()).then((r) => { if (alive) setRows(r) }).catch(() => { if (alive) setRows([]) })
    return () => { alive = false }
  }, [])

  const all = useMemo(() => rows ?? [], [rows])
  const filtered = useMemo(() => all.filter((r) =>
    (!dept || r.dept === dept) && (!type || r.type === type) && (!status || r.status === status)
    && (!q.trim() || `${r.type} ${r.desc} ${r.who} ${r.dept}`.includes(q.trim()))), [all, q, dept, type, status])
  const pages = Math.max(1, Math.ceil(filtered.length / PAGE))
  const cur = Math.min(page, pages)
  const shown = filtered.slice((cur - 1) * PAGE, cur * PAGE)
  const count = (s: Status) => all.filter((r) => r.status === s).length
  const opts = (k: 'dept' | 'type') => [...new Set(all.map((r) => r[k]))].sort()
  const reset = () => { setQ(''); setDept(''); setType(''); setStatus(''); setPage(1) }

  const cards: { s: Status | ''; label: string; hint: string; icon: string; cls: string; n: number }[] = [
    { s: '', label: 'الكل', hint: 'كل الإجراءات', icon: '☰', cls: 'border-slate-200 bg-white text-slate-800', n: all.length },
    { s: 'FOLLOW', label: 'تحتاج متابعة', hint: 'تنتظر قرارك أو متابعة', icon: '🕐', cls: 'border-amber-100 bg-amber-50/70 text-amber-700', n: count('FOLLOW') },
    { s: 'LATE', label: 'متأخرة', hint: 'فات موعدها', icon: '⚠️', cls: 'border-red-100 bg-red-50/70 text-red-700', n: count('LATE') },
    { s: 'PROGRESS', label: 'قيد التنفيذ', hint: 'ديشتغلون عليها', icon: '⚙️', cls: 'border-sky-100 bg-sky-50/70 text-sky-700', n: count('PROGRESS') },
    { s: 'DONE', label: 'خالصة', hint: 'هالأسبوع', icon: '✅', cls: 'border-emerald-100 bg-emerald-50/70 text-emerald-700', n: count('DONE') },
  ]
  const sel = 'rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm'

  return (
    <div className="space-y-4" dir="rtl">
      <div>
        <h2 className="text-2xl font-extrabold text-[#0f2040]">⚡ الإجراءات</h2>
        <p className="text-sm text-slate-500">كل شي ينتظر متابعة أو قرار — بمكان واحد، وكل صف يودّيك لشاشته.</p>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        {cards.map((c) => (
          <button key={c.label} onClick={() => { setStatus(c.s); setPage(1) }}
            className={`flex items-center justify-between rounded-2xl border p-4 text-right transition hover:shadow-md ${c.cls} ${status === c.s ? 'ring-2 ring-brand-400' : ''}`}>
            <div>
              <p className="text-sm font-bold">{c.label}</p>
              <p className="text-3xl font-extrabold tabular-nums">{rows ? c.n : '…'}</p>
              <p className="text-[11px] opacity-70">{c.hint}</p>
            </div>
            <span className="grid h-12 w-12 place-items-center rounded-2xl bg-white/80 text-xl">{c.icon}</span>
          </button>
        ))}
      </div>

      <section className="flex flex-wrap items-end gap-2 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <input value={q} onChange={(e) => { setQ(e.target.value); setPage(1) }} placeholder="🔍 دوّر بالإجراءات…" className={`${sel} min-w-[200px] flex-1`} />
        <label className="text-xs font-bold text-slate-500">القسم<br />
          <select value={dept} onChange={(e) => { setDept(e.target.value); setPage(1) }} className={sel}><option value="">الكل</option>{opts('dept').map((o) => <option key={o}>{o}</option>)}</select>
        </label>
        <label className="text-xs font-bold text-slate-500">نوع الإجراء<br />
          <select value={type} onChange={(e) => { setType(e.target.value); setPage(1) }} className={sel}><option value="">الكل</option>{opts('type').map((o) => <option key={o}>{o}</option>)}</select>
        </label>
        <label className="text-xs font-bold text-slate-500">الحالة<br />
          <select value={status} onChange={(e) => { setStatus(e.target.value as Status | ''); setPage(1) }} className={sel}>
            <option value="">الكل</option>{(Object.keys(STATUS) as Status[]).map((s) => <option key={s} value={s}>{STATUS[s].label}</option>)}
          </select>
        </label>
        <button onClick={reset} className="rounded-xl bg-slate-100 px-4 py-2 text-sm font-bold text-slate-700 hover:bg-slate-200">↺ صفّر</button>
      </section>

      <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-3 flex items-center justify-between">
          <h3 className="font-extrabold text-[#0f2040]">☰ قائمة الإجراءات</h3>
          <span className="text-xs text-slate-500">عدد النتائج: {filtered.length}</span>
        </div>
        {rows === null ? <p className="text-slate-400">ديجمع الإجراءات…</p> : filtered.length === 0 ? (
          <p className="py-8 text-center text-slate-500">ماكو إجراءات بهالفلتر ✅</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full min-w-[820px] text-sm">
              <thead><tr className="bg-slate-50 text-right text-xs text-slate-500">
                <th className="rounded-r-lg p-2">#</th><th className="p-2">نوع الإجراء</th><th className="p-2">الوصف</th><th className="p-2">الموظف</th>
                <th className="p-2">القسم</th><th className="p-2">الموعد</th><th className="p-2">الحالة</th><th className="rounded-l-lg p-2">إجراء</th>
              </tr></thead>
              <tbody>
                {shown.map((r, i) => {
                  const st = STATUS[r.status]
                  return (
                    <tr key={r.key} className="border-b border-slate-100">
                      <td className="p-2 text-slate-400">{(cur - 1) * PAGE + i + 1}</td>
                      <td className="whitespace-nowrap p-2 font-bold text-slate-700">{r.icon} {r.type}</td>
                      <td className="p-2 text-slate-600">{r.desc}</td>
                      <td className="whitespace-nowrap p-2 text-slate-700">👤 {r.who}</td>
                      <td className="whitespace-nowrap p-2 text-slate-500">{r.dept}</td>
                      <td className={`whitespace-nowrap p-2 tabular-nums ${r.status === 'LATE' ? 'font-bold text-red-600' : 'text-slate-600'}`}>📅 {fmtDate(r.due)}</td>
                      <td className="p-2"><span className={`whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-bold ${st.cls}`}>{st.icon} {st.label}</span></td>
                      <td className="p-2">
                        <Link to={r.to} className={`whitespace-nowrap rounded-lg px-3 py-1.5 text-xs font-bold ${r.status === 'DONE' ? 'bg-slate-100 text-slate-700' : 'bg-brand-600 text-white hover:bg-brand-700'}`}>
                          {r.status === 'DONE' ? '👁' : '▶'} {r.act}
                        </Link>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
        {pages > 1 && (
          <div className="mt-3 flex items-center justify-center gap-1">
            <button disabled={cur === 1} onClick={() => setPage(1)} className="rounded-lg border px-2.5 py-1 text-xs disabled:opacity-40">الأول</button>
            {Array.from({ length: pages }, (_, i) => i + 1).filter((p) => Math.abs(p - cur) <= 2).map((p) => (
              <button key={p} onClick={() => setPage(p)} className={`rounded-lg px-3 py-1 text-xs font-bold ${p === cur ? 'bg-brand-600 text-white' : 'border'}`}>{p}</button>
            ))}
            <button disabled={cur === pages} onClick={() => setPage(pages)} className="rounded-lg border px-2.5 py-1 text-xs disabled:opacity-40">الأخير</button>
          </div>
        )}
      </section>
    </div>
  )
}
