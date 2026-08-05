import { useEffect, useState, useRef } from 'react'
import { api, type Product, type ProductAvailability, fileUrl } from '../api'

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const fmt = (n: number) => n.toLocaleString('en-IQ')

// «الحاجة» — القيم بقت مثل ما هي بقاعدة البيانات، بس التسمية تغيّرت.
const NEED: { value: ProductAvailability; label: string }[] = [
  { value: 'ON_DEMAND', label: 'يحتاج أوفرها' },
  { value: 'IN_STOCK', label: 'ما يحتاج أوفرها' },
]
const needLabel = (v: ProductAvailability) => NEED.find((n) => n.value === v)?.label ?? v

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [defaultPrice, setDefaultPrice] = useState(0)
  const [wholesalePrice, setWholesalePrice] = useState(0)
  const [imageBase64, setImageBase64] = useState('')
  const [specs, setSpecs] = useState('')
  const [source, setSource] = useState('')
  const [modelName, setModelName] = useState('')
  const [availability, setAvailability] = useState<ProductAvailability>('ON_DEMAND')
  const [serviceText, setServiceText] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)

  // added=1 — هاي الشاشة تعرض بس المنتجات الي ضافها التقنيين،
  // مو كتالوج عرض السعر كامل. (المضاف هنا يوصل لعرض السعر برضه.)
  const load = () => {
    api.getAddedProducts()
      .then((rows) => setProducts(rows ?? []))
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(() => { load() }, [])

  const patch = async (id: string, data: Parameters<typeof api.updateProduct>[1]) => {
    setBusyId(id)
    try {
      const updated = await api.updateProduct(id, data)
      setProducts((prev) => prev.map((p) => (p.id === id ? updated : p)))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر التعديل')
    } finally {
      setBusyId(null)
    }
  }

  // الصورة تنرفع للتخزين وننحفظ رابطها بس — بدل ما نحشر base64
  // بقاعدة البيانات ونثقّل كل استعلام يجيب المنتج.
  const [uploading, setUploading] = useState(false)
  const handleImageChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const res = await api.uploadFile(file, 'products')
      setImageBase64(res.url)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفع الصورة')
      if (fileRef.current) fileRef.current.value = ''
    } finally {
      setUploading(false)
    }
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setSubmitting(true)
    try {
      await api.createProduct({
        name,
        defaultPrice,
        wholesalePrice: wholesalePrice || undefined,
        imageBase64: imageBase64 || undefined,
        availability,
        serviceText: serviceText.trim() || undefined,
        specs: specs.trim() || undefined,
        source: source.trim() || undefined,
        modelName: modelName.trim() || undefined,
      })
      setName(''); setDefaultPrice(0); setWholesalePrice(0); setImageBase64('')
      setSpecs(''); setSource(''); setModelName(''); setServiceText('')
      setAvailability('ON_DEMAND')
      if (fileRef.current) fileRef.current.value = ''
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف المنتج؟')) return
    try {
      await api.deleteProduct(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const inputStyle: React.CSSProperties = {
    width: '100%', padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px',
    fontSize: '14px', boxSizing: 'border-box', outline: 'none',
  }
  const labelStyle: React.CSSProperties = {
    display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666',
  }

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '24px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إضافة منتج</h1>
        <span style={{ color: GOLD, fontSize: '14px' }}>
          المنتجات الي يضيفها التقنيين — وتظهر تلقائياً بعرض السعر
        </span>
      </div>

      <form
        onSubmit={handleAdd}
        style={{
          background: 'white', border: `2px solid ${PRIMARY}`, borderRadius: '12px',
          padding: '20px', marginBottom: '24px',
        }}
      >
        <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
          إضافة منتج جديد
        </h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', alignItems: 'end' }}>
          {/* الاسم والموديل بصف واحد */}
          <div>
            <label style={labelStyle}>اسم المنتج *</label>
            <input required value={name} onChange={(e) => setName(e.target.value)} style={inputStyle} />
          </div>
          <div>
            <label style={labelStyle}>الموديل</label>
            <input value={modelName} onChange={(e) => setModelName(e.target.value)} style={inputStyle} placeholder="رقم/اسم الموديل" />
          </div>
          <div>
            <label style={labelStyle}>السعر الافتراضي</label>
            <input type="number" min="0" value={defaultPrice} onChange={(e) => setDefaultPrice(Number(e.target.value))} style={inputStyle} />
          </div>
          <div>
            <label style={labelStyle}>سعر الجملة</label>
            <input type="number" min="0" value={wholesalePrice} onChange={(e) => setWholesalePrice(Number(e.target.value))} style={inputStyle} />
          </div>
          <div>
            <label style={labelStyle}>المصدر</label>
            <input value={source} onChange={(e) => setSource(e.target.value)} style={inputStyle} placeholder="المجهز أو بلد المنشأ" />
          </div>
          {/* الخدمة يكتبها التقني بنفسه — ماكو قائمة وماكو اقتراح نظام */}
          <div>
            <label style={labelStyle}>الخدمة</label>
            <input value={serviceText} onChange={(e) => setServiceText(e.target.value)} style={inputStyle} placeholder="اكتب الخدمة الي يخصها المنتج" />
          </div>
          <div>
            <label style={labelStyle}>الحاجة *</label>
            <select value={availability} onChange={(e) => setAvailability(e.target.value as ProductAvailability)} style={inputStyle}>
              {NEED.map((n) => (
                <option key={n.value} value={n.value}>{n.label}</option>
              ))}
            </select>
          </div>
          <div>
            <label style={labelStyle}>صورة المنتج</label>
            <input ref={fileRef} type="file" accept="image/*" onChange={handleImageChange} style={{ ...inputStyle, padding: '7px 12px' }} />
            {uploading && <span style={{ fontSize: '12px', color: '#666' }}>جاري رفع الصورة...</span>}
          </div>
        </div>

        {/* المواصفات خانة كبيرة بصف لحالها */}
        <div style={{ marginTop: '16px' }}>
          <label style={labelStyle}>المواصفات</label>
          <textarea
            value={specs}
            onChange={(e) => setSpecs(e.target.value)}
            rows={5}
            style={{ ...inputStyle, resize: 'vertical', lineHeight: 1.7 }}
            placeholder="اكتب مواصفات المنتج كاملة — مثال: 4MP · ليلي ملون · عدسة 2.8mm · مقاوم للماء"
          />
        </div>

        {imageBase64 && (
          <div style={{ marginTop: '12px' }}>
            <img src={fileUrl(imageBase64)} alt="معاينة" style={{ width: '80px', height: '80px', objectFit: 'cover', borderRadius: '8px', border: `2px solid ${GOLD}` }} />
          </div>
        )}

        <button
          type="submit"
          disabled={submitting || uploading}
          style={{
            marginTop: '16px', width: '100%', background: PRIMARY, color: 'white', border: 'none',
            padding: '12px', borderRadius: '8px', cursor: submitting ? 'not-allowed' : 'pointer',
            fontWeight: 'bold', fontSize: '14px', opacity: submitting ? 0.6 : 1,
          }}
        >
          {submitting ? 'جاري الإضافة...' : 'إضافة منتج'}
        </button>
      </form>

      {loading && <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: '#dc2626' }}>
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: '16px' }}>
          {products.map((product) => (
            <div
              key={product.id}
              style={{ background: 'white', border: '2px solid #e0e0e0', borderRadius: '12px', overflow: 'hidden' }}
            >
              <div style={{
                height: '140px',
                background: product.imageBase64 ? 'none' : `linear-gradient(135deg, ${PRIMARY}, #283593)`,
                display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden',
              }}>
                {product.imageBase64 ? (
                  <img src={fileUrl(product.imageBase64)} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                ) : (
                  <span style={{ color: GOLD, fontSize: '40px' }}>&#128230;</span>
                )}
              </div>
              <div style={{ padding: '12px 16px' }}>
                <h4 style={{ margin: '0 0 6px 0', color: PRIMARY, fontSize: '15px' }}>
                  {product.name}
                  {product.modelName && <span style={{ color: '#777', fontWeight: 'normal' }}> — {product.modelName}</span>}
                </h4>
                {product.serviceText && <p style={{ margin: '0 0 2px 0', fontSize: '12px', color: '#777' }}>الخدمة: {product.serviceText}</p>}
                {product.specs && <p style={{ margin: '0 0 2px 0', fontSize: '12px', color: '#777', whiteSpace: 'pre-wrap' }}>المواصفات: {product.specs}</p>}
                {product.source && <p style={{ margin: '0 0 4px 0', fontSize: '12px', color: '#777' }}>المصدر: {product.source}</p>}
                <p style={{ margin: '0 0 4px 0', fontSize: '16px', fontWeight: 'bold', color: GOLD }}>
                  {fmt(product.defaultPrice ?? 0)} د.ع
                </p>
                {product.wholesalePrice != null && (
                  <p style={{ margin: '0 0 10px 0', fontSize: '12px', color: '#999' }}>
                    سعر الجملة: {fmt(product.wholesalePrice)} د.ع
                  </p>
                )}
                {product.createdByName && (
                  <p style={{ margin: '0 0 8px 0', fontSize: '12px', color: '#555' }}>
                    👤 أضافه: {product.createdByName}
                  </p>
                )}

                {/* الحاجة — زر واحد يقلب بين الخيارين */}
                <button
                  disabled={busyId === product.id}
                  onClick={() => patch(product.id, {
                    availability: product.availability === 'IN_STOCK' ? 'ON_DEMAND' : 'IN_STOCK',
                  })}
                  title="اضغط لتبديل الحاجة"
                  style={{
                    width: '100%', marginBottom: '8px', padding: '7px', borderRadius: '6px',
                    border: 'none', cursor: 'pointer', fontSize: '12px', fontWeight: 'bold',
                    background: product.availability === 'IN_STOCK' ? '#dcfce7' : '#fef3c7',
                    color: product.availability === 'IN_STOCK' ? '#15803d' : '#a16207',
                    opacity: busyId === product.id ? 0.5 : 1,
                  }}
                >
                  {needLabel(product.availability)}
                </button>

                <button
                  onClick={() => handleDelete(product.id)}
                  style={{
                    width: '100%', background: '#fee2e2', color: '#dc2626', border: 'none',
                    padding: '8px', borderRadius: '6px', cursor: 'pointer', fontSize: '13px', fontWeight: 'bold',
                  }}
                >
                  حذف المنتج
                </button>
              </div>
            </div>
          ))}
          {products.length === 0 && (
            <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '40px', color: '#999' }}>
              ماكو منتجات مضافة بعد — أضف أول منتج من الفورم فوك
            </div>
          )}
        </div>
      )}
    </div>
  )
}
