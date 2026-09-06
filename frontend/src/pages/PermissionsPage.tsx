import { useEffect, useState } from 'react'
import { api, type Employee, type MissingRoleDefault, type Permission } from '../api'
import { roleLabels } from '../session'
import { matches } from '../utils/search'

export default function PermissionsPage() {
  // ═══ فحص الصلاحيات الناقصة ═══
  //
  // ⚠️ خريطة صلاحيات الأدوار بالكود **اقتراح مو قاعدة منفَّذة**:
  // إنشاء موظف ما يمنحه ولا صلاحية منها، والتطبيق بضغطة يدوية.
  // فموظف بدور المحاسب ممكن يكون بلا صلاحية «المالية» — والقائمة
  // تعرضله شاشات وكل طلباته تنرفض، وهو ما يعرف ليش.
  //
  // ماكو مكان بالنظام يبيّن هذا، فانضاف هنا.
  const [audit, setAudit] = useState<MissingRoleDefault[] | null>(null)
  const [auditBusy, setAuditBusy] = useState(false)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [employeePerms, setEmployeePerms] = useState<string[]>([])
  const [roleDefaults, setRoleDefaults] = useState<Record<string, string[]>>({})
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [applyingDefaults, setApplyingDefaults] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  useEffect(() => {
    Promise.all([api.getEmployees(), api.getPermissions(), api.getRoleDefaults()])
      .then(([emps, perms, defaults]) => {
        setEmployees(emps)
        setPermissions(perms)
        setRoleDefaults(defaults)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    // Guard-clause reset when selection is cleared, not part of the fetch itself.
    if (!selectedEmployeeId) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setEmployeePerms([])
      return
    }
    api.getEmployeePermissions(selectedEmployeeId)
      .then((perms) => setEmployeePerms(perms.map((p) => p.name)))
      .catch(() => setEmployeePerms([]))
  }, [selectedEmployeeId])

  const selectedEmployee = employees.find((e) => e.id === selectedEmployeeId)
  const defaults = selectedEmployee ? (roleDefaults[selectedEmployee.role] || []) : []

  // درجات عروض الأسعار الثلاث متبادلة (راديو) — تفعيل وحدة يلغي البقية تلقائياً
  // حتى ما ينمنح موظف أكثر من مستوى وحدة بالغلط.
  const quotationTiers = ['quotation_create', 'quotation_edit_own', 'quotation_manage_all']

  const togglePermission = (permName: string) => {
    setEmployeePerms((prev) => {
      if (prev.includes(permName)) return prev.filter((p) => p !== permName)
      if (quotationTiers.includes(permName)) {
        return [...prev.filter((p) => !quotationTiers.includes(p)), permName]
      }
      return [...prev, permName]
    })
    setSuccessMsg(null)
  }

  const handleSave = async () => {
    if (!selectedEmployeeId) return
    setSaving(true)
    setSuccessMsg(null)
    setError(null)
    try {
      const permIds = permissions
        .filter((p) => employeePerms.includes(p.name))
        .map((p) => p.id)
      await api.setEmployeePermissions(selectedEmployeeId, permIds)
      setSuccessMsg('تم حفظ الصلاحيات بنجاح')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ أثناء الحفظ')
    } finally {
      setSaving(false)
    }
  }

  const handleApplyDefaults = async () => {
    if (!selectedEmployeeId) return
    setApplyingDefaults(true)
    setSuccessMsg(null)
    setError(null)
    try {
      const result = await api.applyDefaultPermissions(selectedEmployeeId)
      setEmployeePerms(result.map(p => p.name))
      setSuccessMsg('تم تطبيق الصلاحيات الافتراضية بنجاح')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setApplyingDefaults(false)
    }
  }

  const filteredEmployees = employees.filter((emp) =>
    matches([emp.name, emp.jobTitle, emp.position], search),
  )

  const permissionGroups = [
    {
      title: 'إدارة الموظفين',
      // ⚠️ صلاحيات مستقلة لكل شغلة: staff_management كانت تفتح معاها
      // الإحصائيات وبرامج التدريب وتوجيه المهام — كلهن على مفتاح واحد.
      // انفصلن حتى تنطي الشغلة الي تقصدها بس.
      perms: ['staff_management', 'staff_stats', 'extra_tasks_assign', 'training_manage', 'edit_employee_profile', 'kpi_management', 'kpi_criteria_management', 'inventory', 'tool_requests_approve'],
    },
    {
      title: 'إدارة العمل',
      perms: ['sales_booking', 'manage_customers', 'view_bookings', 'coordinator', 'bookings_archive', 'manage_services', 'mission_tracking'],
    },
    {
      title: 'إدارة المشاريع والخدمات',
      perms: ['project_management', 'project_create_only', 'gps_system', 'content_technician'],
    },
    {
      // ثلاث درجات متدرجة (وحدة وحدة، مو مع بعض) — تفعيل درجة يلغي الثانيتين
      // تلقائياً (منطق راديو)، تطبيقه بـtogglePermission أدناه.
      title: 'عروض الأسعار (اختر درجة وحدة فقط)',
      perms: ['quotation_create', 'quotation_edit_own', 'quotation_manage_all'],
    },
    {
      title: 'المركبات',
      perms: ['vehicle_management'],
    },
    {
      title: 'المشتريات والمواد',
      perms: ['procurement', 'procurement_personal', 'procurement_customer', 'suppliers_management'],
    },
    {
      title: 'الحسابات',
      perms: ['finance', 'expenses', 'complaints'],
    },
    {
      title: 'المراقبة والتدقيق',
      perms: ['monitoring', 'auditing', 'quality_control', 'crew_management'],
    },
    {
      title: 'سلة الليدر',
      perms: ['leader_basket'],
    },
    {
      // الوحدات: كل وحدة إدارية بالقائمة الجانبية (تحت فاصل "── الوحدات ──")
      // إلها صلاحية خاصة بيها تتحكم بظهورها كاملة — منفصلة عن صلاحيات
      // الصفحات الداخلية، حتى منح موظف صلاحية داخلية وحدة ما يفتحله وحدات
      // ثانية بالغلط. حالياً وحدة التقنيين هيه الوحدة الوحيدة المفعّلة بهذا
      // النمط — بقية الوحدات القديمة تعتمد صلاحياتها الحالية لحد ما تنراجع.
      // منح صلاحية وحدة يفتح الوحدة كاملة بكل صفحاتها للموظف
      title: 'الوحدات (منح الوحدة كاملة للموظف)',
      perms: [
        'unit_service', 'unit_technicians', 'unit_design', 'unit_pr', 'unit_quality',
        'unit_monitoring', 'unit_procurement', 'unit_finance', 'unit_hr', 'unit_projects',
      ],
    },
    {
      title: 'سياسة الخصوصية',
      perms: ['privacy_policy_manage'],
    },
  ]

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">إدارة الصلاحيات</h2>
      <p className="mt-1 text-slate-500">تحديد صلاحيات الوصول لكل موظف في الأنظمة المختلفة.</p>

      {/* فحص: منو ناقصه صلاحيات دوره */}
      <div className="mt-4 rounded-xl border p-4"
        style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <p className="text-sm font-bold" style={{ color: 'var(--t-title)' }}>
              🔎 فحص الصلاحيات الناقصة
            </p>
            <p className="mt-0.5 text-xs" style={{ color: 'var(--t-muted)' }}>
              يقارن صلاحيات كل موظف نشط بصلاحيات دوره الافتراضية. الي ناقصه
              شي، الشاشات تطلعله بالقائمة وطلباته تنرفض بلا ما يعرف ليش.
            </p>
          </div>
          <button
            disabled={auditBusy}
            onClick={() => {
              setAuditBusy(true)
              api.auditRoleDefaults().then(setAudit)
                .catch(() => setAudit(null))
                .finally(() => setAuditBusy(false))
            }}
            className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-5 py-2.5 text-sm font-bold text-white disabled:opacity-50">
            {auditBusy ? 'جاري الفحص...' : 'افحص الآن'}
          </button>
        </div>

        {audit && audit.length === 0 && (
          <p className="mt-3 rounded-lg bg-emerald-50 px-3 py-2 text-sm font-bold text-emerald-700">
            ✅ كل موظف نشط عنده صلاحيات دوره كاملة.
          </p>
        )}

        {audit && audit.length > 0 && (
          <div className="mt-3 space-y-2">
            <p className="text-xs font-bold text-red-700">
              {audit.length} موظف ناقصه صلاحيات دوره:
            </p>
            {audit.map((r) => (
              <div key={r.employeeId}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg border px-3 py-2"
                style={{ backgroundColor: 'var(--sf-sunken)', borderColor: 'var(--bd-line)' }}>
                <div className="min-w-0">
                  <p className="text-sm font-bold" style={{ color: 'var(--t-title)' }}>
                    {r.employeeName}{' '}
                    <span className="text-xs font-normal" style={{ color: 'var(--t-muted)' }}>
                      ({roleLabels[r.role] || r.role})
                    </span>
                  </p>
                  <p className="text-xs" style={{ color: 'var(--t-body)' }}>
                    ناقصه {r.missing.length}: {r.missingLabels.join('، ')}
                  </p>
                  {r.notSeeded && r.notSeeded.length > 0 && (
                    /* ⚠️ فرق مهم: هاي مو «الموظف ناقصه» — هاي صلاحية
                       مو موجودة بقاعدة البيانات أصلاً. */
                    <p className="mt-0.5 text-[11px] font-bold text-amber-700">
                      ⚠️ منهن {r.notSeeded.length} مو معرّفة بقاعدة البيانات أصلاً
                    </p>
                  )}
                </div>
                <button
                  onClick={() => {
                    api.applyDefaultPermissions(r.employeeId)
                      .then(() => api.auditRoleDefaults().then(setAudit))
                      .catch((e) => alert(e instanceof Error ? e.message : 'تعذر التطبيق'))
                  }}
                  className="shrink-0 rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white">
                  طبّق الافتراضي
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-4 rounded-lg bg-red-50 p-4 text-red-600">{error}</p>
      )}

      {!loading && (
        <div className="mt-6 space-y-6">
          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <label className="mb-2 block text-sm font-medium text-slate-600">
              البحث عن موظف
            </label>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="اكتب اسم الموظف..."
              className="mb-4 w-full max-w-md rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
            />

            <label className="mb-1 block text-sm font-medium text-slate-600">اختر الموظف</label>
            <select
              value={selectedEmployeeId}
              onChange={(e) => {
                setSelectedEmployeeId(e.target.value)
                setSuccessMsg(null)
                setError(null)
              }}
              className="w-full max-w-md rounded-lg border border-slate-300 px-3 py-2.5 outline-none focus:border-brand-500"
            >
              <option value="">-- اختر موظف --</option>
              {filteredEmployees.map((emp) => (
                <option key={emp.id} value={emp.id}>
                  {emp.name} — {roleLabels[emp.role] || emp.role}
                </option>
              ))}
            </select>
          </div>

          {selectedEmployeeId && selectedEmployee && (
            <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
                <div>
                  <h3 className="text-lg font-bold text-brand-800">
                    صلاحيات: {selectedEmployee.name}
                  </h3>
                  <p className="text-sm text-slate-400">
                    الدور: {roleLabels[selectedEmployee.role] || selectedEmployee.role}
                  </p>
                </div>
                {defaults.length > 0 && (
                  <button
                    onClick={handleApplyDefaults}
                    disabled={applyingDefaults}
                    className="rounded-lg border border-brand-300 bg-brand-50 px-4 py-2 text-sm font-medium text-brand-700 transition-all hover:bg-brand-100 disabled:opacity-50"
                  >
                    {applyingDefaults ? 'جاري التطبيق...' : 'تطبيق الصلاحيات الافتراضية للدور'}
                  </button>
                )}
              </div>

              {[
                ...permissionGroups,
                {
                  title: 'أخرى',
                  // شبكة أمان: أي صلاحية جديدة تُضاف بالباك اند وننسى نحطها بمجموعة
                  // فوق تظهر هنا تلقائياً، بدل ما تختفي بصمت من الصفحة.
                  perms: permissions
                    .map(p => p.name)
                    .filter(name => !permissionGroups.some(g => g.perms.includes(name))),
                },
              ].map((group) => {
                const groupPerms = permissions.filter(p => group.perms.includes(p.name))
                if (!groupPerms.length) return null
                return (
                  <div key={group.title} className="mb-6">
                    <h4 className="mb-3 text-sm font-bold text-slate-500 border-b border-slate-100 pb-2">
                      {group.title}
                    </h4>
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      {groupPerms.map((perm) => {
                        const isActive = employeePerms.includes(perm.name)
                        const isDefault = defaults.includes(perm.name)
                        return (
                          <div
                            key={perm.id}
                            className={`flex items-center justify-between rounded-lg border px-5 py-4 transition-colors ${
                              isActive
                                ? 'border-brand-500 bg-brand-50'
                                : 'border-slate-200 bg-white'
                            }`}
                          >
                            <div className="flex flex-col">
                              <span className={`text-sm font-medium ${isActive ? 'text-brand-800' : 'text-slate-500'}`}>
                                {perm.label}
                              </span>
                              {isDefault && !isActive && (
                                <span className="text-[10px] text-amber-500 mt-0.5">افتراضية لهذا الدور</span>
                              )}
                            </div>
                            <button
                              type="button"
                              role="switch"
                              aria-checked={isActive}
                              onClick={() => togglePermission(perm.name)}
                              className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full transition-colors ${
                                isActive ? 'bg-brand-600' : 'bg-slate-300'
                              }`}
                            >
                              <span
                                className={`pointer-events-none inline-block h-5 w-5 translate-y-0.5 transform rounded-full bg-white shadow ring-0 transition-transform ${
                                  isActive ? '-translate-x-5' : '-translate-x-0.5'
                                }`}
                              />
                            </button>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                )
              })}

              {successMsg && (
                <p className="mt-4 rounded-lg bg-emerald-50 p-3 text-sm font-medium text-emerald-700">
                  {successMsg}
                </p>
              )}

              <div className="mt-6">
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-10 py-2.5 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
                >
                  {saving ? 'جاري الحفظ...' : 'حفظ الصلاحيات'}
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
