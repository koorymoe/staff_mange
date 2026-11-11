import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { api, apiRequest, ensureFileToken, fileUrl } from '../api'
import { useSession } from '../session'
import { buildPlanSetHtml, isImageFile, type PlanPrintQuote } from '../utils/planSetPrint'
import { printProjectQuote } from '../utils/projectQuotePrint'
import { ArtCrop, IconCity, IconFolder, IconPdf, IconUpload } from './design/planArt'
import type { PlanMeta, PlanSetDetail, PlanSetRow } from './design/planTypes'

// ═══ 📐 سيت واحد — (ك) 11-10 بطلب دازاي ═══
// «تختار الي تريده وتسوي ابلود وترفع صورة أو ملف للمخطط… ينحفظ بالنظام… تسوي تنزيل
// ملف PDF بالهوية البصرية بي كل التصاميم بالإضافة الى عرض السعر».
// والشكل كحلي نفس صفحة «سيت المخططات» — بطلبه: «سوي صفحة داخل السيت بنفس الشكل الكحلي».
const ORANGE = 'linear-gradient(90deg,#f59e0b,#ea7a1c)'
const GROUP_ICON: Record<number, string> = {
  1: '🏛️', 2: '🧱', 3: '⚡', 4: '❄️', 5: '📹', 6: '🏗️', 7: '🚪', 8: '📋', 9: '🧮', 10: '🖼️',
}
const FIELD ='rounded-lg border border-slate-200 bg-white p-2 text-sm'

export default function PlanSetPage() {
  const { id = '' } = useParams()
  const [params, setParams] = useSearchParams()
  const nav = useNavigate()
  // (ل) 11-10: النسخ من نموذج جاهز
  const [templates, setTemplates] = useState<PlanSetRow[] | null>(null)
  const [fromTpl, setFromTpl] = useState('')
  const [tplDetail, setTplDetail] = useState<PlanSetDetail | null>(null)
  const [copySecs, setCopySecs] = useState<Set<number>>(new Set())
  const [msg, setMsg] = useState('')
  const { employee, permissions } = useSession()
  const [meta, setMeta] = useState<PlanMeta | null>(null)
  const [d, setD] = useState<PlanSetDetail | null>(null)
  const [reload, setReload] = useState(0)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [group, setGroup] = useState(1)
  const [type, setType] = useState('')
  const [title, setTitle] = useState('')
  const printed = useRef(false)

  useEffect(() => {
    let alive = true
    // (ل) 11-11: وسم الملفات قبل الرسم — بدونه رابط الصورة يطلع بلا ?ft والسيرفر يرفضه (صورة مكسورة)
    Promise.all([apiRequest<PlanMeta>('/plan-sets/meta'), apiRequest<PlanSetDetail>(`/plan-sets/${id}`), ensureFileToken()])
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
      title: d.title, projectName: d.projectName, groups: meta.groups, files: d.items, quote, quoteNote, date: new Date(), noQuote: true,
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

  // 💰 PDF عرض السعر: العرض المربوط بالسيت (أو عرض المشروع) يفتح بقالبه وينطبع.
  // ماكو عرض؟ نفتح عرض جديد مربوط بهالسيت — ينحفظ وينربط لحاله.
  const openQuote = async () => {
    if (!d) return
    let qid = d.quotationId
    if (!qid && d.projectId && canQuotes) {
      try { qid = (await api.getQuotationByProject(d.projectId))?.id ?? null } catch { qid = null }
      if (qid) void apiRequest(`/plan-sets/${id}`, { method: 'PUT', body: JSON.stringify({ quotationId: qid }) }).catch(() => undefined)
    }
    if (qid) nav(`/quotations/project/${qid}?print=1`)
    else nav(`/quotations/project/new?planSet=${id}${d.projectId ? `&project=${d.projectId}` : ''}`)
  }
  const loadTemplates = () => {
    if (templates) return
    apiRequest<PlanSetRow[]>('/plan-sets').then((r) => setTemplates(r.filter((x) => x.isTemplate && !x.archivedAt && x.id !== id))).catch(() => setTemplates([]))
  }
  const chooseTpl = (tid: string) => {
    setFromTpl(tid); setCopySecs(new Set()); setTplDetail(null)
    if (tid) apiRequest<PlanSetDetail>(`/plan-sets/${tid}`).then(setTplDetail).catch((e: Error) => setErr(e.message))
  }
  const doCopy = () => run(async () => {
    const r = await apiRequest<{ copied: number }>(`/plan-sets/${id}/copy`, { method: 'POST', body: JSON.stringify({ fromSetId: fromTpl, sections: [...copySecs] }) })
    setMsg(r.copied ? `انسخ ${r.copied} مخطط من النموذج` : 'كلها موجودة أصلاً — ما انسخ شي')
    setCopySecs(new Set())
  })

  if (!d || !meta) return <p className="rounded-2xl bg-white p-4 text-sm text-slate-500">{err || 'يحمّل…'}</p>
  const byType = new Map<string, typeof d.items>()
  for (const f of d.items) byType.set(f.typeCode, [...(byType.get(f.typeCode) ?? []), f])
  const locked = !!d.archivedAt
  const current = meta.groups.find((g) => g.no === group)

  return (
    <div dir="rtl" className="space-y-4">
      <Link to="/plan-sets" className="inline-flex items-center gap-1 text-xs font-bold text-brand-700">→ كل السيتات</Link>

      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-l text-white from-[#0f1a3d] via-[#15215a] to-[#1b2b6b]">
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
            <button type="button" disabled={busy || d.items.length === 0} onClick={() => void download()} className="flex items-center gap-2 rounded-full px-5 py-2 text-sm font-bold text-white shadow-lg shadow-orange-900/40 disabled:opacity-50" style={{ background: ORANGE }}><IconPdf size={18} /> 📄 PDF التصاميم</button>
            <button type="button" disabled={busy} onClick={() => void openQuote()} className="rounded-full bg-white px-5 py-2 text-sm font-bold text-[#15215a] shadow">💰 {d.quotationId ? 'PDF عرض السعر' : 'عرض السعر'}</button>
            {d.canManage && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-sets/${id}`, { method: 'PUT', body: JSON.stringify({ isTemplate: !d.isTemplate }) }))} className="rounded-full border border-white/30 bg-white/10 px-4 py-2 text-sm font-bold">{d.isTemplate ? '⭐ نموذج جاهز — شيله' : '☆ خلّيه نموذج جاهز'}</button>}
            {d.canManage && <button type="button" disabled={busy} onClick={() => void run(() => apiRequest(`/plan-sets/${id}/archive`, { method: 'POST', body: JSON.stringify({ archived: !locked }) }))} className="rounded-full border border-white/30 bg-white/10 px-4 py-2 text-sm font-bold">{locked ? '↩️ رجّعه' : '🗄️ أرشفة'}</button>}
          </div>
        </div>
      </div>
      {err && <p className="rounded-xl bg-red-50 p-2 text-sm text-red-600">{err}</p>}
      {msg && <p role="status" className="rounded-xl bg-emerald-50 p-2 text-sm font-bold text-emerald-700">{msg}</p>}

      {/* (ل) 11-10 دازاي: «عندي سيتات جاهزة… الزبون طلب الكاميرات والطاقة الشمسية… انسخه واضيفه ع مشروعه» */}
      {d.canManage && !d.isTemplate && !locked && (
        <details className="rounded-2xl border border-amber-200 bg-white p-3" onToggle={(e) => { if ((e.target as HTMLDetailsElement).open) loadTemplates() }}>
          <summary className="cursor-pointer text-sm font-extrabold text-slate-800">📋 خذ من نموذج جاهز</summary>
          <div className="mt-3 space-y-2">
            <select id="ps-copy-from" value={fromTpl} onChange={(e) => chooseTpl(e.target.value)} className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm">
              <option value="">{templates ? (templates.length ? 'اختار النموذج…' : 'ماكو نماذج جاهزة بعد') : 'يحمّل…'}</option>
              {templates?.map((t) => <option key={t.id} value={t.id}>⭐ {t.title} ({t.files} تصميم)</option>)}
            </select>
            {tplDetail && (
              <>
                <div className="grid gap-1 sm:grid-cols-2">
                  {meta.groups.map((g) => {
                    const n = tplDetail.items.filter((f) => f.typeCode.startsWith(`${g.no}.`)).length
                    if (!n) return null
                    return (
                      <label key={g.no} className="flex items-center gap-2 rounded-lg p-2 text-sm hover:bg-slate-50">
                        <input id={`cp-${g.no}`} type="checkbox" checked={copySecs.has(g.no)} onChange={() => setCopySecs((s) => { const x = new Set(s); if (x.has(g.no)) x.delete(g.no); else x.add(g.no); return x })} />
                        {g.no}. {g.name} <span className="text-xs text-slate-500">({n})</span>
                      </label>
                    )
                  })}
                </div>
                <button type="button" disabled={busy || copySecs.size === 0} onClick={() => void doCopy()} className="rounded-xl bg-amber-500 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">انسخ الأقسام المختارة لهالسيت</button>
              </>
            )}
          </div>
        </details>
      )}

      {/* (ك) 11-10: الأقسام العشرة كروت صغيرة مرتبة، ومن تنطب على وحدة تطلع مخططاتها تحت —
          بطلب دازاي «تبويبات كروت صغيرة مرتبة… ومن انطب عليهن يطلع المخططات الي بداخلهم». */}
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5">
        {meta.groups.map((g) => {
          const count = g.types.reduce((n, t) => n + (byType.get(t.code)?.length ?? 0), 0)
          const on = g.no === group
          return (
            <button key={g.no} type="button" onClick={() => { setGroup(g.no); setType(g.types.length === 1 ? g.types[0].code : '') }}
              className={`flex items-center gap-2 rounded-xl border p-2.5 text-start transition ${on ? 'border-brand-700 bg-brand-700 text-white shadow-md' : 'border-slate-200 bg-white hover:border-brand-700'}`}>
              <span className={`grid h-8 w-8 shrink-0 place-items-center rounded-lg text-sm font-extrabold ${on ? 'bg-white/20 text-white' : 'bg-slate-100 text-brand-700'}`}>{g.no}</span>
              <span className="min-w-0 flex-1">
                <b className={`block text-xs leading-snug ${on ? 'text-white' : 'text-slate-800'}`} title={g.name}>{GROUP_ICON[g.no]} {g.name}</b>
                <span className={`text-[11px] ${on ? 'text-white/80' : 'text-slate-400'}`}>{g.types.length} نوع · {count} ملف</span>
              </span>
            </button>
          )
        })}
      </div>

      {current && (
        <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4">
          <div className="flex flex-wrap items-center gap-3">
            <span className="grid h-10 w-10 place-items-center rounded-xl bg-brand-700 text-lg font-extrabold text-white">{current.no}</span>
            <div className="flex-1">
              <h3 className="text-lg font-extrabold text-brand-900">{GROUP_ICON[current.no]} {current.name}</h3>
              <p className="text-xs text-slate-400" dir="ltr">{current.en}</p>
            </div>
          </div>

          {!locked && (
            <div className="space-y-2 rounded-xl border-2 border-dashed border-slate-300 bg-slate-50 p-3"
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => { e.preventDefault(); if (e.dataTransfer.files.length) upload(e.dataTransfer.files) }}>
              <p className="flex items-center gap-2 text-sm font-bold text-slate-700"><span className="text-brand-700"><IconUpload size={18} /></span> رفع مخططات لهذا القسم</p>
              {current.types.length > 1 && (
                <div className="flex flex-wrap gap-1.5">
                  {current.types.map((t) => (
                    <button key={t.code} type="button" onClick={() => setType(t.code)}
                      className={`rounded-full px-3 py-1 text-xs font-bold ${type === t.code ? 'bg-amber-500 text-white' : 'bg-white text-slate-500 border border-slate-200'}`}>
                      {t.code} {t.name}
                    </button>
                  ))}
                </div>
              )}
              <div className="flex flex-wrap items-center gap-2">
                <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="اسم الملف (اختياري)" className={`min-w-48 flex-1 ${FIELD}`} />
                <label className={`inline-flex cursor-pointer items-center gap-2 rounded-full px-5 py-2 text-sm font-bold text-white ${type && !busy ? '' : 'pointer-events-none opacity-40'}`} style={{ background: ORANGE }}>
                  <IconFolder size={18} /> {busy ? 'يرفع…' : type ? `اختيار الملفات (${type})` : 'اختار النوع أول'}
                  <input type="file" multiple accept="image/png,image/jpeg,image/webp,application/pdf,.dwg" className="hidden" onChange={(e) => { if (e.target.files?.length) upload(e.target.files); e.target.value = '' }} />
                </label>
              </div>
              <p className="text-[11px] text-slate-400">اسحب الملفات هنا أو اختارها · PDF، DWG، JPG، PNG · لحد ٢٤ ميگا</p>
            </div>
          )}

          <div className="grid items-start gap-2 md:grid-cols-2">
            {current.types.map((t) => {
              const fs = byType.get(t.code) ?? []
              return (
                <div key={t.code} className={`rounded-xl border p-3 ${type === t.code ? 'border-amber-500' : 'border-slate-200'}`}>
                  <div className="flex items-center gap-2">
                    <span className={`rounded-md px-2 py-0.5 text-[11px] font-bold ${fs.length ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-slate-500'}`} dir="ltr">{t.code}</span>
                    <b className={`flex-1 text-sm ${fs.length ? 'text-slate-800' : 'text-slate-500'}`}>{t.name}</b>
                    <span className="text-[11px] text-slate-400">{fs.length ? `${fs.length} ملف` : 'ماكو بعد'}</span>
                    {!locked && <button type="button" onClick={() => setType(t.code)} className="text-[11px] font-bold text-amber-600">+ رفع</button>}
                  </div>
                  {fs.length > 0 && (
                    <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                      {fs.map((f) => (
                        <div key={f.id} className="overflow-hidden rounded-lg border border-slate-200 bg-white">
                          <a href={fileUrl(f.fileKey)} target="_blank" rel="noreferrer" className="block h-20 bg-slate-50">
                            {isImageFile(f) ? <img src={fileUrl(f.fileKey)} alt="" className="h-full w-full object-cover" /> : <span className="grid h-full place-items-center text-xs font-bold text-brand-700">{f.fileType === 'application/pdf' ? '📄 PDF' : '📐 DWG'}</span>}
                          </a>
                          <div className="flex items-center justify-between gap-1 p-1.5 text-[11px]">
                            <span className="truncate text-slate-700" title={f.title}>{f.title}</span>
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
        </div>
      )}
    </div>
  )
}
