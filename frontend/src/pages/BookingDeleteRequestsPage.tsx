import { Fragment, useEffect, useMemo, useState } from 'react'
import {
  api, bookingDeleteChannelLabels, bookingDeleteTypeLabels,
  type BookingDeleteChannel, type BookingDeleteRequest, type BookingDeleteRequestCounts, type BookingDeleteRequestType,
} from '../api'
import { matches } from '../utils/search'
import { useSaveGuard } from '../useSaveGuard'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'
import SearchBar from '../components/SearchBar'
import SaveError from '../components/SaveError'
import Pager from '../components/Pager'
import RowActions from '../components/RowActions'

/**
 * طلبات حذف الحجوزات.
 *
 * الإداري يطلب من صفحة الحجوزات، والمراقب أو مدير النظام يبت هنا.
 * الحذف ما يترد — فالموافقة تؤرشف الحجز لا تمحيه فعلياً، وما تصير
 * بضغطة زر وحدة بلا سبب.
 *
 * ⚠️ «معلقة» تصنيف فرعي داخل PENDING — الطلب ناقصه معلومات ولسه
 * ما انبتّ فيه، مو حالة رابعة منفصلة.
 */

const STATUS_STYLE: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-800',
  APPROVED: 'bg-red-100 text-red-800',
  REJECTED: 'bg-slate-200 text-slate-600',
}

type TileFilter = 'ALL' | 'APPROVED' | 'NEEDS_INFO' | 'AWAITING' | 'REJECTED'
type SortOrder = 'newest' | 'oldest'

function fmtDate(iso: string) {
  return new Date(iso).toLocaleString('ar-IQ', { dateStyle: 'short', timeStyle: 'short' })
}

export default function BookingDeleteRequestsPage() {
  const [rows, setRows] = useState<BookingDeleteRequest[]>([])
  const [counts, setCounts] = useState<BookingDeleteRequestCounts | null>(null)
  const [loading, setLoading] = useState(true)
  const guard = useSaveGuard()

  const [query, setQuery] = useState('')
  const [tileFilter, setTileFilter] = useState<TileFilter>('ALL')
  const [sort, setSort] = useState<SortOrder>('newest')
  const [dateFilter, setDateFilter] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [channelFilter, setChannelFilter] = useState<BookingDeleteChannel | ''>('')
  const [typeFilter, setTypeFilter] = useState<BookingDeleteRequestType | ''>('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)

  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [noteFor, setNoteFor] = useState<string | null>(null)
  const [note, setNote] = useState('')
  const [infoFor, setInfoFor] = useState<string | null>(null)
  const [infoNote, setInfoNote] = useState('')

  const load = () => {
    Promise.all([api.getBookingDeleteRequests(), api.getBookingDeleteRequestCounts()])
      .then(([r, c]) => { setRows(r); setCounts(c) })
      .catch(() => {})
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  // ⚠️ أي مرشّح يتغيّر يرجّع الصفحة لواحد، وإلا يبقى المستخدم بصفحة
  // متأخرة وأمامه نتيجة ترشيح قليلة فيظن ماكو نتائج.
  const setFilter = <T,>(setter: (v: T) => void) => (v: T) => { setter(v); setPage(1) }

  const decide = async (r: BookingDeleteRequest, approve: boolean) => {
    if (approve && !confirm(`تأكيد حذف الحجز ${r.bookingCode}؟ الحذف ما يترد.`)) return
    if (!approve && noteFor !== r.id) { setNoteFor(r.id); setNote(''); return }
    const ok = await guard.run(approve ? 'تنفيذ الحذف' : 'رفض الطلب', () =>
      api.decideBookingDelete(r.id, approve, note || undefined))
    if (ok) { setNoteFor(null); setNote(''); load() }
  }

  const submitNeedsInfo = async (r: BookingDeleteRequest) => {
    if (infoFor !== r.id) { setInfoFor(r.id); setInfoNote(''); return }
    if (!infoNote.trim()) return
    const ok = await guard.run('تعليم الطلب ناقص معلومات', () =>
      api.setBookingDeleteNeedsInfo(r.id, infoNote.trim()))
    if (ok) { setInfoFor(null); setInfoNote(''); load() }
  }

  const filtered = useMemo(() => {
    const out = rows
      .filter((r) => {
        if (tileFilter === 'ALL') return true
        if (tileFilter === 'APPROVED') return r.status === 'APPROVED'
        if (tileFilter === 'REJECTED') return r.status === 'REJECTED'
        if (tileFilter === 'NEEDS_INFO') return r.status === 'PENDING' && r.needsInfo
        if (tileFilter === 'AWAITING') return r.status === 'PENDING' && !r.needsInfo
        return true
      })
      .filter((r) => !channelFilter || r.channel === channelFilter)
      .filter((r) => !typeFilter || r.requestType === typeFilter)
      .filter((r) => !dateFilter || r.createdAt.slice(0, 10) === dateFilter)
      .filter((r) => !query.trim() || matches([r.bookingCode, r.customerName], query))
    out.sort((a, b) => {
      const da = new Date(a.createdAt).getTime()
      const db = new Date(b.createdAt).getTime()
      return sort === 'newest' ? db - da : da - db
    })
    return out
  }, [rows, tileFilter, channelFilter, typeFilter, dateFilter, query, sort])

  const start = (page - 1) * perPage
  const paged = filtered.slice(start, start + perPage)

  return (
    <div dir="rtl" className="space-y-4">
      <SaveError message={guard.error} onClose={guard.clear} />

      <PageHeader
        title="🗑️ طلبات حذف الحجوزات"
        subtitle="الحجوزات التجريبية والملغاة تنتقل لموافقة لحذف الحجز نهائياً وما ترجع."
      />

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-5">
        <StatTile label="موافق عليها" icon="✅" tone="success" value={counts?.approved ?? '—'}
          onClick={() => setFilter(setTileFilter)('APPROVED')} />
        <StatTile label="معلقة" icon="ℹ️" tone="warning" value={counts?.needsInfo ?? '—'}
          hint="تحتاج معلومات ناقصة" onClick={() => setFilter(setTileFilter)('NEEDS_INFO')} />
        <StatTile label="بانتظار الموافقة" icon="⏳" tone="info" value={counts?.awaitingReview ?? '—'}
          onClick={() => setFilter(setTileFilter)('AWAITING')} />
        <StatTile label="مرفوضة" icon="✕" tone="danger" value={counts?.rejected ?? '—'}
          onClick={() => setFilter(setTileFilter)('REJECTED')} />
        <StatTile label="إجمالي الطلبات" icon="📄" tone="default" value={counts?.total ?? '—'}
          onClick={() => setFilter(setTileFilter)('ALL')} />
      </div>

      <div className="space-y-2">
        <SearchBar value={query} onChange={setFilter(setQuery)} placeholder="ابحث برقم الحجز أو اسم العميل...">
          <select value={sort} onChange={(e) => setSort(e.target.value as SortOrder)}
            className="rounded-xl border px-3 py-2.5 text-sm" style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
            <option value="newest">الأحدث أولاً</option>
            <option value="oldest">الأقدم أولاً</option>
          </select>
          <input type="date" value={dateFilter} onChange={(e) => setFilter(setDateFilter)(e.target.value)}
            title="تاريخ الطلب"
            className="rounded-xl border px-3 py-2.5 text-sm" style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }} />
          <button onClick={() => setShowAdvanced((v) => !v)}
            className={`rounded-xl border px-4 py-2.5 text-sm font-medium ${
              showAdvanced ? 'border-brand-300 bg-brand-50 text-brand-800' : ''}`}
            style={showAdvanced ? undefined : { borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
            ⛃ فلترة متقدمة
          </button>
        </SearchBar>

        {showAdvanced && (
          <div className="flex flex-wrap gap-2 rounded-xl border p-3" style={{ borderColor: 'var(--bd-line)' }}>
            <select value={channelFilter} onChange={(e) => setFilter(setChannelFilter)(e.target.value as BookingDeleteChannel | '')}
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs">
              <option value="">كل القنوات</option>
              {(Object.entries(bookingDeleteChannelLabels) as [BookingDeleteChannel, string][]).map(([k, l]) => (
                <option key={k} value={k}>{l}</option>
              ))}
            </select>
            <select value={typeFilter} onChange={(e) => setFilter(setTypeFilter)(e.target.value as BookingDeleteRequestType | '')}
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs">
              <option value="">كل الأنواع</option>
              {(Object.entries(bookingDeleteTypeLabels) as [BookingDeleteRequestType, string][]).map(([k, l]) => (
                <option key={k} value={k}>{l}</option>
              ))}
            </select>
          </div>
        )}
      </div>

      {loading ? (
        <p className="text-center" style={{ color: 'var(--t-faint)' }}>جاري التحميل...</p>
      ) : (
        <div className="overflow-hidden rounded-2xl border" style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
          {filtered.length === 0 ? (
            <EmptyState
              icon={rows.length === 0 ? '📭' : '🔍'}
              title={rows.length === 0 ? 'ماكو طلبات حذف بعد' : 'ماكو طلب يطابق البحث أو الفلترة'}
              reason={rows.length === 0
                ? 'الطلبات تنسجّل من شاشة الحجوزات أو التنسيق، وتظهر هنا للبتّ فيها.'
                : `اكو ${rows.length} طلب بالمجموع — امسح البحث أو الفلترة.`} />
          ) : (
            <>
              <div className="divide-y" style={{ borderColor: 'var(--bd-line)' }}>
                {paged.map((r) => {
                  const isOpen = expandedId === r.id
                  const decided = r.status !== 'PENDING'
                  return (
                    <Fragment key={r.id}>
                      <div className="p-4">
                        <div className="flex flex-wrap items-start justify-between gap-3">
                          <div className="min-w-0">
                            <p className="font-bold" style={{ color: 'var(--t-title)' }}>
                              {r.bookingCode}
                              <span className="mr-2 text-sm font-normal" style={{ color: 'var(--t-muted)' }}>{r.customerName}</span>
                            </p>
                            <p className="mt-1 text-sm" style={{ color: 'var(--t-body)' }}>📝 {r.reason}</p>
                            <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs" style={{ color: 'var(--t-faint)' }}>
                              <span>🌐 {r.channelLabel || '—'}</span>
                              <span>🏷️ {r.requestTypeLabel || '—'}</span>
                              <span>🕓 {fmtDate(r.createdAt)}</span>
                              <span>طلبه: {r.requestedByName}</span>
                            </div>
                            {r.status === 'PENDING' && r.needsInfo && (
                              <p className="mt-1 text-xs font-bold text-amber-700">
                                ℹ️ معلقة — {r.needsInfoNote}
                                {r.needsInfoByName && ` (${r.needsInfoByName})`}
                              </p>
                            )}
                          </div>
                          <div className="flex shrink-0 items-center gap-2">
                            <span className={`rounded-full px-3 py-1 text-xs font-bold ${STATUS_STYLE[r.status]}`}>
                              {r.status === 'PENDING' && r.needsInfo ? 'معلقة' : r.statusLabel}
                            </span>
                            {decided && (
                              <RowActions actions={[
                                {
                                  label: isOpen ? 'اخفي التفاصيل' : 'عرض التفاصيل',
                                  icon: isOpen ? '▲' : '▼',
                                  onClick: () => setExpandedId(isOpen ? null : r.id),
                                },
                              ]} />
                            )}
                          </div>
                        </div>

                        {r.status === 'PENDING' && (
                          <div className="mt-3 space-y-2">
                            {noteFor === r.id && (
                              <input
                                value={note} onChange={(e) => setNote(e.target.value)} placeholder="سبب الرفض (اختياري)"
                                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-right text-sm outline-none focus:border-brand-500"
                              />
                            )}
                            {infoFor === r.id && (
                              <input
                                value={infoNote} onChange={(e) => setInfoNote(e.target.value)} placeholder="شنو المعلومة الناقصة؟"
                                className="w-full rounded-lg border border-amber-300 px-3 py-2 text-right text-sm outline-none focus:border-amber-500"
                              />
                            )}
                            <div className="flex flex-wrap gap-2">
                              <button disabled={guard.busy} onClick={() => decide(r, true)}
                                className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50">
                                🗑️ وافق واحذف
                              </button>
                              <button disabled={guard.busy} onClick={() => decide(r, false)}
                                className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-50">
                                {noteFor === r.id ? 'تأكيد الرفض' : '✕ رفض الطلب'}
                              </button>
                              <button disabled={guard.busy} onClick={() => submitNeedsInfo(r)}
                                className="rounded-lg border border-amber-300 px-4 py-2 text-sm font-medium text-amber-800 hover:bg-amber-50 disabled:opacity-50">
                                {infoFor === r.id ? 'تأكيد الطلب' : 'ℹ️ اطلب معلومات'}
                              </button>
                            </div>
                          </div>
                        )}
                      </div>

                      {isOpen && decided && (
                        <div className="px-4 pb-4">
                          <div className="rounded-xl p-3 text-xs" style={{ backgroundColor: 'var(--sf-sunken)', color: 'var(--t-muted)' }}>
                            <p>القرار: <b style={{ color: 'var(--t-title)' }}>{r.statusLabel}</b></p>
                            <p className="mt-1">البتّ بواسطة: <b style={{ color: 'var(--t-title)' }}>{r.decidedByName ?? '—'}</b></p>
                            {r.decisionNote && <p className="mt-1">الملاحظة: {r.decisionNote}</p>}
                            {r.needsInfo && (
                              <p className="mt-1">
                                طُلبت معلومات إضافية سابقاً: {r.needsInfoNote}
                                {r.needsInfoByName && ` (${r.needsInfoByName})`}
                              </p>
                            )}
                            <p className="mt-1">حالة الحجز الحالية: {r.bookingStatus}</p>
                          </div>
                        </div>
                      )}
                    </Fragment>
                  )
                })}
              </div>
              <div className="border-t px-4 py-3" style={{ borderColor: 'var(--bd-line)' }}>
                <Pager page={page} perPage={perPage} total={filtered.length} unit="طلب"
                  onPage={setPage} onPerPage={setPerPage} />
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}
