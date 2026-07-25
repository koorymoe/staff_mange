import { useEffect, useState } from 'react'
import { api, type Vehicle, type VehicleMission, type Employee } from '../api'
import { useSession } from '../session'

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
    try {
      await api.startVehicleMission({
        vehicleId: selVehicleId,
        driverId: employee?.id,
        purpose,
        destination,
        startOdometer: Number(startOdometer),
        passengerIds,
      })
      setSelVehicleId(''); setPurpose(''); setDestination(''); setStartOdometer(''); setPassengerIds([])
      loadActive()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'تعذر بدء المهمة')
    } finally {
      setSaving(false)
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
                </tr>
              ))}
              {historyMissions.length === 0 && <tr><td colSpan={10} className="p-4 text-center text-slate-400">لا توجد مهام مكتملة مطابقة</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
