import { useEffect, useState } from 'react'
import { api, type DriverRatingSummary, type Employee, type VehicleMission, type VehicleOption, type VehicleTool } from '../api'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'

// ═══ مهمات المركبات ═══
//
// ⚠️ حراس الباك (main.go + الهاندلر):
//   POST /vehicle-missions              → requireVehicleMgmt؛ driverId لغير نفسه بس لدور ADMIN/MONITOR حقيقي (وإلا 403)
//   PUT  /vehicle-missions/{id}/end     → requireVehicleMgmt؛ السائق نفسه أو ADMIN/MONITOR حقيقي (وإلا 403)
//   GET  /vehicle-missions              → requireVehicleMgmt
//   POST /vehicle-missions/{id}/rating  → requireVehicleMgmt (المهمة لازم COMPLETED وما متقيّمة)
//   POST /vehicle-missions/{id}/tool-check → requireLeader (ADMIN/OWNER أو ليدر)
//   GET  /employees/{id}/driver-rating-summary → requireAuth
// الصفحة كلها ورا RequirePermission vehicle_management.
// المالك صار ضمن isAdminOrMonitor بالخادم (OWNER).

const CARD = 'rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]'
const INPUT = 'rounded-lg border border-slate-300 px-3 py-2'

const fmt = (s: string) => new Date(s).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' })

const RATING_FIELDS = [
  { key: 'commitment', label: 'الالتزام' },
  { key: 'vehicleCare', label: 'المحافظة على السيارة' },
  { key: 'driving', label: 'القيادة' },
  { key: 'cleanliness', label: 'النظافة' },
] as const
type RatingKey = typeof RATING_FIELDS[number]['key']

export default function VehicleMissionsPage() {
  const guard = useSaveGuard()
  const { employee } = useSession()
  const myId = employee?.id ?? ''
  const realRole = employee?.actualRole ?? employee?.role
  // نفس isAdminOrMonitor بالهاندلر — على الدور الحقيقي
  const isAdminOrMonitor = realRole === 'ADMIN' || realRole === 'OWNER' || realRole === 'MONITOR'
  // RequireLeader: ADMIN/OWNER (بالواجهة الاثنين role=ADMIN) أو ليدر
  const canToolCheck = employee?.role === 'ADMIN' || !!employee?.isLeader

  const [vehicles, setVehicles] = useState<VehicleOption[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [reload, setReload] = useState(0)

  const [active, setActive] = useState<VehicleMission[] | null>(null)
  const [history, setHistory] = useState<VehicleMission[] | null>(null)
  const [filters, setFilters] = useState({ vehicleId: '', driverId: '', status: '' as '' | 'IN_PROGRESS' | 'COMPLETED', from: '', to: '' })

  const [form, setForm] = useState({ vehicleId: '', driverId: '', purpose: '', destination: '', startOdometer: '', passengerIds: [] as string[] })
  const [warning, setWarning] = useState<string | null>(null)

  const [ending, setEnding] = useState<VehicleMission | null>(null)
  const [endForm, setEndForm] = useState({ endOdometer: '', notes: '' })

  const [rating, setRating] = useState<VehicleMission | null>(null)
  const [ratingForm, setRatingForm] = useState<Record<RatingKey, number> & { notes: string }>({ commitment: 5, vehicleCare: 5, driving: 5, cleanliness: 5, notes: '' })

  const [toolCheck, setToolCheck] = useState<VehicleMission | null>(null)
  const [tools, setTools] = useState<VehicleTool[] | null>(null)
  const [missing, setMissing] = useState<string[]>([])

  const [summaryFor, setSummaryFor] = useState<{ id: string; name: string } | null>(null)
  const [summary, setSummary] = useState<DriverRatingSummary | null>(null)

  useEffect(() => {
    let alive = true
    api.getVehicleOptions().then((v) => { if (alive) setVehicles(v) }).catch(() => {})
    api.getEmployees().then((e) => { if (alive) setEmployees(e.filter((x) => x.status === 'ACTIVE')) }).catch(() => {})
    return () => { alive = false }
  }, [])

  // المهمات الجارية: الأدمن/المراقب يشوف الكل (يقدر ينهيها)، والباقي مهماته بس
  useEffect(() => {
    let alive = true
    api.getVehicleMissions(isAdminOrMonitor ? { status: 'IN_PROGRESS' } : { status: 'IN_PROGRESS', driverId: myId })
      .then((l) => { if (alive) setActive(l) })
      .catch(() => { if (alive) setActive([]) })
    return () => { alive = false }
  }, [isAdminOrMonitor, myId, reload])

  useEffect(() => {
    let alive = true
    api.getVehicleMissions({
      vehicleId: filters.vehicleId || undefined,
      driverId: filters.driverId || undefined,
      status: filters.status || undefined,
      from: filters.from || undefined,
      to: filters.to ? `${filters.to}T23:59:59` : undefined,
    })
      .then((l) => { if (alive) setHistory(l) })
      .catch(() => { if (alive) setHistory([]) })
    return () => { alive = false }
  }, [filters, reload])

  useEffect(() => {
    if (!summaryFor) return
    let alive = true
    api.getDriverRatingSummary(summaryFor.id)
      .then((s) => { if (alive) setSummary(s) })
      .catch(() => { if (alive) setSummary(null) })
    return () => { alive = false }
  }, [summaryFor, reload])

  useEffect(() => {
    if (!toolCheck) return
    let alive = true
    api.getVehicleTools(toolCheck.vehicleId)
      .then((t) => { if (alive) setTools(t) })
      .catch(() => { if (alive) setTools([]) })
    return () => { alive = false }
  }, [toolCheck])

  const openToolCheck = (m: VehicleMission) => { setTools(null); setMissing([]); setToolCheck(m) }

  const start = async (e: React.FormEvent) => {
    e.preventDefault()
    setWarning(null)
    const res = await guard.run('بدء المهمة', () => api.startVehicleMission({
      vehicleId: form.vehicleId,
      // سائق غير نفسه بس للأدمن/المراقب — غيرهم الحقل ما يطلع أصلاً
      driverId: isAdminOrMonitor && form.driverId ? form.driverId : undefined,
      purpose: form.purpose.trim(),
      destination: form.destination.trim(),
      startOdometer: Number(form.startOdometer),
      passengerIds: form.passengerIds.length ? form.passengerIds : undefined,
    }))
    if (!res) return
    if (res.bookingWarning) setWarning(res.bookingWarning)
    setForm({ vehicleId: '', driverId: '', purpose: '', destination: '', startOdometer: '', passengerIds: [] })
    setReload((n) => n + 1)
    if (res.requiresToolCheck && canToolCheck) openToolCheck(res)
  }

  const end = async () => {
    if (!ending || endForm.endOdometer === '') return
    if (Number(endForm.endOdometer) < ending.startOdometer) {
      alert('عداد النهاية لازم يكون أكبر أو يساوي عداد البداية')
      return
    }
    const id = ending.id
    const ok = await guard.run('إنهاء المهمة', () => api.endVehicleMission(id, { endOdometer: Number(endForm.endOdometer), notes: endForm.notes.trim() || undefined }))
    if (!ok) return
    setEnding(null); setEndForm({ endOdometer: '', notes: '' })
    setReload((n) => n + 1)
  }

  const submitRating = async () => {
    if (!rating) return
    const id = rating.id
    const ok = await guard.run('تقييم المهمة', () => api.createVehicleMissionRating(id, {
      commitment: ratingForm.commitment, vehicleCare: ratingForm.vehicleCare, driving: ratingForm.driving, cleanliness: ratingForm.cleanliness,
      notes: ratingForm.notes.trim() || undefined,
    }))
    if (!ok) return
    setRating(null)
    setReload((n) => n + 1)
  }

  const submitToolCheck = async () => {
    if (!toolCheck) return
    const id = toolCheck.id
    if (await guard.run('حفظ فحص الأدوات', () => api.createVehicleMissionToolCheck(id, missing))) setToolCheck(null)
  }

  // المهمة تنتهي من سائقها أو من أدمن/مراقب حقيقي
  const canEnd = (m: VehicleMission) => m.status === 'IN_PROGRESS' && (m.driverId === myId || isAdminOrMonitor)

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
      <div dir="rtl" className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">مهمات المركبات</h2>
          <p className="mt-1 text-slate-500">بدء وإنهاء المهمات بالعداد، وتقييم السائقين بعد المهمة</p>
        </div>

        {/* المهمات الجارية */}
        <div className={CARD}>
          <h3 className="mb-3 font-bold text-slate-700">🚦 {isAdminOrMonitor ? 'المهمات الجارية' : 'مهمتي الجارية'}</h3>
          {active === null ? (
            <p className="text-sm text-slate-400">جاري التحميل...</p>
          ) : active.length === 0 ? (
            <p className="text-sm text-slate-400">ماكو مهمة جارية</p>
          ) : (
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {active.map((m) => (
                <div key={m.id} className="rounded-xl border border-amber-200 bg-amber-50/60 p-4">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <p className="font-bold text-brand-900">{m.vehicle?.name} <span className="text-xs text-slate-500">{m.vehicle?.plateNumber}</span></p>
                      <p className="text-sm text-slate-600">{m.driver?.name} · {m.destination}</p>
                      <p className="text-xs text-slate-500">{m.purpose}</p>
                      <p className="mt-1 text-xs text-slate-400">بدأت {fmt(m.startedAt)} · عداد {m.startOdometer.toLocaleString()}</p>
                    </div>
                    <div className="flex shrink-0 flex-col gap-1">
                      {canEnd(m) && (
                        <button onClick={() => { setEnding(m); setEndForm({ endOdometer: '', notes: '' }) }} className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white">🏁 إنهاء</button>
                      )}
                      {canToolCheck && (
                        <button onClick={() => openToolCheck(m)} className="rounded-lg bg-white px-3 py-1.5 text-xs font-bold text-brand-700 border border-brand-200">🧰 فحص الأدوات</button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* بدء مهمة */}
        <form onSubmit={start} className={`${CARD} grid grid-cols-1 gap-3 sm:grid-cols-3`}>
          <h3 className="font-bold text-slate-700 sm:col-span-3">🚗 بدء مهمة جديدة</h3>
          <select required value={form.vehicleId} onChange={(e) => setForm({ ...form, vehicleId: e.target.value })} className={INPUT}>
            <option value="">-- اختر السيارة --</option>
            {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name} — {v.plateNumber}</option>)}
          </select>
          <input required type="number" min="0" placeholder="عداد البداية (كم)" value={form.startOdometer} onChange={(e) => setForm({ ...form, startOdometer: e.target.value })} className={INPUT} />
          {isAdminOrMonitor ? (
            <select value={form.driverId} onChange={(e) => setForm({ ...form, driverId: e.target.value })} className={INPUT}>
              <option value="">السائق: أنا</option>
              {employees.filter((x) => x.id !== myId).map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}
            </select>
          ) : (
            <div className="rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-600">السائق: {employee?.name}</div>
          )}
          <input required placeholder="سبب المهمة" value={form.purpose} onChange={(e) => setForm({ ...form, purpose: e.target.value })} className={INPUT} />
          <input required placeholder="الوجهة" value={form.destination} onChange={(e) => setForm({ ...form, destination: e.target.value })} className={INPUT} />
          <select
            value=""
            onChange={(e) => { const id = e.target.value; if (id && !form.passengerIds.includes(id)) setForm({ ...form, passengerIds: [...form.passengerIds, id] }) }}
            className={INPUT}
          >
            <option value="">+ إضافة راكب (اختياري)</option>
            {employees.filter((x) => !form.passengerIds.includes(x.id)).map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}
          </select>
          {form.passengerIds.length > 0 && (
            <div className="flex flex-wrap gap-1 sm:col-span-3">
              {form.passengerIds.map((id) => (
                <button type="button" key={id} onClick={() => setForm({ ...form, passengerIds: form.passengerIds.filter((p) => p !== id) })} className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">
                  {employees.find((x) => x.id === id)?.name ?? id} ✕
                </button>
              ))}
            </div>
          )}
          <button type="submit" disabled={guard.busy} className="sm:col-span-3 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white disabled:opacity-50">بدء المهمة</button>
          {warning && <p className="sm:col-span-3 rounded-lg bg-amber-50 px-3 py-2 text-sm font-bold text-amber-700">⚠️ {warning}</p>}
        </form>

        {/* ملخص تقييم السائق */}
        {summaryFor && (
          <div className={`${CARD} border-brand-100`}>
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-bold text-slate-700">⭐ تقييم السائق: {summaryFor.name}</h3>
              <button onClick={() => { setSummaryFor(null); setSummary(null) }} className="text-xs text-slate-400">✕ إغلاق</button>
            </div>
            {!summary ? (
              <p className="text-sm text-slate-400">جاري التحميل...</p>
            ) : summary.ratingsCount === 0 ? (
              <p className="text-sm text-slate-400">ما عنده تقييمات بعد</p>
            ) : (
              <div className="grid grid-cols-2 gap-3 text-sm sm:grid-cols-6">
                <div className="rounded-lg bg-brand-50 p-3"><p className="text-xs text-brand-700">الإجمالي</p><p className="font-bold text-brand-800">{summary.avgOverall.toFixed(2)} / 5</p></div>
                <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">عدد التقييمات</p><p className="font-bold">{summary.ratingsCount}</p></div>
                <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">الالتزام</p><p className="font-bold">{summary.avgCommitment.toFixed(2)}</p></div>
                <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">المحافظة</p><p className="font-bold">{summary.avgVehicleCare.toFixed(2)}</p></div>
                <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">القيادة</p><p className="font-bold">{summary.avgDriving.toFixed(2)}</p></div>
                <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">النظافة</p><p className="font-bold">{summary.avgCleanliness.toFixed(2)}</p></div>
              </div>
            )}
          </div>
        )}

        {/* السجل */}
        <div className="space-y-3">
          <div className="flex flex-wrap gap-2">
            <select value={filters.vehicleId} onChange={(e) => setFilters({ ...filters, vehicleId: e.target.value })} className={`${INPUT} text-sm`}>
              <option value="">كل السيارات</option>
              {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name}</option>)}
            </select>
            <select value={filters.driverId} onChange={(e) => setFilters({ ...filters, driverId: e.target.value })} className={`${INPUT} text-sm`}>
              <option value="">كل السائقين</option>
              {employees.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}
            </select>
            <select value={filters.status} onChange={(e) => setFilters({ ...filters, status: e.target.value as typeof filters.status })} className={`${INPUT} text-sm`}>
              <option value="">كل الحالات</option>
              <option value="IN_PROGRESS">جارية</option>
              <option value="COMPLETED">منتهية</option>
            </select>
            <input type="date" value={filters.from} onChange={(e) => setFilters({ ...filters, from: e.target.value })} className={`${INPUT} text-sm`} title="من" />
            <input type="date" value={filters.to} onChange={(e) => setFilters({ ...filters, to: e.target.value })} className={`${INPUT} text-sm`} title="إلى" />
          </div>

          <div className="overflow-x-auto rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <table className="w-full text-right text-sm">
              <thead className="bg-slate-50 text-slate-600">
                <tr>
                  <th className="px-4 py-2">السيارة</th>
                  <th className="px-4 py-2">السائق</th>
                  <th className="px-4 py-2">الوجهة / السبب</th>
                  <th className="px-4 py-2">الوقت</th>
                  <th className="px-4 py-2">المسافة</th>
                  <th className="px-4 py-2">التقييم</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {history === null && <tr><td colSpan={6} className="p-4 text-center text-slate-400">جاري التحميل...</td></tr>}
                {history?.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">ماكو مهمات</td></tr>}
                {history?.map((m) => (
                  <tr key={m.id}>
                    <td className="px-4 py-2 font-medium">{m.vehicle?.name ?? '-'}<div className="text-xs text-slate-400">{m.vehicle?.plateNumber}</div></td>
                    <td className="px-4 py-2">
                      {m.driver ? (
                        <button onClick={() => { setSummary(null); setSummaryFor(m.driver) }} className="font-medium text-brand-700 hover:underline" title="ملخص تقييم السائق">
                          {m.driver.name} ⭐
                        </button>
                      ) : '-'}
                      {(m.passengers ?? []).length > 0 && <div className="text-xs text-slate-400">+ {(m.passengers ?? []).map((p) => p.employee?.name).filter(Boolean).join('، ')}</div>}
                    </td>
                    <td className="px-4 py-2">{m.destination}<div className="text-xs text-slate-400">{m.purpose}</div></td>
                    <td className="px-4 py-2 text-xs text-slate-500">{fmt(m.startedAt)}{m.endedAt && <><br />← {fmt(m.endedAt)}</>}</td>
                    <td className="px-4 py-2">
                      {m.status === 'IN_PROGRESS' ? (
                        <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-700">جارية</span>
                      ) : (
                        <span>{m.distanceKm ?? '-'} كم</span>
                      )}
                    </td>
                    <td className="px-4 py-2">
                      {m.rating ? (
                        <span className="text-xs text-slate-600" title={m.rating.notes ?? ''}>
                          {((m.rating.commitment + m.rating.vehicleCare + m.rating.driving + m.rating.cleanliness) / 4).toFixed(1)} / 5
                          {m.rating.ratedBy && <span className="block text-slate-400">{m.rating.ratedBy.name}</span>}
                        </span>
                      ) : m.status === 'COMPLETED' ? (
                        <button onClick={() => { setRating(m); setRatingForm({ commitment: 5, vehicleCare: 5, driving: 5, cleanliness: 5, notes: '' }) }} className="rounded-lg bg-brand-50 px-2 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">⭐ تقييم</button>
                      ) : canEnd(m) ? (
                        <button onClick={() => { setEnding(m); setEndForm({ endOdometer: '', notes: '' }) }} className="rounded-lg bg-brand-600 px-2 py-1 text-xs font-bold text-white">🏁 إنهاء</button>
                      ) : '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {ending && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
              <h3 className="text-lg font-bold text-brand-900">إنهاء مهمة {ending.vehicle?.name}</h3>
              <p className="mt-1 text-sm text-slate-500">عداد البداية: {ending.startOdometer.toLocaleString()}</p>
              <input autoFocus type="number" min={ending.startOdometer} placeholder="عداد النهاية" value={endForm.endOdometer} onChange={(e) => setEndForm({ ...endForm, endOdometer: e.target.value })} className={`mt-4 w-full ${INPUT}`} />
              <textarea placeholder="ملاحظات (اختياري)" value={endForm.notes} onChange={(e) => setEndForm({ ...endForm, notes: e.target.value })} className={`mt-3 h-20 w-full ${INPUT}`} />
              <div className="mt-4 flex gap-3">
                <button onClick={end} disabled={endForm.endOdometer === '' || guard.busy} className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50">إنهاء المهمة</button>
                <button onClick={() => setEnding(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600">إلغاء</button>
              </div>
            </div>
          </div>
        )}

        {rating && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
              <h3 className="text-lg font-bold text-brand-900">تقييم {rating.driver?.name} — {rating.vehicle?.name}</h3>
              <div className="mt-4 space-y-3">
                {RATING_FIELDS.map((f) => (
                  <div key={f.key} className="flex items-center justify-between gap-3">
                    <span className="text-sm text-slate-600">{f.label}</span>
                    <div className="flex gap-1">
                      {[1, 2, 3, 4, 5].map((n) => (
                        <button type="button" key={n} onClick={() => setRatingForm({ ...ratingForm, [f.key]: n })} className={`h-8 w-8 rounded-lg text-sm font-bold ${ratingForm[f.key] >= n ? 'bg-amber-400 text-white' : 'bg-slate-100 text-slate-400'}`}>{n}</button>
                      ))}
                    </div>
                  </div>
                ))}
                <textarea placeholder="ملاحظات (اختياري)" value={ratingForm.notes} onChange={(e) => setRatingForm({ ...ratingForm, notes: e.target.value })} className={`h-20 w-full ${INPUT}`} />
              </div>
              <div className="mt-4 flex gap-3">
                <button onClick={submitRating} disabled={guard.busy} className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50">حفظ التقييم</button>
                <button onClick={() => setRating(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600">إلغاء</button>
              </div>
            </div>
          </div>
        )}

        {toolCheck && canToolCheck && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
              <h3 className="text-lg font-bold text-brand-900">🧰 فحص أدوات {toolCheck.vehicle?.name}</h3>
              <p className="mt-1 text-sm text-slate-500">أشّر الأدوات الناقصة — إذا كلشي موجود احفظ بلا تأشير</p>
              <div className="mt-4 max-h-72 space-y-1 overflow-auto">
                {tools === null ? (
                  <p className="text-sm text-slate-400">جاري التحميل...</p>
                ) : tools.length === 0 ? (
                  <p className="text-sm text-slate-400">ماكو أدوات مسجلة لهاي السيارة</p>
                ) : tools.map((t) => (
                  <label key={t.id} className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-slate-50">
                    <input
                      type="checkbox"
                      checked={missing.includes(t.name)}
                      onChange={(e) => setMissing((prev) => e.target.checked ? [...prev, t.name] : prev.filter((n) => n !== t.name))}
                    />
                    <span>{t.name}</span>
                    <span className="text-xs text-slate-400">× {t.quantity}</span>
                  </label>
                ))}
              </div>
              <div className="mt-4 flex gap-3">
                <button onClick={submitToolCheck} disabled={guard.busy || tools === null} className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50">
                  {missing.length ? `حفظ (${missing.length} ناقصة)` : 'حفظ — كلشي موجود'}
                </button>
                <button onClick={() => setToolCheck(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600">لاحقاً</button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
