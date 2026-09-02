import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MonitorReview, type MonitorStage } from '../api'
import EntityIdentity from '../components/EntityIdentity'
import { formatCustomerCode } from '../utils/identity'
import BookingTimelineView from '../components/BookingTimeline'
import { matches } from '../utils/search'
import { useSaveGuard } from '../useSaveGuard'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import SearchBar from '../components/SearchBar'
import EmptyState from '../components/EmptyState'
import SaveError from '../components/SaveError'
import Pager from '../components/Pager'

// ═══ صندوق المراقب ═══
//
// قبلها المراقب عنده وصول لأغلب الشاشات بس ماكو شي «يوصله»: لازم
// يفتح شاشة شاشة ويخمّن شنو تغيّر. فيقعد ساكت وشغله ما ينعمل.
//
// هنا الشغل يجيه: كل محطة إلها تبويب وعدّاد، وكل صف إما «سليم» أو
// «عندي ملاحظة». والملاحظة تروح للموظف صاحب الشغل وللإدارة — مو
// تبقى مدوّنة بشاشة المراقب بس.
//
// ⚠️⚠️ محطة «مبالغ فاتورة انتعدّلت» (INVOICE_ADJUSTED) كانت
// **موجودة بالخادم وتنكتب فعلاً** (المالك يرجّع فاتورة للمحاسب، أو
// تعديل مبلغ بعد التدقيق) — بس ماكو إلها تبويب هنا، فالصفوف توصل
// وتبقى معلّقة للأبد بلا ما يشوفها المراقب أبداً. تسميتها من
// `MonitorStageLabel` بالخادم نفسه — تسمية وحدة بمكان وحد.

const STAGES: { key: MonitorStage; label: string; hint: string }[] = [
  { key: 'INVOICE_BEFORE_AUDIT', label: '🧾 فاتورة قبل التدقيق', hint: 'الأرقام الأصلية قبل ما يمسّها المحاسب' },
  { key: 'INVOICE_AFTER_AUDIT', label: '✅ فاتورة بعد التدقيق', hint: 'قارن الأرقام — أي تعديل ينبيّن' },
  { key: 'INVOICE_ADJUSTED', label: '✏️ مبالغ فاتورة انتعدّلت', hint: 'المالك رجّعها أو انعدّل مبلغها بعد التدقيق' },
  { key: 'BOOKING_BEFORE_CONFIRM', label: '📅 حجز قبل التثبيت', hint: 'هذا وقت الاعتراض، بعدها تصليح مو منع' },
  { key: 'BOOKING_AFTER_CONFIRM', label: '📌 حجز بعد التثبيت', hint: 'الكادر والموعد النهائي' },
  { key: 'BOOKING_AFTER_COMPLETE', label: '🏁 حجز بعد الإنجاز', hint: 'شنو انعمل فعلاً قبل ما تصير فاتورة' },
  { key: 'PROCUREMENT_FULFILLED', label: '📦 مادة انشترت', hint: 'لحظة صرف الفلوس — الكلفة والمورد' },
  { key: 'QUALITY_VERDICT', label: '⚠️ حكم الجودة', hint: 'انخصمت نقطة من موظف بناءً على كلام زبون' },
  { key: 'GPS_DEVICE_DONE', label: '📡 جهاز جي بي اس انسلّم', hint: 'الجهاز راح للزبون والاشتراك بدأ' },
  { key: 'SOLAR_QUOTED', label: '☀️ منظومة شمسية انتسعّرت', hint: 'السعر انحسب تلقائياً من المخزن — محد شافه قبل الزبون' },
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

/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function MonitorInboxPage({ embedded }: EmbeddedProps = {}) {
  const [stage, setStage] = useState<MonitorStage>('INVOICE_BEFORE_AUDIT')
  const [showDone, setShowDone] = useState(false)
  const [ownerRole, setOwnerRole] = useState('')
  const [query, setQuery] = useState('')
  const [rows, setRows] = useState<MonitorReview[]>([])
  const [counts, setCounts] = useState<Record<string, number>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [pageError, setPageError] = useState<string | null>(null)
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const guard = useSaveGuard()

  // ⚠️ الجلب بمكان **واحد** داخل الـeffect، والقرار يطلب التحديث
  // برفع العدّاد. و`alive` يمنع سباق الطلبات: المراقب يبدّل المحطات
  // بسرعة، وجواب محطة قديمة يوصل متأخر ويطمس الجديد — يعني يشوف
  // مراجعات محطة وهو واقف على محطة ثانية، ويتخذ قرار بالغلط.
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        // ⚠️ ٥٠٠ حد الخادم الأقصى — قبلها ما كانت الواجهة ترسل `limit`
        // فتاخذ الافتراضي ٢٠٠ بلا ترقيم ولا مؤشر إذا اكو أكثر.
        const [list, c] = await Promise.all([
          api.getMonitorReviews({ stage, status: showDone ? '' : 'PENDING', ownerRole, limit: 500 }),
          api.getMonitorReviewCounts(),
        ])
        if (alive) {
          setRows(list)
          setCounts(Object.fromEntries(c.map((x) => [x.stage, x.count])))
          setPageError(null)
        }
      } catch (e) {
        if (alive) setPageError(e instanceof Error ? e.message : 'تعذر جلب الصندوق')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [stage, showDone, ownerRole, reload])

  const decide = async (row: MonitorReview, flag: boolean) => {
    const note = (notes[row.id] || '').trim()
    if (flag && note.length < 5) {
      setPageError('اكتب الملاحظة — الموظف لازم يعرف شنو يصلّح')
      return
    }
    setPageError(null)
    const ok = await guard.run('حفظ القرار', () => api.decideMonitorReview(row.id, { flag, note }))
    if (ok) setReload((n) => n + 1)
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

  const totalPending = useMemo(() => Object.values(counts).reduce((a, b) => a + b, 0), [counts])
  const currentStage = STAGES.find((s) => s.key === stage)

  const filtered = useMemo(() => {
    if (!query.trim()) return rows
    return rows.filter((r) => matches([r.title, r.summary, r.ownerEmployee?.name], query))
  }, [rows, query])

  const start = (page - 1) * perPage
  const paged = filtered.slice(start, start + perPage)

  return (
    <div dir="rtl" className="space-y-4">
      <SaveError message={guard.error ?? pageError} onClose={() => { guard.clear(); setPageError(null) }} />

      {!embedded && (
        <PageHeader
          title="👁️ صندوق المراقب"
          subtitle="الشغل يجيك بمحطاته، ما تدوّر عليه — كل صف إما «سليم» أو «عندي ملاحظة»، والملاحظة توصل صاحب الشغل والإدارة."
        />
      )}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <StatTile label="إجمالي بانتظار القرار" icon="📥" tone="warning" value={totalPending} hint="كل المحطات مجتمعة" />
        <StatTile label={`محطة ${currentStage?.label ?? ''}`}
          icon="🎯" tone="info" value={counts[stage] ?? 0} hint={currentStage?.hint} />
        <StatTile label="بالصفحة الحالية" icon="🔍" tone="default" value={filtered.length} hint="بعد البحث" />
      </div>

      {/* التبويبات */}
      <div className="flex flex-wrap gap-2">
        {STAGES.map((st) => (
          <button
            key={st.key}
            onClick={() => { setStage(st.key); setPage(1) }}
            title={st.hint}
            className={`rounded-lg border px-3 py-2 text-xs font-bold transition-colors ${
              stage === st.key ? 'border-brand-600 bg-brand-600 text-white' : ''
            }`}
            style={stage === st.key ? undefined : { borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}
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

      <SearchBar value={query} onChange={(v) => { setQuery(v); setPage(1) }} placeholder="ابحث بالعنوان أو الملخص أو صاحب الشغل...">
        <label className="flex items-center gap-1.5 text-xs font-bold" style={{ color: 'var(--t-body)' }}>
          <input type="checkbox" checked={showDone} onChange={(e) => { setShowDone(e.target.checked); setPage(1) }} />
          وريني الي انبتّ بيه بعد
        </label>
        <select
          value={ownerRole}
          onChange={(e) => { setOwnerRole(e.target.value); setPage(1) }}
          className="rounded-lg border px-2 py-1.5 text-xs"
          style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}
        >
          <option value="">شغل كل الأدوار</option>
          {Object.entries(ROLE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
      </SearchBar>

      {loading && <p style={{ color: 'var(--t-faint)' }}>جاري التحميل...</p>}
      {!loading && filtered.length === 0 && (
        <EmptyState
          icon="✅"
          title={rows.length === 0 ? 'ماكو شي بهاي المحطة — كلها انبتّ بيها' : 'ماكو صف يطابق البحث أو الفلترة'}
          reason={rows.length === 0
            ? 'الشغل يجيك تلقائياً لهذي المحطة — لمن يصير حدث جديد يطلع هنا.'
            : `اكو ${rows.length} صف بهذي المحطة — امسح البحث أو الفلترة.`} />
      )}

      {!loading && paged.length > 0 && (
        <>
          <div className="space-y-3">
            {paged.map((row) => (
              <div key={row.id} className="rounded-xl border p-4" style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <Link to={linkOf(row)} className="font-bold underline" style={{ color: 'var(--t-title)' }}>{row.title}</Link>
                    <p className="text-xs" style={{ color: 'var(--t-muted)' }}>{row.summary}</p>
                    <p className="mt-1 text-[11px]" style={{ color: 'var(--t-faint)' }}>
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

                {/* قصة الحجز والأوقات — «المراقب يحتاج يشوف كلشي… الإداري
                    شكد تأخر يلا ثبّت الحجز والفني شكد تأخر يله طلع
                    للزبون». تنفتح بالطلب حتى ما نجيب خط زمني لكل صف. */}
                {row.identity && (
                  <details className="mt-2">
                    <summary className="cursor-pointer text-xs font-bold text-brand-700">
                      🕒 شوف قصة الحجز والأوقات
                    </summary>
                    <BookingTimelineView bookingId={row.identity.bookingId} />
                  </details>
                )}

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
                      disabled={guard.busy}
                      className="rounded-lg bg-emerald-700 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                    >
                      سليم ✓
                    </button>
                    <button
                      onClick={() => decide(row, true)}
                      disabled={guard.busy}
                      className="rounded-lg bg-red-600 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                    >
                      عندي ملاحظة ⚠️
                    </button>
                  </div>
                ) : (
                  row.note && (
                    <p className="mt-2 rounded-lg px-3 py-2 text-xs" style={{ backgroundColor: 'var(--sf-sunken)', color: 'var(--t-body)' }}>
                      ملاحظة {row.reviewedBy?.name || 'المراقب'}: {row.note}
                    </p>
                  )
                )}
              </div>
            ))}
          </div>
          <Pager page={page} perPage={perPage} total={filtered.length} unit="صف" onPage={setPage} onPerPage={setPerPage} />
        </>
      )}
    </div>
  )
}
