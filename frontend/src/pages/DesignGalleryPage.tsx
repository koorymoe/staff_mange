import { useEffect, useMemo, useState } from 'react'
import { api, fileUrl, ensureFileToken, type DesignAsset } from '../api'
import PageHeader from '../components/PageHeader'
import StatTile from '../components/StatTile'
import EmptyState from '../components/EmptyState'
import SearchBar from '../components/SearchBar'
import Pager from '../components/Pager'
import SaveError from '../components/SaveError'
import { useSaveGuard } from '../useSaveGuard'
import { matches } from '../utils/search'

/**
 * ═══ معرض التصاميم ═══
 *
 * «معرض تصاميم مستقل» — قراره. المصممة ترفع شغلها ويبقى محفوظاً
 * بمكان واحد، **بلا ربط بحجز** لأنها ما تشوف الحجوزات.
 *
 * ⚠️ **الملف يمرّ على `POST /api/files` الموجود** — المعرض يخزّن
 * المفتاح الراجع منه بس. صفر منطق تخزين جديد.
 *
 * ⚠️ **الأرشفة مو حذف**: التصميم ينزاح من المعرض ويبقى بالقاعدة
 * وبالتخزين، فـ«رجّعه» يرجّعه كامل.
 */
const PER_PAGE_OPTIONS = [12, 24, 48]

export default function DesignGalleryPage() {
  // ⚠️ `null` = لسه ما وصلت، `[]` = وصلت وفاضية.
  const [rows, setRows] = useState<DesignAsset[] | null>(null)
  const [categories, setCategories] = useState<Record<string, string>>({})
  const [showArchived, setShowArchived] = useState(false)
  const [query, setQuery] = useState('')
  const [cat, setCat] = useState('')
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(12)
  const guard = useSaveGuard()

  // نموذج الرفع
  const [open, setOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [newCat, setNewCat] = useState('LOGO')
  const [notes, setNotes] = useState('')
  const [file, setFile] = useState<File | null>(null)

  const load = () => {
    api.getDesignAssets(showArchived).then(setRows).catch(() => setRows([]))
  }
  useEffect(load, [showArchived])
  useEffect(() => {
    // رمز عرض الملفات — بدونه الصور تطلع مكسورة.
    void ensureFileToken()
    api.getDesignCategories().then(setCategories).catch(() => setCategories({}))
  }, [])

  const list = useMemo(() => rows ?? [], [rows])
  const loading = rows === null

  const filtered = useMemo(() => list.filter((a) => {
    if (cat && a.category !== cat) return false
    if (!query.trim()) return true
    return matches([a.title, a.notes, a.uploadedByName], query)
  }), [list, cat, query])

  const start = (page - 1) * perPage
  const paged = filtered.slice(start, start + perPage)

  // هذا الشهر — بلا بيانة «—» لا «٠».
  const thisMonth = useMemo(() => {
    const now = new Date()
    return list.filter((a) => {
      const d = new Date(a.createdAt)
      return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
    }).length
  }, [list])

  const upload = async () => {
    if (!file || !title.trim()) return
    const ok = await guard.run('رفع التصميم', async () => {
      const up = await api.uploadFile(file, 'designs')
      return api.createDesignAsset({
        title: title.trim(),
        category: newCat,
        fileKey: up.url,
        fileType: up.type || undefined,
        notes: notes.trim() || undefined,
      })
    })
    if (ok) {
      setTitle(''); setNotes(''); setFile(null); setOpen(false)
      load()
    }
  }

  const toggleArchive = async (a: DesignAsset) => {
    const archived = !!a.archivedAt
    if (!archived && !confirm(`أرشفة «${a.title}»؟\nينزاح من المعرض ويبقى محفوظاً، وتكدر ترجّعه.`)) return
    const ok = await guard.run(archived ? 'إرجاع التصميم' : 'أرشفة التصميم',
      () => api.setDesignAssetArchived(a.id, !archived))
    if (ok) load()
  }

  const isImage = (a: DesignAsset) => (a.fileType || '').startsWith('image/')

  return (
    <div dir="rtl" className="space-y-4">
      <SaveError message={guard.error} onClose={guard.clear} />
      <PageHeader
        title="🎨 معرض التصاميم"
        subtitle="تصاميم الشركة بمكان واحد — شعارات وبنرات وسوشيال وطباعة. المؤرشف ينزاح من المعرض ويبقى محفوظاً."
      />

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        <StatTile label="تصاميم بالمعرض" icon="🖼️" tone="info" value={list.length} />
        <StatTile label="هذا الشهر" icon="🗓️" tone="default"
          value={thisMonth > 0 ? thisMonth : '—'}
          hint={thisMonth > 0 ? undefined : 'ماكو تصميم انرفع هذا الشهر'} />
        <StatTile label="بالبحث" icon="🔍" tone="default" value={filtered.length} />
      </div>

      <SearchBar value={query} onChange={(v) => { setQuery(v); setPage(1) }}
        placeholder="ابحث بالعنوان أو الملاحظة أو اسم المصمّم...">
        <select value={cat} onChange={(e) => { setCat(e.target.value); setPage(1) }}
          className="rounded-lg border px-2 py-1.5 text-xs"
          style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}>
          <option value="">كل التصنيفات</option>
          {Object.entries(categories).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <label className="flex items-center gap-1.5 text-xs font-bold" style={{ color: 'var(--t-body)' }}>
          <input type="checkbox" checked={showArchived}
            onChange={(e) => { setShowArchived(e.target.checked); setPage(1) }} />
          وريني المؤرشف
        </label>
        <button onClick={() => setOpen((o) => !o)}
          className="rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-bold text-white">
          {open ? '✕ سكّر' : '＋ ارفع تصميماً'}
        </button>
      </SearchBar>

      {open && (
        <div className="rounded-xl border p-4"
          style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--t-muted)' }}>عنوان التصميم *</label>
              <input value={title} onChange={(e) => setTitle(e.target.value)}
                placeholder="مثال: شعار حملة رمضان"
                className="w-full rounded-lg border px-3 py-2 text-sm"
                style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }} />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--t-muted)' }}>التصنيف</label>
              <select value={newCat} onChange={(e) => setNewCat(e.target.value)}
                className="w-full rounded-lg border px-3 py-2 text-sm"
                style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }}>
                {Object.entries(categories).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </div>
            <div className="sm:col-span-2">
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--t-muted)' }}>ملاحظة (اختيارية)</label>
              <input value={notes} onChange={(e) => setNotes(e.target.value)}
                placeholder="شنو المطلوب بالضبط، أو لأي زبون"
                className="w-full rounded-lg border px-3 py-2 text-sm"
                style={{ borderColor: 'var(--bd-line)', backgroundColor: 'var(--sf-card)', color: 'var(--t-body)' }} />
            </div>
            <div className="sm:col-span-2">
              <label className="mb-1 block text-xs font-medium" style={{ color: 'var(--t-muted)' }}>الملف *</label>
              <input type="file" onChange={(e) => setFile(e.target.files?.[0] ?? null)}
                className="w-full text-sm" style={{ color: 'var(--t-body)' }} />
            </div>
          </div>
          <button onClick={upload} disabled={guard.busy || !file || !title.trim()}
            className="mt-3 rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
            {guard.busy ? 'جاري الرفع...' : 'ارفعه للمعرض'}
          </button>
        </div>
      )}

      {loading && <p style={{ color: 'var(--t-faint)' }}>جاري التحميل...</p>}
      {!loading && filtered.length === 0 && (
        <EmptyState icon="🎨" title="ماكو تصاميم"
          reason={list.length === 0
            ? 'المعرض فاضي — ارفع أول تصميم من الزر فوق.'
            : 'ماكو تصميم يطابق البحث أو التصنيف.'} />
      )}

      {!loading && paged.length > 0 && (
        <>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
            {paged.map((a) => (
              <div key={a.id} className="overflow-hidden rounded-xl border"
                style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)', opacity: a.archivedAt ? 0.6 : 1 }}>
                <a href={fileUrl(a.fileKey)} target="_blank" rel="noreferrer"
                  className="block aspect-square overflow-hidden"
                  style={{ backgroundColor: 'var(--sf-sunken)' }}>
                  {isImage(a) ? (
                    <img src={fileUrl(a.fileKey)} alt={a.title} className="h-full w-full object-cover" />
                  ) : (
                    <span className="flex h-full w-full items-center justify-center text-3xl">📄</span>
                  )}
                </a>
                <div className="p-2.5">
                  <p className="truncate text-sm font-bold" style={{ color: 'var(--t-title)' }} title={a.title}>{a.title}</p>
                  <p className="text-[11px]" style={{ color: 'var(--t-faint)' }}>
                    {categories[a.category] || a.category}
                    {a.uploadedByName && <> • {a.uploadedByName}</>}
                  </p>
                  {a.notes && <p className="mt-1 line-clamp-2 text-[11px]" style={{ color: 'var(--t-muted)' }}>{a.notes}</p>}
                  <button onClick={() => toggleArchive(a)} disabled={guard.busy}
                    className={`mt-2 rounded-lg px-2 py-0.5 text-[11px] font-bold ${
                      a.archivedAt ? 'text-emerald-700 hover:bg-emerald-50' : 'text-red-600 hover:bg-red-50'}`}>
                    {a.archivedAt ? '↩︎ رجّعه للمعرض' : '🗄️ أرشفه'}
                  </button>
                </div>
              </div>
            ))}
          </div>
          <Pager page={page} total={filtered.length} perPage={perPage} unit="تصميم"
            onPage={setPage} onPerPage={(n) => { setPerPage(n); setPage(1) }}
            perPageOptions={PER_PAGE_OPTIONS} />
        </>
      )}
    </div>
  )
}
