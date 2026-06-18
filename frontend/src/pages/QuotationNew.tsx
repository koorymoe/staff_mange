import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type Product } from '../api'
import { useSession } from '../session'

interface ItemRow {
  productName: string
  unit: string
  quantity: number
  unitPrice: number
  totalPrice: number
}

const emptyItem = (): ItemRow => ({
  productName: '',
  unit: '',
  quantity: 1,
  unitPrice: 0,
  totalPrice: 0,
})

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const fmt = (n: number) => n.toLocaleString('en-IQ')

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
  const [activeAutocomplete, setActiveAutocomplete] = useState<number | null>(null)
  const [autocompleteFilter, setAutocompleteFilter] = useState('')
  const autocompleteRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    api.getProducts().then(setProducts).catch(() => {})
  }, [])

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (autocompleteRef.current && !autocompleteRef.current.contains(e.target as Node)) {
        setActiveAutocomplete(null)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const today = new Date().toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long', day: 'numeric' })

  const grandTotal = items.reduce((sum, item) => sum + item.totalPrice, 0)
  const discountValue = (grandTotal * discountPercent) / 100
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
      unit: product.unit,
      unitPrice: product.defaultPrice,
    })
    setActiveAutocomplete(null)
    setAutocompleteFilter('')
  }

  const addRow = () => setItems((prev) => [...prev, emptyItem()])
  const removeRow = (index: number) => {
    if (items.length <= 1) return
    setItems((prev) => prev.filter((_, i) => i !== index))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!customerName.trim()) return alert('يرجى إدخال اسم الزبون')
    if (items.every((it) => !it.productName.trim())) return alert('يرجى إضافة منتج واحد على الأقل')

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
      navigate('/quotations')
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handlePrint = () => {
    const printWindow = window.open('', '_blank')
    if (!printWindow) return
    const validItems = items.filter((it) => it.productName.trim())
    const itemsRows = validItems.map((item, i) => `
      <tr>
        <td style="border:1px solid #ddd;padding:8px;text-align:center;">${i + 1}</td>
        <td style="border:1px solid #ddd;padding:8px;text-align:right;">${item.productName}</td>
        <td style="border:1px solid #ddd;padding:8px;text-align:center;">${item.unit}</td>
        <td style="border:1px solid #ddd;padding:8px;text-align:center;">${fmt(item.quantity)}</td>
        <td style="border:1px solid #ddd;padding:8px;text-align:center;">${fmt(item.unitPrice)}</td>
        <td style="border:1px solid #ddd;padding:8px;text-align:center;font-weight:bold;">${fmt(item.totalPrice)}</td>
      </tr>
    `).join('')

    printWindow.document.write(`<!DOCTYPE html><html dir="rtl" lang="ar"><head><meta charset="UTF-8">
      <title>عرض سعر</title>
      <style>
        @page { size: A4; margin: 15mm; }
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; color: #333; margin: 0; padding: 20px; }
        .page { page-break-after: always; }
        .header { background: ${PRIMARY}; color: white; padding: 20px 30px; border-radius: 8px; margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center; }
        .header h1 { margin: 0; font-size: 22px; }
        .header .gold { color: ${GOLD}; font-size: 14px; }
        .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; margin-bottom: 20px; }
        .info-box { border: 2px solid ${PRIMARY}; border-radius: 8px; padding: 12px 16px; }
        .info-box h3 { color: ${PRIMARY}; margin: 0 0 8px 0; font-size: 14px; border-bottom: 2px solid ${GOLD}; padding-bottom: 6px; }
        .info-box p { margin: 4px 0; font-size: 13px; }
        .info-box span.label { color: #666; }
        .info-box span.value { font-weight: bold; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
        thead th { background: ${PRIMARY}; color: white; padding: 10px 8px; font-size: 13px; }
        .totals-box { background: #f8f9fa; border: 2px solid ${PRIMARY}; border-radius: 8px; padding: 16px; }
        .totals-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 14px; }
        .totals-row.net { font-size: 18px; font-weight: bold; color: ${PRIMARY}; border-top: 2px solid ${GOLD}; margin-top: 8px; padding-top: 10px; }
        .terms { margin-top: 20px; }
        .terms h3 { color: ${PRIMARY}; border-bottom: 2px solid ${GOLD}; padding-bottom: 6px; }
        .terms ul { padding-right: 20px; font-size: 13px; line-height: 1.8; }
        .footer { text-align: center; margin-top: 30px; padding-top: 15px; border-top: 2px solid ${GOLD}; font-size: 12px; color: #666; }
      </style></head><body>
      <div class="page">
        <div class="header">
          <div><h1>الأماني - نظام عروض الأسعار الرسمية</h1><span class="gold">عرض سعر رسمي</span></div>
          <div style="text-align:left;"><div style="font-size:13px;">التاريخ: ${today}</div></div>
        </div>
        <div class="info-grid">
          <div class="info-box">
            <h3>بيانات العميل</h3>
            <p><span class="label">الاسم: </span><span class="value">${customerName}</span></p>
            <p><span class="label">الهاتف: </span><span class="value">${customerPhone || '-'}</span></p>
            <p><span class="label">العنوان: </span><span class="value">${customerAddress || '-'}</span></p>
            <p><span class="label">المشروع: </span><span class="value">${projectName || '-'}</span></p>
          </div>
          <div class="info-box">
            <h3>بيانات العرض</h3>
            <p><span class="label">مدة التنفيذ: </span><span class="value">${duration || '-'}</span></p>
          </div>
        </div>
        <table>
          <thead><tr>
            <th style="width:40px;">#</th><th>اسم المادة</th><th>الوحدة</th><th>الكمية</th><th>سعر الوحدة</th><th>المجموع</th>
          </tr></thead>
          <tbody>${itemsRows}</tbody>
        </table>
        <div class="totals-box">
          <div class="totals-row"><span>المجموع الكلي:</span><span>${fmt(grandTotal)} د.ع</span></div>
          <div class="totals-row"><span>الخصم (${discountPercent}%):</span><span>${fmt(discountValue)} د.ع</span></div>
          <div class="totals-row net"><span>الصافي:</span><span>${fmt(netTotal)} د.ع</span></div>
        </div>
      </div>
      <div>
        <div class="header">
          <div><h1>الأماني - ملخص عرض السعر</h1><span class="gold">الشروط والأحكام</span></div>
        </div>
        <div class="totals-box" style="margin-bottom:20px;">
          <div class="totals-row"><span>اسم العميل:</span><span>${customerName}</span></div>
          <div class="totals-row"><span>المشروع:</span><span>${projectName || '-'}</span></div>
          <div class="totals-row"><span>عدد الأصناف:</span><span>${validItems.length}</span></div>
          <div class="totals-row"><span>المجموع الكلي:</span><span>${fmt(grandTotal)} د.ع</span></div>
          <div class="totals-row"><span>الخصم:</span><span>${fmt(discountValue)} د.ع</span></div>
          <div class="totals-row net"><span>المبلغ الصافي:</span><span>${fmt(netTotal)} د.ع</span></div>
        </div>
        ${notes ? `<div class="info-box" style="margin-bottom:20px;"><h3>ملاحظات</h3><p>${notes}</p></div>` : ''}
        <div class="terms">
          <h3>الشروط والأحكام</h3>
          <ul>
            <li>الأسعار المذكورة أعلاه لا تشمل أجور النقل والتركيب ما لم يُذكر خلاف ذلك.</li>
            <li>عرض السعر ساري المفعول لمدة 15 يوم من تاريخه.</li>
            <li>الدفع: 50% مقدم والباقي عند التسليم.</li>
            <li>مدة التنفيذ تبدأ من تاريخ استلام الدفعة الأولى.</li>
            <li>الأسعار قابلة للتغيير حسب تقلبات السوق.</li>
          </ul>
        </div>
        <div class="footer">
          <p>الأماني للمقاولات والخدمات | هاتف: 0770 000 0000</p>
        </div>
      </div>
      </body></html>`)
    printWindow.document.close()
    printWindow.print()
  }

  const filteredProducts = products.filter(p =>
    p.name.includes(autocompleteFilter)
  )

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      {/* Header */}
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white',
        padding: '20px 30px',
        borderRadius: '12px',
        marginBottom: '24px',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
      }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '24px' }}>الأماني - نظام عروض الأسعار الرسمية</h1>
          <span style={{ color: GOLD, fontSize: '14px' }}>عرض سعر جديد</span>
        </div>
        <button
          onClick={() => navigate('/quotations')}
          style={{
            background: 'rgba(255,255,255,0.15)',
            border: '1px solid rgba(255,255,255,0.3)',
            color: 'white',
            padding: '8px 20px',
            borderRadius: '8px',
            cursor: 'pointer',
            fontSize: '14px',
          }}
        >
          رجوع لعروض الأسعار
        </button>
      </div>

      <form onSubmit={handleSubmit}>
        {/* Quote Info Card */}
        <div style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
            بيانات العرض
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>رقم العرض</label>
              <input
                disabled
                placeholder="تلقائي"
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  background: '#f5f5f5', fontSize: '14px', boxSizing: 'border-box' as const,
                }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>التاريخ</label>
              <input
                disabled
                value={today}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  background: '#f5f5f5', fontSize: '14px', boxSizing: 'border-box' as const,
                }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>مدة التنفيذ</label>
              <input
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                placeholder="مثال: 30 يوم"
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
          </div>
        </div>

        {/* Customer Info Card */}
        <div style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
            بيانات العميل
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>اسم العميل *</label>
              <input
                required
                value={customerName}
                onChange={(e) => setCustomerName(e.target.value)}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>رقم الهاتف</label>
              <input
                value={customerPhone}
                onChange={(e) => setCustomerPhone(e.target.value)}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>العنوان / المحافظة</label>
              <input
                value={customerAddress}
                onChange={(e) => setCustomerAddress(e.target.value)}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>اسم المشروع</label>
              <input
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
          </div>
        </div>

        {/* Products Table */}
        <div style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
            <h3 style={{ color: PRIMARY, margin: 0 }}>جدول المواد والأصناف</h3>
            <button
              type="button"
              onClick={addRow}
              style={{
                background: GOLD,
                color: PRIMARY,
                border: 'none',
                padding: '8px 20px',
                borderRadius: '8px',
                cursor: 'pointer',
                fontWeight: 'bold',
                fontSize: '14px',
              }}
            >
              + إضافة صنف
            </button>
          </div>
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ background: PRIMARY, color: 'white' }}>
                  <th style={{ padding: '10px 8px', fontSize: '13px', width: '40px' }}>#</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px', textAlign: 'right' }}>اسم المادة</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px' }}>الوحدة</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px' }}>الكمية</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px' }}>سعر الوحدة</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px' }}>المجموع</th>
                  <th style={{ padding: '10px 8px', fontSize: '13px', width: '50px' }}></th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, index) => (
                  <tr key={index} style={{ borderBottom: '1px solid #eee' }}>
                    <td style={{ padding: '8px', textAlign: 'center', fontWeight: 'bold', color: PRIMARY }}>{index + 1}</td>
                    <td style={{ padding: '8px', position: 'relative' }}>
                      <div ref={activeAutocomplete === index ? autocompleteRef : undefined} style={{ position: 'relative' }}>
                        <input
                          value={item.productName}
                          onChange={(e) => {
                            updateItem(index, { productName: e.target.value })
                            setAutocompleteFilter(e.target.value)
                            setActiveAutocomplete(index)
                          }}
                          onFocus={() => {
                            setActiveAutocomplete(index)
                            setAutocompleteFilter(item.productName)
                          }}
                          placeholder="اكتب اسم المنتج..."
                          style={{
                            width: '100%', padding: '8px 10px', border: '1px solid #ddd', borderRadius: '6px',
                            fontSize: '13px', boxSizing: 'border-box' as const, outline: 'none',
                          }}
                        />
                        {activeAutocomplete === index && filteredProducts.length > 0 && (
                          <div style={{
                            position: 'absolute', top: '100%', right: 0, left: 0,
                            background: 'white', border: `1px solid ${PRIMARY}`, borderRadius: '6px',
                            maxHeight: '180px', overflowY: 'auto', zIndex: 100,
                            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
                          }}>
                            {filteredProducts.map((p) => (
                              <div
                                key={p.id}
                                onClick={() => handleProductSelect(index, p)}
                                style={{
                                  padding: '8px 12px', cursor: 'pointer', fontSize: '13px',
                                  borderBottom: '1px solid #f0f0f0', display: 'flex', justifyContent: 'space-between',
                                }}
                                onMouseEnter={(e) => { e.currentTarget.style.background = '#e8eaf6' }}
                                onMouseLeave={(e) => { e.currentTarget.style.background = 'white' }}
                              >
                                <span>{p.name}</span>
                                <span style={{ color: '#888', fontSize: '12px' }}>{fmt(p.defaultPrice)} د.ع</span>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </td>
                    <td style={{ padding: '8px' }}>
                      <input
                        value={item.unit}
                        onChange={(e) => updateItem(index, { unit: e.target.value })}
                        style={{
                          width: '80px', padding: '8px 10px', border: '1px solid #ddd', borderRadius: '6px',
                          fontSize: '13px', textAlign: 'center', outline: 'none',
                        }}
                      />
                    </td>
                    <td style={{ padding: '8px' }}>
                      <input
                        type="number"
                        min="1"
                        value={item.quantity}
                        onChange={(e) => updateItem(index, { quantity: Number(e.target.value) })}
                        style={{
                          width: '70px', padding: '8px 10px', border: '1px solid #ddd', borderRadius: '6px',
                          fontSize: '13px', textAlign: 'center', outline: 'none',
                        }}
                      />
                    </td>
                    <td style={{ padding: '8px' }}>
                      <input
                        type="number"
                        min="0"
                        value={item.unitPrice}
                        onChange={(e) => updateItem(index, { unitPrice: Number(e.target.value) })}
                        style={{
                          width: '110px', padding: '8px 10px', border: '1px solid #ddd', borderRadius: '6px',
                          fontSize: '13px', textAlign: 'center', outline: 'none',
                        }}
                      />
                    </td>
                    <td style={{ padding: '8px', textAlign: 'center', fontWeight: 'bold', color: PRIMARY, fontSize: '14px' }}>
                      {fmt(item.totalPrice)} د.ع
                    </td>
                    <td style={{ padding: '8px', textAlign: 'center' }}>
                      <button
                        type="button"
                        onClick={() => removeRow(index)}
                        style={{
                          background: '#ef4444', color: 'white', border: 'none',
                          width: '28px', height: '28px', borderRadius: '6px',
                          cursor: 'pointer', fontSize: '16px', lineHeight: '1',
                        }}
                      >
                        x
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Discount & Totals */}
        <div style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
            الخصم والإجماليات
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '16px' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>المجموع الكلي</label>
              <div style={{
                background: '#e8eaf6', padding: '12px', borderRadius: '8px',
                fontWeight: 'bold', fontSize: '16px', color: PRIMARY, textAlign: 'center',
              }}>
                {fmt(grandTotal)} د.ع
              </div>
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>نسبة الخصم %</label>
              <input
                type="number"
                min="0"
                max="100"
                value={discountPercent}
                onChange={(e) => setDiscountPercent(Number(e.target.value))}
                style={{
                  width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
                  fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none', textAlign: 'center',
                }}
                onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
                onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
              />
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>قيمة الخصم</label>
              <div style={{
                background: '#fef2f2', padding: '12px', borderRadius: '8px',
                fontWeight: 'bold', fontSize: '16px', color: '#dc2626', textAlign: 'center',
              }}>
                {fmt(discountValue)} د.ع
              </div>
            </div>
            <div>
              <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>الصافي</label>
              <div style={{
                background: GOLD, padding: '12px', borderRadius: '8px',
                fontWeight: 'bold', fontSize: '18px', color: PRIMARY, textAlign: 'center',
              }}>
                {fmt(netTotal)} د.ع
              </div>
            </div>
          </div>
        </div>

        {/* Notes */}
        <div style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '20px',
        }}>
          <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
            ملاحظات
          </h3>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={4}
            placeholder="أضف ملاحظات إضافية هنا..."
            style={{
              width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
              fontSize: '14px', boxSizing: 'border-box' as const, outline: 'none', resize: 'vertical',
            }}
            onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
            onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
          />
        </div>

        {/* Action Buttons */}
        <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
          <button
            type="button"
            onClick={() => navigate('/products')}
            style={{
              background: 'white',
              color: PRIMARY,
              border: `2px solid ${PRIMARY}`,
              padding: '12px 24px',
              borderRadius: '8px',
              cursor: 'pointer',
              fontWeight: 'bold',
              fontSize: '14px',
            }}
          >
            إدارة المنتجات
          </button>
          <button
            type="submit"
            disabled={submitting}
            style={{
              background: PRIMARY,
              color: 'white',
              border: 'none',
              padding: '12px 32px',
              borderRadius: '8px',
              cursor: submitting ? 'not-allowed' : 'pointer',
              fontWeight: 'bold',
              fontSize: '14px',
              opacity: submitting ? 0.6 : 1,
            }}
          >
            {submitting ? 'جاري الحفظ...' : 'حفظ'}
          </button>
          <button
            type="button"
            onClick={handlePrint}
            style={{
              background: GOLD,
              color: PRIMARY,
              border: 'none',
              padding: '12px 32px',
              borderRadius: '8px',
              cursor: 'pointer',
              fontWeight: 'bold',
              fontSize: '14px',
            }}
          >
            طباعة
          </button>
        </div>
      </form>
    </div>
  )
}
