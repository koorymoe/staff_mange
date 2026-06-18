import { useState, useEffect } from 'react'
import { api } from '../../api'

const OPERATORS = ['ZAIN', 'ASIACELL', 'KOREK', 'OTHER']
const OPERATOR_LABELS: Record<string, string> = { ZAIN: 'زين', ASIACELL: 'آسياسيل', KOREK: 'كورك', OTHER: 'أخرى' }

const emptyForm = { simNumber: '', iccid: '', operator: 'ZAIN', notes: '' }

export default function GpsSims() {
  const [sims, setSims] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [showAdd, setShowAdd] = useState(false)
  const [editSim, setEditSim] = useState<any>(null)
  const [form, setForm] = useState(emptyForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  async function load() {
    setLoading(true)
    try {
      const data = await api.getSimCards()
      setSims(data || [])
    } catch (e) { console.error(e) }
    setLoading(false)
  }

  useEffect(() => { load() }, [])

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true); setError('')
    try {
      await api.createSimCard({ simNumber: form.simNumber, iccid: form.iccid || undefined, operator: form.operator, notes: form.notes || undefined })
      setSuccess('تم إضافة SIM بنجاح')
      setShowAdd(false); setForm(emptyForm); load()
    } catch { setError('حدث خطأ أثناء الإضافة') }
    setSaving(false)
  }

  async function handleEdit(e: React.FormEvent) {
    e.preventDefault()
    if (!editSim) return
    setSaving(true); setError('')
    try {
      await api.updateSimCard(editSim.id, { simNumber: form.simNumber, iccid: form.iccid || undefined, operator: form.operator, notes: form.notes || undefined })
      setSuccess('تم التحديث بنجاح')
      setEditSim(null); setForm(emptyForm); load()
    } catch { setError('حدث خطأ أثناء التحديث') }
    setSaving(false)
  }

  async function handleRelease(sim: any) {
    if (!confirm(`تحرير الخط ${sim.simNumber}؟`)) return
    try {
      await api.updateSimCard(sim.id, { status: 'AVAILABLE', currentCustomerName: null })
      setSuccess('تم تحرير الخط بنجاح'); load()
    } catch { alert('حدث خطأ') }
  }

  function openEdit(sim: any) {
    setEditSim(sim)
    setForm({ simNumber: sim.simNumber, iccid: sim.iccid || '', operator: sim.operator, notes: sim.notes || '' })
    setError('')
  }

  const total = sims.length
  const available = sims.filter(s => s.status === 'AVAILABLE' || s.status === 'available').length
  const inUse = sims.filter(s => s.status === 'IN_USE' || s.status === 'in_use').length
  const isModalOpen = showAdd || !!editSim

  return (
    <div dir="rtl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold" style={{ color: '#1a3a5c' }}>إدارة SIM كارتات 📱</h1>
        <button onClick={() => { setShowAdd(true); setEditSim(null); setForm(emptyForm); setError('') }}
          className="text-sm font-semibold px-5 py-2.5 rounded-xl"
          style={{ background: '#f8fafc', color: '#1a3a5c', border: '1px solid #e2e8f0', cursor: 'pointer' }}>
          + إضافة SIM
        </button>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="rounded-2xl p-5 text-center" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-3xl font-bold" style={{ color: '#1f2937' }}>{total}</div>
          <div className="text-sm mt-1" style={{ color: '#6b7280' }}>إجمالي SIM</div>
        </div>
        <div className="rounded-2xl p-5 text-center" style={{ background: '#f0fdf4', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-3xl font-bold" style={{ color: '#15803d' }}>{available}</div>
          <div className="text-sm mt-1" style={{ color: '#16a34a' }}>متاح</div>
        </div>
        <div className="rounded-2xl p-5 text-center" style={{ background: '#eff6ff', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <div className="text-3xl font-bold" style={{ color: '#1d4ed8' }}>{inUse}</div>
          <div className="text-sm mt-1" style={{ color: '#2563eb' }}>مستخدم</div>
        </div>
      </div>

      {success && <div className="text-sm text-center rounded-lg p-3 mb-4" style={{ background: '#f0fdf4', border: '1px solid #bbf7d0', color: '#15803d' }}>{success}</div>}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
          <div className="rounded-2xl w-full max-w-md m-4" style={{ background: 'white', boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)' }}>
            <div className="p-5 flex items-center justify-between" style={{ borderBottom: '1px solid #e5e7eb' }}>
              <h2 className="text-xl font-bold" style={{ color: '#1a3a5c' }}>{editSim ? 'تعديل SIM' : 'إضافة SIM جديد'}</h2>
              <button onClick={() => { setShowAdd(false); setEditSim(null) }} style={{ color: '#9ca3af', fontSize: '1.5rem', background: 'none', border: 'none', cursor: 'pointer' }}>✕</button>
            </div>
            <form onSubmit={editSim ? handleEdit : handleAdd} className="p-6" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>رقم الخط *</label>
                <input className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                  value={form.simNumber} onChange={e => setForm(p => ({ ...p, simNumber: e.target.value }))} required placeholder="07xxxxxxxxx" />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>رقم الشريحة (ICCID)</label>
                <input className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                  value={form.iccid} onChange={e => setForm(p => ({ ...p, iccid: e.target.value }))} placeholder="اختياري" />
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>المشغل *</label>
                <select className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                  value={form.operator} onChange={e => setForm(p => ({ ...p, operator: e.target.value }))} required>
                  {OPERATORS.map(op => <option key={op} value={op}>{OPERATOR_LABELS[op] || op}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-semibold mb-1" style={{ color: '#374151' }}>ملاحظات</label>
                <input className="w-full text-sm rounded-xl px-4 py-2.5" style={{ border: '1px solid #e5e7eb', outline: 'none' }}
                  value={form.notes} onChange={e => setForm(p => ({ ...p, notes: e.target.value }))} placeholder="اختياري" />
              </div>
              {error && <div className="text-sm text-center rounded-lg p-3" style={{ background: '#fef2f2', border: '1px solid #fecaca', color: '#dc2626' }}>{error}</div>}
              <button type="submit" disabled={saving}
                className="w-full py-3 rounded-xl font-semibold text-white"
                style={{ background: '#1a3a5c', border: 'none', cursor: 'pointer' }}>
                {saving ? 'جارٍ الحفظ...' : editSim ? '✅ حفظ التعديل' : '✅ إضافة SIM'}
              </button>
            </form>
          </div>
        </div>
      )}

      {loading ? (
        <div className="text-center py-20" style={{ color: '#9ca3af' }}>جارٍ التحميل...</div>
      ) : (
        <div className="rounded-2xl overflow-hidden" style={{ background: 'white', boxShadow: '0 1px 3px rgba(0,0,0,0.05)' }}>
          <table className="w-full">
            <thead>
              <tr className="text-right text-sm" style={{ color: '#6b7280', borderBottom: '1px solid #e5e7eb', background: '#f9fafb' }}>
                <th className="px-5 py-3 font-semibold">رقم الخط</th>
                <th className="px-5 py-3 font-semibold">رقم الشريحة (ICCID)</th>
                <th className="px-5 py-3 font-semibold">المشغل</th>
                <th className="px-5 py-3 font-semibold">الحالة</th>
                <th className="px-5 py-3 font-semibold">الزبون الحالي</th>
                <th className="px-5 py-3 font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {sims.map(sim => (
                <tr key={sim.id} style={{ borderBottom: '1px solid #f3f4f6' }}>
                  <td className="px-5 py-4 font-medium" style={{ color: '#1f2937' }}>{sim.simNumber}</td>
                  <td className="px-5 py-4" style={{ color: '#6b7280' }}>{sim.iccid || '-'}</td>
                  <td className="px-5 py-4" style={{ color: '#374151' }}>{OPERATOR_LABELS[sim.operator] || sim.operator}</td>
                  <td className="px-5 py-4">
                    {(sim.status === 'AVAILABLE' || sim.status === 'available')
                      ? <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#f0fdf4', color: '#15803d' }}>متاح</span>
                      : <span className="text-xs px-2 py-0.5 rounded-full" style={{ background: '#eff6ff', color: '#1d4ed8' }}>مستخدم</span>}
                  </td>
                  <td className="px-5 py-4" style={{ color: '#6b7280' }}>{sim.currentCustomerName || '-'}</td>
                  <td className="px-5 py-4">
                    {(sim.status === 'IN_USE' || sim.status === 'in_use') ? (
                      <button onClick={() => handleRelease(sim)} className="text-sm font-medium" style={{ color: '#ea580c', background: 'none', border: 'none', cursor: 'pointer' }}>تحرير الخط</button>
                    ) : (
                      <button onClick={() => openEdit(sim)} className="text-sm font-medium" style={{ color: '#3b82f6', background: 'none', border: 'none', cursor: 'pointer' }}>تعديل</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {sims.length === 0 && <div className="text-center py-12" style={{ color: '#9ca3af' }}>لا توجد SIM كارتات</div>}
        </div>
      )}
    </div>
  )
}
