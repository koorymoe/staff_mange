import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type EmployeeRole, type Permission } from '../api'
import { navItems, isNavVisible, type NavItem } from '../components/navTree'

// ═══ دليل الأدوار والصلاحيات ═══
//
// «أريدك تسويلي خطة لكل دور موجود عندي بالنظام — كل دور شنو صلاحياته،
// شنو يكدر يسوي وشنو ما يكدر».
//
// ⚠️ **متولّدة من الكود مو مكتوبة بالإيد.** الوثيقة المكتوبة بالإيد
// تصير كذب بعد أول تعديل: تنضاف شاشة أو تتغيّر صلاحية وما أحد يتذكر
// يحدّث الورقة. هنا نحسبها من `navItems` بـ`isNavVisible()` — **نفس
// الدالة** الي تبني القائمة الحقيقية وتبني «شوف بعين الموظف».
//
// يعني: لو الوثيقة تگول إن المحاسب يشوف شاشة، فهو **فعلاً** يشوفها —
// لأن نفس السطر الي حسب هذا هو الي يعرض القائمة له.
//
// ⚠️ وتعرض «شنو **ما** يكدر» بعد: الوثيقة الي تعدّد المسموح بس تخلي
// القارئ يخمّن الباقي — وأخطر سؤال بالصلاحيات هو «هل هذا الدور يوصل
// للفلوس؟» ولازم ينجاوب صراحةً.

const ROLE_LABELS: Record<string, string> = {
  OWNER: 'المالك',
  ADMIN: 'مدير النظام',
  HR_COORDINATOR: 'إداري الكوادر / الحجوزات',
  FINANCE: 'المحاسب',
  MONITOR: 'المراقب',
  TECHNICIAN: 'فني',
  TECHNICAL: 'تقني',
  PROJECT_MANAGER: 'مدير مشاريع',
  QUALITY_ENGINEER: 'مهندس جودة',
  ENGINEER: 'مهندس',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  GPS_ADMIN: 'مسؤول جي بي اس',
  DESIGNER: 'مصمم',
  SALES: 'موظف مبيعات',
  SERVICE_MANAGER: 'مسؤول خدمة',
}

// ⚠️ الشاشات الحسّاسة تتأشّر صراحةً: «هل هذا الدور يوصل للفلوس أو
// للصلاحيات أو للحذف؟» أهم سؤال، ولازم ما ينلزم القارئ يستنتجه.
const SENSITIVE: { match: (to: string) => boolean; label: string }[] = [
  { match: (t) => /finance|expenses|revolving|leader-invoices|daily-audit|gps-install-costs/.test(t), label: '💰 فلوس' },
  { match: (t) => /permissions|employees/.test(t), label: '🔑 صلاحيات وحسابات' },
  { match: (t) => /owner-backups|command-code/.test(t), label: '🔐 مالك' },
  { match: (t) => /archive|delete/.test(t), label: '🗑️ حذف وأرشفة' },
]

function flatten(items: NavItem[]): NavItem[] {
  const out: NavItem[] = []
  for (const it of items) {
    if (it.divider) continue
    if (it.to) out.push(it)
    if (it.children) out.push(...flatten(it.children))
  }
  return out
}

export default function RolesGuidePage() {
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [role, setRole] = useState<string>('HR_COORDINATOR')
  const [withGrants, setWithGrants] = useState(false)

  useEffect(() => { api.getPermissions().then(setPermissions).catch(() => setPermissions([])) }, [])

  const all = useMemo(() => flatten(navItems), [])

  // موظف وهمي بالدور المختار — نمرّره لنفس دالة القائمة الحقيقية.
  // ⚠️ isLeader = false: الليدر حالة خاصة تنعرض بدورها المستقل، ودمجها
  // بـ«فني» يخلي الوثيقة تگول إن كل فني يشوف شاشات الليدر.
  const ctxFor = (r: string, grants: string[]) => ({
    // ⚠️ نقلّد التطبيع الي تسويه الجلسة الحقيقية (Layout سطر ٦٣٢):
    // المالك ينمشي بالنظام كـrole='ADMIN' ودوره الحقيقي بـactualRole.
    //
    // انمسك بالفحص: بدون هذا التطبيع الوثيقة كانت تگول إن **المالك
    // يشوف ٤ شاشات بس** — لأن أغلب العناصر مشروطة بـroles:['ADMIN']
    // وسلسلة 'OWNER' ما تطابقها. يعني الوثيقة تكذب على أهم دور.
    employee: {
      role: (r === 'OWNER' ? 'ADMIN' : r) as EmployeeRole,
      actualRole: r as EmployeeRole,
      isLeader: false,
    } as unknown as Employee,
    permissions: grants,
    gpsServiceId: null,
  })

  const grants = withGrants ? permissions.map((p) => p.name) : []
  const ctx = ctxFor(role, grants)

  const visible = all.filter((i) => isNavVisible(i, ctx))
  const hidden = all.filter((i) => !isNavVisible(i, ctx))

  const sensitiveOf = (to: string) => SENSITIVE.filter((s) => s.match(to)).map((s) => s.label)

  // ═══ تصدير Markdown ═══
  // نفس الحساب بالضبط — الملف والشاشة ما يفترقون.
  const exportMd = () => {
    const lines: string[] = ['# دليل الأدوار والصلاحيات', '',
      '> متولّد من الكود (`navItems` + `isNavVisible`) — نفس المصدر الي',
      '> يبني القائمة الحقيقية. أي تعديل بالصلاحيات ينعكس هنا تلقائياً.', '']
    for (const r of Object.keys(ROLE_LABELS)) {
      const c = ctxFor(r, [])
      const vis = all.filter((i) => isNavVisible(i, c))
      const hid = all.filter((i) => !isNavVisible(i, c))
      lines.push(`## ${ROLE_LABELS[r]} (\`${r}\`)`, '',
        `**يشوف ${vis.length} شاشة** بلا أي صلاحية إضافية:`, '')
      for (const i of vis) {
        const tags = sensitiveOf(i.to || '')
        lines.push(`- ${i.label}${tags.length ? ` — ${tags.join(' ')}` : ''}`)
      }
      lines.push('', `**ما يشوف (${hid.length}):** ` +
        hid.slice(0, 25).map((i) => i.label).join(' · ') + (hid.length > 25 ? ' …' : ''), '')
    }
    const blob = new Blob([lines.join('\n')], { type: 'text/markdown;charset=utf-8' })
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = 'دليل-الأدوار.md'
    a.click()
    URL.revokeObjectURL(a.href)
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">📋 دليل الأدوار والصلاحيات</h2>
          <p className="mt-1 text-sm text-slate-500">
            كل دور: شنو يشوف، وشنو <b>ما</b> يشوف. متولّد من الكود — مو مكتوب بالإيد.
          </p>
        </div>
        <button onClick={exportMd} className="rounded-xl bg-[#0f2040] px-4 py-2.5 text-sm font-bold text-white">
          ⬇️ نزّل الدليل (Markdown)
        </button>
      </div>

      {/* ⚠️ نگولها صراحة: الوثيقة تحسب بنفس دالة القائمة الحقيقية.
          بدون هذا التوضيح، القارئ ما يعرف إذا يثق بيها. */}
      <p className="rounded-xl border border-sky-200 bg-sky-50 p-4 text-xs leading-6 text-sky-900">
        هذي الوثيقة <b>تنحسب من نفس الكود</b> الي يبني القائمة لكل موظف (<code>isNavVisible</code>) —
        نفس مصدر «شوف بعين الموظف». يعني لو تگول إن دور يشوف شاشة، فهو <b>فعلاً</b> يشوفها.
        الوثيقة المكتوبة بالإيد تصير كذب بعد أول تعديل.
      </p>

      <div className="flex flex-wrap items-center gap-2">
        <select
          value={role}
          onChange={(e) => setRole(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold"
        >
          {Object.entries(ROLE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <label className="flex items-center gap-1.5 text-xs font-bold text-slate-600">
          <input type="checkbox" checked={withGrants} onChange={(e) => setWithGrants(e.target.checked)} />
          وريني لو انطيته <b>كل</b> الصلاحيات
        </label>
        <span className="mr-auto text-xs text-slate-500">
          يشوف <b className="text-emerald-700">{visible.length}</b> · ما يشوف <b className="text-slate-700">{hidden.length}</b>
        </span>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-2xl border border-emerald-200 bg-white p-4">
          <h3 className="mb-2 font-bold text-emerald-800">✅ يكدر يوصل ({visible.length})</h3>
          <div className="space-y-1">
            {visible.map((i) => {
              const tags = sensitiveOf(i.to || '')
              return (
                <p key={i.to} className="flex flex-wrap items-center gap-1.5 text-xs text-slate-700">
                  <span>{i.label}</span>
                  {tags.map((t) => (
                    <span key={t} className="rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-bold text-amber-800">{t}</span>
                  ))}
                </p>
              )
            })}
            {visible.length === 0 && <p className="text-xs text-slate-400">ما يشوف ولا شاشة.</p>}
          </div>
        </div>

        {/* ⚠️ «ما يكدر» تنعرض بنفس الوزن: أخطر سؤال بالصلاحيات هو
            «هل هذا الدور يوصل للفلوس؟» — ولازم ينجاوب صراحةً. */}
        <div className="rounded-2xl border border-slate-200 bg-white p-4">
          <h3 className="mb-2 font-bold text-slate-700">🚫 ما يكدر يوصل ({hidden.length})</h3>
          <div className="max-h-[420px] space-y-1 overflow-y-auto">
            {hidden.map((i) => {
              const tags = sensitiveOf(i.to || '')
              return (
                <p key={i.to} className="flex flex-wrap items-center gap-1.5 text-xs text-slate-500">
                  <span>{i.label}</span>
                  {tags.map((t) => (
                    <span key={t} className="rounded-full bg-slate-100 px-1.5 py-0.5 text-[10px] font-bold text-slate-500">{t}</span>
                  ))}
                </p>
              )
            })}
            {hidden.length === 0 && <p className="text-xs text-slate-400">يشوف كلشي.</p>}
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-white bg-white p-4">
        <h3 className="mb-2 font-bold text-[#0f2040]">🔑 الصلاحيات المتاحة ({permissions.length})</h3>
        <p className="mb-2 text-xs text-slate-500">
          الصلاحية تُمنح لموظف من شاشة الصلاحيات، وتفتحله شاشات زيادة على دوره.
        </p>
        <div className="flex flex-wrap gap-1.5">
          {permissions.map((p) => (
            <span key={p.id} className="rounded-lg bg-slate-100 px-2 py-1 text-[11px] text-slate-700" title={p.name}>
              {p.label || p.name}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}
