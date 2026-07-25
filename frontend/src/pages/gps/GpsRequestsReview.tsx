import { useEffect, useState } from 'react'
import { api, type Employee, type GpsDeviceRequest } from '../../api'
import { useSession, hasGpsSkill } from '../../session'

const subLabel = (t: string) => t === 'THREE_MONTHS' ? '3 أشهر' : t === 'SIX_MONTHS' ? '6 أشهر' : 'سنوي'
const subDays = (t: string) => t === 'THREE_MONTHS' ? 90 : t === 'SIX_MONTHS' ? 180 : 365

export default function GpsRequestsReview() {
  const { gpsServiceId } = useSession()
  const [requests, setRequests] = useState<GpsDeviceRequest[]>([])
  const [gpsTechnicians, setGpsTechnicians] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<GpsDeviceRequest | null>(null)
  const [activationDate, setActivationDate] = useState('')
  const [checks, setChecks] = useState({ checked: false, activated: false, delivered: false })
  const [scheduledAt, setScheduledAt] = useState('')
  const [assignedTechnicianId, setAssignedTechnicianId] = useState('')
  const [saving, setSaving] = useState(false)

  const load = () => {
    // gpsServiceId can change after mount, so re-arm the loading flag outside the
    // effect's synchronous body (avoids react-hooks/set-state-in-effect) while still
    // showing the spinner before the network calls resolve.
    queueMicrotask(() => setLoading(true))
    Promise.all([
      api.getGpsDevices().then(all => setRequests(all.filter(r => r.status === 'PENDING'))),
      api.getEmployees().then(all => setGpsTechnicians(
        all.filter(e => e.role === 'TECHNICIAN' && hasGpsSkill(e, gpsServiceId))
      )),
    ]).finally(() => setLoading(false))
  }
  useEffect(load, [gpsServiceId])

  const openReview = (req: GpsDeviceRequest) => {
    setSelected(req); setActivationDate(''); setChecks({ checked: false, activated: false, delivered: false })
    setScheduledAt(''); setAssignedTechnicianId('')
  }

  const activate = async () => {
    if (!selected) return
    if (selected.purchaseType === 'DEVICE_ONLY') {
      if (!checks.checked || !checks.activated || !checks.delivered) { alert('يرجى تأكيد جميع الخطوات أولاً'); return }
    } else if (!activationDate) { alert('يرجى إدخال تاريخ التفعيل أولاً'); return }

    setSaving(true)
    try {
      const start = activationDate ? new Date(activationDate) : null
      const end = start ? new Date(start.getTime() + subDays(selected.subscriptionType) * 86400000) : null
      await api.updateGpsDevice(selected.id, {
        status: 'APPROVED',
        isChecked: selected.purchaseType === 'DEVICE_ONLY' ? checks.checked : true,
        isActivated: selected.purchaseType === 'DEVICE_ONLY' ? checks.activated : true,
        subscriptionStart: start ? start.toISOString() : undefined,
        subscriptionEnd: end ? end.toISOString() : undefined,
        activationDate: start ? start.toISOString() : undefined,
        scheduledAt: scheduledAt ? new Date(scheduledAt).toISOString() : undefined,
        assignedTechnicianId: assignedTechnicianId || undefined,
      })
      setSelected(null)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ أثناء التفعيل')
    } finally {
      setSaving(false)
    }
  }

  const typeLabel = (t: string) => t === 'DEVICE_SIM' ? 'جهاز + SIM كارد' : 'جهاز فقط'

  return (
    <div>
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلبات GPS المعلقة 📋</h2>

      {loading && <p className="py-20 text-center text-slate-400">جاري التحميل...</p>}
      {!loading && requests.length === 0 && (
        <div className="rounded-2xl bg-white py-16 text-center shadow-sm">
          <div className="mb-4 text-5xl">✅</div>
          <p className="text-lg text-slate-400">لا توجد طلبات معلقة</p>
        </div>
      )}

      <div className="flex flex-col gap-3">
        {requests.map(req => (
          <div key={req.id} className="flex items-center justify-between rounded-xl bg-white p-5 shadow-sm">
            <div>
              <div className="mb-1 flex items-center gap-3">
                <span className="font-bold text-brand-900">{req.customer.fullName} {req.customer.fatherName} {req.customer.grandfatherName}</span>
                <span className="rounded-full bg-violet-50 px-2 py-0.5 text-xs font-semibold text-violet-700">{typeLabel(req.purchaseType)}</span>
                <span className="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-semibold text-blue-700">{subLabel(req.subscriptionType)}</span>
              </div>
              <div className="flex gap-4 text-sm text-slate-500">
                <span>📞 {req.customer.phone}</span>
                <span>📍 {req.customer.address}</span>
                <span>👤 {req.employee.name}</span>
              </div>
            </div>
            <button onClick={() => openReview(req)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">مراجعة ←</button>
          </div>
        ))}
      </div>

      {selected && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
          <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white shadow-2xl">
            <div className="sticky top-0 flex items-center justify-between border-b border-slate-100 bg-white p-5">
              <h3 className="text-lg font-bold text-brand-900">مراجعة الطلب</h3>
              <button onClick={() => setSelected(null)} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
            </div>
            <div className="space-y-5 p-6">
              <div className="grid grid-cols-2 gap-3 text-sm">
                <p><b className="text-slate-500">الاسم:</b> {selected.customer.fullName} {selected.customer.fatherName} {selected.customer.grandfatherName}</p>
                <p><b className="text-slate-500">الهاتف:</b> {selected.customer.phone}</p>
                <p><b className="text-slate-500">العنوان:</b> {selected.customer.address}</p>
                <p><b className="text-slate-500">رقم الجهاز:</b> {selected.gpsNumber || '-'}</p>
              </div>

              {selected.purchaseType === 'DEVICE_ONLY' ? (
                <div className="rounded-xl bg-emerald-50 p-4">
                  <p className="mb-3 text-sm font-bold text-emerald-800">تأكيد تسليم الجهاز</p>
                  {[['checked', 'تم فحص الجهاز'], ['activated', 'تم تفعيل الجهاز'], ['delivered', 'تم تسليم الجهاز']].map(([key, label]) => (
                    <label key={key} className="mb-2 flex items-center gap-2 text-sm">
                      <input type="checkbox" checked={checks[key as keyof typeof checks]}
                        onChange={e => setChecks(prev => ({ ...prev, [key]: e.target.checked }))} />
                      {label}
                    </label>
                  ))}
                </div>
              ) : (
                <div className="rounded-xl bg-blue-50 p-4">
                  <label className="mb-2 block text-sm font-bold text-blue-800">📅 تاريخ التفعيل *</label>
                  <input type="date" value={activationDate} onChange={e => setActivationDate(e.target.value)}
                    className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm" />
                  {activationDate && (
                    <p className="mt-2 text-xs text-emerald-700">
                      ينتهي: {new Date(new Date(activationDate).getTime() + subDays(selected.subscriptionType) * 86400000).toLocaleDateString('ar-IQ')}
                    </p>
                  )}
                </div>
              )}

              <div className="rounded-xl bg-amber-50 p-4">
                <p className="mb-3 text-sm font-bold text-amber-800">📅 جدولة موعد التركيب (اختياري)</p>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="mb-1 block text-xs font-bold text-slate-500">موعد الزيارة</label>
                    <input type="datetime-local" value={scheduledAt} onChange={e => setScheduledAt(e.target.value)}
                      className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs font-bold text-slate-500">الفني المكلّف</label>
                    <select value={assignedTechnicianId} onChange={e => setAssignedTechnicianId(e.target.value)}
                      className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm">
                      <option value="">— اختر فني —</option>
                      {gpsTechnicians.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
                    </select>
                  </div>
                </div>
              </div>

              {selected.invoicePhotoUrl && (
                <div>
                  <h4 className="mb-2 text-sm font-bold text-brand-900">صورة الفاتورة</h4>
                  <img src={selected.invoicePhotoUrl} className="max-h-56 w-full rounded-xl border object-contain" />
                </div>
              )}
            </div>
            <div className="sticky bottom-0 flex gap-3 border-t border-slate-100 bg-white p-5">
              <button onClick={activate} disabled={saving} className="flex-1 rounded-lg bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {saving ? 'جاري التفعيل...' : 'تفعيل الجهاز ✅'}
              </button>
              <button onClick={() => setSelected(null)} className="rounded-lg border border-slate-200 px-6 py-3 text-sm font-bold text-slate-600 hover:bg-slate-50">إغلاق</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
