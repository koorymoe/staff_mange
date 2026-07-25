import { useEffect, useState } from 'react'
import { api, type Vehicle, type VehicleLog, type VehicleIncident, type VehicleMonthlyStatus, type VehicleDailyRating, type Employee, type VehicleDocument, type VehiclePhoto } from '../api'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

const DOC_TYPE_LABELS: Record<string, string> = {
  INSURANCE: 'تأمين',
  ANNUAL_LICENSE: 'إجازة سنوية',
  INSPECTION: 'فحص دوري',
  OTHER: 'أخرى',
}

function docExpiryColor(expiryDate: string | null): string {
  if (!expiryDate) return ''
  const diffDays = Math.floor((new Date(expiryDate).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  if (diffDays < 0) return 'bg-red-100 text-red-700'
  if (diffDays <= 30) return 'bg-amber-100 text-amber-700'
  return 'bg-emerald-100 text-emerald-700'
}

const LOG_TYPE_LABELS: Record<string, string> = {
  FUEL: 'تعبئة وقود',
  CLEANING: 'تنظيف',
  OIL_CHANGE: 'تبديل زيت',
}

const RATING_FIELDS: { key: keyof typeof EMPTY_RATING_FORM; label: string }[] = [
  { key: 'wash', label: 'غسل السيارة' },
  { key: 'exteriorClean', label: 'نظافة الهيكل الخارجي' },
  { key: 'exteriorCondition', label: 'حالة الهيكل الخارجي' },
  { key: 'tireCondition', label: 'حالة الإطارات' },
  { key: 'glassClean', label: 'تنظيف الزجاج' },
  { key: 'lightsCondition', label: 'حالة اللايتات' },
  { key: 'technicalFaults', label: 'الأعطال الفنية' },
  { key: 'interiorClean', label: 'التنظيف الداخلي' },
  { key: 'seatsCondition', label: 'حالة الكراسي' },
  { key: 'interiorDirt', label: 'الأوساخ الداخلية' },
  { key: 'smell', label: 'الرائحة' },
]

const RATING_GRADE_LABELS: Record<number, string> = { 0: '0 - سيئ جداً', 1: '1 - ضعيف', 2: '2 - متوسط', 3: '3 - جيد', 4: '4 - ممتاز' }

function scoreGrade(score: number): { label: string; color: string } {
  if (score >= 0.85) return { label: 'ممتاز', color: 'bg-emerald-100 text-emerald-700' }
  if (score >= 0.7) return { label: 'جيد', color: 'bg-lime-100 text-lime-700' }
  if (score >= 0.5) return { label: 'مقبول', color: 'bg-amber-100 text-amber-700' }
  return { label: 'ضعيف', color: 'bg-red-100 text-red-700' }
}

const EMPTY_RATING_FORM = {
  wash: '' as number | '', exteriorClean: '' as number | '', exteriorCondition: '' as number | '',
  tireCondition: '' as number | '', glassClean: '' as number | '', lightsCondition: '' as number | '',
  technicalFaults: '' as number | '', interiorClean: '' as number | '', seatsCondition: '' as number | '',
  interiorDirt: '' as number | '', smell: '' as number | '',
}

const currentMonth = () => {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

export default function VehiclesPage() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [logs, setLogs] = useState<VehicleLog[]>([])
  const [incidents, setIncidents] = useState<VehicleIncident[]>([])
  const [monthlyStatus, setMonthlyStatus] = useState<VehicleMonthlyStatus[]>([])
  const [dailyRatings, setDailyRatings] = useState<VehicleDailyRating[]>([])
  const [documents, setDocuments] = useState<VehicleDocument[]>([])
  const [photos, setPhotos] = useState<VehiclePhoto[]>([])
  const [tab, setTab] = useState<'logs' | 'incidents' | 'monthly' | 'rating' | 'documents' | 'photos'>('logs')

  // add-vehicle form
  const [showAddVehicle, setShowAddVehicle] = useState(false)
  const [vName, setVName] = useState('')
  const [vPlate, setVPlate] = useState('')
  const [vColor, setVColor] = useState('')
  const [vType, setVType] = useState('')
  const [vModel, setVModel] = useState('')
  const [vYear, setVYear] = useState('')
  const [vChassis, setVChassis] = useState('')
  const [vEngine, setVEngine] = useState('')
  const [vFuel, setVFuel] = useState('')
  const [vOdometer, setVOdometer] = useState('')
  const [vCondition, setVCondition] = useState('')

  // document form
  const [docType, setDocType] = useState<'INSURANCE' | 'ANNUAL_LICENSE' | 'INSPECTION' | 'OTHER'>('INSURANCE')
  const [docNumber, setDocNumber] = useState('')
  const [docIssue, setDocIssue] = useState('')
  const [docExpiry, setDocExpiry] = useState('')
  const [docFile, setDocFile] = useState('')
  const [savingDoc, setSavingDoc] = useState(false)

  // photo upload
  const [uploadingPhoto, setUploadingPhoto] = useState(false)

  // log form
  const [logType, setLogType] = useState<'FUEL' | 'CLEANING' | 'OIL_CHANGE'>('FUEL')
  const [logOdometer, setLogOdometer] = useState('')
  const [logCost, setLogCost] = useState('')
  const [logNextDue, setLogNextDue] = useState('')
  const [logNotes, setLogNotes] = useState('')

  // incident form
  const [incType, setIncType] = useState<'FAULT' | 'DAMAGE'>('FAULT')
  const [incDesc, setIncDesc] = useState('')
  const [incResponsible, setIncResponsible] = useState('')
  const [incCost, setIncCost] = useState('')

  // monthly form
  const [monMonth, setMonMonth] = useState(currentMonth())
  const [monHasIssue, setMonHasIssue] = useState(false)
  const [monDesc, setMonDesc] = useState('')
  const [monResolved, setMonResolved] = useState(false)
  const [monNotes, setMonNotes] = useState('')

  // daily rating form
  const [ratingForm, setRatingForm] = useState(EMPTY_RATING_FORM)
  const [faultDesc, setFaultDesc] = useState('')
  const [ratingNotes, setRatingNotes] = useState('')
  const [washTechId, setWashTechId] = useState('')
  const [washScore, setWashScore] = useState<0 | 1 | 2 | ''>('')
  const [savingRating, setSavingRating] = useState(false)

  const loadVehicles = () => api.getVehicles().then(setVehicles)

  useEffect(() => {
    loadVehicles()
    api.getEmployees().then(setEmployees)
  }, [])

  useEffect(() => {
    if (!selectedId) return
    api.getVehicleLogs(selectedId).then(setLogs)
    api.getVehicleIncidents(selectedId).then(setIncidents)
    api.getVehicleMonthlyStatus(selectedId).then(setMonthlyStatus)
    api.getVehicleDailyRatings(selectedId).then(setDailyRatings)
    api.getVehicleDocuments(selectedId).then(setDocuments)
    api.getVehiclePhotos(selectedId).then(setPhotos)
  }, [selectedId])

  const selectedVehicle = vehicles.find((v) => v.id === selectedId) || null

  const handleAddVehicle = async (e: React.FormEvent) => {
    e.preventDefault()
    const created = await api.createVehicle({ name: vName, plateNumber: vPlate, color: vColor || undefined, type: vType || undefined })
    const extra: Record<string, string | number> = {}
    if (vModel) extra.model = vModel
    if (vYear) extra.year = Number(vYear)
    if (vChassis) extra.chassisNumber = vChassis
    if (vEngine) extra.engineNumber = vEngine
    if (vFuel) extra.fuelType = vFuel
    if (vOdometer) extra.currentOdometer = Number(vOdometer)
    if (vCondition) extra.condition = vCondition
    if (Object.keys(extra).length > 0) {
      await api.updateVehicle(created.id, extra)
    }
    setVName(''); setVPlate(''); setVColor(''); setVType('')
    setVModel(''); setVYear(''); setVChassis(''); setVEngine(''); setVFuel(''); setVOdometer(''); setVCondition('')
    setShowAddVehicle(false)
    loadVehicles()
  }

  const handleAddDocument = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    setSavingDoc(true)
    try {
      await api.createVehicleDocument(selectedId, {
        documentType: docType,
        documentNumber: docNumber || undefined,
        issueDate: docIssue || undefined,
        expiryDate: docExpiry || undefined,
        fileUrl: docFile || undefined,
      })
      setDocNumber(''); setDocIssue(''); setDocExpiry(''); setDocFile('')
      api.getVehicleDocuments(selectedId).then(setDocuments)
    } finally {
      setSavingDoc(false)
    }
  }

  const handleDeleteDocument = async (docId: string) => {
    if (!selectedId) return
    await api.deleteVehicleDocument(selectedId, docId)
    api.getVehicleDocuments(selectedId).then(setDocuments)
  }

  const handleUploadPhoto = async (file: File | null) => {
    if (!file || !selectedId) return
    setUploadingPhoto(true)
    try {
      const base64 = await fileToBase64(file)
      await api.createVehiclePhoto(selectedId, { url: base64 })
      api.getVehiclePhotos(selectedId).then(setPhotos)
    } finally {
      setUploadingPhoto(false)
    }
  }

  const handleDeletePhoto = async (photoId: string) => {
    if (!selectedId) return
    await api.deleteVehiclePhoto(selectedId, photoId)
    api.getVehiclePhotos(selectedId).then(setPhotos)
  }

  const handleAddLog = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    await api.createVehicleLog(selectedId, {
      type: logType,
      odometer: logOdometer ? Number(logOdometer) : undefined,
      cost: logCost ? Number(logCost) : undefined,
      nextDueAt: logNextDue || undefined,
      notes: logNotes || undefined,
    })
    setLogOdometer(''); setLogCost(''); setLogNextDue(''); setLogNotes('')
    api.getVehicleLogs(selectedId).then(setLogs)
  }

  const handleAddIncident = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId || !incDesc.trim()) return
    await api.createVehicleIncident(selectedId, {
      type: incType,
      description: incDesc,
      responsibleEmployeeId: incResponsible || undefined,
      cost: incCost ? Number(incCost) : undefined,
    })
    setIncDesc(''); setIncResponsible(''); setIncCost('')
    api.getVehicleIncidents(selectedId).then(setIncidents)
  }

  const handleResolveIncident = async (id: string) => {
    if (!selectedId) return
    await api.updateVehicleIncident(id, { status: 'RESOLVED' })
    api.getVehicleIncidents(selectedId).then(setIncidents)
  }

  const handleSetMonthly = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    await api.setVehicleMonthlyStatus(selectedId, {
      month: monMonth, hasIssue: monHasIssue, issueDescription: monDesc || undefined, resolved: monResolved, notes: monNotes || undefined,
    })
    setMonDesc(''); setMonNotes('')
    api.getVehicleMonthlyStatus(selectedId).then(setMonthlyStatus)
  }

  const handleAddRating = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    setSavingRating(true)
    try {
      await api.createVehicleDailyRating(selectedId, {
        wash: ratingForm.wash === '' ? undefined : ratingForm.wash,
        exteriorClean: ratingForm.exteriorClean === '' ? undefined : ratingForm.exteriorClean,
        exteriorCondition: ratingForm.exteriorCondition === '' ? undefined : ratingForm.exteriorCondition,
        tireCondition: ratingForm.tireCondition === '' ? undefined : ratingForm.tireCondition,
        glassClean: ratingForm.glassClean === '' ? undefined : ratingForm.glassClean,
        lightsCondition: ratingForm.lightsCondition === '' ? undefined : ratingForm.lightsCondition,
        technicalFaults: ratingForm.technicalFaults === '' ? undefined : ratingForm.technicalFaults,
        faultDescription: faultDesc || undefined,
        interiorClean: ratingForm.interiorClean === '' ? undefined : ratingForm.interiorClean,
        seatsCondition: ratingForm.seatsCondition === '' ? undefined : ratingForm.seatsCondition,
        interiorDirt: ratingForm.interiorDirt === '' ? undefined : ratingForm.interiorDirt,
        smell: ratingForm.smell === '' ? undefined : ratingForm.smell,
        notes: ratingNotes || undefined,
        technicianRatings: washTechId && washScore !== '' ? [{ employeeId: washTechId, score: washScore }] : undefined,
      })
      setRatingForm(EMPTY_RATING_FORM); setFaultDesc(''); setRatingNotes(''); setWashTechId(''); setWashScore('')
      api.getVehicleDailyRatings(selectedId).then(setDailyRatings)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر حفظ التقييم')
    } finally {
      setSavingRating(false)
    }
  }

  return (
    <div dir="rtl" className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">إدارة المركبات</h2>
          <p className="mt-1 text-slate-500">وقود، تنظيف، تبديل زيت، أعطال، أضرار، وحالة شهرية لكل سيارة</p>
        </div>
        <button
          onClick={() => setShowAddVehicle((v) => !v)}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2 font-medium text-white shadow-md"
        >
          {showAddVehicle ? 'إلغاء' : '+ إضافة سيارة'}
        </button>
      </div>

      {showAddVehicle && (
        <form onSubmit={handleAddVehicle} className="grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4">
          <input required placeholder="اسم السيارة" value={vName} onChange={(e) => setVName(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input required placeholder="رقم اللوحة" value={vPlate} onChange={(e) => setVPlate(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="اللون" value={vColor} onChange={(e) => setVColor(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="النوع" value={vType} onChange={(e) => setVType(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="الموديل" value={vModel} onChange={(e) => setVModel(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input type="number" placeholder="سنة الصنع" value={vYear} onChange={(e) => setVYear(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="رقم الشاصي" value={vChassis} onChange={(e) => setVChassis(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="رقم المحرك" value={vEngine} onChange={(e) => setVEngine(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input placeholder="نوع الوقود" value={vFuel} onChange={(e) => setVFuel(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <input type="number" placeholder="عداد الكيلومترات الحالي" value={vOdometer} onChange={(e) => setVOdometer(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
          <select value={vCondition} onChange={(e) => setVCondition(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
            <option value="">-- حالة السيارة --</option>
            <option value="ممتازة">ممتازة</option>
            <option value="جيدة">جيدة</option>
            <option value="تحتاج صيانة">تحتاج صيانة</option>
            <option value="معطلة">معطلة</option>
          </select>
          <button type="submit" className="sm:col-span-4 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">حفظ السيارة</button>
        </form>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Vehicle list */}
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-1">
          <div className="bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-3 text-white font-bold">السيارات ({vehicles.length})</div>
          <div className="divide-y divide-slate-100">
            {vehicles.map((v) => (
              <button
                key={v.id}
                onClick={() => setSelectedId(v.id)}
                className={`block w-full px-5 py-3 text-right transition-colors ${selectedId === v.id ? 'bg-brand-50' : 'hover:bg-slate-50'}`}
              >
                <p className="font-bold text-slate-800">{v.name}</p>
                <p className="text-xs text-slate-500">{v.plateNumber} {v.color ? `- ${v.color}` : ''}</p>
              </button>
            ))}
            {vehicles.length === 0 && <p className="p-6 text-center text-slate-400">لا توجد سيارات بعد</p>}
          </div>
        </div>

        {/* Detail panel */}
        <div className="lg:col-span-2">
          {!selectedVehicle ? (
            <div className="flex h-full items-center justify-center rounded-xl border border-white bg-white p-10 text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              اختر سيارة لعرض سجلاتها
            </div>
          ) : (
            <div className="space-y-4">
              <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <h3 className="text-lg font-bold text-brand-800">{selectedVehicle.name}</h3>
                <p className="text-sm text-slate-500">{selectedVehicle.plateNumber}</p>
                <div className="mt-2 grid grid-cols-2 gap-1 text-xs text-slate-500 sm:grid-cols-3">
                  {selectedVehicle.model && <p>الموديل: {selectedVehicle.model}</p>}
                  {selectedVehicle.year && <p>سنة الصنع: {selectedVehicle.year}</p>}
                  {selectedVehicle.chassisNumber && <p>رقم الشاصي: {selectedVehicle.chassisNumber}</p>}
                  {selectedVehicle.engineNumber && <p>رقم المحرك: {selectedVehicle.engineNumber}</p>}
                  {selectedVehicle.fuelType && <p>الوقود: {selectedVehicle.fuelType}</p>}
                  <p>عداد الكيلومترات: {selectedVehicle.currentOdometer}</p>
                  {selectedVehicle.condition && <p>الحالة: {selectedVehicle.condition}</p>}
                </div>
              </div>

              <div className="flex flex-wrap gap-2 rounded-xl border border-white bg-white p-2 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                {([
                  { key: 'logs', label: 'وقود / تنظيف / زيت' },
                  { key: 'incidents', label: 'أعطال وأضرار' },
                  { key: 'monthly', label: 'الحالة الشهرية' },
                  { key: 'rating', label: 'التقييم اليومي' },
                  { key: 'documents', label: 'الوثائق' },
                  { key: 'photos', label: 'الصور' },
                ] as const).map((t) => (
                  <button
                    key={t.key}
                    onClick={() => setTab(t.key)}
                    className={`flex-1 rounded-lg px-3 py-2 text-sm font-bold transition-colors ${tab === t.key ? 'bg-brand-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
                  >
                    {t.label}
                  </button>
                ))}
              </div>

              {tab === 'logs' && (
                <div className="space-y-4">
                  <form onSubmit={handleAddLog} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3">
                    <select value={logType} onChange={(e) => setLogType(e.target.value as typeof logType)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="FUEL">تعبئة وقود</option>
                      <option value="CLEANING">تنظيف</option>
                      <option value="OIL_CHANGE">تبديل زيت</option>
                    </select>
                    <input type="number" placeholder="عداد المسافة" value={logOdometer} onChange={(e) => setLogOdometer(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="التكلفة" value={logCost} onChange={(e) => setLogCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">الموعد القادم (اختياري)</label>
                      <input type="date" value={logNextDue} onChange={(e) => setLogNextDue(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <input placeholder="ملاحظات" value={logNotes} onChange={(e) => setLogNotes(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                    <button type="submit" className="sm:col-span-3 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">تسجيل</button>
                  </form>

                  <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <table className="w-full text-right text-sm">
                      <thead className="bg-slate-50 text-slate-600">
                        <tr>
                          <th className="px-4 py-2">النوع</th>
                          <th className="px-4 py-2">التاريخ</th>
                          <th className="px-4 py-2">العداد</th>
                          <th className="px-4 py-2">التكلفة</th>
                          <th className="px-4 py-2">الموعد القادم</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {logs.map((l) => (
                          <tr key={l.id}>
                            <td className="px-4 py-2 font-medium">{LOG_TYPE_LABELS[l.type]}</td>
                            <td className="px-4 py-2 text-slate-500">{new Date(l.performedAt).toLocaleDateString('ar-IQ')}</td>
                            <td className="px-4 py-2 text-slate-500">{l.odometer ?? '-'}</td>
                            <td className="px-4 py-2 text-slate-500">{l.cost ?? '-'}</td>
                            <td className="px-4 py-2 text-slate-500">{l.nextDueAt ? new Date(l.nextDueAt).toLocaleDateString('ar-IQ') : '-'}</td>
                          </tr>
                        ))}
                        {logs.length === 0 && <tr><td colSpan={5} className="p-4 text-center text-slate-400">لا توجد سجلات</td></tr>}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {tab === 'incidents' && (
                <div className="space-y-4">
                  <form onSubmit={handleAddIncident} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
                    <select value={incType} onChange={(e) => setIncType(e.target.value as typeof incType)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="FAULT">عطل</option>
                      <option value="DAMAGE">ضرر (صدمة)</option>
                    </select>
                    <select value={incResponsible} onChange={(e) => setIncResponsible(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="">-- المسبب (اختياري) --</option>
                      {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                    </select>
                    <input required placeholder="الوصف" value={incDesc} onChange={(e) => setIncDesc(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                    <input type="number" placeholder="التكلفة (اختياري)" value={incCost} onChange={(e) => setIncCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <button type="submit" className="rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">تسجيل</button>
                  </form>

                  <div className="space-y-3">
                    {incidents.map((inc) => (
                      <div key={inc.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        <div className="flex items-center justify-between">
                          <span className={`rounded-full px-2 py-1 text-xs font-bold ${inc.type === 'FAULT' ? 'bg-amber-100 text-amber-700' : 'bg-red-100 text-red-700'}`}>
                            {inc.type === 'FAULT' ? 'عطل' : 'ضرر'}
                          </span>
                          <span className={`rounded-full px-2 py-1 text-xs font-bold ${inc.status === 'RESOLVED' ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}>
                            {inc.status === 'RESOLVED' ? 'تمت المعالجة' : 'مفتوح'}
                          </span>
                        </div>
                        <p className="mt-2 text-sm text-slate-700">{inc.description}</p>
                        <div className="mt-1 grid grid-cols-2 gap-1 text-xs text-slate-500">
                          <p>المسبب: {inc.responsibleEmployee?.name || '-'}</p>
                          <p>التكلفة: {inc.cost ?? '-'}</p>
                        </div>
                        {inc.status === 'OPEN' && (
                          <button onClick={() => handleResolveIncident(inc.id)} className="mt-2 rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">
                            تمت المعالجة
                          </button>
                        )}
                      </div>
                    ))}
                    {incidents.length === 0 && <p className="p-4 text-center text-slate-400">لا توجد أعطال أو أضرار مسجلة</p>}
                  </div>
                </div>
              )}

              {tab === 'monthly' && (
                <div className="space-y-4">
                  <form onSubmit={handleSetMonthly} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
                    <input type="month" value={monMonth} onChange={(e) => setMonMonth(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <label className="flex items-center gap-2 text-sm">
                      <input type="checkbox" checked={monHasIssue} onChange={(e) => setMonHasIssue(e.target.checked)} />
                      فيها مشكلة هذا الشهر
                    </label>
                    <input placeholder="وصف المشكلة (إن وجدت)" value={monDesc} onChange={(e) => setMonDesc(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                    <label className="flex items-center gap-2 text-sm">
                      <input type="checkbox" checked={monResolved} onChange={(e) => setMonResolved(e.target.checked)} />
                      تمت المعالجة
                    </label>
                    <input placeholder="ملاحظات" value={monNotes} onChange={(e) => setMonNotes(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <button type="submit" className="sm:col-span-2 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">حفظ حالة الشهر</button>
                  </form>

                  <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <table className="w-full text-right text-sm">
                      <thead className="bg-slate-50 text-slate-600">
                        <tr>
                          <th className="px-4 py-2">الشهر</th>
                          <th className="px-4 py-2">فيها مشكلة</th>
                          <th className="px-4 py-2">الوصف</th>
                          <th className="px-4 py-2">الحالة</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {monthlyStatus.map((m) => (
                          <tr key={m.id}>
                            <td className="px-4 py-2 font-medium">{m.month}</td>
                            <td className="px-4 py-2">{m.hasIssue ? 'نعم' : 'لا'}</td>
                            <td className="px-4 py-2 text-slate-500">{m.issueDescription || '-'}</td>
                            <td className="px-4 py-2">
                              {m.hasIssue && (
                                <span className={`rounded-full px-2 py-1 text-xs font-bold ${m.resolved ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'}`}>
                                  {m.resolved ? 'انعالجت' : 'لم تُعالج'}
                                </span>
                              )}
                            </td>
                          </tr>
                        ))}
                        {monthlyStatus.length === 0 && <tr><td colSpan={4} className="p-4 text-center text-slate-400">لا توجد تقارير شهرية</td></tr>}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              {tab === 'rating' && (
                <div className="space-y-4">
                  <form onSubmit={handleAddRating} className="space-y-4 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <p className="text-sm text-slate-500">قيّم كل بند من 0 (سيئ جداً) إلى 4 (ممتاز) — تكدر تترك أي بند فاضي إذا ما تكدر تقيّمه اليوم.</p>
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                      {RATING_FIELDS.map((f) => (
                        <div key={f.key}>
                          <label className="mb-1 block text-xs font-medium text-slate-500">{f.label}</label>
                          <select
                            value={ratingForm[f.key]}
                            onChange={(e) => setRatingForm((prev) => ({ ...prev, [f.key]: e.target.value === '' ? '' : Number(e.target.value) }))}
                            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                          >
                            <option value="">-- بدون تقييم --</option>
                            {[0, 1, 2, 3, 4].map((n) => <option key={n} value={n}>{RATING_GRADE_LABELS[n]}</option>)}
                          </select>
                        </div>
                      ))}
                    </div>
                    <input placeholder="وصف العطل (إذا اكو عطل فني)" value={faultDesc} onChange={(e) => setFaultDesc(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                    <input placeholder="ملاحظات إضافية" value={ratingNotes} onChange={(e) => setRatingNotes(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />

                    <div className="rounded-xl border border-slate-100 bg-slate-50/50 p-4">
                      <p className="mb-2 text-sm font-bold text-slate-700">تقييم جودة الغسيل (اختياري) — مين غسل السيارة اليوم؟</p>
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                        <select value={washTechId} onChange={(e) => setWashTechId(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                          <option value="">-- اختر الفني --</option>
                          {employees.filter((e) => e.role === 'TECHNICIAN').map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                        </select>
                        <select value={washScore} onChange={(e) => setWashScore(e.target.value === '' ? '' : (Number(e.target.value) as 0 | 1 | 2))} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                          <option value="">-- تقييم الغسيل --</option>
                          <option value="0">0 - لم يغسل</option>
                          <option value="1">1 - غسل غير جيد</option>
                          <option value="2">2 - غسل جيد</option>
                        </select>
                      </div>
                    </div>

                    <button type="submit" disabled={savingRating} className="w-full rounded-lg bg-brand-600 px-5 py-2 font-medium text-white disabled:opacity-50">
                      {savingRating ? 'جاري الحفظ...' : 'حفظ تقييم اليوم'}
                    </button>
                  </form>

                  <div className="space-y-3">
                    {dailyRatings.map((r) => {
                      const grade = r.weightedScore !== null ? scoreGrade(r.weightedScore) : null
                      return (
                        <div key={r.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                          <div className="flex items-center justify-between">
                            <span className="text-sm font-bold text-slate-700">{new Date(r.ratedDate).toLocaleDateString('ar-IQ')}</span>
                            {grade && (
                              <span className={`rounded-full px-3 py-1 text-xs font-bold ${grade.color}`}>
                                {grade.label} — {Math.round(r.weightedScore! * 100)}%
                              </span>
                            )}
                          </div>
                          {r.faultDescription && <p className="mt-2 text-sm text-red-700">عطل: {r.faultDescription}</p>}
                          {r.notes && <p className="mt-1 text-sm text-slate-500">{r.notes}</p>}
                          {r.washRatings.length > 0 && (
                            <div className="mt-2 flex flex-wrap gap-2">
                              {r.washRatings.map((wr) => (
                                <span key={wr.id} className="rounded-full bg-sky-50 px-3 py-1 text-xs font-bold text-sky-700">
                                  {wr.employee?.name || '-'}: غسيل {wr.score}/2
                                </span>
                              ))}
                            </div>
                          )}
                        </div>
                      )
                    })}
                    {dailyRatings.length === 0 && <p className="p-4 text-center text-slate-400">لا توجد تقييمات مسجلة لهذي السيارة</p>}
                  </div>
                </div>
              )}

              {tab === 'documents' && (
                <div className="space-y-4">
                  <form onSubmit={handleAddDocument} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3">
                    <select value={docType} onChange={(e) => setDocType(e.target.value as typeof docType)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="INSURANCE">تأمين</option>
                      <option value="ANNUAL_LICENSE">إجازة سنوية</option>
                      <option value="INSPECTION">فحص دوري</option>
                      <option value="OTHER">أخرى</option>
                    </select>
                    <input placeholder="رقم الوثيقة" value={docNumber} onChange={(e) => setDocNumber(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">تاريخ الإصدار</label>
                      <input type="date" value={docIssue} onChange={(e) => setDocIssue(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">تاريخ الانتهاء</label>
                      <input type="date" value={docExpiry} onChange={(e) => setDocExpiry(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <div className="sm:col-span-2">
                      <label className="mb-1 block text-xs text-slate-500">ملف الوثيقة (اختياري)</label>
                      <input type="file" accept="image/*,.pdf" onChange={(e) => {
                        const f = e.target.files?.[0] || null
                        if (f) fileToBase64(f).then(setDocFile)
                      }} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                    </div>
                    <button type="submit" disabled={savingDoc} className="sm:col-span-3 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white disabled:opacity-50">
                      {savingDoc ? 'جاري الحفظ...' : 'إضافة وثيقة'}
                    </button>
                  </form>

                  <div className="space-y-3">
                    {documents.map((d) => (
                      <div key={d.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-bold text-slate-700">{DOC_TYPE_LABELS[d.documentType] || d.documentType}</span>
                          <div className="flex items-center gap-2">
                            {d.expiryDate && (
                              <span className={`rounded-full px-2 py-1 text-xs font-bold ${docExpiryColor(d.expiryDate)}`}>
                                ينتهي: {new Date(d.expiryDate).toLocaleDateString('ar-IQ')}
                              </span>
                            )}
                            <button onClick={() => handleDeleteDocument(d.id)} className="rounded-lg bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">حذف</button>
                          </div>
                        </div>
                        <div className="mt-1 grid grid-cols-2 gap-1 text-xs text-slate-500">
                          <p>الرقم: {d.documentNumber || '-'}</p>
                          <p>الإصدار: {d.issueDate ? new Date(d.issueDate).toLocaleDateString('ar-IQ') : '-'}</p>
                        </div>
                        {d.fileUrl && (
                          <a href={d.fileUrl} target="_blank" rel="noreferrer" className="mt-2 inline-block text-xs font-bold text-brand-600 hover:underline">عرض الملف</a>
                        )}
                      </div>
                    ))}
                    {documents.length === 0 && <p className="p-4 text-center text-slate-400">لا توجد وثائق مسجلة لهذي السيارة</p>}
                  </div>
                </div>
              )}

              {tab === 'photos' && (
                <div className="space-y-4">
                  <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <label className="mb-1 block text-xs text-slate-500">رفع صورة جديدة</label>
                    <input
                      type="file"
                      accept="image/*"
                      disabled={uploadingPhoto}
                      onChange={(e) => handleUploadPhoto(e.target.files?.[0] || null)}
                      className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                    />
                  </div>
                  <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
                    {photos.map((p) => (
                      <div key={p.id} className="group relative overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        <a href={p.url} target="_blank" rel="noreferrer">
                          <img src={p.url} alt="صورة السيارة" className="h-32 w-full object-cover" />
                        </a>
                        <button
                          onClick={() => handleDeletePhoto(p.id)}
                          className="absolute left-1 top-1 rounded-lg bg-red-600/90 px-2 py-1 text-xs font-bold text-white opacity-0 transition-opacity group-hover:opacity-100"
                        >
                          حذف
                        </button>
                      </div>
                    ))}
                    {photos.length === 0 && <p className="col-span-full p-4 text-center text-slate-400">لا توجد صور مسجلة لهذي السيارة</p>}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
