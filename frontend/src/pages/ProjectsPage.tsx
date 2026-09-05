import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '../session'
import LocationFields from '../components/LocationFields'
import { matches } from '../utils/search'
import BookingCodeChip from '../components/BookingCodeChip'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('authToken')
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options?.headers || {}),
    },
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
  locationUrl: string | null
  createdByName: string | null
  mapLatitude: number | null
  mapLongitude: number | null
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
  bookingId: string | null
  // القائمة ما ترجّع ملفات العقد نفسها (ثقيلة) — بس علمين، والملف ينجلب
  // لما تنفتح نافذة العقد عبر GET /projects/{id}.
  hasContract: boolean
  hasSignedContract: boolean
  responsibleEmployeeId: string | null
  surveyorEmployeeId: string | null
  createdAt: string
  // توجيه المشروع لموظف — يشتغل عليه كأنه عنده إدارة مشاريع، بس على هذا المشروع
  delegatedToEmployeeId: string | null
  delegatedToName: string | null
  delegatedAt: string | null
}

// حجز محول من إدارة الكوادر (الحجوزات الكبيرة تعتبر مشاريع)
interface TransferredBooking {
  id: string
  code: string
  customer: { name: string; phone: string; location?: string | null; code?: string }
  service: { name: string } | null
  address: string | null
  quotedPrice: number | null
  transferToProjects: boolean
  status: string
  createdAt: string
  scheduledAt?: string | null
  notes?: string | null
  adminNotes?: string | null
  confirmedByName?: string | null
  supervisor?: { name: string } | null
}

interface Stats {
  اتصال: number; كشف: number; سعر: number; عقد: number; تنفيذ: number; مكتمل: number; مرفوض: number
}

const STAGES = [
  '1. اتصال بالزبون',
  '2. مرحلة الكشف',
  '3. عرض السعر',
  '4. العقد',
  '5. البدء بالتنفيذ',
  '✅ مكتمل',
  '❌ مرفوض',
]

// أنواع الأعمال كانت قائمة ثابتة بالكود — صارت تنجلب من إعدادات وحدة إدارة
// المشاريع (/project-work-types) حتى يقدر المدير يضيف/يحذف نوع عمل براحته
// بدون تعديل كود. القائمة القديمة تبقى fallback ريثما يوصل جواب السيرفر.
const FALLBACK_WORK_TYPES = ['طاقة شمسية', 'كاميرات', 'بيت ذكي', 'شبكات', 'إنذار حريق', 'أقفال وحاكيات', 'ستلايت', 'منظومة صوت', 'أخرى']

interface ProjectCandidate {
  id: string
  name: string
  role: string
  isLeader: boolean
  isTrainee: boolean
  group: string
  groupLabel: string
  isEngineer: boolean
}

// اسم المرشح بالقائمة يبيّن دوره وحالته — المدير لازم يتأكد إنه اختار
// الشخص الصح، وإذا الموظف لسه متدرب لازم يعرف قبل ما يوجّهه مشروع.
const ROLE_AR: Record<string, string> = {
  ADMIN: 'مدير النظام', OWNER: 'المالك', TECHNICIAN: 'فني', TECHNICAL: 'تقني',
  ENGINEER: 'مهندس', QUALITY_ENGINEER: 'مهندس جودة', SALES: 'مبيعات',
  HR_COORDINATOR: 'كوادر', FINANCE: 'حسابات', MONITOR: 'رقابة',
  PROCUREMENT_ADMIN: 'مخازن', GPS_ADMIN: 'مسؤول GPS', DESIGNER: 'مصمم',
  PROJECT_MANAGER: 'مدير مشاريع',
}
function candidateLabel(c: ProjectCandidate): string {
  const bits = [ROLE_AR[c.role] || c.role]
  if (c.isLeader) bits.push('تيم ليدر')
  if (c.isTrainee) bits.push('⚠ متدرب')
  return `${c.name} — ${bits.join('، ')}`
}

// مرشحو المشروع — "المسؤول عن المشروع" حصراً المهندسون (اللي عندهم مهارات
// تصميم/تخطيط/تنفيذ أو دورهم مهندس)، أما "منفّذ الكشف" فأي موظف بس معروض
// بالتسلسل: مهندسين ← تقنيين ← ليدريه ← فنيين ← إداريين ← مصممين.
function useProjectCandidates(): ProjectCandidate[] {
  const [staff, setStaff] = useState<ProjectCandidate[]>([])
  useEffect(() => {
    request<ProjectCandidate[]>('/project-candidates').then(setStaff).catch(() => {})
  }, [])
  return staff
}

// EmployeeOptions يرسم <optgroup> لكل مجموعة بنفس ترتيب وصولها من السيرفر.
function EmployeeOptions({ candidates }: { candidates: ProjectCandidate[] }) {
  const groups: { label: string; items: ProjectCandidate[] }[] = []
  candidates.forEach(c => {
    const last = groups[groups.length - 1]
    if (last && last.label === c.groupLabel) last.items.push(c)
    else groups.push({ label: c.groupLabel, items: [c] })
  })
  return (
    <>
      {groups.map(g => (
        <optgroup key={g.label} label={g.label}>
          {g.items.map(c => <option key={c.id} value={c.id}>{candidateLabel(c)}</option>)}
        </optgroup>
      ))}
    </>
  )
}

function useProjectWorkTypes(): string[] {
  const [types, setTypes] = useState<string[]>(FALLBACK_WORK_TYPES)
  useEffect(() => {
    request<{ id: string; name: string }[]>('/project-work-types')
      .then((rows) => { if (rows.length > 0) setTypes(rows.map((r) => r.name)) })
      .catch(() => {})
  }, [])
  return types
}

const STAGE_CARDS = [
  { key: 'اتصال', label: 'اتصال', icon: '📞', color: 'bg-[var(--color-brand-500)]' },
  { key: 'كشف', label: 'كشف', icon: '🔍', color: 'bg-green-700' },
  { key: 'سعر', label: 'سعر', icon: '💰', color: 'bg-amber-500' },
  { key: 'عقد', label: 'عقد', icon: '📄', color: 'bg-purple-600' },
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
  { key: 'all', label: 'الكل', color: 'var(--t-muted)' },
  { key: 'مالي', label: 'مالي', color: 'var(--t-success)' },
  { key: 'عميل', label: 'عميل', color: 'var(--t-warning)' },
  { key: 'تقني', label: 'تقني', color: 'var(--t-danger)' },
  { key: 'إداري', label: 'إداري', color: 'var(--t-cyan)' },
  { key: 'ظروف', label: 'ظروف', color: 'var(--t-muted)' },
  { key: 'أخرى', label: 'أخرى', color: 'var(--t-body)' },
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
// mode='delegated' يعرض بس المشاريع الموجّهة للموظف الحالي — نفس الواجهة
// والمراحل والتقارير بالضبط، بس على مشاريعه هو. هذا معنى "كأنه عنده صلاحية
// إدارة مشاريع بس على هذا المشروع".
// ═══ خياران من فوگ بدل بندين بالقائمة ═══
//
// «المشاريع نفس الطريقة — من أضغط عليها تفتح واجهة أختار إضافة
// مشروع لو المشاريع الموجّهة لي».
//
// نفس نمط «مهامي» و«الجرد» و«التقييم»: بند واحد بالقائمة، والاختيار
// جوّا الشاشة. الوضع يجي من الرابط أول مرة (‎/my-projects‎ يفتح على
// الموجّهة لي) ويتبدّل بضغطة بلا ما تطلع للقائمة.
export default function ProjectsPage({ mode: initialMode = 'all' }: { mode?: 'all' | 'delegated' } = {}) {
  const { employee, permissions } = useSession()
  const role = employee?.role
  const [mode, setMode] = useState<'all' | 'delegated'>(initialMode)
  const delegatedMode = mode === 'delegated'
  const canManage = delegatedMode
    || role === 'ADMIN' || role === 'PROJECT_MANAGER' || permissions.includes('project_management')
  // صلاحية "إضافة مشروع فقط": يشوف واجهة إضافة نظيفة بس — بدون إحصائيات ولا
  // قائمة مشاريع ولا تقارير ولا استمارة كشف ولا ترحيل مراحل.
  const addOnly = !canManage && permissions.includes('project_create_only')

  const [projects, setProjects] = useState<Project[]>([])
  const [stats, setStats] = useState<Stats>({ اتصال: 0, كشف: 0, سعر: 0, عقد: 0, تنفيذ: 0, مكتمل: 0, مرفوض: 0 })
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<'main' | 'rejection'>('main')

  // filters
  const [search, setSearch] = useState('')
  const [filterWorkType, setFilterWorkType] = useState('')
  const [filterStage, setFilterStage] = useState('')

  // modals
  const [showAdd, setShowAdd] = useState(false)
  const [editProject, setEditProject] = useState<Project | null>(null)
  const [moveTarget, setMoveTarget] = useState<{ project: Project; nextStage: string } | null>(null)
  const [report, setReport] = useState<{ type: 'survey' | 'visit'; project: Project } | null>(null)
  const [transferred, setTransferred] = useState<TransferredBooking[]>([])
  // تفاصيل الحجز المحوّل — منها يقرر يستلمه كمشروع أو يرجعه لكادر الشد
  const [detailBooking, setDetailBooking] = useState<TransferredBooking | null>(null)
  // المشروع الي راح ينسلّم لموظف
  const [delegateTarget, setDelegateTarget] = useState<Project | null>(null)
  const [returnNote, setReturnNote] = useState('')
  const [returning, setReturning] = useState(false)

  // ⚠️ الجلب ينطلب برفع عدّاد بدل نداء مباشر: نداء دالة تحدّث الحالة
  // من جسم الـeffect يسبّب دورة رسم زايدة، وبنفس الوقت يخلّي مسار
  // الجلب موزّع على مكانين. هسه مكان واحد، وكل الي يريد تحديث
  // ينادي `refresh()`.
  const [reload, setReload] = useState(0)
  const refresh = () => setReload((n) => n + 1)


  // إرجاع الحجز لكادر الشد لما يتبين إنه مو مال مشروع
  const returnBookingToCrew = async (b: TransferredBooking) => {
    // السبب إجباري — إداري الكوادر لازم يعرف ليش رجع له الحجز حتى يتصرف،
    // وإلا يرجع له بلا معلومة ويضيع.
    if (!returnNote.trim()) {
      alert('لازم تكتب سبب الإرجاع — إداري الكوادر يحتاجه حتى يعرف شنو يسوي بالحجز')
      return
    }
    setReturning(true)
    try {
      await request(`/bookings/${b.id}/return-to-crew`, {
        method: 'PUT',
        body: JSON.stringify({ note: returnNote.trim() }),
      })
      // نشيله من القائمة فوراً بدل ما ننتظر إعادة تحميل كاملة
      setTransferred((prev) => prev.filter((x) => x.id !== b.id))
      setDetailBooking(null)
      setReturnNote('')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر إرجاع الحجز')
    } finally {
      setReturning(false)
    }
  }

  const receiveBooking = async (b: TransferredBooking) => {
    if (!confirm(`استلام حجز ${b.customer.name} كمشروع جديد؟`)) return
    try {
      await request('/projects', {
        method: 'POST',
        body: JSON.stringify({
          name: b.customer.name,
          rep: b.customer.name,
          phone: b.customer.phone,
          location: b.address,
          workType: b.service?.name || null,
          bookingId: b.id,
        }),
      })
      refresh()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر استلام الحجز')
    }
  }

  // ⚠️ `delegatedMode` بالاعتماديات: تبديل «المفوّضة لي» يبدّل المسار
  // الي ينجلب منه، وبدونها چان يبقى يعرض بيانات الوضع السابق.
  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        // الطلبين بالتوازي بدل واحد يستنى الثاني — كانوا متسلسلين وهذا
        // يضاعف زمن فتح الصفحة بلا داعي.
        const [data, bookings] = await Promise.all([
          request<{ projects: Project[]; stats: Stats }>(
            delegatedMode ? '/projects/delegated-to-me' : '/projects'),
          // الحجوزات الفعّالة بس — الأرشيف الكامل كان يوصل مئات
          // الكيلوبايتات ويكبر مع الوقت، مع إننا نحتاج بس المحوّلة الي
          // لسه ما انستلمت.
          delegatedMode
            ? Promise.resolve([] as TransferredBooking[])
            : request<TransferredBooking[]>('/bookings?status=PENDING,CONFIRMED'),
        ])
        if (!alive) return
        setProjects(data.projects)
        setStats(data.stats)
        // الحجوزات المحوّلة لإدارة المشاريع والي لسه ما انعمل منها مشروع
        const linked = new Set(data.projects.map((p) => p.bookingId).filter(Boolean))
        setTransferred(bookings.filter((b) =>
          b.transferToProjects && b.status !== 'COMPLETED' && b.status !== 'CANCELLED' && !linked.has(b.id)
        ))
      } catch { /* ignore */ }
      if (alive) setLoading(false)
    })()
    return () => { alive = false }
  }, [reload, delegatedMode])

  const filtered = useMemo(() => {
    return projects.filter(p => {
      const matchSearch = matches([p.name, p.code, p.phone, p.location, p.rep, p.stage], search)
      const matchWork = !filterWorkType || p.workType === filterWorkType
      const matchStage = !filterStage || p.stage.includes(filterStage)
      return matchSearch && matchWork && matchStage
    })
  }, [projects, search, filterWorkType, filterStage])

  const rejectedCount = projects.filter(p => p.stage.includes('مرفوض')).length

  const del = async (id: string) => {
    if (!confirm('حذف المشروع؟')) return
    await request(`/projects/${id}`, { method: 'DELETE' })
    refresh()
  }

  // واجهة "إضافة مشروع فقط" — نظيفة ومبسّطة، بلا أي إحصائيات/قوائم/تقارير
  if (addOnly) {
    return (
      <div dir="rtl" className="mx-auto max-w-lg">
        <div className="rounded-2xl bg-gradient-to-l from-[var(--color-brand-500)] to-[var(--color-brand-900)] p-6 text-white shadow-lg">
          <h1 className="text-2xl font-bold">➕ إضافة مشروع جديد</h1>
          <p className="mt-1 text-sm text-blue-100/90">سجّل بيانات المشروع وموقعه، وفريق إدارة المشاريع راح يتابعه من هناك.</p>
        </div>

        <button
          onClick={() => setShowAdd(true)}
          className="mt-6 w-full rounded-2xl border-2 border-dashed border-[var(--color-brand-500)] bg-white py-10 text-center text-lg font-bold text-[var(--color-brand-900)] transition-colors hover:bg-blue-50"
        >
          <span className="block text-4xl">📋</span>
          اضغط هنا لإضافة مشروع
        </button>

        {showAdd && <AddModal onClose={() => setShowAdd(false)} onSaved={() => { setShowAdd(false); alert('تم إضافة المشروع بنجاح ✅') }} />}
      </div>
    )
  }

  return (
    <div className="space-y-5">
      {/* Header + view tabs */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-bold text-[var(--color-brand-900)]">
          {delegatedMode ? '📤 المشاريع الموجّهة لي' : '📋 إدارة وأرشفة المشاريع'}
        </h1>
        <div className="flex gap-2">
          <button onClick={() => setView('main')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${view === 'main' ? 'bg-[var(--color-brand-500)] text-white' : 'bg-white text-gray-600 border'}`}>
            🏠 لوحة التحكم
          </button>
          <button onClick={() => setView('rejection')}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors flex items-center gap-2 ${view === 'rejection' ? 'bg-red-600 text-white' : 'bg-white text-gray-600 border'}`}>
            🚫 أسباب الرفض
            {rejectedCount > 0 && <span className="bg-red-600 text-white text-xs px-2 py-0.5 rounded-full">{rejectedCount}</span>}
          </button>
        </div>
      </div>

      {/* ═══ الخياران ═══
          ⚠️ «كل المشاريع» ما يطلع إلا لمن يديرها فعلاً — الفني الي
          عنده مشاريع موجّهة له بس ما يشوف مشاريع الشركة كلها. */}
      {canManage && !addOnly && (
        <div className="inline-flex flex-wrap gap-1.5 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-sm">
          {([
            { k: 'all' as const, label: '🏗️ كل المشاريع' },
            { k: 'delegated' as const, label: '📤 المشاريع الموجّهة لي' },
          ]).map((o) => (
            <button
              key={o.k}
              onClick={() => setMode(o.k)}
              className={`rounded-xl px-4 py-2 text-xs font-extrabold transition ${
                mode === o.k
                  ? 'bg-gradient-to-l from-[var(--color-brand-500)] to-[var(--color-brand-800)] text-white shadow-md'
                  : 'text-gray-600 hover:bg-slate-50'
              }`}
            >
              {o.label}
            </button>
          ))}
          {/* «إضافة مشروع» زر مباشر جنب الخيارين — كان مدفون تحت
              الإحصائيات والرسم البياني، والي جاي يضيف مشروع يريده
              بأول الشاشة. */}
          <button
            onClick={() => setShowAdd(true)}
            className="rounded-xl bg-emerald-700 px-4 py-2 text-xs font-extrabold text-white shadow-md hover:bg-emerald-800"
          >
            ＋ إضافة مشروع
          </button>
        </div>
      )}

      {/* حجوزات محولة من إدارة الكوادر — بانتظار الاستلام كمشاريع */}
      {transferred.length > 0 && (
        <div className="rounded-2xl border-2 border-violet-200 bg-violet-50 p-5">
          <h3 className="mb-3 flex items-center gap-2 text-base font-extrabold text-violet-900">
            📥 حجوزات محولة للمشاريع
            <span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-violet-600 text-xs font-black text-white">{transferred.length}</span>
          </h3>
          <div className="flex flex-col gap-2">
            {transferred.map(b => (
              <div key={b.id} className="flex flex-wrap items-center justify-between gap-3 rounded-xl bg-white p-4 shadow-sm">
                <div>
                  <p className="font-bold text-slate-800">{b.customer.name} <span className="text-xs font-normal text-slate-400">(<BookingCodeChip code={b.code} />)</span></p>
                  <p className="text-sm text-slate-500">
                    📞 {b.customer.phone}
                    {b.service ? ` · 🛠️ ${b.service.name}` : ''}
                    {b.address ? ` · 📍 ${b.address}` : ''}
                    {b.quotedPrice ? ` · 💰 ${b.quotedPrice.toLocaleString('ar-IQ')} د.ع` : ''}
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button onClick={() => setDetailBooking(b)}
                    className="rounded-lg border border-violet-300 bg-white px-4 py-2 text-sm font-bold text-violet-700 hover:bg-violet-50">
                    📋 التفاصيل
                  </button>
                  <button onClick={() => receiveBooking(b)}
                    className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-bold text-white hover:bg-violet-700">
                    استلام كمشروع ←
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

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
          onAdd={delegatedMode ? undefined : () => setShowAdd(true)}
          onEdit={(p) => setEditProject(p)}
          onMove={(project, nextStage) => setMoveTarget({ project, nextStage })}
          onReport={(type, project) => setReport({ type, project })}
          onDelete={del}
          onRefresh={refresh}
          onDelegate={delegatedMode ? undefined : (p) => setDelegateTarget(p)}
        />
      ) : (
        <RejectionView projects={projects} />
      )}

      {showAdd && <AddModal onClose={() => setShowAdd(false)} onSaved={() => { setShowAdd(false); refresh() }} />}
      {editProject && <EditModal project={editProject} onClose={() => setEditProject(null)} onSaved={() => { setEditProject(null); refresh() }} />}
      {moveTarget && (
        <MoveModal
          project={moveTarget.project}
          nextStage={moveTarget.nextStage}
          onClose={() => setMoveTarget(null)}
          onSaved={() => { setMoveTarget(null); refresh() }}
        />
      )}
      {report && <ReportModal type={report.type} project={report.project} onClose={() => setReport(null)} />}
      {delegateTarget && (
        <DelegateModal
          project={delegateTarget}
          onClose={() => setDelegateTarget(null)}
          onSaved={() => { setDelegateTarget(null); refresh() }}
        />
      )}

      {/* تفاصيل الحجز المحوّل — يقررمنها: يستلمه كمشروع أو يرجعه لكادر الشد */}
      {detailBooking && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
          <div className="max-h-[88vh] w-full max-w-lg overflow-auto rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-violet-900">تفاصيل الحجز</h3>
            <p className="mt-1 text-sm text-slate-400"><BookingCodeChip code={detailBooking.code} /></p>

            <div className="mt-4 space-y-2 rounded-xl bg-slate-50 p-4 text-sm">
              {[
                ['الزبون', detailBooking.customer.name],
                ['الهاتف', detailBooking.customer.phone],
                ['كود الزبون', detailBooking.customer.code],
                ['الخدمة', detailBooking.service?.name],
                ['العنوان', detailBooking.address || detailBooking.customer.location],
                ['المبلغ المقدّر', detailBooking.quotedPrice ? `${detailBooking.quotedPrice.toLocaleString('ar-IQ')} د.ع` : null],
                ['موعد التنفيذ', detailBooking.scheduledAt ? new Date(detailBooking.scheduledAt).toLocaleString('ar-IQ') : null],
                ['ثبّته', detailBooking.confirmedByName],
                ['المشرف', detailBooking.supervisor?.name],
                ['تاريخ الحجز', new Date(detailBooking.createdAt).toLocaleDateString('ar-IQ')],
              ].filter(([, v]) => v).map(([k, v]) => (
                <div key={k as string} className="flex justify-between gap-4">
                  <span className="text-slate-400">{k}</span>
                  <span className="text-left font-medium text-slate-800">{v}</span>
                </div>
              ))}
              {detailBooking.notes && (
                <div className="border-t border-slate-200 pt-2">
                  <div className="text-slate-400">ملاحظات الحجز</div>
                  <div className="mt-0.5 text-slate-700">{detailBooking.notes}</div>
                </div>
              )}
              {detailBooking.adminNotes && (
                <div className="border-t border-slate-200 pt-2">
                  <div className="text-slate-400">ملاحظات الإداري</div>
                  <div className="mt-0.5 text-slate-700">{detailBooking.adminNotes}</div>
                </div>
              )}
            </div>

            <div className="mt-5 rounded-xl border border-amber-200 bg-amber-50 p-4">
              <p className="text-sm font-bold text-amber-800">مو مال مشروع؟ رجّعه لكادر الشد</p>
              <p className="mt-0.5 text-xs text-amber-700">سبب الإرجاع إجباري — يوصل لإداري الكوادر</p>
              <input
                value={returnNote}
                onChange={(e) => setReturnNote(e.target.value)}
                placeholder="ليش رجعته؟ مثال: شغلة شد عادية مو مشروع"
                className="mt-2 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-amber-500"
              />
              <button
                onClick={() => returnBookingToCrew(detailBooking)}
                disabled={returning || !returnNote.trim()}
                className="mt-2 w-full rounded-lg bg-amber-600 px-4 py-2 text-sm font-bold text-white hover:bg-amber-700 disabled:opacity-50"
              >
                {returning ? 'جاري الإرجاع...' : '↩️ إعادة الترحيل لكادر الشد'}
              </button>
            </div>

            <div className="mt-5 flex gap-3">
              <button
                onClick={() => { const b = detailBooking; setDetailBooking(null); receiveBooking(b) }}
                className="flex-1 rounded-xl bg-violet-600 py-2.5 font-bold text-white hover:bg-violet-700"
              >
                استلام كمشروع ←
              </button>
              <button
                onClick={() => { setDetailBooking(null); setReturnNote('') }}
                className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600 hover:bg-slate-50"
              >
                إغلاق
              </button>
            </div>
          </div>
        </div>
      )}
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
  onAdd?: () => void
  onEdit: (p: Project) => void
  onMove: (p: Project, nextStage: string) => void
  onReport: (type: 'survey' | 'visit', p: Project) => void
  onDelete: (id: string) => void
  onRefresh: () => void
  onDelegate?: (p: Project) => void
}) {
  const { projects, totalCount, stats, canManage } = props
  const maxStat = Math.max(1, ...STAGE_CARDS.map(c => stats[c.key as keyof Stats]))
  const workTypes = useProjectWorkTypes()

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
            {workTypes.map(w => <option key={w} value={w}>{w}</option>)}
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

      {canManage && props.onAdd && (
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
            onEdit={props.onEdit} onMove={props.onMove} onReport={props.onReport} onDelete={props.onDelete}
            onRefresh={props.onRefresh} onDelegate={props.onDelegate} />
        ))}
      </div>
    </>
  )
}

function ProjectCard({ p, canManage, onEdit, onMove, onReport, onDelete, onRefresh, onDelegate }: {
  p: Project; canManage: boolean
  onEdit: (p: Project) => void
  onMove: (p: Project, nextStage: string) => void
  onReport: (type: 'survey' | 'visit', p: Project) => void
  onDelete: (id: string) => void
  onRefresh: () => void
  onDelegate?: (p: Project) => void
}) {
  const navigate = useNavigate()
  const isRejected = p.stage.includes('مرفوض')
  const isCompleted = p.stage.includes('مكتمل')
  const isContractStage = p.stage.includes('عقد')
  const stageIdx = STAGES.indexOf(p.stage)
  const nextStage = STAGES[stageIdx + 1] && stageIdx <= 4 ? STAGES[stageIdx + 1] : null
  const [showContract, setShowContract] = useState(false)

  const handleAdvance = () => {
    if (isContractStage && !p.hasContract) {
      alert('ارفع ملف العقد (PDF) الأول قبل الترحيل لمرحلة التنفيذ.')
      setShowContract(true)
      return
    }
    if (nextStage) onMove(p, nextStage)
  }
  const borderColor = isRejected ? '#6b7280' : isCompleted ? '#16a34a' : 'var(--color-brand-500)'
  const addedDate = p.createdAt ? new Date(p.createdAt).toLocaleString('ar-IQ') : '---'

  return (
    <div className="bg-white rounded-2xl shadow-sm p-5 relative" style={{ borderRight: `10px solid ${borderColor}`, opacity: isRejected ? 0.92 : 1 }}>
      <div className="absolute top-0 left-0 text-white text-xs px-3 py-1 rounded-bl-xl font-bold"
        style={{ background: isRejected ? '#6b7280' : '#f59e0b' }}>
        📅 انتهاء تقريبي: {p.deliveryDate || '---'}
      </div>

      {/* المرحلة بأعلى وسط البطاقة — أهم معلومة، تنقرا بلمحة بدون تدوير */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2">
        <span className="inline-block rounded-b-xl px-4 py-1 text-xs font-extrabold text-white shadow-sm"
          style={{ background: borderColor }}>
          {p.stage}
        </span>
      </div>

      <h3 className="font-bold text-lg mt-4 flex items-center gap-2" style={{ color: borderColor }}>
        {p.name}
        {p.priority === 'عاجل جداً' && <span className="bg-red-600 text-white text-xs px-2 py-0.5 rounded-full animate-pulse">عاجل 🔥</span>}
        <span className="text-xs text-gray-400 font-normal">{p.code}</span>
      </h3>

      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-2 mt-4">
        <Info icon="📞" value={p.phone} />
        <Info icon="👤" value={p.rep} />
        <Info icon="🤝" value={p.refPerson || 'لا يوجد'} />
        <Info icon="🛠️" value={p.workType} />
        <Info icon="📍" value={p.location} />
        {/* مكان المرحلة القديم: منو رحّل الحجز أو أضاف المشروع */}
        <Info icon="🧑‍💼" value={p.createdByName ? `أضافه: ${p.createdByName}` : 'أضافه: غير محدد'} />
        {p.delegatedToName && <Info icon="📤" value={`موجّه إلى: ${p.delegatedToName}`} />}
      </div>

      <div className="mt-4 pt-3 border-t flex flex-wrap items-center justify-between gap-2">
        {canManage && (
          <div className="flex gap-2">
            <button onClick={() => onEdit(p)}
              className="text-sm px-3 py-1.5 rounded-lg border border-blue-300 text-blue-600 hover:bg-blue-50">تعديل ✏️</button>
            <button onClick={() => onDelete(p.id)}
              className="text-sm px-3 py-1.5 rounded-lg border border-red-300 text-red-600 hover:bg-red-50">حذف</button>
            {onDelegate && (
              <button onClick={() => onDelegate(p)}
                className="text-sm px-3 py-1.5 rounded-lg border border-violet-300 text-violet-700 hover:bg-violet-50">
                {p.delegatedToEmployeeId ? '👤 تغيير الموظف الموجّه له' : '📤 توجيه لموظف'}
              </button>
            )}
          </div>
        )}
        <div className="flex gap-2 flex-wrap">
          {stageIdx >= 1 && (
            <button onClick={() => onReport('visit', p)}
              className="text-sm px-3 py-1.5 rounded-lg border font-medium hover:bg-gray-50">تقرير الزيارة</button>
          )}
          {stageIdx >= 2 && (
            <button onClick={() => onReport('survey', p)}
              className="text-sm px-3 py-1.5 rounded-lg bg-green-700 text-white font-medium hover:brightness-110">استمارة الكشف</button>
          )}
          {/* بين الكشف والسعر: يفتح نظام عرض السعر الموجود عدنا، مع تعبئة بيانات
              الزبون والمشروع تلقائياً — بلون مميّز عن باقي الأزرار. */}
          {canManage && (p.stage.includes('كشف') || p.stage.includes('سعر')) && (
            <button
              onClick={() => {
                const q = new URLSearchParams({
                  customerName: p.name,
                  ...(p.phone ? { customerPhone: p.phone } : {}),
                  ...(p.location ? { customerAddress: p.location } : {}),
                  projectName: `${p.code} — ${p.workType || 'مشروع'}`,
                  // حتى يطلع زر "تم" بعرض السعر ويرجعه لإدارة المشاريع
                  // بدل ما يضل يتنقل بالماوس بين الصفحات
                  returnTo: '/projects',
                })
                navigate(`/quotations/new?${q.toString()}`)
              }}
              className="text-sm px-3 py-1.5 rounded-lg bg-cyan-600 text-white font-bold hover:brightness-110">
              🧾 اعمل عرض سعر
            </button>
          )}
          {canManage && (isContractStage || p.hasContract) && (
            <button onClick={() => setShowContract(true)}
              className="text-sm px-3 py-1.5 rounded-lg bg-purple-600 text-white font-medium hover:brightness-110">
              📄 العقد{p.hasContract ? (p.hasSignedContract ? ' (مرفوع وموقّع)' : ' (مرفوع)') : ''}
            </button>
          )}
          {canManage && !isRejected && !isCompleted && nextStage && (
            <button onClick={handleAdvance}
              className="text-sm px-4 py-1.5 rounded-full bg-gray-800 text-white font-bold hover:brightness-110">ترحيل ⮕</button>
          )}
          {canManage && !isRejected && !isCompleted && (
            <button onClick={() => onMove(p, '❌ مرفوض')}
              className="text-sm px-3 py-1.5 rounded-lg bg-gray-500 text-white font-medium hover:brightness-110">رفض</button>
          )}
          {isCompleted && (
            <span className="text-sm px-3 py-1.5 rounded-full bg-green-700 text-white font-bold flex items-center gap-1">✅ تم التسليم والانتهاء</span>
          )}
        </div>
      </div>

      <div className="mt-3 text-center text-xs text-gray-500 bg-gray-50 rounded-lg py-2">🕐 تاريخ الإضافة: {addedDate}</div>

      {showContract && (
        <ContractModal project={p} onClose={() => setShowContract(false)} onSaved={() => { setShowContract(false); onRefresh() }} />
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Contract modal — رفع عقد المشروع كـPDF (قبل التوقيع وبعده)
// ---------------------------------------------------------------------------
function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function ContractModal({ project, onClose, onSaved }: { project: Project; onClose: () => void; onSaved: () => void }) {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'
  const [saving, setSaving] = useState<'plain' | 'signed' | null>(null)
  const [error, setError] = useState('')
  // ملفات العقد ما تجي بالقائمة (ثقيلة) — نجيبها هنا بس لما تنفتح النافذة.
  const [files, setFiles] = useState<{ contractPdfBase64: string | null; signedContractPdfBase64: string | null } | null>(null)
  useEffect(() => {
    request<{ contractPdfBase64: string | null; signedContractPdfBase64: string | null }>(`/projects/${project.id}`)
      .then(setFiles)
      .catch(() => setFiles({ contractPdfBase64: null, signedContractPdfBase64: null }))
  }, [project.id])

  const upload = async (file: File | undefined, field: 'contractPdfBase64' | 'signedContractPdfBase64') => {
    if (!file) return
    if (file.type !== 'application/pdf') { setError('لازم يكون الملف بصيغة PDF'); return }
    setError('')
    setSaving(field === 'contractPdfBase64' ? 'plain' : 'signed')
    try {
      const base64 = await fileToBase64(file)
      await request(`/projects/${project.id}`, { method: 'PUT', body: JSON.stringify({ [field]: base64 }) })
      onSaved()
    } catch (e) {
      setError((e as Error).message || 'تعذر رفع الملف')
    } finally {
      setSaving(null)
    }
  }

  // حذف العقد المرفوع — لمدير النظام حصراً (والسيرفر يفرضها كمان).
  const removeContract = async (which: 'plain' | 'signed') => {
    if (!confirm('حذف ملف العقد هذا نهائياً؟')) return
    setError('')
    try {
      await request(`/projects/${project.id}/contract?which=${which}`, { method: 'DELETE' })
      setFiles(f => f ? { ...f, [which === 'plain' ? 'contractPdfBase64' : 'signedContractPdfBase64']: null } : f)
      onSaved()
    } catch (e) {
      setError((e as Error).message || 'تعذر حذف الملف')
    }
  }

  return (
    <Modal onClose={onClose} title={`عقد المشروع: ${project.name}`}>
      <div className="space-y-4">
        <div>
          <label className="text-xs font-bold text-gray-600">العقد قبل التوقيع (PDF)</label>
          <input
            type="file" accept="application/pdf"
            onChange={(e) => upload(e.target.files?.[0], 'contractPdfBase64')}
            className="mt-1 block w-full text-sm"
          />
          {files?.contractPdfBase64 && (
            <div className="mt-1 flex items-center gap-3">
              <a href={files.contractPdfBase64} download={`عقد-${project.code}.pdf`} className="text-xs text-brand-600 hover:underline">
                📄 تحميل الملف المرفوع حالياً
              </a>
              {isAdmin && (
                <button type="button" onClick={() => removeContract('plain')} className="text-xs font-bold text-red-600 hover:underline">
                  حذف الملف
                </button>
              )}
            </div>
          )}
          {saving === 'plain' && <p className="text-xs text-gray-400">جاري الرفع...</p>}
        </div>
        <div>
          <label className="text-xs font-bold text-gray-600">العقد بعد التوقيع (PDF)</label>
          <input
            type="file" accept="application/pdf"
            onChange={(e) => upload(e.target.files?.[0], 'signedContractPdfBase64')}
            className="mt-1 block w-full text-sm"
          />
          {files?.signedContractPdfBase64 && (
            <div className="mt-1 flex items-center gap-3">
              <a href={files.signedContractPdfBase64} download={`عقد-موقّع-${project.code}.pdf`} className="text-xs text-brand-600 hover:underline">
                📄 تحميل الملف الموقّع حالياً
              </a>
              {isAdmin && (
                <button type="button" onClick={() => removeContract('signed')} className="text-xs font-bold text-red-600 hover:underline">
                  حذف الملف
                </button>
              )}
            </div>
          )}
          {saving === 'signed' && <p className="text-xs text-gray-400">جاري الرفع...</p>}
        </div>
        {error && <p className="text-sm font-bold text-red-600">{error}</p>}
      </div>
    </Modal>
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
  const workTypes = useProjectWorkTypes()
  const [form, setForm] = useState({ name: '', rep: '', phone: '', location: '', workType: 'طاقة شمسية', refPerson: '', priority: 'عادي', deliveryDate: '' })
  const [mapPoint, setMapPoint] = useState<{ lat: number; lng: number } | null>(null)
  const [locationUrl, setLocationUrl] = useState('')
  const [saving, setSaving] = useState(false)
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const save = async () => {
    if (!form.name.trim()) { alert('اسم المؤسسة مطلوب'); return }
    setSaving(true)
    try {
      await request('/projects', {
        method: 'POST',
        body: JSON.stringify({ ...form, mapLatitude: mapPoint?.lat, mapLongitude: mapPoint?.lng, locationUrl: locationUrl || undefined }),
      })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title="إضافة مشروع جديد" wide>
      <div className="grid sm:grid-cols-2 gap-3">
        <Field label="المؤسسة *"><input className="inp" value={form.name} onChange={e => set('name', e.target.value)} /></Field>
        <Field label="الممثل"><input className="inp" value={form.rep} onChange={e => set('rep', e.target.value)} /></Field>
        <Field label="الهاتف"><input className="inp" value={form.phone} onChange={e => set('phone', e.target.value)} /></Field>
        <div className="sm:col-span-2">
          {/* نفس آلية الموردين: عنوان + رابط + خريطة، والرابط يغني عن الخريطة */}
          <LocationFields
            addressLabel="عنوان الموقع"
            address={form.location}
            onAddressChange={(v) => set('location', v)}
            point={mapPoint}
            onPointChange={setMapPoint}
            locationUrl={locationUrl}
            onLocationUrlChange={setLocationUrl}
            resolveUrl={(u) => request<{ lat: number; lng: number }>(`/geo/resolve-map-link?url=${encodeURIComponent(u)}`)}
          />
        </div>
        <Field label="نوع العمل">
          <select className="inp" value={form.workType} onChange={e => set('workType', e.target.value)}>
            {workTypes.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </Field>
        <Field label="الطرف الوسيط"><input className="inp" value={form.refPerson} onChange={e => set('refPerson', e.target.value)} /></Field>
        <Field label="الأولوية">
          <select className="inp" value={form.priority} onChange={e => set('priority', e.target.value)}>
            <option value="عادي">عادي</option>
            <option value="عاجل جداً">عاجل جداً 🔥</option>
          </select>
        </Field>
        <Field label="تاريخ انتهاء المشروع التقريبي"><input type="date" className="inp" value={form.deliveryDate} onChange={e => set('deliveryDate', e.target.value)} /></Field>
      </div>
      <button onClick={save} disabled={saving}
        className="w-full mt-5 py-2.5 rounded-lg bg-[var(--color-brand-500)] text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'حفظ ✅'}
      </button>
    </Modal>
  )
}

// ---------------------------------------------------------------------------
// Edit project modal
// ---------------------------------------------------------------------------
function EditModal({ project, onClose, onSaved }: { project: Project; onClose: () => void; onSaved: () => void }) {
  const workTypes = useProjectWorkTypes()
  const [form, setForm] = useState({
    name: project.name, rep: project.rep || '', phone: project.phone || '',
    location: project.location || '', workType: project.workType || 'طاقة شمسية',
    refPerson: project.refPerson || '', priority: project.priority, deliveryDate: project.deliveryDate || '',
  })
  const [mapPoint, setMapPoint] = useState<{ lat: number; lng: number } | null>(
    project.mapLatitude != null && project.mapLongitude != null ? { lat: project.mapLatitude, lng: project.mapLongitude } : null,
  )
  const [locationUrl, setLocationUrl] = useState(project.locationUrl || '')
  const [saving, setSaving] = useState(false)
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const save = async () => {
    if (!form.name.trim()) { alert('اسم المؤسسة مطلوب'); return }
    setSaving(true)
    try {
      await request(`/projects/${project.id}`, {
        method: 'PUT',
        body: JSON.stringify({ ...form, mapLatitude: mapPoint?.lat, mapLongitude: mapPoint?.lng, locationUrl: locationUrl || null }),
      })
      onSaved()
    } catch (e) { alert((e as Error).message); setSaving(false) }
  }

  return (
    <Modal onClose={onClose} title={`تعديل: ${project.name}`} wide>
      <div className="grid sm:grid-cols-2 gap-3">
        <Field label="المؤسسة *"><input className="inp" value={form.name} onChange={e => set('name', e.target.value)} /></Field>
        <Field label="الممثل"><input className="inp" value={form.rep} onChange={e => set('rep', e.target.value)} /></Field>
        <Field label="الهاتف"><input className="inp" value={form.phone} onChange={e => set('phone', e.target.value)} /></Field>
        <div className="sm:col-span-2">
          {/* نفس آلية الموردين: عنوان + رابط + خريطة، والرابط يغني عن الخريطة */}
          <LocationFields
            addressLabel="عنوان الموقع"
            address={form.location}
            onAddressChange={(v) => set('location', v)}
            point={mapPoint}
            onPointChange={setMapPoint}
            locationUrl={locationUrl}
            onLocationUrlChange={setLocationUrl}
            resolveUrl={(u) => request<{ lat: number; lng: number }>(`/geo/resolve-map-link?url=${encodeURIComponent(u)}`)}
          />
        </div>
        <Field label="نوع العمل">
          <select className="inp" value={form.workType} onChange={e => set('workType', e.target.value)}>
            {workTypes.map(w => <option key={w} value={w}>{w}</option>)}
          </select>
        </Field>
        <Field label="الطرف الوسيط"><input className="inp" value={form.refPerson} onChange={e => set('refPerson', e.target.value)} /></Field>
        <Field label="الأولوية">
          <select className="inp" value={form.priority} onChange={e => set('priority', e.target.value)}>
            <option value="عادي">عادي</option>
            <option value="عاجل جداً">عاجل جداً 🔥</option>
          </select>
        </Field>
        <Field label="تاريخ انتهاء المشروع التقريبي"><input type="date" className="inp" value={form.deliveryDate} onChange={e => set('deliveryDate', e.target.value)} /></Field>
      </div>
      <button onClick={save} disabled={saving}
        className="w-full mt-5 py-2.5 rounded-lg bg-[var(--color-brand-500)] text-white font-bold disabled:opacity-50">
        {saving ? 'جارٍ الحفظ...' : 'حفظ التعديلات ✅'}
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
  const toContract = cur.includes('سعر') && nextStage.includes('عقد')
  const toExec = cur.includes('عقد') && nextStage.includes('تنفيذ')
  const toDone = cur.includes('تنفيذ') && nextStage.includes('مكتمل')

  const candidates = useProjectCandidates()
  // تحديد المسؤول ومنفّذ الكشف يصير هنا (عند الترحيل لمرحلة الكشف)، مو بفورمة
  // إضافة/تعديل المشروع — لأن هذي اللحظة الي ينعرف بيها مين راح يطلع كشف.
  const [responsibleEmployeeId, setResponsibleEmployeeId] = useState(project.responsibleEmployeeId || '')
  const [surveyorEmployeeId, setSurveyorEmployeeId] = useState(project.surveyorEmployeeId || '')
  const [staff, setStaff] = useState(project.staff || '')
  const [location, setLocation] = useState(project.location || '')
  const [time, setTime] = useState('')
  const [task, setTask] = useState(project.task || '')
  const [price, setPrice] = useState(project.price || '')
  const [survey, setSurvey] = useState<string[]>(SURVEY.map((s, i) => project.survey?.[i] || s.opts[0]))
  const [reason, setReason] = useState('')
  const [otherReason, setOtherReason] = useState('')
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

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
      if (toKashf) {
        payload.responsibleEmployeeId = responsibleEmployeeId
        payload.surveyorEmployeeId = surveyorEmployeeId
      }
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
          <Field label="المسؤول عن المشروع (مهندس فقط)">
            <select className="inp" value={responsibleEmployeeId} onChange={e => setResponsibleEmployeeId(e.target.value)}>
              <option value="">-- اختر المهندس --</option>
              <EmployeeOptions candidates={candidates.filter(c => c.isEngineer)} />
            </select>
          </Field>
          <Field label="منفّذ الكشف">
            <select className="inp" value={surveyorEmployeeId} onChange={e => setSurveyorEmployeeId(e.target.value)}>
              <option value="">-- اختر الموظف --</option>
              <EmployeeOptions candidates={candidates} />
            </select>
          </Field>
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

      {/* سعر → عقد : تنبيه لازم يترفع العقد قبل ما يترحّل للتنفيذ */}
      {!isReject && toContract && (
        <div className="bg-purple-50 text-purple-700 rounded-lg p-3 text-center text-sm font-bold">
          📄 بعد الترحيل، ارفع ملف العقد (PDF) من زر "العقد" بالبطاقة — ما تكدر تترحّل لمرحلة التنفيذ قبل رفعه.
        </div>
      )}

      {/* عقد → تنفيذ : الفني المسؤول */}
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

      <button onClick={save} disabled={saving}
        className="w-full mt-3 py-2.5 rounded-lg bg-green-700 text-white font-bold disabled:opacity-50">
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
          <div className="overflow-x-auto">
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
          </div>
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
    if (search.trim()) list = list.filter(r => matches([r.reason], search) ||
      r.projects.some(p => matches([p.name, p.location], search)))
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
        <button onClick={exportCsv} className="px-4 py-2 rounded-full bg-green-700 text-white font-bold text-sm hover:brightness-110">📊 سحب التقرير (Excel)</button>
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
                : { borderColor: 'var(--bd-line)', color: 'var(--t-muted)' }}>
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

// ---------------------------------------------------------------------------
// توجيه المشروع لموظف
// ---------------------------------------------------------------------------
// المدير ما يريد يشتغل على المشروع بنفسه؟ يسلّمه لموظف. القائمة مرتبة من
// المهندسين وصولاً للفنيين. الموظف الموجّه له يشوف المشروع كامل بكل مراحله
// من صفحة "المشاريع الموجّهة لي" ويتحكم بيه — بس بهذا المشروع، مو بكل المشاريع.
function DelegateModal({ project, onClose, onSaved }: {
  project: Project
  onClose: () => void
  onSaved: () => void
}) {
  const candidates = useProjectCandidates()
  const [employeeId, setEmployeeId] = useState(project.delegatedToEmployeeId || '')
  // بحث بالاسم — القائمة فيها كل موظفي الشركة موزّعين على خانات، فبدل
  // ما يدوّر المدير بالقائمة كلها يكتب اسمه ويلكاه.
  const [q, setQ] = useState('')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const save = async (targetId: string) => {
    setSaving(true)
    setError(null)
    try {
      await request(`/projects/${project.id}/delegate`, {
        method: 'PUT',
        body: JSON.stringify({ employeeId: targetId, note: note.trim() }),
      })
      onSaved()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر توجيه المشروع')
    } finally {
      setSaving(false)
    }
  }

  // المرشحون مرتبين أصلاً بالخانات من السيرفر
  const shown = q.trim()
    ? candidates.filter((c) => matches([c.name], q))
    : candidates
  const groups: { label: string; items: ProjectCandidate[] }[] = []
  for (const c of shown) {
    const last = groups[groups.length - 1]
    if (last && last.label === c.groupLabel) last.items.push(c)
    else groups.push({ label: c.groupLabel, items: [c] })
  }

  return (
    <Modal title={`توجيه لموظف: ${project.name}`} onClose={onClose}>
      <div className="space-y-4">
        {project.delegatedToName && (
          <p className="rounded-lg bg-violet-50 p-3 text-sm text-violet-800">
            المشروع موجّه حالياً إلى <b>{project.delegatedToName}</b>
            {project.delegatedAt && ` منذ ${new Date(project.delegatedAt).toLocaleDateString('ar-IQ')}`}
          </p>
        )}

        <Field label="الموظف الموجّه له (كل كوادر الشركة)">
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="🔍 اكتب اسم الموظف للبحث..."
            className="mb-2 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          />
          {q.trim() && shown.length === 0 && (
            <p className="mb-2 text-xs text-red-500">ماكو موظف بهذا الاسم.</p>
          )}
          <select
            value={employeeId}
            onChange={(e) => setEmployeeId(e.target.value)}
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          >
            <option value="">-- اختر موظف --</option>
            {groups.map((g) => (
              <optgroup key={g.label} label={g.label}>
                {g.items.map((c) => (
                  <option key={c.id} value={c.id}>{candidateLabel(c)}</option>
                ))}
              </optgroup>
            ))}
          </select>
        </Field>

        <Field label="ملاحظة للموظف (اختياري)">
          <textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={3}
            placeholder="شنو المطلوب منه بهذا المشروع؟"
            className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          />
        </Field>

        <p className="rounded-lg bg-amber-50 p-3 text-xs text-amber-700">
          الموظف الموجّه له راح يشوف هذا المشروع بكل مراحله وتقاريره من صفحة
          «المشاريع الموجّهة لي»، ويقدر يتابعه ويحرّكه بين المراحل ويسوي عرض سعر —
          بس على هذا المشروع لحاله.
        </p>

        {error && <p className="rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>}

        <div className="flex gap-2">
          <button
            onClick={() => save(employeeId)}
            disabled={saving || !employeeId}
            className="flex-1 rounded-lg bg-violet-600 px-4 py-2 font-medium text-white disabled:opacity-50"
          >
            {saving ? 'جاري الحفظ...' : 'توجيه لموظف'}
          </button>
          {project.delegatedToEmployeeId && (
            <button
              onClick={() => save('')}
              disabled={saving}
              className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-600 disabled:opacity-50"
            >
              إلغاء التوجيه
            </button>
          )}
        </div>
      </div>
    </Modal>
  )
}
