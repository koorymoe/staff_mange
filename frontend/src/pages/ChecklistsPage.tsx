import { useEffect, useRef, useState } from 'react'
import { api, type ProjectChecklist, fileUrl } from '../api'

function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function printBlankChecklist(title: string) {
  const win = window.open('', '_blank')
  if (!win) return
  win.document.write(`
    <html dir="rtl"><head><meta charset="utf-8"><title>كشف - ${title}</title>
    <style>
      body { font-family: sans-serif; padding: 32px; }
      h2 { margin-bottom: 4px; }
      .row { display: flex; gap: 24px; margin: 12px 0; }
      .field { flex: 1; border-bottom: 1px solid #333; padding: 6px 4px; min-height: 28px; }
      .label { font-size: 12px; color: #555; margin-bottom: 2px; }
      .lines .field { min-height: 40px; }
    </style></head><body>
    <h2>كشف: ${title}</h2>
    <p style="color:#777">التاريخ: ______________ / الموقع: ______________</p>
    <div class="row"><div class="field"><div class="label">اسم المهندس</div></div><div class="field"><div class="label">اسم المشروع</div></div></div>
    ${Array.from({ length: 8 }).map(() => '<div class="row lines"><div class="field"></div></div>').join('')}
    <p style="margin-top:24px; color:#777">توقيع المهندس: ______________</p>
    </body></html>
  `)
  win.document.close()
  win.focus()
  win.print()
}

export default function ChecklistsPage() {
  const [checklists, setChecklists] = useState<ProjectChecklist[]>([])
  const [projects, setProjects] = useState<{ id: string; name: string; code: string }[]>([])
  const [loading, setLoading] = useState(true)
  const [title, setTitle] = useState('')
  const [projectId, setProjectId] = useState('')
  const [creating, setCreating] = useState(false)
  const [uploadingId, setUploadingId] = useState<string | null>(null)
  const fileInputRefs = useRef<Record<string, HTMLInputElement | null>>({})

  const load = () => {
    api.getChecklists().then(setChecklists).finally(() => setLoading(false))
  }
  useEffect(load, [])
  useEffect(() => { api.getProjectsBrief().then(setProjects).catch(() => {}) }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    setCreating(true)
    try {
      await api.createChecklist({ title: title.trim(), projectId: projectId || null })
      setTitle('')
      setProjectId('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إنشاء الكشف')
    } finally {
      setCreating(false)
    }
  }

  const handleUpload = async (id: string, files: FileList | null) => {
    if (!files || files.length === 0) return
    setUploadingId(id)
    try {
      const base64s = await Promise.all(Array.from(files).map(fileToBase64))
      await api.addChecklistPhotos(id, base64s)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفع الصور')
    } finally {
      setUploadingId(null)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الكشوفات</h2>
      <p className="mt-1 text-slate-500">
        سوي كشف جديد، اطبعه فارغ حتى يمليه المهندس بالموقع، وبعدها ارفع صور الفورمة المالية.
      </p>

      <form onSubmit={handleCreate} className="mt-6 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-4">
        <input
          required
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="عنوان الكشف (مثال: كشف تسليم موقع)"
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-2"
        />
        <select value={projectId} onChange={(e) => setProjectId(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
          <option value="">-- مشروع (اختياري) --</option>
          {projects.map((p) => <option key={p.id} value={p.id}>{p.code} - {p.name}</option>)}
        </select>
        <button type="submit" disabled={creating} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
          {creating ? 'جاري الإنشاء...' : '+ كشف جديد'}
        </button>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      <div className="mt-6 flex flex-col gap-3">
        {!loading && checklists.length === 0 && <p className="text-slate-400">لا توجد كشوفات بعد.</p>}
        {checklists.map((c) => (
          <div key={c.id} className="rounded-xl border border-white bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <p className="font-bold text-brand-900">{c.title}</p>
                <p className="text-xs text-slate-400">
                  {c.project ? `مشروع: ${c.project.name} — ` : ''}
                  أنشأه: {c.createdBy?.name || '-'} — {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                </p>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => printBlankChecklist(c.title)}
                  className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-bold text-slate-600 hover:bg-slate-50"
                >
                  🖨️ طباعة فورمة فارغة
                </button>
                <button
                  onClick={() => fileInputRefs.current[c.id]?.click()}
                  disabled={uploadingId === c.id}
                  className="rounded-lg bg-brand-50 px-3 py-1.5 text-sm font-bold text-brand-700 hover:bg-brand-100 disabled:opacity-50"
                >
                  {uploadingId === c.id ? 'جاري الرفع...' : '📤 رفع صور الفورمة المملوءة'}
                </button>
                <input
                  ref={(el) => { fileInputRefs.current[c.id] = el }}
                  type="file" accept="image/*" multiple className="hidden"
                  onChange={(e) => handleUpload(c.id, e.target.files)}
                />
              </div>
            </div>
            {c.photoUrls.length > 0 && (
              <div className="mt-3 flex flex-wrap gap-2">
                {c.photoUrls.map((url, i) => (
                  <img key={i} src={fileUrl(url)} className="h-24 w-24 rounded-lg border object-cover" />
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
