import { useEffect, useState } from 'react'
import ProjectPaymentsModal from '../../components/ProjectPaymentsModal'
import { api, fileUrl, storeFile, type ProjectMoney, type ProjectPaymentsState } from '../../api'
import { useSession, isAdminOrOwner } from '../../session'
import LocationFields from '../../components/LocationFields'
import { matches } from '../../utils/search'
import SurveyPhotos from '../../components/SurveyPhotos'
import { askConfirm } from '../../ui/dialogs'
import { swallow } from '../../utils/swallow'
import { dayBaghdad } from '../../utils/format'
import { request, type Project, type ProjectCandidate, candidateLabel, useProjectCandidates, useProjectWorkTypes, SURVEY, REJECTION_REASONS, CATEGORIES, resetMediaOpenCache, type OpenBrief, localDT, useCanUploadProjectSurvey } from './shared'
import { EmployeeOptions, Modal, Field } from './ui'

export function ContractModal({ project, onClose, onSaved }: { project: Project; onClose: () => void; onSaved: () => void }) {
  const { employee } = useSession()
  const isAdmin = isAdminOrOwner(employee)
  const [saving, setSaving] = useState<'plain' | 'signed' | null>(null)
  const [error, setError] = useState('')
  // ملفات العقد ما تجي بالقائمة (ثقيلة) — نجيبها هنا بس لما تنفتح النافذة.
  const [files, setFiles] = useState<{ contractPdfBase64: string | null; signedContractPdfBase64: string | null } | null>(null)
  useEffect(() => {
    request<{ contractPdfBase64: string | null; signedContractPdfBase64: string | null }>(`/projects/${project.id}`)
      .then(setFiles)
      .catch(() => setFiles({ contractPdfBase64: null, signedContractPdfBase64: null }))
  }, [project.id])

  const upload = async (file: File | undefined, field: 'contractPdfBase64' | 'signedContractPdfBase64') => {
    if (!file) return
    if (file.type !== 'application/pdf') { setError('لازم يكون الملف بصيغة PDF'); return }
    setError('')
    setSaving(field === 'contractPdfBase64' ? 'plain' : 'signed')
    try {
      const stored = await storeFile(file, 'projects')
      await request(`/projects/${project.id}`, { method: 'PUT', body: JSON.stringify({ [field]: stored }) })
      onSaved()
    } catch (e) {
      setError((e as Error).message || 'تعذر رفع الملف')
    } finally {
      setSaving(null)
    }
  }

  // حذف العقد المرفوع — لمدير النظام حصراً (والسيرفر يفرضها كمان).
  const removeContract = async (which: 'plain' | 'signed') => {
    if (!(await askConfirm('حذف ملف العقد هذا نهائياً؟'))) return
    setError('')
    try {
      await request(`/projects/${project.id}/contract?which=${which}`, { method: 'DELETE' })
      setFiles(f => f ? { ...f, [which === 'plain' ? 'contractPdfBase64' : 'signedContractPdfBase64']: null } : f)
      onSaved()
    } catch (e) {
      setError((e as Error).message || 'تعذر حذف الملف')
    }
  }

  return (
    <Modal onClose={onClose} title={`عقد المشروع: ${project.name}`}>
      <div className="space-y-4">
        <div>
          <label className="text-xs font-bold text-gray-600">العقد قبل التوقيع (PDF)</label>
          <input
            type="file" accept="application/pdf"
            onChange={(e) => upload(e.target.files?.[0], 'contractPdfBase64')}
            className="mt-1 block w-full text-sm"
          />
          {files?.contractPdfBase64 && (
            <div className="mt-1 flex items-center gap-3">
              <a href={fileUrl(files.contractPdfBase64)} download={`عقد-${project.code}.pdf`} className="text-xs text-brand-600 hover:underline">
                📄 تحميل الملف المرفوع حالياً
              </a>
              {isAdmin && (
                <button type="button" onClick={() => removeContract('plain')} className="text-xs font-bold text-red-600 hover:underline">
                  حذف الملف
                </button>
              )}
            </div>
          )}
          {saving === 'plain' && <p className="text-xs text-gray-400">جاري الرفع...</p>}
        </div>
        <div>
          <label className="text-xs font-bold text-gray-600">العقد بعد التوقيع (PDF)</label>
          <input
            type="file" accept="application/pdf"
            onChange={(e) => upload(e.target.files?.[0], 'signedContractPdfBase64')}
            className="mt-1 block w-full text-sm"
          />
          {files?.signedContractPdfBase64 && (
            <div className="mt-1 flex items-center gap-3">
              <a href={fileUrl(files.signedContractPdfBase64)} download={`عقد-موقّع-${project.code}.pdf`} className="text-xs text-brand-600 hover:underline">
                📄 تحميل الملف الموقّع حالياً
              </a>
              {isAdmin && (
                <button type="button" onClick={() => removeContract('signed')} className="text-xs font-bold text-red-600 hover:underline">
                  حذف الملف
                </button>
              )}
            </div>
          )}
          {saving === 'signed' && <p className="text-xs text-gray-400">جاري الرفع...</p>}
        </div>
        {error && <p className="text-sm font-bold text-red-600">{error}</p>}
      </div>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Add project modal
// ---------------------------------------------------------------------------
export function AddModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const workTypes = useProjectWorkTypes()
  const [form, setForm] = useState({ name: '', rep: '', phone: '', location: '', workType: 'طاقة شمسية', refPerson: '', priority: 'عادي', deliveryDate: '' })
  const [mapPoint, setMapPoint] = useState<{ lat: number; lng: number } | null>(null)
  const [locationUrl, setLocationUrl] = useState('')
  const [saving, setSaving] = useState(false)
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const save = async () => {
    if (!form.name.trim()) { alert('اسم المؤسسة مطلوب'); return }
    setSaving(true)
    try {
      await request('/projects', {
        method: 'POST',
        body: JSON.stringify({ ...form, mapLatitude: mapPoint?.lat, mapLongitude: mapPoint?.lng, locationUrl: locationUrl || undefined }),
      })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title="إضافة مشروع جديد" wide>
      <div className="grid sm:grid-cols-2 gap-3">
        <Field label="المؤسسة *"><input className="inp" value={form.name} onChange={e => set('name', e.target.value)} /></Field>
        <Field label="الممثل"><input className="inp" value={form.rep} onChange={e => set('rep', e.target.value)} /></Field>
        <Field label="الهاتف"><input className="inp" value={form.phone} onChange={e => set('phone', e.target.value)} /></Field>
        <div className="sm:col-span-2">
          {/* نفس آلية الموردين: عنوان + رابط + خريطة، والرابط يغني عن الخريطة */}
          <LocationFields
            addressLabel="عنوان الموقع"
            address={form.location}
            onAddressChange={(v) => set('location', v)}
            point={mapPoint}
            onPointChange={setMapPoint}
            locationUrl={locationUrl}
            onLocationUrlChange={setLocationUrl}
            resolveUrl={(u) => request<{ lat: number; lng: number }>(`/geo/resolve-map-link?url=${encodeURIComponent(u)}`)}
          />
        </div>
        <Field label="نوع العمل">
          <select className="inp" value={form.workType} onChange={e => set('workType', e.target.value)}>
            {workTypes.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </Field>
        <Field label="الطرف الوسيط"><input className="inp" value={form.refPerson} onChange={e => set('refPerson', e.target.value)} /></Field>
        <Field label="الأولوية">
          <select className="inp" value={form.priority} onChange={e => set('priority', e.target.value)}>
            <option value="عادي">عادي</option>
            <option value="عاجل جداً">عاجل جداً 🔥</option>
          </select>
        </Field>
        <Field label="تاريخ انتهاء المشروع التقريبي"><input type="date" className="inp" value={form.deliveryDate} onChange={e => set('deliveryDate', e.target.value)} /></Field>
      </div>
      <button onClick={save} disabled={saving}
        className="w-full mt-5 py-2.5 rounded-lg bg-[var(--color-brand-500)] text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'حفظ ✅'}
      </button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Edit project modal
// ---------------------------------------------------------------------------
// 💰 قيمة العقد الكلية (دازاي 10-31): نفس قيمة «الدفعات» (ProjectContractValue) بس تنكتب من
// نموذج المشروع. الي ما يحق له فلوس المشروع ما يشوف الخانة (الخادم يرفض بلا مخالفة).
export function ContractValueField({ projectId }: { projectId: string }) {
  const [state, setState] = useState<ProjectPaymentsState | null>(null)
  const [value, setValue] = useState('')
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  useEffect(() => {
    let alive = true
    void api.getProjectPayments(projectId).then((s) => {
      if (!alive) return
      setState(s); setValue(s.money.contractValue && !s.money.valueFromText ? String(s.money.contractValue) : '')
    }).catch(swallow)
    return () => { alive = false }
  }, [projectId])
  if (!state) return null
  const m = state.money
  const num = Number(value.replace(/[^\d]/g, ''))
  const save = async () => {
    if (!num) { setMsg('اكتب قيمة العقد'); return }
    setBusy(true); setMsg('')
    try { const s = await api.setProjectContractValue(projectId, num); setState(s); setMsg('انحفظت ✅') }
    catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر الحفظ') }
    finally { setBusy(false) }
  }
  return (
    <div className="sm:col-span-2 rounded-lg border border-emerald-200 bg-emerald-50 p-3">
      <p className="text-xs font-bold text-emerald-900">💰 قيمة العقد الكلية</p>
      {state.canValue ? (
        <div className="mt-2 flex items-end gap-2">
          <input inputMode="numeric" className="inp flex-1" placeholder="مثلاً 500000000" value={value} onChange={(e) => setValue(e.target.value)} />
          <button type="button" onClick={save} disabled={busy} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : 'حفظ القيمة'}</button>
        </div>
      ) : (
        <p className="mt-1 text-sm font-bold">{m.contractValue != null ? `${m.contractValue.toLocaleString('en-US')} د.ع` : '—'}</p>
      )}
      {m.contractValue != null && (
        <p className="mt-1 text-[11px] text-emerald-800">المدفوع {m.paid.toLocaleString('en-US')} · الباقي {Math.max(0, m.contractValue - m.paid).toLocaleString('en-US')} د.ع{m.valueFromText ? ' · (القيمة الحالية مأخوذة من نص المبلغ التخميني)' : ''}</p>
      )}
      {msg && <p className="mt-1 text-[11px] text-slate-700">{msg}</p>}
    </div>
  )
}

export function EditModal({ project, onClose, onSaved }: { project: Project; onClose: () => void; onSaved: () => void }) {
  const workTypes = useProjectWorkTypes()
  const [form, setForm] = useState({
    name: project.name, rep: project.rep || '', phone: project.phone || '',
    location: project.location || '', workType: project.workType || 'طاقة شمسية',
    refPerson: project.refPerson || '', priority: project.priority, deliveryDate: project.deliveryDate || '',
  })
  const [mapPoint, setMapPoint] = useState<{ lat: number; lng: number } | null>(
    project.mapLatitude != null && project.mapLongitude != null ? { lat: project.mapLatitude, lng: project.mapLongitude } : null,
  )
  const [locationUrl, setLocationUrl] = useState(project.locationUrl || '')
  const [saving, setSaving] = useState(false)
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const save = async () => {
    if (!form.name.trim()) { alert('اسم المؤسسة مطلوب'); return }
    setSaving(true)
    try {
      await request(`/projects/${project.id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...form, mapLatitude: mapPoint?.lat, mapLongitude: mapPoint?.lng, locationUrl: locationUrl || null }),
      })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title={`تعديل: ${project.name}`} wide>
      <div className="grid sm:grid-cols-2 gap-3">
        <Field label="المؤسسة *"><input className="inp" value={form.name} onChange={e => set('name', e.target.value)} /></Field>
        <Field label="الممثل"><input className="inp" value={form.rep} onChange={e => set('rep', e.target.value)} /></Field>
        <Field label="الهاتف"><input className="inp" value={form.phone} onChange={e => set('phone', e.target.value)} /></Field>
        <div className="sm:col-span-2">
          {/* نفس آلية الموردين: عنوان + رابط + خريطة، والرابط يغني عن الخريطة */}
          <LocationFields
            addressLabel="عنوان الموقع"
            address={form.location}
            onAddressChange={(v) => set('location', v)}
            point={mapPoint}
            onPointChange={setMapPoint}
            locationUrl={locationUrl}
            onLocationUrlChange={setLocationUrl}
            resolveUrl={(u) => request<{ lat: number; lng: number }>(`/geo/resolve-map-link?url=${encodeURIComponent(u)}`)}
          />
        </div>
        <Field label="نوع العمل">
          <select className="inp" value={form.workType} onChange={e => set('workType', e.target.value)}>
            {workTypes.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </Field>
        <Field label="الطرف الوسيط"><input className="inp" value={form.refPerson} onChange={e => set('refPerson', e.target.value)} /></Field>
        <Field label="الأولوية">
          <select className="inp" value={form.priority} onChange={e => set('priority', e.target.value)}>
            <option value="عادي">عادي</option>
            <option value="عاجل جداً">عاجل جداً 🔥</option>
          </select>
        </Field>
        <Field label="تاريخ انتهاء المشروع التقريبي"><input type="date" className="inp" value={form.deliveryDate} onChange={e => set('deliveryDate', e.target.value)} /></Field>
        <ContractValueField projectId={project.id} />
      </div>
      <button onClick={save} disabled={saving}
        className="w-full mt-5 py-2.5 rounded-lg bg-[var(--color-brand-500)] text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'حفظ التعديلات ✅'}
      </button>
    </Modal>
  )
}

export function MediaTransferModal({ project, onClose, onSaved }: { project: Project; onClose: () => void; onSaved: () => void }) {
  const candidates = useProjectCandidates()
  const [existing, setExisting] = useState<OpenBrief | null>(null)
  const [engineer, setEngineer] = useState(project.responsibleEmployeeId || project.delegatedToEmployeeId || project.surveyorEmployeeId || '')
  const [start, setStart] = useState('')
  const [end, setEnd] = useState(project.deliveryDate || '')
  const [duration, setDuration] = useState(project.time || '')
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)
  useEffect(() => {
    void request<OpenBrief | null>(`/projects/${project.id}/media`).then((b) => {
      if (!b) return
      setExisting(b)
      if (b.engineerId) setEngineer(b.engineerId)
      setStart(localDT(b.startAt))
      if (b.expectedEndAt) setEnd(dayBaghdad(b.expectedEndAt))
      setDuration(b.duration || '')
      setNotes(b.notes || '')
    }).catch(swallow)
  }, [project.id])
  const save = async () => {
    if (!start) { alert('حدد متى يبدي الشغل حتى الإعلام يعرف يجون يصوّرون.'); return }
    setSaving(true)
    try {
      await request(`/projects/${project.id}/media`, { method: 'POST', body: JSON.stringify({
        engineerId: engineer || null, startAt: start, expectedEndAt: end || null, duration: duration || null, notes: notes || null,
      }) })
      resetMediaOpenCache()
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }
  return (
    <Modal onClose={onClose} title={`📸 إعلام: ${project.name}`}>
      <div className="space-y-3">
        <div className="rounded-lg bg-pink-50 p-3 text-center text-sm font-bold text-pink-700">
          {existing ? '📸 المشروع محوّل للإعلام — تكدر تحدّث التفاصيل' : '📸 تحويل المشروع للإعلام والعلاقات العامة'}
        </div>
        <p className="text-xs text-gray-600">ينرسل للإعلام: اسم المشروع ونوعه، الموقع والخريطة، المهندس المشرف ورقمه، متى يبدي الشغل وشكد ياخذ ومتى يخلص — حتى يطلعون يصوّرون الشغل وينشرونه. المشروع يكمّل مراحله عادي.</p>
        <Field label="المهندس المشرف">
          <select className="inp" value={engineer} onChange={e => setEngineer(e.target.value)}>
            <option value="">-- اختر المهندس --</option>
            <EmployeeOptions candidates={candidates} />
          </select>
        </Field>
        <Field label="متى يبدي الشغل *"><input type="datetime-local" className="inp" value={start} onChange={e => setStart(e.target.value)} /></Field>
        <Field label="مدة الشغل (مثلاً: ٣ أيام)"><input className="inp" value={duration} onChange={e => setDuration(e.target.value)} /></Field>
        <Field label="متى يخلص المشروع"><input type="date" className="inp" value={end} onChange={e => setEnd(e.target.value)} /></Field>
        <Field label="ملاحظات للإعلام (شنو يصوّرون، أوقات مناسبة…)"><textarea className="inp" rows={3} value={notes} onChange={e => setNotes(e.target.value)} /></Field>
        <button onClick={() => void save()} disabled={saving}
          className="w-full rounded-lg bg-pink-600 py-2.5 font-bold text-white hover:brightness-110 disabled:opacity-50">
          {saving ? 'جاري الإرسال…' : existing ? 'حدّث التفاصيل' : 'حوّل للإعلام 📸'}
        </button>
      </div>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Move / stage-transition modal
// ---------------------------------------------------------------------------
export function MoveModal({ project, nextStage, onClose, onSaved }: {
  project: Project; nextStage: string; onClose: () => void; onSaved: () => void
}) {
  const cur = project.stage
  const isReject = nextStage.includes('مرفوض')
  const toKashf = cur.includes('اتصال') && nextStage.includes('كشف')
  const toSer = cur.includes('كشف') && nextStage.includes('سعر')
  const canUploadSurvey = useCanUploadProjectSurvey(project)
  const toContract = cur.includes('سعر') && nextStage.includes('عقد')
  const toExec = (cur.includes('تجهيز') || cur.includes('الإعلام')) && nextStage.includes('تنفيذ')
  const toMedia = nextStage.includes('الإعلام')
  const [mediaEngineer, setMediaEngineer] = useState(project.responsibleEmployeeId || project.delegatedToEmployeeId || project.surveyorEmployeeId || '')
  const [mediaStart, setMediaStart] = useState('')
  const [mediaEnd, setMediaEnd] = useState(project.deliveryDate || '')
  const [mediaDuration, setMediaDuration] = useState(project.time || '')
  const toDone = cur.includes('التسليم') && nextStage.includes('مكتمل')
  // قرار (ع) 10-06: «مينتهي المشروع بدون ميكتبون» — الفلوس لازم تنسجّل قبل الإكمال.
  const toComplete = nextStage.includes('مكتمل')
  const [money, setMoney] = useState<ProjectMoney | null>(null)
  const [showPay, setShowPay] = useState(false)
  const [unpaidNote, setUnpaidNote] = useState('')
  const [moneyTick, setMoneyTick] = useState(0)
  useEffect(() => {
    if (toComplete) void api.getProjectPayments(project.id).then((s) => setMoney(s.money)).catch(swallow)
  }, [toComplete, project.id, moneyTick])
  const leftToPay = money?.contractValue != null ? money.contractValue - money.paid : null
  const moneyBlock = toComplete && (!money || money.contractValue == null || money.contractValue <= 0 || money.payments === 0)
  const needNote = toComplete && !moneyBlock && leftToPay != null && leftToPay > 0

  const candidates = useProjectCandidates()
  // تحديد المسؤول ومنفّذ الكشف يصير هنا (عند الترحيل لمرحلة الكشف)، مو بفورمة
  // إضافة/تعديل المشروع — لأن هذي اللحظة الي ينعرف بيها مين راح يطلع كشف.
  const [responsibleEmployeeId, setResponsibleEmployeeId] = useState(project.responsibleEmployeeId || '')
  const [surveyorEmployeeId, setSurveyorEmployeeId] = useState(project.surveyorEmployeeId || '')
  const [staff, setStaff] = useState(project.staff || '')
  const [location, setLocation] = useState(project.location || '')
  const [time, setTime] = useState('')
  const [task, setTask] = useState(project.task || '')
  const [price, setPrice] = useState(project.price || '')
  const [survey, setSurvey] = useState<string[]>(SURVEY.map((s, i) => project.survey?.[i] || s.opts[0]))
  // القائمة تنزل بلا استمارة الكشف (حتى تخف) — نجيبها هنا إذا بنحتاجها.
  useEffect(() => {
    if (!toSer) return
    void request<Project>(`/projects/${project.id}`).then((full) => {
      if (full.survey?.length) setSurvey(SURVEY.map((s, i) => full.survey?.[i] || s.opts[0]))
    }).catch(swallow)
  }, [toSer, project.id])
  const [reason, setReason] = useState('')
  const [otherReason, setOtherReason] = useState('')
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

  const save = async () => {
    if (isReject && !reason) { alert('اختر سبب الرفض!'); return }
    if (toMedia) {
      if (!mediaStart) { alert('حدد متى يبدي الشغل حتى الإعلام يعرف يجون يصوّرون.'); return }
      setSaving(true)
      try {
        await request(`/projects/${project.id}/media`, { method: 'POST', body: JSON.stringify({
          engineerId: mediaEngineer || null, startAt: mediaStart, expectedEndAt: mediaEnd || null,
          duration: mediaDuration || null, notes: notes || null,
        }) })
        onSaved()
      } catch (e) { alert((e as Error).message); setSaving(false) }
      return
    }
    const payload: Record<string, unknown> = { stage: nextStage }
    if (isReject) {
      let finalReason = reason
      if (reason.includes('آخر') && otherReason.trim()) finalReason = otherReason.trim()
      payload.task = `رفض المشروع: ${finalReason}${notes ? ' | ملاحظات: ' + notes : ''}`
    } else {
      if (staff) payload.staff = staff
      if (time) payload.time = time
      if (location) payload.location = location
      if (task) payload.task = task
      if (price) payload.price = price
      if (needNote) payload.unpaidNote = unpaidNote
      if (toSer) payload.survey = survey
      if (toKashf) {
        payload.responsibleEmployeeId = responsibleEmployeeId
        payload.surveyorEmployeeId = surveyorEmployeeId
      }
    }
    setSaving(true)
    try {
      await request(`/projects/${project.id}`, { method: 'PUT', body: JSON.stringify(payload) })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title={`ترحيل إلى: ${nextStage}`}>
      {/* Rejection form */}
      {isReject && (
        <div className="space-y-3">
          <div className="bg-red-50 text-red-700 rounded-lg p-3 text-center text-sm font-bold">⚠️ تحويل المشروع إلى مرفوض</div>
          <Field label="سبب الرفض *">
            <select className="inp border-red-300" value={reason} onChange={e => setReason(e.target.value)}>
              <option value="">-- اختر سبب الرفض --</option>
              {CATEGORIES.filter(c => c.key !== 'all').map(cat => (
                <optgroup key={cat.key} label={cat.label}>
                  {REJECTION_REASONS.filter(r => r.category === cat.key).map(r => (
                    <option key={r.value} value={r.value}>{r.value}</option>
                  ))}
                </optgroup>
              ))}
            </select>
          </Field>
          {reason.includes('آخر') && (
            <input className="inp" placeholder="اكتب السبب بالتفصيل..." value={otherReason} onChange={e => setOtherReason(e.target.value)} />
          )}
          <Field label="ملاحظات إضافية"><textarea className="inp" rows={3} value={notes} onChange={e => setNotes(e.target.value)} /></Field>
        </div>
      )}

      {/* اتصال → كشف : تجهيز الكشف الميداني */}
      {!isReject && toKashf && (
        <div className="space-y-3">
          <div className="bg-green-50 text-green-700 rounded-lg p-3 text-center text-sm font-bold">📋 تجهيز بيانات الكشف الميداني</div>
          <Field label="المسؤول عن المشروع (مهندس فقط)">
            <select className="inp" value={responsibleEmployeeId} onChange={e => setResponsibleEmployeeId(e.target.value)}>
              <option value="">-- اختر المهندس --</option>
              <EmployeeOptions candidates={candidates.filter(c => c.isEngineer)} />
            </select>
          </Field>
          <Field label="منفّذ الكشف">
            <select className="inp" value={surveyorEmployeeId} onChange={e => setSurveyorEmployeeId(e.target.value)}>
              <option value="">-- اختر الموظف --</option>
              <EmployeeOptions candidates={candidates} />
            </select>
          </Field>
          <Field label="فريق الكشف (أسماء الفنيين)"><textarea className="inp" rows={2} value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="مكان/موقع الكشف بالتفصيل"><input className="inp" value={location} onChange={e => setLocation(e.target.value)} /></Field>
          <Field label="تاريخ ووقت الكشف المقرر"><input type="datetime-local" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="ملاحظات خاصة بالكشف"><textarea className="inp" rows={2} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      {/* كشف → سعر : استمارة 17 سؤال + المبلغ */}
      {!isReject && toSer && (
        <div className="space-y-2 max-h-96 overflow-y-auto border rounded-lg p-3">
          {SURVEY.map((item, i) => (
            <div key={i}>
              <label className="text-xs font-bold text-gray-600">{i + 1}. {item.q}</label>
              <select className="inp" value={survey[i]} onChange={e => setSurvey(s => { const n = [...s]; n[i] = e.target.value; return n })}>
                {item.opts.map(o => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
          ))}
          <Field label="المبلغ التخميني"><input className="inp" placeholder="المبلغ المقترح..." value={price} onChange={e => setPrice(e.target.value)} /></Field>
          <SurveyPhotos owner={{ projectId: project.id }} canUpload={canUploadSurvey} />
        </div>
      )}

      {/* سعر → عقد : تنبيه لازم يترفع العقد قبل ما يترحّل للتنفيذ */}
      {!isReject && toContract && <div className="mb-3 grid"><ContractValueField projectId={project.id} /></div>}
      {!isReject && toContract && (
        <div className="bg-purple-50 text-purple-700 rounded-lg p-3 text-center text-sm font-bold">
          📄 بعد الترحيل، ارفع ملف العقد (PDF) من زر "العقد" بالبطاقة — ما تكدر تترحّل لمرحلة تجهيز المنتجات قبل رفعه.
        </div>
      )}

      {/* عقد → الإعلام : تحويل كل التفاصيل للإعلام حتى يطلعون يصوّرون */}
      {!isReject && toMedia && (
        <div className="space-y-3">
          <div className="rounded-lg bg-pink-50 p-3 text-center text-sm font-bold text-pink-700">📸 تحويل المشروع للإعلام والعلاقات العامة</div>
          <p className="text-xs text-gray-600">ينرسل للإعلام: اسم المشروع ونوعه، الموقع والخريطة، المهندس المشرف ورقمه، متى يبدي الشغل وشكد ياخذ ومتى يخلص — حتى يطلعون يصوّرون الشغل وينشرونه.</p>
          <Field label="المهندس المشرف">
            <select className="inp" value={mediaEngineer} onChange={e => setMediaEngineer(e.target.value)}>
              <option value="">-- اختر المهندس --</option>
              <EmployeeOptions candidates={candidates} />
            </select>
          </Field>
          <Field label="متى يبدي الشغل *"><input type="datetime-local" className="inp" value={mediaStart} onChange={e => setMediaStart(e.target.value)} /></Field>
          <Field label="مدة الشغل (مثلاً: ٣ أيام)"><input className="inp" value={mediaDuration} onChange={e => setMediaDuration(e.target.value)} /></Field>
          <Field label="متى يخلص المشروع"><input type="date" className="inp" value={mediaEnd} onChange={e => setMediaEnd(e.target.value)} /></Field>
          <Field label="ملاحظات للإعلام (شنو يصوّرون، أوقات مناسبة…)"><textarea className="inp" rows={3} value={notes} onChange={e => setNotes(e.target.value)} /></Field>
        </div>
      )}

      {/* عقد → تنفيذ : الفني المسؤول */}
      {!isReject && toExec && (
        <Field label="الفني المسؤول عن التنفيذ"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
      )}

      {/* 💰 قفل الإكمال: الفلوس قبل التسليم */}
      {!isReject && toComplete && (
        <div className={`mb-3 rounded-lg p-3 text-sm ${moneyBlock ? 'bg-red-50 text-red-800' : needNote ? 'bg-amber-50 text-amber-900' : 'bg-emerald-50 text-emerald-800'}`}>
          {!money ? <p>🤖 ماتركس يفحص فلوس المشروع…</p> : (
            <>
              <p className="font-bold">🤖 فلوس المشروع: المدفوع {money.paid.toLocaleString('en-US')} من {money.contractValue != null ? money.contractValue.toLocaleString('en-US') : '—'} د.ع</p>
              {moneyBlock && <p className="mt-1">{money.contractValue == null ? 'ماكو قيمة عقد مسجّلة.' : 'ماكو ولا دفعة مسجّلة.'} المشروع ما يكتمل لحد ما تنسجّل الفلوس الي انقبضت.</p>}
              {needNote && (
                <Field label={`باقي ${leftToPay!.toLocaleString('en-US')} د.ع — ليش؟ *`}>
                  <input className="inp" value={unpaidNote} onChange={e => setUnpaidNote(e.target.value)} placeholder="مثلاً: الزبون يدفع الباقي بعد شهر / خصم متفق عليه" />
                </Field>
              )}
              {(moneyBlock || needNote) && (
                <button type="button" onClick={() => setShowPay(true)} className="mt-2 rounded-lg bg-emerald-700 px-3 py-1.5 text-xs font-bold text-white">💰 سجّل الدفعات</button>
              )}
            </>
          )}
        </div>
      )}
      {showPay && <ProjectPaymentsModal projectId={project.id} title={project.name} onClose={() => { setShowPay(false); setMoneyTick((n) => n + 1) }} />}

      {/* تنفيذ → مكتمل : التسليم النهائي */}
      {!isReject && toDone && (
        <div className="space-y-3">
          <div className="bg-green-50 text-green-700 rounded-lg p-3 text-center text-sm font-bold">✅ إنهاء المشروع والتسليم النهائي</div>
          <Field label="رقم ملف التسليم / المسؤول"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="تاريخ التسليم الفعلي"><input type="date" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="ملاحظات نهائية"><textarea className="inp" rows={2} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      {/* fallback generic */}
      {!isReject && !toKashf && !toSer && !toExec && !toDone && !toMedia && (
        <div className="space-y-3">
          <Field label="الموظف المسؤول"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="الموعد"><input type="datetime-local" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="الملاحظات"><textarea className="inp" rows={3} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      <button onClick={save} disabled={saving || moneyBlock || (needNote && unpaidNote.trim().length < 3)}
        className="w-full mt-3 py-2.5 rounded-lg bg-green-700 text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'تأكيد ✅'}
      </button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Report modal (survey / visit)
// ---------------------------------------------------------------------------
export function ReportModal({ type, project: listProject, onClose }: { type: 'survey' | 'visit'; project: Project; onClose: () => void }) {
  // القائمة بلا استمارة الكشف — نجيب المشروع كامل لمن تنفتح.
  const [full, setFull] = useState<Project | null>(null)
  useEffect(() => { void request<Project>(`/projects/${listProject.id}`).then(setFull).catch(swallow) }, [listProject.id])
  const project = full ?? listProject
  const canUpload = useCanUploadProjectSurvey(project)
  return (
    <Modal onClose={onClose} title={type === 'survey' ? 'استمارة الكشف' : 'تقرير الزيارة'} wide>
      <div className="border-2 border-gray-800 p-5 bg-white text-gray-900">
        <div className="grid grid-cols-2 gap-2 border border-gray-800 p-3 text-sm mb-3">
          <div><b>الاسم:</b> {project.name}</div>
          <div><b>الموقع:</b> {project.location || '---'}</div>
          <div><b>الهاتف:</b> {project.phone || '---'}</div>
          <div><b>الممثل:</b> {project.rep || '---'}</div>
        </div>
        {type === 'survey' ? (
          <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead><tr className="bg-gray-100">
              <th className="border border-gray-800 p-2">#</th>
              <th className="border border-gray-800 p-2 text-right">المعيار</th>
              <th className="border border-gray-800 p-2 text-right">الحالة</th>
            </tr></thead>
            <tbody>
              {SURVEY.map((item, i) => (
                <tr key={i}>
                  <td className="border border-gray-800 p-2 text-center">{i + 1}</td>
                  <td className="border border-gray-800 p-2">{item.q}</td>
                  <td className="border border-gray-800 p-2">{project.survey?.[i] || '---'}</td>
                </tr>
              ))}
            </tbody>
          </table>
          </div>
        ) : null}
        {type === 'survey' && <SurveyPhotos owner={{ projectId: project.id }} canUpload={canUpload} />}
        {type === 'survey' ? null : (
          <div className="border border-gray-800 p-3 mt-2">
            <div className="grid grid-cols-2 gap-2 text-sm mb-3">
              <div><b>المسؤول:</b> {project.staff || 'غير محدد'}</div>
              <div><b>الموعد:</b> {project.time || 'غير محدد'}</div>
            </div>
            <b>الملاحظات:</b>
            <p className="mt-1">{project.task || 'لا يوجد'}</p>
          </div>
        )}
      </div>
      <button onClick={() => window.print()} className="w-full mt-4 py-2.5 rounded-lg bg-gray-800 text-white font-bold">🖨️ طباعة</button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// توجيه المشروع لموظف
// ---------------------------------------------------------------------------
// المدير ما يريد يشتغل على المشروع بنفسه؟ يسلّمه لموظف. القائمة مرتبة من
// المهندسين وصولاً للفنيين. الموظف الموجّه له يشوف المشروع كامل بكل مراحله
// من صفحة "المشاريع الموجّهة لي" ويتحكم بيه — بس بهذا المشروع، مو بكل المشاريع.
export function DelegateModal({ project, onClose, onSaved }: {
  project: Project
  onClose: () => void
  onSaved: () => void
}) {
  const candidates = useProjectCandidates()
  const [employeeId, setEmployeeId] = useState(project.delegatedToEmployeeId || '')
  // بحث بالاسم — القائمة فيها كل موظفي الشركة موزّعين على خانات، فبدل
  // ما يدوّر المدير بالقائمة كلها يكتب اسمه ويلكاه.
  const [q, setQ] = useState('')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const save = async (targetId: string) => {
    setSaving(true)
    setError(null)
    try {
      await request(`/projects/${project.id}/delegate`, {
        method: 'PUT',
        body: JSON.stringify({ employeeId: targetId, note: note.trim() }),
      })
      onSaved()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر توجيه المشروع')
    } finally {
      setSaving(false)
    }
  }

  // المرشحون مرتبين أصلاً بالخانات من السيرفر
  const shown = q.trim()
    ? candidates.filter((c) => matches([c.name], q))
    : candidates
  const groups: { label: string; items: ProjectCandidate[] }[] = []
  for (const c of shown) {
    const last = groups[groups.length - 1]
    if (last && last.label === c.groupLabel) last.items.push(c)
    else groups.push({ label: c.groupLabel, items: [c] })
  }

  return (
    <Modal title={`توجيه لموظف: ${project.name}`} onClose={onClose}>
      <div className="space-y-4">
        {project.delegatedToName && (
          <p className="rounded-lg bg-violet-50 p-3 text-sm text-violet-800">
            المشروع موجّه حالياً إلى <b>{project.delegatedToName}</b>
            {project.delegatedAt && ` منذ ${new Date(project.delegatedAt).toLocaleDateString('ar-IQ')}`}
          </p>
        )}

        <Field label="الموظف الموجّه له (كل كوادر الشركة)">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="🔍 اكتب اسم الموظف للبحث..."
            className="mb-2 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          />
          {q.trim() && shown.length === 0 && (
            <p className="mb-2 text-xs text-red-500">ماكو موظف بهذا الاسم.</p>
          )}
          <select
            value={employeeId}
            onChange={(e) => setEmployeeId(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          >
            <option value="">-- اختر موظف --</option>
            {groups.map((g) => (
              <optgroup key={g.label} label={g.label}>
                {g.items.map((c) => (
                  <option key={c.id} value={c.id}>{candidateLabel(c)}</option>
                ))}
              </optgroup>
            ))}
          </select>
        </Field>

        <Field label="ملاحظة للموظف (اختياري)">
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={3}
            placeholder="شنو المطلوب منه بهذا المشروع؟"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          />
        </Field>

        <p className="rounded-lg bg-amber-50 p-3 text-xs text-amber-700">
          الموظف الموجّه له راح يشوف هذا المشروع بكل مراحله وتقاريره من صفحة
          «المشاريع الموجّهة لي»، ويقدر يتابعه ويحرّكه بين المراحل ويسوي عرض سعر —
          بس على هذا المشروع لحاله.
        </p>

        {error && <p className="rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>}

        <div className="flex gap-2">
          <button
            onClick={() => save(employeeId)}
            disabled={saving || !employeeId}
            className="flex-1 rounded-lg bg-violet-600 px-4 py-2 font-medium text-white disabled:opacity-50"
          >
            {saving ? 'جاري الحفظ...' : 'توجيه لموظف'}
          </button>
          {project.delegatedToEmployeeId && (
            <button
              onClick={() => save('')}
              disabled={saving}
              className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-600 disabled:opacity-50"
            >
              إلغاء التوجيه
            </button>
          )}
        </div>
      </div>
    </Modal>
  )
}
