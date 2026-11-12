import { useState } from 'react'
import ProjectPaymentsModal from '../../components/ProjectPaymentsModal'
import { apiRequest } from '../../api'
import { useNavigate } from 'react-router-dom'
import ProjectDelayBanner from '../../components/ProjectDelayBanner'
import { askInput } from '../../ui/dialogs'
import { type Project, type Stats, STAGES, nextProjectStage, useProjectWorkTypes, STAGE_CARDS, PAGE, useMonitorWatch, WATCH_LABEL, useMediaOpen, restoreRejected, useCanPurge, purgeRejected } from './shared'
import { ContractModal, MediaTransferModal } from './modals'
import { Info } from './ui'
import LinkQuotationModal from './LinkQuotationModal'

export function MainView(props: {
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
  // عدد البطاقات المعروضة — يرجع لأول دفعة كل ما تغيّر البحث أو الفلتر.
  const sig = `${props.search}|${props.filterStage}|${props.filterWorkType}`
  const [shown, setShown] = useState({ sig, n: PAGE })
  const shownCount = shown.sig === sig ? shown.n : PAGE
  const { projects, totalCount, stats, canManage } = props
  const maxStat = Math.max(1, ...STAGE_CARDS.map(c => stats[c.key as keyof Stats] ?? 0))
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
      <div className="grid grid-cols-3 sm:grid-cols-6 lg:grid-cols-11 gap-2">
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
            <span className="text-lg font-bold">{stats[c.key as keyof Stats] ?? 0}</span>
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
                  style={{ width: `${((stats[c.key as keyof Stats] ?? 0) / maxStat) * 100}%` }} />
              </div>
              <span className="text-xs font-bold w-8">{stats[c.key as keyof Stats] ?? 0}</span>
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

      {/* Project cards — نرسم أول دفعة بس (قياس ١٠-٧: ٨٠٠ بطاقة = ٣٠ ألف عنصر و٥-٧ ثواني
          فتح). البيانات كلها بالذاكرة، فالعدّادات والفلاتر والبحث تشمل الكل. */}
      <div className="space-y-4">
        {projects.length === 0 && <p className="text-center text-gray-400 py-12 text-lg">لا توجد مشاريع</p>}
        {projects.slice(0, shownCount).map(p => (
          <ProjectCard key={p.id} p={p} canManage={canManage}
            onEdit={props.onEdit} onMove={props.onMove} onReport={props.onReport} onDelete={props.onDelete}
            onRefresh={props.onRefresh} onDelegate={props.onDelegate} />
        ))}
        {projects.length > shownCount && (
          <button type="button" onClick={() => setShown({ sig, n: shownCount + PAGE })}
            className="w-full rounded-xl border border-gray-300 bg-white py-3 text-sm font-bold text-gray-700 hover:bg-gray-50">
            اعرض {Math.min(PAGE, projects.length - shownCount)} ثانية · معروض {shownCount} من {projects.length}
          </button>
        )}
      </div>
    </>
  )
}

export function ProjectCard({ p, canManage, onEdit, onMove, onReport, onDelete, onRefresh, onDelegate }: {
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
  const canPurge = useCanPurge()
  const isCompleted = p.stage.includes('مكتمل')
  const isContractStage = p.stage.includes('عقد')
  const stageIdx = STAGES.indexOf(p.stage)
  const nextStage = nextProjectStage(p.stage)
  const [showContract, setShowContract] = useState(false)
  const [showMedia, setShowMedia] = useState(false)
  const [showPay, setShowPay] = useState(false)
  const [showLinkQ, setShowLinkQ] = useState(false)
  const inMedia = useMediaOpen().has(p.id)
  const watch = useMonitorWatch()[p.id]

  const handleAdvance = () => {
    if (isContractStage && !p.hasContract) {
      alert('ارفع ملف العقد (PDF) الأول قبل الترحيل لمرحلة تجهيز المنتجات.')
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
        style={{ background: isRejected ? '#6b7280' : isCompleted ? '#16a34a' : '#f59e0b' }}>
        {isCompleted ? '✅ منتهي' : `📅 انتهاء تقريبي: ${p.deliveryDate || '---'}`}
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
      {!isRejected && !isCompleted && <ProjectDelayBanner projectId={p.id} />}

      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-2 mt-4">
        <Info icon="📞" value={p.phone} />
        <Info icon="👤" value={p.rep} />
        <Info icon="🤝" value={p.refPerson || 'لا يوجد'} />
        <Info icon="🛠️" value={p.workType} />
        <Info icon="📍" value={p.location} />
        {/* مكان المرحلة القديم: منو رحّل الحجز أو أضاف المشروع */}
        <Info icon="🧑‍💼" value={p.createdByName ? `أضافه: ${p.createdByName}` : 'أضافه: غير محدد'} />
        {p.delegatedToName && <Info icon="📤" value={`موجّه إلى: ${p.delegatedToName}`} />}
        {p.customerContactedAt && (
          <Info icon="📞" value={`تواصل ويا الزبون: ${new Date(p.customerContactedAt.endsWith('Z') ? p.customerContactedAt : p.customerContactedAt + 'Z').toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })}${p.customerContactNote ? ` — ${p.customerContactNote}` : ''}`} />
        )}
      </div>

      <div className="mt-4 pt-3 border-t flex flex-wrap items-center justify-between gap-2">
        {canManage && (
          <div className="flex flex-wrap gap-2">
            {/* (ع 10-16) المراقب يشوف شوكت تواصل المشرف — قبل عرض السعر */}
            {!p.customerContactedAt && !isRejected && !isCompleted && (
              <button onClick={async () => {
                const note = (await askInput('شنو حجيت ويا الزبون؟ (جملة قصيرة)'))
                if (!note || note.trim().length < 5) { if (note !== null) alert('اكتب جملة قصيرة على الأقل'); return }
                try { await apiRequest(`/projects/${p.id}/customer-contacted`, { method: 'POST', body: JSON.stringify({ note: note.trim() }) }); onRefresh() }
                catch (e) { alert(e instanceof Error ? e.message : 'تعذر التسجيل') }
              }}
                className="text-sm px-3 py-1.5 rounded-lg bg-emerald-600 text-white font-bold hover:bg-emerald-700">📞 تواصلت ويا الزبون</button>
            )}
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
                  projectId: p.id,
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
              🧾 عرض السعر
            </button>
          )}
          {/* بلاغ #33: عرض انسوّى قبل ما الزبون يصير مشروع — يرتبط هنا */}
          {canManage && (p.stage.includes('كشف') || p.stage.includes('سعر')) && (
            <button onClick={() => setShowLinkQ(true)}
              className="text-sm px-3 py-1.5 rounded-lg border border-cyan-600 text-cyan-700 font-medium hover:bg-cyan-50">
              🔗 ربط عرض موجود
            </button>
          )}
          {canManage && (isContractStage || p.hasContract) && (
            <button onClick={() => setShowContract(true)}
              className="text-sm px-3 py-1.5 rounded-lg bg-purple-600 text-white font-medium hover:brightness-110">
              📄 العقد{p.hasContract ? (p.hasSignedContract ? ' (مرفوع وموقّع)' : ' (مرفوع)') : ''}
            </button>
          )}
          {/* 💰 الدفعات: فلوس المشروع تنسجّل وتدخل بالإيرادات (قرار (ع) 10-06) */}
          {canManage && !isRejected && (
            <button onClick={() => setShowPay(true)}
              className="text-sm px-3 py-1.5 rounded-lg bg-emerald-700 text-white font-bold hover:brightness-110">💰 الدفعات</button>
          )}
          {/* 📸 إعلام: تحويل المشروع للإعلام والعلاقات العامة حتى يصوّرون الشغل */}
          {canManage && !isRejected && !isCompleted && stageIdx >= 3 && (
            <button onClick={() => setShowMedia(true)}
              className={`text-sm px-3 py-1.5 rounded-lg font-bold hover:brightness-110 ${inMedia ? 'border border-pink-400 bg-pink-50 text-pink-700' : 'bg-pink-600 text-white'}`}>
              {inMedia ? '📸 بالإعلام ✓' : '📸 إعلام'}
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
          {canManage && isRejected && (
            <button onClick={() => void restoreRejected(p, onRefresh)}
              className="text-sm px-3 py-1.5 rounded-lg bg-emerald-600 text-white font-bold hover:brightness-110">↩️ إرجاع للمشاريع</button>
          )}
          {canPurge && isRejected && (
            <button onClick={() => void purgeRejected(p, onRefresh)}
              className="text-sm px-3 py-1.5 rounded-lg bg-red-600 text-white font-bold hover:brightness-110">🗑️ حذف نهائي</button>
          )}
          {isCompleted && (
            <span className="text-sm px-3 py-1.5 rounded-full bg-green-700 text-white font-bold flex items-center gap-1">✅ تم التسليم والانتهاء</span>
          )}
        </div>
      </div>

      {canManage && (p.stage.includes('تنفيذ') || p.stage.includes('التسليم')) && (
        <p className="mt-2 text-center text-xs text-emerald-800">🤖 المشروع ما يكتمل بدون الفلوس — سجّل الدفعات من «💰 الدفعات» أول بأول.</p>
      )}
      {watch && (
        <p className="mt-2 rounded-lg bg-indigo-50 py-1.5 text-center text-xs font-bold text-indigo-900">
          👁️ المراقب متابع — {WATCH_LABEL[watch.action]}{watch.dueAt ? ` لحد ${new Date(watch.dueAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })}` : ''}
        </p>
      )}
      <div className="mt-3 text-center text-xs text-gray-500 bg-gray-50 rounded-lg py-2">🕐 تاريخ الإضافة: {addedDate}</div>

      {showContract && (
        <ContractModal project={p} onClose={() => setShowContract(false)} onSaved={() => { setShowContract(false); onRefresh() }} />
      )}
      {showLinkQ && <LinkQuotationModal project={p} onClose={() => setShowLinkQ(false)} onLinked={() => { setShowLinkQ(false); onRefresh() }} />}
      {showPay && <ProjectPaymentsModal projectId={p.id} title={p.name} onClose={() => setShowPay(false)} />}
      {showMedia && (
        <MediaTransferModal project={p} onClose={() => setShowMedia(false)} onSaved={() => { setShowMedia(false); onRefresh() }} />
      )}
    </div>
  )
}
