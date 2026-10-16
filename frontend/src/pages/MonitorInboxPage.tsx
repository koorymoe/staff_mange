import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type Justification, type MonitorReview, type MonitorStage } from '../api'
import EntityIdentity from '../components/EntityIdentity'
import { formatCustomerCode } from '../utils/identity'
import BookingTimelineView from '../components/BookingTimeline'
import BookingChainView from '../components/BookingChain'
import { matches } from '../utils/search'
import { useSaveGuard } from '../useSaveGuard'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import { useSession } from '../session'
import DuplicateReviewPage from './DuplicateReviewPage'
import SearchBar from '../components/SearchBar'
import EmptyState from '../components/EmptyState'
import SaveError from '../components/SaveError'
import Pager from '../components/Pager'
import { roleLabel, roleLabelShort } from '../roleLabels'
import MatrixNote from '../components/MatrixNote'

// ═══ صندوق المراقب ═══
//
// قبلها المراقب عنده وصول لأغلب الشاشات بس ماكو شي «يوصله»: لازم
// يفتح شاشة شاشة ويخمّن شنو تغيّر. فيقعد ساكت وشغله ما ينعمل.
//
// هنا الشغل يجيه: كل محطة إلها تبويب وعدّاد، وكل صف إما «سليم» أو
// «عندي ملاحظة». والملاحظة تروح للموظف صاحب الشغل وللإدارة — مو
// تبقى مدوّنة بشاشة المراقب بس.
//
// ⚠️ محطة جديدة تعني **سطر وحد هنا** — باقي الشاشة (العرض، الهوية،
// أزرار القرار، الخط الزمني) عام أصلاً ويشتغل بلا أي كود إضافي.
// تسميتها من `MonitorStageLabel` بالخادم نفسه — تسمية وحدة بمكان وحد.

// قرار (ع) 10-04: الصندوق أبسط — انشالت الشمسية، والجي بي اس، والمادة المشتراة
// (صارت بقصة الحجز)، وتعديل مبالغ الفاتورة، وحكم الجودة (مكانه «الجودة والشكاوى»).
// 🔴 قرار (ع) 10-16: «قبل وبعد الغيها» — الصندوق خمس تبويبات بس. الفاتورة
// تجي **بجهة وحدة** بكل تفاصيلها (سعرها، شكد جاب الليدر، شنو غيّر المحاسب)،
// والحجز يجي بكل قصته (منو سجّله، تواصل، الموعد، صباحي/مسائي، مراحل العمل).
// صفوف نفس الفاتورة/الحجز من محطات مختلفة تنلم بكارت واحد، والقرار يشملها.
type TabKey = 'INVOICES' | 'BOOKINGS' | 'AI'
const TABS: { key: TabKey; label: string; hint: string; stages: MonitorStage[] }[] = [
  { key: 'INVOICES', label: '🧾 فواتير المحاسب', hint: 'كل فاتورة بتفاصيلها — سجّل بيها خطأ لو لا', stages: ['INVOICE_BEFORE_AUDIT', 'INVOICE_AFTER_AUDIT'] },
  { key: 'BOOKINGS', label: '📋 الحجوزات', hint: 'كل حجز بقصته كاملة — سليم لو عندك ملاحظة', stages: ['BOOKING_BEFORE_CONFIRM', 'BOOKING_AFTER_CONFIRM', 'BOOKING_AFTER_COMPLETE'] },
  { key: 'AI', label: '🧠 أحكام ماتركس', hint: 'تحليل تلقائي على إشارات الحجوزات والفواتير — سليم يعني الحكم ما يثبت', stages: ['AI_VERDICT'] },
]
// ⚠️ STAGES تبقى للتسمية بس (محطات الخادم كما هي)
const STAGES: { key: MonitorStage; label: string; hint: string }[] = [
  { key: 'INVOICE_BEFORE_AUDIT', label: 'فاتورة قبل التدقيق', hint: '' },
  { key: 'INVOICE_AFTER_AUDIT', label: 'فاتورة بعد التدقيق', hint: '' },
  { key: 'BOOKING_BEFORE_CONFIRM', label: 'حجز قبل التثبيت', hint: '' },
  { key: 'BOOKING_AFTER_CONFIRM', label: 'حجز بعد التثبيت', hint: '' },
  { key: 'BOOKING_AFTER_COMPLETE', label: 'حجز بعد الإنجاز', hint: '' },
  { key: 'AI_VERDICT', label: 'أحكام ماتركس', hint: '' },
]

// ⚠️ **قائمة مرشّح مو خريطة أسماء**: هذولا أدوار أصحاب الشغل الي
// يمرّ بطوابير المراقب — والأسماء نفسها تجي من المصدر الوحيد.
const REVIEWABLE_ROLES = ['HR_COORDINATOR', 'FINANCE', 'TECHNICIAN', 'QUALITY_ENGINEER',
  'PROCUREMENT_ADMIN', 'GPS_ADMIN', 'SERVICE_MANAGER'] as const

/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function MonitorInboxPage({ embedded }: EmbeddedProps = {}) {
  const [tab, setTab] = useState<TabKey>('INVOICES')
  const tabDef = TABS.find((t) => t.key === tab)!
  // الحجوزات: حجوزات اليوم أول (انسجّلت أو موعدها أو انحدثت اليوم) — والكل بزر
  const [onlyToday, setOnlyToday] = useState(true)
  // ترتيب (ع) 10-04: «تدقيق التكرار» تبويب لحاله جنب «أحكام ماتركس» — نفس الشاشة.
  const { employee, permissions } = useSession()
  const canDup = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER' || permissions.includes('duplicate_review')
  const [dupMode, setDupMode] = useState(false)
  // (ع) 10-13: تبريرات الموظفين — ردودهم ترجع هنا والمراقب يحكم
  const [justMode, setJustMode] = useState(false)
  const [showDone, setShowDone] = useState(false)
  const [ownerRole, setOwnerRole] = useState('')
  const [query, setQuery] = useState('')
  const [rows, setRows] = useState<MonitorReview[]>([])
  const [counts, setCounts] = useState<Record<string, number>>({})
  // (ع) 10-16: رقم تبويبي الفواتير والحجوزات = عدد الكارتات، مو عدد المحطات.
  const [cards, setCards] = useState<{ invoices: number; bookings: number } | null>(null)
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
        const [list, c, cards] = await Promise.all([
          Promise.all(tabDef.stages.map((st) => api.getMonitorReviews({ stage: st, status: showDone ? '' : 'PENDING', ownerRole, limit: 500 })))
            .then((ls) => ls.flat().sort((a, b) => Number(b.urgent) - Number(a.urgent) || b.createdAt.localeCompare(a.createdAt))),
          api.getMonitorReviewCounts(),
          api.getMonitorReviewCards().catch(() => null),
        ])
        if (alive) {
          setRows(list)
          setCounts(Object.fromEntries(c.map((x) => [x.stage, x.count])))
          setCards(cards)
          setPageError(null)
        }
      } catch (e) {
        if (alive) setPageError(e instanceof Error ? e.message : 'تعذر جلب الصندوق')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [tab, showDone, ownerRole, reload]) // eslint-disable-line react-hooks/exhaustive-deps


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

  const tabCount = (t: typeof TABS[number]) =>
    t.key === 'INVOICES' && cards ? cards.invoices
      : t.key === 'BOOKINGS' && cards ? cards.bookings
        : t.stages.reduce((a, st) => a + (counts[st] ?? 0), 0)
  const totalPending = TABS.reduce((a, t) => a + tabCount(t), 0)

  const filtered = useMemo(() => {
    if (!query.trim()) return rows
    return rows.filter((r) => matches([r.title, r.summary, r.ownerEmployee?.name, r.identity?.bookingCode, r.identity?.customerName], query))
  }, [rows, query])

  // كارت واحد لكل فاتورة/حجز — صفوف المحطات المختلفة تنلم سوه
  const groups = useMemo(() => {
    const m = new Map<string, MonitorReview[]>()
    for (const r of filtered) {
      const k = tab === 'INVOICES' ? (r.details?.invoice?.invoiceId ?? r.entityId)
        : tab === 'BOOKINGS' ? (r.identity?.bookingId ?? r.entityId) : r.id
      m.set(k, [...(m.get(k) ?? []), r])
    }
    let gs = [...m.entries()].map(([key, rs]) => ({ key, rows: rs }))
    if (tab === 'BOOKINGS' && onlyToday) {
      const today = new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
      const isToday = (s?: string | null) => !!s && new Date(s).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) === today
      gs = gs.filter((g) => g.rows.some((r) => isToday(r.createdAt) || isToday(r.details?.booking?.scheduledAt) || isToday(r.details?.booking?.createdAt)))
    }
    return gs
  }, [filtered, tab, onlyToday])
  const decideGroup = async (rs: MonitorReview[], flag: boolean, key: string) => {
    const note = (notes[key] || '').trim()
    if (flag && note.length < 5) { setPageError('اكتب الملاحظة — الموظف لازم يعرف شنو يصلّح'); return }
    setPageError(null)
    const ok = await guard.run('حفظ القرار', async () => {
      for (const r of rs.filter((x) => x.status === 'PENDING')) await api.decideMonitorReview(r.id, { flag, note })
    })
    if (ok) setReload((n) => n + 1)
  }

  const start = (page - 1) * perPage
  const pagedGroups = groups.slice(start, start + perPage)

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
        <StatTile label="إجمالي بانتظار القرار" icon="📥" tone="warning" value={totalPending} hint="كل الكارتات مجتمعة" />
        <StatTile label={tabDef.label} icon="🎯" tone="info" value={tabCount(tabDef)} hint={tabDef.hint} />
        <StatTile label="بالصفحة الحالية" icon="🔍" tone="default" value={filtered.length} hint="بعد البحث" />
      </div>

      {/* التبويبات */}
      <div className="flex flex-wrap gap-2">
        {TABS.map((st) => (
          <button
            key={st.key}
            onClick={() => { setTab(st.key); setPage(1); setDupMode(false); setJustMode(false) }}
            title={st.hint}
            className={`rounded-lg border px-3 py-2 text-xs font-bold transition-colors ${
              tab === st.key && !dupMode && !justMode ? 'border-brand-600 bg-brand-600 text-white' : ''
            }`}
            style={tab === st.key && !dupMode && !justMode ? undefined : { borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}
          >
            {st.label}
            {tabCount(st) > 0 && (
              <span className={`mr-1.5 rounded-full px-1.5 py-0.5 text-[10px] ${tab === st.key ? 'bg-white text-brand-700' : 'bg-red-100 text-red-700'}`}>
                {tabCount(st)}
              </span>
            )}
          </button>
        ))}
        {canDup && (
          <button onClick={() => { setDupMode(true); setJustMode(false) }} title="حجوزات وزبائن مكررة — ماتركس يكتشفها كل ساعة"
            className={`rounded-lg border px-3 py-2 text-xs font-bold transition-colors ${dupMode ? 'border-violet-600 bg-violet-600 text-white' : ''}`}
            style={dupMode ? undefined : { borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}>
            🤖 التكرار
          </button>
        )}
        <button onClick={() => { setJustMode(true); setDupMode(false) }} title="ردود الموظفين على «طلب تبرير»"
          className={`rounded-lg border px-3 py-2 text-xs font-bold transition-colors ${justMode ? 'border-amber-600 bg-amber-600 text-white' : ''}`}
          style={justMode ? undefined : { borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}>
          📝 التبريرات
        </button>
      </div>

      {dupMode && !justMode && <DuplicateReviewPage />}
      {justMode && <JustificationsPanel />}
      <div className={dupMode || justMode ? 'hidden' : 'contents'}>
      <SearchBar value={query} onChange={(v) => { setQuery(v); setPage(1) }} placeholder="ابحث بالعنوان أو الملخص أو صاحب الشغل...">
        <label className="flex items-center gap-1.5 text-xs font-bold" style={{ color: 'var(--t-body)' }}>
          <input type="checkbox" checked={showDone} onChange={(e) => { setShowDone(e.target.checked); setPage(1) }} />
          وريني الي انبتّ بيه بعد
        </label>
        {tab === 'BOOKINGS' && (
          <label className="flex items-center gap-1.5 text-xs font-bold" style={{ color: 'var(--t-body)' }}>
            <input type="checkbox" checked={onlyToday} onChange={(e) => { setOnlyToday(e.target.checked); setPage(1) }} />
            حجوزات اليوم بس
          </label>
        )}
        <select
          value={ownerRole}
          onChange={(e) => { setOwnerRole(e.target.value); setPage(1) }}
          className="rounded-lg border px-2 py-1.5 text-xs"
          style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}
        >
          <option value="">شغل كل الأدوار</option>
          {REVIEWABLE_ROLES.map((k) => <option key={k} value={k}>{roleLabelShort(k)}</option>)}
        </select>
      </SearchBar>

      {loading && <p style={{ color: 'var(--t-faint)' }}>جاري التحميل...</p>}
      {!loading && groups.length === 0 && (
        <EmptyState
          icon="✅"
          title={rows.length === 0 ? 'ماكو شي هنا — كلها انبتّ بيها' : tab === 'BOOKINGS' && onlyToday ? 'ماكو حجوزات اليوم بانتظارك — شيل «حجوزات اليوم بس» تشوف الباقي' : 'ماكو صف يطابق البحث أو الفلترة'}
          reason={rows.length === 0
            ? 'الشغل يجيك تلقائياً لهذي المحطة — لمن يصير حدث جديد يطلع هنا.'
            : `اكو ${rows.length} صف بهذي المحطة — امسح البحث أو الفلترة.`} />
      )}

      {!loading && pagedGroups.length > 0 && (
        <>
          <div className="space-y-3">
            {pagedGroups.map(({ key: gk, rows: grs }) => { const row = grs.find((x) => x.status === 'PENDING') ?? grs[0]; return (
              /* ⚠️ الصف العاجل («غير مطابق» بفاتورة) يتميّز بإطار أحمر
                 وشارة — والترتيب يجي من الخادم أصلاً (العاجل بأول
                 المعلّق)، فما اكو فرزان يفترقان. */
              <div key={row.id} className="rounded-xl border p-4"
                style={row.urgent && row.status === 'PENDING'
                  ? { backgroundColor: 'var(--sf-card)', borderColor: '#dc2626', borderWidth: 2 }
                  : { backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    {row.urgent && row.status === 'PENDING' && (
                      <span className="mb-1 block w-fit rounded-full bg-red-600 px-2.5 py-0.5 text-[11px] font-bold text-white">
                        🔴 عاجل — يحتاج قرارك الآن
                      </span>
                    )}
                    <Link to={linkOf(row)} className="font-bold underline" style={{ color: 'var(--t-title)' }}>{row.title}</Link>
                    {row.entityType === 'AI_VERDICT'
                      ? <MatrixNote className="mt-1">{row.summary}</MatrixNote>
                      : <p className="text-xs" style={{ color: 'var(--t-muted)' }}>{row.summary}</p>}
                    <p className="mt-1 text-[11px]" style={{ color: 'var(--t-faint)' }}>
                      {row.ownerRole && <>شغل: {roleLabel(row.ownerRole)} </>}
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
                    {row.late && <LateBlock late={row.late} />}
                    {grs.length > 1 && (
                      <p className="mt-1 text-[11px] font-bold text-slate-500">🔗 {grs.length} محطات لنفس {tab === 'INVOICES' ? 'الفاتورة' : 'الحجز'}: {grs.map((x) => STAGES.find((st) => st.key === x.stage)?.label).join(' · ')}</p>
                    )}
                    {row.details?.invoice && <InvoiceDetailBlock d={row.details.invoice} />}
                    {row.details?.booking && tab !== 'INVOICES' && <BookingDetailBlock d={row.details.booking} />}
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
                    <BookingChainView bookingId={row.identity.bookingId} />
                  </details>
                )}

                {row.status === 'PENDING' ? (
                  <div className="mt-3 flex flex-wrap items-center gap-2">
                    <input
                      value={notes[gk] || ''}
                      onChange={(e) => setNotes((p) => ({ ...p, [gk]: e.target.value }))}
                      placeholder="الملاحظة (إجبارية لو أشّرت)"
                      className="min-w-[220px] flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                    <button
                      onClick={() => decideGroup(grs, false, gk)}
                      disabled={guard.busy}
                      className="rounded-lg bg-emerald-700 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                    >
                      {tab === 'INVOICES' ? 'ماكو خطأ ✓' : 'سليم ✓'}
                    </button>
                    <button
                      onClick={() => decideGroup(grs, true, gk)}
                      disabled={guard.busy}
                      className="rounded-lg bg-red-600 px-3 py-2 text-xs font-bold text-white disabled:opacity-50"
                    >
                      {tab === 'INVOICES' ? 'بيها خطأ ⚠️' : 'عندي ملاحظة ⚠️'}
                    </button>
                    {(row.late?.employeeId || row.ownerEmployeeId || row.entityType === 'AI_VERDICT') && (
                      <ActionMenu row={row} note={notes[gk] || ''} />
                    )}
                  </div>
                ) : (
                  row.note && (
                    <p className="mt-2 rounded-lg px-3 py-2 text-xs" style={{ backgroundColor: 'var(--sf-sunken)', color: 'var(--t-body)' }}>
                      ملاحظة {row.reviewedBy?.name || 'المراقب'}: {row.note}
                    </p>
                  )
                )}
              </div>
            ) })}
          </div>
          <Pager page={page} perPage={perPage} total={groups.length} unit={tab === 'INVOICES' ? 'فاتورة' : tab === 'BOOKINGS' ? 'حجز' : 'صف'} onPage={setPage} onPerPage={setPerPage} />
        </>
      )}
      </div>
    </div>
  )
}

// ═══ تفاصيل «تأخر بالخروج للزبون» — (ع): الحجز، والليدر، وزر يودّي،
// وشكد تأخر، وشوكت الموعد، وشنو السبب ═══
const fmtDur = (m: number) => (m >= 60 ? `${Math.floor(m / 60)} س ${m % 60} د` : `${m} د`)
const fmtAt = (s: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'medium', timeStyle: 'short' }) : '—')
function LateBlock({ late }: { late: NonNullable<MonitorReview['late']> }) {
  const cards: { icon: string; label: string; value: string; tone?: string }[] = [
    { icon: '⏰', label: 'شكد تأخر', value: late.minutesLate != null ? fmtDur(late.minutesLate) : '—', tone: 'text-red-700' },
    { icon: '📅', label: 'الموعد المحدد', value: fmtAt(late.scheduledAt) },
    { icon: '🚗', label: 'طلع الساعة', value: fmtAt(late.departedAt) },
    { icon: '🔁', label: 'تأخر بآخر ٣٠ يوم', value: late.count30d != null ? `${late.count30d} مرة` : '—', tone: (late.count30d ?? 0) >= 4 ? 'text-red-700' : undefined },
  ]
  return (
    <div className="mt-2 rounded-xl border border-red-100 bg-red-50/40 p-3">
      <div className="flex flex-wrap items-center gap-2 text-xs">
        {late.employeeName && (
          <Link to={late.employeeId ? `/matrix/employee/${late.employeeId}` : '#'} className="flex items-center gap-1.5 rounded-full bg-white px-2.5 py-1 font-bold text-[#0f2040] ring-1 ring-slate-200 hover:ring-sky-300">
            <span className="grid h-6 w-6 place-items-center rounded-full bg-sky-100 text-[10px] text-sky-800">{late.employeeName.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join(' ')}</span>
            👤 {late.employeeName}
          </Link>
        )}
        {late.bookingCode && (
          <Link to={late.bookingId ? `/bookings?focus=${late.bookingId}` : '/bookings'} className="rounded-full bg-amber-500 px-3 py-1 font-bold text-white hover:bg-amber-600">📋 الحجز {late.bookingCode} — ودّيني ←</Link>
        )}
        {late.threshold != null && <span className="text-[11px] text-slate-500">الحد المسموح {late.threshold} دقيقة</span>}
      </div>
      <div className="mt-2 grid grid-cols-2 gap-2 md:grid-cols-4">
        {cards.map((c) => (
          <div key={c.label} className="rounded-lg bg-white p-2 text-center ring-1 ring-slate-100">
            <span>{c.icon}</span>
            <p className="text-[10px] text-slate-500">{c.label}</p>
            <b className={`text-xs ${c.tone ?? 'text-[#0f2040]'}`}>{c.value}</b>
          </div>
        ))}
      </div>
      <p className="mt-2 text-xs">
        <b className="text-slate-600">💬 سبب التأخير: </b>
        {late.reason ? <span className="text-slate-800">{late.reason}</span> : <span className="text-slate-400">ما كتب الكادر سبب — اسأله واكتبه بالملاحظة تحت.</span>}
      </p>
      <LateHistory late={late} />
    </div>
  )
}

// (ع): «هل صارت هاي الحالة سابقاً؟» — آخر مرات تأخر نفس الموظف، وشنو انحكم عليها.
const REVIEW_LABEL = {
  OK: { text: '✓ سليم', cls: 'bg-emerald-100 text-emerald-700' },
  FLAGGED: { text: '⚠️ ملاحظة', cls: 'bg-red-100 text-red-700' },
  PENDING: { text: '⏳ ما انحكمت', cls: 'bg-slate-100 text-slate-600' },
} as const
function LateHistory({ late }: { late: NonNullable<MonitorReview['late']> }) {
  const items = late.history ?? []
  return (
    <div className="mt-3 rounded-lg bg-white p-2 ring-1 ring-slate-100">
      <div className="mb-1.5 flex flex-wrap items-center justify-between gap-2">
        <b className="text-xs text-slate-700">🔁 صارت قبل؟ <span className="font-normal text-slate-500">(آخر ٩٠ يوم)</span></b>
        {late.employeeId && <Link to={`/matrix/employee/${late.employeeId}`} className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[11px] font-bold text-sky-800 hover:bg-sky-100">📄 تقرير الموظف الكامل ←</Link>}
      </div>
      {items.length === 0
        ? <p className="text-xs text-emerald-700">أول مرة ✅ — ما تأخر بالخروج قبلها.</p>
        : (
          <ul className="space-y-1.5">
            {items.map((h) => {
              const rv = REVIEW_LABEL[h.review ?? 'PENDING']
              return (
                <li key={h.occurredAt + (h.bookingId ?? '')} className="flex flex-wrap items-center gap-x-2 gap-y-1 border-b border-slate-100 pb-1.5 text-[11px] last:border-0 last:pb-0">
                  <span className="text-slate-500">{fmtAt(h.occurredAt)}</span>
                  <b className="text-red-700">{h.minutesLate != null ? fmtDur(h.minutesLate) : '—'}</b>
                  {h.bookingCode && <Link to={h.bookingId ? `/bookings?focus=${h.bookingId}` : '/bookings'} className="font-bold text-amber-700 hover:underline">📋 {h.bookingCode} ←</Link>}
                  <span className={`rounded-full px-2 py-0.5 ${rv.cls}`}>{rv.text}</span>
                  <span className="basis-full text-slate-700">💬 {h.reason ?? <span className="text-slate-400">بلا سبب مكتوب</span>}{h.reviewNote && <span className="text-slate-500"> — ملاحظة المراقب: {h.reviewNote}</span>}</span>
                </li>
              )
            })}
          </ul>
        )}
    </div>
  )
}

// ═══ «اتخاذ إجراء» — تنبيه، طلب تبرير، متابعة، إحالة (بلا غرامة ولا نقاط) ═══
// ═══ التبريرات — (ع) 10-13 ═══
function JustificationsPanel() {
  const [rows, setRows] = useState<Justification[] | null>(null)
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [reload, setReload] = useState(0)
  useEffect(() => { api.getJustifications().then(setRows).catch(() => setRows([])) }, [reload])
  const decide = async (id: string, verdict: 'ACCEPTED' | 'REJECTED') => {
    try { await api.decideJustification(id, verdict, notes[id] || ''); setReload((x) => x + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر') }
  }
  if (!rows) return <p className="py-8 text-center text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) return <p className="py-8 text-center text-slate-500">ماكو طلبات تبرير بعد. اطلبها من «⚡ اتخاذ إجراء ← طلب تبرير».</p>
  const waiting = rows.filter((r) => r.answeredAt && !r.verdict).length
  const fmt = (s: string | null) => s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) : ''
  return (
    <div className="space-y-2">
      <p className="text-sm font-bold text-slate-600">ردود تنتظر حكمك: <b className="text-amber-700">{waiting}</b> · كل الطلبات (آخر ٦٠ يوم): {rows.length}</p>
      {rows.map((r) => (
        <div key={r.id} className={`rounded-xl border p-3 text-sm ${r.answeredAt && !r.verdict ? 'border-amber-300 bg-amber-50' : 'border-slate-200 bg-white'}`}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <b className="text-[#0f2040]">{r.employeeName}</b>
            <span className="text-[11px] text-slate-500">طلبه {r.monitorName || 'المراقب'} · {fmt(r.createdAt)}</span>
          </div>
          <p className="mt-1 text-slate-700">📝 {r.question}</p>
          {!r.answeredAt
            ? <p className="mt-2 text-xs font-bold text-slate-500">⏳ لسه الموظف ما رد</p>
            : <p className="mt-2 rounded-lg bg-sky-50 px-3 py-2 text-sky-900">💬 عذره: «{r.answer}» <span className="text-[11px] text-slate-500">— {fmt(r.answeredAt)}</span></p>}
          {r.verdict && (
            <p className={`mt-2 text-xs font-bold ${r.verdict === 'ACCEPTED' ? 'text-emerald-700' : 'text-red-700'}`}>
              {r.verdict === 'ACCEPTED' ? '✅ النتيجة: مقبول' : '❌ النتيجة: مرفوض'}{r.verdictNote ? ` — ${r.verdictNote}` : ''} · {fmt(r.decidedAt)}
            </p>
          )}
          {r.answeredAt && !r.verdict && (
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <input value={notes[r.id] ?? ''} onChange={(e) => setNotes((n) => ({ ...n, [r.id]: e.target.value }))} placeholder="ملاحظة للموظف (اختياري)"
                className="min-w-[200px] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-xs" />
              <button onClick={() => void decide(r.id, 'ACCEPTED')} className="rounded-lg bg-emerald-700 px-3 py-1.5 text-xs font-bold text-white">✓ مقبول</button>
              <button onClick={() => void decide(r.id, 'REJECTED')} className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-bold text-white">✗ مرفوض</button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

const ACTIONS: { key: 'NOTIFY' | 'JUSTIFY' | 'FOLLOW_UP' | 'ESCALATE'; icon: string; label: string; confirm: string }[] = [
  { key: 'NOTIFY', icon: '🔔', label: 'تنبيه الموظف', confirm: 'ترسل تنبيه للموظف؟' },
  { key: 'JUSTIFY', icon: '📝', label: 'طلب تبرير', confirm: 'تطلب من الموظف يبرّر؟' },
  { key: 'FOLLOW_UP', icon: '📅', label: 'إنشاء متابعة', confirm: 'تسوي مهمة متابعة للموظف (موعدها باچر)؟' },
  { key: 'ESCALATE', icon: '👥', label: 'إحالة للمسؤول', confirm: 'تحيلها للمدير والمالك؟' },
]
function ActionMenu({ row, note }: { row: MonitorReview; note: string }) {
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState<string | null>(null)
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const close = (e: MouseEvent) => { if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false) }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [open])
  const run = async (a: typeof ACTIONS[number]) => {
    const who = row.late?.employeeName ? ` (${row.late.employeeName})` : ''
    if (!confirm(a.confirm + who + (note.trim() ? `\n\nالملاحظة: ${note.trim()}` : ''))) return
    setBusy(true); setOpen(false)
    try { setDone((await api.monitorReviewAction(row.id, a.key, note)).done) } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر') } finally { setBusy(false) }
  }
  return (
    <div ref={ref} className="relative">
      <button type="button" disabled={busy} onClick={() => setOpen((o) => !o)}
        className="flex items-center gap-1 rounded-lg bg-[#1e3a8a] px-3 py-2 text-xs font-bold text-white disabled:opacity-50">
        ⚡ اتخاذ إجراء <span className="text-[10px]">▾</span>
      </button>
      {open && (
        <ul className="absolute bottom-full left-0 z-20 mb-1 w-44 overflow-hidden rounded-xl border border-slate-200 bg-white py-1 shadow-lg">
          {ACTIONS.map((a) => (
            <li key={a.key}><button type="button" onClick={() => run(a)} className="flex w-full items-center gap-2 px-3 py-2 text-right text-xs text-slate-700 hover:bg-slate-50"><span>{a.icon}</span>{a.label}</button></li>
          ))}
        </ul>
      )}
      {done && <p className="absolute top-full left-0 mt-1 whitespace-nowrap text-[11px] font-bold text-emerald-700">✓ {done}</p>}
    </div>
  )
}

// ═══ تفاصيل الفاتورة والحجز بجهة وحدة — (ع) 10-16 ═══
const money = (n?: number | null) => (n == null ? '—' : `${Math.round(n).toLocaleString('en-US')} د.ع`)
function InvoiceDetailBlock({ d }: { d: NonNullable<NonNullable<MonitorReview['details']>['invoice']> }) {
  const changed = d.adjustments.length > 0
  return (
    <div className="mt-2 rounded-xl border border-slate-200 bg-slate-50 p-3 text-xs">
      <p className="mb-1.5 font-extrabold text-[#0f2040]">🧾 الفاتورة {d.isFree && <span className="mr-1 rounded-full bg-violet-100 px-2 py-0.5 text-violet-700">🎁 مجانية</span>}</p>
      <div className="grid grid-cols-2 gap-x-4 gap-y-1 sm:grid-cols-4">
        <span>الليدر: <b>{d.leaderName ?? '—'}</b></span>
        <span>التنفيذ: <b>{money(d.executionCost)}</b></span>
        <span>المواد: <b>{money(d.materialsTotal)}</b></span>
        <span>الخصم: <b>{money(d.discountValue)}</b></span>
        <span>الصافي: <b className="text-[#0f2040]">{money(d.netTotal)}</b></span>
        <span>الليدر استلم من الزبون: <b className={d.collected != null && Math.round(d.collected) !== Math.round(d.netTotal) ? 'text-red-700' : ''}>{money(d.collected)}</b></span>
        <span>رقم الفاتورة المحاسبية: <b className="font-mono">{d.externalNo ?? '—'}</b></span>
      </div>
      {changed ? (
        <div className="mt-2 space-y-1">
          <p className="font-bold text-amber-800">✏️ المحاسب غيّر:</p>
          {d.adjustments.map((a, i) => (
            <p key={i} className="rounded-lg bg-amber-50 px-2 py-1 text-amber-900">
              الصافي <b>{money(a.oldNetTotal)}</b> ← <b>{money(a.newNetTotal)}</b>
              {a.oldExecutionCost !== a.newExecutionCost && <> · التنفيذ {money(a.oldExecutionCost)} ← {money(a.newExecutionCost)}</>}
              {a.oldMaterialsTotal !== a.newMaterialsTotal && <> · المواد {money(a.oldMaterialsTotal)} ← {money(a.newMaterialsTotal)}</>}
              {' '}— {a.byName ?? 'المحاسب'}: «{a.reason}»
            </p>
          ))}
        </div>
      ) : <p className="mt-1.5 text-emerald-700">المحاسب ما غيّر شي بالمبالغ.</p>}
    </div>
  )
}

const SHIFT_AR: Record<string, string> = { MORNING: 'صباحي', EVENING: 'مسائي', NIGHT: 'ليلي' }
function BookingDetailBlock({ d }: { d: NonNullable<NonNullable<MonitorReview['details']>['booking']> }) {
  const at = (s: string | null) => (s ? new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' }) : null)
  const daysTo = d.scheduledAt ? Math.round((new Date(d.scheduledAt).setHours(0, 0, 0, 0) - new Date().setHours(0, 0, 0, 0)) / 86400000) : null
  // سطر الحالة — «حجز جديد ومحد تواصل وياه»، «مثبّت ومحدد بعد يومين»…
  let headline = d.station
  if (!d.confirmedAt && !d.contactedAt && d.status !== 'COMPLETED') headline = '🆕 حجز جديد — محد تواصل ويا الزبون لسه'
  else if (d.status === 'COMPLETED') headline = '🏁 تم الإنجاز'
  else if (d.scheduledAt && daysTo != null) headline = `📌 ${d.station} — ${daysTo === 0 ? 'موعده اليوم' : daysTo === 1 ? 'موعده باچر' : daysTo > 1 ? `موعده بعد ${daysTo} أيام` : `موعده فات من ${-daysTo} يوم`}`
  const steps: [string, string | null][] = [
    ['انسجّل', at(d.createdAt)], ['تواصل', at(d.contactedAt)], ['انثبّت', at(d.confirmedAt)],
    ['الموعد', at(d.scheduledAt)], ['وصل', at(d.arrivedAt)], ['بدا', at(d.startedAt)], ['خلص', at(d.completedAt)],
  ]
  return (
    <div className="mt-2 rounded-xl border border-slate-200 bg-slate-50 p-3 text-xs">
      <p className="mb-1.5 font-extrabold text-[#0f2040]">{headline}</p>
      <div className="grid grid-cols-2 gap-x-4 gap-y-1 sm:grid-cols-4">
        <span>الخدمة: <b>{d.service ?? '—'}</b></span>
        <span>سجّله: <b>{d.createdByName ?? '—'}</b></span>
        <span>ثبّته: <b>{d.confirmedByName ?? '—'}</b></span>
        <span>الشفت: <b>{d.shift ? (SHIFT_AR[d.shift] ?? d.shift) : '—'}</b></span>
        <span>الليدر: <b>{d.leaderName ?? '—'}</b></span>
        <span className="sm:col-span-2">الكادر: <b>{d.crew ?? '—'}</b></span>
        <span>المستلم: <b>{money(d.collected)}</b></span>
        {d.quality && <span className="sm:col-span-4">تقييم الجودة: <b>{d.quality}</b></span>}
      </div>
      <div className="mt-2 flex flex-wrap gap-1">
        {steps.map(([k, v]) => (
          <span key={k} className={`rounded-full px-2 py-0.5 ${v ? 'bg-emerald-50 text-emerald-800' : 'bg-white text-slate-400'}`}>{v ? '✓' : '○'} {k}{v ? `: ${v}` : ''}</span>
        ))}
      </div>
    </div>
  )
}
