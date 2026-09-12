import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, EXTRA_TASK_STATUS, type ExtraTask } from '../api'

// ═══ مهامي الإضافية — لكل موظف ═══
//
// المهام الإضافية كانت تنعرض جوّا شاشة «مهامي»، وهاي محصورة بالفنيين
// والتقنيين. يعني المدير يوجّه مهمة لإداري أو محاسب أو موظف مبيعات
// — **وما توصله** إلا كإشعار يضيع بين باقي الإشعارات.
//
// المهمة الي ما إلها مكان ثابت تنعرض بيه، تنتنسى. والمدير يظن إنه
// وجّه شغل، والموظف ما يدري إن عليه شغل.
//
// فصارت شاشة مستقلة يشوفها **كل** موظف — بلا صلاحية، لأن هاي مهامه
// هو مو مهام غيره. أما **توجيه** المهام لغيره فيحتاج صلاحية
// `extra_tasks_assign`.
//
// ═══ التصميم ═══
// أربع بطاقات أرقام فوگ، شريط تصفية، وبعدها عمودين: القائمة يمين
// وتنبيهات المهام يسار. وبالموبايل ينقلبون عمود واحد والتنبيهات
// تنزل تحت القائمة (مو فوگها — الموظف جاي يشوف مهامه أول).

const STATUS_STYLE: Record<ExtraTask['status'], string> = {
  NEW: 'bg-amber-50 text-amber-700 border-amber-200',
  IN_PROGRESS: 'bg-sky-50 text-sky-700 border-sky-200',
  DONE: 'bg-emerald-50 text-emerald-700 border-emerald-200',
  CANCELLED: 'bg-slate-100 text-slate-500 border-slate-200',
}

// عناوين قريبة من لغة الموظف — «جديدة» ما تگله شنو مطلوب منه،
// «بانتظار البدء» تگله.
const STATUS_LABEL: Record<ExtraTask['status'], string> = {
  ...EXTRA_TASK_STATUS,
  NEW: 'بانتظار البدء',
  DONE: 'مكتمل',
}

const isSameDay = (iso: string | undefined, now: Date) => {
  if (!iso) return false
  const d = new Date(iso)
  return d.getFullYear() === now.getFullYear()
    && d.getMonth() === now.getMonth()
    && d.getDate() === now.getDate()
}

/** «اليوم ١١:٠٠ ص» أوضح للموظف من تاريخ كامل يقراه ويحوّله بمخّه. */
function whenLabel(iso: string | undefined, now: Date): string {
  if (!iso) return 'بلا موعد'
  const d = new Date(iso)
  const time = d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
  const dayDiff = Math.round(
    (new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
      - new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()) / 86400000,
  )
  if (dayDiff === 0) return `اليوم - ${time}`
  if (dayDiff === 1) return `غداً - ${time}`
  if (dayDiff === -1) return `أمس - ${time}`
  return `${d.toLocaleDateString('ar-IQ', { day: 'numeric', month: 'long' })} - ${time}`
}

export default function MyExtraTasksPage() {
  const [tasks, setTasks] = useState<ExtraTask[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<string | null>(null)
  const [doneFor, setDoneFor] = useState<ExtraTask | null>(null)
  const [doneNote, setDoneNote] = useState('')
  const [err, setErr] = useState<string | null>(null)

  // ⚠️ «الآن» ينمسك مرة وحدة بالحالة مو ينقرا بكل رندر: قراءته أثناء
  // الرندر تخلي نفس المكوّن يطلع نتيجتين مختلفتين بنفس اللحظة.
  const [now] = useState(() => new Date())

  // ── التصفية ──
  const [q, setQ] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [dateFilter, setDateFilter] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(5)

  // ⚠️ نجيب **الكل** مرة وحدة ونفلتر محلياً، مو نطلب مع كل تغيير فلتر.
  // عدد مهام الموظف صغير (السيرفر يحدّها بـ٣٠٠)، وطلب جديد على كل حرف
  // يكتبه بالبحث يعني عشرات النداءات على انترنت موبايل.
  // ⚠️ ما نرفع `loading` هنا: الحالة تبدي `true` أصلاً، وإعادة رفعها
  // بجسم الـeffect تولّد رندر متسلسل. وبإعادة التحميل بعد «خلّصتها»
  // ما نريد شاشة تحميل تمسح القائمة — نريد الأرقام تتحدّث بهدوء.
  const load = useCallback(() => {
    api.getMyExtraTasks(true)
      .then(setTasks)
      .catch(() => setTasks([]))
      .finally(() => setLoading(false))
  }, [])

  useEffect(load, [load])

  // نأشّر «شافها» أول ما تنعرض — يقطع نقاش «ما وصلني».
  // ⚠️ للجديدة بس: إعادة التأشير على كل تحميل تدوس على الوقت الأول.
  useEffect(() => {
    tasks.filter((t) => t.status === 'NEW' && !t.seenAt)
      .forEach((t) => { void api.markExtraTaskSeen(t.id).catch(() => undefined) })
  }, [tasks])

  const start = async (t: ExtraTask) => {
    setBusy(t.id)
    try { await api.startExtraTask(t.id); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر البدء') }
    finally { setBusy(null) }
  }

  const complete = async () => {
    if (!doneFor) return
    setErr(null)
    // ⚠️ العدّ بالحروف مو بالبايتات — العربي حرفه بايتين وأكثر.
    if ([...doneNote.trim()].length < 5) {
      setErr('اكتب شنو سويت بالضبط — «تم» ما تكفي')
      return
    }
    setBusy(doneFor.id)
    try {
      await api.completeExtraTask(doneFor.id, doneNote.trim())
      setDoneFor(null); setDoneNote('')
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الإنهاء')
    } finally { setBusy(null) }
  }

  // ── الأرقام الأربعة ──
  const stats = useMemo(() => {
    const weekAgo = new Date(now.getTime() - 7 * 86400000)
    return {
      today: tasks.filter((t) => t.status !== 'CANCELLED' && t.status !== 'DONE' && isSameDay(t.dueAt, now)).length,
      inProgress: tasks.filter((t) => t.status === 'IN_PROGRESS').length,
      waiting: tasks.filter((t) => t.status === 'NEW').length,
      doneWeek: tasks.filter((t) => t.status === 'DONE' && t.doneAt && new Date(t.doneAt) >= weekAgo).length,
    }
  }, [tasks, now])

  // ── التنبيهات ──
  // المستعجلة والي فات موعدها والي موعدها اليوم — مرتّبة بالموعد.
  // هاي مو قائمة ثانية، هاي **نفس** المهام مأشّرة: الموظف عنده ٨ مهام
  // بس اثنتين منهن تحرگ، ولازم يعرفهن بلمحة.
  const alerts = useMemo(
    () => tasks
      .filter((t) => (t.status === 'NEW' || t.status === 'IN_PROGRESS')
        && (t.overdue || t.priority === 'URGENT' || isSameDay(t.dueAt, now)))
      .sort((a, b) => (a.dueAt || '9').localeCompare(b.dueAt || '9'))
      .slice(0, 5),
    [tasks, now],
  )

  // ── التصفية ──
  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase()
    return tasks.filter((t) => {
      if (statusFilter && t.status !== statusFilter) return false
      if (priorityFilter && t.priority !== priorityFilter) return false
      if (dateFilter) {
        if (!t.dueAt) return false
        if (new Date(t.dueAt).toISOString().slice(0, 10) !== dateFilter) return false
      }
      if (needle) {
        const hay = `${t.title} ${t.description || ''} ${t.assignedByName || ''}`.toLowerCase()
        if (!hay.includes(needle)) return false
      }
      return true
    })
  }, [tasks, q, statusFilter, priorityFilter, dateFilter])

  const filtersOn = !!(q || statusFilter || priorityFilter || dateFilter)
  const clearFilters = () => { setQ(''); setStatusFilter(''); setPriorityFilter(''); setDateFilter(''); setPage(1) }

  // ⚠️ الصفحة تنرجع للأول لما تتغير التصفية — وإلا تبقى بصفحة ٣ وتشوف
  // «ماكو نتائج» مع إن النتائج موجودة بصفحة ١.
  const pageCount = Math.max(1, Math.ceil(filtered.length / perPage))
  const safePage = Math.min(page, pageCount)
  const shown = filtered.slice((safePage - 1) * perPage, safePage * perPage)

  return (
    <div dir="rtl" className="mx-auto max-w-6xl space-y-4 p-1">
      <div className="flex items-center gap-3">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-50 text-lg sm:h-11 sm:w-11 sm:text-xl">📋</span>
        <div className="min-w-0">
          <h1 className="text-xl font-black text-[#0f2040] sm:text-2xl">مهامي الإضافية</h1>
          <p className="text-[11px] text-slate-500 sm:text-xs">المهام الموجّهة إليك من المدير خارج الحجوزات العادية</p>
        </div>
      </div>

      {/* ═══ الأرقام الأربعة ═══ */}
      <div className="grid grid-cols-2 gap-2.5 sm:gap-3 lg:grid-cols-4">
        <StatCard label="مهام اليوم" hint="مهام مستحقة اليوم" value={stats.today} icon="📅" tone="sky" />
        <StatCard label="قيد التنفيذ" hint="مهام" value={stats.inProgress} icon="🔄" tone="violet" />
        <StatCard label="بانتظار البدء" hint="مهام" value={stats.waiting} icon="⏱" tone="amber" />
        <StatCard label="مكتملة هذا الأسبوع" hint="مهام منجزة" value={stats.doneWeek} icon="✅" tone="emerald" />
      </div>

      {/* ═══ التصفية ═══
          تشتغل **مباشرة** بلا زر «طبّق»: الموظف يغيّر الحالة ويشوف
          النتيجة، مو يغيّر ويستنى ويتساءل ليش ما تغيّر شي. */}
      <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
        <div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2 lg:grid-cols-5">
          <div className="lg:col-span-2">
            <label className="mb-1 block text-[10px] font-bold text-slate-500">بحث</label>
            <input
              value={q}
              onChange={(e) => { setQ(e.target.value); setPage(1) }}
              placeholder="🔍 ابحث عن مهمة أو مدير..."
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">التاريخ</label>
            <input
              type="date"
              value={dateFilter}
              onChange={(e) => { setDateFilter(e.target.value); setPage(1) }}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">الأولوية</label>
            <select
              value={priorityFilter}
              onChange={(e) => { setPriorityFilter(e.target.value); setPage(1) }}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            >
              <option value="">كل الأولويات</option>
              <option value="URGENT">🔴 مستعجلة</option>
              <option value="NORMAL">عادية</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">الحالة</label>
            <select
              value={statusFilter}
              onChange={(e) => { setStatusFilter(e.target.value); setPage(1) }}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            >
              <option value="">كل الحالات</option>
              <option value="NEW">بانتظار البدء</option>
              <option value="IN_PROGRESS">قيد التنفيذ</option>
              <option value="DONE">مكتمل</option>
              <option value="CANCELLED">ملغاة</option>
            </select>
          </div>
        </div>
        {filtersOn && (
          <div className="mt-2.5 flex flex-wrap items-center justify-between gap-2">
            <p className="text-[11px] text-slate-500">التصفية شغّالة — {filtered.length} من {tasks.length} مهمة</p>
            <button
              onClick={clearFilters}
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-[11px] font-bold text-slate-600 hover:bg-slate-50"
            >
              ✖ مسح التصفية
            </button>
          </div>
        )}
      </div>

      {/* ═══ عمودين: القائمة يمين والتنبيهات يسار ═══ */}
      <div className="grid gap-4 lg:grid-cols-[1fr_300px] lg:items-start">
        <div className="min-w-0 rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_2px_12px_rgba(15,32,64,0.05)] sm:p-4">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 pb-2.5">
            <h3 className="font-bold text-[#0f2040]">قائمة المهام الإضافية</h3>
            {filtered.length > 0 && (
              <p className="text-[11px] text-slate-400">
                عرض {(safePage - 1) * perPage + 1}-{Math.min(safePage * perPage, filtered.length)} من {filtered.length} مهام
              </p>
            )}
          </div>

          {loading && <p className="py-8 text-center text-sm text-slate-400">جاري التحميل...</p>}

          {!loading && filtered.length === 0 && (
            <p className="rounded-xl border border-dashed border-slate-200 bg-slate-50/60 px-3 py-10 text-center text-sm text-slate-400">
              {tasks.length === 0
                ? 'ما انوجّهت إلك ولا مهمة إضافية بعد.'
                : 'ماكو مهام بهذي التصفية — جرّب تمسحها.'}
            </p>
          )}

          <div className="divide-y divide-slate-100">
            {shown.map((t) => (
              <TaskRow
                key={t.id}
                task={t}
                now={now}
                busy={busy === t.id}
                onStart={() => start(t)}
                onComplete={() => { setDoneFor(t); setDoneNote(''); setErr(null) }}
              />
            ))}
          </div>

          {/* ── الترقيم ── */}
          {filtered.length > perPage && (
            <div className="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-slate-100 pt-3">
              <label className="flex items-center gap-1.5 text-[11px] text-slate-500">
                عرض
                <select
                  value={perPage}
                  onChange={(e) => { setPerPage(Number(e.target.value)); setPage(1) }}
                  className="rounded-lg border border-slate-300 px-2 py-1 text-[11px]"
                >
                  {[5, 10, 25].map((n) => <option key={n} value={n}>{n}</option>)}
                </select>
                في الصفحة
              </label>
              <div className="flex items-center gap-1.5">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={safePage === 1}
                  className="rounded-lg border border-slate-300 px-2.5 py-1 text-[11px] font-bold text-slate-600 disabled:opacity-40"
                >
                  السابق
                </button>
                {Array.from({ length: pageCount }, (_, i) => i + 1).map((n) => (
                  <button
                    key={n}
                    onClick={() => setPage(n)}
                    className={`h-7 w-7 rounded-lg text-[11px] font-bold ${
                      n === safePage ? 'bg-[#2c5aad] text-white' : 'border border-slate-300 text-slate-600'
                    }`}
                  >
                    {n}
                  </button>
                ))}
                <button
                  onClick={() => setPage((p) => Math.min(pageCount, p + 1))}
                  disabled={safePage === pageCount}
                  className="rounded-lg border border-slate-300 px-2.5 py-1 text-[11px] font-bold text-slate-600 disabled:opacity-40"
                >
                  التالي
                </button>
              </div>
            </div>
          )}
        </div>

        {/* ── العمود الجانبي ──
            بالموبايل ينزل **تحت** القائمة عن قصد: الموظف فاتح الشاشة
            حتى يشوف مهامه، مو حتى يقرا تنبيهات قبلهن. */}
        <aside className="space-y-3">
          <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
            <h3 className="mb-2.5 flex items-center gap-1.5 text-sm font-bold text-[#0f2040]">
              🔔 تنبيهات المهام
            </h3>
            {alerts.length === 0 ? (
              <p className="text-[11px] text-slate-400">ماكو شي مستعجل — كلشي بموعده.</p>
            ) : (
              <div className="space-y-2.5">
                {alerts.map((t) => (
                  <div key={t.id} className="flex items-start gap-2">
                    <span className={`mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full ${
                      t.overdue ? 'bg-red-500' : t.priority === 'URGENT' ? 'bg-amber-500' : 'bg-sky-500'
                    }`} />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-start justify-between gap-1.5">
                        <p className="min-w-0 flex-1 truncate text-[11px] font-bold text-slate-700">{t.title}</p>
                        <span className={`shrink-0 rounded-md px-1.5 py-0.5 text-[9px] font-bold ${
                          t.overdue ? 'bg-red-100 text-red-700'
                            : t.priority === 'URGENT' ? 'bg-amber-100 text-amber-800'
                            : 'bg-sky-100 text-sky-700'
                        }`}>
                          {t.overdue ? 'فات موعدها' : t.priority === 'URGENT' ? 'مستعجل' : 'اليوم'}
                        </span>
                      </div>
                      <p className="mt-0.5 text-[10px] text-slate-400">{whenLabel(t.dueAt, now)}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
            <h3 className="mb-2 flex items-center gap-1.5 text-sm font-bold text-[#0f2040]">
              ✏️ ملاحظات سريعة
            </h3>
            <p className="text-[11px] leading-relaxed text-slate-500">
              احرص على تحديث حالة المهام باستمرار وإضافة الملاحظات عند الحاجة لتسهيل المتابعة.
              الي تكتبه بخانة «شنو سويت» يوصل المدير مباشرة — فما يحتاج يتصل بيك يسأل.
            </p>
          </div>
        </aside>
      </div>

      {/* ⚠️ الإنجاز يطلب وصف: «تم» بلا شرح ترجّع المدير يسأل بالتلفون
          — وهاي نفس المكالمة الي بنينا الميزة حتى نلغيها. */}
      {doneFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setDoneFor(null)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="font-bold text-[#0f2040]">✓ إنهاء: {doneFor.title}</h3>
            <label className="mb-1 mt-3 block text-xs font-bold text-slate-500">
              شنو سويت بالضبط؟ <span className="text-red-600">*</span>
            </label>
            <textarea
              value={doneNote}
              onChange={(e) => setDoneNote(e.target.value)}
              rows={3}
              placeholder="مثال: خرّجت ٤٢ فاتورة للشهر ورفعتها بالنظام المحاسبي"
              className="w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            {err && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{err}</p>}
            <div className="mt-3 flex gap-2">
              <button
                onClick={complete}
                disabled={busy !== null}
                className="flex-1 rounded-xl bg-emerald-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {busy ? 'جاري الحفظ...' : 'خلّصتها'}
              </button>
              <button onClick={() => setDoneFor(null)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/* ───── بطاقة رقم ───── */

const TONES: Record<string, string> = {
  sky: 'bg-sky-50 text-sky-600',
  violet: 'bg-violet-50 text-violet-600',
  amber: 'bg-amber-50 text-amber-600',
  emerald: 'bg-emerald-50 text-emerald-600',
}

function StatCard({ label, hint, value, icon, tone }: {
  label: string; hint: string; value: number; icon: string; tone: string
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_2px_12px_rgba(15,32,64,0.05)] sm:p-4">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="truncate text-[10px] font-bold text-slate-500 sm:text-[11px]">{label}</p>
          <p className="mt-1 text-2xl font-black text-[#0f2040] sm:text-3xl">{value}</p>
          <p className="text-[10px] text-slate-400">{hint}</p>
        </div>
        <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-xl text-sm sm:h-10 sm:w-10 sm:text-base ${TONES[tone]}`}>
          {icon}
        </span>
      </div>
    </div>
  )
}

/* ───── صف مهمة ─────
   ⚠️ مكوّن **برّا** المكوّن الأب مو جوّاه: التعريف جوّا الأب يخلي React
   يحسبه نوع جديد بكل رندر، فيهدم الصف ويعيد بناءه — والنتيجة إن أي
   خانة كتابة جوّاه تفقد التركيز بعد أول حرف. */
function TaskRow({ task: t, now, busy, onStart, onComplete }: {
  task: ExtraTask; now: Date; busy: boolean; onStart: () => void; onComplete: () => void
}) {
  const [open, setOpen] = useState(false)
  const actionable = t.status === 'NEW' || t.status === 'IN_PROGRESS'

  return (
    <div className={`py-3 ${t.overdue ? 'border-r-2 border-red-400 pr-2.5' : ''}`}>
      <div className="flex items-start gap-2.5">
        <span className="mt-0.5 shrink-0 text-base">
          {t.overdue ? '🔺' : t.status === 'DONE' ? '✅' : t.priority === 'URGENT' ? '🔴' : '📄'}
        </span>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-1.5">
            <b className="text-sm text-[#0f2040]">{t.title}</b>
            <span className={`rounded-lg border px-2 py-0.5 text-[10px] font-bold ${STATUS_STYLE[t.status]}`}>
              {STATUS_LABEL[t.status]}
            </span>
            {t.priority === 'URGENT' && t.status !== 'DONE' && (
              <span className="rounded-lg border border-red-200 bg-red-50 px-2 py-0.5 text-[10px] font-bold text-red-700">مستعجل</span>
            )}
            {t.overdue && (
              <span className="rounded-lg bg-red-600 px-2 py-0.5 text-[10px] font-bold text-white">فات موعدها</span>
            )}
          </div>

          {t.description && <p className="mt-1 text-[11px] leading-relaxed text-slate-500">{t.description}</p>}

          <div className="mt-1.5 flex flex-wrap gap-x-3 gap-y-1 text-[10px] text-slate-400">
            {t.assignedByName && <span>👤 المدير: {t.assignedByName}</span>}
            <span>🕐 {whenLabel(t.dueAt, now)}</span>
            {/* «شافها» يقطع نقاش «ما وصلني» */}
            {!t.seenAt && t.status === 'NEW' && <span>⚪ ما فتحتها بعد</span>}
          </div>

          {t.status === 'DONE' && t.doneNote && (
            <p className="mt-2 rounded-lg bg-emerald-50 px-3 py-2 text-[11px] text-emerald-900">✅ {t.doneNote}</p>
          )}
          {t.status === 'CANCELLED' && (
            <p className="mt-2 rounded-lg bg-slate-100 px-3 py-2 text-[11px] text-slate-600">✖️ انلغت — {t.cancelReason}</p>
          )}

          {open && actionable && (
            <div className="mt-2.5 flex flex-wrap gap-2">
              {t.status === 'NEW' && (
                <button
                  onClick={onStart}
                  disabled={busy}
                  className="rounded-lg border border-brand-300 px-3 py-1.5 text-[11px] font-bold text-brand-700 disabled:opacity-50"
                >
                  ▶ بديت بيها
                </button>
              )}
              <button
                onClick={onComplete}
                className="rounded-lg bg-emerald-700 px-3 py-1.5 text-[11px] font-bold text-white"
              >
                ✓ خلّصتها
              </button>
            </div>
          )}
        </div>

        {actionable && (
          <button
            onClick={() => setOpen((v) => !v)}
            aria-label={open ? 'إخفاء الإجراءات' : 'عرض الإجراءات'}
            className="shrink-0 rounded-lg px-2 py-1 text-slate-400 hover:bg-slate-100"
          >
            {open ? '▲' : '▼'}
          </button>
        )}
      </div>
    </div>
  )
}
