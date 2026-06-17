import { useEffect, useState } from 'react'
import { api, type Product } from '../api'

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [unit, setUnit] = useState('')
  const [defaultPrice, setDefaultPrice] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState('')
  const [editUnit, setEditUnit] = useState('')
  const [editPrice, setEditPrice] = useState(0)

  const load = () => {
    setLoading(true)
    api.getProducts()
      .then(setProducts)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return
    setSubmitting(true)
    try {
      await api.createProduct({ name, unit, defaultPrice })
      setName('')
      setUnit('')
      setDefaultPrice(0)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const startEdit = (product: Product) => {
    setEditingId(product.id)
    setEditName(product.name)
    setEditUnit(product.unit)
    setEditPrice(product.defaultPrice)
  }

  const handleUpdate = async () => {
    if (!editingId) return
    try {
      await api.updateProduct(editingId, { name: editName, unit: editUnit, defaultPrice: editPrice })
      setEditingId(null)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف المنتج؟')) return
    try {
      await api.deleteProduct(id)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة المنتجات</h2>
      <p className="mt-1 text-slate-500">إدارة قائمة المنتجات والأسعار الافتراضية لعروض الأسعار.</p>

      <form
        onSubmit={handleAdd}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">اسم المنتج</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الوحدة</label>
          <input
            value={unit}
            onChange={(e) => setUnit(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
            placeholder="متر، قطعة، ..."
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">السعر الافتراضي</label>
          <input
            type="number"
            min="0"
            value={defaultPrice}
            onChange={(e) => setDefaultPrice(Number(e.target.value))}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="flex items-end">
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الإضافة...' : 'إضافة منتج'}
          </button>
        </div>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">اسم المنتج</th>
                <th className="px-4 py-3 text-sm font-semibold">الوحدة</th>
                <th className="px-4 py-3 text-sm font-semibold">السعر الافتراضي</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {products.map((product) => (
                <tr key={product.id} className="hover:bg-slate-50">
                  {editingId === product.id ? (
                    <>
                      <td className="px-4 py-3">
                        <input
                          value={editName}
                          onChange={(e) => setEditName(e.target.value)}
                          className="w-full rounded border border-slate-300 px-2 py-1 text-sm"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <input
                          value={editUnit}
                          onChange={(e) => setEditUnit(e.target.value)}
                          className="w-24 rounded border border-slate-300 px-2 py-1 text-sm"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <input
                          type="number"
                          value={editPrice}
                          onChange={(e) => setEditPrice(Number(e.target.value))}
                          className="w-28 rounded border border-slate-300 px-2 py-1 text-sm"
                        />
                      </td>
                      <td className="px-4 py-3 flex gap-2">
                        <button
                          onClick={handleUpdate}
                          className="text-sm text-brand-700 hover:text-brand-900"
                        >
                          حفظ
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="text-sm text-slate-500 hover:text-slate-700"
                        >
                          إلغاء
                        </button>
                      </td>
                    </>
                  ) : (
                    <>
                      <td className="px-4 py-3 font-medium">{product.name}</td>
                      <td className="px-4 py-3">{product.unit}</td>
                      <td className="px-4 py-3">{product.defaultPrice.toLocaleString()} د.ع</td>
                      <td className="px-4 py-3 flex gap-3">
                        <button
                          onClick={() => startEdit(product)}
                          className="text-sm text-brand-700 hover:text-brand-900"
                        >
                          تعديل
                        </button>
                        <button
                          onClick={() => handleDelete(product.id)}
                          className="text-sm text-red-600 hover:text-red-800"
                        >
                          حذف
                        </button>
                      </td>
                    </>
                  )}
                </tr>
              ))}
              {products.length === 0 && (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-center text-slate-400">
                    لا توجد منتجات بعد
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
