import { useState, useEffect } from 'react'
import { api } from '../../api'

function formatDate(d: string) {
  if (!d) return '-'
  try { return new Date(d).toLocaleDateString('ar-IQ') } catch { return d }
}

const statusMap: Record<string, { label: string; color: string; bg: string }> = {
  PENDING: { label: 'معلق', color: '#d97706', bg: '#fffbeb' },
  pending: { label: 'معلق', color: '#d97706', bg: '#fffbeb' },
  IN_PROGRESS: { label: 'قيد المعالجة', color: '#7c3aed', bg: '#f5f3ff' },
  in_progress: { label: 'قيد المعالجة', color: '#7c3aed', bg: '#f5f3ff' },
  COMPLETED: { label: 'مكتمل', color: '#16a34a', bg: '#f0fdf4' },
  completed: { label: 'مكتمل', color: '#16a34a', bg: '#f0fdf4' },
}

export default function GpsMaintenance() {
  const [reqs, setReqs] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<any>(null)
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

  async function load() {
    try {
      const data = await api.getGpsMaintenance()
      setReqs(data || [])
    } catch (e) { console.error(e) }
    setLoading(false)
  }
  useEffect(() => { load() }, [])

  async function handleUpdate(newStatus: string) {
    if (!selected) return
    setSaving(true)
    try {
      await api.updateGpsMaintenance(selected.id, { status: newStatus, adminNotes: notes })
      setSelected(null); load()
    } catch (e) { console.error(e); alert('حدث خطأ') }
    setSaving(false)
  }

  const pending = reqs.filter(r => r.status === 'PENDING' || r.status === 'pending').length
  const inProgress = reqs.filter(r => r.status === 'IN_PROGRESS' || r.status === 'in_progress').length
  const completed = reqs.filter(r => r.status === 'COMPLETED' || r.status === 'completed').length

  return (
    <div dir="rtl">
      <h1 className="text-2xl font-bold mb-6" style={{ color: '#1a3a5c' }}>طلبات الصيانة 🔧</h1>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="rounded-2xl p-4 text-center" style={{ background: '#fffbeb', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-2xl font-bold" style={{ color: '#d97706' }}>{pending}</div>
          <div className="text-sm mt-1" style={{ color: '#92400e' }}>معلق</div>
        </div>
        <div className="rounded-2xl p-4 text-center" style={{ background: '#f5f3ff', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-2xl font-bold" style={{ color: '#7c3aed' }}>{inProgress}</div>
          <div className="text-sm mt-1" style={{ color: '#5b21b6' }}>قيد المعالجة</div>
        </div>
        <div className="rounded-2xl p-4 text-center" style={{ background: '#f0fdf4', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-2xl font-bold" style={{ color: '#16a34a' }}>{completed}</div>
          <div className="text-sm mt-1" style={{ color: '#166534' }}>مكتمل</div>
        </div>
      </div>

      {loading ? <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div> : reqs.length === 0 ? (
        <div className="text-center py-16 rounded-2xl" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <p className="text-lg" style={{ color: '#9ca3af' }}>لا توجد طلبات صيانة</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {reqs.map(req => {
            const s = statusMap[req.status] || statusMap.PENDING
            return (
              <div key={req.id} className="rounded-xl p-5 flex items-center justify-between"
                style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
                <div style={{ flex: 1 }}>
                  <div className="flex items-center gap-3 mb-1">
                    <span className="font-bold" style={{ color: '#1f2937' }}>{req.customer ? `${req.customer.fullName} ${req.customer.fatherName}` : '-'}</span>
                    <span className="text-xs px-2 py-0.5 rounded-full" style={{ color: s.color, background: s.bg }}>{s.label}</span>
                  </div>
                  <p className="text-sm mb-1" style={{ color: '#6b7280' }}>🔧 {req.problemDescription}</p>
                  <div className="flex gap-4 text-xs" style={{ color: '#9ca3af' }}>
                    <span>📞 {req.customer?.phone}</span>
                    <span>👤 {req.employee?.name || req.employee?.fullName || '-'}</span>
                    <span>📅 {req.createdAt ? formatDate(req.createdAt) : '-'}</span>
                  </div>
                </div>
                <button onClick={() => { setSelected(req); setNotes(req.adminNotes || '') }}
                  className="text-sm px-4 py-2 mr-4 rounded-lg font-medium"
                  style={{ background: '#f8fafc', color: '#1a3a5c', border: '1px solid #e2e8f0', cursor: 'pointer' }}>
                  تفاصيل ←
                </button>
              </div>
            )
          })}
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
          <div className="rounded-2xl w-full max-w-xl m-4" style={{ background: 'white', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}>
            <div className="p-5 flex items-center justify-between" style={{ borderBottom: '1px solid #e5e7eb' }}>
              <h2 className="text-xl font-bold" style={{ color: '#1a3a5c' }}>تفاصيل طلب الصيانة</h2>
              <button onClick={() => setSelected(null)} style={{ color: '#9ca3af', fontSize: '1.5rem', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
            <div className="p-6" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div className="grid grid-cols-2 gap-3 text-sm p-4 rounded-xl" style={{ background: '#f9fafb' }}>
                <div><span style={{ color: '#6b7280' }}>الاسم: </span><span className="font-semibold">{selected.customer?.fullName} {selected.customer?.fatherName} {selected.customer?.grandfatherName}</span></div>
                <div><span style={{ color: '#6b7280' }}>الهاتف: </span><span className="font-semibold">{selected.customer?.phone}</span></div>
                <div><span style={{ color: '#6b7280' }}>العنوان: </span><span className="font-semibold">{selected.customer?.address}</span></div>
                <div><span style={{ color: '#6b7280' }}>التاريخ: </span><span className="font-semibold">{selected.createdAt ? formatDate(selected.createdAt) : '-'}</span></div>
              </div>
              <div>
                <p className="text-sm font-semibold mb-1" style={{ color: '#374151' }}>وصف المشكلة</p>
                <p className="text-sm p-3 rounded-lg" style={{ background: '#f9fafb', color: '#374151' }}>{selected.problemDescription}</p>
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>ملاحظات الإداري</label>
                <textarea className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none', minHeight: '5rem' }}
                  value={notes} onChange={e => setNotes(e.target.value)} placeholder="أضف ملاحظاتك..." />
              </div>
            </div>
            <div className="p-5 flex gap-3" style={{ borderTop: '1px solid #e5e7eb' }}>
              <button onClick={() => handleUpdate('IN_PROGRESS')} disabled={saving}
                className="flex-1 py-2.5 rounded-lg font-semibold text-sm"
                style={{ background: '#7c3aed15', color: '#7c3aed', border: 'none', cursor: 'pointer' }}>
                قيد المعالجة
              </button>
              <button onClick={() => handleUpdate('COMPLETED')} disabled={saving}
                className="flex-1 py-2.5 rounded-lg font-semibold text-sm"
                style={{ background: '#16a34a15', color: '#16a34a', border: 'none', cursor: 'pointer' }}>
                ✅ مكتمل
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
