import { useEffect, useState } from 'react'
import { api, type ProjectWorkType } from '../api'

// إعدادات وحدة إدارة المشاريع — أنواع الأعمال ("نوع العمل") صارت قابلة
// للإضافة والحذف براحة المدير، بدل قائمة ثابتة بالكود. أي نوع تضيفه هنا
// يظهر فوراً بقائمة "نوع العمل" عند إنشاء/تعديل/فلترة المشاريع.
export default function ProjectWorkTypesSettingsPage() {
  const [types, setTypes] = useState<ProjectWorkType[]>([])
  const [loading, setLoading] = useState(true)
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = () => { api.getProjectWorkTypes().then(setTypes).finally(() => setLoading(false)) }
  useEffect(load, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setSaving(true)
    setError(null)
    try {
      await api.createProjectWorkType(name.trim())
      setName('')
      load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'تعذر إضافة نوع العمل')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('حذف نوع العمل هذا؟ المشاريع الي مستخدمينه بيهه ما تتأثر، بس النوع ما يضل يظهر بالقائمة عند إنشاء مشروع جديد.')) return
    await api.deleteProjectWorkType(id)
    load()
  }

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">إعدادات إدارة المشاريع — أنواع الأعمال</h2>
      <p className="mt-1 text-slate-500">
        أضف أو احذف أنواع الخدمات التي تظهر بقائمة "نوع العمل" عند إنشاء أو تعديل أي مشروع.
      </p>

      <form onSubmit={handleAdd} className="mt-5 flex flex-wrap gap-2 rounded-xl border border-white bg-white p-4 shadow-sm">
        <input
          value={name} onChange={(e) => setName(e.target.value)}
          placeholder="اسم نوع العمل الجديد (مثال: أنظمة الري الذكي)"
          className="min-w-[260px] flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
        />
        <button
          type="submit" disabled={saving}
          className="rounded-lg bg-brand-600 px-5 py-2 text-sm font-bold text-white hover:bg-brand-700 disabled:opacity-50"
        >
          {saving ? 'جاري الإضافة...' : '+ إضافة نوع عمل'}
        </button>
      </form>
      {error && <p className="mt-2 text-sm font-bold text-red-600">{error}</p>}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="mt-6 flex flex-wrap gap-2">
          {types.map((t) => (
            <div key={t.id} className="flex items-center gap-2 rounded-full border border-brand-100 bg-brand-50 px-4 py-2 text-sm font-bold text-brand-800">
              {t.name}
              <button
                onClick={() => handleDelete(t.id)}
                className="text-red-500 hover:text-red-700"
                title="حذف"
              >
                ✕
              </button>
            </div>
          ))}
          {types.length === 0 && <p className="text-slate-400">ما اكو أنواع أعمال مضافة بعد.</p>}
        </div>
      )}
    </div>
  )
}
