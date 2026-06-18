import { useEffect, useState, useRef } from 'react'
import { api, type Product } from '../api'

const BLUE = '#1a237e'
const BLUE2 = '#283593'

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [name, setName] = useState('')
  const [unit, setUnit] = useState('قطعة')
  const [defaultPrice, setDefaultPrice] = useState(0)
  const [imageBase64, setImageBase64] = useState('')
  const [imgPreview, setImgPreview] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [statusMsg, setStatusMsg] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  const load = () => {
    setLoading(true)
    api.getProducts().then(setProducts).catch(() => {}).finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    if (!f) return
    const reader = new FileReader()
    reader.onload = (ev) => {
      const img = new Image()
      img.onload = () => {
        const MAX = 220
        let w = img.width, h = img.height
        if (w > h) { if (w > MAX) { h = Math.round(h * MAX / w); w = MAX } }
        else { if (h > MAX) { w = Math.round(w * MAX / h); h = MAX } }
        const cv = document.createElement('canvas')
        cv.width = w; cv.height = h
        cv.getContext('2d')!.drawImage(img, 0, 0, w, h)
        const b64 = cv.toDataURL('image/jpeg', 0.75)
        setImageBase64(b64)
        setImgPreview(b64)
      }
      img.src = ev.target?.result as string
    }
    reader.readAsDataURL(f)
  }

  const resetForm = () => {
    setName(''); setDefaultPrice(0); setUnit('قطعة'); setImageBase64(''); setImgPreview('')
    if (fileRef.current) fileRef.current.value = ''
  }

  const handleAdd = async () => {
    if (!name.trim()) { setStatusMsg('أدخل اسم المنتج'); return }
    setSubmitting(true)
    setStatusMsg('جاري الحفظ...')
    try {
      await api.createProduct({ name, unit, defaultPrice, imageBase64: imageBase64 || undefined })
      resetForm()
      load()
      setStatusMsg('تم حفظ المنتج بنجاح')
    } catch (e) {
      setStatusMsg(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('حذف المنتج؟')) return
    try {
      await api.deleteProduct(id)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%', padding: '11px 13px', border: '1.5px solid #e0e0e0', borderRadius: 8,
    fontSize: 14, fontFamily: 'inherit', background: '#fff', outline: 'none', boxSizing: 'border-box',
  }

  return (
    <div style={{ maxWidth: 1100, margin: '0 auto' }}>
      {/* Header */}
      <div style={{ background: `linear-gradient(135deg, ${BLUE}, ${BLUE2})`, color: '#fff', padding: '20px 26px', borderRadius: '14px 14px 0 0', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700 }}>&#128230; إدارة قاعدة بيانات المنتجات</h3>
      </div>

      <div style={{ background: '#fff', borderRadius: '0 0 14px 14px', padding: 24, boxShadow: '0 20px 60px rgba(0,0,0,.08)' }}>
        {/* Add Form */}
        <div style={{ background: '#f5f7ff', border: '1px solid #e8eaf6', borderRadius: 10, padding: 20, marginBottom: 24 }}>
          <div style={{ fontSize: 14, fontWeight: 700, color: BLUE, marginBottom: 14, paddingBottom: 10, borderBottom: '2px solid #e8eaf6' }}>+ إضافة منتج جديد</div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>اسم المنتج *</label>
              <input type="text" value={name} onChange={(e) => setName(e.target.value)} placeholder="مثال: لوح شمسي 400 واط" style={inputStyle} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>وحدة القياس</label>
              <input type="text" value={unit} onChange={(e) => setUnit(e.target.value)} style={inputStyle} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>السعر الافتراضي (د.ع)</label>
              <input type="number" min="0" value={defaultPrice} onChange={(e) => setDefaultPrice(Number(e.target.value))} style={inputStyle} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <label style={{ fontWeight: 700, color: '#424242', fontSize: 13 }}>صورة المنتج</label>
              <input type="file" accept="image/*" ref={fileRef} onChange={handleImageChange}
                style={{ ...inputStyle, border: `1.5px dashed ${BLUE}`, cursor: 'pointer', fontSize: 13 }} />
            </div>
          </div>
          {imgPreview && (
            <div style={{ textAlign: 'center', marginTop: 10 }}>
              <img src={imgPreview} alt="" style={{ width: 80, height: 80, objectFit: 'contain', border: '1.5px solid #e0e0e0', borderRadius: 8 }} />
            </div>
          )}
          <button onClick={handleAdd} disabled={submitting}
            style={{ width: '100%', padding: 13, marginTop: 14, background: `linear-gradient(135deg, ${BLUE}, ${BLUE2})`, color: '#fff', border: 'none', borderRadius: 8, fontSize: 14, fontWeight: 700, cursor: 'pointer', fontFamily: 'inherit', opacity: submitting ? 0.6 : 1 }}>
            &#128190; حفظ المنتج في القاعدة
          </button>
          {statusMsg && <div style={{ textAlign: 'center', fontSize: 13, fontWeight: 600, marginTop: 10, color: statusMsg.includes('خطأ') || statusMsg.includes('أدخل') ? '#c62828' : statusMsg.includes('جاري') ? BLUE : '#2e7d32' }}>{statusMsg}</div>}
        </div>

        {/* Products Grid */}
        <div>
          <div style={{ fontSize: 14, fontWeight: 700, color: BLUE, marginBottom: 14, paddingBottom: 10, borderBottom: '2px solid #e8eaf6', display: 'flex', alignItems: 'center', gap: 10 }}>
            المنتجات المخزونة
            <span style={{ background: BLUE, color: '#fff', padding: '2px 12px', borderRadius: 20, fontSize: 12 }}>{products.length}</span>
          </div>

          {loading ? (
            <div style={{ textAlign: 'center', color: '#9e9e9e', padding: 30 }}>جاري التحميل...</div>
          ) : products.length === 0 ? (
            <div style={{ textAlign: 'center', color: '#9e9e9e', padding: 30 }}>
              <div style={{ fontSize: 36, marginBottom: 8 }}>&#128230;</div>لا توجد منتجات بعد!
            </div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(130px, 1fr))', gap: 12 }}>
              {products.map((p) => (
                <div key={p.id} style={{ background: '#fff', border: '1px solid #e0e0e0', borderRadius: 10, overflow: 'hidden', textAlign: 'center' }}>
                  {p.imageBase64 ? (
                    <img src={p.imageBase64} alt="" style={{ width: '100%', height: 90, objectFit: 'contain', background: '#f9f9f9', padding: 5, boxSizing: 'border-box', display: 'block' }} />
                  ) : (
                    <div style={{ width: '100%', height: 90, background: 'linear-gradient(135deg, #e8eaf6, #f5f7ff)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 32 }}>&#128230;</div>
                  )}
                  <div style={{ padding: '8px 6px' }}>
                    <div style={{ fontWeight: 700, fontSize: 12, color: BLUE }}>{p.name}</div>
                    <div style={{ fontSize: 10, color: '#757575' }}>{p.unit}</div>
                    <div style={{ fontSize: 11, color: '#2e7d32', fontWeight: 700, marginTop: 2 }}>{p.defaultPrice.toLocaleString('en-IQ')} د.ع</div>
                    <button onClick={() => handleDelete(p.id)}
                      style={{ background: '#c62828', color: '#fff', border: 'none', padding: '5px 0', borderRadius: 5, cursor: 'pointer', fontSize: 11, fontFamily: 'inherit', marginTop: 6, width: 'calc(100% - 12px)' }}>
                      &#128465; حذف
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
