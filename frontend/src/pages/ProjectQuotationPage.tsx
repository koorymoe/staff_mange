import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { api, apiRequest, ensureFileToken, fileUrl } from '../api'
import type { Product, QuotationItem } from '../apiTypes'
import PageHeader from '../components/PageHeader'
import AnnexTemplatesManager, { type AnnexTemplate } from '../components/AnnexTemplatesManager'
import { useSession } from '../session'
import { arabicAmountWords } from '../utils/arabicWords'
import {
  buildProjectQuoteHtml,
  printProjectQuote,
  type AnnexPart,
  type ProjectCover,
  type ProjectItem,
} from '../utils/projectQuotePrint'

// ═══ «عرض مشاريع» (دازاي 11-06) — بجانب عرض السعر القديم ═══
// الموظف يختار المنتجات من الكتالوج فتنسحب صورتها وموديلها ومواصفاتها وفئتها (الفصل)،
// يكتب الكمية والسعر والغلاف، ويختار قالب الملحق حسب نوع المشروع. الأرقام كلها من الفورمة.

type Row = ProjectItem & { key: number }
let seq = 0
const emptyRow = (): Row => ({ key: ++seq, productName: '', quantity: 1, unitPrice: 0, specs: '', modelName: '', category: '', imageBase64: null, unit: 'قطعة' })
const num = (v: number) => Math.round(v).toLocaleString('en-US')
const input = 'w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm'

export default function ProjectQuotationPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { employee } = useSession()
  const [products, setProducts] = useState<Product[]>([])
  const [templates, setTemplates] = useState<AnnexTemplate[]>([])
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [cover, setCover] = useState<ProjectCover>({ city: 'Karbala, Iraq' })
  // (بلاغ #17) المواصفات تطلع لحالها من الموديل — بلا ما يكتبها الموظف سطر سطر
  const [specsBusy, setSpecsBusy] = useState<number | null>(null)
  const fillSpecs = async (r: Row) => {
    const model = (r.modelName ?? '').trim()
    if (!model || specsBusy !== null) return
    setSpecsBusy(r.key)
    try {
      const out = await api.suggestProductSpecs(r.productName.trim() || model, model)
      if (out.specs.trim()) set(r.key, { specs: out.specs.trim() })
      else setMsg({ ok: false, t: 'ما لگينا مواصفات أكيدة لهذا الموديل — اكتبها يدوياً' })
    } catch (e) {
      setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر جلب المواصفات' })
    } finally {
      setSpecsBusy(null)
    }
  }
  // (بلاغ #13) صورة وحدة للواجهة يختارها الموظف (مثلاً مولّدة بالذكاء) بدل تجميع صور المنتجات
  const [coverBusy, setCoverBusy] = useState(false)
  const uploadCover = async (file?: File) => {
    if (!file) return
    setCoverBusy(true)
    try {
      const up = await api.uploadFile(file, 'quotations')
      setCover((c) => ({ ...c, image: up.url }))
    } catch (e) {
      setMsg({ ok: false, t: e instanceof Error ? e.message : 'تعذر رفع الصورة' })
    } finally {
      setCoverBusy(false)
    }
  }
  const [rows, setRows] = useState<Row[]>([emptyRow()])
  const chapterNames = useMemo(() => [...new Set(rows.map((r) => (r.category ?? '').trim() || 'التجهيزات').filter((n, i, a) => a.indexOf(n) === i))], [rows])
  const [annex, setAnnex] = useState<AnnexPart[]>([])
  const [number, setNumber] = useState('')
  const [msg, setMsg] = useState<{ t: string; ok: boolean } | null>(null)
  const [busy, setBusy] = useState(false)
  const [openSpecs, setOpenSpecs] = useState<number | null>(null)
  const [manage, setManage] = useState(false)
  // (ل) 11-10 دازاي: صور العرض من مكتبة أنواع العمل — كل عرض صورة غير؛ ونماذج عروض يحفظها ماتركس ويبني عليها
  const [params] = useSearchParams()
  const [picBusy, setPicBusy] = useState(false)
  const [workTypes, setWorkTypes] = useState<{ workTypeId: string; name: string; images: number }[]>([])
  const [libType, setLibType] = useState('')
  const [qTemplates, setQTemplates] = useState<{ id: string; name: string; summary: string }[] | null>(null)
  const printedRef = useRef(false)
  const [ftReady, setFtReady] = useState(false)
  const pickImages = async (again: boolean, only?: string) => {
    if (!id) return setMsg({ ok: false, t: 'احفظ العرض أول، بعدين النظام يختار الصور' })
    setPicBusy(true)
    try {
      const out = await apiRequest<Record<string, string>>(`/quotations/${id}/images/pick`, { method: 'POST', body: JSON.stringify({ chapters: only ? [only] : chapterNames, again }) })
      const got = Object.keys(out).length
      setCover((c) => {
        const ch = { ...c.chapters }
        for (const [k, v] of Object.entries(out)) ch[k] = { ...ch[k], image: v }
        return { ...c, chapters: ch }
      })
      setMsg(got ? { ok: true, t: `اختار ${got} صورة — احفظ حتى تثبت` } : { ok: false, t: 'ماكو صور بالمكتبة لهالفصول — ارفع صور لنوع العمل تحت' })
    } catch (e) {
      setMsg({ ok: false, t: (e as Error).message })
    } finally {
      setPicBusy(false)
    }
  }
  const uploadLib = async (files: FileList | null) => {
    if (!files?.length || !libType) return
    setPicBusy(true)
    try {
      const keys: string[] = []
      for (const f of Array.from(files)) keys.push((await api.uploadFile(f, 'work-type-images')).url)
      await apiRequest('/work-type-images', { method: 'POST', body: JSON.stringify({ workTypeId: libType, fileKeys: keys }) })
      setWorkTypes(await apiRequest<{ workTypeId: string; name: string; images: number }[]>('/work-type-images').catch(() => workTypes))
      setMsg({ ok: true, t: `انرفعت ${keys.length} صورة للمكتبة` })
    } catch (e) {
      setMsg({ ok: false, t: (e as Error).message })
    } finally {
      setPicBusy(false)
    }
  }
  const saveAsTemplate = async () => {
    if (!id) return setMsg({ ok: false, t: 'احفظ العرض أول' })
    const name = (cover.title || cover.label || customerName || 'نموذج').trim()
    try {
      await apiRequest('/quotation-templates', { method: 'POST', body: JSON.stringify({ quotationId: id, name }) })
      setMsg({ ok: true, t: `انحفظ كنموذج «${name}» — ماتركس يگدر يبني عليه عروض ثانية` })
    } catch (e) { setMsg({ ok: false, t: (e as Error).message }) }
  }
  const applyTemplate = async (tid: string) => {
    if (!customerName.trim()) return setMsg({ ok: false, t: 'اكتب اسم الزبون أول' })
    try {
      const q = await apiRequest<{ id: string }>(`/quotation-templates/${tid}/use`, { method: 'POST', body: JSON.stringify({ customerName: customerName.trim(), customerPhone: customerPhone.trim(), projectId: params.get('project') ?? '' }) })
      navigate(`/quotations/project/${q.id}`)
    } catch (e) { setMsg({ ok: false, t: (e as Error).message }) }
  }
  const reloadTemplates = () => void apiRequest<AnnexTemplate[]>('/quotation-annex-templates').then(setTemplates).catch(() => {})

  useEffect(() => {
    void api.getProducts().then(setProducts).catch(() => setProducts([]))
    void apiRequest<AnnexTemplate[]>('/quotation-annex-templates')
      .then((t) => {
        setTemplates(t)
        if (!id && t[0]) setAnnex(t[0].parts)
      })
      .catch(() => setTemplates([]))
  }, [id])

  useEffect(() => {
    if (!id) return
    void api.getQuotation(id).then((q) => {
      setNumber(q.quotationNumber)
      setCustomerName(q.customerName)
      setCustomerPhone(q.customerPhone ?? '')
      setCover(q.cover ?? {})
      setAnnex(q.annex ?? [])
      setRows(q.items.map((it) => ({ key: ++seq, productName: it.productName, quantity: it.quantity, unitPrice: it.unitPrice, specs: it.specs ?? '', modelName: it.modelName ?? '', category: it.category ?? '', imageBase64: it.imageBase64 ?? null, unit: it.unit })))
    }).catch((e: Error) => setMsg({ t: e.message, ok: false }))
  }, [id])

  useEffect(() => {
    // صور الفصول تنعرض بـ<img> — تحتاج الوسم قبل الرسم، وإلا تطلع مكسورة لحد ما تتحدث الصفحة
    void ensureFileToken().then(() => setFtReady(true))
    void apiRequest<{ workTypeId: string; name: string; images: number }[]>('/work-type-images').then(setWorkTypes).catch(() => setWorkTypes([]))
    if (!id) void apiRequest<{ id: string; name: string; summary: string }[]>('/quotation-templates').then(setQTemplates).catch(() => setQTemplates([]))
  }, [id])

  const valid = rows.filter((r) => r.productName.trim())
  const total = useMemo(() => valid.reduce((a, r) => a + r.quantity * r.unitPrice, 0), [valid])
  const set = (key: number, patch: Partial<Row>) => setRows((rs) => rs.map((r) => (r.key === key ? { ...r, ...patch } : r)))

  // اختيار منتج من الكتالوج: الصورة والموديل والمواصفات والفئة والسعر تنسحب لحالها
  const pick = (key: number, name: string) => {
    const p = products.find((x) => x.name === name)
    if (!p) return set(key, { productName: name })
    set(key, { productName: p.name, modelName: p.modelName ?? '', specs: p.specs ?? '', category: p.serviceName ?? '', imageBase64: p.imageBase64 ?? null, unitPrice: p.defaultPrice ?? 0, unit: p.unit ?? 'قطعة' })
  }

  const payload = () => ({
    customerName: customerName.trim(),
    customerPhone: customerPhone.trim() || undefined,
    projectName: cover.label || cover.title || undefined,
    items: valid.map((r): Omit<QuotationItem, 'id'> => ({ productName: r.productName.trim(), unit: r.unit ?? 'قطعة', quantity: r.quantity, unitPrice: r.unitPrice, totalPrice: r.quantity * r.unitPrice, imageBase64: r.imageBase64 ?? undefined, modelName: r.modelName || null, specs: r.specs || null, category: r.category || null })),
    grandTotal: total, discountPercent: 0, discountValue: 0, netTotal: total,
    layout: 'PROJECT' as const, cover, annex,
    createdByEmployeeId: employee?.id,
  })

  const save = async () => {
    if (!customerName.trim()) return setMsg({ t: 'اكتب اسم الزبون أو الجهة', ok: false })
    if (!valid.length) return setMsg({ t: 'ضيف منتج واحد على الأقل', ok: false })
    if (total <= 0) return setMsg({ t: 'المجموع صفر — تأكد من الأسعار والكميات', ok: false })
    setBusy(true); setMsg(null)
    try {
      if (id) {
        await apiRequest(`/quotations/${id}`, { method: 'PUT', body: JSON.stringify(payload()) })
        setMsg({ t: 'انحفظ ✅', ok: true })
      } else {
        const q = await apiRequest<{ id: string; quotationNumber: string }>('/quotations', { method: 'POST', body: JSON.stringify(payload()) })
        setNumber(q.quotationNumber)
        const ps = params.get('planSet')
        if (ps) await apiRequest(`/plan-sets/${ps}`, { method: 'PUT', body: JSON.stringify({ quotationId: q.id }) }).catch(() => undefined)
        navigate(`/quotations/project/${q.id}`, { replace: true })
        setMsg({ t: 'انحفظ ✅', ok: true })
      }
    } catch (e) {
      setMsg({ t: (e as Error).message, ok: false })
    } finally {
      setBusy(false)
    }
  }

  const print = () => {
    if (!valid.length || total <= 0) return setMsg({ t: 'ماكو بنود بأسعار حتى ينطبع', ok: false })
    printProjectQuote(buildProjectQuoteHtml({ quotationNumber: number, customerName, cover, items: valid, annex, date: new Date() }, `${import.meta.env.BASE_URL}apple-touch-icon.png`))
  }

  // ?print=1 (من «💰 PDF عرض السعر» بسيت المخططات): ينطبع مرة وحدة بعد ما يحمّل
  useEffect(() => {
    if (params.get('print') !== '1' || printedRef.current || !number || !valid.length) return
    printedRef.current = true
    print()
  })

  const setAnnexItem = (pi: number, ii: number, patch: Partial<{ title: string; text: string }>) =>
    setAnnex((a) => a.map((p, i) => (i !== pi ? p : { ...p, items: p.items.map((it, j) => (j === ii ? { ...it, ...patch } : it)) })))

  return (
    <div dir="rtl" className="space-y-4 p-4">
      <PageHeader title="📘 عرض مشاريع" subtitle={`ملف تجهيز فني وتجاري بالشكل الجديد${number ? ` · رقم ${number}` : ''} — عرض السعر القديم يبقى مثل ما هو`} />
      {msg && <p role={msg.ok ? 'status' : 'alert'} className={`rounded-xl p-3 text-sm ${msg.ok ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'}`}>{msg.t}</p>}

      {!id && qTemplates && qTemplates.length > 0 && (
        <section className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-4 text-sm shadow-sm">
          <b>📋 من نموذج:</b>
          {qTemplates.map((t) => <button key={t.id} type="button" title={t.summary} onClick={() => void applyTemplate(t.id)} className="rounded-full bg-indigo-50 px-3 py-1 text-indigo-800">{t.name}</button>)}
          <small className="text-slate-500">(اكتب اسم الزبون أول — العرض ينبني بنفس البنود والشروط وتگدر تعدّل)</small>
        </section>
      )}

      <section className="grid gap-3 rounded-2xl bg-white p-4 shadow-sm md:grid-cols-2">
        <h2 className="text-base font-bold text-slate-700 md:col-span-2">الغلاف والزبون</h2>
        <label className="text-sm">الزبون / الجهة<input className={input} value={customerName} onChange={(e) => setCustomerName(e.target.value)} /></label>
        <label className="text-sm">الهاتف<input className={input} value={customerPhone} onChange={(e) => setCustomerPhone(e.target.value)} /></label>
        <label className="text-sm">السطر الصغير فوق العنوان<input className={input} placeholder="مثلاً: مشروع جامعة المكين" value={cover.label ?? ''} onChange={(e) => setCover({ ...cover, label: e.target.value })} /></label>
        <label className="text-sm">المدينة<input className={input} value={cover.city ?? ''} onChange={(e) => setCover({ ...cover, city: e.target.value })} /></label>
        <label className="text-sm md:col-span-2">العنوان الكبير<input className={input} placeholder="مثلاً: مشروع تجهيز وتحديث البنية التحتية الرقمية" value={cover.title ?? ''} onChange={(e) => setCover({ ...cover, title: e.target.value })} /></label>
        <div className="text-sm md:col-span-2">
          صورة الواجهة <small className="text-slate-500">(اختياري — صورة وحدة مرتبة للغلاف، وإذا ماكو تنجمع صور المنتجات)</small>
          <div className="mt-1 flex flex-wrap items-center gap-2">
            <input type="file" accept="image/*" disabled={coverBusy} onChange={(e) => void uploadCover(e.target.files?.[0])} className="text-xs" />
            {coverBusy && <span className="text-xs text-slate-500">يرفع…</span>}
            {cover.image && <><span className="text-xs text-emerald-700">✓ انرفعت</span><button type="button" onClick={() => setCover({ ...cover, image: undefined })} className="rounded bg-rose-50 px-2 py-0.5 text-xs text-rose-700">شيلها</button></>}
          </div>
        </div>
        <label className="text-sm md:col-span-2">الوصف<input className={input} placeholder="مثلاً: حلول الحوسبة والمختبرات، والطباعة والأرشفة، والشاشات الذكية" value={cover.subtitle ?? ''} onChange={(e) => setCover({ ...cover, subtitle: e.target.value })} /></label>
      </section>


      <section className="space-y-3 rounded-2xl border-2 border-dashed border-amber-300 bg-amber-50/40 p-4">
        <h2 className="text-base font-bold text-slate-700">🖼️ صور العرض — ارفع أو ارفق صور <small className="font-normal text-slate-500">— لكل فصل صورة تناسب شغله، والنظام يختار صورة غير بكل عرض</small></h2>
        <div className="flex flex-wrap gap-2">
          <button type="button" disabled={picBusy || !id} onClick={() => void pickImages(false)} className="rounded-xl bg-amber-500 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">✨ اختار صور للفصول</button>
          {!id && <span className="self-center text-xs text-slate-500">احفظ العرض أول</span>}
        </div>
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {chapterNames.map((n) => {
            const img = cover.chapters?.[n]?.image
            return (
              <div key={n} className="flex items-center gap-2 rounded-xl bg-white p-2 shadow-sm">
                {img && ftReady ? <img src={fileUrl(img)} alt="" className="h-14 w-20 rounded object-cover" /> : <div className="grid h-14 w-20 place-items-center rounded bg-slate-100 text-xs text-slate-400">بلا صورة</div>}
                <div className="min-w-0 flex-1"><b className="block truncate text-sm">{n}</b>
                  <div className="flex gap-1">
                    <button type="button" disabled={picBusy || !id} onClick={() => void pickImages(true, n)} className="rounded bg-slate-100 px-2 py-0.5 text-xs">🔄 صورة غيرها</button>
                    {img && <button type="button" onClick={() => setCover((c) => ({ ...c, chapters: { ...c.chapters, [n]: { ...c.chapters?.[n], image: undefined } } }))} className="rounded bg-rose-50 px-2 py-0.5 text-xs text-rose-700">شيلها</button>}
                  </div>
                </div>
              </div>
            )
          })}
        </div>
        <div className="flex flex-wrap items-center gap-2 border-t border-amber-200 pt-3 text-sm">
          <span className="font-bold">📤 ارفع صور للمكتبة:</span>
          <select id="lib-type" value={libType} onChange={(e) => setLibType(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm">
            <option value="">نوع العمل…</option>
            {workTypes.map((w) => <option key={w.workTypeId} value={w.workTypeId}>{w.name} ({w.images} صورة)</option>)}
          </select>
          <input type="file" multiple accept="image/*" disabled={!libType || picBusy} onChange={(e) => void uploadLib(e.target.files)} className="text-xs" />
          {picBusy && <span className="text-xs text-slate-500">يشتغل…</span>}
        </div>
      </section>
      {chapterNames.length > 0 && (
        <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
          <h2 className="text-base font-bold text-slate-700">الفصول <small className="font-normal text-slate-500">— اختياري: وسم إنگليزي ووصف قصير لكل فصل (يطلعن بالنظرة الشاملة ورأس الفصل)</small></h2>
          {chapterNames.map((name) => (
            <div key={name} className="grid gap-2 md:grid-cols-6">
              <b className="self-center text-sm text-brand-800 md:col-span-1">{name}</b>
              <input dir="ltr" className={`${input} md:col-span-2`} placeholder="COMPUTING DEVICES" value={cover.chapters?.[name]?.en ?? ''}
                onChange={(e) => setCover((c) => ({ ...c, chapters: { ...c.chapters, [name]: { ...c.chapters?.[name], en: e.target.value } } }))} />
              <input className={`${input} md:col-span-3`} placeholder="مثلاً: حواسيب محمولة، ومحطات مختبرية All-in-One، وحاسبة مخصصة للتصميم" value={cover.chapters?.[name]?.desc ?? ''}
                onChange={(e) => setCover((c) => ({ ...c, chapters: { ...c.chapters, [name]: { ...c.chapters?.[name], desc: e.target.value } } }))} />
            </div>
          ))}
        </section>
      )}

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="text-base font-bold text-slate-700">المنتجات <small className="font-normal text-slate-500">— اختار من الكتالوج وتنسحب الصورة والموديل والمواصفات والفصل لحالها</small></h2>
        <datalist id="pq-products">{products.map((p) => <option key={p.id} value={p.name} />)}</datalist>
        {rows.map((r, i) => (
          <div key={r.key} className="rounded-xl border border-slate-200 p-3">
            <div className="grid items-end gap-2 md:grid-cols-12">
              <span className="text-lg font-black text-amber-500 md:col-span-1">{String(i + 1).padStart(2, '0')}</span>
              <label className="text-xs md:col-span-4">المنتج<input list="pq-products" className={input} value={r.productName} onChange={(e) => pick(r.key, e.target.value)} /></label>
              <label className="text-xs md:col-span-2">الفصل (الفئة)<input className={input} value={r.category ?? ''} onChange={(e) => set(r.key, { category: e.target.value })} /></label>
              <label className="text-xs md:col-span-1">الكمية<input type="number" min={1} className={input} value={r.quantity} onChange={(e) => set(r.key, { quantity: Math.max(1, Number(e.target.value) || 1) })} /></label>
              <label className="text-xs md:col-span-2">السعر المفرد<input type="number" min={0} className={input} value={r.unitPrice} onChange={(e) => set(r.key, { unitPrice: Math.max(0, Number(e.target.value) || 0) })} /></label>
              <div className="flex gap-1 md:col-span-2">
                <button type="button" onClick={() => setOpenSpecs(openSpecs === r.key ? null : r.key)} className="rounded-lg bg-slate-100 px-2 py-1.5 text-xs font-bold">📋 المواصفات</button>
                <button type="button" aria-label="شيل المنتج" onClick={() => setRows((rs) => (rs.length > 1 ? rs.filter((x) => x.key !== r.key) : rs))} className="rounded-lg bg-rose-50 px-2 py-1.5 text-xs font-bold text-rose-700">✕</button>
              </div>
            </div>
            <p className="mt-1 text-xs text-slate-500">الإجمالي: <b>{num(r.quantity * r.unitPrice)}</b> د.ع {r.imageBase64 ? '· 🖼️ بصورته' : '· بدون صورة (تطلع رسمة توضيحية)'}</p>
            {openSpecs === r.key && (
              <div className="mt-2 grid gap-2 md:grid-cols-3">
                <label className="text-xs">الموديل<input dir="ltr" className={input} value={r.modelName ?? ''} onChange={(e) => set(r.key, { modelName: e.target.value })}
                  onBlur={() => { if (!(r.specs ?? '').trim()) void fillSpecs(r) }} />
                  <button type="button" disabled={specsBusy === r.key || !(r.modelName ?? '').trim()} onClick={() => void fillSpecs(r)}
                    className="mt-1 rounded-lg bg-indigo-50 px-2 py-1 text-xs font-bold text-indigo-700 disabled:opacity-50">
                    {specsBusy === r.key ? '…' : '🤖 المواصفات من الموديل'}
                  </button>
                </label>
                <label className="text-xs md:col-span-2">المواصفات (كل سطر «العنوان: القيمة»)<textarea rows={5} className={input} value={r.specs ?? ''} onChange={(e) => set(r.key, { specs: e.target.value })} placeholder={'الشاشة: 15.6 بوصة FHD\nالمعالج: Intel Core i5-13420H'} /></label>
              </div>
            )}
          </div>
        ))}
        <button type="button" onClick={() => setRows((rs) => [...rs, emptyRow()])} className="rounded-lg bg-brand-50 px-3 py-1.5 text-sm font-bold text-brand-700">＋ منتج</button>
        <div className="mt-2 rounded-xl bg-[#0f2040] p-3 text-white">
          <b className="text-xl">{num(total)} د.ع</b>
          <p className="text-xs opacity-80">{total > 0 ? arabicAmountWords(total) : '—'} · {valid.length} بند · {valid.reduce((a, r) => a + r.quantity, 0)} وحدة</p>
        </div>
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="text-base font-bold text-slate-700">الملحق الفني</h2>
          <select className="rounded-lg border border-slate-300 px-2 py-1 text-sm" defaultValue="" onChange={(e) => { const t = templates.find((x) => x.id === e.target.value); if (t) setAnnex(t.parts) }}>
            <option value="" disabled>اختار قالب حسب نوع المشروع…</option>
            {templates.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
          </select>
          <small className="text-slate-500">البنود تنسخ لهذا العرض وتگدر تعدّلها هنا بس.</small>
          <button type="button" onClick={() => setManage(true)} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold">⚙️ قوالب الملحق</button>
        </div>
        {manage && <AnnexTemplatesManager templates={templates} onClose={() => setManage(false)} onSaved={reloadTemplates} />}
        {annex.map((p, pi) => (
          <details key={pi} className="rounded-xl border border-slate-200 p-2" open={pi === 0}>
            <summary className="cursor-pointer font-bold text-[#1e3a8a]">الجزء {pi + 1}: {p.title} ({p.items.length} بند)</summary>
            {p.items.map((it, ii) => (
              <div key={ii} className="mt-2 grid gap-1">
                <div className="flex gap-1"><input className={input} value={it.title} onChange={(e) => setAnnexItem(pi, ii, { title: e.target.value })} />
                  <button type="button" aria-label="شيل البند" onClick={() => setAnnex((a) => a.map((x, i) => (i !== pi ? x : { ...x, items: x.items.filter((_, j) => j !== ii) })))} className="rounded-lg bg-rose-50 px-2 text-xs text-rose-700">✕</button></div>
                <textarea rows={3} className={input} value={it.text} onChange={(e) => setAnnexItem(pi, ii, { text: e.target.value })} />
              </div>
            ))}
            <button type="button" onClick={() => setAnnex((a) => a.map((x, i) => (i !== pi ? x : { ...x, items: [...x.items, { title: 'بند جديد', text: '' }] })))} className="mt-2 rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold">＋ بند</button>
          </details>
        ))}
      </section>

      <div className="flex flex-wrap gap-2">
        <button type="button" disabled={busy} onClick={() => void save()} className="rounded-xl bg-brand-600 px-5 py-2 font-bold text-white disabled:opacity-50">{busy ? 'دا أحفظ…' : '💾 حفظ'}</button>
        <button type="button" onClick={print} className="rounded-xl bg-amber-500 px-5 py-2 font-bold text-white">🖨️ طباعة / PDF</button>
        {id && <button type="button" onClick={() => void saveAsTemplate()} className="rounded-xl border border-indigo-300 bg-indigo-50 px-5 py-2 font-bold text-indigo-800">💾 حفظ كنموذج</button>}
        <button type="button" onClick={() => navigate('/quotations')} className="rounded-xl border px-5 py-2 text-slate-600">رجوع</button>
      </div>
    </div>
  )
}
