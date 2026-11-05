import { useEffect, useState } from 'react'
import { api, apiRequest } from '../../api'
import { useSession, hasRole } from '../../session'
import { roleLabelShort } from '../../roleLabels'
import { askConfirm, askInput } from '../../ui/dialogs'
import { swallow } from '../../utils/swallow'

// ⚠️ چانت عندها طريقة اتصال خاصة (من أصلها كنظام منفصل) — هسه الموحدة.
export const request = apiRequest

// ---------------------------------------------------------------------------
// Types & constants
// ---------------------------------------------------------------------------
export interface Project {
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
  // (ع 10-16) رقابة المشرف: شوكت تواصل ويا الزبون
  customerContactedAt?: string | null
  customerContactNote?: string | null
}

// حجز محول من إدارة الكوادر (الحجوزات الكبيرة تعتبر مشاريع)
export interface TransferredBooking {
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

export interface Stats {
  اتصال: number; كشف: number; سعر: number; عقد: number; الإعلام?: number; تجهيز?: number; تنفيذ: number; تسليم?: number; مكتمل: number; مرفوض: number
}

export const STAGES = [
  '1. اتصال بالزبون',
  '2. مرحلة الكشف',
  '3. عرض السعر',
  '🔁 إعادة التسعير',
  '4. العقد',
  // قرار (ع) 10-05: قبل التنفيذ المشروع يتحوّل للإعلام حتى يصوّرون الشغل.
  '📸 الإعلام',
  // (ع) 10-31: تجهيز المنتجات ورا العقد وقبل التنفيذ، والتسليم والاستلام ورا التنفيذ
  '5. تجهيز المنتجات',
  '6. البدء بالتنفيذ',
  '7. التسليم والاستلام',
  '✅ مكتمل',
  '❌ مرفوض',
]

// المرحلة الجاية لزر «ترحيل ⮕» — بالأسماء مو بالأرقام: الأرقام انزاحت لمّا
// انضافت مراحل (١٠-١٧) فوگف الترحيل من «٥. البدء بالتنفيذ».
// قرار (ع) 10-06: الإعلام زر بخطوات المشروع، مو مرحلة توقفه — العقد يترحّل
// للتنفيذ مباشرة. وإعادة التسعير فرع ينختار يدوياً، والرفض زر مستقل.
export const STAGE_SIDE_BRANCHES = ['🔁 إعادة التسعير', '📸 الإعلام', '❌ مرفوض']

export const STAGE_NEXT_OVERRIDE: Record<string, string> = {
  '4. العقد': '5. تجهيز المنتجات',
  '🔁 إعادة التسعير': '4. العقد',
  '📸 الإعلام': '5. تجهيز المنتجات',
}

export function nextProjectStage(stage: string): string | null {
  if (STAGE_NEXT_OVERRIDE[stage]) return STAGE_NEXT_OVERRIDE[stage]
  const i = STAGES.indexOf(stage)
  if (i < 0) return null
  for (let j = i + 1; j < STAGES.length; j++) {
    if (!STAGE_SIDE_BRANCHES.includes(STAGES[j])) return STAGES[j]
  }
  return null
}

// أنواع الأعمال كانت قائمة ثابتة بالكود — صارت تنجلب من إعدادات وحدة إدارة
// المشاريع (/project-work-types) حتى يقدر المدير يضيف/يحذف نوع عمل براحته
// بدون تعديل كود. القائمة القديمة تبقى fallback ريثما يوصل جواب السيرفر.
export const FALLBACK_WORK_TYPES = ['طاقة شمسية', 'كاميرات', 'بيت ذكي', 'شبكات', 'إنذار حريق', 'أقفال وحاكيات', 'ستلايت', 'منظومة صوت', 'أخرى']

export interface ProjectCandidate {
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
export function candidateLabel(c: ProjectCandidate): string {
  const bits = [roleLabelShort(c.role)]
  if (c.isLeader) bits.push('تيم ليدر')
  if (c.isTrainee) bits.push('⚠ متدرب')
  return `${c.name} — ${bits.join('، ')}`
}

// مرشحو المشروع — "المسؤول عن المشروع" حصراً المهندسون (اللي عندهم مهارات
// تصميم/تخطيط/تنفيذ أو دورهم مهندس)، أما "منفّذ الكشف" فأي موظف بس معروض
// بالتسلسل: مهندسين ← تقنيين ← ليدريه ← فنيين ← إداريين ← مصممين.
export function useProjectCandidates(): ProjectCandidate[] {
  const [staff, setStaff] = useState<ProjectCandidate[]>([])
  useEffect(() => {
    request<ProjectCandidate[]>('/project-candidates').then(setStaff).catch(swallow)
  }, [])
  return staff
}

export function useProjectWorkTypes(): string[] {
  const [types, setTypes] = useState<string[]>(FALLBACK_WORK_TYPES)
  useEffect(() => {
    request<{ id: string; name: string }[]>('/project-work-types')
      .then((rows) => { if (rows.length > 0) setTypes(rows.map((r) => r.name)) })
      .catch(swallow)
  }, [])
  return types
}

export const STAGE_CARDS = [
  { key: 'اتصال', label: 'اتصال', icon: '📞', color: 'bg-[var(--color-brand-500)]' },
  { key: 'كشف', label: 'كشف', icon: '🔍', color: 'bg-green-700' },
  { key: 'سعر', label: 'سعر', icon: '💰', color: 'bg-amber-500' },
  { key: 'عقد', label: 'عقد', icon: '📄', color: 'bg-purple-600' },
  // قرار (ع) 10-06: الإعلام بين العقد والتنفيذ.
  { key: 'الإعلام', label: 'إعلام', icon: '📸', color: 'bg-pink-600' },
  // (ع) 10-31: التجهيز والتسليم يطلعن بالخانات هم
  { key: 'تجهيز', label: 'تجهيز المنتجات', icon: '📦', color: 'bg-orange-500' },
  { key: 'تنفيذ', label: 'تنفيذ', icon: '🛠️', color: 'bg-red-600' },
  { key: 'تسليم', label: 'تسليم واستلام', icon: '🤝', color: 'bg-teal-600' },
  { key: 'مكتمل', label: 'مكتمل', icon: '✅', color: 'bg-blue-600' },
  { key: 'مرفوض', label: 'مرفوض', icon: '❌', color: 'bg-gray-500' },
] as const

// استمارة الكشف الفني (17 سؤال)
export const SURVEY = [
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
export const REJECTION_REASONS = [
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

export const CATEGORIES = [
  { key: 'all', label: 'الكل', color: 'var(--t-muted)' },
  { key: 'مالي', label: 'مالي', color: 'var(--t-success)' },
  { key: 'عميل', label: 'عميل', color: 'var(--t-warning)' },
  { key: 'تقني', label: 'تقني', color: 'var(--t-danger)' },
  { key: 'إداري', label: 'إداري', color: 'var(--t-cyan)' },
  { key: 'ظروف', label: 'ظروف', color: 'var(--t-muted)' },
  { key: 'أخرى', label: 'أخرى', color: 'var(--t-body)' },
]

export function categoryColor(cat: string): string {
  return CATEGORIES.find(c => c.key === cat)?.color || '#dc2626'
}

export function parseRejection(task: string | null): { reason: string; notes: string; category: string } {
  let reasonText = (task || 'سبب غير مُحدد').replace(/رفض المشروع\s*[:-]?\s*/i, '')
  const parts = reasonText.split('|')
  reasonText = parts[0].trim() || 'سبب غير مُحدد'
  const notes = parts[1] ? parts[1].replace('ملاحظات:', '').trim() : ''
  const match = REJECTION_REASONS.find(r => reasonText.includes(r.value) || r.value.includes(reasonText))
  return { reason: reasonText, notes, category: match?.category || 'أخرى' }
}

// ---------------------------------------------------------------------------
// Main dashboard view
// ---------------------------------------------------------------------------
export const PAGE = 30

// ---------------------------------------------------------------------------
// 📸 إعلام — تحويل المشروع للإعلام (قرار (ع) 10-06: زر بخطوات المشروع)
// ---------------------------------------------------------------------------
// 👁️ (دازاي 10-31) علامة «المراقب متابع» على بطاقة المشروع — من إجراءات المراقب على المتأخر
export type MonitorWatch = Record<string, { action: 'FOLLOW_UP' | 'ESCALATE' | 'VIOLATION'; at: string; dueAt?: string }>

export let monitorWatchCache: Promise<MonitorWatch> | null = null

export function useMonitorWatch(): MonitorWatch {
  const [m, setM] = useState<MonitorWatch>({})
  useEffect(() => {
    if (!monitorWatchCache) monitorWatchCache = request<MonitorWatch>('/projects/monitor-watch').catch(() => ({}))
    let alive = true
    void monitorWatchCache.then((x) => { if (alive) setM(x) })
    return () => { alive = false }
  }, [])
  return m
}

export const WATCH_LABEL = { FOLLOW_UP: '⏰ حاطّ مهلة', ESCALATE: '📌 صعّده للإدارة', VIOLATION: '⚖️ سجّل مخالفة' } as const

let mediaOpenCache: Promise<Set<string>> | null = null
// نقل لـ«الإعلام» يغيّر القائمة — الكاش ينصفّر من النافذة (الاستيراد ما ينكتب عليه)
export function resetMediaOpenCache() { mediaOpenCache = null }

export function useMediaOpen(): Set<string> {
  const [set, setSet] = useState<Set<string>>(new Set())
  useEffect(() => {
    if (!mediaOpenCache) mediaOpenCache = request<string[]>('/projects/media/open').then((ids) => new Set(ids)).catch(() => new Set<string>())
    let alive = true
    void mediaOpenCache.then((s) => { if (alive) setSet(s) })
    return () => { alive = false }
  }, [])
  return set
}

export interface OpenBrief { engineerId: string | null; startAt: string | null; expectedEndAt: string | null; duration: string | null; notes: string | null; status: string }

export const localDT = (iso: string | null) => (iso ? new Date(new Date(iso).getTime() - new Date().getTimezoneOffset() * 60000).toISOString().slice(0, 16) : '')

// رفع صور كشف المشروع: منفّذ الكشف أو المسؤول أو الموجّه له، أو إدارة
// المشاريع. الخادم يعيد نفس الفحص — هذا بس حتى ما يطلع زر يفشل.
export function useCanUploadProjectSurvey(project: Project) {
  const { employee, permissions } = useSession()
  if (!employee) return false
  if (['ADMIN', 'PROJECT_MANAGER', 'HR_COORDINATOR'].includes(employee.role)) return true
  if (permissions.includes('project_management')) return true
  return [project.surveyorEmployeeId, project.responsibleEmployeeId, project.delegatedToEmployeeId].includes(employee.id)
}

// إرجاع مشروع مرفوض لمرحلته قبل الرفض — (ع) 10-19
export async function restoreRejected(p: { id: string; name: string }, done: () => void) {
  const reason = (await askInput(`ترجّع «${p.name}» للمشاريع؟ يرجع لنفس المرحلة الي چان بيها قبل الرفض.\nاكتب السبب (مثلاً: الزبون رجع واتصل):`))
  if (reason === null) return
  if (reason.trim().length < 3) { alert('اكتب السبب'); return }
  try {
    const r = await api.restoreProject(p.id, reason.trim())
    alert(`رجع المشروع ✅ — مرحلته هسه: ${r.stage}`)
    done()
  } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر الإرجاع') }
}

// حذف نهائي للمرفوض — (ع) 10-19: «مع طلب السبب». نفس حارس الخادم (ADMIN/PROJECT_MANAGER).
export function useCanPurge() {
  const { employee } = useSession()
  return employee?.role === 'ADMIN' || hasRole(employee, 'PROJECT_MANAGER')
}

export async function purgeRejected(p: { id: string; name: string }, done: () => void) {
  const reason = (await askInput(`🗑️ حذف «${p.name}» نهائياً؟ ينشال من كل الشاشات وما يرجع.\nاكتب سبب الحذف:`))
  if (reason === null) return
  if (reason.trim().length < 3) { alert('اكتب سبب الحذف'); return }
  if (!(await askConfirm(`متأكد؟ «${p.name}» ينحذف نهائياً.`))) return
  try {
    await api.purgeProject(p.id, reason.trim())
    alert('انحذف المشروع ✅')
    done()
  } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر الحذف') }
}
