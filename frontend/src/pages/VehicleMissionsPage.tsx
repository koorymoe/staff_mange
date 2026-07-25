import { useEffect, useState } from 'react'
import { api, type Vehicle, type VehicleMission, type Employee, type VehicleBooking } from '../api'
import { useSession, hasMonitorAccess } from '../session'

const STAR_FIELDS: { key: 'commitment' | 'vehicleCare' | 'driving' | 'cleanliness'; label: string }[] = [
  { key: 'commitment', label: 'الالتزام' },
  { key: 'vehicleCare', label: 'المحافظة على السيارة' },
  { key: 'driving', label: 'القيادة' },
  { key: 'cleanliness', label: 'النظافة' },
]

const bookingStatusLabel: Record<string, string> = {
  PENDING: 'معلّق',
  APPROVED: 'معتمد',
  REJECTED: 'مرفوض',
  CANCELLED: 'ملغى',
}
const bookingStatusClass: Record<string, string> = {
  PENDING: 'bg-amber-50 text-amber-700',
  APPROVED: 'bg-emerald-50 text-emerald-700',
  REJECTED: 'bg-red-50 text-red-700',
  CANCELLED: 'bg-slate-100 text-slate-500',
}

function elapsedSince(iso: string): string {
  const diffMin = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 60000))
  const h = Math.floor(diffMin / 60)
  const m = diffMin % 60
  return h > 0 ? `${h} ساعة و ${m} دقيقة` : `${m} دقيقة`
}

export default function VehicleMissionsPage() {
  const { employee } = useSession()
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [activeMissions, setActiveMissions] = useState<VehicleMission[]>([])
  const [historyMissions, setHistoryMissions] = useState<VehicleMission[]>([])
  const [, setTick] = useState(0)

  // start-mission form
  const [selVehicleId, setSelVehicleId] = useState('')
  const [purpose, setPurpose] = useState('')
  const [destination, setDestination] = useState('')
  const [startOdometer, setStartOdometer] = useState('')
  const [passengerIds, setPassengerIds] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  // end-mission modal
  const [endingId, setEndingId] = useState<string | null>(null)
  const [endOdometer, setEndOdometer] = useState('')
  const [endNotes, setEndNotes] = useState('')
  const [ending, setEnding] = useState(false)

  // history filters
  const [fVehicle, setFVehicle] = useState('')
  const [fDriver, setFDriver] = useState('')
  const [fFrom, setFFrom] = useState('')
  const [fTo, setFTo] = useState('')

  const [startWarning, setStartWarning] = useState('')

  // rating modal
  const [ratingMissionId, setRatingMissionId] = useState<string | null>(null)
  const [ratingScores, setRatingScores] = useState<Record<string, number>>({ commitment: 5, vehicleCare: 5, driving: 5, cleanliness: 5 })
  const [ratingNotes, setRatingNotes] = useState('')
  const [ratingSaving, setRatingSaving] = useState(false)
  const [ratingError, setRatingError] = useState('')

  // bookings
  const isManager = hasMonitorAccess(employee?.role)
  const [bookVehicleId, setBookVehicleId] = useState('')
  const [bookPurpose, setBookPurpose] = useState('')
  const [bookStartAt, setBookStartAt] = useState('')
  const [bookEndAt, setBookEndAt] = useState('')
  const [bookSaving, setBookSaving] = useState(false)
  const [bookError, setBookError] = useState('')
  const [pendingBookings, setPendingBookings] = useState<VehicleBooking[]>([])
  const [myBookings, setMyBookings] = useState<VehicleBooking[]>([])
  const [rejectingId, setRejectingId] = useState<string | null>(null)
  const [rejectReason, setRejectReason] = useState('')

  const loadPendingBookings = () => {
    if (isManager) api.getVehicleBookings({ status: 'PENDING' }).then(setPendingBookings)
  }
  const loadMyBookings = () => {
    if (employee) api.getVehicleBookings({ requestedById: employee.id }).then(setMyBookings)
  }

  const loadActive = () => api.getVehicleMissions({ status: 'IN_PROGRESS' }).then(setActiveMissions)
  const loadHistory = () => api.getVehicleMissions({
    status: 'COMPLETED',
    vehicleId: fVehicle || undefined,
    driverId: fDriver || undefined,
    from: fFrom || undefined,
    to: fTo || undefined,
  }).then(setHistoryMissions)

  useEffect(() => {
    api.getVehicles().then(setVehicles)
    api.getEmployees().then(setEmployees)
    loadActive()
    loadHistory()
    loadPendingBookings()
    loadMyBookings()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    loadHistory()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fVehicle, fDriver, fFrom, fTo])

  useEffect(() => {
    const t = setInterval(() => setTick((v) => v + 1), 30000)
    return () => clearInterval(t)
  }, [])

  const busyVehicleIds = new Set(activeMissions.map((m) => m.vehicleId))
  const selectedVehicle = vehicles.find((v) => v.id === selVehicleId) || null

  const handleSelectVehicle = (id: string) => {
    setSelVehicleId(id)
    const v = vehicles.find((x) => x.id === id)
    setStartOdometer(v ? String(v.currentOdometer) : '')
  }

  const togglePassenger = (id: string) => {
    setPassengerIds((prev) => (prev.includes(id) ? prev.filter((p) => p !== id) : [...prev, id]))
  }

  const handleStart = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!selVehicleId || !purpose.trim() || !destination.trim() || !startOdometer) {
      setError('يرجى تعبئة جميع الحقول المطلوبة')
      return
    }
    setSaving(true)
    setStartWarning('')
    try {
      const result = await api.startVehicleMission({
        vehicleId: selVehicleId,
        driverId: employee?.id,
        purpose,
        destination,
        startOdometer: Number(startOdometer),
        passengerIds,
      })
      if (result.bookingWarning) setStartWarning(result.bookingWarning)
      setSelVehicleId(''); setPurpose(''); setDestination(''); setStartOdometer(''); setPassengerIds([])
      loadActive()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'تعذر بدء المهمة')
    } finally {
      setSaving(false)
    }
  }

  const handleCreateBooking = async (e: React.FormEvent) => {
    e.preventDefault()
    setBookError('')
    if (!bookVehicleId || !bookPurpose.trim() || !bookStartAt || !bookEndAt) {
      setBookError('يرجى تعبئة جميع الحقول المطلوبة')
      return
    }
    setBookSaving(true)
    try {
      await api.createVehicleBooking({ vehicleId: bookVehicleId, purpose: bookPurpose, startAt: bookStartAt, endAt: bookEndAt })
      setBookVehicleId(''); setBookPurpose(''); setBookStartAt(''); setBookEndAt('')
      loadMyBookings()
      loadPendingBookings()
    } catch (err) {
      setBookError(err instanceof Error ? err.message : 'تعذر إنشاء الحجز')
    } finally {
      setBookSaving(false)
    }
  }

  const handleApproveBooking = async (id: string) => {
    try {
      await api.decideVehicleBooking(id, { approve: true })
      loadPendingBookings(); loadMyBookings()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر اعتماد الحجز')
    }
  }

  const handleRejectBooking = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!rejectingId) return
    try {
      await api.decideVehicleBooking(rejectingId, { approve: false, rejectionReason: rejectReason || 'بدون سبب' })
      setRejectingId(null); setRejectReason('')
      loadPendingBookings(); loadMyBookings()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفض الحجز')
    }
  }

  const handleCancelBooking = async (id: string) => {
    try {
      await api.cancelVehicleBooking(id)
      loadMyBookings(); loadPendingBookings()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إلغاء الحجز')
    }
  }

  const handleSubmitRating = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!ratingMissionId) return
    setRatingSaving(true)
    setRatingError('')
    try {
      await api.createVehicleMissionRating(ratingMissionId, {
        commitment: ratingScores.commitment,
        vehicleCare: ratingScores.vehicleCare,
        driving: ratingScores.driving,
        cleanliness: ratingScores.cleanliness,
        notes: ratingNotes || undefined,
      })
      setRatingMissionId(null); setRatingNotes(''); setRatingScores({ commitment: 5, vehicleCare: 5, driving: 5, cleanliness: 5 })
      loadHistory()
    } catch (err) {
      setRatingError(err instanceof Error ? err.message : 'تعذر حفظ التقييم')
    } finally {
      setRatingSaving(false)
    }
  }

  const handleEnd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!endingId || !endOdometer) return
    setEnding(true)
    try {
      await api.endVehicleMission(endingId, { endOdometer: Number(endOdometer), notes: endNotes || undefined })
      setEndingId(null); setEndOdometer(''); setEndNotes('')
      loadActive()
      loadHistory()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إنهاء المهمة')
    } finally {
      setEnding(false)
    }
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">مهام المركبات</h2>
        <p className="mt-1 text-slate-500">بدء مهمة، متابعة المهام النشطة، وسجل المهام المكتملة</p>
      </div>

      {startWarning && (
        <div className="rounded-lg bg-amber-50 px-4 py-3 text-sm font-medium text-amber-800 border border-amber-200">
          {startWarning}
          <button onClick={() => setStartWarning('')} className="mr-3 text-xs underline">إغلاق</button>
        </div>
      )}

      {/* Start mission form */}
      <form onSubmit={handleStart} className="grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
        <h3 className="sm:col-span-2 text-lg font-bold text-brand-800">بدء مهمة جديدة</h3>
        {error && <p className="sm:col-span-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{error}</p>}

        <div>
          <label className="mb-1 block text-xs text-slate-500">السيارة</label>
          <select value={selVehicleId} onChange={(e) => handleSelectVehicle(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2">
            <option value="">-- اختر سيارة --</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id} disabled={busyVehicleIds.has(v.id)}>
                {v.name} - {v.plateNumber} {busyVehicleIds.has(v.id) ? '(بمهمة حالياً)' : ''}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="mb-1 block text-xs text-slate-500">عداد البداية</label>
          <input
            type="number"
            value={startOdometer}
            onChange={(e) => setStartOdometer(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2"
          />
          {selectedVehicle && <p className="mt-1 text-xs text-slate-400">آخر عداد مسجل: {selectedVehicle.currentOdometer}</p>}
        </div>

        <input required placeholder="سبب المهمة" value={purpose} onChange={(e) => setPurpose(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
        <input required placeholder="الوجهة" value={destination} onChange={(e) => setDestination(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />

        <div className="sm:col-span-2">
          <label className="mb-2 block text-xs font-bold text-slate-600">المرافقون (اختياري)</label>
          <div className="grid max-h-40 grid-cols-2 gap-2 overflow-y-auto rounded-lg border border-slate-200 p-3 sm:grid-cols-3">
            {employees.map((emp) => (
              <label key={emp.id} className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={passengerIds.includes(emp.id)} onChange={() => togglePassenger(emp.id)} />
                {emp.name}
              </label>
            ))}
          </div>
        </div>

        <button type="submit" disabled={saving} className="sm:col-span-2 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white shadow-md disabled:opacity-50">
          {saving ? 'جاري البدء...' : 'بدء المهمة'}
        </button>
      </form>

      {/* Active missions */}
      <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-3 text-white font-bold">المهام النشطة ({activeMissions.length})</div>
        <table className="w-full text-right text-sm">
          <thead className="bg-slate-50 text-slate-600">
            <tr>
              <th className="px-4 py-2">السيارة</th>
              <th className="px-4 py-2">السائق</th>
              <th className="px-4 py-2">الغرض</th>
              <th className="px-4 py-2">الوجهة</th>
              <th className="px-4 py-2">منذ</th>
              <th className="px-4 py-2"></th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {activeMissions.map((m) => (
              <tr key={m.id}>
                <td className="px-4 py-2 font-medium">{m.vehicle?.name || '-'}</td>
                <td className="px-4 py-2 text-slate-500">{m.driver?.name || '-'}</td>
                <td className="px-4 py-2 text-slate-500">{m.purpose}</td>
                <td className="px-4 py-2 text-slate-500">{m.destination}</td>
                <td className="px-4 py-2 text-slate-500">{elapsedSince(m.startedAt)}</td>
                <td className="px-4 py-2">
                  <button
                    onClick={() => { setEndingId(m.id); setEndOdometer(''); setEndNotes('') }}
                    className="rounded-lg bg-red-50 px-3 py-1 text-xs font-bold text-red-700 hover:bg-red-100"
                  >
                    إنهاء المهمة
                  </button>
                </td>
              </tr>
            ))}
            {activeMissions.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">لا توجد مهام نشطة حالياً</td></tr>}
          </tbody>
        </table>
      </div>

      {/* End mission modal */}
      {endingId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <form onSubmit={handleEnd} dir="rtl" className="w-full max-w-md space-y-3 rounded-xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-bold text-brand-800">إنهاء المهمة</h3>
            <input
              required
              type="number"
              placeholder="عداد النهاية"
              value={endOdometer}
              onChange={(e) => setEndOdometer(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
            />
            <input
              placeholder="ملاحظات (اختياري)"
              value={endNotes}
              onChange={(e) => setEndNotes(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
            />
            <div className="flex gap-2">
              <button type="submit" disabled={ending} className="flex-1 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white disabled:opacity-50">
                {ending ? 'جاري الإنهاء...' : 'تأكيد الإنهاء'}
              </button>
              <button type="button" onClick={() => setEndingId(null)} className="rounded-lg border border-slate-300 px-5 py-2 font-medium text-slate-600">
                إلغاء
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Vehicle bookings */}
      <div className="space-y-4">
        <h3 className="text-lg font-bold text-brand-800">حجوزات المركبات</h3>

        <form onSubmit={handleCreateBooking} className="grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
          <h4 className="sm:col-span-2 font-bold text-brand-700">حجز سيارة</h4>
          {bookError && <p className="sm:col-span-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{bookError}</p>}
          <select value={bookVehicleId} onChange={(e) => setBookVehicleId(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
            <option value="">-- اختر سيارة --</option>
            {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name} - {v.plateNumber}</option>)}
          </select>
          <input required placeholder="سبب الحجز" value={bookPurpose} onChange={(e) => setBookPurpose(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <div>
            <label className="mb-1 block text-xs text-slate-500">من</label>
            <input required type="datetime-local" value={bookStartAt} onChange={(e) => setBookStartAt(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-500">إلى</label>
            <input required type="datetime-local" value={bookEndAt} onChange={(e) => setBookEndAt(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
          </div>
          <button type="submit" disabled={bookSaving} className="sm:col-span-2 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white shadow-md disabled:opacity-50">
            {bookSaving ? 'جاري الإرسال...' : 'إرسال طلب الحجز'}
          </button>
        </form>

        {isManager && (
          <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-3 text-white font-bold">الحجوزات المعلقة ({pendingBookings.length})</div>
            <table className="w-full text-right text-sm">
              <thead className="bg-slate-50 text-slate-600">
                <tr>
                  <th className="px-4 py-2">السيارة</th>
                  <th className="px-4 py-2">مقدم الطلب</th>
                  <th className="px-4 py-2">السبب</th>
                  <th className="px-4 py-2">من</th>
                  <th className="px-4 py-2">إلى</th>
                  <th className="px-4 py-2"></th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {pendingBookings.map((b) => (
                  <tr key={b.id}>
                    <td className="px-4 py-2 font-medium">{b.vehicle?.name || '-'}</td>
                    <td className="px-4 py-2 text-slate-500">{b.requestedBy?.name || '-'}</td>
                    <td className="px-4 py-2 text-slate-500">{b.purpose}</td>
                    <td className="px-4 py-2 text-slate-500">{new Date(b.startAt).toLocaleString('ar-IQ')}</td>
                    <td className="px-4 py-2 text-slate-500">{new Date(b.endAt).toLocaleString('ar-IQ')}</td>
                    <td className="px-4 py-2">
                      <div className="flex gap-2">
                        <button onClick={() => handleApproveBooking(b.id)} className="rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">اعتماد</button>
                        <button onClick={() => { setRejectingId(b.id); setRejectReason('') }} className="rounded-lg bg-red-50 px-3 py-1 text-xs font-bold text-red-700 hover:bg-red-100">رفض</button>
                      </div>
                    </td>
                  </tr>
                ))}
                {pendingBookings.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">لا توجد حجوزات معلقة</td></tr>}
              </tbody>
            </table>
          </div>
        )}

        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-3 text-white font-bold">حجوزاتي ({myBookings.length})</div>
          <table className="w-full text-right text-sm">
            <thead className="bg-slate-50 text-slate-600">
              <tr>
                <th className="px-4 py-2">السيارة</th>
                <th className="px-4 py-2">السبب</th>
                <th className="px-4 py-2">من</th>
                <th className="px-4 py-2">إلى</th>
                <th className="px-4 py-2">الحالة</th>
                <th className="px-4 py-2"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {myBookings.map((b) => (
                <tr key={b.id}>
                  <td className="px-4 py-2 font-medium">{b.vehicle?.name || '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{b.purpose}</td>
                  <td className="px-4 py-2 text-slate-500">{new Date(b.startAt).toLocaleString('ar-IQ')}</td>
                  <td className="px-4 py-2 text-slate-500">{new Date(b.endAt).toLocaleString('ar-IQ')}</td>
                  <td className="px-4 py-2">
                    <span className={`rounded-full px-2 py-1 text-xs font-bold ${bookingStatusClass[b.status]}`}>{bookingStatusLabel[b.status]}</span>
                    {b.status === 'REJECTED' && b.rejectionReason && <p className="mt-1 text-xs text-slate-400">{b.rejectionReason}</p>}
                  </td>
                  <td className="px-4 py-2">
                    {(b.status === 'PENDING' || b.status === 'APPROVED') && new Date(b.startAt) > new Date() && (
                      <button onClick={() => handleCancelBooking(b.id)} className="rounded-lg bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">إلغاء</button>
                    )}
                  </td>
                </tr>
              ))}
              {myBookings.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">لا توجد حجوزات</td></tr>}
            </tbody>
          </table>
        </div>
      </div>

      {/* Reject booking modal */}
      {rejectingId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <form onSubmit={handleRejectBooking} dir="rtl" className="w-full max-w-md space-y-3 rounded-xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-bold text-brand-800">رفض الحجز</h3>
            <input
              placeholder="سبب الرفض"
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
            />
            <div className="flex gap-2">
              <button type="submit" className="flex-1 rounded-lg bg-red-600 px-5 py-2 font-medium text-white">تأكيد الرفض</button>
              <button type="button" onClick={() => setRejectingId(null)} className="rounded-lg border border-slate-300 px-5 py-2 font-medium text-slate-600">إلغاء</button>
            </div>
          </form>
        </div>
      )}

      {/* History */}
      <div className="space-y-3">
        <h3 className="text-lg font-bold text-brand-800">سجل المهام</h3>
        <div className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4">
          <select value={fVehicle} onChange={(e) => setFVehicle(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">-- كل السيارات --</option>
            {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name}</option>)}
          </select>
          <select value={fDriver} onChange={(e) => setFDriver(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">-- كل السائقين --</option>
            {employees.map((emp) => <option key={emp.id} value={emp.id}>{emp.name}</option>)}
          </select>
          <input type="date" value={fFrom} onChange={(e) => setFFrom(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
          <input type="date" value={fTo} onChange={(e) => setFTo(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>

        <div className="overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full text-right text-sm">
            <thead className="bg-slate-50 text-slate-600">
              <tr>
                <th className="px-4 py-2">السيارة</th>
                <th className="px-4 py-2">السائق</th>
                <th className="px-4 py-2">الغرض</th>
                <th className="px-4 py-2">الوجهة</th>
                <th className="px-4 py-2">البداية</th>
                <th className="px-4 py-2">النهاية</th>
                <th className="px-4 py-2">العداد (بداية/نهاية)</th>
                <th className="px-4 py-2">المسافة</th>
                <th className="px-4 py-2">المرافقون</th>
                <th className="px-4 py-2">ملاحظات</th>
                <th className="px-4 py-2">تقييم السائق</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {historyMissions.map((m) => (
                <tr key={m.id}>
                  <td className="px-4 py-2 font-medium">{m.vehicle?.name || '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{m.driver?.name || '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{m.purpose}</td>
                  <td className="px-4 py-2 text-slate-500">{m.destination}</td>
                  <td className="px-4 py-2 text-slate-500">{new Date(m.startedAt).toLocaleString('ar-IQ')}</td>
                  <td className="px-4 py-2 text-slate-500">{m.endedAt ? new Date(m.endedAt).toLocaleString('ar-IQ') : '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{m.startOdometer} / {m.endOdometer ?? '-'}</td>
                  <td className="px-4 py-2 font-bold text-brand-700">{m.distanceKm != null ? `${m.distanceKm} كم` : '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{m.passengers.map((p) => p.employee?.name).filter(Boolean).join(', ') || '-'}</td>
                  <td className="px-4 py-2 text-slate-500">{m.notes || '-'}</td>
                  <td className="px-4 py-2 text-slate-500">
                    {m.rating ? (
                      <div className="text-xs">
                        <div>التزام: {m.rating.commitment} | سيارة: {m.rating.vehicleCare}</div>
                        <div>قيادة: {m.rating.driving} | نظافة: {m.rating.cleanliness}</div>
                      </div>
                    ) : (
                      <button
                        onClick={() => { setRatingMissionId(m.id); setRatingScores({ commitment: 5, vehicleCare: 5, driving: 5, cleanliness: 5 }); setRatingNotes(''); setRatingError('') }}
                        className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
                      >
                        تقييم
                      </button>
                    )}
                  </td>
                </tr>
              ))}
              {historyMissions.length === 0 && <tr><td colSpan={11} className="p-4 text-center text-slate-400">لا توجد مهام مكتملة مطابقة</td></tr>}
            </tbody>
          </table>
        </div>
      </div>

      {/* Rating modal */}
      {ratingMissionId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <form onSubmit={handleSubmitRating} dir="rtl" className="w-full max-w-md space-y-3 rounded-xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-bold text-brand-800">تقييم السائق</h3>
            {ratingError && <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{ratingError}</p>}
            {STAR_FIELDS.map((f) => (
              <div key={f.key} className="flex items-center justify-between">
                <label className="text-sm text-slate-600">{f.label}</label>
                <select
                  value={ratingScores[f.key]}
                  onChange={(e) => setRatingScores((prev) => ({ ...prev, [f.key]: Number(e.target.value) }))}
                  className="rounded-lg border border-slate-300 px-3 py-1"
                >
                  {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                </select>
              </div>
            ))}
            <input
              placeholder="ملاحظات (اختياري)"
              value={ratingNotes}
              onChange={(e) => setRatingNotes(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2"
            />
            <div className="flex gap-2">
              <button type="submit" disabled={ratingSaving} className="flex-1 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white disabled:opacity-50">
                {ratingSaving ? 'جاري الحفظ...' : 'حفظ التقييم'}
              </button>
              <button type="button" onClick={() => setRatingMissionId(null)} className="rounded-lg border border-slate-300 px-5 py-2 font-medium text-slate-600">
                إلغاء
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  )
}
