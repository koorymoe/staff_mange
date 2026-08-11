import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MonitorReview, type MonitorStage } from '../api'
import EntityIdentity from '../components/EntityIdentity'
import { formatCustomerCode } from '../utils/identity'

// ═══ صندوق المراقب ═══
//
// قبلها المراقب عنده وصول لأغلب الشاشات بس ماكو شي «يوصله»: لازم
// يفتح شاشة شاشة ويخمّن شنو تغيّر. فيقعد ساكت وشغله ما ينعمل.
//
// هنا الشغل يجيه: كل محطة إلها تبويب وعدّاد، وكل صف إما «سليم» أو
// «عندي ملاحظة». والملاحظة تروح للموظف صاحب الشغل وللإدارة — مو
// تبقى مدوّنة بشاشة المراقب بس.

const STAGES: { key: MonitorStage; label: string; hint: string }[] = [
  { key: 'INVOICE_BEFORE_AUDIT', label: '🧾 فاتورة قبل التدقيق', hint: 'الأرقام الأصلية قبل ما يمسّها المحاسب' },
  { key: 'INVOICE_AFTER_AUDIT', label: '✅ فاتورة بعد التدقيق', hint: 'قارن الأرقام — أي تعديل ينبيّن' },
  { key: 'BOOKING_BEFORE_CONFIRM', label: '📅 حجز قبل التثبيت', hint: 'هذا وقت الاعتراض، بعدها تصليح مو منع' },
  { key: 'BOOKING_AFTER_CONFIRM', label: '📌 حجز بعد التثبيت', hint: 'الكادر والموعد النهائي' },
  { key: 'BOOKING_AFTER_COMPLETE', label: '🏁 حجز بعد الإنجاز', hint: 'شنو انعمل فعلاً قبل ما تصير فاتورة' },
  { key: 'PROCUREMENT_FULFILLED', label: '📦 مادة انشترت', hint: 'لحظة صرف الفلوس — الكلفة والمورد' },
  { key: 'QUALITY_VERDICT', label: '⚠️ حكم الجودة', hint: 'انخصمت نقطة من موظف بناءً على كلام زبون' },
  { key: 'GPS_DEVICE_DONE', label: '📡 جهاز جي بي اس انسلّم', hint: 'الجهاز راح للزبون والاشتراك بدأ' },
]

const ROLE_LABELS: Record<string, string> = {
  HR_COORDINATOR: 'إداري الحجوزات',
  FINANCE: 'المحاسب',
  TECHNICIAN: 'الفني / الليدر',
  QUALITY_ENGINEER: 'مهندس الجودة',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  GPS_ADMIN: 'إداري الجي بي اس',
  SERVICE_MANAGER: 'مسؤول الخدمة',
}

export default function MonitorInboxPage() {
  const [stage, setStage] = useState<MonitorStage>('INVOICE_BEFORE_AUDIT')
  const [showDone, setShowDone] = useState(false)
  const [ownerRole, setOwnerRole] = useState('')
  const [rows, setRows] = useState<MonitorReview[]>([])
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [list, c] = await Promise.all([
        api.getMonitorReviews({ stage, status: showDone ? '' : 'PENDING', ownerRole }),
        api.getMonitorReviewCounts(),
      ])
      setRows(list)
      setCounts(Object.fromEntries(c.map((x) => [x.stage, x.count])))
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر جلب الصندوق')
    } finally {
      setLoading(false)
    }
  }, [stage, showDone, ownerRole])

  useEffect(() => { void load() }, [load])

  const decide = async (row: MonitorReview, flag: boolean) => {
    const note = (notes[row.id] || '').trim()
    if (flag && note.length < 5) {
      setError('اكتب الملاحظة — الموظف لازم يعرف شنو يصلّح')
      return
    }
    setBusy(row.id)
    setError(null)
    try {
      await api.decideMonitorReview(row.id, { flag, note })
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر حفظ القرار')
    } finally { setBusy(null) }
  }

  // رابط الكيان: الحجز نفتحه بشاشة الحجوزات، والفاتورة بشاشة الفواتير
  const linkOf = (row: MonitorReview) => {
    switch (row.entityType) {
      case 'BOOKING': return `/bookings?focus=${row.entityId}`
      case 'LEADER_INVOICE': return `/leader-invoices?focus=${row.entityId}`
      // صف التعديل مفتاحه معرّف **التعديل** مو الفاتورة (وإلا الفهرس
      // الفريد يبلع التعديل الثاني)، فالرابط يروح للحجز لو موجود.
      case 'INVOICE_ADJUSTMENT':
        return row.identity ? `/bookings?focus=${row.identity.bookingId}` : '/leader-invoices'
      case 'PROCUREMENT': return '/procurement'
      case 'QUALITY_FOLLOW_UP': return '/quality-follow-ups'
      case 'GPS_DEVICE': return '/gps'
      default: return '#'
    }
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">👁️ صندوق المراقب</h2>
        <p className="mt-1 text-slate-500">
          الشغل يجيك بمحطاته. كل صف إما «سليم» أو «عندي ملاحظة» — والملاحظة توصل صاحب الشغل والإدارة.
        </p>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      {/* التبويبات */}
      <div className="flex flex-wrap gap-2">
        {STAGES.map((st) => (
          <button
            key={st.key}
            onClick={() => setStage(st.key)}
            title={st.hint}
            className={`rounded-lg border px-3 py-2 text-xs font-bold ${
              stage === st.key ? 'border-brand-600 bg-brand-600 text-white' : 'border-slate-300 bg-white text-slate-700'
            }`}
          >
            {st.label}
            {counts[st.key] > 0 && (
              <span className={`mr-1.5 rounded-full px-1.5 py-0.5 text-[10px] ${stage === st.key ? 'bg-white text-brand-700' : 'bg-red-100 text-red-700'}`}>
                {counts[st.key]}
              </span>
            )}
          </button>
        ))}
      </div>

      <div className="flex flex-wrap items-center gap-3 text-xs">
        <label className="flex items-center gap-1.5 font-bold text-slate-600">
          <input type="checkbox" checked={showDone} onChange={(e) => setShowDone(e.target.checked)} />
          وريني الي انبتّ بيه بعد
        </label>
        <select
          value={ownerRole}
          onChange={(e) => setOwnerRole(e.target.value)}
          className="rounded-lg border border-slate-300 px-2 py-1.5"
        >
          <option value="">شغل كل الأدوار</option>
          {Object.entries(ROLE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <p className="text-slate-400">{STAGES.find((s) => s.key === stage)?.hint}</p>
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          ماكو شي بهاي المحطة — كلها انبتّ بيها ✓
        </p>
      )}

      <div className="space-y-3">
        {rows.map((row) => (
          <div key={row.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <Link to={linkOf(row)} className="font-bold text-[#0f2040] underline">{row.title}</Link>
                <p className="text-xs text-slate-500">{row.summary}</p>
                <p className="mt-1 text-[11px] text-slate-400">
                  {row.ownerRole && <>شغل: {ROLE_LABELS[row.ownerRole] || row.ownerRole} </>}
                  {row.ownerEmployee && <>({row.ownerEmployee.name}) </>}
                  • {new Date(row.createdAt).toLocaleString('en-GB')}
                </p>
                {/* المراقب كان يقرا «فاتورة الليدر» وبس، ولازم يفتح كل صف
                    حتى يعرف عن منو يحچي. الهوية تجي جاهزة من السيرفر. */}
                {/* رقم الفاتورة المحاسبية الي ثبّته المحاسب — المراقب
                    يدقّق وراه، وبدون الرقم ما يكدر يطابق فاتورتنا
                    بفاتورة النظام الخارجي. */}
                {row.identity?.externalInvoiceNumber && (
                  <p className="mt-1.5 inline-block rounded-lg bg-emerald-50 px-2.5 py-1 text-xs font-bold text-emerald-800">
                    🧾 رقم الفاتورة المحاسبية: <span className="font-mono">{row.identity.externalInvoiceNumber}</span>
                  </p>
                )}
                {row.identity && (
                  <EntityIdentity
                    variant="full"
                    className="mt-2"
                    fields={{
                      bookingCode: row.identity.bookingCode,
                      customerCode: formatCustomerCode({ customerCode: row.identity.customerCode }),
                      customerName: row.identity.customerName,
                      customerPhone: row.identity.customerPhone || undefined,
                      address: row.identity.address || undefined,
                      leaderName: row.identity.leaderName || undefined,
                    }}
                  />
                )}
              </div>
              {row.status !== 'PENDING' && (
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${row.status === 'FLAGGED' ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'}`}>
                  {row.status === 'FLAGGED' ? '⚠️ عليه ملاحظة' : '✓ سليم'}
                </span>
              )}
            </div>

            {row.status === 'PENDING' ? (
              <div className="mt-3 flex flex-wrap items-center gap-2">
                <input
                  value={notes[row.id] || ''}
                  onChange={(e) => setNotes((p) => ({ ...p, [row.id]: e.target.value }))}
                  placeholder="الملاحظة (إجبارية لو أشّرت)"
                  className="min-w-[220px] flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                />
                <button
                  onClick={() => decide(row, false)}
                  disabled={busy === row.id}
                  className="rounded-lg bg-emerald-600 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                >
                  سليم ✓
                </button>
                <button
                  onClick={() => decide(row, true)}
                  disabled={busy === row.id}
                  className="rounded-lg bg-red-600 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                >
                  عندي ملاحظة ⚠️
                </button>
              </div>
            ) : (
              row.note && (
                <p className="mt-2 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-700">
                  ملاحظة {row.reviewedBy?.name || 'المراقب'}: {row.note}
                </p>
              )
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
