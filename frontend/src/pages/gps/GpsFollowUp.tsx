import { useEffect, useMemo, useState } from 'react'
import {
  api,
  type GpsSubscriptionFollowUp,
  type GpsFollowUpOutcome,
  type GpsFollowUpStage,
} from '../../api'
import { useSession } from '../../session'

/**
 * متابعة تجديد اشتراكات الجي بي اس.
 *
 * الدورة الي اتفقنا عليها:
 *   انتهى الاشتراك → ٤٠ يوم → مهندس الجودة يتصل ويسجّل النتيجة
 *   → إذا رفض: ٤٠ يوم ثانية → الشريحة تحتاج حرق → مسؤول الجي بي اس
 *   يحرقها ويحرّرها فترجع للشرائح المتوفرة.
 *
 * كل الأرقام والمراحل تجي محسوبة من السيرفر — الواجهة ما تعيد حسابها حتى
 * ما يختلف الرقم بين شاشة وشاشة حسب توقيت جهاز المستخدم.
 */

const STAGE_STYLES: Record<GpsFollowUpStage, { bg: string; border: string; text: string; icon: string }> = {
  CALL_DUE: { bg: 'bg-amber-50', border: 'border-amber-300', text: 'text-amber-900', icon: '📞' },
  WAITING: { bg: 'bg-sky-50', border: 'border-sky-300', text: 'text-sky-900', icon: '⏳' },
  BURN_DUE: { bg: 'bg-red-50', border: 'border-red-300', text: 'text-red-900', icon: '🔥' },
  RESOLVED: { bg: 'bg-emerald-50', border: 'border-emerald-300', text: 'text-emerald-900', icon: '✅' },
  GRACE: { bg: 'bg-slate-50', border: 'border-slate-300', text: 'text-slate-700', icon: '🕐' },
}

const OUTCOMES: { value: GpsFollowUpOutcome; label: string; hint: string }[] = [
  { value: 'WILL_RENEW', label: 'راح يجدد', hint: 'يطلع من قائمة المتابعة' },
  { value: 'WILL_MOVE', label: 'راح يحرّك', hint: 'يطلع من قائمة المتابعة' },
  { value: 'REFUSED', label: 'ما يريد يجدد', hint: 'تبدي مهلة ٤٠ يوم ثانية بعدها تنحرق الشريحة' },
  { value: 'NO_ANSWER', label: 'ما رد على الاتصال', hint: 'يضل بقائمة الاتصال' },
]

export default function GpsFollowUp() {
  const { permissions, employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'OWNER'
  const canBurn = isAdmin || permissions.includes('gps_system')
  const canCall = isAdmin || permissions.includes('quality_control') || permissions.includes('gps_system')

  const [rows, setRows] = useState<GpsSubscriptionFollowUp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [callTarget, setCallTarget] = useState<GpsSubscriptionFollowUp | null>(null)
  const [outcome, setOutcome] = useState<GpsFollowUpOutcome>('WILL_RENEW')
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

  const load = () => {
    queueMicrotask(() => setLoading(true))
    api.getSubscriptionFollowUps()
      .then(setRows)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب قائمة المتابعة'))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const groups = useMemo(() => ({
    CALL_DUE: rows.filter((r) => r.stage === 'CALL_DUE'),
    BURN_DUE: rows.filter((r) => r.stage === 'BURN_DUE'),
    WAITING: rows.filter((r) => r.stage === 'WAITING'),
    GRACE: rows.filter((r) => r.stage === 'GRACE'),
    RESOLVED: rows.filter((r) => r.stage === 'RESOLVED'),
  }), [rows])

  const submitCall = async () => {
    if (!callTarget) return
    setSaving(true)
    try {
      await api.createDeviceFollowUp(callTarget.deviceRequestId, outcome, notes || undefined)
      setCallTarget(null); setNotes(''); setOutcome('WILL_RENEW')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل الاتصال')
    } finally { setSaving(false) }
  }

  const doSim = async (row: GpsSubscriptionFollowUp, action: 'burn' | 'release') => {
    if (!row.simCardId) return
    const word = action === 'burn' ? 'حرق' : 'تحرير'
    if (!confirm(`متأكد تريد ${word} الشريحة ${row.simNumber}؟`)) return
    try {
      if (action === 'burn') await api.burnSimCard(row.simCardId)
      else await api.releaseSimCard(row.simCardId)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : `تعذر ${word} الشريحة`)
    }
  }

  const renderRow = (row: GpsSubscriptionFollowUp) => {
    const st = STAGE_STYLES[row.stage]
    return (
      <div className={`rounded-xl border ${st.border} ${st.bg} p-4`}>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <p className={`font-bold ${st.text}`}>{row.customerName || 'زبون غير معروف'}</p>
            <p className="mt-0.5 text-sm text-slate-600" dir="ltr">{row.customerPhone || '—'}</p>
            <p className="mt-1 text-xs text-slate-500">
              {row.simNumber ? <>الشريحة: <span dir="ltr">{row.simNumber}</span></> : 'ما اكو شريحة مربوطة'}
              {row.gpsNumber && <> · جهاز: <span dir="ltr">{row.gpsNumber}</span></>}
            </p>
            {row.lastOutcomeLabel && (
              <p className="mt-1 text-xs text-slate-500">
                آخر اتصال: {row.lastOutcomeLabel}
                {row.lastCalledAt && ` — ${new Date(row.lastCalledAt).toLocaleDateString('ar-IQ')}`}
              </p>
            )}
          </div>
          <div className="text-left">
            <span className={`inline-block rounded-full bg-white px-3 py-1 text-sm font-bold ${st.text}`}>
              صارله {row.daysSinceExpiry} يوم
            </span>
            {row.daysUntilNextStep > 0 && (
              <p className="mt-1 text-xs text-slate-500">باقي {row.daysUntilNextStep} يوم للمرحلة الجاية</p>
            )}
          </div>
        </div>

        <div className="mt-3 flex flex-wrap gap-2">
          {canCall && row.stage !== 'RESOLVED' && (
            <button
              onClick={() => { setCallTarget(row); setOutcome('WILL_RENEW'); setNotes('') }}
              className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white hover:bg-brand-700"
            >
              📞 سجّل نتيجة الاتصال
            </button>
          )}
          {canBurn && row.stage === 'BURN_DUE' && row.simStatus !== 'BURNED' && (
            <button
              onClick={() => doSim(row, 'burn')}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
            >
              🔥 أشّر إنها انحرقت
            </button>
          )}
          {canBurn && row.simCardId && row.simStatus !== 'AVAILABLE' && (
            <button
              onClick={() => doSim(row, 'release')}
              className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
            >
              ♻️ تحرير الشريحة
            </button>
          )}
        </div>
      </div>
    )
  }

  const renderSection = (stage: GpsFollowUpStage, title: string, subtitle: string) => {
    const list = groups[stage]
    const st = STAGE_STYLES[stage]
    return (
      <section className="rounded-2xl bg-white p-6 shadow-sm">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h2 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>{st.icon} {title} ({list.length})</h2>
            <p className="mt-0.5 text-sm text-slate-500">{subtitle}</p>
          </div>
        </div>
        {list.length === 0
          ? <p className="rounded-xl bg-slate-50 p-6 text-center text-sm text-slate-400">ماكو ولا حالة هنا</p>
          : <div className="space-y-3">{list.map((r) => <div key={r.deviceRequestId}>{renderRow(r)}</div>)}</div>}
      </section>
    )
  }

  if (loading) return <div dir="rtl" className="p-8 text-center text-slate-500">جاري التحميل...</div>
  if (error) return <div dir="rtl" className="rounded-xl bg-red-50 p-6 text-center text-red-700">{error}</div>

  return (
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">🔄 متابعة تجديد الاشتراكات</h1>
        <p className="mt-1 text-sm text-blue-200">
          بعد ٤٠ يوم من انتهاء الاشتراك يتصل مهندس الجودة بالزبون. إذا ما راد يجدد، ننتظر ٤٠ يوم
          ثانية وبعدها الشريحة تحتاج حرق.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
        {([
          ['CALL_DUE', 'مستحق الاتصال'],
          ['BURN_DUE', 'شرائح تحتاج حرق'],
          ['WAITING', 'بانتظار المهلة الثانية'],
          ['GRACE', 'بفترة السماح'],
          ['RESOLVED', 'تم الحسم'],
        ] as [GpsFollowUpStage, string][]).map(([stage, label]) => (
          <div key={stage} className={`rounded-2xl border ${STAGE_STYLES[stage].border} ${STAGE_STYLES[stage].bg} p-4`}>
            <div className="text-2xl">{STAGE_STYLES[stage].icon}</div>
            <p className={`mt-1 text-3xl font-bold ${STAGE_STYLES[stage].text}`}>{groups[stage].length}</p>
            <p className="mt-1 text-xs font-medium text-slate-600">{label}</p>
          </div>
        ))}
      </div>

      {renderSection('CALL_DUE', 'مستحق الاتصال', 'صارلهم ٤٠ يوم فما فوق — شغل مهندس الجودة')}
      {renderSection('BURN_DUE', 'شرائح تحتاج حرق', 'خلصت المهلة الثانية والزبون ما راد يجدد — شغل مسؤول الجي بي اس')}
      {renderSection('WAITING', 'بانتظار المهلة الثانية', 'اتصلنا والزبون رفض — ننتظر لين تكمل ٨٠ يوم')}
      {renderSection('GRACE', 'بفترة السماح', 'انتهى اشتراكهم بس ما وصلوا ٤٠ يوم بعد')}
      {renderSection('RESOLVED', 'تم الحسم', 'راح يجدد أو يحرّك')}

      {callTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setCallTarget(null)}>
          <div dir="rtl" className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>نتيجة الاتصال</h3>
            <p className="mt-1 text-sm text-slate-600">
              {callTarget.customerName} — <span dir="ltr">{callTarget.customerPhone}</span>
              {' · '}صارله {callTarget.daysSinceExpiry} يوم
            </p>

            <div className="mt-4 space-y-2">
              {OUTCOMES.map((o) => (
                <label
                  key={o.value}
                  className={`flex cursor-pointer items-start gap-3 rounded-xl border p-3 transition-colors ${
                    outcome === o.value ? 'border-brand-500 bg-brand-50' : 'border-slate-200 hover:bg-slate-50'
                  }`}
                >
                  <input
                    type="radio" name="outcome" value={o.value} checked={outcome === o.value}
                    onChange={() => setOutcome(o.value)} className="mt-1"
                  />
                  <span>
                    <span className="block font-medium text-slate-800">{o.label}</span>
                    <span className="block text-xs text-slate-500">{o.hint}</span>
                  </span>
                </label>
              ))}
            </div>

            <textarea
              value={notes} onChange={(e) => setNotes(e.target.value)} rows={3}
              placeholder="ملاحظات (اختياري)"
              className="mt-4 w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
            />

            <div className="mt-4 flex gap-3">
              <button
                onClick={submitCall} disabled={saving}
                className="flex-1 rounded-lg bg-brand-600 px-4 py-3 font-medium text-white hover:bg-brand-700 disabled:opacity-50"
              >
                {saving ? 'جاري الحفظ...' : 'حفظ'}
              </button>
              <button
                onClick={() => setCallTarget(null)}
                className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700 hover:bg-slate-50"
              >
                إلغاء
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
