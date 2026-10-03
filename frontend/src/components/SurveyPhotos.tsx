import { useEffect, useState } from 'react'
import { api, type SurveyPhoto, type SurveyPhotoOwner } from '../api'
import { useSession } from '../session'

// صور الكشف — للحجز والمشروع. العرض والتنزيل يقرّرهن الخادم بكل طلب
// (المراقب، المحاسب، المدير، إداري الحجوزات، والكادر الي رفعها)، فلو
// رجع ٤٠٣ المكوّن يختفي بهدوء بدل ما يطلّع خطأ لموظف ما يخصّه.
export default function SurveyPhotos({ owner, canUpload = false, title = '📷 صور الكشف' }: {
  owner: SurveyPhotoOwner
  canUpload?: boolean
  title?: string
}) {
  const { employee } = useSession()
  const ownerKey = owner.bookingId ? `b:${owner.bookingId}` : `p:${owner.projectId}`
  const [photos, setPhotos] = useState<SurveyPhoto[] | null>(null)
  const [thumbs, setThumbs] = useState<Record<string, string>>({})
  const [hidden, setHidden] = useState(false)
  const [busy, setBusy] = useState(false)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    api.getSurveyPhotos(owner)
      .then((list) => { if (alive) { setPhotos(list); setHidden(false) } })
      .catch(() => { if (alive) setHidden(true) })
    return () => { alive = false }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ownerKey, reload])

  useEffect(() => {
    if (!photos?.length) return
    let alive = true
    const made: string[] = []
    Promise.all(photos.map((p) =>
      api.surveyPhotoBlob(p.id).then((b) => {
        const u = URL.createObjectURL(b)
        made.push(u)
        return [p.id, u] as const
      }).catch(() => null),
    )).then((pairs) => {
      if (alive) setThumbs(Object.fromEntries(pairs.filter((x): x is readonly [string, string] => !!x)))
    })
    return () => {
      alive = false
      made.forEach((u) => URL.revokeObjectURL(u))
    }
  }, [photos])

  const download = async (p: SurveyPhoto) => {
    try {
      const blob = await api.surveyPhotoBlob(p.id, true)
      const u = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = u
      a.download = p.fileName || `كشف-${p.id}.jpg`
      document.body.appendChild(a)
      a.click()
      a.remove()
      setTimeout(() => URL.revokeObjectURL(u), 1000)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تنزيل الصورة')
    }
  }

  const downloadAll = async () => {
    for (const p of photos || []) await download(p)
  }

  const upload = async (files: FileList | null) => {
    if (!files?.length) return
    setBusy(true)
    const errors: string[] = []
    for (const f of Array.from(files)) {
      try { await api.uploadSurveyPhoto(owner, f) } catch (e) {
        errors.push(`${f.name}: ${e instanceof Error ? e.message : 'تعذر الرفع'}`)
      }
    }
    setBusy(false)
    setReload((n) => n + 1)
    if (errors.length) alert(errors.join('\n'))
  }

  const remove = async (p: SurveyPhoto) => {
    if (!window.confirm('حذف هاي الصورة من الكشف؟')) return
    try {
      await api.deleteSurveyPhoto(p.id)
      setReload((n) => n + 1)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حذف الصورة')
    }
  }

  if (hidden) return null
  const canDelete = (p: SurveyPhoto) => p.uploadedById === employee?.id || employee?.role === 'ADMIN'

  return (
    <div className="mt-3 rounded-xl border border-slate-200 bg-slate-50 p-3">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-bold text-slate-700">{title} {photos ? `(${photos.length})` : ''}</p>
        <div className="flex flex-wrap gap-2">
          {!!photos?.length && (
            <button type="button" onClick={downloadAll}
              className="rounded-lg bg-white px-2.5 py-1 text-xs font-bold text-slate-700 ring-1 ring-slate-200 hover:bg-slate-100">
              ⬇ تنزيل الكل
            </button>
          )}
          {canUpload && (
            <label className={`cursor-pointer rounded-lg bg-[var(--color-brand-600,#2563eb)] px-2.5 py-1 text-xs font-bold text-white ${busy ? 'opacity-60' : ''}`}>
              {busy ? 'جاري الرفع…' : '＋ أرفق صور'}
              <input type="file" accept="image/*" multiple className="hidden" disabled={busy}
                onChange={(e) => { upload(e.target.files); e.target.value = '' }} />
            </label>
          )}
        </div>
      </div>
      {photos === null ? (
        <p className="text-xs text-slate-400">جاري التحميل…</p>
      ) : photos.length === 0 ? (
        <p className="text-xs text-slate-400">ماكو صور مرفقة بعد.</p>
      ) : (
        <div className="grid grid-cols-3 gap-2 sm:grid-cols-4">
          {photos.map((p) => (
            <div key={p.id} className="overflow-hidden rounded-lg bg-white ring-1 ring-slate-200">
              {thumbs[p.id]
                ? <a href={thumbs[p.id]} target="_blank" rel="noreferrer"><img src={thumbs[p.id]} alt={p.fileName} className="h-24 w-full object-cover" /></a>
                : <div className="h-24 w-full animate-pulse bg-slate-100" />}
              <div className="flex items-center justify-between gap-1 px-1.5 py-1">
                <span className="truncate text-[10px] text-slate-500" title={p.uploadedByName}>{p.uploadedByName}</span>
                <span className="flex gap-1">
                  <button type="button" onClick={() => download(p)} title="تنزيل" className="text-xs text-slate-600 hover:text-slate-900">⬇</button>
                  {canDelete(p) && (
                    <button type="button" onClick={() => remove(p)} title="حذف" className="text-xs text-red-500 hover:text-red-700">✕</button>
                  )}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
