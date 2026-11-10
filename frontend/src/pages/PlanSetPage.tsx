import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { api, apiRequest, fileUrl } from '../api'
import { useSession } from '../session'
import { buildPlanSetHtml, isImageFile, type PlanPrintQuote } from '../utils/planSetPrint'
import { printProjectQuote } from '../utils/projectQuotePrint'
import type { PlanMeta, PlanSetDetail } from './design/planTypes'

// ═══ 📐 سيت واحد — (ك) 11-10 بطلب دازاي ═══
// «تختار الي تريده وتسوي ابلود وترفع صورة أو ملف للمخطط… ينحفظ بالنظام… تسوي تنزيل
// ملف PDF بالهوية البصرية بي كل التصاميم بالإضافة الى عرض السعر».
export default function PlanSetPage() {
  const { id = '' } = useParams()
  const [params, setParams] = useSearchParams()
  const { employee, permissions } = useSession()
  const [meta, setMeta] = useState<PlanMeta | null>(null)
  const [d, setD] = useState<PlanSetDetail | null>(null)
  const [reload, setReload] = useState(0)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [type, setType] = useState('')
  const [title, setTitle] = useState('')
  const printed = useRef(false)

  useEffect(() => {
    let alive = true
    Promise.all([apiRequest<PlanMeta>('/plan-sets/meta'), apiRequest<PlanSetDetail>(`/plan-sets/${id}`)])
      .then(([m, s]) => { if (alive) { setMeta(m); setD(s) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [id, reload])

  const run = async (fn: () => Promise<unknown>) => {
    setBusy(true); setErr('')
    try { await fn(); setReload((n) => n + 1) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const upload = (files: FileList | File[]) => {
    if (!type) { setErr('اختار نوع المخطط أول'); return }
    void run(async () => {
      for (const file of Array.from(files)) {
        const up = await api.uploadFile(file, 'designs')
        await apiRequest(`/plan-sets/${id}/files`, { method: 'POST', body: JSON.stringify({ typeCode: type, title: title.trim() || file.name, fileKey: up.url, fileType: up.type }) })
      }
      setTitle('')
    })
  }

  // عرض السعر: ينجاب بس للي يكدر يشوف العروض كلها — غيره الخادم يرجّع 403 وتنحسب مخالفة.
  const canQuotes = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
    || permissions.includes('quotation_manage_all') || permissions.includes('quotation_system')
  const download = useCallback(async () => {
    if (!d || !meta) return
    let quote: PlanPrintQuote | null = null
    let quoteNote = ''
    if (!d.projectId) quoteNote = 'السيت مو مربوط بمشروع، فماكو عرض سعر.'
    else if (!canQuotes) quoteNote = 'عرض السعر يحتاج صلاحية العروض — اطلبه من الإدارة.'
    else {
      try {
        const q = await api.getQuotationByProject(d.projectId)
        if (q) quote = { quotationNumber: q.quotationNumber, customerName: q.customerName, items: q.items, grandTotal: q.grandTotal, discountValue: q.discountValue, netTotal: q.netTotal }
        else quoteNote = 'ماكو عرض سعر لهذا المشروع بعد.'
      } catch { quoteNote = 'تعذر جلب عرض السعر.' }
    }
    printProjectQuote(buildPlanSetHtml({
      title: d.title, projectName: d.projectName, groups: meta.groups, files: d.items, quote, quoteNote, date: new Date(),
    }, `${import.meta.env.BASE_URL}apple-touch-icon.png`))
  }, [d, meta, canQuotes])

  // «⬇️ تحميل» من كرت السيت يفتح هنا بـ?pdf=1 — نطبع مرة وحدة بعد ما يوصل السيت.
  useEffect(() => {
    if (params.get('pdf') === '1' && d && meta && !printed.current) {
      printed.current = true
      setParams({}, { replace: true })
      void download()
    }
  }, [params, d, meta, download, setParams])

  if (!d || !meta) return <p className="p-4 text-sm text-slate-500">{err || 'يحمّل…'}</p>
  const byType = new Map<string, typeof d.items>()
  for (const f of d.items) byType.set(f.typeCode, [...(byType.get(f.typeCode) ?? []), f])
  const locked = !!d.archivedAt

  return (
    <div dir="rtl" className="space-y-4">
      <Link to="/plan-sets" className="text-xs text-brand-700 underline">← كل السيتات</Link>
      <div className="flex flex-wrap items-start justify-between gap-3 rounded-2xl border border-slate-200 bg-white p-4">
        <div>
          <h2 className="text-xl font-extrabold text-brand-900">📐 {d.title}</h2>
          <p className="text-xs text-slate-500">{d.projectCode ? `${d.projectCode} — ${d.projectName ?? ''} · ` : ''}{d.items.length} مخطط{d.createdByName ? ` · فتحه ${d.createdByName}` : ''}</p>
          {d.notes && <p className="mt-1 text-sm text-slate-600">{d.notes}</p>}
        </div>
        <div className="flex flex-wrap gap-2">
          <button type="button" disabled={busy || d.items.length === 0} onClick={() => void download()} className="rounded-lg bg-amber-500 px-3 py-1.5 text-sm font-bold text-white disabled:opacity-50">⬇️ تنزيل PDF</button>
          {d.canManage && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-sets/${id}/archive`, { method: 'POST', body: JSON.stringify({ archived: !locked }) }))} className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-bold">{locked ? '↩️ رجّعه' : '🗄️ أرشفة'}</button>}
        </div>
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}

      {!locked && (
        <div className="space-y-2 rounded-2xl border-2 border-dashed border-slate-300 bg-white p-4"
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => { e.preventDefault(); if (e.dataTransfer.files.length) upload(e.dataTransfer.files) }}>
          <p className="font-bold text-slate-800">☁️ رفع مخططات جديدة</p>
          <div className="flex flex-wrap gap-2">
            <select value={type} onChange={(e) => setType(e.target.value)} className="min-w-56 flex-1 rounded-lg border border-slate-200 p-2 text-sm">
              <option value="">اختار نوع المخطط…</option>
              {meta.groups.map((g) => (
                <optgroup key={g.no} label={`${g.no}. ${g.name}`}>
                  {g.types.map((t) => <option key={t.code} value={t.code}>{t.code} {t.name}</option>)}
                </optgroup>
              ))}
            </select>
            <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="اسم الملف (اختياري)" className="flex-1 rounded-lg border border-slate-200 p-2 text-sm" />
            <label className={`cursor-pointer rounded-lg px-4 py-2 text-sm font-bold text-white ${type && !busy ? 'bg-[#0f2040]' : 'pointer-events-none bg-slate-400'}`}>
              📁 اختيار الملفات
              <input type="file" multiple accept="image/png,image/jpeg,image/webp,application/pdf,.dwg" className="hidden" onChange={(e) => { if (e.target.files?.length) upload(e.target.files); e.target.value = '' }} />
            </label>
          </div>
          <p className="text-xs text-slate-500">اسحب الملفات هنا أو اختارها · PDF، DWG، JPG، PNG · لحد ٢٤ ميگا. الصور تطلع صفحات بالـPDF، وPDF وDWG يطلعن بقائمة المرفقات.</p>
        </div>
      )}

      {meta.groups.map((g) => {
        const has = g.types.some((t) => byType.has(t.code))
        return (
          <details key={g.no} open={has} className="rounded-2xl border border-slate-200 bg-white">
            <summary className="cursor-pointer p-3 font-bold text-brand-900">{g.no}. {g.name} <span className="text-xs font-normal text-slate-400" dir="ltr">{g.en}</span>
              <span className="ms-2 text-xs text-slate-500">({g.types.reduce((n, t) => n + (byType.get(t.code)?.length ?? 0), 0)})</span></summary>
            <div className="space-y-2 border-t border-slate-100 p-3">
              {g.types.map((t) => {
                const fs = byType.get(t.code) ?? []
                return (
                  <div key={t.code}>
                    <p className={`text-sm ${fs.length ? 'font-bold text-slate-800' : 'text-slate-400'}`}>{t.code} {t.name}</p>
                    {fs.length > 0 && (
                      <div className="mt-1 grid gap-2 sm:grid-cols-3 lg:grid-cols-4">
                        {fs.map((f) => (
                          <div key={f.id} className="overflow-hidden rounded-lg border border-slate-200">
                            <a href={fileUrl(f.fileKey)} target="_blank" rel="noreferrer" className="block h-28 bg-slate-50">
                              {isImageFile(f) ? <img src={fileUrl(f.fileKey)} alt="" className="h-full w-full object-contain" /> : <span className="grid h-full place-items-center text-sm font-bold text-slate-500">{f.fileType === 'application/pdf' ? '📄 PDF' : '📐 DWG'}</span>}
                            </a>
                            <div className="flex items-center justify-between gap-1 p-2 text-xs">
                              <span className="truncate" title={f.title}>{f.title}</span>
                              {d.canManage && !locked && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-files/${f.id}`, { method: 'DELETE' }))} className="shrink-0 text-red-600">شيل</button>}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </details>
        )
      })}
    </div>
  )
}
