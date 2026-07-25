import { useEffect, useState, useCallback } from 'react'
import { api, type Booking, type Employee, type Stats, type VehicleScoreSummary, type TechnicianWashSummary } from '../api'
import { useSession } from '../session'
import { useNavigate } from 'react-router-dom'

type Tab = 'overview' | 'crews' | 'audit' | 'finance' | 'vehicles'

export default function MonitorDashboard() {
  const { employee: currentUser } = useSession()
  const navigate = useNavigate()
  const [tab, setTab] = useState<Tab>('overview')
  const [stats, setStats] = useState<Stats | null>(null)
  const [bookings, setBookings] = useState<Booking[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [vehicleScores, setVehicleScores] = useState<VehicleScoreSummary[]>([])
  const [techWashSummaries, setTechWashSummaries] = useState<TechnicianWashSummary[]>([])

  const load = useCallback(() => {
    const monthAgo = new Date()
    monthAgo.setMonth(monthAgo.getMonth() - 1)
    const since = monthAgo.toISOString().slice(0, 10)
    Promise.all([
      api.getStats().catch(() => null),
      api.getBookings().catch(() => []),
      api.getEmployees().catch(() => []),
      api.getVehicleScoreSummaries(since).catch(() => []),
      api.getTechnicianWashSummaries(since).catch(() => []),
    ]).then(([s, b, e, vs, tw]) => {
      setStats(s as Stats | null)
      setBookings(b as Booking[])
      setEmployees(e as Employee[])
      setVehicleScores(vs as VehicleScoreSummary[])
      setTechWashSummaries(tw as TechnicianWashSummary[])
    }).finally(() => setLoading(false))
  }, [])

  useEffect(load, [load])

  if (!currentUser) return null

  const todayStr = new Date().toDateString()
  const todayBookings = bookings.filter(b => new Date(b.createdAt).toDateString() === todayStr)
  const activeBookings = bookings.filter(b => b.status === 'CONFIRMED' || b.status === 'IN_PROGRESS')
  const completedBookings = bookings.filter(b => b.status === 'COMPLETED')
  const pendingBookings = bookings.filter(b => b.status === 'PENDING')
  const unverifiedBookings = completedBookings.filter(b => !b.amountVerified)
  const onDutyEmployees = employees.filter(e => e.onDuty && e.status === 'ACTIVE')
  const techs = employees.filter(e => e.role === 'TECHNICIAN' || e.role === 'PROJECT_MANAGER')
  const coordinators = employees.filter(e => e.role === 'HR_COORDINATOR')
  const sales = employees.filter(e => e.role === 'SALES')

  const tabs: { key: Tab; label: string; count?: number }[] = [
    { key: 'overview', label: 'نظرة عامة' },
    { key: 'crews', label: 'الكوادر الميدانية', count: activeBookings.length },
    { key: 'audit', label: 'التدقيق', count: unverifiedBookings.length },
    { key: 'finance', label: 'المالية' },
    { key: 'vehicles', label: 'تقييم السيارات والفنيين' },
  ]

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">لوحة المراقبة والتدقيق</h2>
          <p className="mt-1 text-slate-500">متابعة شاملة لجميع العمليات والموظفين</p>
        </div>
        <button onClick={load} className="rounded-lg border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-50">
          تحديث البيانات
        </button>
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <>
          {/* Tabs */}
          <div className="mb-6 flex gap-2 overflow-x-auto rounded-xl bg-white p-1.5 shadow-[0_1px_3px_rgba(0,0,0,0.06)]">
            {tabs.map(t => (
              <button
                key={t.key}
                onClick={() => setTab(t.key)}
                className={`flex items-center gap-2 rounded-lg px-4 py-2.5 text-sm font-bold transition ${
                  tab === t.key ? 'bg-brand-500 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'
                }`}
              >
                {t.label}
                {t.count !== undefined && t.count > 0 && (
                  <span className={`inline-flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-[11px] font-black ${
                    tab === t.key ? 'bg-white/20 text-white' : 'bg-red-100 text-red-600'
                  }`}>{t.count}</span>
                )}
              </button>
            ))}
          </div>

          {/* ═══ Overview Tab ═══ */}
          {tab === 'overview' && (
            <div className="space-y-6">
              {/* Summary Cards */}
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {[
                  { label: 'حجوزات اليوم', value: todayBookings.length, color: 'text-blue-600', bg: 'bg-blue-50' },
                  { label: 'بانتظار التثبيت', value: pendingBookings.length, color: 'text-amber-600', bg: 'bg-amber-50' },
                  { label: 'جاري التنفيذ', value: activeBookings.length, color: 'text-violet-600', bg: 'bg-violet-50' },
                  { label: 'موظفين بالدوام', value: onDutyEmployees.length, color: 'text-emerald-600', bg: 'bg-emerald-50' },
                ].map(c => (
                  <div key={c.label} className={`rounded-2xl ${c.bg} p-4 text-center`}>
                    <p className={`text-3xl font-black ${c.color}`}>{c.value}</p>
                    <p className="mt-1 text-xs font-medium text-slate-500">{c.label}</p>
                  </div>
                ))}
              </div>

              {/* Employee Status Grid */}
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-4 text-lg font-bold text-brand-900">حالة الموظفين</h3>
                <div className="space-y-4">
                  {/* By Role */}
                  {[
                    { label: 'الفنيين / مدراء المشاريع', list: techs },
                    { label: 'الإداريين', list: coordinators },
                    { label: 'المبيعات', list: sales },
                  ].map(group => group.list.length > 0 && (
                    <div key={group.label}>
                      <p className="mb-2 text-sm font-bold text-slate-500">{group.label}</p>
                      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                        {group.list.map(emp => (
                          <div key={emp.id} className="flex items-center justify-between rounded-xl border border-slate-100 px-3 py-2">
                            <div className="flex items-center gap-2">
                              <span className={`h-2.5 w-2.5 rounded-full ${emp.onDuty ? 'bg-emerald-500' : 'bg-slate-300'}`} />
                              <span className="text-sm font-medium text-slate-700">{emp.name}</span>
                            </div>
                            <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                              emp.onDuty ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-400'
                            }`}>
                              {emp.onDuty ? 'بالدوام' : 'خارج الدوام'}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Recent Activity */}
              {stats?.recentBookings && stats.recentBookings.length > 0 && (
                <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                  <h3 className="mb-4 text-lg font-bold text-brand-900">آخر الحجوزات</h3>
                  <div className="space-y-2">
                    {stats.recentBookings.slice(0, 10).map(b => (
                      <div key={b.id} className="flex items-center justify-between rounded-xl border border-slate-100 px-4 py-2.5">
                        <div className="flex items-center gap-3 text-right">
                          <span className="font-mono text-sm font-bold text-brand-600">{b.code}</span>
                          <span className="text-sm text-slate-700">{b.customerName}</span>
                          {b.serviceName && <span className="rounded-lg bg-slate-100 px-2 py-0.5 text-[11px] text-slate-500">{b.serviceName}</span>}
                        </div>
                        <div className="flex items-center gap-2">
                          <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                            b.status === 'PENDING' ? 'bg-amber-100 text-amber-700' :
                            b.status === 'CONFIRMED' ? 'bg-blue-100 text-blue-700' :
                            b.status === 'IN_PROGRESS' ? 'bg-violet-100 text-violet-700' :
                            b.status === 'COMPLETED' ? 'bg-emerald-100 text-emerald-700' :
                            'bg-red-100 text-red-700'
                          }`}>{
                            b.status === 'PENDING' ? 'بانتظار' :
                            b.status === 'CONFIRMED' ? 'مثبت' :
                            b.status === 'IN_PROGRESS' ? 'جاري' :
                            b.status === 'COMPLETED' ? 'مكتمل' : 'ملغي'
                          }</span>
                          <span className="text-[11px] text-slate-400">{new Date(b.createdAt).toLocaleDateString('ar-IQ')}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Stats Summary */}
              {stats && (
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                  {[
                    { label: 'إجمالي الحجوزات', value: stats.totals.totalBookings },
                    { label: 'إجمالي العملاء', value: stats.totals.totalCustomers },
                    { label: 'المنجزة', value: stats.totals.completedBookings },
                    { label: 'إجمالي الإيرادات', value: `${(stats.totals.totalRevenue || 0).toLocaleString('ar-IQ')} د.ع` },
                    { label: 'إيرادات غير مدققة', value: `${(stats.totals.unverifiedRevenue || 0).toLocaleString('ar-IQ')} د.ع` },
                    { label: 'حجوزات عاجلة', value: stats.totals.urgentPending },
                  ].map(s => (
                    <div key={s.label} className="rounded-xl bg-white p-4 text-center shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
                      <p className="text-xl font-black text-brand-800">{s.value}</p>
                      <p className="text-xs text-slate-400">{s.label}</p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* ═══ Crews Tab ═══ */}
          {tab === 'crews' && (
            <div className="space-y-4">
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-1 text-lg font-bold text-brand-900">الكوادر الميدانية الحالية</h3>
                <p className="mb-4 text-sm text-slate-400">الحجوزات المثبتة والجاري تنفيذها مع الكوادر المعينة</p>

                {activeBookings.length === 0 ? (
                  <p className="rounded-xl bg-slate-50 p-8 text-center text-slate-400">لا توجد حجوزات نشطة حالياً</p>
                ) : (
                  <div className="space-y-3">
                    {activeBookings.map(b => (
                      <div key={b.id} className={`rounded-xl border-2 p-4 ${
                        b.status === 'IN_PROGRESS' ? 'border-violet-200 bg-violet-50/30' : 'border-blue-200 bg-blue-50/30'
                      }`}>
                        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                          <div className="flex items-center gap-3">
                            <span className="font-mono text-sm font-bold text-brand-600">{b.code}</span>
                            <span className="font-bold text-brand-800">{b.customer.name}</span>
                            {b.service && <span className="rounded-lg bg-white px-2 py-0.5 text-xs text-slate-600">{b.service.name}</span>}
                          </div>
                          <div className="flex items-center gap-2">
                            <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${
                              b.status === 'IN_PROGRESS' ? 'bg-violet-500 text-white' : 'bg-blue-500 text-white'
                            }`}>{b.status === 'IN_PROGRESS' ? 'جاري التنفيذ' : 'مثبت'}</span>
                            {b.status === 'IN_PROGRESS' && b.responseMinutes != null && (
                              <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${b.responseMinutes > 15 ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'}`}>
                                استجابة الفريق: {b.responseMinutes} د
                              </span>
                            )}
                            {b.status === 'CONFIRMED' && b.materialsReadyAt && (
                              <span className="rounded-full bg-amber-100 px-2.5 py-1 text-xs font-bold text-amber-700">
                                ⏰ بانتظار انطلاق الفريق
                              </span>
                            )}
                            {b.scheduledAt && (
                              <span className="text-xs text-slate-500">
                                {new Date(b.scheduledAt).toLocaleString('ar-IQ', { weekday: 'short', hour: '2-digit', minute: '2-digit' })}
                              </span>
                            )}
                          </div>
                        </div>

                        {/* Crew */}
                        <div className="flex flex-wrap gap-2">
                          {b.assignments.length === 0 ? (
                            <span className="rounded-lg bg-amber-100 px-3 py-1.5 text-xs font-bold text-amber-700">لم يتم تعيين كادر بعد</span>
                          ) : b.assignments.map(a => (
                            <span key={a.id} className="rounded-lg bg-white px-3 py-1.5 text-xs font-medium text-slate-700 shadow-sm">
                              {a.role === 'TECH_1' ? 'فني ١' : a.role === 'TECH_2' ? 'فني ٢' : 'فني ٣'}: <strong>{a.employee.name}</strong>
                              {a.employee.onDuty
                                ? <span className="mr-1 inline-block h-2 w-2 rounded-full bg-emerald-500" />
                                : <span className="mr-1 inline-block h-2 w-2 rounded-full bg-red-400" />
                              }
                            </span>
                          ))}
                          {b.projectSupervisor && (
                            <span className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-medium text-brand-700 shadow-sm">
                              مشرف: <strong>{b.projectSupervisor.name}</strong>
                            </span>
                          )}
                        </div>

                        {/* Details row */}
                        <div className="mt-2 flex flex-wrap gap-3 text-xs text-slate-500">
                          {b.address && <span>العنوان: {b.address}</span>}
                          {b.quotedPrice && <span>التكلفة: {b.quotedPrice.toLocaleString('ar-IQ')} د.ع</span>}
                          {b.assignedVehicle && <span>السيارة: {b.assignedVehicle}</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Technician Performance */}
              {stats?.technicianStats && stats.technicianStats.length > 0 && (
                <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                  <h3 className="mb-4 text-lg font-bold text-brand-900">أداء الفنيين</h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-slate-100 text-right text-xs text-slate-400">
                          <th className="pb-2 pr-3">الفني</th>
                          <th className="pb-2">الحالة</th>
                          <th className="pb-2">مُسند</th>
                          <th className="pb-2">مكتمل</th>
                          <th className="pb-2">الإيرادات</th>
                        </tr>
                      </thead>
                      <tbody>
                        {stats.technicianStats.map(t => (
                          <tr key={t.employeeId} className="border-b border-slate-50">
                            <td className="py-2.5 pr-3 font-medium text-slate-700">{t.name}</td>
                            <td>
                              <span className={`inline-flex h-2 w-2 rounded-full ${t.onDuty ? 'bg-emerald-500' : 'bg-slate-300'}`} />
                            </td>
                            <td className="text-slate-600">{t.totalAssigned}</td>
                            <td className="font-bold text-emerald-600">{t.completed}</td>
                            <td className="text-slate-600">{t.revenueHandled.toLocaleString('ar-IQ')}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ═══ Audit Tab ═══ */}
          {tab === 'audit' && (
            <div className="space-y-4">
              {/* Coordinator Audit */}
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-1 text-lg font-bold text-brand-900">تدقيق عمل الإداريين</h3>
                <p className="mb-4 text-sm text-slate-400">هل الإداري وجّه الكوادر بشكل صحيح؟</p>

                {stats?.coordinatorStats && stats.coordinatorStats.length > 0 ? (
                  <div className="space-y-2">
                    {stats.coordinatorStats.map(c => (
                      <div key={c.employeeId} className="flex items-center justify-between rounded-xl border border-slate-100 px-4 py-3">
                        <span className="font-medium text-slate-700">{c.name}</span>
                        <div className="flex items-center gap-4 text-sm">
                          <span className="text-slate-500">إجمالي التثبيت: <strong className="text-brand-700">{c.totalConfirmed}</strong></span>
                          <span className="text-slate-500">اليوم: <strong className="text-blue-600">{c.today}</strong></span>
                          <span className="text-slate-500">هذا الشهر: <strong className="text-emerald-600">{c.thisMonth}</strong></span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : <p className="text-sm text-slate-400">لا توجد بيانات</p>}
              </div>

              {/* Bookings without crew */}
              {(() => {
                const noCrew = activeBookings.filter(b => b.assignments.length === 0)
                if (noCrew.length === 0) return null
                return (
                  <div className="rounded-2xl border-2 border-amber-200 bg-amber-50 p-5">
                    <h3 className="mb-3 text-lg font-bold text-amber-800">حجوزات مثبتة بدون كادر</h3>
                    <div className="space-y-2">
                      {noCrew.map(b => (
                        <div key={b.id} className="flex items-center justify-between rounded-xl bg-white px-4 py-3">
                          <div className="flex items-center gap-3">
                            <span className="font-mono text-sm font-bold text-brand-600">{b.code}</span>
                            <span className="text-sm text-slate-700">{b.customer.name}</span>
                          </div>
                          <button onClick={() => navigate('/coordinator')} className="rounded-lg bg-amber-500 px-3 py-1 text-xs font-bold text-white hover:bg-amber-600">
                            توجيه كادر
                          </button>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })()}

              {/* Completed but unverified */}
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-1 text-lg font-bold text-brand-900">تدقيق المبالغ</h3>
                <p className="mb-4 text-sm text-slate-400">حجوزات مكتملة بانتظار تدقيق المحاسب</p>

                {unverifiedBookings.length === 0 ? (
                  <div className="rounded-xl bg-emerald-50 p-4 text-center text-sm font-bold text-emerald-600">
                    جميع المبالغ مدققة
                  </div>
                ) : (
                  <div className="space-y-2">
                    {unverifiedBookings.map(b => (
                      <div key={b.id} className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-amber-200 bg-amber-50/50 px-4 py-3">
                        <div className="flex items-center gap-3">
                          <span className="font-mono text-sm font-bold text-brand-600">{b.code}</span>
                          <span className="text-sm text-slate-700">{b.customer.name}</span>
                        </div>
                        <div className="flex items-center gap-3 text-sm">
                          <span className="text-slate-500">المستلم: <strong className="text-brand-700">{(b.amountCollected || 0).toLocaleString('ar-IQ')}</strong></span>
                          <span className="text-slate-500">المقدم: <strong className="text-brand-700">{(b.advancePaid || 0).toLocaleString('ar-IQ')}</strong></span>
                          <span className="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-700">بانتظار التدقيق</span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Technician completion audit */}
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-1 text-lg font-bold text-brand-900">تدقيق إنجاز الفنيين</h3>
                <p className="mb-4 text-sm text-slate-400">الحجوزات المكتملة وتفاصيل الإنجاز</p>

                {completedBookings.length === 0 ? (
                  <p className="text-sm text-slate-400">لا توجد حجوزات مكتملة</p>
                ) : (
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {completedBookings.slice(0, 20).map(b => (
                      <div key={b.id} className="rounded-xl border border-slate-100 px-4 py-3">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                          <div className="flex items-center gap-3">
                            <span className="font-mono text-sm font-bold text-brand-600">{b.code}</span>
                            <span className="text-sm font-medium text-slate-700">{b.customer.name}</span>
                            {b.service && <span className="text-xs text-slate-400">{b.service.name}</span>}
                          </div>
                          <div className="flex items-center gap-2">
                            {b.completedAt && (
                              <span className="text-xs text-slate-400">
                                أُنجز: {new Date(b.completedAt).toLocaleString('ar-IQ')}
                              </span>
                            )}
                            <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                              b.amountVerified ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
                            }`}>{b.amountVerified ? 'مدقق' : 'غير مدقق'}</span>
                          </div>
                        </div>
                        {b.completionNotes && (
                          <p className="mt-1 text-xs text-slate-500">ملاحظات: {b.completionNotes}</p>
                        )}
                        <div className="mt-1 flex flex-wrap gap-2">
                          {b.assignments.map(a => (
                            <span key={a.id} className="text-xs text-slate-500">{a.employee.name}</span>
                          ))}
                          {b.quotedPrice && <span className="text-xs text-slate-400">التكلفة: {b.quotedPrice.toLocaleString('ar-IQ')}</span>}
                          {b.amountCollected && <span className="text-xs text-slate-400">المستلم: {b.amountCollected.toLocaleString('ar-IQ')}</span>}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Sales audit */}
              {stats?.salesStats && stats.salesStats.length > 0 && (
                <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                  <h3 className="mb-1 text-lg font-bold text-brand-900">تدقيق عمل المبيعات</h3>
                  <p className="mb-4 text-sm text-slate-400">أداء موظفي المبيعات وعدد الحجوزات</p>
                  <div className="space-y-2">
                    {stats.salesStats.map(s => (
                      <div key={s.employeeId} className="flex items-center justify-between rounded-xl border border-slate-100 px-4 py-3">
                        <span className="font-medium text-slate-700">{s.name}</span>
                        <div className="flex items-center gap-4 text-sm">
                          <span className="text-slate-500">إجمالي: <strong className="text-brand-700">{s.totalTransferred}</strong></span>
                          <span className="text-slate-500">مثبت: <strong className="text-emerald-600">{s.confirmed}</strong></span>
                          <span className="text-slate-500">اليوم: <strong className="text-blue-600">{s.today}</strong></span>
                          <span className="text-slate-500">الشهر: <strong className="text-violet-600">{s.thisMonth}</strong></span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ═══ Finance Tab ═══ */}
          {tab === 'finance' && (
            <div className="space-y-4">
              {stats && (
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div className="rounded-2xl bg-gradient-to-bl from-emerald-500 to-emerald-600 p-5 text-white">
                    <p className="text-sm font-medium text-emerald-100">إجمالي الإيرادات</p>
                    <p className="mt-1 text-2xl font-black">{(stats.totals.totalRevenue || 0).toLocaleString('ar-IQ')} <span className="text-sm">د.ع</span></p>
                  </div>
                  <div className="rounded-2xl bg-gradient-to-bl from-amber-500 to-amber-600 p-5 text-white">
                    <p className="text-sm font-medium text-amber-100">غير مدققة</p>
                    <p className="mt-1 text-2xl font-black">{(stats.totals.unverifiedRevenue || 0).toLocaleString('ar-IQ')} <span className="text-sm">د.ع</span></p>
                  </div>
                  <div className="rounded-2xl bg-gradient-to-bl from-blue-500 to-blue-600 p-5 text-white">
                    <p className="text-sm font-medium text-blue-100">مدققة</p>
                    <p className="mt-1 text-2xl font-black">{((stats.totals.totalRevenue || 0) - (stats.totals.unverifiedRevenue || 0)).toLocaleString('ar-IQ')} <span className="text-sm">د.ع</span></p>
                  </div>
                </div>
              )}

              {/* Service Breakdown */}
              {stats?.serviceBreakdown && stats.serviceBreakdown.length > 0 && (
                <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                  <h3 className="mb-4 text-lg font-bold text-brand-900">توزيع الحجوزات حسب الخدمة</h3>
                  <div className="space-y-2">
                    {stats.serviceBreakdown.map(s => {
                      const maxSvc = Math.max(1, ...stats.serviceBreakdown.map(x => x.count))
                      return (
                        <div key={s.serviceId || 'none'} className="flex items-center gap-3">
                          <span className="w-32 text-sm font-medium text-slate-600 text-right">{s.name}</span>
                          <div className="flex-1">
                            <div className="h-6 overflow-hidden rounded-full bg-slate-100">
                              <div className="flex h-full items-center rounded-full bg-brand-500 px-2 text-xs font-bold text-white" style={{ width: `${(s.count / maxSvc) * 100}%`, minWidth: '2rem' }}>
                                {s.count}
                              </div>
                            </div>
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>
              )}

              {/* Completed bookings revenue detail */}
              <div className="rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.04)]">
                <h3 className="mb-4 text-lg font-bold text-brand-900">تفاصيل الإيرادات</h3>
                {completedBookings.length === 0 ? (
                  <p className="text-sm text-slate-400">لا توجد حجوزات مكتملة</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-slate-100 text-right text-xs text-slate-400">
                          <th className="pb-2 pr-3">الكود</th>
                          <th className="pb-2">الزبون</th>
                          <th className="pb-2">التكلفة المقدرة</th>
                          <th className="pb-2">المستلم</th>
                          <th className="pb-2">المقدم</th>
                          <th className="pb-2">الحالة</th>
                        </tr>
                      </thead>
                      <tbody>
                        {completedBookings.slice(0, 30).map(b => (
                          <tr key={b.id} className="border-b border-slate-50">
                            <td className="py-2 pr-3 font-mono font-bold text-brand-600">{b.code}</td>
                            <td className="text-slate-700">{b.customer.name}</td>
                            <td className="text-slate-600">{(b.quotedPrice || 0).toLocaleString('ar-IQ')}</td>
                            <td className="font-bold text-brand-700">{(b.amountCollected || 0).toLocaleString('ar-IQ')}</td>
                            <td className="text-slate-600">{(b.advancePaid || 0).toLocaleString('ar-IQ')}</td>
                            <td>
                              <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${
                                b.amountVerified ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
                              }`}>{b.amountVerified ? 'مدقق' : 'بانتظار'}</span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* ═══ Vehicles & Technician Wash Rating Tab (تذكير بس — بدون تعديل تلقائي للراتب) ═══ */}
          {tab === 'vehicles' && (
            <div className="space-y-6">
              <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
                ⚠️ هذا تذكير فقط لآخر 30 يوم — النتائج والرواتب المقترحة هنا معلوماتية بس، ماكو أي تعديل تلقائي براتب أي موظف. لو تريد تزيد راتب فني، سوّيها يدوياً من صفحة إدارة الكوادر.
              </div>

              <div className="rounded-2xl border border-slate-100 bg-white p-5 shadow-[0_1px_3px_rgba(0,0,0,0.06)]">
                <h3 className="mb-4 font-bold text-slate-800">متوسط تقييم السيارات (آخر 30 يوم)</h3>
                <div className="overflow-x-auto">
                  <table className="w-full text-right text-sm">
                    <thead className="text-slate-500">
                      <tr>
                        <th className="pb-2">السيارة</th>
                        <th className="pb-2">عدد التقييمات</th>
                        <th className="pb-2">متوسط النتيجة</th>
                      </tr>
                    </thead>
                    <tbody>
                      {vehicleScores.map(v => (
                        <tr key={v.vehicleId} className="border-b border-slate-50">
                          <td className="py-2 font-bold text-slate-800">{v.vehicleName}</td>
                          <td className="text-slate-600">{v.ratingsCount}</td>
                          <td>
                            <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                              v.averageScore >= 0.85 ? 'bg-emerald-100 text-emerald-700' :
                              v.averageScore >= 0.7 ? 'bg-lime-100 text-lime-700' :
                              v.averageScore >= 0.5 ? 'bg-amber-100 text-amber-700' : 'bg-red-100 text-red-700'
                            }`}>{Math.round(v.averageScore * 100)}%</span>
                          </td>
                        </tr>
                      ))}
                      {vehicleScores.length === 0 && (
                        <tr><td colSpan={3} className="py-6 text-center text-slate-400">لا توجد تقييمات مسجلة آخر 30 يوم</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              <div className="rounded-2xl border border-slate-100 bg-white p-5 shadow-[0_1px_3px_rgba(0,0,0,0.06)]">
                <h3 className="mb-4 font-bold text-slate-800">الراتب المقترح للفنيين حسب جودة الغسيل (آخر 30 يوم)</h3>
                <div className="overflow-x-auto">
                  <table className="w-full text-right text-sm">
                    <thead className="text-slate-500">
                      <tr>
                        <th className="pb-2">الفني</th>
                        <th className="pb-2">سيارات غسلها</th>
                        <th className="pb-2">مجموع النقاط</th>
                        <th className="pb-2">الراتب المقترح</th>
                      </tr>
                    </thead>
                    <tbody>
                      {techWashSummaries.map(t => (
                        <tr key={t.employeeId} className="border-b border-slate-50">
                          <td className="py-2 font-bold text-slate-800">{t.employeeName}</td>
                          <td className="text-slate-600">{t.vehiclesWashed}</td>
                          <td className="text-slate-600">{t.totalPoints}</td>
                          <td className="font-bold text-brand-700">
                            {t.suggestedWage.toLocaleString('ar-IQ')} د.ع
                            {t.suggestedWage >= t.monthlyCap && <span className="mr-1 text-[10px] text-amber-600">(بلغ الحد الأعلى)</span>}
                          </td>
                        </tr>
                      ))}
                      {techWashSummaries.length === 0 && (
                        <tr><td colSpan={4} className="py-6 text-center text-slate-400">لا توجد تقييمات غسيل مسجلة آخر 30 يوم</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
