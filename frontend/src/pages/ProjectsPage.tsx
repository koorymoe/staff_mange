import { useDeferredValue, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '../session'
import { matches } from '../utils/search'
import BookingCodeChip from '../components/BookingCodeChip'
import { askConfirm } from '../ui/dialogs'
import { request, type Project, type TransferredBooking, type Stats } from './projects/shared'
import { MainView } from './projects/ProjectCard'
import { AddModal, EditModal, MoveModal, ReportModal, DelegateModal } from './projects/modals'
import { RejectionView } from './projects/RejectionView'

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
  const goto = useNavigate()

  // ═══ خانة «عرض سعر» ═══
  //
  // «ما أريدها إدارة المشاريع — أريد تسويلي خانة هنانه بدل كل
  // المشاريع تخليلي عرض سعر حتى يكدرون يسوون عرض سعر… وهاي الخانة
  // تطلع مال عرض السعر: من أني أنطي صلاحية عرض سعر تطلعله هنانه».
  //
  // ⚠️ **بالصلاحية مو بالدور**: تطلع لصاحب أي صلاحية عرض سعر، ومنو
  // ما عنده ما يشوفها أصلاً — بدل ما يضغط ويرجع «غير مصرح لك».
  //
  // ⚠️ وبرّه شرط `canManage`: صاحب عرض السعر ممكن ما يدير مشاريع
  // إطلاقاً، فلو خلّيناها جوّا شريط الخيارين ما تظهرله أبداً.
  //
  // ⚠️ وما ننسخ شاشة عرض السعر هنا — الخانة **تودّي** لشاشتها
  // `/quotations` الموجودة بحارسها. نسختان تفترقن أول تعديل.
  const canQuote = role === 'ADMIN' || role === 'OWNER'
    || ['quotation_create', 'quotation_edit_own', 'quotation_manage_all', 'quotation_system']
      .some((p) => permissions.includes(p))
  // (ع) 10-14: المالك والمدير يشوفون كل المشاريع — ماكو «موجّهة لي» إلهم
  const isTop = role === 'ADMIN' || role === 'OWNER'
  const [mode, setMode] = useState<'all' | 'delegated'>(isTop ? 'all' : initialMode)
  const delegatedMode = mode === 'delegated'
  const canManage = delegatedMode
    || role === 'ADMIN' || role === 'PROJECT_MANAGER' || permissions.includes('project_management')
  // صلاحية "إضافة مشروع فقط": يشوف واجهة إضافة نظيفة بس — بدون إحصائيات ولا
  // قائمة مشاريع ولا تقارير ولا استمارة كشف ولا ترحيل مراحل.
  const addOnly = !canManage && permissions.includes('project_create_only')

  const [projects, setProjects] = useState<Project[]>([])
  const [stats, setStats] = useState<Stats>({ اتصال: 0, كشف: 0, سعر: 0, عقد: 0, تجهيز: 0, تنفيذ: 0, تسليم: 0, مكتمل: 0, مرفوض: 0 })
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<'main' | 'rejection'>('main')

  // filters
  const [search, setSearch] = useState('')
  const [filterWorkType, setFilterWorkType] = useState('')
  const [filterStage, setFilterStage] = useState('')

  // modals
  // «➕ إضافة مشروع جديد» من قائمة التقني (/projects?add=1) تفتح النافذة مباشرة.
  const [showAdd, setShowAdd] = useState(() => new URLSearchParams(window.location.search).get('add') === '1')
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
    if (!(await askConfirm(`استلام حجز ${b.customer.name} كمشروع جديد؟`))) return
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

  // البحث يتأخر شوية عن الكتابة: الحقل يبقى سريع، والفلترة تلحگه (قياس ١٠-٧: ٨٠٠
  // مشروع چانت تاخذ نص ثانية بكل حرف).
  const deferredSearch = useDeferredValue(search)
  const filtered = useMemo(() => {
    return projects.filter(p => {
      const matchSearch = matches([p.name, p.code, p.phone, p.location, p.rep, p.stage], deferredSearch)
      const matchWork = !filterWorkType || p.workType === filterWorkType
      const matchStage = !filterStage || p.stage.includes(filterStage) || (filterStage === 'سعر' && p.stage.includes('التسعير'))
      return matchSearch && matchWork && matchStage
    })
  }, [projects, deferredSearch, filterWorkType, filterStage])

  const rejectedCount = projects.filter(p => p.stage.includes('مرفوض')).length

  const del = async (id: string) => {
    if (!(await askConfirm('حذف المشروع؟'))) return
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
            ...(isTop ? [] : [{ k: 'delegated' as const, label: '📤 المشاريع الموجّهة لي' }]),
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

      {/* خانة «عرض سعر» — تطلع بالصلاحية وحدها، حتى لمن ما يدير مشاريع */}
      {canQuote && (
        <div className="inline-flex flex-wrap items-center gap-2 rounded-2xl border border-amber-200 bg-amber-50 p-1.5 shadow-sm">
          <button
            onClick={() => goto('/quotations')}
            className="rounded-xl bg-amber-600 px-4 py-2 text-xs font-extrabold text-white shadow-md transition hover:bg-amber-700"
          >
            📄 عرض سعر
          </button>
          <span className="px-2 text-[11px] font-bold text-amber-800">سوّي عرض سعر للزبون</span>
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
        <RejectionView projects={projects} canManage={canManage} onRefresh={refresh} />
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
