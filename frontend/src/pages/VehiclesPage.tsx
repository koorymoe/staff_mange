import { useEffect, useState, useMemo } from 'react'
import { api, type Vehicle, type VehicleLog, type VehicleIncident, type VehicleMonthlyStatus, type VehicleDailyRating, type Employee, type VehicleIncidentAttachment, type VehicleAlert, type VehicleExpenseSummary, type EmployeeFuelStat } from '../api'
import { useSession } from '../session'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

const LOG_TYPE_LABELS: Record<string, string> = {
  FUEL: 'تعبئة وقود',
  CLEANING: 'تنظيف',
  OIL_CHANGE: 'تبديل زيت',
  MAINTENANCE: 'صيانة عامة',
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
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [logs, setLogs] = useState<VehicleLog[]>([])
  const [incidents, setIncidents] = useState<VehicleIncident[]>([])
  const [monthlyStatus, setMonthlyStatus] = useState<VehicleMonthlyStatus[]>([])
  const [dailyRatings, setDailyRatings] = useState<VehicleDailyRating[]>([])
  const [incidentAttachments, setIncidentAttachments] = useState<Record<string, VehicleIncidentAttachment[]>>({})
  const [alerts, setAlerts] = useState<VehicleAlert[]>([])
  const [alertsOpen, setAlertsOpen] = useState(false)
  const [tab, setTab] = useState<'logs' | 'oil' | 'incidents' | 'monthly' | 'rating'>('logs')

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



  // log form
  const [logType, setLogType] = useState<'FUEL' | 'CLEANING' | 'OIL_CHANGE' | 'MAINTENANCE'>('FUEL')
  // كلفة التنظيف — تنكتب بتبويب التقييم وتنحفظ كسجل CLEANING
  const [cleaningCost, setCleaningCost] = useState('')
  // كل تبويب يعرض سجلاته هو فقط: الوقود بتبويبه، والزيت/الصيانة بتبويب الدهن.
  // سجلات التنظيف تظهر بتبويب الدهن كمان حتى ما تختفي من أي مكان.
  const visibleLogs = useMemo(
    () => logs.filter((l) => (tab === 'logs' ? l.type === 'FUEL' : l.type !== 'FUEL')),
    [logs, tab],
  )
  const [logOdometer, setLogOdometer] = useState('')
  const [logCost, setLogCost] = useState('')
  const [logNextDue, setLogNextDue] = useState('')
  const [logNextDueOdometer, setLogNextDueOdometer] = useState('')
  const [logNotes, setLogNotes] = useState('')
  // تفاصيل تعبئة الوقود
  const [logLiters, setLogLiters] = useState('')
  const [logFilledBy, setLogFilledBy] = useState('')
  const [logReceiptNo, setLogReceiptNo] = useState('')
  const [logStation, setLogStation] = useState('')
  const [logReceiptPhoto, setLogReceiptPhoto] = useState('')
  // تعديل سجل موجود
  const [editLog, setEditLog] = useState<VehicleLog | null>(null)
  const [editForm, setEditForm] = useState({ odometer: '', cost: '', liters: '', filledBy: '', receiptNo: '', station: '', notes: '' })
  const [savingLog, setSavingLog] = useState(false)
  // صورة الوصل المعروضة بنافذة (تنجلب عند الطلب فقط)
  const [receiptPhotoView, setReceiptPhotoView] = useState<string | null>(null)
  // إحصائية "منو عبّأ وكم مرة" بالشهر
  const [fuelStats, setFuelStats] = useState<EmployeeFuelStat[]>([])

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
  }, [selectedId])

  useEffect(() => {
    if (!selectedId) return
    api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
    api.getEmployeeFuelStats({ vehicleId: selectedId, month: expenseMonth }).then(setFuelStats).catch(() => setFuelStats([]))
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

  const handleAddLog = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedId) return
    setFuelAnomalyWarning(null)
    const result = await api.createVehicleLog(selectedId, {
      // تبويب الوقود ما بيه اختيار نوع — النوع ثابت FUEL، وتبويب الدهن
      // يستخدم النوع المختار (زيت أو صيانة عامة).
      type: tab === 'logs' ? 'FUEL' : logType,
      odometer: logOdometer ? Number(logOdometer) : undefined,
      cost: logCost ? Number(logCost) : undefined,
      // موعد الصيانة القادمة خاص بتبويب الدهن فقط
      nextDueAt: tab === 'oil' ? (logNextDue || undefined) : undefined,
      nextDueOdometer: tab === 'oil' && logNextDueOdometer ? Number(logNextDueOdometer) : undefined,
      notes: logNotes || undefined,
      // تفاصيل الوقود
      liters: tab === 'logs' && logLiters ? Number(logLiters) : undefined,
      filledByEmployeeId: tab === 'logs' ? (logFilledBy || undefined) : undefined,
      receiptNumber: tab === 'logs' ? (logReceiptNo.trim() || undefined) : undefined,
      stationName: tab === 'logs' ? (logStation.trim() || undefined) : undefined,
      receiptPhotoBase64: tab === 'logs' ? (logReceiptPhoto || undefined) : undefined,
    })
    if (result.fuelAnomaly?.isAnomaly) {
      setFuelAnomalyWarning(
        `⚠️ هذا المبلغ أعلى من المعدل المعتاد لهذي السيارة بـ ${Math.round(result.fuelAnomaly.percentAboveAvg)}%`
      )
    }
    setLogOdometer(''); setLogCost(''); setLogNextDue(''); setLogNextDueOdometer(''); setLogNotes('')
    setLogLiters(''); setLogFilledBy(''); setLogReceiptNo(''); setLogStation(''); setLogReceiptPhoto('')
    api.getVehicleLogs(selectedId).then(setLogs)
    api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
    api.getEmployeeFuelStats({ vehicleId: selectedId, month: expenseMonth }).then(setFuelStats).catch(() => setFuelStats([]))
    loadAlerts()
  }

  const openEditLog = (l: VehicleLog) => {
    setEditLog(l)
    setEditForm({
      odometer: l.odometer != null ? String(l.odometer) : '',
      cost: l.cost != null ? String(l.cost) : '',
      liters: l.liters != null ? String(l.liters) : '',
      filledBy: l.filledByEmployeeId || '',
      receiptNo: l.receiptNumber || '',
      station: l.stationName || '',
      notes: l.notes || '',
    })
  }

  const saveEditLog = async () => {
    if (!editLog || !selectedId) return
    setSavingLog(true)
    try {
      await api.updateVehicleLog(selectedId, editLog.id, {
        odometer: editForm.odometer ? Number(editForm.odometer) : undefined,
        cost: editForm.cost ? Number(editForm.cost) : undefined,
        liters: editForm.liters ? Number(editForm.liters) : undefined,
        filledByEmployeeId: editForm.filledBy || undefined,
        receiptNumber: editForm.receiptNo.trim() || undefined,
        stationName: editForm.station.trim() || undefined,
        notes: editForm.notes.trim() || undefined,
      })
      setEditLog(null)
      api.getVehicleLogs(selectedId).then(setLogs)
      api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
      api.getEmployeeFuelStats({ vehicleId: selectedId, month: expenseMonth }).then(setFuelStats).catch(() => setFuelStats([]))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر حفظ التعديل')
    } finally {
      setSavingLog(false)
    }
  }

  const handleDeleteLog = async (id: string) => {
    if (!selectedId || !confirm('حذف هذا السجل نهائياً؟')) return
    try {
      await api.deleteVehicleLog(selectedId, id)
      api.getVehicleLogs(selectedId).then(setLogs)
      api.getVehicleExpenseSummary(selectedId, { month: expenseMonth }).then(setExpenseSummary)
      api.getEmployeeFuelStats({ vehicleId: selectedId, month: expenseMonth }).then(setFuelStats).catch(() => setFuelStats([]))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر الحذف')
    }
  }

  const viewReceiptPhoto = async (l: VehicleLog) => {
    if (!selectedId) return
    try {
      const res = await api.getVehicleLogReceiptPhoto(selectedId, l.id)
      if (res.receiptPhotoBase64) setReceiptPhotoView(res.receiptPhotoBase64)
      else alert('ما اكو صورة وصل لهذا السجل')
    } catch {
      alert('تعذر جلب صورة الوصل')
    }
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
      // كلفة التنظيف تنسجل كسجل CLEANING حتى تدخل بملخص مصاريف السيارة
      if (cleaningCost.trim() && Number(cleaningCost) > 0) {
        await api.createVehicleLog(selectedId, {
          type: 'CLEANING',
          cost: Number(cleaningCost),
          notes: 'تنظيف مسجّل مع تقييم اليوم',
        })
        api.getVehicleLogs(selectedId).then(setLogs)
      }
      setRatingForm(EMPTY_RATING_FORM); setFaultDesc(''); setRatingNotes(''); setCleaningCost('')
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
          <p className="mt-1 text-slate-500">وقود، دهن، أعطال، أضرار، وتقييم يومي لكل سيارة</p>
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
                {/* الوقود انفصل عن الزيت (كانوا تبويب واحد)، والتنظيف انتقل
                    للتقييم اليومي. الوثائق والصور والإطارات/البطاريات انشالت
                    من الواجهة — مو مطلوبة بهذي الصفحة. */}
                {([
                  { key: 'logs', label: 'الوقود' },
                  { key: 'oil', label: 'الدهن (تبديل زيت)' },
                  { key: 'incidents', label: 'أعطال وأضرار' },
                  { key: 'monthly', label: 'الحالة الشهرية' },
                  { key: 'rating', label: 'التقييم اليومي والتنظيف' },
                ] as const).map((t) => (
                  <button
                    key={t.key}
                    onClick={() => { setTab(t.key); if (t.key === 'oil') setLogType('OIL_CHANGE') }}
                    className={`flex-1 rounded-lg px-3 py-2 text-sm font-bold transition-colors ${tab === t.key ? 'bg-brand-600 text-white' : 'text-slate-600 hover:bg-slate-50'}`}
                  >
                    {t.label}
                  </button>
                ))}
              </div>

              {(tab === 'logs' || tab === 'oil') && (
                <div className="space-y-4">
                  <form onSubmit={handleAddLog} className="grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-3">
                    {/* الوقود تبويب لحاله، والدهن تبويبه لحاله — نوع السجل
                        ينتحدد من التبويب بدل قائمة منسدلة تخلط الاثنين. */}
                    {tab === 'logs' ? (
                      <div className="rounded-lg bg-brand-50 px-3 py-2 text-sm font-bold text-brand-700">⛽ تعبئة وقود</div>
                    ) : (
                      <select value={logType} onChange={(e) => setLogType(e.target.value as typeof logType)} className="rounded-lg border border-slate-300 px-3 py-2">
                        <option value="OIL_CHANGE">تبديل زيت (دهن)</option>
                        <option value="MAINTENANCE">صيانة عامة</option>
                      </select>
                    )}
                    <input type="number" placeholder="عداد المسافة" value={logOdometer} onChange={(e) => setLogOdometer(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />
                    <input type="number" placeholder="التكلفة" value={logCost} onChange={(e) => setLogCost(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2" />

                    {/* موعد الصيانة القادمة مالو محل بتعبئة الوقود — يبقى
                        بتبويب الدهن حيث الصيانة الفعلية. */}
                    {tab === 'oil' ? (
                      <>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">تاريخ الصيانة القادمة (اختياري)</label>
                          <input type="date" value={logNextDue} onChange={(e) => setLogNextDue(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">العداد القادم للصيانة (اختياري)</label>
                          <input type="number" placeholder="العداد القادم للصيانة" value={logNextDueOdometer} onChange={(e) => setLogNextDueOdometer(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                        </div>
                      </>
                    ) : (
                      <>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">عدد اللترات</label>
                          <input type="number" step="0.01" min="0" placeholder="مثال: 45.5" value={logLiters} onChange={(e) => setLogLiters(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">منو عبّأ؟</label>
                          <select value={logFilledBy} onChange={(e) => setLogFilledBy(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2">
                            <option value="">-- اختر الموظف --</option>
                            {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                          </select>
                        </div>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">رقم الوصل</label>
                          <input placeholder="رقم الوصل" value={logReceiptNo} onChange={(e) => setLogReceiptNo(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                        </div>
                        <div>
                          <label className="mb-1 block text-xs text-slate-500">اسم المحطة</label>
                          <input placeholder="اسم المحطة" value={logStation} onChange={(e) => setLogStation(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                        </div>
                        <div className="sm:col-span-3">
                          <label className="mb-1 block text-xs text-slate-500">صورة الوصل (اختياري)</label>
                          <input
                            type="file" accept="image/*"
                            onChange={async (e) => {
                              const f = e.target.files?.[0]
                              setLogReceiptPhoto(f ? await fileToBase64(f) : '')
                            }}
                            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                          />
                          {logReceiptPhoto && <p className="mt-1 text-xs font-bold text-emerald-700">✓ انرفعت صورة الوصل</p>}
                        </div>
                      </>
                    )}
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

                  {/* منو عبّأ وكم مرة بهذا الشهر — الشهر نفسه مال ملخص المصاريف */}
                  {tab === 'logs' && (
                    <div className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                      <h4 className="mb-3 font-bold text-slate-700">تعبئة الوقود حسب الموظف ({expenseMonth})</h4>
                      {fuelStats.length === 0 ? (
                        <p className="text-sm text-slate-400">ما اكو تعبئة مسجلة بهذا الشهر</p>
                      ) : (
                        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                          {fuelStats.map((s) => (
                            <div key={s.employeeId} className="rounded-xl border border-slate-100 bg-slate-50/60 p-4">
                              <div className="font-bold text-brand-900">{s.employeeName}</div>
                              <div className="mt-1 text-sm text-slate-600">
                                عبّأ <span className="font-bold text-brand-700">{s.fillCount}</span> مرة
                              </div>
                              <div className="mt-0.5 text-xs text-slate-400">
                                {s.totalLiters.toLocaleString()} لتر · {s.totalCost.toLocaleString()} د.ع
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <table className="w-full text-right text-sm">
                      <thead className="bg-slate-50 text-slate-600">
                        <tr>
                          <th className="px-4 py-2">النوع</th>
                          <th className="px-4 py-2">التاريخ</th>
                          <th className="px-4 py-2">العداد</th>
                          <th className="px-4 py-2">التكلفة</th>
                          {tab === 'logs' ? (
                            <>
                              <th className="px-4 py-2">اللترات</th>
                              <th className="px-4 py-2">منو عبّأ</th>
                              <th className="px-4 py-2">الوصل / المحطة</th>
                            </>
                          ) : (
                            <>
                              <th className="px-4 py-2">الموعد القادم</th>
                              <th className="px-4 py-2">العداد القادم</th>
                            </>
                          )}
                          <th className="px-4 py-2">إجراءات</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {visibleLogs.map((l) => (
                          <tr key={l.id}>
                            <td className="px-4 py-2 font-medium">{LOG_TYPE_LABELS[l.type]}</td>
                            <td className="px-4 py-2 text-slate-500">{new Date(l.performedAt).toLocaleDateString('ar-IQ')}</td>
                            <td className="px-4 py-2 text-slate-500">{l.odometer ?? '-'}</td>
                            <td className="px-4 py-2 text-slate-500">{l.cost ?? '-'}</td>
                            {tab === 'logs' ? (
                              <>
                                <td className="px-4 py-2 font-bold text-brand-700">{l.liters ?? '-'}</td>
                                <td className="px-4 py-2 text-slate-600">{l.filledByName || '-'}</td>
                                <td className="px-4 py-2 text-slate-500">
                                  {l.receiptNumber && <div>وصل: {l.receiptNumber}</div>}
                                  {l.stationName && <div className="text-xs text-slate-400">{l.stationName}</div>}
                                  {l.hasReceiptPhoto && (
                                    <button onClick={() => viewReceiptPhoto(l)} className="mt-0.5 text-xs font-bold text-brand-600 hover:underline">
                                      📷 شوف صورة الوصل
                                    </button>
                                  )}
                                  {!l.receiptNumber && !l.stationName && !l.hasReceiptPhoto && '-'}
                                </td>
                              </>
                            ) : (
                              <>
                                <td className="px-4 py-2 text-slate-500">{l.nextDueAt ? new Date(l.nextDueAt).toLocaleDateString('ar-IQ') : '-'}</td>
                                <td className="px-4 py-2 text-slate-500">{l.nextDueOdometer ?? '-'}</td>
                              </>
                            )}
                            <td className="px-4 py-2">
                              <div className="flex gap-1">
                                <button onClick={() => openEditLog(l)} className="rounded-lg bg-brand-50 px-2 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">✎ تعديل</button>
                                {isAdmin && (
                                  <button onClick={() => handleDeleteLog(l.id)} className="rounded-lg bg-red-50 px-2 py-1 text-xs font-bold text-red-600 hover:bg-red-100">🗑</button>
                                )}
                              </div>
                            </td>
                          </tr>
                        ))}
                        {visibleLogs.length === 0 && <tr><td colSpan={tab === 'logs' ? 8 : 7} className="p-4 text-center text-slate-400">لا توجد سجلات</td></tr>}
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

                    {/* التنظيف انتقل هنا من تبويب الوقود — التقييم هو محله
                        الطبيعي لأنه مربوط بتقييم جودة الغسيل فوق. */}
                    <div className="rounded-xl border border-slate-100 bg-slate-50/50 p-4">
                      <p className="mb-2 text-sm font-bold text-slate-700">تكلفة التنظيف (اختياري) — تنسجل بمصاريف السيارة</p>
                      <input
                        type="number" min="0"
                        placeholder="كلفة التنظيف بالدينار"
                        value={cleaningCost}
                        onChange={(e) => setCleaningCost(e.target.value)}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
                      />
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

            </div>
          )}
        </div>
      </div>

      {editLog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-brand-900">تعديل سجل — {LOG_TYPE_LABELS[editLog.type]}</h3>
            <p className="mt-1 text-sm text-slate-400">{new Date(editLog.performedAt).toLocaleDateString('ar-IQ')}</p>

            <div className="mt-4 grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs text-slate-500">عداد المسافة</label>
                <input type="number" value={editForm.odometer} onChange={(e) => setEditForm({ ...editForm, odometer: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
              </div>
              <div>
                <label className="mb-1 block text-xs text-slate-500">التكلفة</label>
                <input type="number" value={editForm.cost} onChange={(e) => setEditForm({ ...editForm, cost: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
              </div>
              {editLog.type === 'FUEL' && (
                <>
                  <div>
                    <label className="mb-1 block text-xs text-slate-500">عدد اللترات</label>
                    <input type="number" step="0.01" value={editForm.liters} onChange={(e) => setEditForm({ ...editForm, liters: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-slate-500">منو عبّأ</label>
                    <select value={editForm.filledBy} onChange={(e) => setEditForm({ ...editForm, filledBy: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2">
                      <option value="">-- اختر الموظف --</option>
                      {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-slate-500">رقم الوصل</label>
                    <input value={editForm.receiptNo} onChange={(e) => setEditForm({ ...editForm, receiptNo: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-slate-500">اسم المحطة</label>
                    <input value={editForm.station} onChange={(e) => setEditForm({ ...editForm, station: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
                  </div>
                </>
              )}
              <div className="col-span-2">
                <label className="mb-1 block text-xs text-slate-500">ملاحظات</label>
                <input value={editForm.notes} onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })} className="w-full rounded-lg border border-slate-300 px-3 py-2" />
              </div>
            </div>

            <div className="mt-5 flex gap-3">
              <button onClick={saveEditLog} disabled={savingLog} className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50">
                {savingLog ? 'جاري الحفظ...' : 'حفظ التعديل'}
              </button>
              <button onClick={() => setEditLog(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600 hover:bg-slate-50">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {receiptPhotoView && (
        <div onClick={() => setReceiptPhotoView(null)} className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="max-h-full max-w-3xl overflow-auto rounded-2xl bg-white p-3">
            <img src={receiptPhotoView} alt="صورة الوصل" className="max-w-full rounded-lg" />
            <button onClick={() => setReceiptPhotoView(null)} className="mt-3 w-full rounded-xl border border-slate-300 py-2 font-medium text-slate-600">إغلاق</button>
          </div>
        </div>
      )}
    </div>
  )
}
