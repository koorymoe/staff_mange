import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type AccessGroup, type Employee, type EmployeeAccess, type PermissionAuditRow, type RoleDriftRow } from '../api'
import { roleLabel } from '../roleLabels'

// ═══ 🔐 «الصلاحيات والوصول» — تصوّر ساسكي، بناء لوفي (دازاي 10-26) ═══
// صلاحيات الموظف بمجموعات (وحزمة كاملة بضغطة)، والتعديل مسودة لحد «راجع واحفظ»،
// ومطابقة الصنف، والسجل. الخادم يتحقق من كلشي (نسخة → ٤٠٩ إذا زميل عدّل).

type Tab = 'EMP' | 'DRIFT' | 'LOG'
const SOURCE: Record<string, { t: string; c: string }> = {
  TEMPLATE: { t: 'من قالب منصبه', c: 'bg-sky-50 text-sky-700' },
  MANUAL: { t: 'منح يدوي', c: 'bg-amber-50 text-amber-800' },
  EXCEPTION: { t: 'استثناء معتمد', c: 'bg-violet-50 text-violet-700' },
}
const AGENT_INFO: Record<string, string> = {
  agent_matrix: 'ماتركس — مراقبة وتحليل وتنبيهات',
  agent_quality: 'ابو الجودة — متابعة الزبائن بعد الشغل',
  agent_booking: 'ابو الحجوزات — تثبيت الحجوزات ويا الزبون',
}

export default function AccessPage() {
  const [tab, setTab] = useState<Tab>('EMP')
  const [catalog, setCatalog] = useState<AccessGroup[]>([])
  useEffect(() => { api.getPermissionCatalog().then(setCatalog).catch(() => {}) }, [])
  const label = useMemo(() => Object.fromEntries(catalog.flatMap((g) => g.perms.map((p) => [p.name, p.label]))), [catalog])
  const btn = (t: Tab, s: string) => (
    <button type="button" onClick={() => setTab(t)}
      className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === t ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200'}`}>{s}</button>
  )
  return (
    <div dir="rtl" className="space-y-4 pb-28">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">🔐 الصلاحيات والوصول</h1>
        <p className="mt-1 text-sm text-blue-100">كل صلاحية بمجموعتها، وتعرف ليش عند الموظف، وتطابق كل موظف ويا صنفه.</p>
      </div>
      <div className="flex flex-wrap gap-2">
        {btn('EMP', '👤 صلاحيات الموظف')}{btn('DRIFT', '⚖️ مطابقة الصنف')}{btn('LOG', '📜 السجل')}
        <Link to="/permissions" className="ms-auto self-center text-xs text-slate-400 underline">الشاشة القديمة</Link>
      </div>
      {tab === 'EMP' && <EmployeeTab catalog={catalog} />}
      {tab === 'DRIFT' && <DriftTab label={label} />}
      {tab === 'LOG' && <LogTab label={label} />}
    </div>
  )
}

function EmployeeTab({ catalog }: { catalog: AccessGroup[] }) {
  const [emps, setEmps] = useState<Employee[]>([])
  const [empQ, setEmpQ] = useState('')
  const [acc, setAcc] = useState<EmployeeAccess | null>(null)
  const [draft, setDraft] = useState<Set<string>>(new Set())
  const [q, setQ] = useState('')
  const [open, setOpen] = useState<Record<string, boolean>>({})
  const [review, setReview] = useState(false)
  const [reason, setReason] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [temp, setTemp] = useState<{ perm: string; until: string; reason: string } | null>(null)

  useEffect(() => { api.getEmployees().then(setEmps).catch(() => {}) }, [])
  const load = useCallback((id: string) => {
    setErr('')
    api.getEmployeeAccess(id).then((a) => { setAcc(a); setDraft(new Set(a.perms.map((p) => p.name))) }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
  }, [])

  const have = useMemo(() => new Set(acc?.perms.map((p) => p.name) ?? []), [acc])
  const src = useMemo(() => Object.fromEntries((acc?.perms ?? []).map((p) => [p.name, p])), [acc])
  const add = [...draft].filter((n) => !have.has(n))
  const remove = [...have].filter((n) => !draft.has(n))
  const dirty = add.length + remove.length > 0
  const toggle = (n: string) => setDraft((d) => { const x = new Set(d); if (x.has(n)) x.delete(n); else x.add(n); return x })
  const label = Object.fromEntries(catalog.flatMap((g) => g.perms.map((p) => [p.name, p.label])))

  const save = async () => {
    if (!acc) return
    setBusy(true); setErr('')
    try {
      const a = await api.saveEmployeeAccess(acc.employee.id, { add, remove, reason: reason.trim(), version: acc.version })
      setAcc(a); setDraft(new Set(a.perms.map((p) => p.name))); setReview(false); setReason('')
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const saveTemp = async () => {
    if (!acc || !temp) return
    setBusy(true); setErr('')
    try {
      await api.addPermissionException({ employeeId: acc.employee.id, permission: temp.perm, reason: temp.reason.trim(),
        expiresAt: temp.until ? new Date(temp.until + 'T23:59:00').toISOString() : undefined, version: acc.version })
      setTemp(null); load(acc.employee.id)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  const empList = emps.filter((e) => !empQ || e.name.includes(empQ)).slice(0, 30)
  return (
    <div className="space-y-3">
      <div className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
        <input value={empQ} onChange={(e) => setEmpQ(e.target.value)} placeholder="🔎 اختار الموظف…"
          className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
        {empQ && !acc?.employee.name.includes(empQ) && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {empList.map((e) => (
              <button key={e.id} type="button" onClick={() => { if (dirty && !window.confirm('عندك تغييرات ما محفوظة — تتركها؟')) return; setEmpQ(e.name); load(e.id) }}
                className="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold text-slate-700 hover:bg-slate-200">{e.name}</button>
            ))}
          </div>
        )}
      </div>
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm font-bold text-red-700">{err}</p>}
      {acc && (
        <>
          <div className="flex flex-wrap items-center gap-3 rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <b className="text-lg text-[#0f2040]">{acc.employee.name}</b>
            <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold">{roleLabel(acc.employee.role)}{acc.employee.isLeader ? ' · ليدر' : ''}</span>
            {acc.missingFromTemplate.length > 0 && (
              <span className="text-xs text-amber-700">ناقصه {acc.missingFromTemplate.length} من قالب منصبه</span>
            )}
            <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="دوّر على صلاحية…"
              className="ms-auto w-56 rounded-xl border border-slate-300 px-3 py-1.5 text-sm" />
          </div>
          {catalog.map((g) => {
            const perms = g.perms.filter((p) => !q || p.label.includes(q) || p.name.includes(q))
            if (perms.length === 0) return null
            const on = g.perms.filter((p) => draft.has(p.name)).length
            const grantable = g.perms.filter((p) => p.grantable)
            const full = grantable.every((p) => draft.has(p.name))
            const isOpen = open[g.key] ?? !!q
            return (
              <div key={g.key} className="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-200">
                <div className="flex flex-wrap items-center gap-2 p-3">
                  <button type="button" onClick={() => setOpen((o) => ({ ...o, [g.key]: !isOpen }))} className="flex flex-1 items-center gap-2 text-right">
                    <b className="text-[#0f2040]">{g.title}</b>
                    <span className="text-xs text-slate-500">{on} من {g.perms.length} مفعّلة</span>
                    <span className="text-slate-400">{isOpen ? '▲' : '▼'}</span>
                  </button>
                  {grantable.length > 1 && (
                    full
                      ? <span className="rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700">✔ الحزمة مفعّلة</span>
                      : <button type="button" onClick={() => setDraft((d) => new Set([...d, ...grantable.map((p) => p.name)]))}
                          className="rounded-lg border border-[#0f2040] px-3 py-1 text-xs font-bold text-[#0f2040]">انطِ الحزمة كاملة</button>
                  )}
                </div>
                {isOpen && (
                  <ul className="divide-y divide-slate-100 border-t border-slate-100">
                    {g.key === 'agents' && (
                      <li className="bg-slate-50 px-4 py-2 text-xs text-slate-500">
                        التوجيه ما يمنح الموظف صلاحيات إضافية، ولا يوسّع صلاحيات الوكيل. أسماء الوكلاء يغيّرها المالك من الإعدادات.
                        <span className="mr-2 rounded bg-slate-200 px-2 py-0.5 font-bold text-slate-600">🗂️ ايريس — 🔒 للمدير وحده</span>
                      </li>
                    )}
                    {perms.map((p) => {
                      const s = src[p.name]
                      const changed = draft.has(p.name) !== have.has(p.name)
                      return (
                        <li key={p.name} className={`flex flex-wrap items-center gap-2 px-4 py-2.5 ${changed ? 'bg-amber-50' : ''} ${!p.grantable ? 'opacity-60' : ''}`}>
                          <div className="min-w-0 flex-1">
                            <p className="text-sm font-bold text-slate-800">{p.label}</p>
                            <p className="text-[11px] text-slate-400">
                              {AGENT_INFO[p.name] ?? p.name}
                              {s && <span className={`mr-2 rounded px-1.5 py-0.5 font-bold ${SOURCE[s.source].c}`}>{SOURCE[s.source].t}{s.expiresAt ? ` · لحد ${s.expiresAt.slice(0, 10)}` : ''}</span>}
                              {!p.grantable && <span className="mr-2 text-red-600">🔒 {p.lockReason}</span>}
                            </p>
                          </div>
                          {p.grantable && !have.has(p.name) && (
                            <button type="button" onClick={() => setTemp({ perm: p.name, until: '', reason: '' })} className="text-[11px] text-violet-700 underline">مؤقتة</button>
                          )}
                          <button type="button" disabled={!p.grantable} onClick={() => toggle(p.name)} aria-pressed={draft.has(p.name)}
                            className={`relative h-6 w-11 rounded-full transition ${draft.has(p.name) ? 'bg-emerald-500' : 'bg-slate-300'} disabled:cursor-not-allowed`}>
                            <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all ${draft.has(p.name) ? 'right-0.5' : 'right-[22px]'}`} />
                          </button>
                        </li>
                      )
                    })}
                  </ul>
                )}
              </div>
            )
          })}
        </>
      )}

      {dirty && (
        <div className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200 bg-white/95 p-3 shadow-lg backdrop-blur">
          <div className="mx-auto flex max-w-4xl flex-wrap items-center gap-2" dir="rtl">
            <b className="text-sm text-[#0f2040]">{add.length} إضافة · {remove.length} إزالة</b>
            <button type="button" onClick={() => setReview(true)} className="ms-auto rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white">راجع واحفظ</button>
            <button type="button" onClick={() => acc && setDraft(new Set(acc.perms.map((p) => p.name)))} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
          </div>
        </div>
      )}

      {review && (
        <Modal onClose={() => setReview(false)}>
          <h3 className="text-lg font-bold text-[#0f2040]">راجع قبل الحفظ — {acc?.employee.name}</h3>
          {add.length > 0 && <Diff title="راح تنضاف" tone="emerald" items={add.map((n) => label[n] ?? n)} />}
          {remove.length > 0 && <Diff title="راح تنشال" tone="red" items={remove.map((n) => label[n] ?? n)} />}
          <textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={2} placeholder="السبب (يبقى بالسجل)"
            className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
          {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
          <div className="mt-3 flex gap-2">
            <button type="button" disabled={busy} onClick={() => void save()} className="flex-1 rounded-xl bg-emerald-600 py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : 'احفظ التغييرات'}</button>
            <button type="button" onClick={() => setReview(false)} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">رجوع</button>
          </div>
        </Modal>
      )}

      {temp && (
        <Modal onClose={() => setTemp(null)}>
          <h3 className="text-lg font-bold text-[#0f2040]">⏳ صلاحية مؤقتة: {label[temp.perm] ?? temp.perm}</h3>
          <p className="mt-1 text-xs text-slate-500">تنمنح هسه، وتنشال لحالها بآخر اليوم الي تختاره (مثلاً تغطية إجازة).</p>
          <input type="date" value={temp.until} onChange={(e) => setTemp({ ...temp, until: e.target.value })} className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
          <input value={temp.reason} onChange={(e) => setTemp({ ...temp, reason: e.target.value })} placeholder="السبب" className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
          {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
          <div className="mt-3 flex gap-2">
            <button type="button" disabled={busy || !temp.until || temp.reason.trim().length < 3} onClick={() => void saveTemp()}
              className="flex-1 rounded-xl bg-violet-600 py-2 text-sm font-bold text-white disabled:opacity-50">انطيها مؤقتة</button>
            <button type="button" onClick={() => setTemp(null)} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">رجوع</button>
          </div>
        </Modal>
      )}
    </div>
  )
}

function DriftTab({ label }: { label: Record<string, string> }) {
  const [rows, setRows] = useState<RoleDriftRow[] | null>(null)
  const [openId, setOpenId] = useState('')
  const [confirm, setConfirm] = useState<RoleDriftRow | null>(null)
  const [reason, setReason] = useState('')
  const [err, setErr] = useState('')
  const load = useCallback(() => { api.getRoleDrift().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  useEffect(() => { load() }, [load])
  const L = (n: string) => label[n] ?? n
  const align = async () => {
    if (!confirm) return
    setErr('')
    try { await api.alignRole(confirm.employeeId, reason.trim(), confirm.version); setConfirm(null); setReason(''); load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }
  if (!rows) return <p className="py-10 text-center text-slate-400">{err || 'جاري التحميل…'}</p>
  const n = (s: RoleDriftRow['status']) => rows.filter((r) => r.status === s).length
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-3 gap-2 text-center">
        <div className="rounded-2xl bg-emerald-50 p-3"><b className="text-2xl text-emerald-700">{n('MATCH')}</b><p className="text-xs text-emerald-700">مطابقون</p></div>
        <div className="rounded-2xl bg-amber-50 p-3"><b className="text-2xl text-amber-700">{n('DIFF')}</b><p className="text-xs text-amber-700">عندهم اختلاف</p></div>
        <div className="rounded-2xl bg-slate-100 p-3"><b className="text-2xl text-slate-600">{n('NO_TEMPLATE')}</b><p className="text-xs text-slate-600">بلا قالب</p></div>
      </div>
      {rows.filter((r) => r.status !== 'MATCH').map((r) => (
        <div key={r.employeeId} className="rounded-2xl bg-white p-4 shadow-sm ring-1 ring-slate-200">
          <div className="flex flex-wrap items-center gap-2">
            <b className="text-[#0f2040]">{r.name}</b><span className="text-xs text-slate-500">· {roleLabel(r.role)}</span>
            {r.status === 'NO_TEMPLATE'
              ? <span className="text-xs font-bold text-slate-500">يحتاج تحديد صنفه</span>
              : <span className="text-xs text-amber-700">{r.extra.length} زايدة · {r.missing.length} ناقصة</span>}
            <span className="ms-auto flex gap-2">
              <button type="button" onClick={() => setOpenId(openId === r.employeeId ? '' : r.employeeId)} className="rounded-lg border border-slate-300 px-3 py-1 text-xs font-bold text-slate-600">شوف الفرق</button>
              <button type="button" disabled={r.status === 'NO_TEMPLATE'} onClick={() => setConfirm(r)}
                className="rounded-lg bg-[#0f2040] px-3 py-1 text-xs font-bold text-white disabled:opacity-40">طابقه ويا صنفه</button>
            </span>
          </div>
          {openId === r.employeeId && (
            <div className="mt-3 grid gap-2 sm:grid-cols-3">
              <Diff title="راح تنضاف" tone="emerald" items={r.missing.map(L)} />
              <Diff title="راح تنشال" tone="red" items={r.extra.map(L)} />
              <Diff title="استثناءات معتمدة (تبقى)" tone="violet" items={r.exceptions.map(L)} />
            </div>
          )}
        </div>
      ))}
      {confirm && (
        <Modal onClose={() => setConfirm(null)}>
          <h3 className="text-lg font-bold text-[#0f2040]">مطابقة {confirm.name} ويا صنفه</h3>
          <p className="mt-2 text-sm">راح تنضاف <b>{confirm.missing.length}</b> وتنشال <b>{confirm.extra.length}</b>. الاستثناءات المعتمدة والصلاحيات المقفولة ما تنلمس. متأكد؟</p>
          <textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={2} placeholder="سبب التغيير"
            className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm" />
          {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
          <div className="mt-3 flex gap-2">
            <button type="button" disabled={reason.trim().length < 3} onClick={() => void align()} className="flex-1 rounded-xl bg-emerald-600 py-2 text-sm font-bold text-white disabled:opacity-50">تأكيد المطابقة</button>
            <button type="button" onClick={() => setConfirm(null)} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">رجوع</button>
          </div>
        </Modal>
      )}
    </div>
  )
}

function LogTab({ label }: { label: Record<string, string> }) {
  const [rows, setRows] = useState<PermissionAuditRow[] | null>(null)
  useEffect(() => { api.getPermissionAuditLog().then(setRows).catch(() => setRows([])) }, [])
  if (!rows) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="rounded-2xl bg-white py-10 text-center text-slate-500 ring-1 ring-slate-200">ماكو تغييرات مسجّلة بعد.</p>
  const A = { ADD: ['➕', 'text-emerald-700'], REMOVE: ['➖', 'text-red-700'], EXPIRE: ['⏳', 'text-violet-700'] } as const
  return (
    <ul className="space-y-1.5">
      {rows.map((r) => (
        <li key={r.id} className="flex flex-wrap items-center gap-2 rounded-xl bg-white px-4 py-2 text-sm shadow-sm ring-1 ring-slate-200">
          <span className={A[r.action][1]}>{A[r.action][0]}</span>
          <b>{r.employeeName ?? '—'}</b>
          <span className="text-slate-700">{label[r.permission] ?? r.permission}</span>
          {r.reason && <span className="text-xs text-slate-500">— {r.reason}</span>}
          <span className="ms-auto text-xs text-slate-400">{r.actorName ?? 'النظام'} · {new Date(r.createdAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
        </li>
      ))}
    </ul>
  )
}

function Diff({ title, items, tone }: { title: string; items: string[]; tone: 'emerald' | 'red' | 'violet' }) {
  const c = { emerald: 'bg-emerald-50 text-emerald-800', red: 'bg-red-50 text-red-800', violet: 'bg-violet-50 text-violet-800' }[tone]
  return (
    <div className={`mt-2 rounded-xl p-3 text-sm ${c}`}>
      <p className="font-bold">{title} ({items.length})</p>
      {items.length === 0 ? <p className="text-xs opacity-60">—</p> : <ul className="mt-1 list-inside list-disc text-xs">{items.map((x) => <li key={x}>{x}</li>)}</ul>}
    </div>
  )
}

function Modal({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl" onClick={onClose}>
      <div className="max-h-[85vh] w-full max-w-md overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>{children}</div>
    </div>
  )
}
