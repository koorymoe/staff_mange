import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../../api'
import { useSession } from '../../session'

const IRAQI_GOVERNORATES = [
  'بغداد', 'البصرة', 'نينوى', 'أربيل', 'النجف', 'كربلاء', 'الأنبار', 'ذي قار',
  'ديالى', 'كركوك', 'بابل', 'واسط', 'صلاح الدين', 'القادسية', 'ميسان',
  'المثنى', 'دهوك', 'السليمانية',
]

const PROBLEM_TYPES = [
  ['تأخير في التفعيل', 'delay_activation'],
  ['عطل في الجهاز', 'device_malfunction'],
  ['كسر في الجهاز', 'device_broken'],
  ['انقطاع الإشارة', 'signal_loss'],
  ['ضياع الجهاز', 'device_lost'],
  ['أخرى', 'other'],
] as const

export default function GpsMaintenanceRequestPage() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({
    fullName: '', fatherName: '', grandfatherName: '', phone: '', governorate: '', address: '',
    problemType: PROBLEM_TYPES[0][0] as string, description: '',
  })

  const handleSubmit = async () => {
    if (!employee) return
    if (!form.fullName || !form.phone || !form.governorate || !form.address) {
      setError('يرجى تعبئة كل الحقول المطلوبة'); return
    }
    setLoading(true); setError('')
    try {
      const customer = await api.createGpsCustomer({
        fullName: form.fullName, fatherName: form.fatherName, grandfatherName: form.grandfatherName,
        phone: form.phone, governorate: form.governorate, address: form.address,
      })
      await api.createGpsMaintenance({
        customerId: customer.id, employeeId: employee.id,
        problemDescription: `[${form.problemType}] ${form.description}`.trim(),
      })
      setSuccess(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setSuccess(false)
    setForm({ fullName: '', fatherName: '', grandfatherName: '', phone: '', governorate: '', address: '', problemType: PROBLEM_TYPES[0][0], description: '' })
  }

  if (success) {
    return (
      <div className="mx-auto mt-20 max-w-lg rounded-2xl bg-white p-8 text-center shadow-sm">
        <div className="mb-4 text-6xl">✅</div>
        <h2 className="mb-2 text-2xl font-bold text-brand-900">تم إرسال طلب الصيانة!</h2>
        <p className="mb-6 text-slate-500">سيتم مراجعته من قبل الموظف الإداري</p>
        <div className="flex justify-center gap-3">
          <button onClick={() => navigate('/')} className="rounded-lg bg-slate-100 px-5 py-2.5 text-sm font-bold text-slate-600 hover:bg-slate-200">الرئيسية</button>
          <button onClick={resetForm} className="rounded-lg bg-brand-500 px-5 py-2.5 text-sm font-bold text-white hover:bg-brand-600">طلب جديد</button>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-2xl">
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلب صيانة 🔧</h2>

      <div className="mb-4 rounded-2xl bg-white p-5 shadow-sm">
        <h3 className="mb-4 font-bold text-brand-900">معلومات الزبون</h3>
        <div className="grid grid-cols-3 gap-3">
          <input placeholder="الاسم الأول *" value={form.fullName} onChange={e => setForm(p => ({ ...p, fullName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
          <input placeholder="اسم الأب" value={form.fatherName} onChange={e => setForm(p => ({ ...p, fatherName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
          <input placeholder="اسم الجد" value={form.grandfatherName} onChange={e => setForm(p => ({ ...p, grandfatherName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
        </div>
        <div className="mt-3 grid grid-cols-2 gap-3">
          <input placeholder="رقم الهاتف *" value={form.phone} onChange={e => setForm(p => ({ ...p, phone: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
          <select value={form.governorate} onChange={e => setForm(p => ({ ...p, governorate: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm">
            <option value="">اختر المحافظة *</option>
            {IRAQI_GOVERNORATES.map(g => <option key={g} value={g}>{g}</option>)}
          </select>
        </div>
        <input placeholder="العنوان (المنطقة/الشارع) *" value={form.address} onChange={e => setForm(p => ({ ...p, address: e.target.value }))} className="mt-3 w-full rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
      </div>

      <div className="mb-4 rounded-2xl bg-white p-5 shadow-sm">
        <h3 className="mb-4 font-bold text-brand-900">وصف المشكلة</h3>
        <label className="mb-1 block text-xs font-bold text-slate-500">نوع المشكلة *</label>
        <select value={form.problemType} onChange={e => setForm(p => ({ ...p, problemType: e.target.value }))} className="mb-3 w-full rounded-lg border border-slate-200 px-3 py-2.5 text-sm">
          {PROBLEM_TYPES.map(([label]) => <option key={label} value={label}>{label}</option>)}
        </select>
        <label className="mb-1 block text-xs font-bold text-slate-500">تفاصيل إضافية (اختياري)</label>
        <textarea rows={4} value={form.description} onChange={e => setForm(p => ({ ...p, description: e.target.value }))} placeholder="أي تفاصيل إضافية..." className="w-full rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
      </div>

      {error && <p className="mb-4 rounded-lg bg-red-50 p-3 text-center text-sm text-red-600">{error}</p>}
      <button onClick={handleSubmit} disabled={loading} className="w-full rounded-xl bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
        {loading ? 'جاري الإرسال...' : 'إرسال طلب الصيانة ←'}
      </button>
    </div>
  )
}
