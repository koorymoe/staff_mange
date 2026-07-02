import { useRef, useState } from 'react'
import { api } from '../../api'
import { useSession } from '../../session'
import { useNavigate } from 'react-router-dom'

const IRAQI_GOVERNORATES = [
  'بغداد', 'البصرة', 'نينوى', 'أربيل', 'النجف', 'كربلاء', 'الأنبار', 'ذي قار',
  'ديالى', 'كركوك', 'بابل', 'واسط', 'صلاح الدين', 'القادسية', 'ميسان',
  'المثنى', 'دهوك', 'السليمانية',
]

type PurchaseType = 'DEVICE_SIM' | 'DEVICE_ONLY'
type Step = 'select_type' | 'form'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

export default function GpsPurchase() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>('select_type')
  const [purchaseType, setPurchaseType] = useState<PurchaseType>('DEVICE_SIM')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const [form, setForm] = useState({
    fullName: '', fatherName: '', grandfatherName: '', phone: '', governorate: '', address: '',
    subscriptionType: '' as '' | 'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY',
    gpsNumber: '', residenceCardNumber: '',
  })

  const [files, setFiles] = useState<Record<string, string>>({})

  const handleFile = async (key: string, file: File | null) => {
    if (!file) return
    const base64 = await fileToBase64(file)
    setFiles(prev => ({ ...prev, [key]: base64 }))
  }

  const handleSubmit = async () => {
    setError('')
    if (form.phone.length < 10) { setError('رقم الهاتف غير صحيح'); return }
    if (purchaseType === 'DEVICE_SIM' && !form.subscriptionType) { setError('يرجى اختيار نوع الاشتراك'); return }
    if (purchaseType === 'DEVICE_SIM' && (!files.idFront || !files.idBack || !files.resFront || !files.resBack)) {
      setError('يرجى رفع جميع صور الوثائق المطلوبة'); return
    }
    if (!files.invoice) { setError('يرجى رفع صورة الفاتورة'); return }
    if (!employee) return

    setLoading(true)
    try {
      const customer = await api.createGpsCustomer({
        fullName: form.fullName, fatherName: form.fatherName, grandfatherName: form.grandfatherName,
        phone: form.phone, governorate: form.governorate, address: form.address,
        idCardFrontUrl: files.idFront || null, idCardBackUrl: files.idBack || null,
        residenceCardFrontUrl: files.resFront || null, residenceCardBackUrl: files.resBack || null,
      })

      await api.createGpsDevice({
        customerId: customer.id, employeeId: employee.id,
        purchaseType, subscriptionType: form.subscriptionType || 'THREE_MONTHS',
        invoicePhotoUrl: files.invoice, gpsNumber: form.gpsNumber, residenceCardNumber: form.residenceCardNumber,
      })

      setSuccess(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ، يرجى المحاولة مجدداً')
    } finally {
      setLoading(false)
    }
  }

  const resetForm = () => {
    setSuccess(false); setStep('select_type')
    setForm({ fullName: '', fatherName: '', grandfatherName: '', phone: '', governorate: '', address: '', subscriptionType: '', gpsNumber: '', residenceCardNumber: '' })
    setFiles({})
  }

  if (success) {
    return (
      <div className="mx-auto mt-20 max-w-lg rounded-2xl bg-white p-8 text-center shadow-sm">
        <div className="mb-4 text-6xl">✅</div>
        <h2 className="mb-2 text-2xl font-bold text-brand-900">تم إرسال الطلب!</h2>
        <p className="mb-6 text-slate-500">تم إرسال الطلب للإداري للمراجعة والتفعيل</p>
        <div className="flex justify-center gap-3">
          <button onClick={() => navigate('/')} className="rounded-lg bg-slate-100 px-5 py-2.5 text-sm font-bold text-slate-600 hover:bg-slate-200">الرئيسية</button>
          <button onClick={resetForm} className="rounded-lg bg-brand-500 px-5 py-2.5 text-sm font-bold text-white hover:bg-brand-600">طلب جديد</button>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-3xl">
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلب شراء جهاز GPS</h2>

      {step === 'select_type' && (
        <div className="grid grid-cols-2 gap-4">
          {[
            { type: 'DEVICE_SIM' as PurchaseType, title: 'جهاز + SIM كارد', desc: 'جهاز GPS مع شريحة اتصال واشتراك', icon: '📡' },
            { type: 'DEVICE_ONLY' as PurchaseType, title: 'جهاز فقط', desc: 'جهاز GPS بدون شريحة (شريحة الزبون)', icon: '📦' },
          ].map(opt => (
            <button key={opt.type} onClick={() => { setPurchaseType(opt.type); setStep('form') }}
              className="rounded-2xl border-2 border-transparent bg-white p-6 text-right shadow-sm transition-all hover:border-brand-200 hover:shadow-md">
              <div className="mb-3 text-4xl">{opt.icon}</div>
              <h3 className="mb-1 text-lg font-bold text-brand-900">{opt.title}</h3>
              <p className="text-sm text-slate-400">{opt.desc}</p>
            </button>
          ))}
        </div>
      )}

      {step === 'form' && (
        <div className="flex flex-col gap-5">
          <button onClick={() => setStep('select_type')} className="w-fit text-sm text-brand-600 hover:underline">← رجوع</button>

          <div className="rounded-2xl bg-white p-5 shadow-sm">
            <h3 className="mb-4 font-bold text-brand-900">معلومات الزبون</h3>
            <div className="grid grid-cols-3 gap-3">
              <input placeholder="الاسم الأول *" value={form.fullName} onChange={e => setForm(p => ({ ...p, fullName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
              <input placeholder="اسم الأب *" value={form.fatherName} onChange={e => setForm(p => ({ ...p, fatherName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
              <input placeholder="اسم الجد *" value={form.grandfatherName} onChange={e => setForm(p => ({ ...p, grandfatherName: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
            </div>
            <div className="mt-3 grid grid-cols-2 gap-3">
              <input placeholder="رقم الهاتف *" value={form.phone} onChange={e => setForm(p => ({ ...p, phone: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
              <select value={form.governorate} onChange={e => setForm(p => ({ ...p, governorate: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm">
                <option value="">اختر المحافظة *</option>
                {IRAQI_GOVERNORATES.map(g => <option key={g} value={g}>{g}</option>)}
              </select>
            </div>
            <input placeholder="العنوان (المنطقة / الشارع) *" value={form.address} onChange={e => setForm(p => ({ ...p, address: e.target.value }))} className="mt-3 w-full rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
            <div className="mt-3 grid grid-cols-2 gap-3">
              <input placeholder="رقم الجهاز (GPS/IMEI) *" value={form.gpsNumber} onChange={e => setForm(p => ({ ...p, gpsNumber: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
              <input placeholder="رقم بطاقة السكن *" value={form.residenceCardNumber} onChange={e => setForm(p => ({ ...p, residenceCardNumber: e.target.value }))} className="rounded-lg border border-slate-200 px-3 py-2.5 text-sm" />
            </div>
          </div>

          {purchaseType === 'DEVICE_SIM' && (
            <div className="rounded-2xl bg-white p-5 shadow-sm">
              <h3 className="mb-4 font-bold text-brand-900">نوع الاشتراك</h3>
              <div className="grid grid-cols-3 gap-3">
                {([['THREE_MONTHS', '3 أشهر'], ['SIX_MONTHS', '6 أشهر'], ['YEARLY', 'سنوي']] as const).map(([val, label]) => (
                  <button key={val} type="button" onClick={() => setForm(p => ({ ...p, subscriptionType: val }))}
                    className={`rounded-xl border-2 p-3 text-center font-bold transition-all ${form.subscriptionType === val ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-500'}`}>
                    {label}
                  </button>
                ))}
              </div>
            </div>
          )}

          <div className="rounded-2xl bg-white p-5 shadow-sm">
            <h3 className="mb-4 font-bold text-brand-900">المستندات المطلوبة</h3>
            {purchaseType === 'DEVICE_SIM' && (
              <div className="mb-4 grid grid-cols-2 gap-3">
                <FileUpload label="الهوية - أمامي" preview={files.idFront} onChange={f => handleFile('idFront', f)} />
                <FileUpload label="الهوية - خلفي" preview={files.idBack} onChange={f => handleFile('idBack', f)} />
                <FileUpload label="بطاقة السكن - أمامي" preview={files.resFront} onChange={f => handleFile('resFront', f)} />
                <FileUpload label="بطاقة السكن - خلفي" preview={files.resBack} onChange={f => handleFile('resBack', f)} />
              </div>
            )}
            <FileUpload label="صورة الفاتورة *" preview={files.invoice} onChange={f => handleFile('invoice', f)} />
          </div>

          {error && <p className="rounded-lg bg-red-50 p-3 text-center text-sm text-red-600">{error}</p>}
          <button onClick={handleSubmit} disabled={loading}
            className="rounded-xl bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {loading ? 'جاري الإرسال...' : 'إرسال الطلب للإداري ←'}
          </button>
        </div>
      )}
    </div>
  )
}

function FileUpload({ label, preview, onChange }: { label: string; preview?: string; onChange: (f: File | null) => void }) {
  const inputRef = useRef<HTMLInputElement>(null)
  return (
    <div>
      <label className="mb-1 block text-xs font-medium text-slate-500">{label}</label>
      <div onClick={() => inputRef.current?.click()}
        className="flex min-h-[110px] cursor-pointer items-center justify-center rounded-xl border-2 border-dashed border-slate-200 p-3 text-center transition-all hover:border-brand-300"
        style={preview ? { borderColor: '#2c5aad', borderStyle: 'solid' } : undefined}>
        {preview ? <img src={preview} alt={label} className="max-h-24 rounded-lg object-contain" /> :
          <div className="text-slate-400"><div className="mb-1 text-2xl">📷</div><p className="text-xs">اضغط لرفع الصورة</p></div>}
      </div>
      <input ref={inputRef} type="file" accept="image/*" className="hidden" onChange={e => onChange(e.target.files?.[0] || null)} />
    </div>
  )
}
