import { useEffect, useState } from 'react'
import { api, type Vehicle, type VehicleLog, type VehicleIncident, type VehicleMonthlyStatus, type VehicleDailyRating, type Employee, type VehicleDocument, type VehiclePhoto, type VehicleIncidentAttachment, type VehiclePart, type VehicleAlert, type VehicleExpenseSummary } from '../api'
import { useSession } from '../session'

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
  MAINTENANCE: 'صيانة عامة',
}

const PART_TYPE_LABELS: Record<string, string> = {
  TIRE: 'إطار',
  BATTERY: 'بطارية',
}

const ALERT_TYPE_LABELS: Record<string, string> = {
  MAINTENANCE: 'صيانة',
  PART: 'قطعة',
  DOCUMENT: 'وثيقة',
  FUEL_ANOMALY: 'شذوذ وقود',
}

const INCIDENT_TYPE_LABELS: Record<string, string> = {
  FAULT: 'عطل',
  DAMAGE: 'ضرر',
  ACCIDENT: 'حادث',
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
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  // Snapshot "now" once per mount so "days since install" is pure during render;
  // doesn't need live ticking since it's a day-granularity display, not a countdown.
  const [now] = useState(() => Date.now())
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [logs, setLogs] = useState<VehicleLog[]>([])
  const [incidents, setIncidents] = useState<VehicleIncident[]>([])
  const [monthlyStatus, setMonthlyStatus] = useState<VehicleMonthlyStatus[]>([])
  const [dailyRatings, setDailyRatings] = useState<VehicleDailyRating[]>([])
  const [documents, setDocuments] = useState<VehicleDocument[]>([])
  const [photos, setPhotos] = useState<VehiclePhoto[]>([])
  const [parts, setParts] = useState<VehiclePart[]>([])
  const [incidentAttachments, setIncidentAttachments] = useState<Record<string, VehicleIncidentAttachment[]>>({})
  const [alerts, setAlerts] = useState<VehicleAlert[]>([])
  const [alertsOpen, setAlertsOpen] = useState(false)
  const [tab, setTab] = useState<'logs' | 'incidents' | 'monthly' | 'rating' | 'documents' | 'photos' | 'parts'>('logs')

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
  const [logType, setLogType] = useState<'FUEL' | 'CLEANING' | 'OIL_CHANGE' | 'MAINTENANCE'>('FUEL')
  const [logOdometer, setLogOdometer] = useState('')
  const [logCost, setLogCost] = useState('')
  const [logNextDue, setLogNextDue] = useState('')
  const [logNextDueOdometer, setLogNextDueOdometer] = useState('')
  const [logNotes, setLogNotes] = useState('')

  // incident form
  const [incType, setIncType] = useState<'FAULT' | 'DAMAGE' | 'ACCIDENT'>('FAULT')
  const [incDesc, setIncDesc] = useState('')
  const [incResponsible, setIncResponsible] = useState('')
  const [incCost, setIncCost] = useState('')
  const [incLocation, setIncLocation] = useState('')
  const [incDriver, setIncDriver] = useState('')
  const [incPeoplePresent, setIncPeoplePresent] = useState('')
  const [incPoliceReport, setIncPoliceReport] = useState('')
  const [incRepairCost, setIncRepairCost] = useState('')
  const [uploadingAttachmentFor, setUploadingAttachmentFor] = useState<string | null>(null)

  // expense summary
  const [expenseMonth, setExpenseMonth] = useState(currentMonth())
  const [expenseSummary, setExpenseSummary] = useState<VehicleExpenseSummary | null>(null)
  const [fuelAnomalyWarning, setFuelAnomalyWarning] = useState<string | null>(null)

  // part form
  const [partType, setPartType] = useState<'TIRE' | 'BATTERY'>('TIRE')
  const [partInstalledAt, setPartInstalledAt] = useState('')
  const [partInstalledOdometer, setPartInstalledOdometer] = useState('')
  const [partLifespanKm, setPartLifespanKm] = useState('')
  const [partLifespanMonths, setPartLifespanMonths] = useState('')
  const [partNotes, setPartNotes] = useState('')
  const [partCost, setPartCost] = useState('')
  const [savingPart, setSavingPart] = useState(false)

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
  const [washTechId2, setWashTechId2] = useState('')
  const [washScore2, setWashScore2] = useState<0 | 1 | 2 | ''>('')
  const [savingRating, setSavingRating] = useState(false)

  // غسيل السيارات فنيّين عاديين بس — لا تيم ليدر، لا مصمم، لا مسؤول خدمة، لا إداري.
  const washableTechnicians = employees.filter((e) => e.role === 'TECHNICIAN' && !e.isLeader)

  const loadVehicles = () => api.getVehicles().then(setVehicles)

  const loadAlerts = () => api.getVehicleAlerts().then(setAlerts)

  useEffect(() => {
    loadVehicles()
    api.getEmployees().then(setEmployees)
    loadAlerts()
  }, [])

  const loadIncidentsWithAttachments = (vehicleId: string) => {
    api.getVehicleIncidents(vehicleId).then((list) => {
      setIncidents(list)
      list.forEach((inc) => {
        api.getVehicleIncidentAttachments(inc.id).then((atts) =>
          setIncidentAttachments((prev) => ({ ...prev, [inc.id]: atts }))
        )
      })
    })
  }

  useEffect(() => {
    if (!selectedId) return
    api.getVehicleLogs(selectedId).then(setLogs)
    loadIncidentsWithAttachments(selectedId)
    api.getVehicleMonthlyStatus(selectedId).then(setMonthlyStatus)
    api.getVehicleDailyRatings(selectedId).then(setDailyRatings)
    api.getVehicleDocuments(selectedId).then(setDocuments)
    api.getVehiclePhotos(selectedId).then(setPhotos)
    api.getVehicleParts(selectedId).then(setParts)
  }, [selectedId])

  useEffect(() => {
    if (!selectedId) return
    api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
  }, [selectedId, expenseMonth])

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

  const [showEditVehicle, setShowEditVehicle] = useState(false)
  const [editDraft, setEditDraft] = useState<Partial<Vehicle>>({})
  const [deleting, setDeleting] = useState(false)

  const openEditVehicle = () => {
    if (!selectedVehicle) return
    setEditDraft({ ...selectedVehicle })
    setShowEditVehicle(true)
  }

  const handleUpdateVehicle = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    await api.updateVehicle(selectedId, {
      name: editDraft.name,
      plateNumber: editDraft.plateNumber,
      color: editDraft.color || undefined,
      type: editDraft.type || undefined,
      model: editDraft.model || undefined,
      year: editDraft.year || undefined,
      chassisNumber: editDraft.chassisNumber || undefined,
      engineNumber: editDraft.engineNumber || undefined,
      fuelType: editDraft.fuelType || undefined,
      currentOdometer: editDraft.currentOdometer,
      condition: editDraft.condition || undefined,
    })
    setShowEditVehicle(false)
    loadVehicles()
  }

  const handleDeleteVehicle = async () => {
    if (!selectedId || !selectedVehicle) return
    if (!confirm(`حذف السيارة "${selectedVehicle.name}" (${selectedVehicle.plateNumber}) نهائياً؟ هذا الإجراء لا يمكن التراجع عنه.`)) return
    setDeleting(true)
    try {
      await api.deleteVehicle(selectedId)
      setSelectedId(null)
      loadVehicles()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setDeleting(false)
    }
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
    setFuelAnomalyWarning(null)
    const result = await api.createVehicleLog(selectedId, {
      type: logType,
      odometer: logOdometer ? Number(logOdometer) : undefined,
      cost: logCost ? Number(logCost) : undefined,
      nextDueAt: logNextDue || undefined,
      nextDueOdometer: logNextDueOdometer ? Number(logNextDueOdometer) : undefined,
      notes: logNotes || undefined,
    })
    if (result.fuelAnomaly?.isAnomaly) {
      setFuelAnomalyWarning(
        `⚠️ هذا المبلغ أعلى من المعدل المعتاد لهذي السيارة بـ ${Math.round(result.fuelAnomaly.percentAboveAvg)}%`
      )
    }
    setLogOdometer(''); setLogCost(''); setLogNextDue(''); setLogNextDueOdometer(''); setLogNotes('')
    api.getVehicleLogs(selectedId).then(setLogs)
    api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
    loadAlerts()
  }

  const handleAddIncident = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId || !incDesc.trim()) return
    await api.createVehicleIncident(selectedId, {
      type: incType,
      description: incDesc,
      responsibleEmployeeId: incResponsible || undefined,
      cost: incCost ? Number(incCost) : undefined,
      location: incType === 'ACCIDENT' ? (incLocation || undefined) : undefined,
      driverId: incType === 'ACCIDENT' ? (incDriver || undefined) : undefined,
      peoplePresent: incType === 'ACCIDENT' ? (incPeoplePresent || undefined) : undefined,
      policeReportNumber: incType === 'ACCIDENT' ? (incPoliceReport || undefined) : undefined,
      repairCost: incType === 'ACCIDENT' && incRepairCost ? Number(incRepairCost) : undefined,
    })
    setIncDesc(''); setIncResponsible(''); setIncCost('')
    setIncLocation(''); setIncDriver(''); setIncPeoplePresent(''); setIncPoliceReport(''); setIncRepairCost('')
    loadIncidentsWithAttachments(selectedId)
    if (selectedId) api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
  }

  const handleResolveIncident = async (id: string) => {
    if (!selectedId) return
    await api.updateVehicleIncident(id, { status: 'RESOLVED' })
    loadIncidentsWithAttachments(selectedId)
  }

  const handleUploadIncidentAttachment = async (incidentId: string, file: File | null) => {
    if (!file) return
    setUploadingAttachmentFor(incidentId)
    try {
      const base64 = await fileToBase64(file)
      const mediaType = file.type.startsWith('video/') ? 'VIDEO' : 'IMAGE'
      await api.createVehicleIncidentAttachment(incidentId, { url: base64, mediaType })
      const atts = await api.getVehicleIncidentAttachments(incidentId)
      setIncidentAttachments((prev) => ({ ...prev, [incidentId]: atts }))
    } finally {
      setUploadingAttachmentFor(null)
    }
  }

  const handleDeleteIncidentAttachment = async (incidentId: string, attachmentId: string) => {
    await api.deleteVehicleIncidentAttachment(incidentId, attachmentId)
    const atts = await api.getVehicleIncidentAttachments(incidentId)
    setIncidentAttachments((prev) => ({ ...prev, [incidentId]: atts }))
  }

  const handleAddPart = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId || !partInstalledOdometer) return
    setSavingPart(true)
    try {
      await api.createVehiclePart(selectedId, {
        partType,
        installedAt: partInstalledAt || undefined,
        installedOdometer: Number(partInstalledOdometer),
        expectedLifespanKm: partLifespanKm ? Number(partLifespanKm) : undefined,
        expectedLifespanMonths: partLifespanMonths ? Number(partLifespanMonths) : undefined,
        notes: partNotes || undefined,
        cost: partCost ? Number(partCost) : undefined,
      })
      setPartInstalledAt(''); setPartInstalledOdometer(''); setPartLifespanKm(''); setPartLifespanMonths(''); setPartNotes(''); setPartCost('')
      api.getVehicleParts(selectedId).then(setParts)
      loadAlerts()
    } finally {
      setSavingPart(false)
    }
  }

  const handleReplacePart = async (partId: string, type: 'TIRE' | 'BATTERY') => {
    if (!selectedId) return
    await api.replaceVehiclePart(partId)
    api.getVehicleParts(selectedId).then(setParts)
    loadAlerts()
    // تحضير نموذج إضافة قطعة جديدة بنفس النوع كتذكير للمراقب لتسجيل التبديل فوراً
    setPartType(type)
    setPartInstalledAt(new Date().toISOString().slice(0, 10))
    setPartInstalledOdometer(String(selectedVehicle?.currentOdometer ?? ''))
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
        technicianRatings: (() => {
          const ratings = [
            washTechId && washScore !== '' ? { employeeId: washTechId, score: washScore } : null,
            washTechId2 && washScore2 !== '' ? { employeeId: washTechId2, score: washScore2 } : null,
          ].filter((r): r is { employeeId: string; score: 0 | 1 | 2 } => r !== null)
          return ratings.length > 0 ? ratings : undefined
        })(),
      })
      setRatingForm(EMPTY_RATING_FORM); setFaultDesc(''); setRatingNotes('')
      setWashTechId(''); setWashScore(''); setWashTechId2(''); setWashScore2('')
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

      {alerts.length > 0 && (
        <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <button
            onClick={() => setAlertsOpen((v) => !v)}
            className="flex w-full items-center justify-between px-5 py-3"
          >
            <span className="flex items-center gap-2 font-bold text-slate-800">
              تنبيهات الصيانة والوثائق
              <span className={`rounded-full px-2 py-0.5 text-xs font-bold text-white ${alerts.some((a) => a.severity === 'danger') ? 'bg-red-600' : 'bg-amber-500'}`}>
                {alerts.length}
              </span>
            </span>
            <span className="text-slate-400">{alertsOpen ? '▲' : '▼'}</span>
          </button>
          {alertsOpen && (
            <div className="divide-y divide-slate-100 border-t border-slate-100">
              {alerts.map((a, i) => (
                <div key={i} className="flex items-center justify-between px-5 py-2 text-sm">
                  <div>
                    <span className="font-bold text-slate-700">{a.vehicleName}</span>
                    <span className="mx-2 text-slate-400">·</span>
                    <span className="text-slate-500">{ALERT_TYPE_LABELS[a.alertType] || a.alertType}: {a.message}</span>
                  </div>
                  <span className={`rounded-full px-2 py-1 text-xs font-bold ${a.severity === 'danger' ? 'bg-red-100 text-red-700' : 'bg-amber-100 text-amber-700'}`}>
                    {a.severity === 'danger' ? 'عاجل' : 'قريب'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

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
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="text-lg font-bold text-brand-800">{selectedVehicle.name}</h3>
                    <p className="text-sm text-slate-500">{selectedVehicle.plateNumber}</p>
                  </div>
                  {isAdmin && (
                    <div className="flex shrink-0 gap-2">
                      <button
                        onClick={openEditVehicle}
                        className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-100"
                      >
                        تعديل
                      </button>
                      <button
                        onClick={handleDeleteVehicle}
                        disabled={deleting}
                        className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100 disabled:opacity-50"
                      >
                        {deleting ? 'جارٍ الحذف...' : 'حذف'}
                      </button>
                    </div>
                  )}
                </div>
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

              {showEditVehicle && isAdmin && (
                <form onSubmit={handleUpdateVehicle} className="grid grid-cols-1 gap-3 rounded-xl border border-brand-200 bg-brand-50/40 p-5 sm:grid-cols-3">
                  <input required placeholder="اسم السيارة" value={editDraft.name || ''} onChange={(e) => setEditDraft((p) => ({ ...p, name: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input required placeholder="رقم اللوحة" value={editDraft.plateNumber || ''} onChange={(e) => setEditDraft((p) => ({ ...p, plateNumber: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="اللون" value={editDraft.color || ''} onChange={(e) => setEditDraft((p) => ({ ...p, color: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="النوع" value={editDraft.type || ''} onChange={(e) => setEditDraft((p) => ({ ...p, type: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="الموديل" value={editDraft.model || ''} onChange={(e) => setEditDraft((p) => ({ ...p, model: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="سنة الصنع" type="number" value={editDraft.year || ''} onChange={(e) => setEditDraft((p) => ({ ...p, year: e.target.value ? Number(e.target.value) : undefined }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="رقم الشاصي" value={editDraft.chassisNumber || ''} onChange={(e) => setEditDraft((p) => ({ ...p, chassisNumber: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="رقم المحرك" value={editDraft.engineNumber || ''} onChange={(e) => setEditDraft((p) => ({ ...p, engineNumber: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="نوع الوقود" value={editDraft.fuelType || ''} onChange={(e) => setEditDraft((p) => ({ ...p, fuelType: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="عداد الكيلومترات" type="number" value={editDraft.currentOdometer ?? ''} onChange={(e) => setEditDraft((p) => ({ ...p, currentOdometer: e.target.value ? Number(e.target.value) : undefined }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <input placeholder="الحالة" value={editDraft.condition || ''} onChange={(e) => setEditDraft((p) => ({ ...p, condition: e.target.value }))} className="rounded-lg border border-slate-300 px-3 py-2" />
                  <div className="flex gap-2 sm:col-span-3">
                    <button type="submit" className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white hover:bg-brand-700">حفظ التعديلات</button>
                    <button type="button" onClick={() => setShowEditVehicle(false)} className="rounded-lg bg-slate-100 px-4 py-2 text-sm font-bold text-slate-600 hover:bg-slate-200">إلغاء</button>
                  </div>
                </form>
              )}

              <div className="flex flex-wrap gap-2 rounded-xl border border-white bg-white p-2 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                {([
                  { key: 'logs', label: 'وقود / تنظيف / زيت' },
                  { key: 'incidents', label: 'أعطال وأضرار' },
                  { key: 'monthly', label: 'الحالة الشهرية' },
                  { key: 'rating', label: 'التقييم اليومي' },
                  { key: 'documents', label: 'الوثائق' },
                  { key: 'photos', label: 'الصور' },
                  { key: 'parts', label: 'الإطارات والبطاريات' },
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
                      <option value="MAINTENANCE">صيانة عامة</option>
                    </select>
                    <input type="number" placeholder="عداد المسافة" value={logOdometer} onChange={(e) => setLogOdometer(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="التكلفة" value={logCost} onChange={(e) => setLogCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">تاريخ الصيانة القادمة (اختياري)</label>
                      <input type="date" value={logNextDue} onChange={(e) => setLogNextDue(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">العداد القادم للصيانة (اختياري)</label>
                      <input type="number" placeholder="العداد القادم للصيانة" value={logNextDueOdometer} onChange={(e) => setLogNextDueOdometer(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <input placeholder="ملاحظات" value={logNotes} onChange={(e) => setLogNotes(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                    <button type="submit" className="sm:col-span-3 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">تسجيل</button>
                    {fuelAnomalyWarning && (
                      <p className="sm:col-span-3 rounded-lg bg-amber-50 px-3 py-2 text-sm font-bold text-amber-700">{fuelAnomalyWarning}</p>
                    )}
                  </form>

                  <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <div className="mb-3 flex items-center justify-between">
                      <h4 className="font-bold text-slate-700">ملخص مصاريف السيارة</h4>
                      <input type="month" value={expenseMonth} onChange={(e) => setExpenseMonth(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1 text-sm" />
                    </div>
                    {expenseSummary ? (
                      <div className="grid grid-cols-2 gap-3 text-sm sm:grid-cols-3">
                        <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">وقود</p><p className="font-bold text-slate-800">{expenseSummary.fuelCost.toLocaleString()}</p></div>
                        <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">صيانة</p><p className="font-bold text-slate-800">{expenseSummary.maintenanceCost.toLocaleString()}</p></div>
                        <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">قطع (إطارات/بطاريات)</p><p className="font-bold text-slate-800">{expenseSummary.partsCost.toLocaleString()}</p></div>
                        <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">حوادث/أعطال</p><p className="font-bold text-slate-800">{expenseSummary.incidentCost.toLocaleString()}</p></div>
                        <div className="rounded-lg bg-slate-50 p-3"><p className="text-xs text-slate-500">تنظيف</p><p className="font-bold text-slate-800">{expenseSummary.cleaningCost.toLocaleString()}</p></div>
                        <div className="rounded-lg bg-brand-50 p-3"><p className="text-xs text-brand-700">الإجمالي</p><p className="font-bold text-brand-800">{expenseSummary.totalCost.toLocaleString()}</p></div>
                        {expenseSummary.avgCostPerKm != null && (
                          <div className="rounded-lg bg-slate-50 p-3 sm:col-span-3"><p className="text-xs text-slate-500">متوسط التكلفة لكل كم ({expenseSummary.distanceKm} كم بهذي الفترة)</p><p className="font-bold text-slate-800">{expenseSummary.avgCostPerKm.toFixed(1)}</p></div>
                        )}
                      </div>
                    ) : (
                      <p className="text-sm text-slate-400">جاري التحميل...</p>
                    )}
                  </div>

                  <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <table className="w-full text-right text-sm">
                      <thead className="bg-slate-50 text-slate-600">
                        <tr>
                          <th className="px-4 py-2">النوع</th>
                          <th className="px-4 py-2">التاريخ</th>
                          <th className="px-4 py-2">العداد</th>
                          <th className="px-4 py-2">التكلفة</th>
                          <th className="px-4 py-2">الموعد القادم</th>
                          <th className="px-4 py-2">العداد القادم</th>
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
                            <td className="px-4 py-2 text-slate-500">{l.nextDueOdometer ?? '-'}</td>
                          </tr>
                        ))}
                        {logs.length === 0 && <tr><td colSpan={6} className="p-4 text-center text-slate-400">لا توجد سجلات</td></tr>}
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
                      <option value="ACCIDENT">حادث</option>
                    </select>
                    <select value={incResponsible} onChange={(e) => setIncResponsible(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="">-- المسبب (اختياري) --</option>
                      {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                    </select>
                    <input required placeholder="الوصف" value={incDesc} onChange={(e) => setIncDesc(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                    <input type="number" placeholder="التكلفة (اختياري)" value={incCost} onChange={(e) => setIncCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    {incType === 'ACCIDENT' && (
                      <>
                        <input placeholder="موقع الحادث" value={incLocation} onChange={(e) => setIncLocation(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                        <select value={incDriver} onChange={(e) => setIncDriver(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2">
                          <option value="">-- السائق وقت الحادث (اختياري) --</option>
                          {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                        </select>
                        <input placeholder="الأشخاص الموجودين" value={incPeoplePresent} onChange={(e) => setIncPeoplePresent(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 sm:col-span-2" />
                        <input placeholder="رقم تقرير الشرطة (اختياري)" value={incPoliceReport} onChange={(e) => setIncPoliceReport(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                        <input type="number" placeholder="تكلفة الإصلاح (اختياري)" value={incRepairCost} onChange={(e) => setIncRepairCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                      </>
                    )}
                    <button type="submit" className="rounded-lg bg-brand-600 px-5 py-2 font-medium text-white">تسجيل</button>
                  </form>

                  <div className="space-y-3">
                    {incidents.map((inc) => (
                      <div key={inc.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        <div className="flex items-center justify-between">
                          <span className={`rounded-full px-2 py-1 text-xs font-bold ${inc.type === 'FAULT' ? 'bg-amber-100 text-amber-700' : inc.type === 'ACCIDENT' ? 'bg-purple-100 text-purple-700' : 'bg-red-100 text-red-700'}`}>
                            {INCIDENT_TYPE_LABELS[inc.type] || inc.type}
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
                        {inc.type === 'ACCIDENT' && (
                          <div className="mt-2 grid grid-cols-2 gap-1 rounded-lg bg-purple-50/50 p-2 text-xs text-slate-600">
                            {inc.location && <p>موقع الحادث: {inc.location}</p>}
                            {inc.driver && <p>السائق: {inc.driver.name}</p>}
                            {inc.peoplePresent && <p className="col-span-2">الأشخاص الموجودين: {inc.peoplePresent}</p>}
                            {inc.policeReportNumber && <p>رقم تقرير الشرطة: {inc.policeReportNumber}</p>}
                            {inc.repairCost != null && <p>تكلفة الإصلاح: {inc.repairCost}</p>}
                          </div>
                        )}
                        {inc.status === 'OPEN' && (
                          <button onClick={() => handleResolveIncident(inc.id)} className="mt-2 rounded-lg bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">
                            تمت المعالجة
                          </button>
                        )}

                        <div className="mt-3 border-t border-slate-100 pt-3">
                          <label className="mb-1 block text-xs text-slate-500">إرفاق صورة/فيديو للعطل أو الضرر</label>
                          <input
                            type="file"
                            accept="image/*,video/*"
                            disabled={uploadingAttachmentFor === inc.id}
                            onChange={(e) => handleUploadIncidentAttachment(inc.id, e.target.files?.[0] || null)}
                            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-xs"
                          />
                          {(incidentAttachments[inc.id]?.length ?? 0) > 0 && (
                            <div className="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-4">
                              {incidentAttachments[inc.id].map((att) => (
                                <div key={att.id} className="group relative overflow-hidden rounded-lg border border-slate-200">
                                  {att.mediaType === 'VIDEO' ? (
                                    <video src={att.url} className="h-20 w-full object-cover" muted />
                                  ) : (
                                    <a href={att.url} target="_blank" rel="noreferrer">
                                      <img src={att.url} alt="مرفق" className="h-20 w-full object-cover" />
                                    </a>
                                  )}
                                  <button
                                    onClick={() => handleDeleteIncidentAttachment(inc.id, att.id)}
                                    className="absolute left-1 top-1 rounded-lg bg-red-600/90 px-1.5 py-0.5 text-[10px] font-bold text-white opacity-0 transition-opacity group-hover:opacity-100"
                                  >
                                    حذف
                                  </button>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
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
                      <p className="mb-2 text-sm font-bold text-slate-700">تقييم جودة الغسيل (اختياري) — مين غسل السيارة اليوم؟ (يقدر يكونون فنيّين اثنين بنفس الوقت)</p>
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                        <select value={washTechId} onChange={(e) => setWashTechId(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                          <option value="">-- اختر الفني الأول --</option>
                          {washableTechnicians.filter((e) => e.id !== washTechId2).map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                        </select>
                        <select value={washScore} onChange={(e) => setWashScore(e.target.value === '' ? '' : (Number(e.target.value) as 0 | 1 | 2))} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                          <option value="">-- تقييم غسيل الفني الأول --</option>
                          <option value="0">0 - لم يغسل</option>
                          <option value="1">1 - غسل غير جيد</option>
                          <option value="2">2 - غسل جيد</option>
                        </select>
                        <select value={washTechId2} onChange={(e) => setWashTechId2(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                          <option value="">-- اختر الفني الثاني (اختياري) --</option>
                          {washableTechnicians.filter((e) => e.id !== washTechId).map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                        </select>
                        <select value={washScore2} onChange={(e) => setWashScore2(e.target.value === '' ? '' : (Number(e.target.value) as 0 | 1 | 2))} disabled={!washTechId2} className="rounded-lg border border-slate-300 px-3 py-2 text-sm disabled:opacity-50">
                          <option value="">-- تقييم غسيل الفني الثاني --</option>
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

              {tab === 'parts' && (
                <div className="space-y-4">
                  <form onSubmit={handleAddPart} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3">
                    <select value={partType} onChange={(e) => setPartType(e.target.value as typeof partType)} className="rounded-lg border border-slate-300 px-3 py-2">
                      <option value="TIRE">إطار</option>
                      <option value="BATTERY">بطارية</option>
                    </select>
                    <div>
                      <label className="mb-1 block text-xs text-slate-500">تاريخ التركيب</label>
                      <input type="date" value={partInstalledAt} onChange={(e) => setPartInstalledAt(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                    </div>
                    <input required type="number" placeholder="عداد التركيب" value={partInstalledOdometer} onChange={(e) => setPartInstalledOdometer(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="العمر المتوقع (كم)" value={partLifespanKm} onChange={(e) => setPartLifespanKm(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="العمر المتوقع (شهور)" value={partLifespanMonths} onChange={(e) => setPartLifespanMonths(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input placeholder="ملاحظات" value={partNotes} onChange={(e) => setPartNotes(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="التكلفة (اختياري)" value={partCost} onChange={(e) => setPartCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <button type="submit" disabled={savingPart} className="sm:col-span-3 rounded-lg bg-brand-600 px-5 py-2 font-medium text-white disabled:opacity-50">
                      {savingPart ? 'جاري الحفظ...' : 'إضافة قطعة'}
                    </button>
                  </form>

                  <div className="space-y-3">
                    {parts.map((p) => {
                      const distanceSince = selectedVehicle ? selectedVehicle.currentOdometer - p.installedOdometer : null
                      const daysSince = Math.floor((now - new Date(p.installedAt).getTime()) / (1000 * 60 * 60 * 24))
                      return (
                        <div key={p.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-slate-700">{PART_TYPE_LABELS[p.partType]}</span>
                            <div className="flex items-center gap-2">
                              {p.replacedAt ? (
                                <span className="rounded-full bg-slate-100 px-2 py-1 text-xs font-bold text-slate-500">تم الاستبدال</span>
                              ) : p.dueSoon ? (
                                <span className="rounded-full bg-red-100 px-2 py-1 text-xs font-bold text-red-700">قريب الاستحقاق</span>
                              ) : (
                                <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-bold text-emerald-700">سليم</span>
                              )}
                              {!p.replacedAt && (
                                <button onClick={() => handleReplacePart(p.id, p.partType)} className="rounded-lg bg-amber-50 px-2 py-1 text-xs font-bold text-amber-700 hover:bg-amber-100">
                                  تم الاستبدال
                                </button>
                              )}
                            </div>
                          </div>
                          <div className="mt-2 grid grid-cols-2 gap-1 text-xs text-slate-500 sm:grid-cols-4">
                            <p>تاريخ التركيب: {new Date(p.installedAt).toLocaleDateString('ar-IQ')}</p>
                            <p>عداد التركيب: {p.installedOdometer}</p>
                            <p>المسافة منذ التركيب: {distanceSince ?? '-'} كم</p>
                            <p>الزمن منذ التركيب: {daysSince} يوم</p>
                            {p.expectedLifespanKm != null && <p>العمر المتوقع: {p.expectedLifespanKm} كم</p>}
                            {p.expectedLifespanMonths != null && <p>العمر المتوقع: {p.expectedLifespanMonths} شهر</p>}
                            {p.notes && <p className="sm:col-span-2">ملاحظات: {p.notes}</p>}
                          </div>
                        </div>
                      )
                    })}
                    {parts.length === 0 && <p className="p-4 text-center text-slate-400">لا توجد إطارات أو بطاريات مسجلة لهذي السيارة</p>}
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
