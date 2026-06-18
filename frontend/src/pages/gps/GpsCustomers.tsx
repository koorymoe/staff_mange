import { useState, useEffect } from 'react'
import { api } from '../../api'

function formatDate(d: string) {
  if (!d) return '-'
  try { return new Date(d).toLocaleDateString('ar-IQ') } catch { return d }
}

function getRemainingDays(endDate: string): number {
  const end = new Date(endDate)
  const now = new Date()
  return Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
}

export default function GpsCustomers() {
  const [customers, setCustomers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [deleting, setDeleting] = useState(false)
  const [editCustomer, setEditCustomer] = useState<any>(null)
  const [editDate, setEditDate] = useState('')
  const [saving, setSaving] = useState(false)

  async function load() {
    try {
      const [custs, devices] = await Promise.all([api.getGpsCustomers(), api.getGpsDevices()])
      const deviceMap: Record<string, any> = {}
      for (const d of (devices || [])) {
        if (d.customerId && d.status === 'DELIVERED') deviceMap[d.customerId] = d
      }
      setCustomers((custs || []).map((c: any) => ({
        ...c,
        subscription: deviceMap[c.id] ? {
          id: deviceMap[c.id].id,
          gpsNumber: deviceMap[c.id].gpsNumber,
          simNumber: deviceMap[c.id].simCard?.simNumber,
          subscriptionEnd: deviceMap[c.id].subscriptionEnd,
        } : null
      })))
    } catch (e) { console.error(e) }
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  const filtered = customers.filter(c => {
    const name = `${c.fullName || ''} ${c.fatherName || ''} ${c.grandfatherName || ''}`.toLowerCase()
    const gps = c.subscription?.gpsNumber || ''
    return name.includes(search.toLowerCase()) || (c.phone || '').includes(search) || gps.includes(search)
  })

  function toggleSelect(id: string) {
    setSelected(prev => { const next = new Set(prev); next.has(id) ? next.delete(id) : next.add(id); return next })
  }

  function toggleAll() {
    if (selected.size === filtered.length) setSelected(new Set())
    else setSelected(new Set(filtered.map(c => c.id)))
  }

  async function handleDeleteSelected() {
    if (selected.size === 0) return
    const names = customers.filter(c => selected.has(c.id)).map(c => `${c.fullName} ${c.fatherName}`).join('، ')
    if (!confirm(`هل تريد حذف ${selected.size} زبون؟\n${names}`)) return
    setDeleting(true)
    // No delete API available, just remove from UI
    setCustomers(prev => prev.filter(c => !selected.has(c.id)))
    setSelected(new Set())
    setDeleting(false)
  }

  function openEdit(c: any) {
    setEditCustomer(c)
    setEditDate(c.subscription?.subscriptionEnd || '')
  }

  async function handleSaveDate() {
    if (!editCustomer?.subscription?.id || !editDate) return
    setSaving(true)
    try {
      await api.updateGpsDevice(editCustomer.subscription.id, { subscriptionEnd: editDate })
      setCustomers(prev => prev.map(c =>
        c.id === editCustomer.id
          ? { ...c, subscription: c.subscription ? { ...c.subscription, subscriptionEnd: editDate } : c.subscription }
          : c
      ))
    } catch (e) { console.error(e); alert('حدث خطأ أثناء الحفظ') }
    setEditCustomer(null)
    setEditDate('')
    setSaving(false)
  }

  function getStatusBadge(sub: any) {
    if (!sub?.subscriptionEnd) return <span className="text-xs" style={{ color: '#d1d5db' }}>-</span>
    const days = getRemainingDays(sub.subscriptionEnd)
    if (days < 0) return <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#fef2f2', color: '#dc2626' }}>منتهي</span>
    if (days < 30) return <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#fffbeb', color: '#d97706' }}>قارب الانتهاء</span>
    return <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#f0fdf4', color: '#16a34a' }}>نشط</span>
  }

  return (
    <div dir="rtl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#1a3a5c' }}>الزبائن 👥</h1>
        <span className="text-sm" style={{ color: '#9ca3af' }}>{customers.length} زبون</span>
      </div>

      <div className="flex gap-3 mb-4 items-center">
        <input
          className="text-sm rounded-xl px-4 py-2.5 w-full max-w-sm"
          style={{ border: '1px solid #e5e7eb', outline: 'none' }}
          value={search} onChange={e => setSearch(e.target.value)}
          placeholder="بحث بالاسم أو رقم الهاتف أو رقم GPS..."
        />
        {selected.size > 0 && (
          <button onClick={handleDeleteSelected} disabled={deleting}
            className="text-sm font-semibold text-white px-6 py-2.5 rounded-xl"
            style={{ background: '#dc2626', border: 'none', cursor: 'pointer' }}>
            {deleting ? 'جارٍ الحذف...' : `🗑️ حذف المحددين (${selected.size})`}
          </button>
        )}
      </div>

      {editCustomer && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.4)' }}>
          <div className="rounded-2xl p-6 max-w-sm w-full mx-4" style={{ background: 'white', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}>
            <h3 className="font-bold text-lg mb-1" style={{ color: '#1a3a5c' }}>تعديل تاريخ الانتهاء</h3>
            <p className="text-sm mb-4" style={{ color: '#6b7280' }}>{editCustomer.fullName} {editCustomer.fatherName} {editCustomer.grandfatherName}</p>
            <div className="mb-4">
              <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>تاريخ انتهاء الاشتراك</label>
              <input type="date" className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                value={editDate} onChange={e => setEditDate(e.target.value)} />
              {editDate && (
                <div className="mt-2 text-sm font-semibold" style={{ color: getRemainingDays(editDate) < 0 ? '#dc2626' : '#16a34a' }}>
                  الأيام المتبقية: {(() => { const d = getRemainingDays(editDate); return d < 0 ? `${Math.abs(d)}-` : d })()}
                </div>
              )}
            </div>
            <div className="flex gap-3">
              <button onClick={handleSaveDate} disabled={saving || !editDate}
                className="flex-1 py-2 rounded-xl font-semibold text-white"
                style={{ background: '#1a3a5c', border: 'none', cursor: 'pointer' }}>
                {saving ? 'جارٍ الحفظ...' : '✅ حفظ'}
              </button>
              <button onClick={() => setEditCustomer(null)}
                className="flex-1 py-2 rounded-xl font-semibold"
                style={{ background: '#f1f5f9', color: '#64748b', border: 'none', cursor: 'pointer' }}>
                إلغاء
              </button>
            </div>
          </div>
        </div>
      )}

      {loading ? <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div> : (
        <div className="rounded-2xl overflow-x-auto" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <table className="w-full" style={{ minWidth: '800px' }}>
            <thead>
              <tr className="text-right text-sm" style={{ color: '#6b7280', borderBottom: '1px solid #e5e7eb', background: '#f9fafb' }}>
                <th className="px-4 py-3">
                  <input type="checkbox" checked={selected.size === filtered.length && filtered.length > 0} onChange={toggleAll} className="w-4 h-4 cursor-pointer" />
                </th>
                <th className="px-5 py-3 font-semibold">الاسم الكامل</th>
                <th className="px-5 py-3 font-semibold">رقم الهاتف</th>
                <th className="px-5 py-3 font-semibold">رقم الشريحة</th>
                <th className="px-5 py-3 font-semibold">رقم الجهاز</th>
                <th className="px-5 py-3 font-semibold">تاريخ انتهاء الاشتراك</th>
                <th className="px-5 py-3 font-semibold">الأيام المتبقية</th>
                <th className="px-5 py-3 font-semibold">الحالة</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map(c => (
                <tr key={c.id} style={{ borderBottom: '1px solid #f3f4f6', background: selected.has(c.id) ? '#fef2f2' : '' }}>
                  <td className="px-4 py-4">
                    <input type="checkbox" checked={selected.has(c.id)} onChange={() => toggleSelect(c.id)} className="w-4 h-4 cursor-pointer" />
                  </td>
                  <td className="px-5 py-4 font-medium" style={{ color: '#1f2937' }}>
                    <div className="flex items-center gap-2">
                      <span>{c.fullName} {c.fatherName} {c.grandfatherName}</span>
                      {c.subscription && (
                        <button onClick={() => openEdit(c)} className="text-xs font-semibold" style={{ color: '#60a5fa', background: 'none', border: 'none', cursor: 'pointer' }}>✏️</button>
                      )}
                    </div>
                  </td>
                  <td className="px-5 py-4" style={{ color: '#6b7280' }}>{c.phone}</td>
                  <td className="px-5 py-4 font-bold" style={{ color: '#1a3a5c' }}>{c.subscription?.simNumber || '-'}</td>
                  <td className="px-5 py-4 text-sm" style={{ color: '#6b7280' }}>{c.subscription?.gpsNumber || '-'}</td>
                  <td className="px-5 py-4 text-sm font-medium" style={{ color: '#dc2626' }}>
                    {c.subscription?.subscriptionEnd ? formatDate(c.subscription.subscriptionEnd) : '-'}
                  </td>
                  <td className="px-5 py-4 font-bold text-sm">
                    {c.subscription?.subscriptionEnd ? (() => {
                      const d = getRemainingDays(c.subscription.subscriptionEnd)
                      return <span style={{ color: d < 0 ? '#dc2626' : d < 30 ? '#d97706' : '#16a34a' }}>{d < 0 ? `${Math.abs(d)}-` : d}</span>
                    })() : '-'}
                  </td>
                  <td className="px-5 py-4">{getStatusBadge(c.subscription)}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {filtered.length === 0 && <div className="text-center py-12" style={{ color: '#9ca3af' }}>لا توجد نتائج</div>}
        </div>
      )}
    </div>
  )
}
