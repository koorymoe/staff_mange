import { useEffect, useState, useCallback } from 'react'
import { useSession } from '../session'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('authToken')
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options?.headers || {}),
    },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

interface Specialty {
  id: string
  name: string
}

interface Rating {
  id: string
  value: number
  note: string | null
  ratedByName: string
  createdAt: string
}

interface Supplier {
  id: string
  companyName: string
  ownerName: string
  phone: string
  lat: number | null
  lng: number | null
  isMaterialSupplier: boolean
  isContractor: boolean
  traderTypes: string[]
  notes: string | null
  specialties: Specialty[]
  avgRating: number
  ratingCount: number
  ratings: Rating[]
}

const traderTypeLabels: Record<string, string> = {
  'وكيل': 'وكيل',
  'موزع': 'موزع',
  'تاجر': 'تاجر',
}

const traderTypeOptions = ['وكيل', 'موزع', 'تاجر']

function Stars({ value, size = 'text-lg' }: { value: number; size?: string }) {
  return (
    <span className={size}>
      {[1, 2, 3, 4, 5].map((i) => (
        <span key={i} className={i <= Math.round(value) ? 'text-yellow-400' : 'text-slate-300'}>★</span>
      ))}
    </span>
  )
}

function StarSelector({ value, onChange }: { value: number; onChange: (v: number) => void }) {
  return (
    <span className="text-2xl cursor-pointer">
      {[1, 2, 3, 4, 5].map((i) => (
        <span key={i} onClick={() => onChange(i)} className={i <= value ? 'text-yellow-400' : 'text-slate-300'}>★</span>
      ))}
    </span>
  )
}

export default function SuppliersPage() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN'

  const [suppliers, setSuppliers] = useState<Supplier[]>([])
  const [specialties, setSpecialties] = useState<Specialty[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Filters
  const [search, setSearch] = useState('')
  const [filterSpecialty, setFilterSpecialty] = useState('')
  const [filterType, setFilterType] = useState('')

  // Modals
  const [showForm, setShowForm] = useState(false)
  const [editingSupplier, setEditingSupplier] = useState<Supplier | null>(null)
  const [detailSupplier, setDetailSupplier] = useState<Supplier | null>(null)
  const [showRatingModal, setShowRatingModal] = useState(false)
  const [ratingTarget, setRatingTarget] = useState<Supplier | null>(null)

  // Form state
  const [formCompanyName, setFormCompanyName] = useState('')
  const [formOwnerName, setFormOwnerName] = useState('')
  const [formPhone, setFormPhone] = useState('')
  const [formLat, setFormLat] = useState('')
  const [formLng, setFormLng] = useState('')
  const [formIsMaterial, setFormIsMaterial] = useState(false)
  const [formIsContractor, setFormIsContractor] = useState(false)
  const [formTraderTypes, setFormTraderTypes] = useState<string[]>([])
  const [formNotes, setFormNotes] = useState('')
  const [formSpecialtyIds, setFormSpecialtyIds] = useState<string[]>([])
  const [formSubmitting, setFormSubmitting] = useState(false)

  // Rating form
  const [ratingValue, setRatingValue] = useState(0)
  const [ratingNote, setRatingNote] = useState('')
  const [ratingSubmitting, setRatingSubmitting] = useState(false)

  // Specialty management
  const [newSpecialtyName, setNewSpecialtyName] = useState('')

  const loadData = useCallback(() => {
    Promise.all([
      request<Supplier[]>('/suppliers'),
      request<Specialty[]>('/suppliers/specialties'),
    ])
      .then(([s, sp]) => { setSuppliers(s); setSpecialties(sp) })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => { loadData() }, [loadData])

  const filtered = suppliers.filter((s) => {
    if (search) {
      const q = search.toLowerCase()
      if (!s.companyName.toLowerCase().includes(q) && !s.ownerName.toLowerCase().includes(q) && !s.phone.includes(q)) return false
    }
    if (filterSpecialty && !s.specialties.some((sp) => sp.id === filterSpecialty)) return false
    if (filterType === 'material' && !s.isMaterialSupplier) return false
    if (filterType === 'contractor' && !s.isContractor) return false
    return true
  })

  const totalSpecialties = new Set(suppliers.flatMap((s) => s.specialties.map((sp) => sp.id))).size
  const avgRating = suppliers.length ? (suppliers.reduce((a, s) => a + s.avgRating, 0) / suppliers.length).toFixed(1) : '0'

  const resetForm = () => {
    setFormCompanyName(''); setFormOwnerName(''); setFormPhone('')
    setFormLat(''); setFormLng(''); setFormIsMaterial(false); setFormIsContractor(false)
    setFormTraderTypes([]); setFormNotes(''); setFormSpecialtyIds([])
  }

  const openAdd = () => {
    resetForm(); setEditingSupplier(null); setShowForm(true)
  }

  const openEdit = (s: Supplier) => {
    setFormCompanyName(s.companyName); setFormOwnerName(s.ownerName); setFormPhone(s.phone)
    setFormLat(s.lat?.toString() || ''); setFormLng(s.lng?.toString() || '')
    setFormIsMaterial(s.isMaterialSupplier); setFormIsContractor(s.isContractor)
    setFormTraderTypes(s.traderTypes || []); setFormNotes(s.notes || '')
    setFormSpecialtyIds(s.specialties.map((sp) => sp.id))
    setEditingSupplier(s); setShowForm(true)
  }

  const handleSubmitForm = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormSubmitting(true)
    const body = {
      companyName: formCompanyName,
      ownerName: formOwnerName,
      phone: formPhone,
      lat: formLat ? parseFloat(formLat) : null,
      lng: formLng ? parseFloat(formLng) : null,
      isMaterialSupplier: formIsMaterial,
      isContractor: formIsContractor,
      traderTypes: formTraderTypes,
      notes: formNotes || null,
      specialtyIds: formSpecialtyIds,
    }
    try {
      if (editingSupplier) {
        await request(`/suppliers/${editingSupplier.id}`, { method: 'PUT', body: JSON.stringify(body) })
      } else {
        await request('/suppliers', { method: 'POST', body: JSON.stringify(body) })
      }
      setShowForm(false); loadData()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally { setFormSubmitting(false) }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف هذا المورد؟')) return
    try {
      await request(`/suppliers/${id}`, { method: 'DELETE' })
      setDetailSupplier(null); loadData()
    } catch (err) { alert(err instanceof Error ? err.message : 'حدث خطأ') }
  }

  const handleRate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!ratingTarget || !ratingValue || !employee) return
    setRatingSubmitting(true)
    try {
      await request(`/suppliers/${ratingTarget.id}/rate`, {
        method: 'POST',
        body: JSON.stringify({ value: ratingValue, note: ratingNote || null, ratedById: employee.id, ratedByName: employee.name }),
      })
      setShowRatingModal(false); setRatingValue(0); setRatingNote('')
      loadData()
    } catch (err) { alert(err instanceof Error ? err.message : 'حدث خطأ') }
    finally { setRatingSubmitting(false) }
  }

  const handleAddSpecialty = async () => {
    if (!newSpecialtyName.trim()) return
    try {
      await request('/suppliers/specialties', { method: 'POST', body: JSON.stringify({ name: newSpecialtyName.trim() }) })
      setNewSpecialtyName(''); loadData()
    } catch (err) { alert(err instanceof Error ? err.message : 'حدث خطأ') }
  }

  const handleDeleteSpecialty = async (id: string) => {
    if (!confirm('هل أنت متأكد من حذف هذا التخصص؟')) return
    try {
      await request(`/suppliers/specialties/${id}`, { method: 'DELETE' })
      loadData()
    } catch (err) { alert(err instanceof Error ? err.message : 'حدث خطأ') }
  }

  const toggleTraderType = (t: string) => {
    setFormTraderTypes((prev) => prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t])
  }

  const toggleSpecialty = (id: string) => {
    setFormSpecialtyIds((prev) => prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id])
  }

  // Modal backdrop
  const Modal = ({ children, onClose }: { children: React.ReactNode; onClose: () => void }) => (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
    </div>
  )

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">الموردون</h2>
      <p className="mt-1 text-slate-500">إدارة الموردين والمقاولين وتقييمهم</p>

      {/* Stats bar */}
      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        {[
          { label: 'إجمالي الموردين', value: suppliers.length, color: 'text-brand-600' },
          { label: 'التخصصات', value: totalSpecialties, color: 'text-emerald-600' },
          { label: 'متوسط التقييم', value: avgRating, color: 'text-yellow-600' },
        ].map((stat) => (
          <div key={stat.label} className="rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <p className="text-sm text-slate-500">{stat.label}</p>
            <p className={`mt-1 text-2xl font-bold ${stat.color}`}>{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Search/filter bar */}
      <div className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4">
        <input
          placeholder="بحث (اسم الشركة، المالك، الهاتف)..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
        />
        <select value={filterSpecialty} onChange={(e) => setFilterSpecialty(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
          <option value="">كل التخصصات</option>
          {specialties.map((sp) => <option key={sp.id} value={sp.id}>{sp.name}</option>)}
        </select>
        <select value={filterType} onChange={(e) => setFilterType(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
          <option value="">كل الأنواع</option>
          <option value="material">مورد مواد</option>
          <option value="contractor">منفذ</option>
        </select>
        <button onClick={openAdd}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30">
          + إضافة مورد
        </button>
      </div>

      {/* Loading/error */}
      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {/* Supplier cards grid */}
      {!loading && !error && (
        <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((s) => (
            <div key={s.id}
              onClick={() => setDetailSupplier(s)}
              className="cursor-pointer rounded-xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] transition-all hover:shadow-lg hover:shadow-brand-900/10 hover:-translate-y-0.5">
              <div className="flex items-start justify-between">
                <h3 className="text-lg font-bold text-brand-800">{s.companyName}</h3>
                <Stars value={s.avgRating} size="text-sm" />
              </div>
              <div className="mt-2 flex flex-wrap gap-1">
                {s.specialties.map((sp) => (
                  <span key={sp.id} className="rounded-full bg-brand-50 px-2 py-0.5 text-xs font-medium text-brand-700">{sp.name}</span>
                ))}
              </div>
              <div className="mt-3 space-y-1 text-sm text-slate-600">
                <p><span className="text-slate-400">المالك: </span>{s.ownerName}</p>
                <p><span className="text-slate-400">الهاتف: </span><a href={`tel:${s.phone}`} className="text-brand-600 hover:underline" onClick={(e) => e.stopPropagation()}>{s.phone}</a></p>
              </div>
              <div className="mt-3 flex flex-wrap gap-1">
                {s.isMaterialSupplier && <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">مورد مواد</span>}
                {s.isContractor && <span className="rounded-full bg-orange-50 px-2 py-0.5 text-xs font-medium text-orange-700">منفذ</span>}
                {(s.traderTypes || []).map((t) => (
                  <span key={t} className="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">{traderTypeLabels[t] || t}</span>
                ))}
              </div>
              {s.lat && s.lng && (
                <a href={`https://www.google.com/maps?q=${s.lat},${s.lng}`} target="_blank" rel="noreferrer"
                  onClick={(e) => e.stopPropagation()}
                  className="mt-2 inline-block text-xs text-brand-500 hover:underline">
                  عرض الموقع على الخريطة
                </a>
              )}
            </div>
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-xl border border-white bg-white p-8 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-slate-400">لا يوجد موردين</p>
            </div>
          )}
        </div>
      )}

      {/* Specialties management (admin only) */}
      {isAdmin && (
        <div className="mt-8 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="text-lg font-bold text-brand-800">إدارة التخصصات</h3>
          <div className="mt-4 flex gap-2">
            <input
              placeholder="اسم التخصص الجديد..."
              value={newSpecialtyName}
              onChange={(e) => setNewSpecialtyName(e.target.value)}
              className="flex-1 rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
              onKeyDown={(e) => e.key === 'Enter' && handleAddSpecialty()}
            />
            <button onClick={handleAddSpecialty}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30">
              إضافة
            </button>
          </div>
          <div className="mt-4 overflow-hidden rounded-lg border border-slate-200">
            <div className="overflow-x-auto">
            <table className="w-full text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">التخصص</th>
                  <th className="px-4 py-3 text-sm font-semibold w-24">إجراء</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {specialties.map((sp) => (
                  <tr key={sp.id} className="hover:bg-slate-50">
                    <td className="px-4 py-3">{sp.name}</td>
                    <td className="px-4 py-3">
                      <button onClick={() => handleDeleteSpecialty(sp.id)} className="text-sm text-red-500 hover:text-red-700">حذف</button>
                    </td>
                  </tr>
                ))}
                {specialties.length === 0 && (
                  <tr><td colSpan={2} className="px-4 py-6 text-center text-slate-400">لا توجد تخصصات</td></tr>
                )}
              </tbody>
            </table>
            </div>
          </div>
        </div>
      )}

      {/* Add/Edit Modal */}
      {showForm && (
        <Modal onClose={() => setShowForm(false)}>
          <h3 className="text-xl font-bold text-brand-800 mb-4">{editingSupplier ? 'تعديل المورد' : 'إضافة مورد جديد'}</h3>
          <form onSubmit={handleSubmitForm} className="space-y-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">اسم الشركة</label>
                <input required value={formCompanyName} onChange={(e) => setFormCompanyName(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">اسم المالك</label>
                <input required value={formOwnerName} onChange={(e) => setFormOwnerName(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف</label>
                <input required value={formPhone} onChange={(e) => setFormPhone(e.target.value)}
                  className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="mb-1 block text-sm font-medium text-slate-600">خط العرض</label>
                  <input value={formLat} onChange={(e) => setFormLat(e.target.value)} placeholder="Lat"
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
                </div>
                <div>
                  <label className="mb-1 block text-sm font-medium text-slate-600">خط الطول</label>
                  <input value={formLng} onChange={(e) => setFormLng(e.target.value)} placeholder="Lng"
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
                </div>
              </div>
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-slate-600">التخصصات</label>
              <div className="flex flex-wrap gap-2">
                {specialties.map((sp) => (
                  <button key={sp.id} type="button" onClick={() => toggleSpecialty(sp.id)}
                    className={`rounded-full px-3 py-1 text-sm font-medium transition-all ${
                      formSpecialtyIds.includes(sp.id)
                        ? 'bg-brand-500 text-white'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    }`}>
                    {sp.name}
                  </button>
                ))}
              </div>
            </div>

            <div className="flex flex-wrap gap-4">
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={formIsMaterial} onChange={(e) => setFormIsMaterial(e.target.checked)}
                  className="h-4 w-4 rounded border-slate-300 text-brand-500" />
                مورد مواد
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={formIsContractor} onChange={(e) => setFormIsContractor(e.target.checked)}
                  className="h-4 w-4 rounded border-slate-300 text-brand-500" />
                منفذ
              </label>
            </div>

            <div>
              <label className="mb-2 block text-sm font-medium text-slate-600">نوع التاجر</label>
              <div className="flex flex-wrap gap-2">
                {traderTypeOptions.map((t) => (
                  <button key={t} type="button" onClick={() => toggleTraderType(t)}
                    className={`rounded-full px-3 py-1 text-sm font-medium transition-all ${
                      formTraderTypes.includes(t)
                        ? 'bg-blue-500 text-white'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    }`}>
                    {t}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظات</label>
              <textarea value={formNotes} onChange={(e) => setFormNotes(e.target.value)} rows={3}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
            </div>

            <div className="flex gap-3">
              <button type="submit" disabled={formSubmitting}
                className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50">
                {formSubmitting ? 'جاري الحفظ...' : editingSupplier ? 'تحديث' : 'حفظ'}
              </button>
              <button type="button" onClick={() => setShowForm(false)}
                className="rounded-lg border border-slate-300 px-6 py-2 font-medium text-slate-600 transition-all hover:bg-slate-50">
                إلغاء
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* Detail Modal */}
      {detailSupplier && (
        <Modal onClose={() => setDetailSupplier(null)}>
          <div className="flex items-start justify-between">
            <h3 className="text-xl font-bold text-brand-800">{detailSupplier.companyName}</h3>
            <Stars value={detailSupplier.avgRating} />
          </div>
          <p className="text-sm text-slate-400 mt-1">{detailSupplier.ratingCount} تقييم</p>

          <div className="mt-4 grid grid-cols-1 gap-2 text-sm sm:grid-cols-2">
            <p><span className="text-slate-400">المالك: </span>{detailSupplier.ownerName}</p>
            <p><span className="text-slate-400">الهاتف: </span><a href={`tel:${detailSupplier.phone}`} className="text-brand-600 hover:underline">{detailSupplier.phone}</a></p>
            {detailSupplier.lat && detailSupplier.lng && (
              <p><span className="text-slate-400">الموقع: </span><a href={`https://www.google.com/maps?q=${detailSupplier.lat},${detailSupplier.lng}`} target="_blank" rel="noreferrer" className="text-brand-500 hover:underline">عرض على الخريطة</a></p>
            )}
            {detailSupplier.notes && <p className="sm:col-span-2"><span className="text-slate-400">ملاحظات: </span>{detailSupplier.notes}</p>}
          </div>

          <div className="mt-3 flex flex-wrap gap-1">
            {detailSupplier.specialties.map((sp) => (
              <span key={sp.id} className="rounded-full bg-brand-50 px-2 py-0.5 text-xs font-medium text-brand-700">{sp.name}</span>
            ))}
            {detailSupplier.isMaterialSupplier && <span className="rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700">مورد مواد</span>}
            {detailSupplier.isContractor && <span className="rounded-full bg-orange-50 px-2 py-0.5 text-xs font-medium text-orange-700">منفذ</span>}
            {(detailSupplier.traderTypes || []).map((t) => (
              <span key={t} className="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">{traderTypeLabels[t] || t}</span>
            ))}
          </div>

          {/* Rating history */}
          <h4 className="mt-5 font-bold text-brand-800">سجل التقييمات</h4>
          <div className="mt-2 divide-y divide-slate-100 max-h-48 overflow-y-auto">
            {(detailSupplier.ratings || []).map((r) => (
              <div key={r.id} className="py-2 text-sm">
                <div className="flex items-center justify-between">
                  <Stars value={r.value} size="text-xs" />
                  <span className="text-xs text-slate-400">{new Date(r.createdAt).toLocaleDateString('ar-IQ')}</span>
                </div>
                <p className="text-slate-500 mt-0.5"><span className="text-slate-400">{r.ratedByName}: </span>{r.note || '-'}</p>
              </div>
            ))}
            {(!detailSupplier.ratings || detailSupplier.ratings.length === 0) && (
              <p className="py-3 text-center text-slate-400 text-sm">لا توجد تقييمات</p>
            )}
          </div>

          {/* Actions */}
          <div className="mt-5 flex gap-3 flex-wrap">
            <button onClick={() => { setRatingTarget(detailSupplier); setRatingValue(0); setRatingNote(''); setShowRatingModal(true) }}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30">
              إضافة تقييم
            </button>
            {isAdmin && (
              <>
                <button onClick={() => { setDetailSupplier(null); openEdit(detailSupplier) }}
                  className="rounded-lg border border-brand-500 px-4 py-2 text-sm font-medium text-brand-600 transition-all hover:bg-brand-50">
                  تعديل
                </button>
                <button onClick={() => handleDelete(detailSupplier.id)}
                  className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-600 transition-all hover:bg-red-50">
                  حذف
                </button>
              </>
            )}
            <button onClick={() => setDetailSupplier(null)}
              className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 transition-all hover:bg-slate-50">
              إغلاق
            </button>
          </div>
        </Modal>
      )}

      {/* Rating Modal */}
      {showRatingModal && ratingTarget && (
        <Modal onClose={() => setShowRatingModal(false)}>
          <h3 className="text-xl font-bold text-brand-800 mb-4">تقييم {ratingTarget.companyName}</h3>
          <form onSubmit={handleRate} className="space-y-4">
            <div>
              <label className="mb-2 block text-sm font-medium text-slate-600">التقييم</label>
              <StarSelector value={ratingValue} onChange={setRatingValue} />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">ملاحظة (اختياري)</label>
              <textarea value={ratingNote} onChange={(e) => setRatingNote(e.target.value)} rows={2}
                className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
            </div>
            <div className="flex gap-3">
              <button type="submit" disabled={ratingSubmitting || !ratingValue}
                className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50">
                {ratingSubmitting ? 'جاري الحفظ...' : 'حفظ التقييم'}
              </button>
              <button type="button" onClick={() => setShowRatingModal(false)}
                className="rounded-lg border border-slate-300 px-6 py-2 font-medium text-slate-600 transition-all hover:bg-slate-50">
                إلغاء
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
