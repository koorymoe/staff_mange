import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type Employee, type RemoteEntry, type RemoteSummary } from '../api'
import { useSession } from '../session'
import { loadFailed } from '../netErrors'

// ═══ ساعات العمل من البيت ═══
// المسؤول (صلاحية remote_hours_manage) يختار موظف ويضيفله ساعات/دقائق/ثواني
// ليوم معيّن، والنظام يجمع. بنهاية الشهر ينزّل Excel.

const hms = (sec: number) => `${Math.floor(sec / 3600)}:${String(Math.floor((sec % 3600) / 60)).padStart(2, '0')}:${String(sec % 60).padStart(2, '0')}`
const today = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export default function RemoteHoursPage() {
  const { employee: me } = useSession()
  const isAdmin = me?.role === 'ADMIN' || me?.actualRole === 'OWNER'
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
  const loadSummary = useCallback(() => { api.getRemoteSummary(month).then(setSummary).catch(loadFailed) }, [month])
  const loadEmp = useCallback(() => {
    if (!sel) return
    api.getRemoteHours(month, sel.id).then((r) => { setEntries(r.entries); setTotal(r.totalSeconds) }).catch(loadFailed)
  }, [month, sel])
  useEffect(() => { loadSummary() }, [loadSummary])
  useEffect(() => { loadEmp() }, [loadEmp])

  const matches = useMemo(() => q.trim() ? employees.filter((e) => e.name.includes(q.trim())).slice(0, 8) : [], [q, employees])
  const n = (v: string) => Math.max(0, Math.floor(Number(v) || 0))

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

  const days = new Set(entries.map((e) => e.workDate.slice(0, 10))).size
  const inp = 'rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
  return (
    <div className="space-y-4" dir="rtl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-extrabold text-[#0f2040]">🏠 ساعات العمل من البيت</h1>
          <p className="text-sm text-slate-500">ضيف الساعات يومياً، والنظام يجمعها بالساعة والدقيقة والثانية.</p>
        </div>
        <div className="flex items-center gap-2">
          <input type="month" value={month} onChange={(e) => setMonth(e.target.value)} className={inp} />
          <button onClick={() => api.exportRemoteHours(month).catch(loadFailed)} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-sm font-bold text-white">⬇️ Excel كل الموظفين</button>
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_1fr]">
        <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
          <h2 className="mb-2 font-extrabold text-[#0f2040]">➕ إضافة ساعات</h2>
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

        <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
          {!sel ? <p className="text-sm text-slate-400">اختار موظف حتى تشوف عدّاده.</p> : (
            <>
              <div className="flex items-start justify-between gap-2">
                <div>
                  <h2 className="font-extrabold text-[#0f2040]">⏱️ عدّاد {sel.name} — {month}</h2>
                  <p className="font-mono text-4xl font-black tabular-nums text-sky-700" dir="ltr">{hms(total)}</p>
                  <p className="text-xs text-slate-500">ساعات:دقائق:ثواني · {days} يوم · {entries.length} تسجيل</p>
                </div>
                <button onClick={() => api.exportRemoteHours(month, sel.id, sel.name).catch(loadFailed)} className="rounded-lg border border-emerald-600 px-3 py-1.5 text-xs font-bold text-emerald-700">⬇️ Excel هالموظف</button>
              </div>
              <table className="mt-3 w-full text-sm">
                <thead className="text-xs text-slate-500"><tr><th className="text-right">اليوم</th><th>المدة</th><th className="text-right">ملاحظة</th><th className="text-right">سجّلها</th><th /></tr></thead>
                <tbody>
                  {entries.map((e) => (
                    <tr key={e.id} className="border-t border-slate-100">
                      <td className="py-1">{e.workDate.slice(0, 10)}</td>
                      <td className="text-center font-mono" dir="ltr">{hms(e.seconds)}</td>
                      <td className="text-xs text-slate-600">{e.note}</td>
                      <td className="text-xs text-slate-400">{e.addedBy}</td>
                      <td>{(isAdmin || e.addedById === me?.id) && <button onClick={() => del(e)} className="text-xs text-red-600">حذف</button>}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}
        </section>
      </div>

      <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="mb-2 font-extrabold text-[#0f2040]">📋 ملخص {month}</h2>
        {summary.length === 0 ? <p className="text-sm text-slate-400">ماكو ساعات مسجّلة هالشهر.</p> : (
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
