import { useState } from 'react'
import { api, isStealthSession } from '../api'
import { useSession } from '../session'

// ═══ 🔍 وضع الفحص — (ع) 10-23 ═══
// «نفس المعلومات الي تضهرله تضهرلي بحسابه، بس اي اجراء اسوي ما اريده ينحسب».
// الخادم يحميه: المالك وحده، ساعتين، بلا حضور ولا «متصل»، وأي تعديل ينرفض
// (ما ينحفظ شي باسم الموظف ولا تنحسب مخالفة). هنا: الشريط والزر.

/** الشريط فوق الشاشة طول ما المالك يفحص حساب موظف. */
export function InspectBar() {
  if (!isStealthSession()) return null
  let name = ''
  try { name = sessionStorage.getItem('inspectName') || '' } catch { /* */ }
  return (
    <div className="sticky top-0 z-[80] mb-3 flex flex-wrap items-center justify-center gap-3 rounded-xl bg-slate-900 px-4 py-2 text-sm text-white" dir="rtl">
      <span className="font-bold">🔍 وضع الفحص — تشوف حساب {name || 'الموظف'} مثل ما يشوفه</span>
      <span className="text-xs text-slate-300">قراءة بس · أي تعديل ما ينحفظ ولا ينحسب عليه</span>
      <button type="button" onClick={() => api.inspectExit()}
        className="min-h-9 rounded-lg bg-amber-400 px-4 text-sm font-extrabold text-slate-900">⏏️ رجوع لحسابي</button>
    </div>
  )
}

/** زر «🔍 افحص حسابه» — للمالك بس، وما يطلع على حسابه هو. */
export function InspectButton({ employeeId, name }: { employeeId: string; name?: string }) {
  const { employee } = useSession()
  const [busy, setBusy] = useState(false)
  if (employee?.actualRole !== 'OWNER' || isStealthSession() || !employeeId || employeeId === employee.id) return null
  const go = async () => {
    setBusy(true)
    try { await api.inspectEnter(employeeId) } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر الفتح'); setBusy(false) }
  }
  return (
    <button type="button" disabled={busy} onClick={() => void go()} title={name ? `تشوف حساب ${name} مثل ما يشوفه — قراءة بس` : undefined}
      className="min-h-10 rounded-xl bg-slate-900 px-4 text-sm font-extrabold text-white disabled:opacity-50">
      {busy ? '…' : '🔍 افحص حسابه'}
    </button>
  )
}
