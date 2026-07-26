import { useState } from 'react'
import { api, type Division, type Employee, type EmployeeRole, type Permission } from '../api'

const roleLabels: Record<EmployeeRole, string> = {
  ADMIN: 'مدير النظام',
  SALES: 'موظف مبيعات',
  HR_COORDINATOR: 'إداري الكوادر',
  TECHNICIAN: 'فني',
  PROJECT_MANAGER: 'مدير مشاريع',
  MONITOR: 'مراقب / مدقق',
  FINANCE: 'محاسب',
  GPS_ADMIN: 'مسؤول GPS',
  QUALITY_ENGINEER: 'مهندس جودة',
  ENGINEER: 'مهندس',
  PROCUREMENT_ADMIN: 'إداري الكميات',
  DESIGNER: 'مصمم',
  SERVICE_MANAGER: 'مسؤول خدمة',
  OWNER: 'مالك النظام',
}

// دور OWNER محجوز لحساب مالك النظام الوحيد المزروع مباشرة بقاعدة البيانات —
// محد يقدر يمنحه لموظف من الواجهة، حتى الأدمن نفسه.
// دور "مسؤول GPS" ما ينعطى بعد الآن لموظف جديد — الجي بي اس صارت خدمة وحدة
// بين عدة خدمات، والمسؤولية عنها تنعطى عن طريق "مسؤولو الخدمات" (صلاحية
// gps_system) مو دور وظيفي منفصل، حتى تنسجم مع "مسؤول خدمة" العام.
// "مهندس" و"مصمم" ينعطون مباشرة وقت الإنشاء بدون صلاحيات افتراضية —
// تُمنح لهم يدوياً من خطوة الصلاحيات بآخر الفورم (نفس نمط إداري الكميات).
const creatableRoles = (Object.keys(roleLabels) as EmployeeRole[]).filter(r => r !== 'OWNER' && r !== 'GPS_ADMIN')

function calcHours(start: string, end: string): number | null {
  if (!start || !end) return null
  const [sh, sm] = start.split(':').map(Number)
  const [eh, em] = end.split(':').map(Number)
  let minutes = (eh * 60 + em) - (sh * 60 + sm)
  if (minutes <= 0) minutes += 24 * 60 // يمتد لليوم التالي
  return Math.round((minutes / 60) * 10) / 10
}

const steps = ['الشعبة', 'الاسم', 'الشهادة والمنصب', 'الراتب', 'الدوام', 'الدور الوظيفي', 'بيانات الدخول', 'الصلاحيات'] as const

export default function AddEmployeeWizard({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [step, setStep] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // أول سؤال إلزامي بالفورم بالضبط متل ما طلب صاحب العمل: "شعبة هندسية" أو
  // "ديكور" — لا تظهر بقية الحقول قبل ما تنعطى إجابة هنا (division يبدأ null
  // عمداً، مو قيمة افتراضية، حتى "التالي" ما يشتغل بدون اختيار صريح).
  const [division, setDivision] = useState<Division | null>(null)
  const [name, setName] = useState('')
  const [certificate, setCertificate] = useState('')
  const [position, setPosition] = useState('')
  const [phone, setPhone] = useState('')
  const [jobTitle, setJobTitle] = useState('')
  const [salary, setSalary] = useState('')
  const [shiftStart, setShiftStart] = useState('08:00')
  const [shiftEnd, setShiftEnd] = useState('16:00')
  const [role, setRole] = useState<EmployeeRole>('TECHNICIAN')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')

  const [createdEmployee, setCreatedEmployee] = useState<Employee | null>(null)
  const [allPermissions, setAllPermissions] = useState<Permission[]>([])
  const [selectedPermIds, setSelectedPermIds] = useState<string[]>([])

  const hours = calcHours(shiftStart, shiftEnd)
  const isDecor = division === 'DECOR'

  const canProceed = () => {
    if (step === 0) return division !== null
    if (step === 1) return name.trim().length > 0
    if (step === 6) return username.trim().length > 0 && password.trim().length >= 4
    return true
  }

  const next = () => setStep(s => Math.min(s + 1, steps.length - 1))
  const back = () => setStep(s => Math.max(s - 1, 0))

  const handleCreate = async () => {
    setSubmitting(true)
    setError(null)
    try {
      const shift = hours !== null && hours <= 9 ? (Number(shiftStart.split(':')[0]) < 14 ? 'MORNING' : 'EVENING') : 'MORNING'
      // موظفو الديكور ينعطون دور TECHNICIAN دائماً (يعيد استخدام صلاحيات الفني
      // العادي المحدودة جداً أصلاً) بغض النظر عمن اختير بخطوة "الدور الوظيفي" —
      // الباك إند يفرض هذا برضه (دفاع مضاعف)، هذا فقط يطابق ما سيُحفظ فعلياً.
      const effectiveRole: EmployeeRole = isDecor ? 'TECHNICIAN' : role
      const created = await api.createEmployee({
        name,
        certificate: certificate || null,
        position: position || null,
        phone: phone || null,
        jobTitle: jobTitle || undefined,
        salary: salary ? Number(salary) : undefined,
        shift,
        shiftStart,
        shiftEnd,
        role: effectiveRole,
        division: division ?? 'ENGINEERING',
        username,
        password,
      })
      setCreatedEmployee(created)
      const perms = await api.getPermissions()
      setAllPermissions(perms)
      const roleNames = new Set(
        // نعرض تلقائياً الصلاحيات الافتراضية لهذا الدور كبداية مقترحة، والمدير يقدر يعدلها
        (await api.getRoleDefaults())[effectiveRole] || []
      )
      setSelectedPermIds(perms.filter(p => roleNames.has(p.name)).map(p => p.id))
      next()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ أثناء إنشاء الموظف')
    } finally {
      setSubmitting(false)
    }
  }

  const handleFinish = async () => {
    if (!createdEmployee) return
    setSubmitting(true)
    try {
      await api.setEmployeePermissions(createdEmployee.id, selectedPermIds)
      onCreated()
      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ أثناء حفظ الصلاحيات')
    } finally {
      setSubmitting(false)
    }
  }

  const togglePerm = (id: string) => {
    setSelectedPermIds(prev => prev.includes(id) ? prev.filter(p => p !== id) : [...prev, id])
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-xl rounded-2xl bg-white shadow-2xl overflow-hidden">
        {/* Progress header */}
        <div className="bg-gradient-to-l from-[#2c5aad] to-[#1e3f7a] p-5 text-white">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-bold">إضافة موظف جديد</h3>
            <button onClick={onClose} className="text-white/70 hover:text-white text-xl leading-none">×</button>
          </div>
          <div className="flex items-center gap-1.5">
            {steps.map((s, i) => (
              <div key={s} className={`h-1.5 flex-1 rounded-full transition-colors ${i <= step ? 'bg-white' : 'bg-white/20'}`} />
            ))}
          </div>
          <p className="mt-2 text-xs text-blue-100">خطوة {step + 1} من {steps.length} — {steps[step]}</p>
        </div>

        <div className="p-6 min-h-[220px]">
          {error && <p className="mb-4 rounded-lg bg-red-50 p-3 text-xs text-red-600">{error}</p>}

          {step === 0 && (
            <div className="flex flex-col gap-3">
              <label className="text-xs font-bold text-slate-500">هل هذا الموظف من الشعبة الهندسية أم شعبة الديكور؟ *</label>
              <p className="text-xs text-slate-400">هذا أول سؤال لازم يُجاب عليه — بقية بيانات الموظف تظهر بعده مباشرة، وكل شعبة إلها كتالوج مهارات مختلف تماماً.</p>
              <div className="grid grid-cols-2 gap-3">
                <button type="button" onClick={() => setDivision('ENGINEERING')}
                  className={`rounded-xl border-2 px-4 py-5 text-sm font-bold transition-all ${
                    division === 'ENGINEERING' ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-500 hover:bg-slate-50'
                  }`}>
                  🛠️ شعبة هندسية
                  <p className="mt-1 text-[11px] font-normal text-slate-400">كاميرات / شبكات / GPS وما شابه</p>
                </button>
                <button type="button" onClick={() => setDivision('DECOR')}
                  className={`rounded-xl border-2 px-4 py-5 text-sm font-bold transition-all ${
                    division === 'DECOR' ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-500 hover:bg-slate-50'
                  }`}>
                  🎨 شعبة ديكور
                  <p className="mt-1 text-[11px] font-normal text-slate-400">حدادة / نجارة / صباغة / سيراميك / لبخ / سباكة / جبس بورد</p>
                </button>
              </div>
            </div>
          )}

          {step === 1 && (
            <div className="flex flex-col gap-3">
              <label className="text-xs font-bold text-slate-500">اسم الموظف *</label>
              <input autoFocus value={name} onChange={e => setName(e.target.value)} placeholder="مثال: أحمد محمد"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
              <label className="text-xs font-bold text-slate-500 mt-2">رقم الهاتف</label>
              <input value={phone} onChange={e => setPhone(e.target.value)} placeholder="07xxxxxxxxx"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
            </div>
          )}

          {step === 2 && (
            <div className="flex flex-col gap-3">
              <label className="text-xs font-bold text-slate-500">الشهادة</label>
              <input value={certificate} onChange={e => setCertificate(e.target.value)} placeholder="مثال: بكالوريوس هندسة"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
              <label className="text-xs font-bold text-slate-500 mt-2">المنصب / التخصص</label>
              <input value={position} onChange={e => setPosition(e.target.value)} placeholder="مثال: كاميرات مراقبة" list="position-options"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
              <datalist id="position-options">
                <option value="إداري كوادر" />
                <option value="تقني" />
                <option value="مهندس" />
                <option value="مصمم" />
              </datalist>
              <label className="text-xs font-bold text-slate-500 mt-2">العنوان الوظيفي</label>
              <input value={jobTitle} onChange={e => setJobTitle(e.target.value)} placeholder="مثال: فني أول"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
            </div>
          )}

          {step === 3 && (
            <div className="flex flex-col gap-3">
              <label className="text-xs font-bold text-slate-500">الراتب الشهري (د.ع)</label>
              <input type="number" value={salary} onChange={e => setSalary(e.target.value)} placeholder="0"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
            </div>
          )}

          {step === 4 && (
            <div className="flex flex-col gap-3">
              <p className="text-xs text-slate-400">حدد دوام الموظف الفعلي من والى — يدعم الدوام الطويل (16 ساعة مثلاً) بدون التقيد بـ"صباحي/مسائي" فقط.</p>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-bold text-slate-500">من الساعة</label>
                  <input type="time" value={shiftStart} onChange={e => setShiftStart(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
                </div>
                <div>
                  <label className="text-xs font-bold text-slate-500">إلى الساعة</label>
                  <input type="time" value={shiftEnd} onChange={e => setShiftEnd(e.target.value)}
                    className="mt-1 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
                </div>
              </div>
              {hours !== null && (
                <div className="rounded-xl bg-blue-50 p-3 text-center text-sm font-bold text-brand-700">
                  إجمالي ساعات الدوام: {hours} ساعة {hours >= 12 && '⚠️ دوام طويل'}
                </div>
              )}
            </div>
          )}

          {step === 5 && (
            <div className="flex flex-col gap-3">
              {isDecor ? (
                <div className="rounded-xl bg-blue-50 p-4 text-sm text-brand-700">
                  موظفو شعبة الديكور يُمنحون تلقائياً دور "فني" العادي (بدون أي صلاحيات إدارية) —
                  يستلمون فقط المهام التي يوجّهها لهم منسق الحجوزات، بالضبط متل فنيي الشعبة الهندسية.
                </div>
              ) : (
                <>
                  <label className="text-xs font-bold text-slate-500">الدور الوظيفي بالنظام</label>
                  <div className="grid grid-cols-2 gap-2">
                    {creatableRoles.map(key => {
                      const label = roleLabels[key]
                      return (
                        <button key={key} type="button" onClick={() => setRole(key)}
                          className={`rounded-xl border px-3 py-2.5 text-xs font-bold transition-all ${
                            role === key ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-500 hover:bg-slate-50'
                          }`}>
                          {label}
                        </button>
                      )
                    })}
                  </div>
                </>
              )}
            </div>
          )}

          {step === 6 && (
            <div className="flex flex-col gap-3">
              <label className="text-xs font-bold text-slate-500">اسم المستخدم *</label>
              <input value={username} onChange={e => setUsername(e.target.value)} placeholder="username"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
              <label className="text-xs font-bold text-slate-500 mt-2">كلمة المرور *</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="4 أحرف على الأقل"
                className="rounded-xl border border-slate-200 px-4 py-3 text-sm outline-none focus:border-brand-500" />
            </div>
          )}

          {step === 7 && (
            <div className="flex flex-col gap-3">
              <p className="text-xs text-slate-400">حددنا صلاحيات مقترحة حسب الدور — عدّلها حسب الحاجة قبل الحفظ النهائي.</p>
              <div className="flex max-h-64 flex-col gap-1.5 overflow-y-auto">
                {allPermissions.map(p => (
                  <label key={p.id} className="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
                    <input type="checkbox" checked={selectedPermIds.includes(p.id)} onChange={() => togglePerm(p.id)} />
                    {p.label}
                  </label>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className="flex items-center justify-between border-t border-slate-100 p-4">
          <button onClick={step === 0 ? onClose : back}
            className="rounded-lg px-4 py-2 text-sm font-bold text-slate-500 hover:bg-slate-100">
            {step === 0 ? 'إلغاء' : 'رجوع'}
          </button>

          {step < 6 && (
            <button onClick={next} disabled={!canProceed()}
              className="rounded-lg bg-brand-500 px-6 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-40">
              التالي
            </button>
          )}
          {step === 6 && (
            <button onClick={handleCreate} disabled={!canProceed() || submitting}
              className="rounded-lg bg-brand-500 px-6 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-40">
              {submitting ? 'جاري الإنشاء...' : 'إنشاء الموظف والمتابعة للصلاحيات'}
            </button>
          )}
          {step === 7 && (
            <button onClick={handleFinish} disabled={submitting}
              className="rounded-lg bg-emerald-500 px-6 py-2 text-sm font-bold text-white hover:bg-emerald-600 disabled:opacity-40">
              {submitting ? 'جاري الحفظ...' : 'حفظ الصلاحيات وإنهاء'}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
