import { useEffect, useState } from 'react'
import { api, type ProductRequest } from '../api'
import { useSession } from '../session'

export default function ProductRequestsPage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || employee?.role === 'PROCUREMENT_ADMIN' || permissions.includes('unit_technicians')
  const isAdmin = employee?.role === 'ADMIN'

  const [items, setItems] = useState<ProductRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ productName: '', specs: '', source: '', model: '', category: '', price: '' })
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)

  const load = () => { api.getProductRequests().then(setItems).finally(() => setLoading(false)) }
  useEffect(load, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.productName.trim()) return
    setSaving(true)
    try {
      await api.createProductRequest({
        productName: form.productName.trim(),
        specs: form.specs.trim() || undefined,
        source: form.source.trim() || undefined,
        model: form.model.trim() || undefined,
        category: form.category.trim() || undefined,
        price: form.price ? Number(form.price) : undefined,
      })
      setForm({ productName: '', specs: '', source: '', model: '', category: '', price: '' })
      setShowForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة الطلب')
    } finally {
      setSaving(false)
    }
  }

  const resolve = async (id: string, approve: boolean) => {
    setBusyId(id)
    try {
      if (approve) await api.approveProductRequest(id)
      else await api.rejectProductRequest(id)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const statusLabel: Record<string, string> = { PENDING: '⏳ بالانتظار', APPROVED: '✔ موافق عليه', REJECTED: '✘ مرفوض' }
  const statusColor: Record<string, string> = { PENDING: 'text-amber-600', APPROVED: 'text-emerald-600', REJECTED: 'text-red-600' }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة المنتجات</h2>
      <p className="mt-1 text-slate-500">اقتراح منتج جديد يُضاف لكتالوج النظام — يراجعه المدير ويوافق أو يرفض.</p>

      {canAdd && (
        <div className="mt-4 mb-4">
          <button onClick={() => setShowForm((v) => !v)} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600">
            {showForm ? '× إغلاق' : '+ طلب منتج جديد'}
          </button>
          {showForm && (
            <form onSubmit={handleCreate} className="mt-3 grid grid-cols-1 gap-3 rounded-xl border border-white bg-white p-5 shadow-sm sm:grid-cols-3">
              <input required value={form.productName} onChange={(e) => setForm({ ...form, productName: e.target.value })} placeholder="اسم المنتج" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-3" />
              <input value={form.specs} onChange={(e) => setForm({ ...form, specs: e.target.value })} placeholder="المواصفات" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.source} onChange={(e) => setForm({ ...form, source: e.target.value })} placeholder="المصدر" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.model} onChange={(e) => setForm({ ...form, model: e.target.value })} placeholder="الموديل" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} placeholder="التصنيف/الدورة" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <input value={form.price} type="number" onChange={(e) => setForm({ ...form, price: e.target.value })} placeholder="السعر" className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
              <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50 sm:col-span-3">
                {saving ? 'جاري الحفظ...' : 'إرسال الطلب'}
              </button>
            </form>
          )}
        </div>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && items.length === 0 && <p className="text-slate-400">لا توجد طلبات منتجات بعد.</p>}

      <div className="overflow-x-auto rounded-xl border border-white bg-white shadow-sm">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="bg-brand-900 text-white">
              <th className="p-2 text-right">المنتج</th>
              <th className="p-2 text-right">المواصفات</th>
              <th className="p-2 text-right">المصدر</th>
              <th className="p-2 text-right">الموديل</th>
              <th className="p-2 text-right">التصنيف</th>
              <th className="p-2 text-right">السعر</th>
              <th className="p-2 text-right">الحالة</th>
              <th className="p-2 text-right">طلبه</th>
              {isAdmin && <th className="p-2 text-right">إجراء</th>}
            </tr>
          </thead>
          <tbody>
            {items.map((p) => (
              <tr key={p.id} className="border-b border-slate-100">
                <td className="p-2 font-bold">{p.productName}</td>
                <td className="p-2 text-slate-500">{p.specs || '—'}</td>
                <td className="p-2 text-slate-500">{p.source || '—'}</td>
                <td className="p-2 text-slate-500">{p.model || '—'}</td>
                <td className="p-2 text-slate-500">{p.category || '—'}</td>
                <td className="p-2 text-slate-500">{p.price != null ? p.price.toLocaleString('en-IQ') : '—'}</td>
                <td className={`p-2 font-bold ${statusColor[p.status]}`}>{statusLabel[p.status]}</td>
                <td className="p-2 text-slate-400">{p.requestedBy?.name || '—'}</td>
                {isAdmin && (
                  <td className="p-2">
                    {p.status === 'PENDING' && (
                      <div className="flex gap-1">
                        <button onClick={() => resolve(p.id, true)} disabled={busyId === p.id} className="rounded bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700 hover:bg-emerald-100">✔</button>
                        <button onClick={() => resolve(p.id, false)} disabled={busyId === p.id} className="rounded bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">✘</button>
                      </div>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
