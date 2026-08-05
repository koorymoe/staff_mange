import { Fragment, useCallback, useEffect, useState } from 'react'
import { api, complaintTypeLabels, type Complaint, type ComplaintCustomerStat, type ComplaintType, type Customer, type Employee } from '../api'
import { useSession } from '../session'

const statusLabels: Record<Complaint['status'], string> = {
  NEW: 'جديدة',
  IN_PROGRESS: 'قيد المعالجة',
  RESOLVED: 'تم الحل',
  CLOSED: 'مغلقة',
}

const statusColors: Record<Complaint['status'], string> = {
  NEW: 'bg-yellow-100 text-yellow-800',
  IN_PROGRESS: 'bg-blue-100 text-blue-800',
  RESOLVED: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

// أبو الجودة والمراقب المدقق (والأدمن) يشوفون متابعة/إدارة الشكاوى بس —
// باقي الأدوار الي عندها صلاحية الشكاوى (مثلاً المبيعات) تشوف تسجيل شكوى
// جديدة بس، بدون واجهة المتابعة والإدارة.
const trackingRoles = ['QUALITY_ENGINEER', 'MONITOR', 'ADMIN', 'OWNER']
// الاتصال بالزبون مو حكر على مهندس الجودة: أي موظف ينطيه المدير صلاحية
// complaint_contact يكدر يتصل ويأشر النتيجة، واسمه ينحفظ كدام الشكوى.
const contactRoles = ['QUALITY_ENGINEER', 'MONITOR', 'ADMIN', 'OWNER']
// مهندس الجودة (والأدمن) يشوف تفاصيل الزبون كاملة ويتواصل معه مباشرة.
// المراقب المدقق يشوف بس "إنذار" إنه اكو شكوى تحتاج متابعة، بدون تفاصيل
// الزبون — دوره رقابي بس، مو التواصل المباشر.
const fullDetailRoles = ['QUALITY_ENGINEER', 'ADMIN']

// حالة الاتصال — قائمة منسدلة بخيارين صريحين بدل السويچ: الموظف يختار
// شنو سوّى، ما يخمّن شنو معنى اليمين واليسار. أخضر «تم» وأحمر «لم يتم».
//
// `canEdit` false يعني الموظف يشوف الحالة بس ما يغيّرها — الاتصال
// بالزبون صلاحية مستقلة (complaint_contact) ينطيها المدير لمن يريد.
function ContactSelect({ complaint, busy, canEdit, onChange }: {
  complaint: Complaint
  busy: boolean
  canEdit: boolean
  onChange: (contacted: boolean) => void
}) {
  const on = !!complaint.contactedAt
  const tone = on
    ? 'border-emerald-300 bg-emerald-50 text-emerald-700'
    : 'border-red-300 bg-red-50 text-red-700'
  if (!canEdit) {
    return (
      <span className={`inline-block rounded-full border px-3 py-1 text-xs font-bold ${tone}`}>
        {on ? '✔ تم الاتصال' : '✕ لم يتم الاتصال'}
      </span>
    )
  }
  return (
    <select
      value={on ? 'YES' : 'NO'}
      disabled={busy}
      onChange={(e) => onChange(e.target.value === 'YES')}
      className={`rounded-lg border px-2 py-1.5 text-xs font-bold outline-none disabled:opacity-50 ${tone}`}
    >
      <option value="NO">✕ لم يتم الاتصال</option>
      <option value="YES">✔ تم الاتصال</option>
    </select>
  )
}

export default function ComplaintsPage() {
  const { employee: currentUser, permissions } = useSession()
  const canContact = !!currentUser
    && (contactRoles.includes(currentUser.role) || permissions.includes('complaint_contact'))
  const canTrack = (!!currentUser && trackingRoles.includes(currentUser.role)) || canContact
  const canSeeFullDetail = !!currentUser && fullDetailRoles.includes(currentUser.role)
  const [complaints, setComplaints] = useState<Complaint[]>([])
  const [customers, setCustomers] = useState<Customer[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [stats, setStats] = useState<ComplaintCustomerStat[]>([])
  const [contactBusy, setContactBusy] = useState<string | null>(null)
  // الزبون المفتوح بتقرير الشكاوى — نعرض شكاواه وحدة وحدة تحته
  const [expandedCustomer, setExpandedCustomer] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)

  // Form state — رقم الهاتف والاسم حقلين منفصلين، وترابطهم ثنائي الاتجاه:
  // اكتب رقم صحيح يعبّي الاسم تلقائياً، واكتب اسم موجود يعبّي الرقم تلقائياً،
  // وبنفس الوقت الاثنين قابلين للتعديل اليدوي بأي وقت.
  const [phone, setPhone] = useState('')
  const [name, setName] = useState('')
  const [foundCustomer, setFoundCustomer] = useState<Customer | null>(null)
  const [searchingCustomer, setSearchingCustomer] = useState(false)
  const [complaintType, setComplaintType] = useState<ComplaintType | ''>('')
  const [relatedEmployeeId, setRelatedEmployeeId] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Resolution modal
  const [resolvingId, setResolvingId] = useState<string | null>(null)
  const [resolution, setResolution] = useState('')

  const load = () => {
    Promise.all([api.getComplaints(), api.getCustomers(), api.getEmployees()])
      .then(([comps, custs, emps]) => {
        setComplaints(comps)
        setCustomers(custs)
        setEmployees(emps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  // تقرير عدد شكاوى كل زبون — منفصل تماماً عن إحصائيات الحجوزات.
  // ينعاد تحميله بعد تأشير الاتصال حتى عمود «الحالة» يتحدث فوراً.
  const loadStats = useCallback(() => {
    if (canTrack) api.getComplaintStats().then(setStats).catch(() => {})
  }, [canTrack])

  useEffect(load, [])
  useEffect(loadStats, [loadStats])
  // الأدوار الي بس تسجل شكوى (بدون متابعة) تشوف الفورم مباشرة بدون زر تبديل
  useEffect(() => {
    // Auto-opening the complaint form for roles without a tracking view is a
    // derived-state sync from the session role, not a fetch.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (currentUser && !canTrack) setShowForm(true)
  }, [currentUser, canTrack])

  // رقم الهاتف → يعبّي الاسم تلقائياً (بحث بالباك إند لأنه أدق من القائمة المحلية)
  const handlePhoneChange = async (value: string) => {
    setPhone(value)
    setFoundCustomer(null)
    if (!/^\d{11}$/.test(value.trim())) return
    setSearchingCustomer(true)
    try {
      const c = await api.lookupCustomer(value.trim())
      if (c) {
        setFoundCustomer(c)
        setName(c.name)
      }
    } finally {
      setSearchingCustomer(false)
    }
  }

  // الاسم → يعبّي رقم الهاتف تلقائياً إذا طابق زبون موجود بالضبط
  const handleNameChange = (value: string) => {
    setName(value)
    const match = customers.find((c) => c.name.trim() === value.trim())
    if (match) {
      setFoundCustomer(match)
      setPhone(match.phone)
    } else {
      setFoundCustomer(null)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!currentUser || !complaintType) return
    // نتأكد من تطابق الزبون بلحظة الإرسال (مو بس الاعتماد على آخر بحث)، لأنه
    // ممكن المستخدم يعدل الحقلين يدوياً بعد آخر تطابق تلقائي.
    const matched =
      foundCustomer ||
      customers.find((c) => c.phone.trim() === phone.trim() || c.name.trim() === name.trim())
    if (!matched) {
      alert('تعذر تحديد الزبون — تأكد من رقم الهاتف أو الاسم مطابق لزبون مسجل بالنظام')
      return
    }
    setSubmitting(true)
    try {
      await api.createComplaint({
        customerId: matched.id,
        type: complaintType,
        description,
        relatedEmployeeId: relatedEmployeeId || undefined,
        createdByEmployeeId: currentUser.id,
      })
      setPhone('')
      setName('')
      setFoundCustomer(null)
      setComplaintType('')
      setRelatedEmployeeId('')
      setDescription('')
      setShowForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  // تأشير الاتصال بالزبون — النظام يخزن منو اتصل ومتى، ويطلع كدام الشكوى
  const setContacted = async (c: Complaint, contacted: boolean) => {
    setContactBusy(c.id)
    try {
      const updated = await api.setComplaintContacted(c.id, contacted)
      setComplaints((prev) => prev.map((x) => (x.id === c.id ? updated : x)))
      loadStats()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تحديث حالة الاتصال')
    } finally {
      setContactBusy(null)
    }
  }

  const saveNotes = async (c: Complaint, value: string) => {
    if ((c.notes || '') === value.trim()) return
    try {
      const updated = await api.setComplaintNotes(c.id, value.trim())
      setComplaints((prev) => prev.map((x) => (x.id === c.id ? updated : x)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ الملاحظات')
    }
  }

  const handleAssign = async (complaintId: string, employeeId: string) => {
    try {
      await api.updateComplaint(complaintId, { assignedToEmployeeId: employeeId })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleStatusChange = async (complaintId: string, status: Complaint['status']) => {
    try {
      await api.updateComplaint(complaintId, { status })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleResolve = async () => {
    if (!resolvingId || !resolution.trim()) return
    try {
      await api.resolveComplaint(resolvingId, resolution)
      setResolvingId(null)
      setResolution('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">شكاوى الصيانة</h2>
          <p className="mt-1 text-slate-500">إدارة شكاوى العملاء ومتابعة حالتها.</p>
        </div>
        {!canTrack && (
          <button
            onClick={() => setShowForm(!showForm)}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
          >
            شكوى جديدة
          </button>
        )}
      </div>

      {/* New complaint form */}
      {!canTrack && showForm && (
        <form
          onSubmit={handleSubmit}
          className="mt-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
        >
          <h3 className="mb-4 text-lg font-bold text-brand-800">تسجيل شكوى جديدة</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">رقم هاتف الزبون</label>
              <input
                value={phone}
                onChange={(e) => handlePhoneChange(e.target.value)}
                placeholder="أدخل رقم الهاتف"
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              />
              {searchingCustomer && <p className="mt-1 text-xs text-slate-400">جاري البحث...</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون</label>
              <input
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="أدخل اسم الزبون"
                list="complaint-customer-names"
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              />
              <datalist id="complaint-customer-names">
                {customers.map((c) => <option key={c.id} value={c.name} />)}
              </datalist>
            </div>
          </div>
          {foundCustomer && (
            <p className="mt-2 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">
              الزبون: {foundCustomer.name} - {foundCustomer.phone} (كود: {foundCustomer.code})
            </p>
          )}

          <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">نوع الشكوى</label>
              <select
                required
                value={complaintType}
                onChange={(e) => setComplaintType(e.target.value as ComplaintType)}
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              >
                <option value="">اختر نوع الشكوى</option>
                {Object.entries(complaintTypeLabels).map(([k, v]) => (
                  <option key={k} value={k}>{v}</option>
                ))}
              </select>
            </div>
            {complaintType && complaintType !== 'OTHER' && (
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">
                  الموظف المتسبب (اختياري)
                </label>
                <select
                  value={relatedEmployeeId}
                  onChange={(e) => setRelatedEmployeeId(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                >
                  <option value="">غير محدد / لا يتذكر</option>
                  {employees.map((emp) => (
                    <option key={emp.id} value={emp.id}>{emp.name}</option>
                  ))}
                </select>
              </div>
            )}
          </div>

          <div className="mt-4">
            <label className="mb-1 block text-sm font-medium text-slate-600">تفاصيل إضافية (اختياري)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
            />
          </div>

          <div className="mt-4">
            <button
              type="submit"
              disabled={submitting || !complaintType}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
            >
              {submitting ? 'جاري الحفظ...' : 'تسجيل الشكوى'}
            </button>
          </div>
        </form>
      )}

      {/* Resolution modal */}
      {canTrack && resolvingId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-bold text-brand-800">حل الشكوى</h3>
            <textarea
              value={resolution}
              onChange={(e) => setResolution(e.target.value)}
              rows={4}
              placeholder="أدخل تفاصيل الحل..."
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
            />
            <div className="mt-4 flex gap-3">
              <button
                onClick={handleResolve}
                className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md"
              >
                حفظ الحل
              </button>
              <button
                onClick={() => { setResolvingId(null); setResolution('') }}
                className="rounded-lg border border-slate-300 px-6 py-2 font-medium text-slate-600"
              >
                إلغاء
              </button>
            </div>
          </div>
        </div>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {canTrack && stats.length > 0 && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="border-b border-slate-100 px-4 py-3 text-sm font-bold text-brand-800">
            تقرير الشكاوى حسب الزبون
          </h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right text-sm">
              <thead className="bg-slate-50">
                <tr>
                  <th className="px-4 py-2 font-semibold text-slate-600">الزبون</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">الهاتف</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">عدد الشكاوى</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">الحالة</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {stats.map((s) => {
                  const rows = complaints.filter((c) => c.customer?.id === s.customerId)
                  const open = expandedCustomer === s.customerId
                  return (
                    <Fragment key={s.customerId}>
                      <tr
                        className="cursor-pointer transition-colors hover:bg-slate-50"
                        onClick={() => setExpandedCustomer(open ? null : s.customerId)}
                      >
                        <td className="px-4 py-2 font-medium">
                          <span className="ml-1 text-slate-400">{open ? '▾' : '◂'}</span>
                          {s.customerName}
                        </td>
                        <td className="px-4 py-2 text-slate-500">{s.customerPhone}</td>
                        <td className="px-4 py-2 font-bold text-brand-700">{s.complaintCount}</td>
                        {/* الحالة = هل انتصلنا بيه، مو «مفتوحة حالياً» */}
                        <td className="px-4 py-2">
                          {s.notContactedCount > 0 ? (
                            <span className="rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                              ✕ لم يتم الاتصال{s.notContactedCount > 1 ? ` (${s.notContactedCount})` : ''}
                            </span>
                          ) : (
                            <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
                              ✔ تم الاتصال
                            </span>
                          )}
                        </td>
                      </tr>
                      {/* شكاوى هذا الزبون وحدة وحدة: كل شكوى إلها قائمة
                          منسدلة مستقلة واسم الموظف الي اتصل وحلّها. */}
                      {open && (
                        <tr className="bg-slate-50">
                          <td colSpan={4} className="px-4 py-3">
                            {rows.length === 0 && <p className="text-xs text-slate-400">ماكو تفاصيل متاحة إلك.</p>}
                            <div className="space-y-2">
                              {rows.map((c) => (
                                <div key={c.id} className="flex flex-wrap items-center justify-between gap-3 rounded-lg bg-white px-3 py-2 shadow-sm">
                                  <div className="text-xs text-slate-500">
                                    {c.contactedAt ? (
                                      <>اتصل: <b className="text-slate-700">{c.contactedByName || '—'}</b>
                                        {' · '}{new Date(c.contactedAt).toLocaleDateString('ar-IQ')}</>
                                    ) : 'ما اتصل بيه أحد بعد'}
                                  </div>
                                  <div className="flex items-center gap-3">
                                    <ContactSelect
                                      complaint={c}
                                      busy={contactBusy === c.id}
                                      canEdit={canContact}
                                      onChange={(v) => setContacted(c, v)}
                                    />
                                    <span className="text-sm font-medium text-brand-800">
                                      {complaintTypeLabels[c.type] || c.type}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* المراقب: يشوف الشكاوى ويتحكم بسويچ الاتصال وبالملاحظات — بدون
          تفاصيل الزبون، لأن دوره رقابي مو تواصل مباشر. */}
      {canTrack && !loading && !error && !canSeeFullDetail && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الشكوى</th>
                  <th className="px-4 py-3 text-sm font-semibold">الاتصال بالزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">منو اتصل</th>
                  <th className="px-4 py-3 text-sm font-semibold">ملاحظات الزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                  <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {complaints.map((c) => (
                  <tr key={c.id} className="transition-colors hover:bg-slate-50">
                    <td className="px-4 py-3">
                      <span className="font-medium text-brand-800">{complaintTypeLabels[c.type] || c.type}</span>
                      {!c.contactedAt && (
                        <div className="text-xs text-amber-700">⚠️ تحتاج متابعة</div>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <ContactSelect complaint={c} busy={contactBusy === c.id} canEdit={canContact} onChange={(v) => setContacted(c, v)} />
                    </td>
                    <td className="px-4 py-3 text-xs text-slate-500">
                      {c.contactedAt
                        ? <>{c.contactedByName || '—'}<div className="text-[11px] text-slate-400">{new Date(c.contactedAt).toLocaleDateString('ar-IQ')}</div></>
                        : '—'}
                    </td>
                    <td className="max-w-[220px] px-4 py-3">
                      <input
                        defaultValue={c.notes || ''}
                        onBlur={(e) => saveNotes(c, e.target.value)}
                        placeholder="ملاحظات الزبون..."
                        className="w-full rounded-lg border border-slate-200 px-2 py-1 text-xs outline-none focus:border-brand-500"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[c.status]}`}>
                        {statusLabels[c.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-500">
                      {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                  </tr>
                ))}
                {complaints.length === 0 && (
                  <tr>
                    <td colSpan={6} className="px-4 py-6 text-center text-slate-400">
                      لا توجد شكاوى بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {canSeeFullDetail && !loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">الوصف</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                  <th className="px-4 py-3 text-sm font-semibold">الاتصال بالزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">ملاحظات الزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">المسؤول</th>
                  <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحل</th>
                  <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {complaints.map((c) => (
                  <tr key={c.id} className="transition-colors hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium">
                      {c.customer?.name || '—'}
                      <div className="text-xs font-normal text-slate-400">{c.customer?.phone}</div>
                    </td>
                    <td className="max-w-xs px-4 py-3 text-slate-600">
                      <span className="font-medium text-brand-800">{complaintTypeLabels[c.type] || c.type}</span>
                      {c.relatedEmployee && (
                        <div className="text-xs text-amber-600">الموظف المتسبب: {c.relatedEmployee.name}</div>
                      )}
                      {c.description && <div className="truncate text-xs text-slate-400">{c.description}</div>}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[c.status]}`}>
                        {statusLabels[c.status]}
                      </span>
                    </td>
                    {/* سويچ الاتصال — أخضر «تم» وأحمر «لم يتم»، ويخزن منو اتصل */}
                    <td className="px-4 py-3">
                      <ContactSelect complaint={c} busy={contactBusy === c.id} canEdit={canContact} onChange={(v) => setContacted(c, v)} />
                      {c.contactedAt && (
                        <div className="mt-1 text-[11px] text-slate-500">
                          اتصل: <b className="text-slate-600">{c.contactedByName || '—'}</b>
                          {' · '}{new Date(c.contactedAt).toLocaleDateString('ar-IQ')}
                        </div>
                      )}
                    </td>
                    <td className="max-w-[200px] px-4 py-3">
                      <input
                        defaultValue={c.notes || ''}
                        onBlur={(e) => saveNotes(c, e.target.value)}
                        placeholder="ملاحظات الزبون..."
                        className="w-full rounded-lg border border-slate-200 px-2 py-1 text-xs outline-none focus:border-brand-500"
                      />
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={c.assignedToEmployeeId || ''}
                        onChange={(e) => handleAssign(c.id, e.target.value)}
                        className="rounded-lg border border-slate-300 px-2 py-1 text-sm outline-none focus:border-brand-500"
                      >
                        <option value="">غير محدد</option>
                        {employees.map((emp) => (
                          <option key={emp.id} value={emp.id}>{emp.name}</option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3 text-slate-500">
                      {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td className="max-w-xs truncate px-4 py-3 text-slate-500">{c.resolution || '-'}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        {c.status === 'NEW' && (
                          <button
                            onClick={() => handleStatusChange(c.id, 'IN_PROGRESS')}
                            className="rounded-lg bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 hover:bg-blue-200"
                          >
                            بدء المعالجة
                          </button>
                        )}
                        {(c.status === 'NEW' || c.status === 'IN_PROGRESS') && (
                          <button
                            onClick={() => setResolvingId(c.id)}
                            className="rounded-lg bg-green-100 px-3 py-1 text-xs font-medium text-green-700 hover:bg-green-200"
                          >
                            حل
                          </button>
                        )}
                        {c.status === 'RESOLVED' && (
                          <button
                            onClick={() => handleStatusChange(c.id, 'CLOSED')}
                            className="rounded-lg bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-200"
                          >
                            إغلاق
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {complaints.length === 0 && (
                  <tr>
                    <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                      لا توجد شكاوى بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
