import { useEffect, useMemo, useState } from 'react'
import { api, type VehicleBooking, type VehicleOption } from '../api'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'

// ═══ حجز المركبات ═══
//
// ⚠️ حراس الباك (main.go):
//   POST /vehicle-bookings            → requireVehicleMgmt
//   PUT  /vehicle-bookings/{id}/decide → requireVehicleMgmt
//   PUT  /vehicle-bookings/{id}/cancel → requireVehicleMgmt + (صاحب الطلب أو دور ADMIN/MONITOR حقيقي)
//   GET  /vehicle-bookings            → requireAuth
// الصفحة كلها ورا RequirePermission vehicle_management، فالطلب والقرار
// يطلعون لكل من يوصل. الإلغاء لغير صاحب الطلب بس للأدمن/المراقب الحقيقي —
// المالك صار ضمن isAdminOrMonitor بالخادم (OWNER).

const STATUS_LABELS: Record<VehicleBooking['status'], { label: string; cls: string }> = {
  PENDING: { label: 'بانتظار القرار', cls: 'bg-amber-100 text-amber-700' },
  APPROVED: { label: 'معتمد', cls: 'bg-emerald-100 text-emerald-700' },
  REJECTED: { label: 'مرفوض', cls: 'bg-red-100 text-red-700' },
  CANCELLED: { label: 'ملغي', cls: 'bg-slate-100 text-slate-500' },
}

const CARD = 'rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]'
const INPUT = 'rounded-lg border border-slate-300 px-3 py-2'

const fmt = (s: string) => new Date(s).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' })

type Row = VehicleBooking & { started: boolean }

export default function VehicleBookingsPage() {
  const guard = useSaveGuard()
  const { employee } = useSession()
  const myId = employee?.id ?? ''
  const realRole = employee?.actualRole ?? employee?.role
  // نفس isAdminOrMonitor بالهاندلر — على الدور الحقيقي
  const canCancelOthers = realRole === 'ADMIN' || realRole === 'OWNER' || realRole === 'MONITOR'

  const [vehicles, setVehicles] = useState<VehicleOption[]>([])
  const [bookings, setBookings] = useState<Row[] | null>(null)
  const [reload, setReload] = useState(0)
  const [view, setView] = useState<'mine' | 'pending' | 'all'>('mine')
  const [statusFilter, setStatusFilter] = useState('')
  const [vehicleFilter, setVehicleFilter] = useState('')

  const [form, setForm] = useState({ vehicleId: '', purpose: '', startAt: '', endAt: '' })
  const [rejecting, setRejecting] = useState<Row | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  useEffect(() => {
    let alive = true
    api.getVehicleOptions().then((v) => { if (alive) setVehicles(v) }).catch(() => {})
    return () => { alive = false }
  }, [])

  useEffect(() => {
    let alive = true
    const filters: Parameters<typeof api.getVehicleBookings>[0] = {}
    if (view === 'mine') filters.requestedById = myId
    if (view === 'pending') filters.status = 'PENDING'
    else if (statusFilter) filters.status = statusFilter
    if (vehicleFilter) filters.vehicleId = vehicleFilter
    api.getVehicleBookings(filters)
      .then((list) => {
        if (!alive) return
        // «بدأ وقته» ينحسب هنا — Date.now ممنوعة وقت الرسم
        const now = Date.now()
        setBookings(list.map((b) => ({ ...b, started: new Date(b.startAt).getTime() <= now })))
      })
      .catch(() => { if (alive) setBookings([]) })
    return () => { alive = false }
  }, [view, statusFilter, vehicleFilter, myId, reload])

  const pendingCount = useMemo(() => (bookings ?? []).filter((b) => b.status === 'PENDING').length, [bookings])

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (form.endAt <= form.startAt) {
      alert('وقت النهاية لازم يكون بعد البداية')
      return
    }
    const ok = await guard.run('إرسال طلب الحجز', () => api.createVehicleBooking({
      vehicleId: form.vehicleId,
      purpose: form.purpose.trim(),
      startAt: `${form.startAt}:00`,
      endAt: `${form.endAt}:00`,
    }))
    if (!ok) return
    setForm({ vehicleId: '', purpose: '', startAt: '', endAt: '' })
    setReload((n) => n + 1)
  }

  const approve = async (b: Row) => {
    if (!confirm(`اعتماد حجز ${b.vehicle?.name ?? ''} لـ${b.requestedBy?.name ?? ''}؟`)) return
    if (await guard.run('اعتماد الحجز', () => api.decideVehicleBooking(b.id, { approve: true }))) setReload((n) => n + 1)
  }

  const reject = async () => {
    if (!rejecting || !rejectReason.trim()) return
    const id = rejecting.id
    if (await guard.run('رفض الحجز', () => api.decideVehicleBooking(id, { approve: false, rejectionReason: rejectReason.trim() }))) {
      setRejecting(null); setRejectReason('')
      setReload((n) => n + 1)
    }
  }

  const cancel = async (b: Row) => {
    if (!confirm('إلغاء هذا الحجز؟')) return
    if (await guard.run('إلغاء الحجز', () => api.cancelVehicleBooking(b.id))) setReload((n) => n + 1)
  }

  // نفس شروط CancelBooking بالسيرفس: صاحبه (أو أدمن/مراقب) + معلّق/معتمد + ما بدأ وقته
  const canCancel = (b: Row) =>
    (b.requestedById === myId || canCancelOthers) && (b.status === 'PENDING' || b.status === 'APPROVED') && !b.started

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
      <div dir="rtl" className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">حجز المركبات</h2>
          <p className="mt-1 text-slate-500">اطلب سيارة لفترة محددة، والمسؤول يعتمد أو يرفض</p>
        </div>

        <form onSubmit={submit} className={`${CARD} grid grid-cols-1 gap-3 sm:grid-cols-2`}>
          <h3 className="font-bold text-slate-700 sm:col-span-2">🚗 طلب حجز جديد</h3>
          <select required value={form.vehicleId} onChange={(e) => setForm({ ...form, vehicleId: e.target.value })} className={INPUT}>
            <option value="">-- اختر السيارة --</option>
            {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name} — {v.plateNumber}</option>)}
          </select>
          <input required placeholder="سبب الحجز" value={form.purpose} onChange={(e) => setForm({ ...form, purpose: e.target.value })} className={INPUT} />
          <div>
            <label className="mb-1 block text-xs text-slate-500">من</label>
            <input required type="datetime-local" value={form.startAt} onChange={(e) => setForm({ ...form, startAt: e.target.value })} className={`w-full ${INPUT}`} />
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-500">إلى</label>
            <input required type="datetime-local" value={form.endAt} onChange={(e) => setForm({ ...form, endAt: e.target.value })} className={`w-full ${INPUT}`} />
          </div>
          <button type="submit" disabled={guard.busy} className="sm:col-span-2 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white disabled:opacity-50">إرسال الطلب</button>
        </form>

        <div className="flex flex-wrap items-center gap-2 rounded-xl border border-white bg-white p-2 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          {([
            { key: 'mine', label: 'طلباتي' },
            { key: 'pending', label: 'بانتظار القرار' },
            { key: 'all', label: 'كل الحجوزات' },
          ] as const).map((t) => (
            <button
              key={t.key}
              onClick={() => setView(t.key)}
              className={`flex-1 rounded-lg px-3 py-2 text-sm font-bold transition-colors ${view === t.key ? 'bg-brand-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
            >
              {t.label}{t.key === 'pending' && view === 'pending' && bookings ? ` (${pendingCount})` : ''}
            </button>
          ))}
        </div>

        <div className="flex flex-wrap gap-3">
          <select value={vehicleFilter} onChange={(e) => setVehicleFilter(e.target.value)} className={`${INPUT} text-sm`}>
            <option value="">كل السيارات</option>
            {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name}</option>)}
          </select>
          {view !== 'pending' && (
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className={`${INPUT} text-sm`}>
              <option value="">كل الحالات</option>
              {Object.entries(STATUS_LABELS).map(([k, v]) => <option key={k} value={k}>{v.label}</option>)}
            </select>
          )}
        </div>

        <div className="overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full text-right text-sm">
            <thead className="bg-slate-50 text-slate-600">
              <tr>
                <th className="px-4 py-2">السيارة</th>
                <th className="px-4 py-2">الطالب</th>
                <th className="px-4 py-2">السبب</th>
                <th className="px-4 py-2">الفترة</th>
                <th className="px-4 py-2">الحالة</th>
                <th className="px-4 py-2">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {bookings === null && <tr><td colSpan={6} className="p-4 text-center text-slate-400">جاري التحميل...</td></tr>}
              {bookings?.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">ماكو حجوزات</td></tr>}
              {bookings?.map((b) => {
                const st = STATUS_LABELS[b.status]
                return (
                  <tr key={b.id}>
                    <td className="px-4 py-2 font-medium">{b.vehicle?.name ?? '-'}<div className="text-xs text-slate-400">{b.vehicle?.plateNumber}</div></td>
                    <td className="px-4 py-2">{b.requestedBy?.name ?? '-'}</td>
                    <td className="px-4 py-2 text-slate-600">{b.purpose}</td>
                    <td className="px-4 py-2 text-xs text-slate-500">{fmt(b.startAt)}<br />← {fmt(b.endAt)}</td>
                    <td className="px-4 py-2">
                      <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${st.cls}`}>{st.label}</span>
                      {b.approvedBy && <div className="mt-1 text-xs text-slate-400">بواسطة {b.approvedBy.name}</div>}
                      {b.rejectionReason && <div className="mt-1 text-xs text-red-500">السبب: {b.rejectionReason}</div>}
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex flex-wrap gap-1">
                        {b.status === 'PENDING' && (
                          <>
                            <button onClick={() => approve(b)} className="rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">✓ اعتماد</button>
                            <button onClick={() => { setRejecting(b); setRejectReason('') }} className="rounded-lg bg-red-50 px-2 py-1 text-xs font-bold text-red-600 hover:bg-red-100">✕ رفض</button>
                          </>
                        )}
                        {canCancel(b) && (
                          <button onClick={() => cancel(b)} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">إلغاء</button>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        {rejecting && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
              <h3 className="text-lg font-bold text-brand-900">رفض حجز {rejecting.vehicle?.name}</h3>
              <p className="mt-1 text-sm text-slate-500">{rejecting.requestedBy?.name} · {rejecting.purpose}</p>
              <textarea autoFocus value={rejectReason} onChange={(e) => setRejectReason(e.target.value)} placeholder="سبب الرفض (إجباري)" className={`mt-4 h-24 w-full ${INPUT}`} />
              <div className="mt-4 flex gap-3">
                <button onClick={reject} disabled={!rejectReason.trim() || guard.busy} className="flex-1 rounded-xl bg-red-600 py-2.5 font-bold text-white disabled:opacity-50">تأكيد الرفض</button>
                <button onClick={() => setRejecting(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600">إلغاء</button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
