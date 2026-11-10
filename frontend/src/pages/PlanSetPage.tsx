import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { api, apiRequest, fileUrl } from '../api'
import { useSession } from '../session'
import { buildPlanSetHtml, isImageFile, type PlanPrintQuote } from '../utils/planSetPrint'
import { printProjectQuote } from '../utils/projectQuotePrint'
import { ArtCrop, IconCity, IconFolder, IconPdf, IconUpload } from './design/planArt'
import type { PlanMeta, PlanSetDetail } from './design/planTypes'

// ═══ 📐 سيت واحد — (ك) 11-10 بطلب دازاي ═══
// «تختار الي تريده وتسوي ابلود وترفع صورة أو ملف للمخطط… ينحفظ بالنظام… تسوي تنزيل
// ملف PDF بالهوية البصرية بي كل التصاميم بالإضافة الى عرض السعر».
// والشكل كحلي نفس صفحة «سيت المخططات» — بطلبه: «سوي صفحة داخل السيت بنفس الشكل الكحلي».
const ORANGE = 'linear-gradient(90deg,#f59e0b,#ea7a1c)'
const FIELD = 'rounded-lg border border-[#2a3a6e] bg-[#0b1430] p-2 text-sm text-white placeholder:text-white/40'

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

  if (!d || !meta) return <p className="rounded-2xl bg-[#0b1430] p-4 text-sm text-white/70">{err || 'يحمّل…'}</p>
  const byType = new Map<string, typeof d.items>()
  for (const f of d.items) byType.set(f.typeCode, [...(byType.get(f.typeCode) ?? []), f])
  const locked = !!d.archivedAt

  return (
    <div dir="rtl" className="-m-2 space-y-4 rounded-3xl bg-[#0b1430] p-3 text-white sm:p-4">
      <Link to="/plan-sets" className="inline-flex items-center gap-1 text-xs font-bold text-[#b9c6ff]">→ كل السيتات</Link>

      <div className="relative overflow-hidden rounded-3xl border border-[#2a3a6e] bg-gradient-to-l from-[#0f1a3d] via-[#15215a] to-[#1b2b6b]">
        <div className="absolute inset-y-0 left-0 hidden w-[45%] sm:block">
          <ArtCrop part="banner" className="h-full w-full opacity-70" />
          <div className="absolute inset-0 bg-gradient-to-l from-[#15215a] via-transparent to-transparent" />
        </div>
        <div className="relative flex flex-wrap items-center justify-between gap-4 p-5 sm:max-w-[62%]">
          <div className="flex items-center gap-3">
            <span className="text-amber-400"><IconCity size={40} /></span>
            <div>
              <h2 className="text-2xl font-extrabold">{d.title}</h2>
              <p className="text-xs text-white/70">{d.projectCode ? `${d.projectCode} — ${d.projectName ?? ''} · ` : ''}{d.items.length} مخطط{d.createdByName ? ` · فتحه ${d.createdByName}` : ''}{locked ? ' · مؤرشف' : ''}</p>
              {d.notes && <p className="mt-1 text-sm text-white/80">{d.notes}</p>}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <button type="button" disabled={busy || d.items.length === 0} onClick={() => void download()} className="flex items-center gap-2 rounded-full px-5 py-2 text-sm font-bold text-white shadow-lg shadow-orange-900/40 disabled:opacity-50" style={{ background: ORANGE }}><IconPdf size={18} /> تنزيل PDF</button>
            {d.canManage && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-sets/${id}/archive`, { method: 'POST', body: JSON.stringify({ archived: !locked }) }))} className="rounded-full border border-[#3a4c8a] bg-[#1e2b5c] px-4 py-2 text-sm font-bold">{locked ? '↩️ رجّعه' : '🗄️ أرشفة'}</button>}
          </div>
        </div>
      </div>
      {err && <p className="rounded-xl bg-red-500/15 p-2 text-sm text-red-300">{err}</p>}

      {!locked && (
        <div className="space-y-3 rounded-2xl border-2 border-dashed border-[#3a4c8a] bg-[#111c3d]/90 p-5 text-center"
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => { e.preventDefault(); if (e.dataTransfer.files.length) upload(e.dataTransfer.files) }}>
          <span className="mx-auto grid h-14 w-14 place-items-center rounded-full bg-[#24336b] text-[#c9d3ff]"><IconUpload size={28} /></span>
          <p className="font-bold">رفع مخططات جديدة</p>
          <p className="text-xs text-white/60">اختار نوع المخطط، وبعدين اسحب الملفات هنا أو اختارها</p>
          <div className="mx-auto flex max-w-3xl flex-wrap justify-center gap-2">
            <select value={type} onChange={(e) => setType(e.target.value)} className={`min-w-56 flex-1 ${FIELD}`}>
              <option value="">اختار نوع المخطط…</option>
              {meta.groups.map((g) => (
                <optgroup key={g.no} label={`${g.no}. ${g.name}`}>
                  {g.types.map((t) => <option key={t.code} value={t.code}>{t.code} {t.name}</option>)}
                </optgroup>
              ))}
            </select>
            <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="اسم الملف (اختياري)" className={`flex-1 ${FIELD}`} />
          </div>
          <div className="flex flex-wrap justify-center gap-1 text-[11px] font-bold">
            {['PDF', 'DWG', 'JPG', 'PNG'].map((x) => <span key={x} className="rounded bg-[#1e2b5c] px-2 py-0.5 text-amber-400">{x}</span>)}
          </div>
          <label className={`mx-auto inline-flex cursor-pointer items-center gap-2 rounded-full px-6 py-2 text-sm font-bold text-white ${type && !busy ? '' : 'pointer-events-none opacity-40'}`} style={{ background: ORANGE }}>
            <IconFolder size={18} /> {busy ? 'يرفع…' : 'اختيار الملفات'}
            <input type="file" multiple accept="image/png,image/jpeg,image/webp,application/pdf,.dwg" className="hidden" onChange={(e) => { if (e.target.files?.length) upload(e.target.files); e.target.value = '' }} />
          </label>
          <p className="text-[11px] text-white/50">لحد ٢٤ ميگا · الصور تطلع صفحات بالـPDF، وPDF وDWG يطلعن بقائمة المرفقات</p>
        </div>
      )}

      <div className="grid items-start gap-3 lg:grid-cols-2">
        {meta.groups.map((g) => {
          const count = g.types.reduce((n, t) => n + (byType.get(t.code)?.length ?? 0), 0)
          return (
            <details key={g.no} open={count > 0} className="overflow-hidden rounded-2xl border border-[#2a3a6e] bg-[#111c3d]/90">
              <summary className="flex cursor-pointer items-center gap-3 p-3">
                <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-[#2b3074] font-extrabold">{g.no}</span>
                <span className="flex-1"><b className="block">{g.name}</b><span className="text-[11px] text-white/50" dir="ltr">{g.en}</span></span>
                <span className={`rounded-full px-3 py-0.5 text-xs font-bold ${count ? 'bg-amber-500 text-white' : 'bg-[#1e2b5c] text-white/50'}`}>{count}</span>
              </summary>
              <div className="space-y-2 border-t border-[#2a3a6e] p-3">
                {g.types.map((t) => {
                  const fs = byType.get(t.code) ?? []
                  return (
                    <div key={t.code}>
                      <p className={`flex items-center gap-2 text-sm ${fs.length ? 'font-bold text-white' : 'text-white/45'}`}>
                        <span className={`h-2 w-2 rounded-full ${fs.length ? 'bg-emerald-400' : 'bg-white/20'}`} />{t.code} {t.name}
                      </p>
                      {fs.length > 0 && (
                        <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                          {fs.map((f) => (
                            <div key={f.id} className="overflow-hidden rounded-xl border border-[#2a3a6e] bg-[#0f1a3a]">
                              <a href={fileUrl(f.fileKey)} target="_blank" rel="noreferrer" className="block h-24 bg-[#0b1430]">
                                {isImageFile(f) ? <img src={fileUrl(f.fileKey)} alt="" className="h-full w-full object-cover" /> : <span className="grid h-full place-items-center text-sm font-bold text-[#c9d3ff]">{f.fileType === 'application/pdf' ? '📄 PDF' : '📐 DWG'}</span>}
                              </a>
                              <div className="flex items-center justify-between gap-1 p-2 text-xs">
                                <span className="truncate" title={f.title}>{f.title}</span>
                                {d.canManage && !locked && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-files/${f.id}`, { method: 'DELETE' }))} className="shrink-0 text-red-300">شيل</button>}
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
    </div>
  )
}
