import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type DesignForm } from '../api'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'

const PRIMARY = '#47528f'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--design-ink)'
const GOLD = '#c97a3a'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-warm-ink)'

function publicUrl(token: string) {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  return `${window.location.origin}${base}/design-forms/public/${token}`
}

// معاينة مصغّرة تطابق شكل صفحة الفورمة الحقيقية الي توصل الزبون (نفس هيدر
// الأماني الكحلي والذهبي واسم الشركة عربي/انكليزي) — حتى يشوف المدير شكل
// كل فورمة بهويتها البصرية من نفس صفحة الإدارة، بدون ما يفتح الرابط العام.
function FormBrandPreview({ name }: { name: string }) {
  return (
    <div style={{ borderRadius: '12px', overflow: 'hidden', border: '1px solid #fbede2', boxShadow: '0 2px 10px rgba(71,82,143,0.10)' }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #2f3868)`,
        padding: '14px 16px', color: '#fff', position: 'relative', overflow: 'hidden',
      }}>
        <div style={{ position: 'absolute', top: 0, right: 0, bottom: 0, width: '6px', background: `linear-gradient(${GOLD}, ${PRIMARY})` }} />
        <div style={{ fontSize: '11px', fontWeight: 700, lineHeight: 1.4 }}>
          شركة الأماني للتجارة العامة والاستثمارات العقارية
        </div>
        <div style={{ fontSize: '8px', fontWeight: 600, color: GOLD_TEXT, marginTop: '2px' }}>
          Al-Amani for General Trading &amp; Real Estate
        </div>
        <div style={{ marginTop: '8px', fontSize: '14px', fontWeight: 700 }}>{name || 'اسم الفورمة'}</div>
      </div>
      <div style={{ background: 'var(--sf-card)', padding: '10px 14px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
        <div style={{ height: '8px', width: '60%', background: 'var(--sf-sunken)', borderRadius: '4px' }} />
        <div style={{ height: '18px', width: '100%', background: 'var(--sf-sunken)', border: '1px solid #eef0f8', borderRadius: '6px' }} />
        <div style={{ marginTop: '4px', height: '20px', width: '100%', background: GOLD, opacity: 0.85, borderRadius: '6px' }} />
      </div>
    </div>
  )
}

export default function DesignFormsListPage() {
  // كل حفظ بهاي الشاشة يمر من هنا — الفشل ينعرض بدل ما ينبلع
  const guard = useSaveGuard()
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
    if (!(await guard.run('حذف النموذج', () => api.deleteDesignForm(id)))) return
    load()
  }

  const handleCopy = async (form: DesignForm) => {
    await navigator.clipboard.writeText(publicUrl(form.publicToken))
    setCopiedId(form.id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
    <div dir="rtl">
      <h2 className="text-2xl font-bold" style={{ color: PRIMARY_TEXT }}>وحدة التصميم — فورمة التصميم</h2>
      <p className="mt-1 text-slate-500">
        كل فورمة مستقلة بأسئلتها الخاصة ورابطها العام — كلهن بنفس الهوية البصرية لشركة الأماني، وترسلها للزبون مباشرة بدون ما يشوف أي شي ثاني من النظام.
      </p>

      <div className="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-[280px_1fr]">
        <FormBrandPreview name={name} />
        <form onSubmit={handleCreate} className="flex flex-col justify-center gap-2 rounded-xl border border-white bg-white p-4 shadow-sm">
          <label className="text-xs font-bold text-slate-500">اسم الفورمة الجديدة — المعاينة تتحدث وأنت تكتب</label>
          <input
            value={name} onChange={(e) => setName(e.target.value)}
            placeholder="مثال: طلب تصميم هوية بصرية"
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <button
            type="submit" disabled={creating}
            style={{ background: PRIMARY }}
            className="mt-1 rounded-lg px-5 py-2 text-sm font-bold text-white disabled:opacity-50"
          >
            {creating ? 'جاري الإنشاء...' : '+ إنشاء الفورمة'}
          </button>
        </form>
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {!loading && forms.length === 0 && (
        <div className="mt-6 rounded-xl border border-white bg-white p-8 text-center shadow-sm">
          <p className="text-slate-400">ما عندك أي فورمة بعد — سوّي فورمة جديدة وابدأ تضيفلها الأسئلة.</p>
        </div>
      )}

      <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {forms.map((f) => (
          <div key={f.id} className="flex flex-col gap-3 rounded-xl border border-white bg-white p-3 shadow-sm">
            <FormBrandPreview name={f.name} />
            <p className="text-xs text-slate-400">أُنشئت: {new Date(f.createdAt).toLocaleDateString('ar-IQ')}</p>
            <div className="flex flex-wrap gap-2">
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
    </>
  )
}
