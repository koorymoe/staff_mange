import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type Employee, type RemoteEntry, type RemoteSummary, type RemoteTimer } from '../api'
import { useSession } from '../session'
import { loadFailed } from '../netErrors'

// ═══ ساعات العمل من البيت ═══
// صلاحيتان:
//  • remote_hours_self   — الموظف لنفسه: عدّاد حي (ابدأ/وقّف/إنهاء) + إضافة يدوية.
//  • remote_hours_manage — المسؤول: كل الموظفين، والإكسل، والتصفير (يوم ٢٧ وطالع).
// التصفير ما يحذف: يسكّر الفترة، فالعدّادات تبدي من صفر والسجلات تبقى بالإكسل.

const RESET_DAY = 27
const hms = (sec: number) => `${Math.floor(sec / 3600)}:${String(Math.floor((sec % 3600) / 60)).padStart(2, '0')}:${String(sec % 60).padStart(2, '0')}`
const today = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
const baghdadDay = () => Number(new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad', day: '2-digit' }))
const n = (v: string) => Math.max(0, Math.floor(Number(v) || 0))
const inp = 'rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
const card = 'rounded-2xl border border-slate-200 bg-white p-4 shadow-sm'

export default function RemoteHoursPage() {
  const { employee: me, permissions } = useSession()
  const isAdmin = me?.role === 'ADMIN' || me?.actualRole === 'OWNER'
  const canManage = isAdmin || permissions.includes('remote_hours_manage')
  const canSelf = isAdmin || permissions.includes('remote_hours_self') || permissions.includes('remote_hours_manage')
  return (
    <div className="space-y-4" dir="rtl">
      <div>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🏠 ساعات العمل من البيت</h1>
        <p className="text-sm text-slate-500">شغّل العدّاد أو ضيف الساعات بيدك، والنظام يجمعها بالساعة والدقيقة والثانية.</p>
      </div>
      {canSelf && <MyHours meId={me?.id ?? ''} />}
      {canManage && <ManageHours meId={me?.id ?? ''} isAdmin={isAdmin} />}
    </div>
  )
}

// ═══ ساعاتي: العدّاد الحي + الإضافة اليدوية ═══
function MyHours({ meId }: { meId: string }) {
  const [timer, setTimer] = useState<RemoteTimer | null>(null)
  const [base, setBase] = useState(0) // الثانية عند آخر جواب من الخادم
  const [tick, setTick] = useState(0)
  const [entries, setEntries] = useState<RemoteEntry[]>([])
  const [total, setTotal] = useState(0)
  const [busy, setBusy] = useState(false)
  const [form, setForm] = useState({ date: today(), h: '', m: '', s: '', note: '' })
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)

  const apply = (t: RemoteTimer) => { setTimer(t); setBase(Date.now()); setTick(0) }
  const load = useCallback(() => {
    api.getMyRemote().then((r) => { apply(r.timer); setEntries(r.entries); setTotal(r.totalSeconds) }).catch(loadFailed)
  }, [])
  useEffect(() => { load() }, [load])
  // العدّاد يمشي بالشاشة كل ثانية — والحساب الحقيقي بالخادم.
  useEffect(() => {
    if (!timer?.running) return
    const id = window.setInterval(() => setTick(Math.floor((Date.now() - base) / 1000)), 1000)
    return () => window.clearInterval(id)
  }, [timer?.running, base])
  const shown = (timer?.elapsed ?? 0) + (timer?.running ? tick : 0)

  const act = async (a: 'start' | 'pause' | 'finish') => {
    let note: string | undefined
    if (a === 'finish') {
      const v = prompt(`تنهي العدّاد وتنضاف ${hms(shown)} لساعاتك؟\nاكتب شنو اشتغلت (اختياري):`, '')
      if (v === null) return
      note = v.trim() || undefined
    }
    setBusy(true); setMsg(null)
    try {
      const r = await api.remoteTimer(a, note)
      apply(r.timer)
      if (a === 'finish') { setMsg({ ok: true, t: `انضافت ${hms(r.saved ?? 0)} لساعاتك ✅` }); load() }
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذّر' }) } finally { setBusy(false) }
  }
  const addManual = async () => {
    const b = { date: form.date, hours: n(form.h), minutes: n(form.m), seconds: n(form.s), note: form.note.trim() || undefined }
    if (b.hours + b.minutes + b.seconds === 0) { setMsg({ ok: false, t: 'اكتب المدة' }); return }
    setBusy(true); setMsg(null)
    try {
      const r = await api.addMyRemote(b)
      setMsg({ ok: true, t: `انضافت ${hms(r.seconds)} بيوم ${form.date}` })
      setForm({ ...form, h: '', m: '', s: '', note: '' }); load()
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذّر الحفظ' }) } finally { setBusy(false) }
  }
  const del = async (e: RemoteEntry) => {
    if (!confirm(`تحذف ${hms(e.seconds)} من يوم ${e.workDate.slice(0, 10)}؟`)) return
    try { await api.deleteRemoteHours(e.id); load() } catch (err) { alert(err instanceof Error ? err.message : 'تعذّر الحذف') }
  }

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <section className={card}>
        <h2 className="mb-2 font-extrabold text-[#0f2040]">⏱️ عدّادي</h2>
        <div className="rounded-2xl bg-slate-900 p-5 text-center">
          <p className={`font-mono text-5xl font-black tabular-nums ${timer?.running ? 'text-emerald-400' : 'text-sky-300'}`} dir="ltr">{hms(shown)}</p>
          <p className="mt-1 text-xs text-slate-400">{timer?.running ? '🟢 العدّاد شغّال' : shown > 0 ? '⏸️ واگف — كمّل أو أنهِ' : 'ابدأ لمن تبدي شغل'}</p>
        </div>
        <div className="mt-3 grid grid-cols-3 gap-2">
          {timer?.running
            ? <button disabled={busy} onClick={() => act('pause')} className="rounded-xl bg-amber-500 py-2.5 font-extrabold text-white disabled:opacity-50">⏸️ توقّف</button>
            : <button disabled={busy} onClick={() => act('start')} className="rounded-xl bg-emerald-600 py-2.5 font-extrabold text-white disabled:opacity-50">▶️ {shown > 0 ? 'كمّل' : 'ابدأ'}</button>}
          <button disabled={busy || shown === 0} onClick={() => act('finish')} className="col-span-2 rounded-xl bg-brand-600 py-2.5 font-extrabold text-white disabled:opacity-40">✅ إنهاء وحفظ بالجدول</button>
        </div>
        <p className="mt-2 text-[11px] text-slate-400">«إنهاء» يضيف الوقت لجدول ساعاتك ويصفّر العدّاد. العدّاد محفوظ بالسيرفر — لو سدّيت الجهاز ما يضيع.</p>

        <h3 className="mb-1 mt-4 text-sm font-extrabold text-[#0f2040]">✍️ أو ضيف بيدك</h3>
        <div className="flex flex-wrap items-end gap-2">
          <label className="text-xs text-slate-600">اليوم<input type="date" max={today()} value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} className={`${inp} block`} /></label>
          {([['h', 'ساعات'], ['m', 'دقائق'], ['s', 'ثواني']] as const).map(([k, l]) => (
            <label key={k} className="text-xs text-slate-600">{l}<input type="number" min={0} max={k === 'h' ? 24 : 59} value={form[k]} onChange={(e) => setForm({ ...form, [k]: e.target.value })} className={`${inp} block w-16 text-center`} /></label>
          ))}
        </div>
        <input value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} placeholder="ملاحظة (اختياري) — شنو اشتغلت" className={`${inp} mt-2 w-full`} />
        <button disabled={busy} onClick={addManual} className="mt-2 rounded-lg bg-slate-800 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">➕ أضف</button>
        {msg && <p className={`mt-2 text-sm ${msg.ok ? 'text-emerald-700' : 'text-red-600'}`}>{msg.t}</p>}
      </section>

      <section className={card}>
        <h2 className="font-extrabold text-[#0f2040]">📒 ساعاتي (من آخر تصفير)</h2>
        <p className="font-mono text-3xl font-black tabular-nums text-sky-700" dir="ltr">{hms(total)}</p>
        <EntriesTable entries={entries} canDelete={(e) => e.addedById === meId} onDelete={del} showWho={false} />
      </section>
    </div>
  )
}

function EntriesTable({ entries, canDelete, onDelete, showWho }: { entries: RemoteEntry[]; canDelete: (e: RemoteEntry) => boolean; onDelete: (e: RemoteEntry) => void; showWho: boolean }) {
  if (entries.length === 0) return <p className="mt-2 text-sm text-slate-400">ماكو ساعات بعد.</p>
  return (
    <table className="mt-2 w-full text-sm">
      <thead className="text-xs text-slate-500"><tr><th className="text-right">اليوم</th><th>المدة</th><th className="text-right">ملاحظة</th>{showWho && <th className="text-right">سجّلها</th>}<th /></tr></thead>
      <tbody>
        {entries.map((e) => (
          <tr key={e.id} className="border-t border-slate-100">
            <td className="py-1">{e.workDate.slice(0, 10)} {e.source === 'TIMER' && <span title="من العدّاد">⏱️</span>}</td>
            <td className="text-center font-mono" dir="ltr">{hms(e.seconds)}</td>
            <td className="text-xs text-slate-600">{e.note}</td>
            {showWho && <td className="text-xs text-slate-400">{e.addedBy}</td>}
            <td>{canDelete(e) && <button onClick={() => onDelete(e)} className="text-xs text-red-600">حذف</button>}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

// ═══ المسؤول: كل الموظفين ═══
function ManageHours({ meId, isAdmin }: { meId: string; isAdmin: boolean }) {
  const [month, setMonth] = useState(today().slice(0, 7))
  const [employees, setEmployees] = useState<Employee[]>([])
  const [q, setQ] = useState('')
  const [sel, setSel] = useState<Employee | null>(null)
  const [entries, setEntries] = useState<RemoteEntry[]>([])
  const [total, setTotal] = useState(0)
  const [summary, setSummary] = useState<RemoteSummary[]>([])
  const [form, setForm] = useState({ date: today(), h: '', m: '', s: '', note: '' })
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => { api.getEmployees().then((l) => setEmployees(l.filter((e) => e.status === 'ACTIVE'))).catch(loadFailed) }, [])
  const loadSummary = useCallback(() => { api.getRemoteSummary().then(setSummary).catch(loadFailed) }, [])
  const loadEmp = useCallback(() => {
    if (!sel) return
    api.getRemoteHours(sel.id).then((r) => { setEntries(r.entries); setTotal(r.totalSeconds) }).catch(loadFailed)
  }, [sel])
  useEffect(() => { loadSummary() }, [loadSummary])
  useEffect(() => { loadEmp() }, [loadEmp])

  const matches = useMemo(() => q.trim() ? employees.filter((e) => e.name.includes(q.trim())).slice(0, 8) : [], [q, employees])

  const add = async () => {
    if (!sel) return
    const body = { employeeId: sel.id, date: form.date, hours: n(form.h), minutes: n(form.m), seconds: n(form.s), note: form.note.trim() || undefined }
    if (body.hours + body.minutes + body.seconds === 0) { setMsg({ ok: false, t: 'اكتب المدة' }); return }
    setBusy(true); setMsg(null)
    try {
      const r = await api.addRemoteHours(body)
      setMsg({ ok: true, t: `انضاف ${hms(r.seconds)} لـ${sel.name} بيوم ${form.date}` })
      setForm({ ...form, h: '', m: '', s: '', note: '' })
      loadEmp(); loadSummary()
    } catch (e) { setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذّر الحفظ' }) } finally { setBusy(false) }
  }
  const del = async (e: RemoteEntry) => {
    if (!confirm(`تحذف ${hms(e.seconds)} من يوم ${e.workDate.slice(0, 10)}؟`)) return
    try { await api.deleteRemoteHours(e.id); loadEmp(); loadSummary() } catch (err) { alert(err instanceof Error ? err.message : 'تعذّر الحذف') }
  }
  const reset = async () => {
    const all = summary.reduce((s, r) => s + r.seconds, 0)
    if (!confirm(`تصفّر ساعات كل الموظفين (${summary.length} موظف، ${hms(all)})؟\n\nنزّل الإكسل أول. التصفير ما يحذف — السجلات تبقى بإكسل الشهر، بس العدّادات تبدي من صفر.`)) return
    if (!confirm('متأكد؟ هالخطوة ما ترجع.')) return
    try { const r = await api.resetRemoteHours(); alert(`تصفّرت — ${r.closed} تسجيل انقفل.`); setSel(null); setEntries([]); setTotal(0); loadSummary() } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر التصفير') }
  }
  const showReset = baghdadDay() >= RESET_DAY

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-slate-200 bg-slate-50 p-3">
        <h2 className="font-extrabold text-[#0f2040]">👥 كل الموظفين (مسؤول)</h2>
        <div className="flex flex-wrap items-center gap-2">
          <input type="month" value={month} onChange={(e) => setMonth(e.target.value)} className={inp} title="شهر الإكسل" />
          <button onClick={() => api.exportRemoteHours(month).catch(loadFailed)} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-sm font-bold text-white">⬇️ Excel كل الموظفين</button>
          {showReset && <button onClick={reset} className="rounded-lg bg-red-600 px-3 py-1.5 text-sm font-bold text-white">🔄 تصفير كل الساعات</button>}
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <section className={card}>
          <h2 className="mb-2 font-extrabold text-[#0f2040]">➕ إضافة ساعات لموظف</h2>
          <div className="relative">
            <input value={sel ? sel.name : q} onChange={(e) => { setSel(null); setQ(e.target.value) }} placeholder="🔍 ابحث عن الموظف بالاسم…" className={`${inp} w-full`} />
            {!sel && matches.length > 0 && (
              <ul className="absolute z-10 mt-1 w-full overflow-hidden rounded-lg border border-slate-200 bg-white shadow">
                {matches.map((e) => <li key={e.id}><button onClick={() => { setSel(e); setQ('') }} className="w-full px-3 py-1.5 text-right text-sm hover:bg-sky-50">{e.name} <span className="text-xs text-slate-400">{e.position ?? ''}</span></button></li>)}
              </ul>
            )}
          </div>
          {sel && (
            <div className="mt-3 space-y-2">
              <label className="block text-xs text-slate-600">اليوم <input type="date" max={today()} value={form.date} onChange={(e) => setForm({ ...form, date: e.target.value })} className={`${inp} mr-2`} /></label>
              <div className="flex items-end gap-2">
                {([['h', 'ساعات'], ['m', 'دقائق'], ['s', 'ثواني']] as const).map(([k, l]) => (
                  <label key={k} className="text-xs text-slate-600">{l}
                    <input type="number" min={0} max={k === 'h' ? 24 : 59} value={form[k]} onChange={(e) => setForm({ ...form, [k]: e.target.value })} className={`${inp} block w-20 text-center`} />
                  </label>
                ))}
              </div>
              <input value={form.note} onChange={(e) => setForm({ ...form, note: e.target.value })} placeholder="ملاحظة (اختياري) — شنو اشتغل" className={`${inp} w-full`} />
              <button disabled={busy} onClick={add} className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">➕ أضف</button>
              {msg && <p className={`text-sm ${msg.ok ? 'text-emerald-700' : 'text-red-600'}`}>{msg.t}</p>}
            </div>
          )}
        </section>

        <section className={card}>
          {!sel ? <p className="text-sm text-slate-400">اختار موظف حتى تشوف عدّاده.</p> : (
            <>
              <div className="flex items-start justify-between gap-2">
                <div>
                  <h2 className="font-extrabold text-[#0f2040]">⏱️ عدّاد {sel.name}</h2>
                  <p className="font-mono text-4xl font-black tabular-nums text-sky-700" dir="ltr">{hms(total)}</p>
                  <p className="text-xs text-slate-500">من آخر تصفير · {new Set(entries.map((e) => e.workDate.slice(0, 10))).size} يوم · {entries.length} تسجيل</p>
                </div>
                <button onClick={() => api.exportRemoteHours(month, sel.id, sel.name).catch(loadFailed)} className="rounded-lg border border-emerald-600 px-3 py-1.5 text-xs font-bold text-emerald-700">⬇️ Excel هالموظف</button>
              </div>
              <EntriesTable entries={entries} canDelete={(e) => isAdmin || e.addedById === meId} onDelete={del} showWho />
            </>
          )}
        </section>
      </div>

      <section className={card}>
        <h2 className="mb-2 font-extrabold text-[#0f2040]">📋 ملخص الفترة الحالية (من آخر تصفير)</h2>
        {summary.length === 0 ? <p className="text-sm text-slate-400">ماكو ساعات مسجّلة بالفترة الحالية.</p> : (
          <table className="w-full text-sm">
            <thead className="text-xs text-slate-500"><tr><th className="text-right">الموظف</th><th>الأيام</th><th>المجموع</th></tr></thead>
            <tbody>
              {summary.map((s) => (
                <tr key={s.employeeId} className="cursor-pointer border-t border-slate-100 hover:bg-sky-50" onClick={() => { const e = employees.find((x) => x.id === s.employeeId); if (e) setSel(e) }}>
                  <td className="py-1 font-bold">{s.name}</td><td className="text-center">{s.days}</td><td className="text-center font-mono" dir="ltr">{hms(s.seconds)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  )
}
