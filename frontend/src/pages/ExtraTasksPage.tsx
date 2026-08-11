import { useCallback, useEffect, useState } from 'react'
import { api, EXTRA_TASK_STATUS, type Employee, type ExtraTask } from '../api'

// ═══ المهام الإضافية — شاشة المدير ═══
//
// «يجي المدير يوجّه أحد الموظفين شيسوي؟ يختار موظف ويوجّهله عمل يطلعله
// بالنظام… مثلاً أوجّهه على تخريج فواتير».
//
// قبلها: الشغل الي مو حجز ينقال بالتلفون. المدير ما يتذكر منو وجّه
// بشنو، والموظف يگول «ما وصلني»، وماكو أثر يفصل بينهم.
//
// ⚠️ هنا التوجيه والمتابعة بس. التنفيذ (بديت / خلّصت) بشاشة الموظف —
// والسيرفر ما يخلي أحد ينهي مهمة مو إله.

const STATUS_STYLE: Record<string, string> = {
  NEW: 'bg-sky-100 text-sky-800',
  IN_PROGRESS: 'bg-amber-100 text-amber-800',
  DONE: 'bg-emerald-100 text-emerald-800',
  CANCELLED: 'bg-slate-200 text-slate-600',
}

export default function ExtraTasksPage() {
  const [tasks, setTasks] = useState<ExtraTask[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)

  // نموذج التوجيه
  const [title, setTitle] = useState('')
  const [desc, setDesc] = useState('')
  const [assignee, setAssignee] = useState('')
  const [priority, setPriority] = useState<'NORMAL' | 'URGENT'>('NORMAL')
  const [dueAt, setDueAt] = useState('')
  const [busy, setBusy] = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    api.getExtraTasks({ status: statusFilter })
      .then(setTasks)
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب'))
      .finally(() => setLoading(false))
  }, [statusFilter])

  useEffect(load, [load])
  useEffect(() => { api.getEmployees().then(setEmployees).catch(() => setEmployees([])) }, [])

  const assign = async () => {
    setErr(null)
    if (!title.trim()) { setErr('اكتب عنوان المهمة'); return }
    if (!assignee) { setErr('اختار الموظف'); return }
    setBusy(true)
    try {
      await api.createExtraTask({
        title: title.trim(),
        description: desc.trim() || undefined,
        assignedToId: assignee,
        priority,
        // datetime-local ينطي وقتاً بلا منطقة — نحوّله ISO كامل
        dueAt: dueAt ? new Date(dueAt).toISOString() : null,
      })
      setTitle(''); setDesc(''); setDueAt(''); setPriority('NORMAL')
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر التوجيه')
    } finally { setBusy(false) }
  }

  const cancel = async (t: ExtraTask) => {
    const reason = prompt(`ليش تلغي «${t.title}»؟`)
    if (!reason) return
    try {
      await api.cancelExtraTask(t.id, reason)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الإلغاء')
    }
  }

  const open = tasks.filter((t) => t.status === 'NEW' || t.status === 'IN_PROGRESS')

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📋 المهام الإضافية</h2>
        <p className="mt-1 text-sm text-slate-500">
          وجّه شغل لموظف — شغل مو مربوط بحجز. يوصله بإشعار ويطلع بشاشته.
        </p>
      </div>

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}

      {/* ═══ التوجيه ═══ */}
      <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <h3 className="mb-3 font-bold text-[#0f2040]">وجّه مهمة جديدة</h3>
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-xs font-bold text-slate-500">المهمة *</label>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="مثال: تخريج فواتير الشهر"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-bold text-slate-500">الموظف *</label>
            <select
              value={assignee}
              onChange={(e) => setAssignee(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            >
              <option value="">— اختار الموظف —</option>
              {employees.map((e) => (
                <option key={e.id} value={e.id}>{e.name}{e.jobTitle ? ` — ${e.jobTitle}` : ''}</option>
              ))}
            </select>
          </div>
          <div className="sm:col-span-2">
            <label className="mb-1 block text-xs font-bold text-slate-500">تفاصيل (اختياري)</label>
            <textarea
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
              rows={2}
              placeholder="شنو المطلوب بالضبط؟"
              className="w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-bold text-slate-500">الأهمية</label>
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as 'NORMAL' | 'URGENT')}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            >
              <option value="NORMAL">عادية</option>
              <option value="URGENT">🔴 مستعجلة</option>
            </select>
          </div>
          <div>
            {/* ⚠️ «بلا موعد» حالة مشروعة — ما نفرض تاريخاً وهمياً يخلي
                المهمة تطلع «متأخرة» وهي أصلاً ما إلها موعد. */}
            <label className="mb-1 block text-xs font-bold text-slate-500">آخر موعد (اختياري)</label>
            <input
              type="datetime-local"
              value={dueAt}
              onChange={(e) => setDueAt(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
          </div>
        </div>
        <button
          onClick={assign}
          disabled={busy}
          className="mt-3 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2.5 text-sm font-bold text-white disabled:opacity-50"
        >
          {busy ? 'جاري التوجيه...' : '📤 وجّه المهمة'}
        </button>
      </div>

      {/* ═══ المتابعة ═══ */}
      <div className="flex flex-wrap items-center gap-2">
        {[['', 'الكل'], ['NEW', 'جديدة'], ['IN_PROGRESS', 'قيد التنفيذ'], ['DONE', 'منجزة'], ['CANCELLED', 'ملغاة']].map(([k, label]) => (
          <button
            key={k}
            onClick={() => setStatusFilter(k)}
            className={`rounded-lg px-3 py-1.5 text-xs font-bold ${statusFilter === k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600'}`}
          >
            {label}
          </button>
        ))}
        <span className="mr-auto text-xs text-slate-500">{open.length} مهمة مفتوحة</span>
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && tasks.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">ماكو مهام بهذا الفلتر.</p>
      )}

      <div className="space-y-3">
        {tasks.map((t) => (
          <div key={t.id} className={`rounded-2xl border bg-white p-4 shadow-[0_2px_12px_rgba(0,0,0,0.06)] ${t.overdue ? 'border-red-300' : 'border-white'}`}>
            <div className="flex flex-wrap items-center gap-2">
              <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${STATUS_STYLE[t.status]}`}>
                {EXTRA_TASK_STATUS[t.status]}
              </span>
              {t.priority === 'URGENT' && <span className="rounded-full bg-red-100 px-2.5 py-1 text-xs font-bold text-red-700">🔴 مستعجلة</span>}
              {t.overdue && <span className="rounded-full bg-red-600 px-2.5 py-1 text-xs font-bold text-white">⏰ فات موعدها</span>}
              <b className="text-[#0f2040]">{t.title}</b>
              {(t.status === 'NEW' || t.status === 'IN_PROGRESS') && (
                <button onClick={() => cancel(t)} className="mr-auto rounded-lg px-2.5 py-1 text-xs font-bold text-red-600 hover:bg-red-50">
                  إلغاء
                </button>
              )}
            </div>

            {t.description && <p className="mt-1.5 text-sm text-slate-600">{t.description}</p>}

            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
              <span>👤 <b className="text-slate-700">{t.assignedToName || '—'}</b></span>
              {t.assignedByName && <span>📤 وجّهها: {t.assignedByName}</span>}
              {t.dueAt && <span>⏳ آخر موعد: {new Date(t.dueAt).toLocaleString('en-GB')}</span>}
              {/* «شافها» يقطع نقاش «ما وصلني» */}
              <span>{t.seenAt ? `👁️ شافها ${new Date(t.seenAt).toLocaleString('en-GB')}` : '⚪ ما فتحها بعد'}</span>
            </div>

            {t.doneNote && (
              <p className="mt-2 rounded-lg bg-emerald-50 px-3 py-2 text-xs text-emerald-900">
                ✅ {t.doneNote}
                {t.doneAt && <span className="mr-2 text-emerald-700">· {new Date(t.doneAt).toLocaleString('en-GB')}</span>}
              </p>
            )}
            {t.cancelReason && (
              <p className="mt-2 rounded-lg bg-slate-100 px-3 py-2 text-xs text-slate-700">✖️ انلغت: {t.cancelReason}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
