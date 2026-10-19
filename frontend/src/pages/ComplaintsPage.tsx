import { useCallback, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, complaintTypeLabels, type Complaint, type ComplaintCustomerStat, type ComplaintType, type Customer, type Employee } from '../api'
import { useSession } from '../session'
import ComplaintsTracking from '../components/ComplaintsTracking'



// أبو الجودة والمراقب المدقق (والأدمن) يشوفون متابعة/إدارة الشكاوى بس —
// باقي الأدوار الي عندها صلاحية الشكاوى (مثلاً المبيعات) تشوف تسجيل شكوى
// جديدة بس، بدون واجهة المتابعة والإدارة.
const trackingRoles = ['QUALITY_ENGINEER', 'MONITOR', 'ADMIN', 'OWNER']
// صاحب صلاحية monitoring = مراقب (ع 10-16)
// الاتصال بالزبون مو حكر على مهندس الجودة: أي موظف ينطيه المدير صلاحية
// complaint_contact يكدر يتصل ويأشر النتيجة، واسمه ينحفظ كدام الشكوى.
const contactRoles = ['QUALITY_ENGINEER', 'MONITOR', 'ADMIN', 'OWNER']
// مهندس الجودة (والأدمن) يشوف تفاصيل الزبون كاملة ويتواصل معه مباشرة.
// المراقب المدقق يشوف بس "إنذار" إنه اكو شكوى تحتاج متابعة، بدون تفاصيل
// الزبون — دوره رقابي بس، مو التواصل المباشر.
const fullDetailRoles = ['QUALITY_ENGINEER', 'ADMIN']

export default function ComplaintsPage() {
  const { employee: currentUser, permissions } = useSession()
  const canContact = !!currentUser
    && (contactRoles.includes(currentUser.role) || permissions.includes('complaint_contact') || permissions.includes('monitoring'))
  const canTrack = (!!currentUser && trackingRoles.includes(currentUser.role)) || canContact
  const canSeeFullDetail = !!currentUser && fullDetailRoles.includes(currentUser.role)
  // ⚠️ نفس قائمة المسار بالخادم بالضبط (main.go: ADMIN/OWNER/MONITOR).
  // المالك ينوصل بدوره المطبَّع ADMIN، وactualRole للتوثيق.
  const canAudit = !!currentUser
    && (['ADMIN', 'OWNER', 'MONITOR'].includes(currentUser.role) || permissions.includes('monitoring'))
  const [complaints, setComplaints] = useState<Complaint[]>([])
  // ═══ الفلاتر (ع) 10-15 — تنحفظ بالرابط حتى «تابع» من قائمة الإجراءات يفتحها مفلترة ═══
  const [params, setParams] = useSearchParams()
  const fl = {
    type: params.get('type') ?? '', vip: params.get('vip') ?? '', status: params.get('status') ?? 'OPEN',
    late: params.get('late') === '1', rel: params.get('rel') ?? '', asg: params.get('asg') ?? '',
    from: params.get('from') ?? '', to: params.get('to') ?? '', q: params.get('q') ?? '',
  }
  const setFl = (k: string, v: string) => {
    const n = new URLSearchParams(params)
    if (v) n.set(k, v); else n.delete(k)
    setParams(n, { replace: true })
  }
  const [nowMs] = useState(() => Date.now())
  const [customers, setCustomers] = useState<Customer[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [stats, setStats] = useState<ComplaintCustomerStat[]>([])
  const [contactBusy, setContactBusy] = useState<string | null>(null)
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
  // ⚠️ التقييم لازم يوصل للخادم بنفس الطلب. TypeScript يقبل دالة
  // بمعاملين مكان دالة بثلاثة بلا خطأ — فبدون هذا المعامل يضيع
  // التقييم بصمت والمهندس يظن إنه انحفظ.
  const setContacted = async (c: Complaint, contacted: boolean, rating?: number | null) => {
    setContactBusy(c.id)
    try {
      const updated = await api.setComplaintContacted(c.id, contacted, rating)
      setComplaints((prev) => prev.map((x) => (x.id === c.id ? updated : x)))
      loadStats()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تحديث حالة الاتصال')
    } finally {
      setContactBusy(null)
    }
  }

  // «الزبون ما رد»: ٣ محاولات بينهن ٢٤ ساعة، والثالثة تسقط الشكوى — (ع) 10-19
  const noAnswer = async (c: Complaint) => {
    const left = 3 - (c.noAnswerCount ?? 0)
    if (!confirm(left <= 1 ? 'هاي ثالث مرة الزبون ما يرد — الشكوى راح تسقط وتنحسب «تواصلنا والزبون ما رد». متأكد؟' : `تسجّل «الزبون ما رد»؟ الشكوى تتأجل ٢٤ ساعة وترجعلك تتصل مرة ثانية (باقي ${left - 1} محاولة قبل ما تسقط).`)) return
    setContactBusy(c.id)
    try {
      const updated = await api.complaintNoAnswer(c.id)
      setComplaints((prev) => prev.map((x) => (x.id === c.id ? updated : x)))
      loadStats()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التسجيل')
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

  const isLate = (c: Complaint) => (c.status === 'NEW' || c.status === 'IN_PROGRESS') &&
    nowMs - new Date(c.createdAt).getTime() > (c.isVip ? 24 : 72) * 3600_000
  const filtered = useMemo(() => complaints.filter((c) => {
    if (fl.type && c.type !== fl.type) return false
    if (fl.vip === '1' && !c.isVip) return false
    if (fl.vip === '0' && c.isVip) return false
    // (ع) 10-16: الافتراضي «المفتوحة» — الشغل الي لازم يخلص، مو ٦٠٠ شكوى قديمة.
    if (fl.status === 'OPEN') { if (c.status !== 'NEW' && c.status !== 'IN_PROGRESS') return false }
    else if (fl.status && fl.status !== 'ALL' && c.status !== fl.status) return false
    if (fl.late && !isLate(c)) return false
    if (fl.rel && c.relatedEmployee?.id !== fl.rel) return false
    if (fl.asg === 'none' && c.assignedToEmployee) return false
    if (fl.asg && fl.asg !== 'none' && c.assignedToEmployee?.id !== fl.asg) return false
    const day = new Date(c.createdAt).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
    if (fl.from && day < fl.from) return false
    if (fl.to && day > fl.to) return false
    if (fl.q) {
      const q = fl.q.trim().toLowerCase()
      const hay = [c.customer?.name, c.customer?.phone, (c as Complaint & { booking?: { code?: string } }).booking?.code, c.description].join(' ').toLowerCase()
      if (!hay.includes(q)) return false
    }
    return true
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }), [complaints, params, nowMs])
  const filteredStats = useMemo(() => {
    const ids = new Set(filtered.map((c) => c.customer?.id))
    return stats.filter((s) => ids.has(s.customerId))
  }, [filtered, stats])
  const anyFilter = [...params.keys()].length > 0
  const vipCount = complaints.filter((c) => c.isVip && (c.status === 'NEW' || c.status === 'IN_PROGRESS')).length
  const lateCount = complaints.filter(isLate).length

  return (
    <div>
      {/* (ع) 10-13: بوضع المتابعة، «ComplaintsTracking» عنده ترويسته — فما نكررها */}
      {!canTrack && (
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
      )}

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

{canTrack && (
        <ComplaintsTracking
          toolbar={
        <div className="mb-3 flex flex-wrap items-center gap-2 rounded-xl bg-white p-3 shadow-sm">
          <select value={fl.type} onChange={(e) => setFl('type', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="">كل أنواع المشكلة</option>
            {Object.entries(complaintTypeLabels).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
          <select value={fl.vip} onChange={(e) => setFl('vip', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="">كل الزبائن</option>
            <option value="1">👑 شخصيات مهمة بس</option>
            <option value="0">الزبائن العاديين</option>
          </select>
          <select value={fl.status} onChange={(e) => setFl('status', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="OPEN">المفتوحة (جديدة + قيد المعالجة)</option>
            <option value="ALL">كل الحالات</option>
            <option value="NEW">جديدة</option><option value="IN_PROGRESS">قيد المعالجة</option>
            <option value="RESOLVED">محلولة</option><option value="CLOSED">مغلقة</option>
          </select>
          <select value={fl.rel} onChange={(e) => setFl('rel', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="">كل الموظفين المسبّبين</option>
            {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <select value={fl.asg} onChange={(e) => setFl('asg', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
            <option value="">كل المسؤولين</option>
            <option value="none">بلا مسؤول</option>
            {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <label className="text-xs text-slate-500">من</label>
          <input type="date" value={fl.from} onChange={(e) => setFl('from', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
          <label className="text-xs text-slate-500">إلى</label>
          <input type="date" value={fl.to} onChange={(e) => setFl('to', e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
          <input value={fl.q} onChange={(e) => setFl('q', e.target.value)} placeholder="بحث: زبون، رقم، كود حجز" className="min-w-[180px] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
          <button onClick={() => setFl('late', fl.late ? '' : '1')} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${fl.late ? 'bg-red-600 text-white' : 'bg-red-50 text-red-700'}`}>⏰ المتأخرة ({lateCount})</button>
          <button onClick={() => setFl('vip', fl.vip === '1' ? '' : '1')} className={`rounded-lg px-3 py-1.5 text-sm font-bold ${fl.vip === '1' ? 'bg-amber-500 text-white' : 'bg-amber-50 text-amber-800'}`}>👑 مفتوحة لشخصيات مهمة ({vipCount})</button>
          {anyFilter && <button onClick={() => setParams({}, { replace: true })} className="text-sm text-slate-500 underline">مسح الفلاتر</button>}
          <span className="rounded-lg bg-amber-50 px-3 py-1.5 text-xs font-bold text-amber-800" title="الزبون ما رد — تنتظر الاتصال الجاي (مو عاجلة)">📵 تنتظر إعادة اتصال ({complaints.filter((c) => !c.contactedAt && !c.droppedNoAnswerAt && c.nextCallAt && new Date(c.nextCallAt) > new Date()).length})</span>
          <span className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700" title="تواصلنا ٣ مرات والزبون ما رد — الشكوى سقطت">🔕 سقطت لأن الزبون ما رد ({complaints.filter((c) => c.droppedNoAnswerAt).length})</span>
          <span className="mr-auto text-xs text-slate-500">{filtered.length} من {complaints.length} شكوى</span>
        </div>
                }
          complaints={filtered}
          stats={filteredStats}
          employees={employees}
          canContact={canContact}
          canAudit={canAudit}
          canSeeFullDetail={canSeeFullDetail}
          busyId={contactBusy}
          onContact={setContacted}
          onNoAnswer={noAnswer}
          onNotes={saveNotes}
          onAssign={handleAssign}
          onStatus={handleStatusChange}
          onResolveClick={(id: string) => setResolvingId(id)}
          onAudited={() => { load(); loadStats() }}
        />
      )}
    </div>
  )
}
