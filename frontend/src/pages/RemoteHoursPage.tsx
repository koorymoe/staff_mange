import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type Employee, type RemoteEntry, type RemoteSummary, type RemoteTimer, type RemotePlan, type RemoteHistory } from '../api'
import { useSession, isAdminOrOwner } from '../session'
import { loadFailed } from '../netErrors'
import { hms, todayBaghdad as today } from '../utils/format'
import { card } from '../ui/classes'
import { askConfirm } from '../ui/dialogs'
import PctWhy from '../ui/PctWhy'

// ═══ ساعات العمل من البيت ═══
// صلاحيتان:
//  • remote_hours_self   — الموظف لنفسه: عدّاد حي (ابدأ/وقّف/إنهاء) + إضافة يدوية.
//  • remote_hours_manage — المسؤول: كل الموظفين والإكسل. التصفير للمالك بس، بأي وقت.
// التصفير ما يحذف: يسكّر الفترة، فالعدّادات تبدي من صفر والسجلات تبقى بالإكسل.

const n = (v: string) => Math.max(0, Math.floor(Number(v) || 0))
const inp = 'rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
export default function RemoteHoursPage() {
  const { employee: me, permissions } = useSession()
  const isAdmin = isAdminOrOwner(me)
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
      {(me?.role === 'OWNER' || me?.actualRole === 'OWNER') && <OwnerHistory />}
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
  const [plan, setPlan] = useState<RemotePlan | null>(null)
  const [planForm, setPlanForm] = useState({ hours: '', rate: '' })
  const [planMsg, setPlanMsg] = useState('')

  const apply = (t: RemoteTimer) => { setTimer(t); setBase(Date.now()); setTick(0) }
  const load = useCallback(() => {
    api.getMyRemote().then((r) => { apply(r.timer); setEntries(r.entries); setTotal(r.totalSeconds); setPlan(r.plan ?? null) }).catch(loadFailed)
  }, [])
  useEffect(() => { load() }, [load])
  // العدّاد يمشي بالشاشة كل ثانية — والحساب الحقيقي بالخادم.
  useEffect(() => {
    if (!timer?.running) return
    const id = window.setInterval(() => setTick(Math.floor((Date.now() - base) / 1000)), 1000)
    return () => window.clearInterval(id)
  }, [timer?.running, base])
  const shown = (timer?.elapsed ?? 0) + (timer?.running ? tick : 0)

  // (ع) 10-16: «الإنهاء وحفظ بالجدول جاي تعلك» — جان يسأل بـprompt() المتصفح،
  // والمتصفح (أو التطبيق المنصّب) أحياناً يمنع النوافذ فالزر ما يسوي شي.
  // هسه التأكيد والملاحظة جوّا الصفحة.
  const [confirmFinish, setConfirmFinish] = useState(false)
  const [finishNote, setFinishNote] = useState('')
  const act = async (a: 'start' | 'pause' | 'finish') => {
    const note = a === 'finish' ? finishNote.trim() || undefined : undefined
    setBusy(true); setMsg(null)
    try {
      const r = await api.remoteTimer(a, note)
      apply(r.timer)
      if (a === 'finish') { setMsg({ ok: true, t: `انضافت ${hms(r.saved ?? 0)} لساعاتك ✅` }); setConfirmFinish(false); setFinishNote(''); load() }
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
    if (!(await askConfirm(`تحذف ${hms(e.seconds)} من يوم ${e.workDate.slice(0, 10)}؟`))) return
    try { await api.deleteRemoteHours(e.id); load() } catch (err) { alert(err instanceof Error ? err.message : 'تعذّر الحذف') }
  }

  // 🎯 (دازاي 11-01) المطلوب بالشهر: عدّاد تنازلي، وبعد ما يعبر وقت إضافي — والراتب = الساعات × السعر
  const worked = total + shown
  const target = (plan?.targetHours ?? 0) * 3600
  const left = target - worked
  const rateIQD = plan?.hourlyRateIQD ?? 0
  const pay = (sec: number) => (sec / 3600) * rateIQD
  // (دازاي 11-01) فلوس بالفواصل والكسور ويا كلمة «دينار» — حتى ما يصير «62؟ الف؟»
  const money = (n: number, digits = 2) => `${n.toLocaleString('en-US', { minimumFractionDigits: digits, maximumFractionDigits: digits })} دينار`
  // «2.715» بالعراقي = ألفين وسبعمية وخمسطعش (النقطة فاصل آلاف)، مو ٢ وكسور
  const parseRate = (v: string) => {
    const t = v.trim().replace(/[,،\s]/g, '')
    return Number(/^\d{1,3}(\.\d{3})+$/.test(t) ? t.replace(/\./g, '') : t.replace(/[^\d.]/g, ''))
  }
  const savePlan = async () => {
    const hours = Number(planForm.hours), rate = parseRate(planForm.rate)
    if (!(hours > 0 && hours <= 400) || !(rate >= 0)) { setPlanMsg('اكتب الساعات المطلوبة بالشهر وسعر الساعة'); return }
    // (ع) 11-08: سعر الساعة بالدينار كامل — 2.7 يعني 2700
    if (rate > 0 && rate < 100) { setPlanMsg(`اكتب سعر الساعة بالدينار كامل — مثلاً ${(rate * 1000).toLocaleString('en-US')} مو ${rate}`); return }
    try { const r = await api.setRemotePlan({ targetHours: hours, hourlyRateIQD: rate }); setPlan(r.plan); setPlanMsg('') }
    catch (e) { setPlanMsg(e instanceof Error ? e.message : 'تعذر الحفظ') }
  }
  const planCard = (
    <section className={`${card} lg:col-span-2`}>
      <h2 className="mb-2 font-extrabold text-[#0f2040]">🎯 المطلوب مني بالشهر وراتبي</h2>
      {!plan ? (
        <div className="flex flex-wrap items-end gap-2">
          <label className="text-xs text-slate-600">الساعات المطلوبة بالشهر<input type="number" min={1} max={400} value={planForm.hours} onChange={(e) => setPlanForm({ ...planForm, hours: e.target.value })} className={`${inp} block w-28 text-center`} /></label>
          <label className="text-xs text-slate-600">سعر الساعة (د.ع)<input inputMode="numeric" value={planForm.rate} onChange={(e) => setPlanForm({ ...planForm, rate: e.target.value })} placeholder="مثلاً 2700" className={`${inp} block w-32 text-center`} dir="ltr" /></label>
          <button onClick={savePlan} className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white">حفظ</button>
          {planForm.rate && <p className="w-full text-xs font-bold text-brand-700">يعني: {money(parseRate(planForm.rate), 0)} بالساعة</p>}
          <p className="w-full text-[11px] text-slate-500">تحددها مرة وحدة — بعدها التعديل من إدارة الساعات.</p>
          {planMsg && <p className="w-full text-sm font-bold text-red-600">{planMsg}</p>}
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-3">
          <div className={`rounded-2xl p-4 text-center ${left > 0 ? 'bg-slate-900' : 'bg-emerald-900'}`}>
            <p className="text-xs text-slate-300">{left > 0 ? '⏳ باقي عليك' : '⚡ وقت إضافي'}</p>
            <p className={`font-mono text-4xl font-black tabular-nums ${left > 0 ? 'text-amber-300' : 'text-emerald-300'}`} dir="ltr">{hms(Math.abs(left))}</p>
            <p className="mt-1 text-[11px] text-slate-400">من {plan.targetHours} ساعة بالشهر</p>
          </div>
          <div className="rounded-2xl bg-slate-50 p-4 text-center">
            <p className="text-xs text-slate-500">💰 راتبي لهسه</p>
            <p className="text-2xl font-black tabular-nums text-[#0f2040]" dir="ltr">{money(pay(worked))}</p>
            <p className="text-[11px] text-slate-500">{hms(worked)} × {money(rateIQD, 0)} بالساعة</p>
            <p className="mt-1 text-[11px] font-bold text-slate-600">الحد: {money(plan.targetHours * rateIQD, 0)} ({plan.targetHours} ساعة)</p>
          </div>
          <div className="rounded-2xl bg-emerald-50 p-4 text-center">
            <p className="text-xs text-emerald-800">⚡ منها وقت إضافي</p>
            <p className="text-2xl font-black tabular-nums text-emerald-700" dir="ltr">{money(pay(Math.max(0, -left)))}</p>
            <p className="text-[11px] text-emerald-800">{hms(Math.max(0, -left))} فوك المطلوب</p>
          </div>
          <div className="grid grid-cols-3 gap-2 rounded-2xl border border-amber-200 bg-amber-50 p-3 text-center sm:col-span-3">
            <div><p className="text-[11px] text-amber-800">⏱️ كل ساعة</p><p className="font-black tabular-nums text-amber-900" dir="ltr">{money(rateIQD)}</p></div>
            <div><p className="text-[11px] text-amber-800">كل دقيقة</p><p className="font-black tabular-nums text-amber-900" dir="ltr">{money(rateIQD / 60)}</p></div>
            <div><p className="text-[11px] text-amber-800">كل ثانية</p><p className="font-black tabular-nums text-amber-900" dir="ltr">{money(rateIQD / 3600, 4)}</p></div>
          </div>
          {rateIQD > 0 && rateIQD < 100 && (
            <p className="rounded-lg bg-red-50 p-2 text-xs font-bold text-red-700 sm:col-span-3">⚠️ سعر ساعتك مسجّل {money(rateIQD)} بس — إذا تقصد {money(rateIQD * 1000, 0)} گول لإدارة الساعات تعدّله.</p>
          )}
          <p className="text-[11px] text-slate-400 sm:col-span-3">الحساب من آخر تصفية للساعات، ويمشي ويا العدّاد لحظة بلحظة.</p>
        </div>
      )}
    </section>
  )

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      {planCard}
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
          <button disabled={busy || shown === 0} onClick={() => setConfirmFinish(true)} className="col-span-2 rounded-xl bg-brand-600 py-2.5 font-extrabold text-white disabled:opacity-40">✅ إنهاء وحفظ بالجدول</button>
        </div>
        {confirmFinish && (
          <div className="mt-3 rounded-xl border border-brand-200 bg-brand-50 p-3">
            <p className="text-sm font-extrabold text-[#0f2040]">تنهي العدّاد وتنضاف {hms(shown)} لساعاتك؟</p>
            <input value={finishNote} onChange={(e) => setFinishNote(e.target.value)} placeholder="شنو اشتغلت (اختياري)"
              className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm" />
            <div className="mt-2 flex gap-2">
              <button disabled={busy} onClick={() => act('finish')} className="flex-1 rounded-lg bg-brand-600 py-2 text-sm font-extrabold text-white disabled:opacity-50">{busy ? 'جاري الحفظ…' : '✅ إي، احفظ'}</button>
              <button disabled={busy} onClick={() => setConfirmFinish(false)} className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
            </div>
            {msg && !msg.ok && <p className="mt-2 text-sm font-bold text-red-600">{msg.t}</p>}
          </div>
        )}
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
        <EntriesTable entries={entries} canDelete={(e) => e.addedById === meId} onDelete={del} onSaved={load} showWho={false} />
      </section>
    </div>
  )
}

// قرار (ع) 10-08: عرضين — «حسب الجلسات» (كل توقف وبدء) و«حسب الأيام» (مجموع كل يوم).
function EntriesTable({ entries, canDelete, onDelete, onSaved, showWho }: { entries: RemoteEntry[]; canDelete: (e: RemoteEntry) => boolean; onDelete: (e: RemoteEntry) => void; onSaved: () => void; showWho: boolean }) {
  const [view, setView] = useState<'sessions' | 'days'>('sessions')
  // ✏️ تعديل بنفس المكان — بلا window.prompt (محجوب على بعض الأجهزة)
  const [editing, setEditing] = useState<{ id: string; h: string; m: string; note: string } | null>(null)
  const [saving, setSaving] = useState(false)
  const [editErr, setEditErr] = useState('')
  const startEdit = (e: RemoteEntry) => {
    setEditErr('')
    setEditing({ id: e.id, h: String(Math.floor(e.seconds / 3600)), m: String(Math.floor((e.seconds % 3600) / 60)), note: e.note ?? '' })
  }
  const saveEdit = async () => {
    if (!editing) return
    setSaving(true); setEditErr('')
    try {
      await api.updateRemoteHours(editing.id, { hours: Number(editing.h) || 0, minutes: Number(editing.m) || 0, seconds: 0, note: editing.note })
      setEditing(null); onSaved()
    } catch (err) { setEditErr(err instanceof Error ? err.message : 'تعذّر الحفظ') } finally { setSaving(false) }
  }
  const days = useMemo(() => {
    const m = new Map<string, { seconds: number; n: number }>()
    entries.forEach((e) => {
      const d = e.workDate.slice(0, 10)
      const cur = m.get(d) ?? { seconds: 0, n: 0 }
      m.set(d, { seconds: cur.seconds + e.seconds, n: cur.n + 1 })
    })
    return [...m.entries()].sort((a, b) => a[0].localeCompare(b[0]))
  }, [entries])
  if (entries.length === 0) return <p className="mt-2 text-sm text-slate-400">ماكو ساعات بعد.</p>
  const tab = (k: 'sessions' | 'days', label: string) => (
    <button type="button" onClick={() => setView(k)}
      className={`rounded-lg px-3 py-1 text-xs font-bold ${view === k ? 'bg-sky-700 text-white' : 'bg-slate-100 text-slate-600'}`}>{label}</button>
  )
  return (
    <>
    <div className="mt-3 flex gap-2">{tab('sessions', '⏱️ حسب الجلسات')}{tab('days', '📅 حسب الأيام')}</div>
    {view === 'days' ? (
      <table className="mt-2 w-full text-sm">
        <thead className="text-xs text-slate-500"><tr><th className="text-right">اليوم</th><th>مجموع الساعات</th><th className="text-right">عدد الجلسات</th></tr></thead>
        <tbody>
          {days.map(([d, v]) => (
            <tr key={d} className="border-t border-slate-100">
              <td className="py-1">{d}</td>
              <td className="text-center font-mono font-bold" dir="ltr">{hms(v.seconds)}</td>
              <td className="text-xs text-slate-500">{v.n}</td>
            </tr>
          ))}
        </tbody>
      </table>
    ) : (
    // (ع) 10-23: «نستخدم أصابع للضغط» — بطاقة لكل جلسة وأزرار كبيرة (٤٤px) بدل سطر ناعم.
    <ul className="mt-2 space-y-2">
      {entries.map((e) => (
        <li key={e.id} className="rounded-xl border border-slate-200 bg-white p-3">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <span className="text-sm font-bold text-slate-700">{e.workDate.slice(0, 10)} {e.source === 'TIMER' && <span title="من العدّاد">⏱️</span>}</span>
            <span className="font-mono text-lg font-extrabold text-slate-900" dir="ltr">{hms(e.seconds)}</span>
            {showWho && <span className="text-xs text-slate-400">سجّلها: {e.addedBy}</span>}
          </div>
          {e.note && <p className="mt-1 text-sm text-slate-600">{e.note}</p>}
          {editing?.id === e.id ? (
            <div className="mt-3 space-y-2">
              <div className="flex gap-2">
                <label className="flex-1 text-sm font-bold text-slate-600">ساعات
                  <input type="number" inputMode="numeric" min={0} max={24} value={editing.h} onChange={(x) => setEditing({ ...editing, h: x.target.value })}
                    className="mt-1 w-full rounded-lg border border-slate-300 p-2.5 text-base" /></label>
                <label className="flex-1 text-sm font-bold text-slate-600">دقائق
                  <input type="number" inputMode="numeric" min={0} max={59} value={editing.m} onChange={(x) => setEditing({ ...editing, m: x.target.value })}
                    className="mt-1 w-full rounded-lg border border-slate-300 p-2.5 text-base" /></label>
              </div>
              <input value={editing.note} onChange={(x) => setEditing({ ...editing, note: x.target.value })} placeholder="ملاحظة (اختياري)"
                className="w-full rounded-lg border border-slate-300 p-2.5 text-base" />
              {editErr && <p className="text-sm font-bold text-red-600">{editErr}</p>}
              <div className="flex gap-2">
                <button type="button" disabled={saving} onClick={saveEdit} className="min-h-11 flex-1 rounded-xl bg-brand-600 px-4 text-base font-extrabold text-white disabled:opacity-50">{saving ? 'جاري الحفظ…' : '✅ احفظ'}</button>
                <button type="button" disabled={saving} onClick={() => setEditing(null)} className="min-h-11 rounded-xl border border-slate-300 px-5 text-base font-bold text-slate-600">إلغاء</button>
              </div>
            </div>
          ) : canDelete(e) && (
            <div className="mt-3 flex gap-2">
              <button type="button" onClick={() => startEdit(e)} className="min-h-11 flex-1 rounded-xl border-2 border-sky-600 px-4 text-base font-extrabold text-sky-700">✏️ تعديل</button>
              <button type="button" onClick={() => onDelete(e)} className="min-h-11 flex-1 rounded-xl bg-red-600 px-4 text-base font-extrabold text-white">🗑️ حذف</button>
            </div>
          )}
        </li>
      ))}
    </ul>
    )}
    </>
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
  // (دازاي 11-01) إدارة الساعات تعدّل المطلوب وسعر الساعة لأي موظف
  const [empPlan, setEmpPlan] = useState({ hours: '', rate: '' })
  const [planNote, setPlanNote] = useState('')

  useEffect(() => { api.getEmployees().then((l) => setEmployees(l.filter((e) => e.status === 'ACTIVE'))).catch(loadFailed) }, [])
  const loadSummary = useCallback(() => { api.getRemoteSummary().then(setSummary).catch(loadFailed) }, [])
  const loadEmp = useCallback(() => {
    if (!sel) return
    api.getRemoteHours(sel.id).then((r) => { setEntries(r.entries); setTotal(r.totalSeconds); setEmpPlan({ hours: r.plan ? String(r.plan.targetHours) : '', rate: r.plan ? String(r.plan.hourlyRateIQD) : '' }); setPlanNote('') }).catch(loadFailed)
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
    if (!(await askConfirm(`تحذف ${hms(e.seconds)} من يوم ${e.workDate.slice(0, 10)}؟`))) return
    try { await api.deleteRemoteHours(e.id); loadEmp(); loadSummary() } catch (err) { alert(err instanceof Error ? err.message : 'تعذّر الحذف') }
  }
  const reset = async () => {
    const all = summary.reduce((s, r) => s + r.seconds, 0)
    if (!(await askConfirm(`تصفّر ساعات كل الموظفين (${summary.length} موظف، ${hms(all)})؟\n\nنزّل الإكسل أول. التصفير ما يحذف — السجلات تبقى بإكسل الشهر، بس العدّادات تبدي من صفر.`))) return
    if (!(await askConfirm('متأكد؟ هالخطوة ما ترجع.'))) return
    try { const r = await api.resetRemoteHours(); alert(`تصفّرت — ${r.closed} تسجيل انقفل.`); setSel(null); setEntries([]); setTotal(0); loadSummary() } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر التصفير') }
  }
  // (دازاي 10-26) «الزر يطلع فقط للمالك بكل وقت»
  const showReset = useSession().employee?.actualRole === 'OWNER'

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
              <div className="my-3 flex flex-wrap items-end gap-2 rounded-xl border border-amber-200 bg-amber-50 p-3">
                <span className="w-full text-xs font-bold text-amber-900">🎯 المطلوب بالشهر وسعر الساعة</span>
                <label className="text-xs text-slate-600">ساعات<input type="number" min={1} max={400} value={empPlan.hours} onChange={(e) => setEmpPlan({ ...empPlan, hours: e.target.value })} className={`${inp} block w-24 text-center`} /></label>
                <label className="text-xs text-slate-600">سعر الساعة (دينار)<input value={empPlan.rate} onChange={(e) => setEmpPlan({ ...empPlan, rate: e.target.value })} placeholder="مثلاً 2700" className={`${inp} block w-32 text-center`} dir="ltr" /></label>
                <button className="rounded-lg bg-amber-600 px-3 py-2 text-sm font-bold text-white" onClick={async () => {
                  const hours = Number(empPlan.hours), rate = Number(empPlan.rate.replace(/[,،\s]/g, ''))
                  if (!(hours > 0 && hours <= 400) || !(rate >= 0)) { setPlanNote('اكتب الساعات وسعر الساعة'); return }
                  if (rate > 0 && rate < 100) { setPlanNote(`اكتب سعر الساعة بالدينار كامل — مثلاً ${(rate * 1000).toLocaleString('en-US')} مو ${rate}`); return }
                  try { await api.setRemotePlan({ employeeId: sel.id, targetHours: hours, hourlyRateIQD: rate }); setPlanNote(`✅ انحفظ: ${rate.toLocaleString('en-US')} دينار بالساعة`) }
                  catch (e) { setPlanNote(e instanceof Error ? e.message : 'تعذر الحفظ') }
                }}>حفظ</button>
                {planNote && <span className="w-full text-xs font-bold text-slate-700">{planNote}</span>}
              </div>
              <EntriesTable entries={entries} canDelete={(e) => isAdmin || e.addedById === meId} onDelete={del} onSaved={() => { loadEmp(); loadSummary() }} showWho />
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

// ═══ 📅 سجل الأشهر السابقة — المالك بس (دازاي 11-02) ═══
// «اريد اشوف دوام الموظفين للأشهر السابقة شكد جان». الفلوس بسعر الساعة الحالي ولحد المطلوب.
const monthAr = (m: string) => new Date(m + '-15T12:00:00').toLocaleDateString('ar-IQ', { month: 'long', year: 'numeric' })
const iqd = (v: number) => `${Math.round(v).toLocaleString('en-US')} دينار`

function OwnerHistory() {
  const [month, setMonth] = useState('')
  const [emp, setEmp] = useState('')
  const [data, setData] = useState<RemoteHistory | null>(null)
  useEffect(() => { api.getRemoteHistory(month, emp).then(setData).catch(loadFailed) }, [month, emp])
  if (!data) return null
  const cur = data.month ?? ''
  return (
    <div className={`${card} space-y-3`}>
      <h2 className="text-lg font-extrabold text-[#0f2040]">📅 سجل الأشهر السابقة <span className="text-xs font-normal text-slate-400">(إلك بس)</span></h2>
      {data.months.length === 0 ? <p className="text-sm text-slate-400">بعد ماكو دوام من البيت مسجّل.</p> : (
        <div className="flex flex-wrap gap-2">
          {data.months.map((m) => (
            <button key={m.month} onClick={() => { setMonth(m.month); setEmp('') }}
              className={`rounded-xl px-3 py-2 text-right text-xs ${m.month === cur ? 'bg-[#0f2040] text-white' : 'bg-slate-50 text-slate-700 ring-1 ring-slate-200'}`}>
              <b className="block text-sm">{monthAr(m.month)}</b>
              {hms(m.seconds)} · {m.employees} موظف · {m.days} يوم
            </button>
          ))}
        </div>
      )}
      {data.rows && data.rows.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="text-xs text-slate-500"><tr><th className="p-2 text-right">الموظف</th><th>الأيام</th><th>الساعات</th><th>المطلوب</th><th>الإنجاز</th><th>المستحق</th></tr></thead>
            <tbody>
              {data.rows.map((r) => {
                const target = (r.targetHours ?? 0) * 3600
                const pct = target > 0 ? Math.round((r.seconds / target) * 100) : null
                const pay = r.hourlyRateIQD ? (Math.min(r.seconds, target || r.seconds) / 3600) * r.hourlyRateIQD : null
                return (
                  <tr key={r.employeeId} onClick={() => setEmp(emp === r.employeeId ? '' : r.employeeId)} className={`cursor-pointer border-t border-slate-100 hover:bg-slate-50 ${emp === r.employeeId ? 'bg-sky-50' : ''}`}>
                    <td className="p-2 font-bold">{r.name}</td>
                    <td className="p-2 text-center">{r.days}</td>
                    <td className="p-2 text-center font-mono" dir="ltr">{hms(r.seconds)}</td>
                    <td className="p-2 text-center">{r.targetHours ? `${r.targetHours} ساعة` : '—'}</td>
                    <td className={`p-2 text-center font-bold ${pct === null ? '' : pct >= 100 ? 'text-emerald-700' : pct >= 70 ? 'text-amber-700' : 'text-red-700'}`}>{pct === null ? '—' : (
                      <PctWhy title={`إنجاز ساعات ${r.name}`} value={pct} suffix="٪"
                        formula={<>اشتغل {hms(r.seconds)} من البيت بـ{monthAr(cur)}، والمطلوب {r.targetHours} ساعة ({hms(target)}) → {hms(r.seconds)} ÷ {hms(target)} × 100 = <b>{pct}٪</b></>}
                        meaning={<>شكد كمّل من الساعات المطلوبة منه بهذا الشهر. 100٪ وفوق أخضر (كمّل)، من 70٪ أصفر، أقل أحمر. المستحق ينحسب على الساعات لحد المطلوب بس.</>}
                        source={<>جلسات الدوام من البيت المسجّلة للموظف بهذا الشهر (التايمر)، والساعات المطلوبة من خطة الدوام من البيت المكتوبة إله.</>} />
                    )}</td>
                    <td className="p-2 text-center">{pay === null ? '—' : iqd(pay)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          <p className="mt-1 text-[11px] text-slate-400">اضغط على موظف تشوف أيامه. المستحق محسوب بسعر الساعة الحالي ولحد الساعات المطلوبة.</p>
        </div>
      )}
      {emp && data.entries && (
        <div className="rounded-xl bg-slate-50 p-3 text-xs">
          {data.entries.map((e) => (
            <div key={e.id} className="flex gap-3 border-b border-slate-100 py-1 last:border-0">
              <span className="w-24 font-mono text-slate-500" dir="ltr">{e.workDate.slice(0, 10)}</span>
              <span className="w-20 font-mono font-bold" dir="ltr">{hms(e.seconds)}</span>
              <span className="text-slate-500">{e.source === 'TIMER' ? '⏱️ عدّاد' : `✍️ ${e.addedBy}`}{e.note && ` — ${e.note}`}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
