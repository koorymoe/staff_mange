import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type QualityFollowUp } from '../api'
import EntityIdentity from '../components/EntityIdentity'
import { formatCustomerCode } from '../utils/identity'
import BookingExecutionSummary from '../components/BookingExecutionSummary'
import { useSession } from '../session'

const statusLabels: Record<QualityFollowUp['status'], string> = {
  PENDING: 'بانتظار التواصل',
  CONTACTED_OK: 'تم التواصل - كله تمام',
  CONTACTED_ISSUE: 'تم التواصل - اكو مشكلة',
  CONVERTED: 'تحول لحجز جديد',
  CLOSED: 'مغلقة',
}

const statusColors: Record<QualityFollowUp['status'], string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  CONTACTED_OK: 'bg-green-100 text-green-800',
  CONTACTED_ISSUE: 'bg-red-100 text-red-800',
  CONVERTED: 'bg-blue-100 text-blue-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function QualityFollowUpsPage({ embedded }: EmbeddedProps = {}) {
  const { employee } = useSession()
  // مهندس الجودة (والأدمن) يتواصلون مع الزبون مباشرة ويشوفون تفاصيله كاملة.
  // المراقب المدقق يشوف بس تقرير عام (كم متابعة قيد الانتظار وكم فيها مشكلة).
  const canContact = employee?.role === 'QUALITY_ENGINEER' || employee?.role === 'ADMIN'

  const [items, setItems] = useState<QualityFollowUp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notesDraft, setNotesDraft] = useState<Record<string, string>>({})

  const load = () => {
    api.getQualityFollowUps()
      .then(setItems)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const [busy, setBusy] = useState<string | null>(null)

  // ── حكم الجودة ──
  // التقرير الإيجابي ما يترتب عليه شي. السلبي يخصم نقطة «شكوى الزبائن»
  // من الليدر — إلا إذا المهندس شك بالزبون وطلب كشف، وقتها الغرامة
  // تنتظر لحد ما يطلع أحد يشوف بعينه.
  const verdict = async (id: string, reportType: 'POSITIVE' | 'NEGATIVE', needsInspection = false) => {
    const notes = (notesDraft[id] || '').trim()
    if (reportType === 'NEGATIVE' && notes.length < 5) {
      alert('اكتب شنو كالك الزبون — التقرير السلبي يخصم نقطة من الليدر')
      return
    }
    setBusy(id)
    try {
      setItems(await api.qualityVerdict(id, { reportType, notes, needsInspection }))
      setNotesDraft({ ...notesDraft, [id]: '' })
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل التقرير')
    } finally { setBusy(null) }
  }

  const inspect = async (id: string, result: 'CUSTOMER_RIGHT' | 'CUSTOMER_WRONG') => {
    const notes = (notesDraft[id] || '').trim()
    if (notes.length < 5) { alert('اكتب شنو شفت بالكشف — هذا الي يعتمد عليه القرار'); return }
    setBusy(id)
    try {
      setItems(await api.qualityInspect(id, { result, notes }))
      setNotesDraft({ ...notesDraft, [id]: '' })
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل نتيجة الكشف')
    } finally { setBusy(null) }
  }

  const handleUpdate = async (id: string, status: QualityFollowUp['status']) => {
    try {
      await api.updateQualityFollowUp(id, { status, contactNotes: notesDraft[id] })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  if (!canContact) {
    const pending = items.filter((i) => i.status === 'PENDING').length
    const issues = items.filter((i) => i.status === 'CONTACTED_ISSUE').length
    return (
      <div>
        {!embedded && (
          <>
            <h2 className="text-2xl font-bold text-brand-900">متابعة الجودة بعد الحجوزات</h2>
            <p className="mt-1 text-slate-500">تقرير عام — التواصل مع الزبائن مسؤولية مهندس الجودة.</p>
          </>
        )}
        {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
        {!loading && (
          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-3xl font-bold text-amber-600">{pending}</p>
              <p className="mt-1 text-sm text-slate-500">بانتظار تواصل مهندس الجودة</p>
            </div>
            <div className="rounded-xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-3xl font-bold text-red-600">{issues}</p>
              <p className="mt-1 text-sm text-slate-500">فيهن مشكلة تحتاج متابعة</p>
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div>
      {!embedded && (
        <>
          <h2 className="text-2xl font-bold text-brand-900">متابعة الجودة بعد الحجوزات</h2>
          <p className="mt-1 text-slate-500">تواصل مع الزبائن اللي اكتمل حجزهم، وتأكد ما اكو مشاكل.</p>
        </>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      {!loading && !error && (
        <div className="mt-6 space-y-4">
          {items.map((item) => (
            <div key={item.id} className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-bold text-brand-900">
                    {item.customer.name} — {item.customer.phone}
                  </p>
                  <p className="text-sm text-slate-500">
                    الحجز: {item.booking.code} — {item.booking.service?.name || 'بدون خدمة محددة'}
                  </p>
                  {/* هوية كاملة: مهندس الجودة كان يشوف كود الحجز والخدمة بس */}
                  <EntityIdentity
                    fields={{
                      bookingCode: item.booking.code,
                      customerCode: formatCustomerCode(item.customer),
                      customerName: item.customer?.name,
                      customerPhone: item.customer?.phone,
                      serviceName: item.booking.service?.name,
                    }}
                    variant="full"
                    className="mt-2"
                  />
                </div>
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[item.status]}`}>
                  {statusLabels[item.status]}
                </span>
              </div>

              {/* تفاصيل المشروع والمبالغ — حتى يعرف شنو انتفق عليه وشكد انستلم
                  فعلاً، ويقدر يكتب تفاصيل الفارق وهو يتصل بالزبون */}
              {item.financials && (
                <div className="mt-3 rounded-xl bg-slate-50 p-4">
                  <p className="mb-2 text-xs font-bold text-slate-500">تفاصيل المشروع والمبالغ</p>
                  <div className="grid grid-cols-1 gap-x-6 gap-y-1 text-sm sm:grid-cols-2">
                    {item.financials.projectCode && (
                      <p className="text-slate-600">المشروع: <span className="font-medium text-slate-800">{item.financials.projectCode} — {item.financials.projectName}</span>
                        {item.financials.projectStage && <span className="mr-1 text-xs text-slate-400">({item.financials.projectStage})</span>}</p>
                    )}
                    {item.financials.location && <p className="text-slate-600">الموقع: <span className="font-medium text-slate-800">{item.financials.location}</span></p>}
                    {item.financials.workDetails && <p className="text-slate-600 sm:col-span-2">تفاصيل العمل: <span className="font-medium text-slate-800">{item.financials.workDetails}</span></p>}
                  </div>
                  <div className="mt-3 grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <div className="rounded-lg bg-white p-3">
                      <p className="text-xs text-slate-500">المتفق عليه</p>
                      <p className="font-bold text-slate-800">{item.financials.agreedTotal.toLocaleString('en-US')}</p>
                    </div>
                    <div className="rounded-lg bg-white p-3">
                      <p className="text-xs text-slate-500">العربون</p>
                      <p className="font-bold text-slate-800">{(item.financials.advancePaid || 0).toLocaleString('en-US')}</p>
                    </div>
                    <div className="rounded-lg bg-white p-3">
                      <p className="text-xs text-slate-500">المستلم</p>
                      <p className="font-bold text-slate-800">{(item.financials.amountCollected || 0).toLocaleString('en-US')}</p>
                    </div>
                    <div className={`rounded-lg p-3 ${item.financials.difference > 0 ? 'bg-amber-100' : item.financials.difference < 0 ? 'bg-red-100' : 'bg-emerald-100'}`}>
                      <p className="text-xs text-slate-600">الفارق</p>
                      <p className={`font-bold ${item.financials.difference > 0 ? 'text-amber-900' : item.financials.difference < 0 ? 'text-red-900' : 'text-emerald-900'}`}>
                        {item.financials.difference.toLocaleString('en-US')}
                      </p>
                      <p className="text-[10px] text-slate-500">
                        {item.financials.difference > 0 ? 'باقي بذمة الزبون' : item.financials.difference < 0 ? 'انستلم أكثر من المتفق' : 'مسدّد بالكامل'}
                      </p>
                    </div>
                  </div>
                </div>
              )}
              {/* منو طلع، ومتى بدا وخلّص — قبل ما يتصل بالزبون */}
              <BookingExecutionSummary exec={item.execution} />

              {/* الزبون الي سبق واشتكى كذباً — تحذير قبل ما يبني عليه حكم */}
              {(item.customer.falseClaimCount || 0) > 0 && (
                <p className="mt-3 rounded-lg border border-orange-300 bg-orange-50 px-3 py-2 text-xs font-bold text-orange-900">
                  ⚠️ هذا الزبون انكشف {item.customer.falseClaimCount} مرة إن شكواه ما كانت صحيحة —
                  خذ حذرك وفكّر بالكشف قبل ما تخصم من الكادر.
                </p>
              )}

              {item.status === 'PENDING' && (
                <div className="mt-4 space-y-3">
                  <textarea
                    placeholder="ملاحظات التواصل مع الزبون..."
                    value={notesDraft[item.id] || ''}
                    onChange={(e) => setNotesDraft({ ...notesDraft, [item.id]: e.target.value })}
                    rows={2}
                    className="w-full rounded-lg border border-gray-300 px-4 py-2 text-right outline-none focus:border-brand-500"
                  />
                  <div className="flex flex-wrap gap-2">
                    <button
                      onClick={() => verdict(item.id, 'POSITIVE')}
                      disabled={busy === item.id}
                      className="rounded-lg bg-green-100 px-4 py-2 text-sm font-bold text-green-800 hover:bg-green-200 disabled:opacity-50"
                    >
                      ✅ تقرير إيجابي
                    </button>
                    <button
                      onClick={() => verdict(item.id, 'NEGATIVE')}
                      disabled={busy === item.id}
                      className="rounded-lg bg-red-100 px-4 py-2 text-sm font-bold text-red-800 hover:bg-red-200 disabled:opacity-50"
                    >
                      ⛔ تقرير سلبي — خصم نقطة من الليدر
                    </button>
                    {/* الزبون أحياناً يجذب. هذا الزر يوقف الغرامة لحد ما
                        يطلع أحد يشوف بعينه — بدل ما نظلم الليدر أو ننطي
                        الزبون سلاح يستعمله كل مرة. */}
                    <button
                      onClick={() => verdict(item.id, 'NEGATIVE', true)}
                      disabled={busy === item.id}
                      className="rounded-lg border border-amber-400 bg-amber-50 px-4 py-2 text-sm font-bold text-amber-900 hover:bg-amber-100 disabled:opacity-50"
                    >
                      🔍 سلبي — بس يحتاج كشف (بلا خصم الحين)
                    </button>
                  </div>
                  <p className="text-[11px] text-slate-400">
                    التقرير السلبي يخصم نقطة «شكوى الزبائن» من الليدر فوراً. لو تشك بالزبون، اختار «يحتاج كشف».
                  </p>
                </div>
              )}

              {item.inspectionStatus === 'PENDING' && (
                <div className="mt-4 rounded-xl border-2 border-amber-400 bg-amber-50 p-3">
                  <p className="text-sm font-bold text-amber-900">🔍 بانتظار الكشف الميداني</p>
                  <p className="mt-1 text-xs text-amber-800">
                    الغرامة موقوفة لحد ما يطلع أحد يشوف. بعد الكشف اكتب شنو شفت واختار النتيجة:
                  </p>
                  <textarea
                    placeholder="شنو شفت بالكشف؟ (إلزامي)"
                    value={notesDraft[item.id] || ''}
                    onChange={(e) => setNotesDraft({ ...notesDraft, [item.id]: e.target.value })}
                    rows={2}
                    className="mt-2 w-full rounded-lg border border-amber-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500"
                  />
                  <div className="mt-2 flex flex-wrap gap-2">
                    <button
                      onClick={() => inspect(item.id, 'CUSTOMER_RIGHT')}
                      disabled={busy === item.id}
                      className="rounded-lg bg-red-100 px-4 py-2 text-sm font-bold text-red-800 hover:bg-red-200 disabled:opacity-50"
                    >
                      كلام الزبون صح — يتغرّم الليدر
                    </button>
                    <button
                      onClick={() => inspect(item.id, 'CUSTOMER_WRONG')}
                      disabled={busy === item.id}
                      className="rounded-lg bg-slate-100 px-4 py-2 text-sm font-bold text-slate-800 hover:bg-slate-200 disabled:opacity-50"
                    >
                      كلام الزبون كذب — علامة على الزبون
                    </button>
                  </div>
                </div>
              )}

              {/* نتيجة الحكم بعد ما ينبتّ بيه */}
              {item.reportType && item.inspectionStatus !== 'PENDING' && (
                <div className="mt-3 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs">
                  <span className={`font-bold ${item.reportType === 'POSITIVE' ? 'text-emerald-700' : 'text-red-700'}`}>
                    {item.reportType === 'POSITIVE' ? '✅ تقرير إيجابي' : '⛔ تقرير سلبي'}
                  </span>
                  {item.penalizedEmployee && (
                    <span className="text-slate-600"> — انخصمت نقطة من {item.penalizedEmployee.name}</span>
                  )}
                  {item.inspectionResult === 'CUSTOMER_WRONG' && (
                    <span className="text-orange-700"> — الكشف بيّن إن الزبون ما كان صادق، ومحد انغرم</span>
                  )}
                  {item.inspectedBy && (
                    <span className="text-slate-400"> · كشف: {item.inspectedBy.name}</span>
                  )}
                </div>
              )}

              {item.status === 'CONTACTED_ISSUE' && (
                <div className="mt-4 flex flex-wrap items-center gap-2">
                  <Link
                    to={`/sales?customerId=${item.customer.id}`}
                    onClick={() => handleUpdate(item.id, 'CONVERTED')}
                    className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white"
                  >
                    تحويل لحجز جديد
                  </Link>
                  <button
                    onClick={() => handleUpdate(item.id, 'CLOSED')}
                    className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600"
                  >
                    إغلاق بدون حجز
                  </button>
                </div>
              )}

              {item.contactNotes && (
                <p className="mt-3 rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-600">
                  {item.contactNotes}
                </p>
              )}
            </div>
          ))}
          {items.length === 0 && (
            <p className="rounded-xl border border-white bg-white p-6 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              لا توجد متابعات جودة بعد
            </p>
          )}
        </div>
      )}
    </div>
  )
}
