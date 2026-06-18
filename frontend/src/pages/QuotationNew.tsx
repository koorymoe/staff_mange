import { useEffect, useState, useRef, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Product, type Quotation } from '../api'
import { useSession } from '../session'

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
})

/* ───── colours ───── */
const BLUE = '#1a237e'
const BLUE2 = '#283593'
const GOLD = '#ffd54f'

export default function QuotationNew() {
  const navigate = useNavigate()
  const { employee } = useSession()
  const [products, setProducts] = useState<Product[]>([])
  const [customerName, setCustomerName] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [customerAddress, setCustomerAddress] = useState('')
  const [projectName, setProjectName] = useState('')
  const [items, setItems] = useState<ItemRow[]>([emptyItem()])
  const [discountPercent, setDiscountPercent] = useState(0)
  const [duration, setDuration] = useState('')
  const [notes, setNotes] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [quotationNumber, setQuotationNumber] = useState('')
  const [statusMsg, setStatusMsg] = useState<{ text: string; type: 'ok' | 'err' } | null>(null)

  // autocomplete
  const [acIndex, setAcIndex] = useState<number | null>(null)
  const [acMatches, setAcMatches] = useState<Product[]>([])
  const [acPos, setAcPos] = useState<{ top: number; left: number; width: number }>({ top: 0, left: 0, width: 0 })

  useEffect(() => {
    api.getProducts().then(setProducts).catch(() => {})
    // generate next quotation number
    api.getQuotations().then((qs: Quotation[]) => {
      const nums = qs.map((q) => {
        const m = q.quotationNumber?.match(/(\d+)$/)
        return m ? parseInt(m[1], 10) : 0
      })
      const next = (Math.max(0, ...nums) + 1).toString().padStart(4, '0')
      setQuotationNumber(`AM-${new Date().getFullYear()}-${next}`)
    }).catch(() => {
      setQuotationNumber(`AM-${new Date().getFullYear()}-0001`)
    })
  }, [])

  const grandTotal = items.reduce((sum, item) => sum + item.totalPrice, 0)
  const discountValue = Math.round((grandTotal * discountPercent) / 100)
  const netTotal = grandTotal - discountValue
  const todayStr = new Date().toISOString().split('T')[0]

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

  const selectProduct = (index: number, product: Product) => {
    updateItem(index, {
      productName: product.name,
      unit: product.unit,
      unitPrice: product.defaultPrice,
      imageBase64: product.imageBase64 || undefined,
    })
    setAcIndex(null)
    setAcMatches([])
  }

  const handleNameInput = (index: number, val: string) => {
    updateItem(index, { productName: val })
    if (!val.trim()) { setAcIndex(null); setAcMatches([]); return }
    const matches = products.filter((p) => p.name.toLowerCase().includes(val.toLowerCase())).slice(0, 8)
    if (matches.length) { setAcIndex(index); setAcMatches(matches) } else { setAcIndex(null); setAcMatches([]) }
  }

  const addRow = () => setItems((prev) => [...prev, emptyItem()])
  const removeRow = (index: number) => {
    if (items.length <= 1) { showStatus('لا يمكن الحذف', 'err'); return }
    setItems((prev) => prev.filter((_, i) => i !== index))
  }

  const showStatus = (text: string, type: 'ok' | 'err') => {
    setStatusMsg({ text, type })
    setTimeout(() => setStatusMsg(null), 6000)
  }

  const handleSave = async () => {
    if (!customerName.trim()) { showStatus('الرجاء إدخال اسم الزبون', 'err'); return }
    if (items.every((it) => !it.productName.trim())) { showStatus('الرجاء إضافة منتج واحد على الأقل', 'err'); return }
    setSubmitting(true)
    try {
      await api.createQuotation({
        customerName,
        customerPhone: customerPhone || undefined,
        customerAddress: customerAddress || undefined,
        projectName: projectName || undefined,
        items: items.filter((it) => it.productName.trim()),
        grandTotal,
        discountPercent,
        discountValue,
        netTotal,
        duration: duration || undefined,
        notes: notes || undefined,
        createdByEmployeeId: employee?.id,
      })
      showStatus('تم الحفظ بنجاح', 'ok')
    } catch (e) {
      showStatus(e instanceof Error ? e.message : 'حدث خطأ', 'err')
    } finally {
      setSubmitting(false)
    }
  }

  const handlePrint = () => {
    if (!customerName.trim()) { showStatus('الرجاء إدخال اسم الزبون قبل الطباعة', 'err'); return }

    const printItems = items.filter((it) => it.productName.trim())
    const printWindow = window.open('', '_blank')
    if (!printWindow) return

    // Build print rows
    let tableRows = ''
    printItems.forEach((item, i) => {
      const imgCell = item.imageBase64
        ? `<img src="${item.imageBase64}" style="width:55px;height:55px;object-fit:contain;border-radius:6px;border:1px solid #e0e0e0;background:#fafafa;display:block;margin:0 auto;">`
        : `<div style="width:55px;height:55px;background:#f0f2f8;border-radius:6px;display:flex;align-items:center;justify-content:center;font-size:26px;margin:0 auto;">&#128230;</div>`
      tableRows += `<tr>
        <td style="text-align:center;color:#47528f;font-weight:600;padding:8px 6px;">${i + 1}</td>
        <td style="text-align:center;padding:4px;">${imgCell}</td>
        <td style="text-align:right;padding-right:12px;color:#47528f;font-weight:600;font-size:15px;">${item.productName}</td>
        <td style="text-align:center;color:#47528f;font-weight:600;">${item.unit}</td>
        <td style="text-align:center;color:#47528f;font-weight:600;">${item.quantity}</td>
        <td style="text-align:center;color:#47528f;font-weight:600;">${item.unitPrice.toLocaleString('en-IQ')}</td>
        <td style="text-align:center;color:#47528f;font-weight:600;">${item.totalPrice.toLocaleString('en-IQ')}</td>
      </tr>`
    })
    // Grand total row
    tableRows += `<tr style="background-color:#d4ddef;">
      <td colspan="6" style="text-align:right;padding:11px 24px 11px 6px;font-weight:700;font-size:16px;color:#47528f;background-color:#d4ddef;">المجموع الكلي</td>
      <td style="text-align:center;font-weight:700;font-size:16px;padding:11px 6px;color:#47528f;background-color:#d4ddef;">${grandTotal.toLocaleString('en-IQ')}</td>
    </tr>`

    const termsArr = [
      'مدة التنفيذ: ' + (duration || '---'),
      'جميع الأسعار بالدينار العراقي.',
      'الأسعار قابلة للتغيير دون إشعار مسبق.',
      'يتم الدفع كاملاً قبل أو عند التسليم.',
      'الضمان يشمل عيوب التصنيع فقط ولا يشمل سوء الاستخدام.',
      'العرض ساري لمدة أسبوعين من تاريخه.',
    ]
    const termsHtml = termsArr.map((t, i) => `<li style="position:relative;padding-right:28px;margin-bottom:7px;font-size:13px;line-height:1.7;color:#47528f;"><span style="position:absolute;right:0;top:0;font-weight:700;color:#47528f;">${i + 1}.</span>${t}</li>`).join('')

    const html = `<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
<meta charset="UTF-8">
<title>عرض سعر - ${quotationNumber}</title>
<style>
@import url('https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800;900&display=swap');
*{margin:0;padding:0;box-sizing:border-box;}
body{font-family:'Cairo',Tahoma,sans-serif;background:#fff;margin:0;padding:0;}
@page{size:A4;margin:0;}
.page{width:210mm;height:297mm;position:relative;overflow:hidden;page-break-after:always;background:#fff;}
.page:last-child{page-break-after:auto;}
.vstrip{position:absolute;top:0;right:0;width:26mm;height:297mm;z-index:10;pointer-events:none;}
.vstrip img{width:100%;height:100%;object-fit:cover;display:block;}
.fbanner{position:absolute;bottom:0;left:0;right:0;height:19mm;z-index:10;pointer-events:none;}
.fbanner img{width:100%;height:100%;object-fit:cover;display:block;}
.content{position:relative;z-index:5;padding:10mm 30mm 21mm 10mm;height:100%;display:flex;flex-direction:column;}
.header{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:6px;padding-bottom:8px;border-bottom:1.5px solid #47528f;}
.header-right{text-align:right;flex:1.5;}
.header-left{text-align:left;flex:1;}
.header-company-ar{font-size:17px;font-weight:700;color:#47528f;line-height:1.4;}
.header-company-en{font-size:12px;font-weight:600;color:#c97a3a;margin-top:2px;}
.header-left-text{font-size:17px;font-weight:700;color:#47528f;line-height:1.4;}
.info-row{display:flex;justify-content:space-between;margin:16px 0 12px 0;}
.info-item{flex:1 1 0;min-width:0;display:flex;flex-direction:column;gap:6px;padding:4px 0;}
.info-label{font-size:13px;font-weight:700;color:#47528f;white-space:nowrap;}
.info-value{font-size:14px;font-weight:400;color:#47528f;margin-top:5px;}
.title-right{font-size:15px;font-weight:700;color:#47528f;text-align:right;margin:10px 0 6px 0;}
.title-center{font-size:15px;font-weight:700;color:#47528f;text-align:center;margin:12px 0 8px 0;}
.hr{border:none;border-top:1.5px solid #47528f;margin:8px 0;}
.hr-light{border:none;border-top:1px solid #a0a8c8;margin:10px 0;}
table{width:100%;border-collapse:collapse;margin:10px 0 6px 0;font-size:14px;}
thead th{background:#fbede2;color:#47528f;padding:9px 6px;text-align:center;font-weight:700;font-size:13px;border:none;}
.summary-list{margin:8px 0 12px 0;width:70%;}
.summary-item{display:flex;justify-content:space-between;align-items:center;padding:7px 0;font-size:15px;color:#47528f;}
.summary-item.highlight{font-weight:700;}
.sum-label{font-weight:600;font-size:15px;}
.sum-value{font-weight:900;font-size:16px;direction:rtl;}
.notes-title{font-size:15px;font-weight:700;color:#47528f;margin-bottom:0;text-align:right;}
.notes-content{font-size:14px;color:#47528f;padding:10px 0 6px 0;text-align:right;}
.thank-section{margin-top:auto;text-align:center;padding:14px 10px 6px 10px;}
.thank-title{font-size:17px;font-weight:700;color:#47528f;margin-bottom:10px;}
.thank-text{font-size:13px;line-height:1.9;color:#47528f;margin-bottom:12px;}
.signature{font-size:16px;font-weight:700;color:#c97a3a;}
@media print{
  body{background:#fff;margin:0;padding:0;-webkit-print-color-adjust:exact;print-color-adjust:exact;}
  .page{width:210mm;height:297mm;overflow:visible;page-break-after:always;page-break-inside:avoid;}
  *{-webkit-print-color-adjust:exact !important;print-color-adjust:exact !important;}
}
</style>
</head>
<body>
<!-- PAGE 1 -->
<div class="page">
  <div class="content">
    <div class="header">
      <div class="header-right">
        <div class="header-company-ar">شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div class="header-company-en">Al-Amani for General Trading, Real Estate & Commercial Agencies LLC</div>
      </div>
      <div class="header-left">
        <div class="header-left-text">متخصصون في منظومات<br>الطاقة الشمسية والشبكات</div>
      </div>
    </div>
    <div class="info-row">
      <div class="info-item"><div class="info-label">اسم المشروع والموقع:</div><div class="info-value">${projectName || '---'}</div></div>
      <div class="info-item"><div class="info-label">اسم الزبون:</div><div class="info-value">${customerName}</div></div>
      <div class="info-item"><div class="info-label">رقم الهاتف:</div><div class="info-value">${customerPhone || '---'}</div></div>
      <div class="info-item"><div class="info-label">العنوان/المحافظة:</div><div class="info-value">${customerAddress || '---'}</div></div>
    </div>
    <div class="title-right">عرض السعر الاولي</div>
    <hr class="hr">
    <div class="title-right">تفاصيل المنتجات والخدمات</div>
    <table>
      <thead><tr>
        <th style="width:6%">NO.</th>
        <th style="width:60px">الصورة</th>
        <th style="width:30%;text-align:right;padding-right:12px;">البيان/المنتج/الخدمة</th>
        <th style="width:11%">الوحدة</th>
        <th style="width:9%">العدد</th>
        <th style="width:14%">السعر (د.ع)</th>
        <th style="width:14%">الاجمالي (د.ع)</th>
      </tr></thead>
      <tbody>${tableRows}</tbody>
    </table>
  </div>
</div>
<!-- PAGE 2 -->
<div class="page">
  <div class="content">
    <div class="header">
      <div class="header-right">
        <div class="header-company-ar">شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div class="header-company-en">Al-Amani for General Trading, Real Estate & Commercial Agencies LLC</div>
      </div>
      <div class="header-left">
        <div class="header-left-text">متخصصون في منظومات<br>الطاقة الشمسية والشبكات</div>
      </div>
    </div>
    <div class="info-row" style="width:85%;gap:12px;margin:14px 0 10px 0;">
      <div class="info-item"><div class="info-label">اسم الزبون</div><div class="info-value">${customerName}</div></div>
      <div class="info-item"><div class="info-label">رقم العرض:</div><div class="info-value">${quotationNumber}</div></div>
      <div class="info-item"><div class="info-label">التاريخ:</div><div class="info-value">${todayStr}</div></div>
    </div>
    <hr class="hr">
    <div class="title-right">ملخص المبالغ</div>
    <div class="summary-list">
      <div class="summary-item"><span class="sum-label">اجمالي قيمة العرض</span><span class="sum-value">د.ع ${grandTotal.toLocaleString('en-IQ')}</span></div>
      <div class="summary-item"><span class="sum-label">نسبة الخصم</span><span class="sum-value">${discountPercent}%</span></div>
      <div class="summary-item"><span class="sum-label">قيمة الخصم</span><span class="sum-value">د.ع ${discountValue.toLocaleString('en-IQ')}</span></div>
      <div class="summary-item highlight"><span class="sum-label">الصافي بعد الخصم</span><span class="sum-value">د.ع ${netTotal.toLocaleString('en-IQ')}</span></div>
    </div>
    <hr class="hr-light">
    <div class="notes-title">ملاحظات خاصة بالعرض:</div>
    <hr class="hr-light" style="margin:6px 0 8px 0;">
    <div class="notes-content">${notes || 'لا توجد ملاحظات'}</div>
    <hr class="hr-light">
    <div class="title-center">شروط واحكام عرض السعر</div>
    <ol style="list-style:none;padding:0;margin:8px 0;">${termsHtml}</ol>
    <div class="thank-section">
      <div class="thank-title">كلمة شكر وتقدير</div>
      <div class="thank-text">تتقدم شركة الأماني للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية بخالص الشكر والتقدير على ثقتكم الكريمة، ونتطلع إلى أن نكون عند حسن ظنكم في تقديم أفضل الحلول التقنية والخدمات الهندسية المتخصصة. نؤكد لكم أننا نضع رضاكم في مقدمة أولوياتنا.</div>
      <div class="signature">مع خالص التحية والاحترام - إدارة شركة الأماني</div>
    </div>
  </div>
</div>
</body></html>`

    printWindow.document.write(html)
    printWindow.document.close()
    printWindow.onload = () => { printWindow.print() }
  }

  /* ───── input ref for autocomplete positioning ───── */
  const nameInputRef = useCallback((el: HTMLInputElement | null, index: number) => {
    if (el) {
      el.dataset.rowIndex = String(index)
    }
  }, [])

  const positionAc = (inputEl: HTMLInputElement) => {
    const r = inputEl.getBoundingClientRect()
    setAcPos({ top: r.bottom + window.scrollY, left: r.left + window.scrollX, width: r.width })
  }

  /* ───── styles ───── */
  const inputStyle: React.CSSProperties = {
    width: '100%', padding: '12px 14px', border: '1.5px solid #e0e0e0', borderRadius: 8,
    fontSize: 14, fontFamily: 'inherit', background: '#fafafa', outline: 'none',
  }
  const tableInputStyle: React.CSSProperties = {
    width: '100%', padding: 8, border: '1.5px solid #e0e0e0', borderRadius: 6,
    textAlign: 'center', fontSize: 13, fontFamily: 'inherit', background: '#fafafa', outline: 'none',
  }

  return (
    <div style={{ maxWidth: 1100, margin: '0 auto', background: '#fff', borderRadius: 12, boxShadow: '0 8px 30px rgba(0,0,0,.08)', overflow: 'hidden' }}>
      {/* ── Header ── */}
      <div style={{ background: `linear-gradient(135deg, ${BLUE}, ${BLUE2})`, color: 'white', padding: 35, textAlign: 'center' }}>
        <div style={{ fontSize: 32, fontWeight: 800, marginBottom: 4 }}>شركة الأماني</div>
        <div style={{ fontSize: 16, fontWeight: 600, opacity: 0.95, color: '#ffecb3' }}>للتجارة العامة والاستثمارات العقارية والوكالات التجارية محدودة المسؤولية</div>
        <div style={{ marginTop: 14, fontSize: 14, fontWeight: 600, borderTop: '1px solid rgba(255,255,255,.25)', display: 'inline-block', paddingTop: 10, opacity: 0.9 }}>نظام إصدار عروض الأسعار الرسمية</div>
      </div>

      <div style={{ padding: 40 }}>
        {/* ── Quotation Info ── */}
        <div style={{ background: '#fff', border: '1px solid #e0e0e0', borderRadius: 10, padding: 28, marginBottom: 28 }}>
          <div style={{ fontSize: 15, color: BLUE, fontWeight: 700, marginBottom: 22, paddingBottom: 12, borderBottom: '2px solid #e8eaf6' }}>معلومات العرض الأساسية</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 20 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>رقم العرض *</label>
              <input type="text" readOnly value={quotationNumber} style={{ ...inputStyle, background: '#eceff1', fontWeight: 700, color: BLUE }} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>التاريخ</label>
              <input type="date" readOnly value={todayStr} style={inputStyle} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>مدة التنفيذ *</label>
              <input type="text" value={duration} onChange={(e) => setDuration(e.target.value)} placeholder="مثال: 10 أيام عمل" style={inputStyle} />
            </div>
          </div>
        </div>

        {/* ── Customer Card ── */}
        <div style={{ background: '#fff', border: `2px solid ${BLUE}`, borderRadius: 12, marginBottom: 28, overflow: 'hidden' }}>
          <div style={{ background: `linear-gradient(90deg, ${BLUE}, ${BLUE2})`, color: '#fff', padding: '16px 28px', display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{ fontSize: 22 }}>&#127970;</span>
            <span style={{ fontSize: 15, fontWeight: 700 }}>بيانات الزبون</span>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)' }}>
            {[
              { label: 'اسم الزبون *', value: customerName, set: setCustomerName, placeholder: 'الاسم الكامل...' },
              { label: 'رقم الهاتف', value: customerPhone, set: setCustomerPhone, placeholder: '0770-XXX-XXXX' },
              { label: 'العنوان / المحافظة', value: customerAddress, set: setCustomerAddress, placeholder: 'مثال: كربلاء' },
            ].map((f, i) => (
              <div key={i} style={{ padding: '20px 24px', borderLeft: i < 2 ? '1px solid #e8eaf6' : 'none', borderBottom: '1px solid #e8eaf6', display: 'flex', flexDirection: 'column', gap: 6 }}>
                <label style={{ fontSize: 11, fontWeight: 700, color: BLUE }}>{f.label}</label>
                <input type="text" value={f.value} onChange={(e) => f.set(e.target.value)} placeholder={f.placeholder}
                  style={{ border: 'none', borderBottom: '2px solid #bdbdbd', borderRadius: 0, padding: '8px 0', fontSize: 14, background: 'transparent', fontWeight: 600, outline: 'none', fontFamily: 'inherit' }} />
              </div>
            ))}
            <div style={{ gridColumn: '1/-1', borderTop: '1px solid #e8eaf6', background: '#f5f5f5', padding: '20px 24px', display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontSize: 11, fontWeight: 700, color: BLUE }}>اسم المشروع والموقع</label>
              <input type="text" value={projectName} onChange={(e) => setProjectName(e.target.value)} placeholder="مثال: منظومة طاقة شمسية..."
                style={{ border: 'none', borderBottom: '2px solid #bdbdbd', borderRadius: 0, padding: '8px 0', fontSize: 14, background: 'transparent', fontWeight: 600, outline: 'none', fontFamily: 'inherit' }} />
            </div>
          </div>
        </div>

        {/* ── Products Table ── */}
        <div style={{ background: '#fff', border: '1px solid #e0e0e0', borderRadius: 10, padding: 28, marginBottom: 28 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 22, paddingBottom: 14, borderBottom: '2px solid #e8eaf6', flexWrap: 'wrap', gap: 12 }}>
            <div style={{ fontSize: 15, color: BLUE, fontWeight: 700 }}>تفاصيل المنتجات والخدمات</div>
            <button type="button" onClick={addRow} style={{ padding: '10px 22px', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 13, fontWeight: 700, fontFamily: 'inherit', background: '#2e7d32', color: 'white', display: 'inline-flex', alignItems: 'center', gap: 6 }}>+ إضافة بند جديد</button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'separate', borderSpacing: 0, marginBottom: 15, border: '1px solid #e0e0e0', borderRadius: 8, overflow: 'hidden' }}>
              <thead>
                <tr>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '5%' }}>م</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '28%' }}>البيان / المنتج / الخدمة</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '13%' }}>الوحدة</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '13%' }}>العدد</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '15%' }}>السعر (د.ع)</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '17%' }}>الإجمالي (د.ع)</th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '5%' }}></th>
                  <th style={{ background: BLUE, color: 'white', padding: '14px 10px', fontSize: 12, fontWeight: 700, textAlign: 'center', width: '8%' }}>الصورة</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={index} style={{ borderBottom: '1px solid #eee' }}>
                    <td style={{ padding: '12px 10px', textAlign: 'center', fontWeight: 700, color: BLUE, fontSize: 13 }}>{index + 1}</td>
                    <td style={{ padding: '12px 10px', position: 'relative' }}>
                      <input
                        type="text"
                        value={item.productName}
                        onChange={(e) => handleNameInput(index, e.target.value)}
                        onFocus={(e) => { if (item.productName.trim()) handleNameInput(index, item.productName); positionAc(e.currentTarget) }}
                        onBlur={() => setTimeout(() => { setAcIndex(null); setAcMatches([]) }, 180)}
                        placeholder="ابحث عن منتج..."
                        autoComplete="off"
                        style={{ ...tableInputStyle, textAlign: 'right' }}
                        ref={(el) => { if (el) el.dataset.rowIndex = String(index) }}
                      />
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center' }}>
                      <input type="text" value={item.unit} onChange={(e) => updateItem(index, { unit: e.target.value })} style={{ ...tableInputStyle, width: 80 }} />
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center' }}>
                      <input type="number" min="1" value={item.quantity} onChange={(e) => updateItem(index, { quantity: Number(e.target.value) })} style={{ ...tableInputStyle, width: 80 }} />
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center' }}>
                      <input type="number" min="0" value={item.unitPrice} onChange={(e) => updateItem(index, { unitPrice: Number(e.target.value) })} style={{ ...tableInputStyle, width: 110 }} />
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center' }}>
                      <input type="text" readOnly value={item.totalPrice.toLocaleString('en-IQ')} style={{ ...tableInputStyle, background: '#eceff1', fontWeight: 700, color: BLUE, borderColor: '#cfd8dc' }} />
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center' }}>
                      <button type="button" onClick={() => removeRow(index)} style={{ background: '#c62828', color: 'white', padding: '6px 14px', fontSize: 12, borderRadius: 6, border: 'none', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit' }}>حذف</button>
                    </td>
                    <td style={{ padding: '12px 10px', textAlign: 'center', verticalAlign: 'middle' }}>
                      {item.imageBase64 ? (
                        <img src={item.imageBase64} alt="" style={{ width: 50, height: 50, objectFit: 'contain', borderRadius: 6, border: '1px solid #e0e0e0', background: '#f9f9f9', cursor: 'pointer' }} />
                      ) : (
                        <div style={{ width: 50, height: 50, background: '#e8eaf6', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 24, margin: '0 auto' }}>&#128230;</div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Autocomplete Dropdown */}
          {acIndex !== null && acMatches.length > 0 && (
            <div style={{
              position: 'fixed', background: '#fff', border: `1.5px solid ${BLUE}`, borderRadius: '0 0 10px 10px',
              boxShadow: '0 8px 25px rgba(26,35,126,.2)', maxHeight: 250, overflowY: 'auto', zIndex: 9999,
              direction: 'rtl', top: acPos.top, left: acPos.left, width: acPos.width,
            }}>
              {acMatches.map((p) => (
                <div key={p.id} onMouseDown={(e) => { e.preventDefault(); selectProduct(acIndex, p) }}
                  style={{ padding: '9px 12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 10, borderBottom: '1px solid #f5f5f5' }}>
                  {p.imageBase64 ? (
                    <img src={p.imageBase64} alt="" style={{ width: 38, height: 38, objectFit: 'contain', borderRadius: 6, border: '1px solid #e0e0e0', flexShrink: 0 }} />
                  ) : (
                    <div style={{ width: 38, height: 38, background: '#e8eaf6', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 20, flexShrink: 0 }}>&#128230;</div>
                  )}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, color: BLUE, fontSize: 13, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{p.name}</div>
                    <div style={{ fontSize: 11, color: '#757575' }}>{p.defaultPrice.toLocaleString('en-IQ')} د.ع  {p.unit}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* ── Discount Card ── */}
        <div style={{ background: 'linear-gradient(135deg, #fff8e1, #ffecb3)', border: '2px solid #ffa000', borderRadius: 10, padding: 28, marginBottom: 28 }}>
          <div style={{ fontSize: 15, color: '#e65100', fontWeight: 700, marginBottom: 22, paddingBottom: 12, borderBottom: '2px solid #ffe082' }}>الخصم والصافي</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', maxWidth: 650, margin: '0 auto', gap: 20 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>نسبة الخصم (%)</label>
              <input type="number" min="0" max="100" value={discountPercent} onChange={(e) => setDiscountPercent(Number(e.target.value))} style={inputStyle} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>قيمة الخصم (د.ع)</label>
              <input type="text" readOnly value={discountValue.toLocaleString('en-IQ') + ' د.ع'} style={{ ...inputStyle, background: '#fff', fontWeight: 700, color: '#c62828', textAlign: 'center' }} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>الصافي بعد الخصم (د.ع)</label>
              <input type="text" readOnly value={netTotal.toLocaleString('en-IQ') + ' د.ع'} style={{ ...inputStyle, background: '#fff', fontWeight: 700, color: '#2e7d32', textAlign: 'center' }} />
            </div>
          </div>
        </div>

        {/* ── Notes ── */}
        <div style={{ background: '#fff', border: '1px solid #e0e0e0', borderRadius: 10, padding: 28, marginBottom: 28 }}>
          <div style={{ fontSize: 15, color: BLUE, fontWeight: 700, marginBottom: 22, paddingBottom: 12, borderBottom: '2px solid #e8eaf6' }}>ملاحظات خاصة بالعرض</div>
          <textarea value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="اكتب هنا أي ملاحظات إضافية..."
            style={{ ...inputStyle, resize: 'vertical', minHeight: 90 }} />
        </div>

        {/* ── Grand Total ── */}
        <div style={{ background: `linear-gradient(135deg, ${BLUE}, ${BLUE2})`, color: 'white', padding: 30, borderRadius: 12, textAlign: 'center', marginTop: 25 }}>
          <div style={{ fontSize: 14, fontWeight: 600, opacity: 0.9, marginBottom: 10 }}>الإجمالي الكلي للعرض</div>
          <div style={{ fontSize: 40, fontWeight: 800, color: GOLD }}>{grandTotal.toLocaleString('en-IQ')} د.ع</div>
        </div>

        {/* ── Actions Bar ── */}
        <div style={{ display: 'flex', justifyContent: 'center', gap: 18, marginTop: 30, flexWrap: 'wrap' }}>
          <button type="button" onClick={() => navigate('/products')}
            style={{ padding: '16px 32px', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 15, fontWeight: 700, fontFamily: 'inherit', background: 'linear-gradient(135deg, #00695c, #00796b)', color: 'white', boxShadow: '0 4px 12px rgba(0,105,92,.25)', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            &#128230; إدارة المنتجات
          </button>
          <button type="button" onClick={handleSave} disabled={submitting}
            style={{ padding: '16px 40px', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 15, fontWeight: 700, fontFamily: 'inherit', background: `linear-gradient(135deg, ${BLUE}, ${BLUE2})`, color: 'white', boxShadow: '0 4px 12px rgba(26,35,126,.25)', display: 'inline-flex', alignItems: 'center', gap: 6, opacity: submitting ? 0.5 : 1 }}>
            &#128190; {submitting ? 'جاري الحفظ...' : 'حفظ العرض'}
          </button>
          <button type="button" onClick={handlePrint}
            style={{ padding: '16px 40px', border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 15, fontWeight: 700, fontFamily: 'inherit', background: 'linear-gradient(135deg, #e65100, #ef6c00)', color: 'white', boxShadow: '0 4px 12px rgba(230,81,0,.25)', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            &#128424; طباعة العرض الرسمي
          </button>
        </div>

        {/* ── Status Message ── */}
        {statusMsg && (
          <div style={{
            padding: 16, borderRadius: 8, marginTop: 25, textAlign: 'center', fontWeight: 700, fontSize: 14,
            border: '1.5px solid',
            background: statusMsg.type === 'ok' ? '#e8f5e9' : '#ffebee',
            color: statusMsg.type === 'ok' ? '#2e7d32' : '#c62828',
            borderColor: statusMsg.type === 'ok' ? '#a5d6a7' : '#ef9a9a',
          }}>
            {statusMsg.text}
          </div>
        )}
      </div>
    </div>
  )
}
