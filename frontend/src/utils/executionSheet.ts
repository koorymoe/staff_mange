// منطق «تقرير تنفيذ المشروع» — مفصول عن الواجهة (react-refresh).
import type { Booking } from '../api'

export interface SheetCrewRow { employeeId: string; name: string; from: string; to: string; days: string; rating: string }
export interface DecorItem { item: string; qty: string; unit: string }
export interface DecorMaterial { name: string; qty: string }
export interface ExecutionSheetData {
  v: 1
  // (ع 10-16) تقرير الديكور: نفس النموذج، بس بنود الشغل الإنشائي بدل الأجهزة
  kind?: 'DECOR'
  decor?: { items: DecorItem[]; materials: DecorMaterial[] }
  header: { date: string; customer: string; phone: string; location: string; workType: string; customerCode: string; responsible: string }
  time: { startH: string; startM: string; startNote: string; endH: string; endM: string; endNote: string }
  data: {
    workDays: string; landmark: string; startDate: string; endDate: string; systemType: string; devices: string
    wiringMeters: string; extras: string; obstacles: string; projectStatus: string; programming: string; codeGiven: string
  }
  crew: SheetCrewRow[]
  checks: Record<CheckKey, boolean>
  suggestion: string
  amountReceived: string
}

export type CheckKey = 'photosDevices' | 'photosWiring' | 'photosController' | 'systemRunning' | 'explained' | 'cleaned' | 'toolsTaken' | 'handedOver'
export const CHECKS: { key: CheckKey; label: string; hint: string }[] = [
  { key: 'photosDevices', label: 'تصوير الاجهزة', hint: 'لا تنسى التصوير كل الاجهزة' },
  { key: 'photosWiring', label: 'تصوير التسليك', hint: 'لا تنسى تصوير الاسلاك' },
  { key: 'photosController', label: 'تصوير الجهاز المسيطر', hint: 'صور الجهاز المسيطرة و الشاشاة ان وجدت' },
  { key: 'systemRunning', label: 'تشغيل المنظومة', hint: 'شغل المنظومة قبل الخروج' },
  { key: 'explained', label: 'شرح المنظومة للزبون', hint: 'اشرح بالتفصيل عن المنظومة' },
  { key: 'cleaned', label: 'تم تنظيف الموقع', hint: 'نظف مكانك بعد اكمال العمل لان انت نظيف' },
  { key: 'toolsTaken', label: 'اخذ جميع المعدات و عدم نسيان شيء', hint: 'معداتك و ادواتك لك لا تنساهم' },
]
// إكمال شغل الديكور — قبل/بعد، النظافة، العدة، والتسليم للشعبة الهندسية
export const DECOR_CHECKS: { key: CheckKey; label: string; hint: string }[] = [
  { key: 'photosDevices', label: 'تصوير الموقع قبل الشغل', hint: 'صوّر المكان قبل ما تبدي' },
  { key: 'photosWiring', label: 'تصوير الشغل بعد ما خلص', hint: 'صوّر الشغل المنجز من كل الجهات' },
  { key: 'cleaned', label: 'تم تنظيف الموقع', hint: 'نظّف مكانك بعد إكمال العمل' },
  { key: 'toolsTaken', label: 'اخذ جميع المعدات و عدم نسيان شيء', hint: 'معداتك و ادواتك لك لا تنساهم' },
  { key: 'handedOver', label: 'تسليم الموقع للشعبة الهندسية', hint: 'إذا الشغل تحضير للشعبة — بلّغهم إنه جاهز' },
]
export const PROJECT_STATUS = ['تم اكمال العمل 100%', 'انجاز جزئي', 'متوقف'] as const
export const PROGRAMMING = ['كملت الاعدادات', 'ما كملت', 'ما تحتاج برمجة']
export const RATINGS = ['ممتاز', 'جيد جداً', 'جيد', 'مقبول', 'ضعيف']
export const HOURS = Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0'))
export const MINUTES = Array.from({ length: 12 }, (_, i) => String(i * 5).padStart(2, '0'))
export const DAYS = Array.from({ length: 30 }, (_, i) => String(i + 1))

export const DECOR_UNITS = ['متر', 'متر مربع', 'قطعة', 'كغم', 'يوم']
// بنود «بيانات التنفيذ» بنفس ترتيب الشيت وتلميحاته
export type DataKey = keyof ExecutionSheetData['data']
export const DATA_ROWS: { key: DataKey; label: string; hint: string; kind: 'text' | 'num' | 'date' | 'days' | 'system' | 'status' | 'prog' }[] = [
  { key: 'workDays', label: 'عدد ايام العمل', hint: 'عدد الايام العمل بالمضبوط', kind: 'days' },
  { key: 'landmark', label: 'موقع المشروع اقرب نقطة دالة', hint: '', kind: 'text' },
  { key: 'startDate', label: 'تاريخ البدء', hint: 'ادخل الزمن الحقيقي', kind: 'date' },
  { key: 'endDate', label: 'تاريخ الانتهاء', hint: 'يا ساعة كملت', kind: 'date' },
  { key: 'systemType', label: 'نوع المنظومة', hint: 'شوكت خلص البطل', kind: 'system' },
  { key: 'devices', label: 'عدد الاجهزة النهائية', hint: 'كل الاجهزة الي استخدمتها', kind: 'num' },
  { key: 'wiringMeters', label: 'مسافات التسليك', hint: 'الرقم التقريبي بالمتر', kind: 'num' },
  { key: 'extras', label: 'اضافات اخرى', hint: 'اي اضافة من الزبون اذكرها', kind: 'text' },
  { key: 'obstacles', label: 'معرقلات العمل', hint: 'شنو الي اخرك و شنو تحتاج', kind: 'text' },
  { key: 'projectStatus', label: 'حالة المشروع', hint: 'تم اكمال العمل 100%', kind: 'status' },
  { key: 'programming', label: 'برمجة الاجهزة', hint: 'كملت الاعدادات ؟', kind: 'prog' },
  { key: 'codeGiven', label: 'اعطاء الكود للزبون', hint: 'اذكر الكود الخاص بالزبون و يجب ان يكون رقم الهاتف و am', kind: 'text' },
]

// بيانات تنفيذ الديكور — نفس المفاتيح (حتى ماتركس والمراقب يقرون نفس الأماكن)، بمعنى إنشائي
export const DECOR_DATA_ROWS: typeof DATA_ROWS = [
  { key: 'workDays', label: 'عدد ايام العمل', hint: 'عدد الايام العمل بالمضبوط', kind: 'days' },
  { key: 'systemType', label: 'المهنة / نوع الشغل', hint: 'حدادة، نجارة، صبغ، جبس…', kind: 'text' },
  { key: 'landmark', label: 'موقع العمل اقرب نقطة دالة', hint: '', kind: 'text' },
  { key: 'startDate', label: 'تاريخ البدء', hint: '', kind: 'date' },
  { key: 'endDate', label: 'تاريخ الانتهاء', hint: '', kind: 'date' },
  { key: 'extras', label: 'الشغل المنجز بالتفصيل', hint: 'شنو سويت بالضبط (مثلاً قاعدة حديد للداتا شو)', kind: 'text' },
  { key: 'obstacles', label: 'معرقلات العمل', hint: 'إذا أكو شي أخّرك', kind: 'text' },
  { key: 'projectStatus', label: 'حالة الشغل', hint: '', kind: 'status' },
]
export const rowsFor = (s: ExecutionSheetData) => (s.kind === 'DECOR' ? DECOR_DATA_ROWS : DATA_ROWS)
export const checksFor = (s: ExecutionSheetData) => (s.kind === 'DECOR' ? DECOR_CHECKS : CHECKS)

export const today = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export function emptySheet(b: Booking): ExecutionSheetData {
  const crew: SheetCrewRow[] = []
  const seen = new Set<string>()
  const push = (id?: string | null, name?: string | null) => {
    if (!id || seen.has(id) || crew.length >= 7) return
    seen.add(id); crew.push({ employeeId: id, name: name || '', from: today(), to: today(), days: '1', rating: '' })
  }
  push(b.leaderId, b.leaderName)
  for (const a of b.assignments || []) push(a.employee?.id, a.employee?.name)
  const svc = (b.services && b.services.length ? b.services.map((s) => s.name).join(' + ') : b.service?.name) || ''
  return {
    v: 1,
    header: {
      date: today(), customer: b.customer?.name || '', phone: b.customer?.phone || '',
      location: b.address || '', workType: svc, customerCode: String(b.customer?.customerCode ?? ''),
      responsible: b.leaderName || b.projectSupervisor?.name || '',
    },
    time: { startH: '', startM: '', startNote: '', endH: '', endM: '', endNote: '' },
    data: {
      workDays: '1', landmark: '', startDate: today(), endDate: today(), systemType: svc, devices: '',
      wiringMeters: '', extras: '', obstacles: '', projectStatus: '', programming: '', codeGiven: '',
    },
    crew,
    checks: Object.fromEntries([...CHECKS, ...DECOR_CHECKS].map((c) => [c.key, false])) as Record<CheckKey, boolean>,
    suggestion: '', amountReceived: '',
    ...(b.bookingType === 'DECOR' ? { kind: 'DECOR' as const, decor: { items: [{ item: '', qty: '', unit: 'متر' }], materials: [] } } : {}),
  }
}

/** شنو ناقص قبل الإرسال — نفس الخانات الإجبارية بالشيت */
export function sheetMissing(s: ExecutionSheetData): string | null {
  if (!s.time.startH || !s.time.startM) return 'اذكر وقت البدء بالعمل (الساعة والدقيقة)'
  if (!s.time.endH || !s.time.endM) return 'اذكر وقت انتهاء العمل الكلي (الساعة والدقيقة)'
  if (!s.data.projectStatus) return 'اختار حالة المشروع'
  if (s.data.projectStatus !== PROJECT_STATUS[0] && s.data.obstacles.trim().length < 3) return 'اكتب معرقلات العمل — شنو الي أخرك وشنو تحتاج'
  if (s.kind === 'DECOR') {
    if (s.data.extras.trim().length < 3) return 'اكتب الشغل المنجز بالتفصيل'
    if (!s.decor?.items.some((x) => x.item.trim() && x.qty.trim())) return 'اكتب بند واحد على الأقل بالقياسات والكميات'
  }
  return null
}

export const inp = 'w-full rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-sm outline-none focus:border-brand-500'
/** يقرا sheet من الخادم (JSON) بأمان */
export function asSheet(raw: unknown): ExecutionSheetData | null {
  return raw && typeof raw === 'object' && (raw as { v?: number }).v === 1 ? (raw as ExecutionSheetData) : null
}
