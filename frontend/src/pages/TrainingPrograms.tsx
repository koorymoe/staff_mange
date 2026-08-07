import { useEffect, useMemo, useState } from 'react'
import {
  api, DEPARTMENTS,
  type Employee, type SaveTrainingProgramInput, type SkillWithService, type TrainingProgram,
} from '../api'
import { useSession } from '../session'

// ═══ برامج التدريب ═══
//
// منقولة من نظام الطاقة الشمسية، بس على موظفي الشركة كلهم ومهاراتها
// كلها — مو على قائمة موظفين منفصلة.
//
// الفرق الجوهري عن النظام القديم: إكمال البرنامج **يمنح المهارات
// فعلاً** لكل المشاركين. بالقديم «إصدار الشهادات» جان يغيّر حالة
// البرنامج بس، فالموظف يتدرّب والنظام يضل ما يعرف إنه صار يعرف.

const LEVELS = ['مبتدئ', 'متوسط', 'متقدم'] as const
const STATUSES = ['قيد التخطيط', 'جاري التنفيذ', 'مكتمل'] as const
const STEPS = ['التسجيل', 'النظرية', 'عملي', 'تقييم', 'شهادة']

const iqd = (n: number) => `${Math.round(n || 0).toLocaleString('en-US')} د.ع`

export default function TrainingPrograms() {
  const { employee, permissions } = useSession()
  const canManage =
    employee?.role === 'ADMIN' || employee?.role === 'OWNER' || employee?.role === 'HR_COORDINATOR' ||
    permissions.includes('staff_management') || permissions.includes('content_technician')

  const [programs, setPrograms] = useState<TrainingProgram[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [skills, setSkills] = useState<SkillWithService[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<TrainingProgram | 'new' | null>(null)
  const [statusFilter, setStatusFilter] = useState('الكل')
  const [skillFilter, setSkillFilter] = useState('الكل')

  const load = () =>
    Promise.all([
      api.getTrainingPrograms().then(setPrograms).catch(() => {}),
      api.getEmployees().then((e) => setEmployees(e.filter((x) => x.status === 'ACTIVE'))).catch(() => {}),
      api.getSkills().then(setSkills).catch(() => {}),
    ])

  useEffect(() => { load().finally(() => setLoading(false)) }, [])

  const shown = programs
    .filter((p) => statusFilter === 'الكل' || p.status === statusFilter)
    .filter((p) => skillFilter === 'الكل' || p.skills.some((s) => s.name === skillFilter))

  const counts = {
    total: programs.length,
    running: programs.filter((p) => p.status === 'جاري التنفيذ').length,
    done: programs.filter((p) => p.status === 'مكتمل').length,
    trainees: new Set(programs.flatMap((p) => p.participants.map((x) => x.employeeId))).size,
  }

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-lime-600 to-emerald-600 p-5 text-white shadow-lg">
        <h1 className="text-xl font-extrabold">🎓 التدريب والتطوير المهني</h1>
        <p className="mt-1 text-xs opacity-90">
          برامج تدريبية بمهارات مستهدفة — وإكمال البرنامج يمنح مهاراته للمتدربين تلقائياً
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Box label="إجمالي البرامج" value={counts.total} tone="lime" />
        <Box label="جارية التنفيذ" value={counts.running} tone="blue" />
        <Box label="مكتملة" value={counts.done} tone="emerald" />
        <Box label="متدربين" value={counts.trainees} tone="amber" />
      </div>

      <div className="rounded-2xl bg-white p-4 shadow-sm">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">الحالة</label>
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
              <option>الكل</option>
              {STATUSES.map((s) => <option key={s}>{s}</option>)}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-slate-600">المهارة المستهدفة</label>
            <select value={skillFilter} onChange={(e) => setSkillFilter(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
              <option>الكل</option>
              {skills.map((s) => <option key={s.id}>{s.name}</option>)}
            </select>
          </div>
          {canManage && (
            <div className="flex items-end">
              <button onClick={() => setEditing('new')} className="w-full rounded-xl bg-gradient-to-l from-lime-600 to-emerald-600 px-4 py-2.5 text-sm font-bold text-white">
                + برنامج تدريبي جديد
              </button>
            </div>
          )}
        </div>
      </div>

      {editing && (
        <ProgramForm
          program={editing === 'new' ? null : editing}
          employees={employees}
          skills={skills}
          onClose={() => setEditing(null)}
          onSaved={async () => { await load(); setEditing(null) }}
        />
      )}

      {loading && <div className="rounded-2xl bg-white p-8 text-center text-sm text-slate-500">جاري التحميل...</div>}

      {!loading && shown.length === 0 && (
        <div className="rounded-2xl bg-white p-10 text-center">
          <div className="text-4xl">🎓</div>
          <p className="mt-3 text-sm font-bold text-slate-600">ماكو برامج تدريبية</p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {shown.map((p) => (
          <div key={p.id} className="overflow-hidden rounded-2xl bg-white shadow-sm">
            <div className="flex items-center justify-between bg-gradient-to-l from-lime-600 to-emerald-600 p-4 text-white">
              <StatusBadge status={p.status} />
              <div className="text-right">
                <div className="text-base font-black">{p.name}</div>
                <div className="text-[11px] opacity-90">
                  {p.startDate ? new Date(p.startDate).toLocaleDateString('ar-IQ') : 'بلا موعد'} · {p.durationDays} أيام
                  {p.endDate ? ` · تنتهي ${new Date(p.endDate).toLocaleDateString('ar-IQ')}` : ''}
                </div>
              </div>
            </div>

            <div className="p-4">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-bold text-slate-600">
                  👥 {p.participants.length} متدرب
                </span>
                <span className={`rounded-full px-2.5 py-1 text-[11px] font-black ${
                  p.level === 'متقدم' ? 'bg-red-100 text-red-700' : p.level === 'متوسط' ? 'bg-amber-100 text-amber-700' : 'bg-blue-100 text-blue-700'
                }`}>{p.level}</span>
              </div>

              <div className="mb-3">
                <div className="mb-1 text-[11px] font-bold text-slate-500">المهارات الي راح ينالها المتدرّب:</div>
                <div className="flex flex-wrap gap-1">
                  {p.skills.length === 0
                    ? <span className="text-[11px] text-slate-400">ماكو مهارات مرتبطة</span>
                    : p.skills.map((s) => (
                      <span key={s.skillId} className="rounded-full border border-violet-200 bg-violet-50 px-2 py-0.5 text-[10px] font-bold text-violet-700">
                        {s.name}
                      </span>
                    ))}
                </div>
              </div>

              {/* المراحل الخمس */}
              <div className="mb-3 flex items-center justify-between">
                {STEPS.map((label, i) => {
                  const threshold = i * 25
                  const done = p.progress >= threshold && (i === 0 || p.progress >= threshold)
                  return (
                    <div key={label} className="flex flex-1 flex-col items-center">
                      <div className={`flex h-7 w-7 items-center justify-center rounded-full text-[11px] font-black ${
                        done ? 'bg-emerald-500 text-white' : 'border-2 border-slate-200 bg-white text-slate-400'
                      }`}>{done ? '✓' : i + 1}</div>
                      <div className={`mt-1 text-[9px] font-bold ${done ? 'text-emerald-600' : 'text-slate-400'}`}>{label}</div>
                    </div>
                  )
                })}
              </div>

              <div className="mb-3 h-2 overflow-hidden rounded-full bg-slate-100">
                <div className="h-full rounded-full bg-emerald-500" style={{ width: `${p.progress}%` }} />
              </div>

              <div className="flex flex-wrap justify-between gap-2 text-[11px] text-slate-500">
                <span>💰 {iqd(p.cost)}</span>
                <span>🎯 نجاح {p.passRate}%</span>
                <span>👤 {p.instructorName || 'بلا مدرّب'}</span>
                <span>🏢 {p.targetDepartment || 'كل الأقسام'}</span>
              </div>

              {p.participants.length > 0 && (
                <div className="mt-3 max-h-24 overflow-y-auto rounded-lg bg-slate-50 p-2">
                  {p.participants.map((x) => (
                    <div key={x.employeeId} className="text-[11px] text-slate-600">
                      {x.passed ? '✅' : '•'} {x.name} <span className="text-slate-400">{x.department || ''}</span>
                    </div>
                  ))}
                </div>
              )}

              {canManage && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {p.status !== 'مكتمل' && (
                    <button
                      onClick={async () => {
                        if (!confirm(`تكمل «${p.name}»؟\n\nراح تنمنح مهارات البرنامج (${p.skills.length}) لكل المشاركين (${p.participants.length}) بملفاتهم.`)) return
                        try { await api.completeTrainingProgram(p.id); await load() }
                        catch (e) { alert(e instanceof Error ? e.message : 'تعذر الإكمال') }
                      }}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white"
                    >
                      🎓 أكمله وامنح المهارات
                    </button>
                  )}
                  <button onClick={() => setEditing(p)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-600">✏️ تعديل</button>
                  <button
                    onClick={async () => {
                      if (!confirm(`تريد تحذف «${p.name}»؟`)) return
                      try { await api.deleteTrainingProgram(p.id); await load() }
                      catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحذف') }
                    }}
                    className="rounded-lg border border-red-300 px-3 py-1.5 text-xs font-bold text-red-600"
                  >🗑️</button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const map: Record<string, string> = {
    'مكتمل': 'bg-white/25',
    'جاري التنفيذ': 'bg-amber-400 text-amber-950',
    'قيد التخطيط': 'bg-white/20',
  }
  return <span className={`rounded-full px-3 py-1 text-[11px] font-black ${map[status] || 'bg-white/20'}`}>{status}</span>
}

function Box({ label, value, tone }: { label: string; value: number; tone: string }) {
  const tones: Record<string, string> = {
    lime: 'border-lime-200 bg-lime-50 text-lime-700',
    blue: 'border-blue-200 bg-blue-50 text-blue-700',
    emerald: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    amber: 'border-amber-200 bg-amber-50 text-amber-700',
  }
  return (
    <div className={`rounded-2xl border-2 p-4 text-center ${tones[tone]}`}>
      <div className="text-2xl font-black">{value}</div>
      <div className="mt-1 text-[11px] font-medium opacity-80">{label}</div>
    </div>
  )
}

function ProgramForm({
  program, employees, skills, onClose, onSaved,
}: {
  program: TrainingProgram | null
  employees: Employee[]
  skills: SkillWithService[]
  onClose: () => void
  onSaved: () => Promise<void>
}) {
  const [f, setF] = useState<SaveTrainingProgramInput>({
    name: program?.name ?? '',
    level: program?.level ?? 'مبتدئ',
    durationDays: program?.durationDays ?? 3,
    startDate: program?.startDate ? program.startDate.slice(0, 10) : '',
    targetDepartment: program?.targetDepartment ?? '',
    instructorId: program?.instructorId ?? '',
    objectives: program?.objectives ?? '',
    content: program?.content ?? '',
    passRate: program?.passRate ?? 80,
    cost: program?.cost ?? 0,
    status: program?.status ?? 'قيد التخطيط',
    progress: program?.progress ?? 0,
    participantIds: program?.participants.map((p) => p.employeeId) ?? [],
    skillIds: program?.skills.map((s) => s.skillId) ?? [],
  })
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const set = <K extends keyof SaveTrainingProgramInput>(k: K, v: SaveTrainingProgramInput[K]) =>
    setF((p) => ({ ...p, [k]: v }))

  const toggle = (key: 'participantIds' | 'skillIds', id: string) =>
    setF((p) => ({ ...p, [key]: p[key].includes(id) ? p[key].filter((x) => x !== id) : [...p[key], id] }))

  // تاريخ النهاية معروض بس — السيرفر يحسبه من البداية والمدة حتى ما
  // يصير عدنا رقمين ما يتفقون
  const endPreview = useMemo(() => {
    if (!f.startDate || f.durationDays < 1) return ''
    const d = new Date(f.startDate)
    d.setDate(d.getDate() + f.durationDays - 1)
    return d.toLocaleDateString('ar-IQ')
  }, [f.startDate, f.durationDays])

  const byCategory = useMemo(() => {
    const m: Record<string, SkillWithService[]> = {}
    for (const s of skills) (m[s.category] ||= []).push(s)
    return m
  }, [skills])

  const save = async () => {
    if (!f.name.trim()) { setErr('اسم البرنامج مطلوب'); return }
    setBusy(true); setErr(null)
    try {
      if (program) await api.updateTrainingProgram(program.id, f)
      else await api.createTrainingProgram(f)
      await onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  return (
    <div className="rounded-2xl border-2 border-lime-400 bg-white p-5 shadow-lg">
      <h3 className="mb-4 text-base font-extrabold text-[#0f2040]">
        {program ? '✏️ تعديل البرنامج' : '➕ برنامج تدريبي جديد'}
      </h3>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-4">
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium text-slate-600">اسم البرنامج *</label>
          <input value={f.name} onChange={(e) => set('name', e.target.value)} placeholder="دورة تركيب الألواح الشمسية المتقدمة" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">المستوى</label>
          <select value={f.level} onChange={(e) => set('level', e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            {LEVELS.map((l) => <option key={l}>{l}</option>)}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">المدة (أيام)</label>
          <input type="number" min={1} value={f.durationDays} onChange={(e) => set('durationDays', +e.target.value || 1)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">تاريخ البداية</label>
          <input type="date" value={f.startDate} onChange={(e) => set('startDate', e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">تاريخ النهاية (يُحسب تلقائياً)</label>
          <input value={endPreview} readOnly className="w-full rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-500" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">القسم المستهدف</label>
          <select value={f.targetDepartment} onChange={(e) => set('targetDepartment', e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">كل الأقسام</option>
            {DEPARTMENTS.map((d) => <option key={d}>{d}</option>)}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">المدرّب</label>
          <select value={f.instructorId} onChange={(e) => set('instructorId', e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">— اختر —</option>
            {employees.map((e) => <option key={e.id} value={e.id}>{e.name}{e.jobTitle ? ` — ${e.jobTitle}` : ''}</option>)}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">نسبة النجاح %</label>
          <input type="number" min={0} max={100} value={f.passRate} onChange={(e) => set('passRate', +e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">التكلفة (د.ع)</label>
          <input type="number" min={0} value={f.cost} onChange={(e) => set('cost', +e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">الحالة</label>
          <select value={f.status} onChange={(e) => set('status', e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            {STATUSES.map((s) => <option key={s}>{s}</option>)}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">التقدّم %</label>
          <input type="number" min={0} max={100} value={f.progress} onChange={(e) => set('progress', +e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
      </div>

      <div className="mt-4 rounded-xl border-2 border-violet-200 bg-violet-50/50 p-3">
        <h4 className="mb-2 text-sm font-bold text-violet-800">
          🎯 المهارات المستهدفة — تنمنح للمتدربين لمن يكتمل البرنامج
        </h4>
        <div className="max-h-56 overflow-y-auto">
          {Object.entries(byCategory).map(([cat, list]) => (
            <div key={cat} className="mb-2">
              <div className="mb-1 border-b border-violet-200 pb-1 text-xs font-black text-violet-700">{cat}</div>
              <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
                {list.map((s) => (
                  <label key={s.id} className="flex cursor-pointer items-center gap-2 rounded-lg bg-white px-2 py-1.5 text-xs">
                    <input type="checkbox" checked={f.skillIds.includes(s.id)} onChange={() => toggle('skillIds', s.id)} />
                    <span className="font-bold text-slate-700">{s.name}</span>
                    {s.serviceName && <span className="text-[10px] text-slate-400">{s.serviceName}</span>}
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="mt-4 rounded-xl border-2 border-slate-200 bg-slate-50 p-3">
        <h4 className="mb-2 text-sm font-bold text-slate-700">👥 المشاركين ({f.participantIds.length})</h4>
        <div className="grid max-h-48 grid-cols-1 gap-1 overflow-y-auto sm:grid-cols-3">
          {employees.map((e) => (
            <label key={e.id} className="flex cursor-pointer items-center gap-2 rounded-lg bg-white px-2 py-1.5 text-xs">
              <input type="checkbox" checked={f.participantIds.includes(e.id)} onChange={() => toggle('participantIds', e.id)} />
              <span className="font-bold text-slate-700">{e.name}</span>
              {e.department && <span className="text-[10px] text-slate-400">{e.department}</span>}
            </label>
          ))}
        </div>
      </div>

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">الأهداف التدريبية</label>
          <textarea value={f.objectives} onChange={(e) => set('objectives', e.target.value)} rows={3} placeholder="- تعلم أساسيات تركيب الألواح&#10;- فهم أنظمة الأمان الكهربائية" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">المحتوى / الخطة الزمنية</label>
          <textarea value={f.content} onChange={(e) => set('content', e.target.value)} rows={3} placeholder="اليوم 1: مقدمة نظرية&#10;اليوم 2: تطبيق عملي" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <button onClick={save} disabled={busy} className="rounded-lg bg-gradient-to-l from-lime-600 to-emerald-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">
          {busy ? 'جاري الحفظ...' : 'حفظ البرنامج'}
        </button>
        <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
        {err && <span className="text-xs font-bold text-red-600">{err}</span>}
      </div>
    </div>
  )
}
