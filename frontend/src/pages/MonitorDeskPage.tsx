// ═══ مكتب المراقب ═══
//
// «رتّبلي الشغل كله بواجهة المراقب المدقق».
//
// شغل المراقب چان موزّعاً على **عشر شاشات**، ويشوف المشكلة بشاشة
// ولازم يطلع لشاشة ثانية حتى يتصرّف. والخادم أصلاً مصمَّم عكس هذا —
// تعليق جدول `MonitorReview` يقول حرفياً «الشغل يجي له، مو هو يدور
// عليه». الواجهة بنت عشر أبواب بدل باب واحد.
//
// ⚠️⚠️ **المكتب يضمّ الشاشات — ما ينسخها.** كل قسم هو **الشاشة
// نفسها** بخاصية `embedded` تخفي ترويستها بس. نسخ المحتوى يعني إن
// أول تصحيح يوصل نسخة وينسى الثانية، فالمراقب يشوف بلاغاً بشاشة
// ومحلولاً بالثانية — ويفقد الثقة بالاثنتين.
//
// ⚠️ والنمط موجود بالنظام أصلاً: `MyTasks` تضمّ شاشة فواتير الليدر
// بنفس الطريقة. فما نخترع نمطاً — نمشي على الي انبنى.
//
// ⚠️ والشاشات القديمة **تبقى بمساراتها**: اكو روابط محفوظة بمتصفحات
// الموظفين، وحذف مسار يعني صفحة بيضاء بلا تفسير.

import { lazy, Suspense, useEffect, useState } from 'react'

import { api, type MonitorDeskCounts } from '../api'
import { useSession } from '../session'

const MonitorInboxPage = lazy(() => import('./MonitorInboxPage'))
const AuditIssuesPage = lazy(() => import('./AuditIssuesPage'))
const MonitorCrewBookingsPage = lazy(() => import('./MonitorCrewBookingsPage'))
const QualityFollowUpsPage = lazy(() => import('./QualityFollowUpsPage'))
const LeaderInvoicesListPage = lazy(() => import('./LeaderInvoicesListPage'))
const MonitorGpsPage = lazy(() => import('./MonitorGpsPage'))
const MonitorInventoryPage = lazy(() => import('./MonitorInventoryPage'))
const ComplaintsPage = lazy(() => import('./ComplaintsPage'))

type SectionId = 'inbox' | 'issues' | 'crew' | 'quality' | 'invoices' | 'gps' | 'inventory'

interface Section {
  id: SectionId
  label: string
  icon: string
  /** سطر يقول **شنو القرار** المطلوب بهالقسم — مو وصفاً عاماً. */
  todo: string
}

const SECTIONS: Section[] = [
  { id: 'inbox', label: 'صندوق المراقب', icon: '👁️', todo: 'كل صف: سليم أو عندي ملاحظة' },
  { id: 'issues', label: 'بلاغات التدقيق', icon: '💸', todo: 'دقّق البلاغ على الليدر وسكّره' },
  { id: 'invoices', label: 'الفواتير', icon: '🧾', todo: 'الفواتير الي أرسلها المحاسب لمراجعتك' },
  // ⚠️ الجودة والشكاوى **شاشة وحدة** للمراقب — «ماريد خيارين بلقائمة
  // الجانبية». الاثنان شغل جودة واحد: الشكوى تجي، والمتابعة تحسمها.
  { id: 'quality', label: 'الجودة والشكاوى', icon: '⭐', todo: 'تواصل مع الزبون وسجّل الحكم · والشكاوى الواردة' },
  { id: 'crew', label: 'تنسيق الحجوزات', icon: '📋', todo: 'حجوزات تنتظر تثبيت الإداري' },
  // ⚠️ الجي بي اس **نتائج بس**: ست شاشات تشغيلية انخفت عن المراقب
  // بالقائمة، وصارت تبويباً واحداً هنا. وما انزادت ولا محطة بصندوقه —
  // «ما تخلي وحدة هيج، ماريد أزيد ازدحامها».
  { id: 'gps', label: 'الجي بي اس', icon: '📡', todo: 'اشتراكات قربت تنتهي ومشاكل مفتوحة — عرض بلا تنفيذ' },
  // ⚠️ **متابعة** الجرد، مو شاشة `/inventory` مال أبو الكميات (الي
  // يضيف ويحذف أدوات ويوافق على الطلبات). المراقب يشوف منو جرد ومنو
  // ما جرد وشنو ناقص وليش — ويحاسب، ما ينفّذ.
  { id: 'inventory', label: 'متابعة الجرد', icon: '📦', todo: 'منو جرد ومنو ما جرد، وشنو ناقص وليش' },
]

// عدّاد كل قسم — واحد لكل طابور، من مسار عدّ واحد يجمع الخمسة
// (`GET /monitor-desk/counts`) بدل ما يفتح المراقب كل تبويب ليشوف
// بنفسه. تعريف كل رقم منسوخ من نفس فلترة الشاشة المستقلة بالخادم —
// مو محسوب هنا بمنطق ثانٍ.
// ⚠️ الجي بي اس **مو بالخريطة قصداً**: هذا قسم عرض لا طابور قرار،
// وشارة رقمية عليه توحي إن اكو شي ينتظر ضغطة وماكو. وحطّه على
// `inbox` چان يعرض عدّاد الصندوق على تبويب الجي بي اس — رقم كاذب.
const COUNT_KEY: Partial<Record<SectionId, keyof MonitorDeskCounts>> = {
  inbox: 'inbox', issues: 'issues', invoices: 'invoices', quality: 'quality', crew: 'crew',
}

export default function MonitorDeskPage() {
  const { employee, permissions } = useSession()
  const [active, setActive] = useState<SectionId>('inbox')
  const [counts, setCounts] = useState<MonitorDeskCounts | null>(null)

  useEffect(() => {
    let alive = true
    api.getMonitorDeskCounts()
      .then((c) => { if (alive) setCounts(c) })
      .catch(() => { /* العدّاد زينة — غيابه ما يمنع الشغل */ })
    return () => { alive = false }
  }, [])

  const inboxCount = counts?.inbox ?? null

  // ⚠️⚠️ **دور أو صلاحية — نفس منطق القائمة.** الاعتماد على الصلاحية
  // وحدها يخفي القسم عن مراقب حقيقي: صفوف الصلاحيات تنعطى بالطلب،
  // والمراقب الي دوره MONITOR يوصل الشاشة أصلاً بدونها. (وهذا الي طلع
  // بالفحص: التبويبان انخفوا عن حساب مراقب سليم.)
  const role = employee?.actualRole ?? employee?.role
  const isMon = role === 'MONITOR' || role === 'ADMIN' || role === 'OWNER'
  const canQuality = isMon || role === 'QUALITY_ENGINEER' || permissions.includes('quality_control')
  // ⚠️⚠️ **المكتب ما يصير يفوّت حارساً.** شاشة «تنسيق الحجوزات»
  // المستقلة محروسة بـ`RequirePermission("crew_management")`، فلو
  // ضمّيناها هنا بشرط الدور، المراقب يشوف بالمكتب شي تمنعه الشاشة —
  // وهذا باب خلفي بنيناه بيدنا وإحنا نرتّب.
  //
  // (والفحص مسكها: التبويب ظهر لمراقب والشاشة المستقلة رفضته.)
  const canCrew = permissions.includes('crew_management')
  // نفس منطق بقية الأقسام: دور المراقب أو صلاحية الجي بي اس صراحةً.
  const canGps = isMon || permissions.includes('gps_system')
  // «متابعة الجرد» صلاحية مستقلة تنمنح فرد-فرد · و`inventory` صلاحية
  // المراقب الافتراضية · والليدر يشوف جروده هو (الخادم يفرّق).
  const canInventory = isMon || permissions.includes('inventory_follow') || permissions.includes('inventory')
  // ⚠️ القسم الي ما عنده صلاحيته **ما يظهر أصلاً**: قسم فاضي بلا
  // تفسير يخلّي المراقب يظن النظام مكسوراً، وقسم يفتح ويرفض أسوأ.
  const shown = SECTIONS.filter((s) =>
    (s.id !== 'quality' || canQuality) && (s.id !== 'crew' || canCrew) && (s.id !== 'gps' || canGps)
    && (s.id !== 'inventory' || canInventory))

  const cur = shown.find((s) => s.id === active) ?? shown[0]

  return (
    <div dir="rtl" className="space-y-4">
      {/* ═══ الترويسة ═══
          ⚠️ بنفس هوية شاشات المحاسب (التدرّج) — الشاشتان شغل واحد،
          واختلاف الهوية بينهما يخلّي المراقب يحس إنه انتقل لنظام ثاني. */}
      <div
        className="relative overflow-hidden rounded-2xl p-6 shadow-md"
        style={{ background: 'linear-gradient(135deg, #1a3a5c 0%, #24507e 55%, #2f6ba8 100%)' }}
      >
        <div className="relative flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-black text-white">🗂️ مكتب المراقب</h1>
            <p className="mt-1 max-w-2xl text-sm text-blue-100">
              كل شغلك بمكان واحد — الصندوق والبلاغات والفواتير والجودة والتنسيق.
              كل قسم فيه <b className="text-white">قرار</b>، وتسويه من هنا بلا ما تطلع.
            </p>
          </div>
          {inboxCount !== null && inboxCount > 0 && (
            <div className="rounded-xl bg-amber-400/20 px-4 py-2 text-center ring-1 ring-amber-200/40 backdrop-blur">
              <p className="text-2xl font-black leading-none text-amber-100">{inboxCount}</p>
              <p className="mt-1 text-[11px] text-amber-50">بانتظار قرارك</p>
            </div>
          )}
        </div>
      </div>

      {/* ═══ الأقسام ═══
          ⚠️ تبويبات مو صفحة طويلة: خمسة طوابير فوق بعض تعني تمريراً
          بلا نهاية، والمراقب يضيع بأيّها هو. */}
      <div className="flex flex-wrap gap-2">
        {shown.map((s) => (
          <button
            key={s.id}
            onClick={() => setActive(s.id)}
            className={`rounded-xl border px-4 py-2 text-sm font-bold transition-all ${
              active === s.id
                ? 'border-transparent bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow'
                : 'border-slate-200 bg-white text-slate-600 hover:border-brand-300 hover:text-brand-700'
            }`}
          >
            {s.icon} {s.label}
            {counts && COUNT_KEY[s.id] && counts[COUNT_KEY[s.id]!] > 0 && (
              <span className={`mr-1.5 rounded-full px-1.5 py-0.5 text-[10.5px] tabular-nums ${
                active === s.id ? 'bg-white/25' : 'bg-amber-100 text-amber-800'}`}>{counts[COUNT_KEY[s.id]!]}</span>
            )}
          </button>
        ))}
      </div>

      {/* ⚠️ سطر يقول **شنو القرار** المطلوب بالقسم — عنوان بلا مهمة
          يخلّي المراقب يفتح ويتفرّج. */}
      {cur && (
        <p className="rounded-xl border border-slate-200 bg-slate-50 px-4 py-2 text-[12.5px] font-bold text-slate-600">
          {cur.icon} {cur.todo}
        </p>
      )}

      <Suspense fallback={<p className="py-16 text-center text-slate-400">جاري التحميل…</p>}>
        {active === 'inbox' && <MonitorInboxPage embedded />}
        {active === 'issues' && <AuditIssuesPage embedded />}
        {active === 'invoices' && <LeaderInvoicesListPage embedded />}
        {active === 'quality' && canQuality && (
          <div className="space-y-4">
            <QualityFollowUpsPage embedded />
            {/* ⚠️ الشكاوى تحتها بنفس التبويب — انشالت من القائمة
                الجانبية للمراقب، فلو ما انعرضت هنا يخسر وصوله كلياً. */}
            <ComplaintsPage />
          </div>
        )}
        {active === 'crew' && canCrew && <MonitorCrewBookingsPage embedded />}
        {active === 'gps' && canGps && <MonitorGpsPage embedded />}
        {active === 'inventory' && canInventory && <MonitorInventoryPage embedded />}
      </Suspense>
    </div>
  )
}
