import { useEffect, useState } from 'react'
import { api, type Announcement } from '../api'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'

/**
 * إدارة الإعلانات — المالك ومدير النظام حصراً.
 *
 * كل إعلان شغّال يمر بالشريط المتحرك كدام كل الموظفين.
 */
export default function AnnouncementsPage() {
  // كل حفظ بهاي الشاشة يمر من هنا — الفشل ينعرض بدل ما ينبلع
  const guard = useSaveGuard()
  const [items, setItems] = useState<Announcement[]>([])
  const [body, setBody] = useState('')
  const [busy, setBusy] = useState(false)
  const [loading, setLoading] = useState(true)

  const load = () => {
    api.getAnnouncements(true)
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const publish = async () => {
    if (!body.trim()) return
    setBusy(true)
    try {
      await api.createAnnouncement(body.trim())
      setBody('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر النشر')
    } finally {
      setBusy(false)
    }
  }

  const toggle = async (a: Announcement) => {
    if (!(await guard.run('تغيير حالة الإعلان', () => api.setAnnouncementActive(a.id, !a.active)))) return
    load()
  }

  const remove = async (a: Announcement) => {
    if (!confirm('حذف الإعلان نهائياً؟')) return
    if (!(await guard.run('حذف الإعلان', () => api.deleteAnnouncement(a.id)))) return
    load()
  }

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
    <div dir="rtl" className="space-y-6">
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">📢 لوحة الإعلانات</h1>
        <p className="mt-1 text-sm text-blue-200">
          الإعلان الشغّال يمر بشريط متحرك كدام كل الموظفين. الموظف يقدر يخفيه، بس يرجع
          يظهرله بعد ما يسجّل خروج ويرجع يدخل.
        </p>
      </div>

      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <label className="mb-2 block text-sm font-bold text-slate-700">نص الإعلان الجديد</label>
        <textarea
          value={body} onChange={(e) => setBody(e.target.value)} rows={3}
          placeholder="مثال: اجتماع عام يوم الخميس الساعة ٩ صباحاً بمقر الشركة"
          className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
        />
        <button
          onClick={publish} disabled={busy || !body.trim()}
          className="mt-3 rounded-lg px-5 py-2.5 font-medium text-white disabled:opacity-50"
          style={{ backgroundColor: '#1a3a5c' }}
        >
          {busy ? 'جاري النشر...' : '📢 انشر الإعلان'}
        </button>
      </div>

      {loading && <p className="text-center text-slate-400">جاري التحميل...</p>}

      <div className="space-y-3">
        {items.map((a) => (
          <div key={a.id} className="flex flex-wrap items-start justify-between gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="min-w-0">
              <p className={`font-medium ${a.active ? 'text-slate-800' : 'text-slate-400 line-through'}`}>{a.body}</p>
              <p className="mt-1 text-xs text-slate-400">
                {a.createdByName} · {new Date(a.createdAt).toLocaleDateString('ar-IQ')}
              </p>
            </div>
            <div className="flex shrink-0 gap-2">
              <button onClick={() => toggle(a)}
                className={`rounded-lg px-3 py-1.5 text-xs font-bold ${
                  a.active ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'
                }`}>
                {a.active ? '✔ شغّال' : 'موقّف'}
              </button>
              <button onClick={() => remove(a)}
                className="rounded-lg bg-red-50 px-3 py-1.5 text-xs font-bold text-red-700 hover:bg-red-100">
                حذف
              </button>
            </div>
          </div>
        ))}
        {!loading && items.length === 0 && (
          <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">ماكو إعلانات بعد</p>
        )}
      </div>
    </div>
    </>
  )
}
