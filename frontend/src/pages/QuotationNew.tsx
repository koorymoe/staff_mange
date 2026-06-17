import { useEffect, useState } from 'react'
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

  useEffect(() => {
    api.getProducts().then(setProducts).catch(() => {})
  }, [])

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

  const handleProductSelect = (index: number, productName: string) => {
    const product = products.find((p) => p.name === productName)
    if (product) {
      updateItem(index, {
        productName: product.name,
        unit: product.unit,
        unitPrice: product.defaultPrice,
      })
    } else {
      updateItem(index, { productName })
    }
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
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-brand-900">عرض سعر جديد</h2>
        <button
          onClick={() => navigate('/quotations')}
          className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
        >
          رجوع
        </button>
      </div>

      <form onSubmit={handleSubmit} className="mt-6 space-y-6">
        {/* Customer Info */}
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-4 text-lg font-bold text-brand-800">بيانات الزبون</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم الزبون *</label>
              <input
                required
                value={customerName}
                onChange={(e) => setCustomerName(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف</label>
              <input
                value={customerPhone}
                onChange={(e) => setCustomerPhone(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">العنوان</label>
              <input
                value={customerAddress}
                onChange={(e) => setCustomerAddress(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">اسم المشروع</label>
              <input
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
          </div>
        </div>

        {/* Items Table */}
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-lg font-bold text-brand-800">المواد والخدمات</h3>
            <button
              type="button"
              onClick={addRow}
              className="rounded-lg bg-brand-50 px-4 py-2 text-sm font-medium text-brand-700 hover:bg-brand-100"
            >
              + إضافة سطر
            </button>
          </div>
          <table className="w-full text-right">
            <thead className="border-b border-slate-200">
              <tr>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600">المنتج</th>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600">الوحدة</th>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600">الكمية</th>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600">سعر الوحدة</th>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600">المجموع</th>
                <th className="px-3 py-2 text-sm font-semibold text-slate-600"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {items.map((item, index) => (
                <tr key={index}>
                  <td className="px-3 py-2">
                    <input
                      list={`products-${index}`}
                      value={item.productName}
                      onChange={(e) => handleProductSelect(index, e.target.value)}
                      className="w-full rounded border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                      placeholder="اكتب أو اختر منتج"
                    />
                    <datalist id={`products-${index}`}>
                      {products.map((p) => (
                        <option key={p.id} value={p.name} />
                      ))}
                    </datalist>
                  </td>
                  <td className="px-3 py-2">
                    <input
                      value={item.unit}
                      onChange={(e) => updateItem(index, { unit: e.target.value })}
                      className="w-24 rounded border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                    />
                  </td>
                  <td className="px-3 py-2">
                    <input
                      type="number"
                      min="1"
                      value={item.quantity}
                      onChange={(e) => updateItem(index, { quantity: Number(e.target.value) })}
                      className="w-20 rounded border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                    />
                  </td>
                  <td className="px-3 py-2">
                    <input
                      type="number"
                      min="0"
                      value={item.unitPrice}
                      onChange={(e) => updateItem(index, { unitPrice: Number(e.target.value) })}
                      className="w-28 rounded border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                    />
                  </td>
                  <td className="px-3 py-2 font-medium">
                    {item.totalPrice.toLocaleString()} د.ع
                  </td>
                  <td className="px-3 py-2">
                    <button
                      type="button"
                      onClick={() => removeRow(index)}
                      className="text-sm text-red-500 hover:text-red-700"
                    >
                      حذف
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Totals */}
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-4 text-lg font-bold text-brand-800">الإجماليات</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">المجموع الكلي</label>
              <p className="rounded-lg bg-slate-50 px-3 py-2 font-bold text-brand-900">
                {grandTotal.toLocaleString()} د.ع
              </p>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">نسبة الخصم %</label>
              <input
                type="number"
                min="0"
                max="100"
                value={discountPercent}
                onChange={(e) => setDiscountPercent(Number(e.target.value))}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">قيمة الخصم</label>
              <p className="rounded-lg bg-slate-50 px-3 py-2 font-bold text-red-600">
                {discountValue.toLocaleString()} د.ع
              </p>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الصافي</label>
              <p className="rounded-lg bg-brand-50 px-3 py-2 font-bold text-brand-900">
                {netTotal.toLocaleString()} د.ع
              </p>
            </div>
          </div>
        </div>

        {/* Duration & Notes */}
        <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">مدة التنفيذ</label>
              <input
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
                placeholder="مثال: 30 يوم"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظات</label>
              <textarea
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              />
            </div>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-8 py-2.5 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'حفظ'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/quotations')}
            className="rounded-lg border border-slate-300 px-6 py-2.5 font-medium text-slate-700 hover:bg-slate-50"
          >
            رجوع
          </button>
        </div>
      </form>
    </div>
  )
}
