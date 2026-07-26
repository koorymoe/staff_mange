import { useEffect, useState, useRef } from 'react'
import { api, type Product } from '../api'

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const fmt = (n: number) => n.toLocaleString('en-IQ')

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [unit, setUnit] = useState('قطعة')
  const [defaultPrice, setDefaultPrice] = useState(0)
  const [wholesalePrice, setWholesalePrice] = useState(0)
  const [imageBase64, setImageBase64] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const load = () => {
    api.getProducts()
      .then(setProducts)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    const reader = new FileReader()
    reader.onload = () => setImageBase64(reader.result as string)
    reader.readAsDataURL(file)
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setSubmitting(true)
    try {
      await api.createProduct({
        name,
        unit: unit || 'قطعة',
        defaultPrice,
        wholesalePrice: wholesalePrice || undefined,
        imageBase64: imageBase64 || undefined,
      })
      setName('')
      setUnit('قطعة')
      setDefaultPrice(0)
      setWholesalePrice(0)
      setImageBase64('')
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

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      {/* Header */}
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white',
        padding: '20px 30px',
        borderRadius: '12px',
        marginBottom: '24px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إدارة المنتجات</h1>
        <span style={{ color: GOLD, fontSize: '14px' }}>إضافة وإدارة المنتجات والأسعار الافتراضية</span>
      </div>

      {/* Add Product Form */}
      <form
        onSubmit={handleAdd}
        style={{
          background: 'white',
          border: `2px solid ${PRIMARY}`,
          borderRadius: '12px',
          padding: '20px',
          marginBottom: '24px',
        }}
      >
        <h3 style={{ color: PRIMARY, margin: '0 0 16px 0', borderBottom: `2px solid ${GOLD}`, paddingBottom: '8px' }}>
          إضافة منتج جديد
        </h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))', gap: '16px', alignItems: 'end' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>اسم المنتج *</label>
            <input
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              style={inputStyle}
              onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
              onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
            />
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>الوحدة</label>
            <input
              value={unit}
              onChange={(e) => setUnit(e.target.value)}
              style={inputStyle}
              placeholder="قطعة"
              onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
              onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
            />
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>السعر الافتراضي</label>
            <input
              type="number"
              min="0"
              value={defaultPrice}
              onChange={(e) => setDefaultPrice(Number(e.target.value))}
              style={inputStyle}
              onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
              onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
            />
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>سعر الجملة</label>
            <input
              type="number"
              min="0"
              value={wholesalePrice}
              onChange={(e) => setWholesalePrice(Number(e.target.value))}
              style={inputStyle}
              onFocus={(e) => { e.target.style.borderColor = PRIMARY }}
              onBlur={(e) => { e.target.style.borderColor = '#ddd' }}
            />
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>صورة المنتج</label>
            <input
              ref={fileRef}
              type="file"
              accept="image/*"
              onChange={handleImageChange}
              style={{ ...inputStyle, padding: '7px 12px' }}
            />
          </div>
          <div>
            <button
              type="submit"
              disabled={submitting}
              style={{
                width: '100%',
                background: PRIMARY,
                color: 'white',
                border: 'none',
                padding: '12px',
                borderRadius: '8px',
                cursor: submitting ? 'not-allowed' : 'pointer',
                fontWeight: 'bold',
                fontSize: '14px',
                opacity: submitting ? 0.6 : 1,
              }}
            >
              {submitting ? 'جاري الإضافة...' : 'إضافة منتج'}
            </button>
          </div>
        </div>
        {imageBase64 && (
          <div style={{ marginTop: '12px' }}>
            <img src={imageBase64} alt="معاينة" style={{ width: '80px', height: '80px', objectFit: 'cover', borderRadius: '8px', border: `2px solid ${GOLD}` }} />
          </div>
        )}
      </form>

      {loading && <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: '#dc2626' }}>
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
          gap: '16px',
        }}>
          {products.map((product) => (
            <div
              key={product.id}
              style={{
                background: 'white',
                border: '2px solid #e0e0e0',
                borderRadius: '12px',
                overflow: 'hidden',
                transition: 'border-color 0.2s',
              }}
              onMouseEnter={(e) => { e.currentTarget.style.borderColor = PRIMARY }}
              onMouseLeave={(e) => { e.currentTarget.style.borderColor = '#e0e0e0' }}
            >
              {/* Product Image */}
              <div style={{
                height: '140px',
                background: product.imageBase64 ? 'none' : `linear-gradient(135deg, ${PRIMARY}, #283593)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                overflow: 'hidden',
              }}>
                {product.imageBase64 ? (
                  <img src={product.imageBase64} alt={product.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                ) : (
                  <span style={{ color: GOLD, fontSize: '40px' }}>&#128230;</span>
                )}
              </div>
              {/* Product Info */}
              <div style={{ padding: '12px 16px' }}>
                <h4 style={{ margin: '0 0 6px 0', color: PRIMARY, fontSize: '15px' }}>{product.name}</h4>
                <p style={{ margin: '0 0 4px 0', fontSize: '13px', color: '#666' }}>الوحدة: {product.unit || 'قطعة'}</p>
                <p style={{
                  margin: '0 0 4px 0', fontSize: '16px', fontWeight: 'bold',
                  color: GOLD,
                }}>
                  {fmt(product.defaultPrice ?? 0)} د.ع
                </p>
                {product.wholesalePrice != null && (
                  <p style={{ margin: '0 0 10px 0', fontSize: '12px', color: '#999' }}>
                    سعر الجملة: {fmt(product.wholesalePrice)} د.ع
                  </p>
                )}
                <button
                  onClick={() => handleDelete(product.id)}
                  style={{
                    width: '100%',
                    background: '#fee2e2',
                    color: '#dc2626',
                    border: 'none',
                    padding: '8px',
                    borderRadius: '6px',
                    cursor: 'pointer',
                    fontSize: '13px',
                    fontWeight: 'bold',
                  }}
                >
                  حذف المنتج
                </button>
              </div>
            </div>
          ))}
          {products.length === 0 && (
            <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '40px', color: '#999' }}>
              لا توجد منتجات بعد
            </div>
          )}
        </div>
      )}
    </div>
  )
}
