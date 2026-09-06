import { Fragment, useMemo, useState } from 'react'
import {
  api, auditVerdictLabels, eventKindLabels,
  type AuditVerdict, type Complaint, type ComplaintCustomerStat,
  type ComplaintEvent, type Employee,
} from '../api'
import { matches } from '../utils/search'
import { useSaveGuard } from '../useSaveGuard'
import PageHeader from './PageHeader'
import StatTile from './StatTile'
import EmptyState from './EmptyState'
import SearchBar from './SearchBar'
import SaveError from './SaveError'
import Pager from './Pager'
import RowActions from './RowActions'

// ═══ متابعة الشكاوى وتدقيق مهندسي الجودة ═══
//
// صف لكل **زبون** مو لكل شكوى: الزبون الي اشتكى أربع مرات مشكلة
// وحدة، وأربعة صفوف متفرقة تخفي هذا.
//
// ⚠️ وبما إن الصف مجمّع، «حالة الشكوى» و«مهندس الجودة» يعودان
// لآخر شكوى — ومكتوبة صراحةً بالعمود. بلا هذا تصير قراءتان
// مختلفتان لنفس السطر: «٤ شكاوى» وحالة وحدة بلا ما يبين لأيّتهن.

const statusLabels: Record<Complaint['status'], string> = {
  NEW: 'جديدة', IN_PROGRESS: 'قيد المعالجة', RESOLVED: 'تم الحل', CLOSED: 'مغلقة',
}

const VERDICT_TONE: Record<AuditVerdict, string> = {
  NEEDS_FOLLOWUP: 'bg-amber-100 text-amber-800',
  RECHECK_RATING: 'bg-violet-100 text-violet-800',
  APPROVED: 'bg-emerald-100 text-emerald-700',
}

/** نجوم للقراءة السريعة — والرقم جنبها لأن النجوم وحدها ما تنقرا بدقة. */
function Stars({ value }: { value: number | null }) {
  if (value == null) {
    // ⚠️ «—» مو «٠»: ماكو تقييم **مو** تقييم واطي. الصفر يظلم المهندس.
    return <span className="text-slate-400" title="ما انسجّل تقييم">—</span>
  }
  const full = Math.round(value)
  return (
    <span className="whitespace-nowrap">
      <b className="text-slate-700">{value.toFixed(1)}</b>
      <span className="text-slate-400">/5</span>{' '}
      <span className="text-amber-500">{'★'.repeat(full)}{'☆'.repeat(5 - full)}</span>
    </span>
  )
}

export default function ComplaintsTracking({
  complaints, stats, employees, canContact, canAudit, canSeeFullDetail,
  busyId, onContact, onNotes, onAssign, onStatus, onResolveClick, onAudited,
}: {
  complaints: Complaint[]
  stats: ComplaintCustomerStat[]
  employees: Employee[]
  canContact: boolean
  canAudit: boolean
  canSeeFullDetail: boolean
  busyId: string | null
  onContact: (c: Complaint, contacted: boolean, rating?: number | null) => void
  onNotes: (c: Complaint, value: string) => void
  onAssign: (complaintId: string, employeeId: string) => void
  onStatus: (complaintId: string, status: Complaint['status']) => void
  onResolveClick: (id: string) => void
  onAudited: () => void
}) {
  // ⚠️ التبويب **عرض مو صلاحية**: تبويب التدقيق ما يطلع لمهندس
  // الجودة أصلاً، والخادم يرفض حكمه حتى لو انفتح بالإيد.
  const [tab, setTab] = useState<'audit' | 'direct'>(canAudit ? 'audit' : 'direct')
  const [query, setQuery] = useState('')
  const [onlyPending, setOnlyPending] = useState(false)
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [panelOpen, setPanelOpen] = useState(true)
  // ⚠️ الملخص يتمدّد **جوّا الصف**. چانت التفاصيل تنفتح بلوحة
  // أسفل الصفحة بس، فالضغط على صف بعيد يخلي المستخدم ينزل
  // يدوّرها — نفس الشكوى الي انصلّحت بشاشة الكي بي اي.
  const [expandedCustomer, setExpandedCustomer] = useState<string | null>(null)
  const [events, setEvents] = useState<ComplaintEvent[] | null>(null)
  const [rating, setRating] = useState<string>('')
  const [auditNote, setAuditNote] = useState('')
  const guard = useSaveGuard()

  const byCustomer = useMemo(() => {
    const m = new Map<string, Complaint[]>()
    for (const c of complaints) {
      const k = c.customer?.id
      if (!k) continue
      m.set(k, [...(m.get(k) ?? []), c])
    }
    return m
  }, [complaints])

  // ⚠️ المتوسط من التقييمات **الموجودة** بس — الفاضيات ما تنحسب
  // صفراً. ولمن ماكو ولا تقييم يبقى null فتنعرض «—».
  const overallRating = useMemo(() => {
    const xs = complaints.map((c) => c.customerRating).filter((r): r is number => r != null)
    return xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : null
  }, [complaints])

  const openCount = complaints.filter((c) => c.status === 'NEW' || c.status === 'IN_PROGRESS').length
  const waitingContact = complaints.filter(
    (c) => !c.contactedAt && (c.status === 'NEW' || c.status === 'IN_PROGRESS')).length
  const contacted30 = stats.reduce((n, s) => n + s.contactedLast30, 0)

  const shown = stats
    .filter((s) => !onlyPending || s.notContactedCount > 0)
    .filter((s) => !query.trim() || matches([s.customerName, s.customerPhone], query))
  const start = (page - 1) * perPage
  const paged = shown.slice(start, start + perPage)

  const selected = complaints.find((c) => c.id === selectedId) ?? null
  const siblings = selected?.customer?.id ? (byCustomer.get(selected.customer.id) ?? []) : []

  const openRow = (s: ComplaintCustomerStat) => {
    const list = byCustomer.get(s.customerId) ?? []
    const target = s.latestComplaintId ?? list[0]?.id ?? null
    setSelectedId(target)
    setPanelOpen(true)
    setEvents(null)
    const c = complaints.find((x) => x.id === target)
    setRating(c?.customerRating ? String(c.customerRating) : '')
    setAuditNote(c?.auditNote ?? '')
  }

  const loadLog = async () => {
    if (!selected) return
    const ev = await guard.run('جلب السجل', () => api.getComplaintEvents(selected.id))
    if (ev) setEvents(ev)
  }

  const submitAudit = async (verdict: AuditVerdict) => {
    if (!selected) return
    const ok = await guard.run('حفظ التدقيق', () =>
      api.auditComplaint(selected.id, verdict, auditNote))
    if (ok) { onAudited(); setEvents(null) }
  }

  return (
    <div className="mt-6 space-y-4">
      <SaveError message={guard.error} onClose={guard.clear} />

      <PageHeader
        title="شكاوى الصيانة"
        subtitle="إدارة شكاوى العملاء ومتابعة الجودة وحالة المعالجة"
      />

      {/* التبويبان + سطر يشرح شنو تسوي بهذا الوضع */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex gap-2">
          {canAudit && (
            <button onClick={() => setTab('audit')}
              className={`rounded-xl px-5 py-2.5 text-sm font-bold transition ${
                tab === 'audit' ? 'bg-[#0f2040] text-white shadow' : 'border bg-white text-slate-600'}`}>
              🛡️ تدقيق مهندسي الجودة
            </button>
          )}
          <button onClick={() => setTab('direct')}
            className={`rounded-xl px-5 py-2.5 text-sm font-bold transition ${
              tab === 'direct' ? 'bg-[#0f2040] text-white shadow' : 'border bg-white text-slate-600'}`}>
            👤 متابعة الجودة المباشرة
          </button>
        </div>
        <p className="max-w-md rounded-xl border px-3 py-2 text-[11px] leading-relaxed"
          style={{ borderColor: 'var(--bd-line)', color: 'var(--t-muted)' }}>
          {tab === 'audit'
            ? 'ⓘ عرض شكاوى العملاء للتدقّق من تواصل مهندسي الجودة وتقييم الأداء. افتح أي شكوى لمراجعة التفاصيل والقيام بالإجراءات اللازمة.'
            : 'ⓘ تواصل مع الزبون، سجّل تقييمه من ٥، واكتب ملاحظاته. التقييم ينحفظ مع تأشيرة التواصل بنفس الخطوة.'}
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatTile label="متوسط التقييم" icon="⭐" tone="violet"
          value={overallRating == null ? '—' : `${overallRating.toFixed(1)} / 5`}
          hint={overallRating == null ? 'ما انسجّل ولا تقييم بعد' : 'من التقييمات المسجّلة'} />
        <StatTile label="بانتظار التواصل" icon="🕐" tone="warning" value={waitingContact}
          hint="من الشكاوى المفتوحة" onClick={() => { setOnlyPending(!onlyPending); setPage(1) }} />
        <StatTile label="تم التواصل" icon="📞" tone="success" value={contacted30}
          hint="خلال آخر ٣٠ يوم" />
        <StatTile label="الشكاوى المفتوحة" icon="💬" tone="info" value={openCount}
          hint="شكاوى تحتاج متابعة" />
      </div>

      <SearchBar value={query} onChange={(v) => { setQuery(v); setPage(1) }}
        placeholder="ابحث باسم الزبون أو الهاتف...">
        <button onClick={() => { setOnlyPending(!onlyPending); setPage(1) }}
          className={`rounded-xl border px-4 py-2.5 text-sm font-medium ${
            onlyPending ? 'border-amber-300 bg-amber-50 text-amber-800' : ''}`}
          style={onlyPending ? undefined : { borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
          ⛃ {onlyPending ? 'الي ينتظرون تواصل فقط' : 'فلترة'}
        </button>
      </SearchBar>

      <div className="overflow-hidden rounded-2xl border"
        style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
        {shown.length === 0 ? (
          <EmptyState
            icon={stats.length === 0 ? '📭' : '🔍'}
            title={stats.length === 0 ? 'ماكو شكاوى بعد' : 'ماكو زبون يطابق البحث أو الفلترة'}
            reason={stats.length === 0
              ? 'الشكاوى تنسجّل من شاشة تسجيل الشكوى، وتظهر هنا بصف لكل زبون.'
              : `اكو ${stats.length} زبون بالمجموع — امسح البحث أو الفلترة.`} />
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right text-sm">
                <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                  <tr>
                    <th className="px-3 py-3 font-semibold">الزبون</th>
                    <th className="px-3 py-3 font-semibold">الهاتف</th>
                    <th className="px-3 py-3 font-semibold">عدد الشكاوى</th>
                    <th className="px-3 py-3 font-semibold">حالة آخر شكوى</th>
                    <th className="px-3 py-3 font-semibold">مهندس الجودة</th>
                    <th className="px-3 py-3 font-semibold">تم التواصل؟</th>
                    <th className="px-3 py-3 font-semibold">وقت آخر تواصل</th>
                    <th className="px-3 py-3 font-semibold">تقييم الزبون</th>
                    <th className="px-3 py-3 font-semibold">متابعة المدقق</th>
                    <th className="px-3 py-3 font-semibold">إجراءات</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {paged.map((s) => {
                    const latest = complaints.find((c) => c.id === s.latestComplaintId)
                    const v = latest?.auditVerdict ?? null
                    const isSel = selected?.customer?.id === s.customerId
                    const rowOpen = expandedCustomer === s.customerId
                    const list = byCustomer.get(s.customerId) ?? []
                    return (
                    <Fragment key={s.customerId}>
                      <tr
                        onClick={() => { setExpandedCustomer(rowOpen ? null : s.customerId); openRow(s) }}
                        className={`cursor-pointer transition-colors hover:bg-slate-50 ${
                          isSel ? 'bg-brand-50/60' : ''}`}>
                        <td className="px-3 py-2.5 font-medium">{s.customerName}</td>
                        <td className="px-3 py-2.5 text-slate-500">{s.customerPhone}</td>
                        <td className="px-3 py-2.5 font-bold text-brand-700">{s.complaintCount}</td>
                        <td className="px-3 py-2.5">
                          <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-700">
                            {s.latestStatus ? statusLabels[s.latestStatus] : '—'}
                          </span>
                        </td>
                        <td className="px-3 py-2.5 text-slate-600">{s.latestEngineer || '—'}</td>
                        <td className="px-3 py-2.5">
                          {s.notContactedCount > 0 ? (
                            <span className="font-bold text-red-600">
                              ✕ لا{s.notContactedCount > 1 ? ` (${s.notContactedCount})` : ''}
                            </span>
                          ) : <span className="font-bold text-emerald-600">✓ نعم</span>}
                        </td>
                        <td className="px-3 py-2.5 text-slate-500">
                          {s.lastContactAt
                            ? new Date(s.lastContactAt).toLocaleString('ar-IQ', {
                                dateStyle: 'short', timeStyle: 'short' })
                            : '—'}
                        </td>
                        <td className="px-3 py-2.5"><Stars value={s.avgRating} /></td>
                        <td className="px-3 py-2.5">
                          {v ? (
                            <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${VERDICT_TONE[v]}`}>
                              {auditVerdictLabels[v]}
                            </span>
                          ) : (
                            <span className="rounded-full bg-amber-100 px-2.5 py-1 text-xs font-bold text-amber-800">
                              مطلوب متابعة
                            </span>
                          )}
                        </td>
                        <td className="px-3 py-2.5">
                          {/* ⚠️ چان هذا الزر يستدعي openRow() بس — يعني
                              يفتح لوحة بأسفل الصفحة، فالمستخدم يضغط
                              وما يشوف شي يتحرّك ويظن الزر ميّت. */}
                          <RowActions actions={[
                            {
                              label: rowOpen ? 'اخفي الملخص' : 'الملخص هنا',
                              icon: rowOpen ? '▲' : '▼',
                              onClick: () => { setExpandedCustomer(rowOpen ? null : s.customerId); openRow(s) },
                            },
                            {
                              label: 'التفاصيل الكاملة',
                              icon: '🗂️',
                              onClick: () => {
                                openRow(s)
                                setPanelOpen(true)
                                // ننزّله للوحة بدل ما ندور عليها.
                                setTimeout(() => document
                                  .getElementById('complaint-detail-panel')
                                  ?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 60)
                              },
                            },
                            {
                              label: 'عرض السجل',
                              icon: '📄',
                              onClick: () => {
                                openRow(s)
                                setPanelOpen(true)
                                const id = s.latestComplaintId ?? list[0]?.id
                                if (id) api.getComplaintEvents(id).then(setEvents).catch(() => setEvents([]))
                                setTimeout(() => document
                                  .getElementById('complaint-detail-panel')
                                  ?.scrollIntoView({ behavior: 'smooth', block: 'center' }), 60)
                              },
                            },
                          ]} />
                        </td>
                      </tr>

                      {rowOpen && (
                        <tr>
                          <td colSpan={10} className="bg-slate-50 px-4 py-3">
                            <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs">
                              <span style={{ color: 'var(--t-muted)' }}>
                                شكاوى الزبون: <b style={{ color: 'var(--t-title)' }}>{s.complaintCount}</b>
                              </span>
                              <span style={{ color: 'var(--t-muted)' }}>
                                مفتوحة: <b style={{ color: 'var(--t-title)' }}>{s.openCount}</b>
                              </span>
                              <span style={{ color: 'var(--t-muted)' }}>
                                ما انتصل بيها: <b className={s.notContactedCount > 0 ? 'text-red-600' : 'text-emerald-600'}>
                                  {s.notContactedCount}</b>
                              </span>
                              <span style={{ color: 'var(--t-muted)' }}>
                                تحتاج تدقيق: <b style={{ color: 'var(--t-title)' }}>{s.needsAuditCount}</b>
                              </span>
                              {latest && (
                                <span className="truncate" style={{ color: 'var(--t-muted)' }}>
                                  آخر شكوى: <b style={{ color: 'var(--t-body)' }}>{latest.description}</b>
                                </span>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                    )
                  })}
                </tbody>
              </table>
            </div>
            <div className="border-t px-4 py-3" style={{ borderColor: 'var(--bd-line)' }}>
              <Pager page={page} perPage={perPage} total={shown.length} unit="زبون"
                onPage={setPage} onPerPage={setPerPage} />
            </div>
          </>
        )}
      </div>

      {selected && (
        <div id="complaint-detail-panel" className="overflow-hidden rounded-2xl border"
          style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
          <div className="flex flex-wrap items-center justify-between gap-2 border-b px-4 py-3"
            style={{ borderColor: 'var(--bd-line)' }}>
            <div className="flex items-center gap-2">
              <button onClick={() => setPanelOpen(!panelOpen)}
                className="text-slate-400 transition-transform"
                style={{ transform: panelOpen ? 'rotate(0deg)' : 'rotate(180deg)' }}>▲</button>
              <h3 className="text-sm font-bold" style={{ color: 'var(--t-title)' }}>
                تفاصيل الشكوى المحددة
              </h3>
            </div>
            <span className={`rounded-full px-3 py-1 text-xs font-bold ${
              selected.contactedAt ? 'bg-emerald-100 text-emerald-700' : 'bg-red-100 text-red-700'}`}>
              {selected.contactedAt ? 'تم التواصل' : 'لم يتم التواصل'}
            </span>
          </div>

          {panelOpen && (
            <div className="space-y-3 p-4">
              {siblings.length > 1 && (
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-xs" style={{ color: 'var(--t-muted)' }}>
                    عند الزبون {siblings.length} شكاوى:
                  </span>
                  {siblings.map((c, i) => (
                    <button key={c.id}
                      onClick={() => { setSelectedId(c.id); setEvents(null)
                        setRating(c.customerRating ? String(c.customerRating) : '')
                        setAuditNote(c.auditNote ?? '') }}
                      className={`rounded-lg px-3 py-1 text-xs font-bold ${
                        c.id === selected.id ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>
                      #{i + 1} · {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                    </button>
                  ))}
                </div>
              )}

              <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
                <Box label="الزبون">
                  <p className="font-bold" style={{ color: 'var(--t-title)' }}>{selected.customer?.name}</p>
                  <p className="text-xs" style={{ color: 'var(--t-muted)' }}>📞 {selected.customer?.phone}</p>
                  <p className="mt-1 text-xs" style={{ color: 'var(--t-faint)' }}>{selected.description}</p>
                </Box>
                <Box label="مهندس الجودة المسؤول">
                  <p style={{ color: 'var(--t-body)' }}>{selected.assignedToEmployee?.name || '— ما انكلّف أحد'}</p>
                  {canSeeFullDetail && (
                    <select value={selected.assignedToEmployeeId ?? ''}
                      onChange={(e) => onAssign(selected.id, e.target.value)}
                      className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1 text-xs">
                      <option value="">— اختر —</option>
                      {employees.filter((e) => e.status === 'ACTIVE')
                        .map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                    </select>
                  )}
                </Box>
                <Box label="هل تواصل مع الزبون؟">
                  <p className={`font-bold ${selected.contactedAt ? 'text-emerald-600' : 'text-red-600'}`}>
                    {selected.contactedAt ? `✓ نعم — ${selected.contactedByName ?? ''}` : '✕ لا'}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--t-faint)' }}>
                    {selected.contactedAt
                      ? new Date(selected.contactedAt).toLocaleString('ar-IQ')
                      : 'ما انتصل بيه بعد'}
                  </p>
                </Box>
                <Box label="نتيجة التقييم">
                  <Stars value={selected.customerRating} />
                  {canContact && (
                    <div className="mt-2 flex items-center gap-1">
                      <select value={rating} onChange={(e) => setRating(e.target.value)}
                        className="rounded-lg border border-slate-300 px-2 py-1 text-xs">
                        <option value="">بلا تقييم</option>
                        {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n} / 5</option>)}
                      </select>
                      <button
                        disabled={busyId === selected.id}
                        onClick={() => onContact(selected, true, rating ? Number(rating) : null)}
                        className="rounded-lg bg-emerald-600 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">
                        سجّل التواصل والتقييم
                      </button>
                    </div>
                  )}
                </Box>
              </div>

              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <Box label="ملاحظات الزبون">
                  <textarea defaultValue={selected.notes ?? ''} rows={2}
                    disabled={!canContact}
                    onBlur={(e) => onNotes(selected, e.target.value)}
                    placeholder="شنو كال الزبون؟"
                    className="w-full rounded-lg border border-slate-300 px-2 py-1 text-xs disabled:bg-slate-50" />
                </Box>
                <Box label="ملاحظات المدقق">
                  {canAudit ? (
                    <>
                      <textarea value={auditNote} rows={2}
                        onChange={(e) => setAuditNote(e.target.value)}
                        placeholder="ملاحظتك على شغل مهندس الجودة..."
                        className="w-full rounded-lg border border-slate-300 px-2 py-1 text-xs" />
                      <div className="mt-1 flex flex-wrap gap-1">
                        {(Object.keys(auditVerdictLabels) as AuditVerdict[]).map((v) => (
                          <button key={v} onClick={() => submitAudit(v)} disabled={guard.busy}
                            className={`rounded-lg px-3 py-1 text-xs font-bold disabled:opacity-50 ${VERDICT_TONE[v]}`}>
                            {auditVerdictLabels[v]}
                          </button>
                        ))}
                      </div>
                    </>
                  ) : (
                    <p className="text-xs" style={{ color: 'var(--t-muted)' }}>
                      {selected.auditNote || '— ماكو ملاحظة تدقيق'}
                      {selected.auditedBy && (
                        <span className="block text-[11px]" style={{ color: 'var(--t-faint)' }}>
                          دقّقها {selected.auditedBy.name}
                        </span>
                      )}
                    </p>
                  )}
                </Box>
              </div>

              <div className="flex flex-wrap items-center justify-between gap-2 rounded-xl px-3 py-2"
                style={{ backgroundColor: 'var(--sf-sunken)' }}>
                <p className="text-xs" style={{ color: 'var(--t-muted)' }}>
                  آخر إجراء:{' '}
                  <b style={{ color: 'var(--t-body)' }}>
                    {selected.resolvedAt ? 'انحلّت'
                      : selected.auditedAt ? `تدقيق — ${auditVerdictLabels[selected.auditVerdict as AuditVerdict] ?? ''}`
                      : selected.contactedAt ? 'تواصل مع الزبون' : 'انفتحت الشكوى'}
                  </b>
                </p>
                <div className="flex gap-2">
                  {canSeeFullDetail && selected.status !== 'RESOLVED' && (
                    <button onClick={() => onResolveClick(selected.id)}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white">
                      علّمها محلولة
                    </button>
                  )}
                  {canSeeFullDetail && (
                    <select value={selected.status}
                      onChange={(e) => onStatus(selected.id, e.target.value as Complaint['status'])}
                      className="rounded-lg border border-slate-300 px-2 py-1 text-xs">
                      {(Object.keys(statusLabels) as Complaint['status'][]).map((k) => (
                        <option key={k} value={k}>{statusLabels[k]}</option>
                      ))}
                    </select>
                  )}
                  <button onClick={loadLog}
                    className="rounded-lg border px-3 py-1.5 text-xs font-bold"
                    style={{ borderColor: 'var(--bd-line)', color: 'var(--t-body)' }}>
                    📄 عرض السجل
                  </button>
                </div>
              </div>

              {events && (
                <div className="rounded-xl border p-3" style={{ borderColor: 'var(--bd-line)' }}>
                  {events.length === 0 ? (
                    // ⚠️ سجل فارغ **مو** معناه ماكو أحد اشتغل: السجل
                    // انبنى مؤخراً والشكاوى القديمة ما إلها أحداث.
                    // بلا هذا السطر يستنتج المدقق إهمالاً ما صار.
                    <p className="text-xs" style={{ color: 'var(--t-muted)' }}>
                      ماكو أحداث مسجّلة لهذي الشكوى — السجل يبدي من تاريخ تفعيله،
                      والشكاوى الأقدم منه ما إلها أحداث محفوظة.
                    </p>
                  ) : (
                    <ol className="space-y-1.5">
                      {events.map((e) => (
                        <li key={e.id} className="flex flex-wrap items-baseline gap-2 text-xs">
                          <span style={{ color: 'var(--t-faint)' }}>
                            {new Date(e.createdAt).toLocaleString('ar-IQ')}
                          </span>
                          <b style={{ color: 'var(--t-title)' }}>{eventKindLabels[e.kind] ?? e.kind}</b>
                          {e.byName && <span style={{ color: 'var(--t-muted)' }}>— {e.byName}</span>}
                          {e.detail && <span style={{ color: 'var(--t-body)' }}>· {e.detail}</span>}
                        </li>
                      ))}
                    </ol>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function Box({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border p-3"
      style={{ backgroundColor: 'var(--sf-sunken)', borderColor: 'var(--bd-line)' }}>
      <p className="mb-1 text-[11px] font-bold" style={{ color: 'var(--t-muted)' }}>{label}</p>
      {children}
    </div>
  )
}
