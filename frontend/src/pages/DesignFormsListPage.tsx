import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type DesignForm } from '../api'

const PRIMARY = '#47528f'
const GOLD = '#c97a3a'

function publicUrl(token: string) {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  return `${window.location.origin}${base}/design-forms/public/${token}`
}

export default function DesignFormsListPage() {
  const navigate = useNavigate()
  const [forms, setForms] = useState<DesignForm[]>([])
  const [loading, setLoading] = useState(true)
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const load = () => { api.getDesignForms().then(setForms).finally(() => setLoading(false)) }
  useEffect(load, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setCreating(true)
    try {
      const form = await api.createDesignForm(name.trim())
      setName('')
      load()
      navigate(`/design-forms/${form.id}`)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إنشاء الفورمة')
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('حذف هذه الفورمة نهائياً؟ راح تنحذف كل أسئلتها وأجوبتها معها.')) return
    await api.deleteDesignForm(id)
    load()
  }

  const handleCopy = async (form: DesignForm) => {
    await navigator.clipboard.writeText(publicUrl(form.publicToken))
    setCopiedId(form.id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold" style={{ color: PRIMARY }}>وحدة التصميم — فورمة التصميم</h2>
      <p className="mt-1 text-slate-500">
        كل فورمة مستقلة بأسئلتها الخاصة ورابطها العام — ترسله للزبون مباشرة بدون ما يشوف أي شي ثاني من النظام.
      </p>

      <form onSubmit={handleCreate} className="mt-5 flex flex-wrap gap-2 rounded-xl border border-white bg-white p-4 shadow-sm">
        <input
          value={name} onChange={(e) => setName(e.target.value)}
          placeholder="اسم الفورمة الجديدة (مثال: طلب تصميم هوية بصرية)"
          className="min-w-[260px] flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />
        <button
          type="submit" disabled={creating}
          style={{ background: PRIMARY }}
          className="rounded-lg px-5 py-2 text-sm font-bold text-white disabled:opacity-50"
        >
          {creating ? 'جاري الإنشاء...' : '+ فورمة جديدة'}
        </button>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {!loading && forms.length === 0 && (
        <div className="mt-6 rounded-xl border border-white bg-white p-8 text-center shadow-sm">
          <p className="text-slate-400">ما عندك أي فورمة بعد — سوّي فورمة جديدة وابدأ تضيفلها الأسئلة.</p>
        </div>
      )}

      <div className="mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
        {forms.map((f) => (
          <div key={f.id} className="rounded-xl border border-white bg-white p-4 shadow-sm">
            <div className="flex items-start justify-between gap-2">
              <div>
                <p className="font-bold" style={{ color: PRIMARY }}>{f.name}</p>
                <p className="mt-0.5 text-xs text-slate-400">أُنشئت: {new Date(f.createdAt).toLocaleDateString('ar-IQ')}</p>
              </div>
              <span className="h-3 w-3 shrink-0 rounded-full" style={{ background: GOLD }} title="هوية الأماني البصرية" />
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <button onClick={() => navigate(`/design-forms/${f.id}`)} className="rounded-lg px-3 py-1.5 text-xs font-bold text-white" style={{ background: PRIMARY }}>
                الأسئلة والترتيب
              </button>
              <button onClick={() => navigate(`/design-forms/${f.id}/submissions`)} className="rounded-lg bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-700 hover:bg-slate-200">
                الأجوبة المستلمة
              </button>
              <button onClick={() => handleCopy(f)} className="rounded-lg px-3 py-1.5 text-xs font-bold text-white" style={{ background: copiedId === f.id ? '#2e7d32' : GOLD }}>
                {copiedId === f.id ? 'تم النسخ ✓' : 'نسخ رابط الزبون'}
              </button>
              <button onClick={() => handleDelete(f.id)} className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100">
                حذف
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
