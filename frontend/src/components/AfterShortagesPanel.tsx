import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { useSession, hasRole, isAdminOrOwner } from '../session'
import { fmtDateTime } from '../utils/date'

// ═══ نواقص العدّة بعد الشغل — «مبرَّر» (دازاي 11-18) ═══
// النقص الي يلگاه الفني بعدّته بعد الحجز ينحسب عليه بالتقييم، إلا إذا إداري الكميات
// گال «مبرَّر» (مثلاً زميل أخذ الأداة منه). الخادم يمنع الموظف يبرر نقصه هو.
interface Row { bookingId: string; code: string | null; employeeId: string; employee: string; missingItems: string | null; checkedAt: string; justifiedAt: string | null }

export default function AfterShortagesPanel() {
  const { employee, permissions } = useSession()
  const can = isAdminOrOwner(employee) || hasRole(employee, 'PROCUREMENT_ADMIN') || permissions.includes('inventory')
  const [rows, setRows] = useState<Row[] | null>(null)
  const [busy, setBusy] = useState('')
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    apiRequest<{ items: Row[] }>('/inventory/after-shortages').then((d) => setRows(d.items)).catch(() => setRows([]))
  }, [])
  useEffect(() => { if (can) load() }, [can, load])
  if (!can || !rows || rows.length === 0) return null
  const justify = async (r: Row) => {
    const k = r.bookingId + r.employeeId
    setBusy(k); setErr('')
    try { await apiRequest('/inventory/after-shortages/justify', { method: 'POST', body: JSON.stringify({ bookingId: r.bookingId, employeeId: r.employeeId }) }); load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy('') }
  }
  const open = rows.filter((r) => !r.justifiedAt).length
  return (
    <div dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-4">
      <p className="font-extrabold text-[#0f2040]">🧰 نواقص العدّة بعد الشغل {open > 0 && <span className="text-amber-700">({open} تنتظر)</span>}</p>
      <p className="mb-2 text-xs text-slate-500">النقص ينحسب على الفني بالتقييم. إذا عنده مبرر (زميل أخذها، أو انصرفت بالشغل) اضغط «مبرَّر» فما ينحسب.</p>
      {err && <p className="mb-2 text-xs text-red-600">{err}</p>}
      <ul className="divide-y divide-slate-100">
        {rows.map((r) => (
          <li key={r.bookingId + r.employeeId} className="flex flex-wrap items-center justify-between gap-2 py-2">
            <span className="text-sm">
              <b>{r.employee}</b>{r.code ? ` · حجز ${r.code}` : ''}
              {r.missingItems && <span className="block text-xs text-amber-800">ناقص: {r.missingItems}</span>}
              <span className="block text-[11px] text-slate-400">{fmtDateTime(r.checkedAt)}</span>
            </span>
            {r.justifiedAt
              ? <span className="rounded-full bg-sky-50 px-3 py-1 text-xs font-bold text-sky-700">✓ مبرَّر</span>
              : <button type="button" disabled={busy === r.bookingId + r.employeeId} onClick={() => void justify(r)}
                  className="rounded-lg bg-emerald-600 px-3 py-1 text-xs font-bold text-white hover:bg-emerald-700 disabled:opacity-50">مبرَّر — ما ينحسب</button>}
          </li>
        ))}
      </ul>
    </div>
  )
}
