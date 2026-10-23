import { useEffect, useState } from 'react'
import { api, type AttendanceClaim } from '../api'

// ═══ طلبات تصحيح الانصراف — للمراقب ═══
// موظف انسجّل انصرافه تلقائياً وگال «چنت أشتغل لحد كذا» وماتركس ما لگى دليل
// كافي بالنظام — المراقب يقرر.
const t = (s: string | null) => (s ? new Date(s).toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit' }) : '—')

export default function AttendanceClaims() {
  const [rows, setRows] = useState<AttendanceClaim[] | null>(null)
  const load = () => { void api.getAttendanceClaims('PENDING').then(setRows).catch(() => setRows([])) }
  useEffect(load, [])
  if (!rows) return <p className="text-xs text-slate-400">…</p>
  const decide = async (id: string, approve: boolean) => { await api.decideAttendanceClaim(id, approve); load() }
  return (
    <div dir="rtl" className="space-y-2">
      <p className="text-sm font-extrabold text-slate-800">🕓 طلبات تصحيح الانصراف</p>
      <p className="text-[11px] text-slate-500">موظفين انسجّل انصرافهم تلقائي وگالوا «چنت أشتغل بعدها»، وماتركس ما لگى دليل كافي بالنظام. وافق إذا متأكد، أو ارفض.</p>
      {rows.length === 0 && <p className="rounded-xl bg-emerald-50 p-3 text-sm text-emerald-800">✅ ماكو طلبات تنتظر.</p>}
      {rows.map((c) => (
        <div key={c.id} className="rounded-xl border border-amber-200 bg-amber-50/50 p-3 text-sm">
          <p><b>{c.name}</b> · انصراف تلقائي {t(c.autoAt)} · يگول اشتغل لحد <b>{t(c.claimedUntil)}</b></p>
          {c.note && <p className="text-slate-700">«{c.note}»</p>}
          <p className="text-xs text-violet-800">🤖 {c.evidence ? `آخر دليل بالنظام: ${c.evidence} (${t(c.evidenceAt)}) — صحّحت لحده.` : 'ما لگيت أي شغل مسجّل بالنظام بعد الانصراف التلقائي.'}</p>
          <div className="mt-2 flex gap-2">
            <button type="button" onClick={() => void decide(c.id, true)} className="rounded-lg bg-emerald-700 px-3 py-1 text-xs font-bold text-white">✔ وافق</button>
            <button type="button" onClick={() => void decide(c.id, false)} className="rounded-lg border border-slate-300 px-3 py-1 text-xs font-bold">✕ ارفض</button>
          </div>
        </div>
      ))}
    </div>
  )
}
