import { useEffect, useRef, useState } from 'react'
import { api, type TechShowcaseItem, fileUrl } from '../api'
import { useSession } from '../session'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

export default function TechShowcasePage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || permissions.includes('content_technician')
  const [items, setItems] = useState<TechShowcaseItem[]>([])
  const [loading, setLoading] = useState(true)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [creating, setCreating] = useState(false)
  const [uploadingId, setUploadingId] = useState<string | null>(null)
  const fileInputRefs = useRef<Record<string, HTMLInputElement | null>>({})

  const load = () => {
    api.getTechShowcase().then(setItems).finally(() => setLoading(false))
  }
  useEffect(load, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    setCreating(true)
    try {
      await api.createTechShowcaseItem({ title: title.trim(), description: description.trim() || undefined })
      setTitle('')
      setDescription('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة العمل')
    } finally {
      setCreating(false)
    }
  }

  const handleUpload = async (id: string, files: FileList | null) => {
    if (!files || files.length === 0) return
    setUploadingId(id)
    try {
      const base64s = await Promise.all(Array.from(files).map(fileToBase64))
      await api.addTechShowcaseMedia(id, base64s)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفع الوسائط')
    } finally {
      setUploadingId(null)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">معرض أعمال التقنيين</h2>
      <p className="mt-1 text-slate-500">
        نماذج أعمال وأفكار وتصاميم جديدة يعرضها الفنيون — إلهام لكل الفريق.
      </p>

      {canAdd && (
        <form onSubmit={handleCreate} className="mt-6 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-3">
          <input
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="عنوان العمل (مثال: تركيب جديد لمنظومة شمسية)"
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <input
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="وصف مختصر (اختياري)"
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <button type="submit" disabled={creating} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {creating ? 'جاري الإضافة...' : '+ إضافة عمل جديد'}
          </button>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {!loading && items.length === 0 && <p className="text-slate-400">لا توجد أعمال بعد.</p>}
        {items.map((it) => (
          <div key={it.id} className="rounded-xl border border-white bg-white p-4 shadow-sm">
            <p className="font-bold text-brand-900">{it.title}</p>
            {it.description && <p className="mt-1 text-sm text-slate-500">{it.description}</p>}
            <p className="mt-1 text-xs text-slate-400">{it.employee?.name || '-'} — {new Date(it.createdAt).toLocaleDateString('ar-IQ')}</p>
            {it.mediaUrls.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {it.mediaUrls.map((url, i) => (
                  <img key={i} src={fileUrl(url)} className="h-20 w-20 rounded-lg border object-cover" />
                ))}
              </div>
            )}
            {canAdd && (
              <>
                <button
                  onClick={() => fileInputRefs.current[it.id]?.click()}
                  disabled={uploadingId === it.id}
                  className="mt-3 rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700 hover:bg-brand-100 disabled:opacity-50"
                >
                  {uploadingId === it.id ? 'جاري الرفع...' : '📤 رفع صور/فيديو'}
                </button>
                <input
                  ref={(el) => { fileInputRefs.current[it.id] = el }}
                  type="file" accept="image/*,video/*" multiple className="hidden"
                  onChange={(e) => handleUpload(it.id, e.target.files)}
                />
              </>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
