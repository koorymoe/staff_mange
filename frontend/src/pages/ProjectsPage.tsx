import { useEffect, useMemo, useState } from 'react'
import { useSession } from '../session'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

// ---------------------------------------------------------------------------
// Types & constants
// ---------------------------------------------------------------------------
interface Project {
  id: string
  code: string
  name: string
  rep: string | null
  phone: string | null
  location: string | null
  workType: string | null
  refPerson: string | null
  stage: string
  price: string | null
  staff: string | null
  time: string | null
  task: string | null
  priority: string
  deliveryDate: string | null
  survey: (string | null)[] | null
  sentToGroup: boolean
  createdAt: string
}

interface Stats {
  اتصال: number; كشف: number; سعر: number; تنفيذ: number; مكتمل: number; مرفوض: number
}

const STAGES = [
  '1. اتصال بالزبون',
  '2. مرحلة الكشف',
  '3. عرض السعر',
  '4. البدء بالتنفيذ',
  '✅ مكتمل',
  '❌ مرفوض',
]

const WORK_TYPES = ['طاقة شمسية', 'كاميرات', 'بيت ذكي', 'شبكات', 'إنذار حريق', 'أقفال وحاكيات', 'ستلايت', 'منظومة صوت', 'أخرى']

const STAGE_CARDS = [
  { key: 'اتصال', label: 'اتصال', icon: '📞', color: 'bg-[var(--color-brand-500)]' },
  { key: 'كشف', label: 'كشف', icon: '🔍', color: 'bg-green-600' },
  { key: 'سعر', label: 'سعر', icon: '💰', color: 'bg-amber-500' },
  { key: 'تنفيذ', label: 'تنفيذ', icon: '🛠️', color: 'bg-red-600' },
  { key: 'مكتمل', label: 'مكتمل', icon: '✅', color: 'bg-blue-600' },
  { key: 'مرفوض', label: 'مرفوض', icon: '❌', color: 'bg-gray-500' },
] as const

// استمارة الكشف الفني (17 سؤال)
const SURVEY = [
  { q: 'تفاصيل الموقع', opts: ['تحت الإنشاء', 'بناء مكتمل', 'ترميم/تطوير'] },
  { q: 'طرق التسديد المالي', opts: ['نقد (كاش)', 'أقساط شهري', 'دفعات إنجاز'] },
  { q: 'نمط وجدية الزبون', opts: ['جدي جداً', 'متردد/مقارنة', 'استفسار فقط'] },
  { q: 'مراحل عمل إضافية', opts: ['لا يوجد', 'تأسيس فقط', 'تجهيز وتشغيل'] },
  { q: 'الاحتياج لكيبل ضوئي', opts: ['متوفر بالموقع', 'يحتاج سحب جديد', 'لا يحتاج'] },
  { q: 'مسافات التسليك الكهربائي', opts: ['قصيرة (<20م)', 'متوسطة', 'يحتاج الى ذرع طويلة جداً'] },
  { q: 'عدد المسؤولين في الموقع', opts: ['شخص واحد', 'لجنة فنية'] },
  { q: 'وجود عروض منافسة', opts: ['لا يوجد', 'نعم (شركات أخرى)', 'قيد المقارنة'] },
  { q: 'مستوى صعوبة التنفيذ', opts: ['سهل/مباشر', 'متوسط', 'معقد/فني'] },
  { q: 'توفر مخططات هندسية', opts: ['متوفرة ورقياً', 'متوفرة PDF', 'غير متوفرة'] },
  { q: 'معايير فنية محددة', opts: ['حسب اختيارنا', 'حسب مواصفاتهم', 'قياسية'] },
  { q: 'وجود جهة قانونية/استشارية', opts: ['لا يوجد', 'مهندس مشرف', 'مكتب استشاري'] },
  { q: 'توفر مصدر كهرباء', opts: ['متوفر مستمر', 'مولدة فقط', 'غير متوفر حالياً'] },
  { q: 'ارتفاعات تزيد عن 4 متر', opts: ['لا توجد', 'نعم (أماكن محددة)', 'نعم (كامل الموقع)'] },
  { q: 'الحاجة لرافعة أو سكلة', opts: ['لا يحتاج', 'سكلة حديدية', 'رافعة هيدروليك'] },
  { q: 'ماركات أجهزة مخصصة', opts: ['أي ماركة جيدة', 'ماركة محددة بالاسم', 'أعلى جودة'] },
  { q: 'مكائن عمل ثقيلة بالموقع', opts: ['لا يوجد', 'نعم (تعيق العمل)', 'نعم (تساعد بالعمل)'] },
]

// قاعدة بيانات أسباب الرفض المنظمة
const REJECTION_REASONS = [
  { value: 'السعر يتجاوز ميزانية العميل', category: 'مالي' },
  { value: 'طريقة الدفع غير مناسبة', category: 'مالي' },
  { value: 'العميل وجد عرضاً أرخص', category: 'مالي' },
  { value: 'العميل ينتظر تمويل', category: 'مالي' },
  { value: 'عدم جدية العميل', category: 'عميل' },
  { value: 'العميل ألغى المشروع', category: 'عميل' },
  { value: 'تأخر العميل في القرار', category: 'عميل' },
  { value: 'العميل غير متوفر', category: 'عميل' },
  { value: 'خلاف بين أصحاب المشروع', category: 'عميل' },
  { value: 'الموقع غير جاهز', category: 'تقني' },
  { value: 'مواصفات غير متوفرة', category: 'تقني' },
  { value: 'صعوبة التنفيذ', category: 'تقني' },
  { value: 'عدم توفر المواد', category: 'تقني' },
  { value: 'الكهرباء غير متوفرة', category: 'تقني' },
  { value: 'عدم حصول على تصاميم', category: 'إداري' },
  { value: 'المشروع متوقف رسمياً', category: 'إداري' },
  { value: 'المناقصة فازت بها شركة', category: 'إداري' },
  { value: 'ظروف خارجة عن الإرادة', category: 'ظروف' },
  { value: 'سبب آخر', category: 'أخرى' },
]

const CATEGORIES = [
  { key: 'all', label: 'الكل', color: '#475569' },
  { key: 'مالي', label: 'مالي', color: '#16a34a' },
  { key: 'عميل', label: 'عميل', color: '#d97706' },
  { key: 'تقني', label: 'تقني', color: '#dc2626' },
  { key: 'إداري', label: 'إداري', color: '#0891b2' },
  { key: 'ظروف', label: 'ظروف', color: '#6b7280' },
  { key: 'أخرى', label: 'أخرى', color: '#1f2937' },
]

function categoryColor(cat: string): string {
  return CATEGORIES.find(c => c.key === cat)?.color || '#dc2626'
}

function parseRejection(task: string | null): { reason: string; notes: string; category: string } {
  let reasonText = (task || 'سبب غير مُحدد').replace(/رفض المشروع\s*[:-]?\s*/i, '')
  const parts = reasonText.split('|')
  reasonText = parts[0].trim() || 'سبب غير مُحدد'
  const notes = parts[1] ? parts[1].replace('ملاحظات:', '').trim() : ''
  const match = REJECTION_REASONS.find(r => reasonText.includes(r.value) || r.value.includes(reasonText))
  return { reason: reasonText, notes, category: match?.category || 'أخرى' }
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------
export default function ProjectsPage() {
  const { employee } = useSession()
  const role = employee?.role
  const canManage = role === 'ADMIN' || role === 'PROJECT_MANAGER'

  const [projects, setProjects] = useState<Project[]>([])
  const [stats, setStats] = useState<Stats>({ اتصال: 0, كشف: 0, سعر: 0, تنفيذ: 0, مكتمل: 0, مرفوض: 0 })
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<'main' | 'rejection'>('main')

  // filters
  const [search, setSearch] = useState('')
  const [filterWorkType, setFilterWorkType] = useState('')
  const [filterStage, setFilterStage] = useState('')

  // modals
  const [showAdd, setShowAdd] = useState(false)
  const [moveTarget, setMoveTarget] = useState<{ project: Project; nextStage: string } | null>(null)
  const [report, setReport] = useState<{ type: 'survey' | 'visit'; project: Project } | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const data = await request<{ projects: Project[]; stats: Stats }>('/projects')
      setProjects(data.projects)
      setStats(data.stats)
    } catch { /* ignore */ }
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  const filtered = useMemo(() => {
    const s = search.toLowerCase().trim()
    return projects.filter(p => {
      const matchSearch = !s ||
        (p.name?.toLowerCase().includes(s)) ||
        (p.phone?.toLowerCase().includes(s)) ||
        (p.location?.toLowerCase().includes(s)) ||
        (p.stage?.toLowerCase().includes(s))
      const matchWork = !filterWorkType || p.workType === filterWorkType
      const matchStage = !filterStage || p.stage.includes(filterStage)
      return matchSearch && matchWork && matchStage
    })
  }, [projects, search, filterWorkType, filterStage])

  const rejectedCount = projects.filter(p => p.stage.includes('مرفوض')).length

  const del = async (id: string) => {
    if (!confirm('حذف المشروع؟')) return
    await request(`/projects/${id}`, { method: 'DELETE' })
    load()
  }

  return (
    <div className="space-y-5">
      {/* Header + view tabs */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-bold text-[var(--color-brand-900)]">📋 إدارة وأرشفة المشاريع</h1>
        <div className="flex gap-2">
          <button onClick={() => setView('main')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${view === 'main' ? 'bg-[var(--color-brand-500)] text-white' : 'bg-white text-gray-600 border'}`}>
            🏠 لوحة التحكم
          </button>
          <button onClick={() => setView('rejection')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 ${view === 'rejection' ? 'bg-red-600 text-white' : 'bg-white text-gray-600 border'}`}>
            🚫 أسباب الرفض
            {rejectedCount > 0 && <span className="bg-red-500 text-white text-xs px-2 py-0.5 rounded-full">{rejectedCount}</span>}
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="w-8 h-8 border-4 border-[var(--color-brand-500)] border-t-transparent rounded-full animate-spin" />
        </div>
      ) : view === 'main' ? (
        <MainView
          projects={filtered}
          totalCount={projects.length}
          stats={stats}
          canManage={canManage}
          search={search} setSearch={setSearch}
          filterWorkType={filterWorkType} setFilterWorkType={setFilterWorkType}
          filterStage={filterStage} setFilterStage={setFilterStage}
          onAdd={() => setShowAdd(true)}
          onMove={(project, nextStage) => setMoveTarget({ project, nextStage })}
          onReport={(type, project) => setReport({ type, project })}
          onDelete={del}
        />
      ) : (
        <RejectionView projects={projects} />
      )}

      {showAdd && <AddModal onClose={() => setShowAdd(false)} onSaved={() => { setShowAdd(false); load() }} />}
      {moveTarget && (
        <MoveModal
          project={moveTarget.project}
          nextStage={moveTarget.nextStage}
          onClose={() => setMoveTarget(null)}
          onSaved={() => { setMoveTarget(null); load() }}
        />
      )}
      {report && <ReportModal type={report.type} project={report.project} onClose={() => setReport(null)} />}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Main dashboard view
// ---------------------------------------------------------------------------
function MainView(props: {
  projects: Project[]
  totalCount: number
  stats: Stats
  canManage: boolean
  search: string; setSearch: (v: string) => void
  filterWorkType: string; setFilterWorkType: (v: string) => void
  filterStage: string; setFilterStage: (v: string) => void
  onAdd: () => void
  onMove: (p: Project, nextStage: string) => void
  onReport: (type: 'survey' | 'visit', p: Project) => void
  onDelete: (id: string) => void
}) {
  const { projects, totalCount, stats, canManage } = props
  const maxStat = Math.max(1, ...STAGE_CARDS.map(c => stats[c.key as keyof Stats]))

  return (
    <>
      {/* Search */}
      <div className="bg-white rounded-xl shadow-sm border p-3 flex items-center gap-2">
        <span className="text-gray-400">🔍</span>
        <input value={props.search} onChange={e => props.setSearch(e.target.value)}
          placeholder="البحث في المشاريع..." className="flex-1 outline-none text-sm bg-transparent" />
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border p-4 grid sm:grid-cols-2 gap-3">
        <div>
          <label className="text-xs text-gray-500 font-medium">نوع العمل</label>
          <select value={props.filterWorkType} onChange={e => props.setFilterWorkType(e.target.value)}
            className="w-full mt-1 border rounded-lg px-3 py-2 text-sm">
            <option value="">📋 كل أنواع العمل</option>
            {WORK_TYPES.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-500 font-medium">المرحلة</label>
          <select value={props.filterStage} onChange={e => props.setFilterStage(e.target.value)}
            className="w-full mt-1 border rounded-lg px-3 py-2 text-sm">
            <option value="">📊 كل المراحل</option>
            {STAGE_CARDS.map(s => <option key={s.key} value={s.key}>{s.icon} {s.label}</option>)}
          </select>
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-3 sm:grid-cols-7 gap-2">
        <button onClick={() => { props.setFilterStage(''); props.setSearch('') }}
          className="bg-gray-800 text-white rounded-xl p-3 flex flex-col items-center justify-center hover:brightness-110 transition">
          <span className="text-xl">🗂️</span>
          <span className="text-xs mt-1">الكل</span>
          <span className="text-lg font-bold">{totalCount}</span>
        </button>
        {STAGE_CARDS.map(c => (
          <button key={c.key} onClick={() => props.setFilterStage(c.key)}
            className={`${c.color} text-white rounded-xl p-3 flex flex-col items-center justify-center hover:brightness-110 transition`}>
            <span className="text-xl">{c.icon}</span>
            <span className="text-xs mt-1">{c.label}</span>
            <span className="text-lg font-bold">{stats[c.key as keyof Stats]}</span>
          </button>
        ))}
      </div>

      {/* Simple pipeline bar chart */}
      <div className="bg-white rounded-xl shadow-sm border p-5">
        <h2 className="text-sm font-semibold mb-3 text-gray-600">توزيع المشاريع على المراحل</h2>
        <div className="space-y-2">
          {STAGE_CARDS.map(c => (
            <div key={c.key} className="flex items-center gap-3">
              <span className="text-xs w-16 text-gray-600">{c.icon} {c.label}</span>
              <div className="flex-1 h-5 bg-gray-100 rounded-full overflow-hidden">
                <div className={`h-full ${c.color} rounded-full transition-all`}
                  style={{ width: `${(stats[c.key as keyof Stats] / maxStat) * 100}%` }} />
              </div>
              <span className="text-xs font-bold w-8">{stats[c.key as keyof Stats]}</span>
            </div>
          ))}
        </div>
      </div>

      {canManage && (
        <button onClick={props.onAdd}
          className="w-full py-3 rounded-xl bg-[var(--color-brand-500)] text-white font-bold shadow hover:brightness-110 transition">
          + إضافة مشروع جديد
        </button>
      )}

      {/* Project cards */}
      <div className="space-y-4">
        {projects.length === 0 && <p className="text-center text-gray-400 py-12 text-lg">لا توجد مشاريع</p>}
        {projects.map(p => (
          <ProjectCard key={p.id} p={p} canManage={canManage}
            onMove={props.onMove} onReport={props.onReport} onDelete={props.onDelete} />
        ))}
      </div>
    </>
  )
}

function ProjectCard({ p, canManage, onMove, onReport, onDelete }: {
  p: Project; canManage: boolean
  onMove: (p: Project, nextStage: string) => void
  onReport: (type: 'survey' | 'visit', p: Project) => void
  onDelete: (id: string) => void
}) {
  const isRejected = p.stage.includes('مرفوض')
  const isCompleted = p.stage.includes('مكتمل')
  const stageIdx = STAGES.indexOf(p.stage)
  const nextStage = STAGES[stageIdx + 1] && stageIdx < 3 ? STAGES[stageIdx + 1] : null
  const borderColor = isRejected ? '#6b7280' : isCompleted ? '#16a34a' : 'var(--color-brand-500)'
  const addedDate = p.createdAt ? new Date(p.createdAt).toLocaleString('ar-IQ') : '---'

  return (
    <div className="bg-white rounded-2xl shadow-sm p-5 relative" style={{ borderRight: `10px solid ${borderColor}`, opacity: isRejected ? 0.92 : 1 }}>
      <div className="absolute top-0 left-0 text-white text-xs px-3 py-1 rounded-bl-xl font-bold"
        style={{ background: isRejected ? '#6b7280' : '#f59e0b' }}>
        📅 تسليم: {p.deliveryDate || '---'}
      </div>

      <h3 className="font-bold text-lg mt-4 flex items-center gap-2" style={{ color: borderColor }}>
        {p.name}
        {p.priority === 'عاجل جداً' && <span className="bg-red-500 text-white text-xs px-2 py-0.5 rounded-full animate-pulse">عاجل 🔥</span>}
        <span className="text-xs text-gray-400 font-normal">{p.code}</span>
      </h3>

      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-2 mt-4">
        <Info icon="📞" value={p.phone} />
        <Info icon="👤" value={p.rep} />
        <Info icon="🤝" value={p.refPerson || 'لا يوجد'} />
        <Info icon="🛠️" value={p.workType} />
        <Info icon="📍" value={p.location} />
        <Info icon="📊" value={p.stage} />
      </div>

      <div className="mt-4 pt-3 border-t flex flex-wrap items-center justify-between gap-2">
        {canManage && (
          <button onClick={() => onDelete(p.id)}
            className="text-sm px-3 py-1.5 rounded-lg border border-red-300 text-red-600 hover:bg-red-50">حذف</button>
        )}
        <div className="flex gap-2 flex-wrap">
          {stageIdx >= 1 && (
            <button onClick={() => onReport('visit', p)}
              className="text-sm px-3 py-1.5 rounded-lg border font-medium hover:bg-gray-50">تقرير الزيارة</button>
          )}
          {stageIdx >= 2 && (
            <button onClick={() => onReport('survey', p)}
              className="text-sm px-3 py-1.5 rounded-lg bg-green-600 text-white font-medium hover:brightness-110">استمارة الكشف</button>
          )}
          {canManage && !isRejected && !isCompleted && nextStage && (
            <button onClick={() => onMove(p, nextStage)}
              className="text-sm px-4 py-1.5 rounded-full bg-gray-800 text-white font-bold hover:brightness-110">ترحيل ⮕</button>
          )}
          {canManage && !isRejected && !isCompleted && (
            <button onClick={() => onMove(p, '❌ مرفوض')}
              className="text-sm px-3 py-1.5 rounded-lg bg-gray-500 text-white font-medium hover:brightness-110">رفض</button>
          )}
          {isCompleted && (
            <span className="text-sm px-3 py-1.5 rounded-full bg-green-600 text-white font-bold flex items-center gap-1">✅ تم التسليم والانتهاء</span>
          )}
        </div>
      </div>

      <div className="mt-3 text-center text-xs text-gray-500 bg-gray-50 rounded-lg py-2">🕐 تاريخ الإضافة: {addedDate}</div>
    </div>
  )
}

function Info({ icon, value }: { icon: string; value: string | null }) {
  return (
    <div className="flex items-center gap-2 bg-gray-50 rounded-lg p-2 text-sm text-gray-700">
      <span>{icon}</span>
      <span className="truncate">{value || '---'}</span>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Add project modal
// ---------------------------------------------------------------------------
function AddModal({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [form, setForm] = useState({ name: '', rep: '', phone: '', location: '', workType: 'طاقة شمسية', refPerson: '', priority: 'عادي', deliveryDate: '' })
  const [saving, setSaving] = useState(false)
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const save = async () => {
    if (!form.name.trim()) { alert('اسم المؤسسة مطلوب'); return }
    setSaving(true)
    try {
      await request('/projects', { method: 'POST', body: JSON.stringify(form) })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title="إضافة مشروع جديد" wide>
      <div className="grid sm:grid-cols-2 gap-3">
        <Field label="المؤسسة *"><input className="inp" value={form.name} onChange={e => set('name', e.target.value)} /></Field>
        <Field label="الممثل"><input className="inp" value={form.rep} onChange={e => set('rep', e.target.value)} /></Field>
        <Field label="الهاتف"><input className="inp" value={form.phone} onChange={e => set('phone', e.target.value)} /></Field>
        <Field label="الموقع"><input className="inp" value={form.location} onChange={e => set('location', e.target.value)} /></Field>
        <Field label="نوع العمل">
          <select className="inp" value={form.workType} onChange={e => set('workType', e.target.value)}>
            {WORK_TYPES.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </Field>
        <Field label="الطرف الوسيط"><input className="inp" value={form.refPerson} onChange={e => set('refPerson', e.target.value)} /></Field>
        <Field label="الأولوية">
          <select className="inp" value={form.priority} onChange={e => set('priority', e.target.value)}>
            <option value="عادي">عادي</option>
            <option value="عاجل جداً">عاجل جداً 🔥</option>
          </select>
        </Field>
        <Field label="تاريخ التسليم"><input type="date" className="inp" value={form.deliveryDate} onChange={e => set('deliveryDate', e.target.value)} /></Field>
      </div>
      <button onClick={save} disabled={saving}
        className="w-full mt-5 py-2.5 rounded-lg bg-[var(--color-brand-500)] text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'حفظ ✅'}
      </button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Move / stage-transition modal
// ---------------------------------------------------------------------------
function MoveModal({ project, nextStage, onClose, onSaved }: {
  project: Project; nextStage: string; onClose: () => void; onSaved: () => void
}) {
  const cur = project.stage
  const isReject = nextStage.includes('مرفوض')
  const toKashf = cur.includes('اتصال') && nextStage.includes('كشف')
  const toSer = cur.includes('كشف') && nextStage.includes('سعر')
  const toExec = cur.includes('سعر') && nextStage.includes('تنفيذ')
  const toDone = cur.includes('تنفيذ') && nextStage.includes('مكتمل')

  const [staff, setStaff] = useState(project.staff || '')
  const [location, setLocation] = useState(project.location || '')
  const [time, setTime] = useState('')
  const [task, setTask] = useState(project.task || '')
  const [price, setPrice] = useState(project.price || '')
  const [survey, setSurvey] = useState<string[]>(SURVEY.map((s, i) => project.survey?.[i] || s.opts[0]))
  const [reason, setReason] = useState('')
  const [otherReason, setOtherReason] = useState('')
  const [notes, setNotes] = useState('')
  const [sent, setSent] = useState(false)
  const [saving, setSaving] = useState(false)

  // زر "إرسال للجروب" يظهر فقط بانتقال اتصال→كشف ويمنع التأكيد حتى يُضغط
  const needsGroupSend = toKashf
  const canConfirm = !needsGroupSend || sent

  const save = async () => {
    if (isReject && !reason) { alert('اختر سبب الرفض!'); return }
    const payload: Record<string, unknown> = { stage: nextStage }
    if (isReject) {
      let finalReason = reason
      if (reason.includes('آخر') && otherReason.trim()) finalReason = otherReason.trim()
      payload.task = `رفض المشروع: ${finalReason}${notes ? ' | ملاحظات: ' + notes : ''}`
    } else {
      if (staff) payload.staff = staff
      if (time) payload.time = time
      if (location) payload.location = location
      if (task) payload.task = task
      if (price) payload.price = price
      if (toSer) payload.survey = survey
      if (needsGroupSend) payload.sentToGroup = true
    }
    setSaving(true)
    try {
      await request(`/projects/${project.id}`, { method: 'PUT', body: JSON.stringify(payload) })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title={`ترحيل إلى: ${nextStage}`}>
      {/* Rejection form */}
      {isReject && (
        <div className="space-y-3">
          <div className="bg-red-50 text-red-700 rounded-lg p-3 text-center text-sm font-bold">⚠️ تحويل المشروع إلى مرفوض</div>
          <Field label="سبب الرفض *">
            <select className="inp border-red-300" value={reason} onChange={e => setReason(e.target.value)}>
              <option value="">-- اختر سبب الرفض --</option>
              {CATEGORIES.filter(c => c.key !== 'all').map(cat => (
                <optgroup key={cat.key} label={cat.label}>
                  {REJECTION_REASONS.filter(r => r.category === cat.key).map(r => (
                    <option key={r.value} value={r.value}>{r.value}</option>
                  ))}
                </optgroup>
              ))}
            </select>
          </Field>
          {reason.includes('آخر') && (
            <input className="inp" placeholder="اكتب السبب بالتفصيل..." value={otherReason} onChange={e => setOtherReason(e.target.value)} />
          )}
          <Field label="ملاحظات إضافية"><textarea className="inp" rows={3} value={notes} onChange={e => setNotes(e.target.value)} /></Field>
        </div>
      )}

      {/* اتصال → كشف : تجهيز الكشف الميداني */}
      {!isReject && toKashf && (
        <div className="space-y-3">
          <div className="bg-green-50 text-green-700 rounded-lg p-3 text-center text-sm font-bold">📋 تجهيز بيانات الكشف الميداني</div>
          <Field label="فريق الكشف (أسماء الفنيين)"><textarea className="inp" rows={2} value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="مكان/موقع الكشف بالتفصيل"><input className="inp" value={location} onChange={e => setLocation(e.target.value)} /></Field>
          <Field label="تاريخ ووقت الكشف المقرر"><input type="datetime-local" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="ملاحظات خاصة بالكشف"><textarea className="inp" rows={2} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      {/* كشف → سعر : استمارة 17 سؤال + المبلغ */}
      {!isReject && toSer && (
        <div className="space-y-2 max-h-96 overflow-y-auto border rounded-lg p-3">
          {SURVEY.map((item, i) => (
            <div key={i}>
              <label className="text-xs font-bold text-gray-600">{i + 1}. {item.q}</label>
              <select className="inp" value={survey[i]} onChange={e => setSurvey(s => { const n = [...s]; n[i] = e.target.value; return n })}>
                {item.opts.map(o => <option key={o} value={o}>{o}</option>)}
              </select>
            </div>
          ))}
          <Field label="المبلغ التخميني"><input className="inp" placeholder="المبلغ المقترح..." value={price} onChange={e => setPrice(e.target.value)} /></Field>
        </div>
      )}

      {/* سعر → تنفيذ : الفني المسؤول */}
      {!isReject && toExec && (
        <Field label="الفني المسؤول عن التنفيذ"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
      )}

      {/* تنفيذ → مكتمل : التسليم النهائي */}
      {!isReject && toDone && (
        <div className="space-y-3">
          <div className="bg-green-50 text-green-700 rounded-lg p-3 text-center text-sm font-bold">✅ إنهاء المشروع والتسليم النهائي</div>
          <Field label="رقم ملف التسليم / المسؤول"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="تاريخ التسليم الفعلي"><input type="date" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="ملاحظات نهائية"><textarea className="inp" rows={2} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      {/* fallback generic */}
      {!isReject && !toKashf && !toSer && !toExec && !toDone && (
        <div className="space-y-3">
          <Field label="الموظف المسؤول"><input className="inp" value={staff} onChange={e => setStaff(e.target.value)} /></Field>
          <Field label="الموعد"><input type="datetime-local" className="inp" value={time} onChange={e => setTime(e.target.value)} /></Field>
          <Field label="الملاحظات"><textarea className="inp" rows={3} value={task} onChange={e => setTask(e.target.value)} /></Field>
        </div>
      )}

      {needsGroupSend && (
        <button onClick={() => setSent(s => !s)}
          className={`w-full my-3 py-2.5 rounded-lg font-bold border-2 ${sent ? 'bg-green-600 text-white border-green-600' : 'border-amber-400 text-amber-600'}`}>
          {sent ? '✅ تم الإرسال للجروب' : 'إرسال البيانات للجروب؟'}
        </button>
      )}

      <button onClick={save} disabled={saving || !canConfirm}
        className="w-full mt-3 py-2.5 rounded-lg bg-green-600 text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'تأكيد ✅'}
      </button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Report modal (survey / visit)
// ---------------------------------------------------------------------------
function ReportModal({ type, project, onClose }: { type: 'survey' | 'visit'; project: Project; onClose: () => void }) {
  return (
    <Modal onClose={onClose} title={type === 'survey' ? 'استمارة الكشف' : 'تقرير الزيارة'} wide>
      <div className="border-2 border-gray-800 p-5 bg-white text-gray-900">
        <div className="grid grid-cols-2 gap-2 border border-gray-800 p-3 text-sm mb-3">
          <div><b>الاسم:</b> {project.name}</div>
          <div><b>الموقع:</b> {project.location || '---'}</div>
          <div><b>الهاتف:</b> {project.phone || '---'}</div>
          <div><b>الممثل:</b> {project.rep || '---'}</div>
        </div>
        {type === 'survey' ? (
          <table className="w-full text-sm border-collapse">
            <thead><tr className="bg-gray-100">
              <th className="border border-gray-800 p-2">#</th>
              <th className="border border-gray-800 p-2 text-right">المعيار</th>
              <th className="border border-gray-800 p-2 text-right">الحالة</th>
            </tr></thead>
            <tbody>
              {SURVEY.map((item, i) => (
                <tr key={i}>
                  <td className="border border-gray-800 p-2 text-center">{i + 1}</td>
                  <td className="border border-gray-800 p-2">{item.q}</td>
                  <td className="border border-gray-800 p-2">{project.survey?.[i] || '---'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className="border border-gray-800 p-3 mt-2">
            <div className="grid grid-cols-2 gap-2 text-sm mb-3">
              <div><b>المسؤول:</b> {project.staff || 'غير محدد'}</div>
              <div><b>الموعد:</b> {project.time || 'غير محدد'}</div>
            </div>
            <b>الملاحظات:</b>
            <p className="mt-1">{project.task || 'لا يوجد'}</p>
          </div>
        )}
      </div>
      <button onClick={() => window.print()} className="w-full mt-4 py-2.5 rounded-lg bg-gray-800 text-white font-bold">🖨️ طباعة</button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Rejection analytics view
// ---------------------------------------------------------------------------
function RejectionView({ projects }: { projects: Project[] }) {
  const [category, setCategory] = useState('all')
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<'count' | 'date' | 'name'>('count')
  const [expanded, setExpanded] = useState<string | null>(null)

  const rejected = projects.filter(p => p.stage.includes('مرفوض'))
  const total = projects.length || 1

  const grouped = useMemo(() => {
    const map: Record<string, { reason: string; category: string; projects: Project[]; lastDate: string | null }> = {}
    for (const p of rejected) {
      const { reason, category: cat } = parseRejection(p.task)
      if (category !== 'all' && cat !== category) continue
      if (!map[reason]) map[reason] = { reason, category: cat, projects: [], lastDate: null }
      map[reason].projects.push(p)
      const d = p.createdAt
      if (d && (!map[reason].lastDate || d > map[reason].lastDate!)) map[reason].lastDate = d
    }
    let list = Object.values(map)
    const s = search.toLowerCase()
    if (s) list = list.filter(r => r.reason.toLowerCase().includes(s) ||
      r.projects.some(p => p.name?.toLowerCase().includes(s) || p.location?.toLowerCase().includes(s)))
    if (sort === 'count') list.sort((a, b) => b.projects.length - a.projects.length)
    else if (sort === 'date') list.sort((a, b) => new Date(b.lastDate || 0).getTime() - new Date(a.lastDate || 0).getTime())
    else list.sort((a, b) => a.reason.localeCompare(b.reason, 'ar'))
    return list
  }, [rejected, category, search, sort])

  const exportCsv = () => {
    if (rejected.length === 0) { alert('لا توجد مشاريع مرفوضة لتصديرها!'); return }
    let csv = '﻿الفئة,السبب,اسم المشروع,الموقع,الهاتف,التاريخ,ملاحظات\n'
    for (const p of rejected) {
      const { reason, notes, category: cat } = parseRejection(p.task)
      if (category !== 'all' && cat !== category) continue
      const date = p.createdAt ? new Date(p.createdAt).toLocaleDateString('ar-IQ') : ''
      csv += `"${cat}","${reason}","${p.name}","${p.location || ''}","${p.phone || ''}","${date}","${notes}"\n`
    }
    const link = document.createElement('a')
    link.href = encodeURI('data:text/csv;charset=utf-8,' + csv)
    link.download = `تقرير_رفض_المشاريع_${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link); link.click(); document.body.removeChild(link)
  }

  const lastDate = grouped[0]?.lastDate ? new Date(grouped[0].lastDate).toLocaleDateString('ar-IQ') : '-'

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-bold text-red-600">🚫 أسباب رفض المشاريع</h2>
        <button onClick={exportCsv} className="px-4 py-2 rounded-full bg-green-600 text-white font-bold text-sm hover:brightness-110">📊 سحب التقرير (Excel)</button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <RejStat label="إجمالي المرفوضات" value={rejected.length} color="from-red-500 to-red-700" />
        <RejStat label="نسبة الرفض" value={`${Math.round((rejected.length / total) * 100)}%`} color="from-amber-400 to-orange-500" />
        <RejStat label="الفئات المختلفة" value={new Set(rejected.map(p => parseRejection(p.task).category)).size} color="from-gray-500 to-gray-700" />
        <RejStat label="آخر رفض" value={lastDate} color="from-cyan-500 to-cyan-700" small />
      </div>

      {/* Category filter */}
      <div className="bg-white rounded-xl shadow-sm border p-4 space-y-3">
        <div className="flex flex-wrap gap-2">
          {CATEGORIES.map(c => (
            <button key={c.key} onClick={() => setCategory(c.key)}
              className="px-4 py-1.5 rounded-full text-sm font-bold border-2 transition"
              style={category === c.key
                ? { background: c.color, color: '#fff', borderColor: c.color }
                : { borderColor: '#e5e7eb', color: '#475569' }}>
              {c.label}
            </button>
          ))}
        </div>
        <div className="grid sm:grid-cols-2 gap-3">
          <input className="inp" placeholder="ابحث في الأسباب أو المشاريع..." value={search} onChange={e => setSearch(e.target.value)} />
          <select className="inp" value={sort} onChange={e => setSort(e.target.value as typeof sort)}>
            <option value="count">الأكثر تكراراً</option>
            <option value="date">الأحدث</option>
            <option value="name">أبجدياً</option>
          </select>
        </div>
      </div>

      {/* Reasons list */}
      {grouped.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <div className="text-5xl mb-3">✅</div>
          <p className="text-lg">لا توجد مشاريع مرفوضة بهذه الفئة</p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {grouped.map(item => {
            const color = categoryColor(item.category)
            const open = expanded === item.reason
            return (
              <div key={item.reason} className="bg-white rounded-2xl shadow-sm p-5 cursor-pointer"
                style={{ borderRight: `6px solid ${color}` }}
                onClick={() => setExpanded(open ? null : item.reason)}>
                <div className="flex items-start justify-between">
                  <span className="text-xs px-2 py-1 rounded-full text-white font-bold" style={{ background: color }}>{item.category}</span>
                  <span className="text-xs px-3 py-1 rounded-full text-white font-bold" style={{ background: color }}>{item.projects.length} مشروع</span>
                </div>
                <h3 className="font-bold text-gray-800 mt-3">{item.reason}</h3>
                <div className="text-xs text-gray-500 mt-2 flex gap-3">
                  <span>📅 {item.lastDate ? new Date(item.lastDate).toLocaleDateString('ar-IQ') : 'غير معروف'}</span>
                  <span>📂 {item.projects.length} حالة</span>
                </div>
                {open && (
                  <div className="bg-gray-50 rounded-lg p-3 mt-3 space-y-2" style={{ borderRight: `3px solid ${color}` }}>
                    <div className="text-xs font-bold text-red-600">المشاريع المرفوضة:</div>
                    {item.projects.map(p => (
                      <div key={p.id} className="flex justify-between items-center border-b last:border-0 pb-1 text-sm">
                        <div>
                          <div className="font-bold">{p.name}</div>
                          <div className="text-xs text-gray-500">{p.location || 'لا يوجد موقع'}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

function RejStat({ label, value, color, small }: { label: string; value: string | number; color: string; small?: boolean }) {
  return (
    <div className={`bg-gradient-to-br ${color} text-white rounded-xl p-4 text-center`}>
      <div className={`font-bold ${small ? 'text-lg' : 'text-3xl'} my-1`}>{value}</div>
      <div className="text-xs opacity-90">{label}</div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Shared UI bits
// ---------------------------------------------------------------------------
function Modal({ children, title, onClose, wide }: { children: React.ReactNode; title: string; onClose: () => void; wide?: boolean }) {
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/50 overflow-y-auto py-8 px-4" onClick={onClose}>
      <div className={`bg-white rounded-2xl shadow-xl w-full ${wide ? 'max-w-2xl' : 'max-w-md'} p-6`} onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b pb-3 mb-4">
          <h3 className="text-lg font-bold text-[var(--color-brand-900)]">{title}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-700 text-xl">✕</button>
        </div>
        {children}
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs text-gray-500 font-medium">{label}</span>
      <div className="mt-1">{children}</div>
    </label>
  )
}
