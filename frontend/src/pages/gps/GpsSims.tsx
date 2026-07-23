import { useState, useEffect } from 'react'
import { api } from '../../api'

export default function GpsSims() {
  const [sims, setSims] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [simNumber, setSimNumber] = useState('')
  const [iccid, setIccid] = useState('')
  const [operator, setOperator] = useState('ZAIN')

  const load = () => {
    setLoading(true)
    api.getSimCards()
      .then(setSims)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createSimCard({ simNumber, iccid, operator })
      setSimNumber(''); setIccid(''); setOperator('ZAIN')
      setShowForm(false)
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const toggleStatus = async (sim: any) => {
    try {
      await api.updateSimCard(sim.id, { status: sim.status === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE' })
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    }
  }

  const operatorLabel: Record<string, string> = { ZAIN: 'زين', ASIACELL: 'آسياسيل', KOREK: 'كورك', OTHER: 'أخرى' }

  const totalSims = sims.length
  const availableSims = sims.filter(s => s.status !== 'ACTIVE').length
  const inUseSims = sims.filter(s => s.status === 'ACTIVE').length

  return (
    <div dir="rtl">
      {/* Header */}
      <div className="mb-6 rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">📱 إدارة شرائح SIM</h1>
        <p className="mt-1 text-blue-200 text-sm">إضافة وإدارة شرائح الاتصال</p>
      </div>

      {/* Stats Cards */}
      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <span className="text-2xl">📊</span>
            <div>
              <p className="text-2xl font-bold" style={{ color: '#1a3a5c' }}>{totalSims}</p>
              <p className="text-sm text-slate-500">إجمالي الشرائح</p>
            </div>
          </div>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🟢</span>
            <div>
              <p className="text-2xl font-bold text-green-600">{availableSims}</p>
              <p className="text-sm text-slate-500">متوفرة</p>
            </div>
          </div>
        </div>
        <div className="rounded-2xl bg-white p-5 shadow-sm">
          <div className="flex items-center gap-3">
            <span className="text-2xl">🔵</span>
            <div>
              <p className="text-2xl font-bold text-blue-600">{inUseSims}</p>
              <p className="text-sm text-slate-500">مستخدمة</p>
            </div>
          </div>
        </div>
      </div>

      {/* Add Button */}
      <div className="mb-6">
        <button
          onClick={() => setShowForm(!showForm)}
          className="rounded-2xl px-6 py-3 text-sm font-bold text-white shadow-sm transition-all hover:opacity-90"
          style={{ backgroundColor: '#1a3a5c' }}
        >
          {showForm ? '✕ إلغاء' : '➕ إضافة شريحة'}
        </button>
      </div>

      {/* Add Form */}
      {showForm && (
        <div className="mb-6 rounded-2xl bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-bold" style={{ color: '#1a3a5c' }}>➕ شريحة جديدة</h3>
          <form onSubmit={handleSubmit} className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">رقم الشريحة *</label>
              <input required value={simNumber} onChange={(e) => setSimNumber(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">ICCID *</label>
              <input required value={iccid} onChange={(e) => setIccid(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">المشغل</label>
              <select value={operator} onChange={(e) => setOperator(e.target.value)} className="w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-400">
                <option value="ZAIN">زين</option>
                <option value="ASIACELL">آسياسيل</option>
                <option value="KOREK">كورك</option>
                <option value="OTHER">أخرى</option>
              </select>
            </div>
            <div className="sm:col-span-3">
              <button type="submit" disabled={submitting} className="rounded-2xl px-8 py-3 text-sm font-bold text-white shadow-sm disabled:opacity-50" style={{ backgroundColor: '#1a3a5c' }}>
                {submitting ? 'جاري الحفظ...' : '💾 حفظ الشريحة'}
              </button>
            </div>
          </form>
        </div>
      )}

      {loading && <div className="flex items-center justify-center p-12"><div className="text-lg text-slate-500">جاري التحميل...</div></div>}
      {error && <div className="rounded-2xl bg-red-50 p-6 text-red-600 shadow-sm">{error}</div>}

      {/* SIM Table */}
      {!loading && !error && (
        <div className="overflow-hidden rounded-2xl bg-white shadow-sm">
          <div className="overflow-x-auto">
          <table className="w-full text-right">
            <thead>
              <tr style={{ backgroundColor: '#1a3a5c' }}>
                <th className="px-4 py-3 text-sm font-semibold text-white">رقم الشريحة</th>
                <th className="px-4 py-3 text-sm font-semibold text-white">ICCID</th>
                <th className="px-4 py-3 text-sm font-semibold text-white">المشغل</th>
                <th className="px-4 py-3 text-sm font-semibold text-white">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold text-white">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold text-white">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {sims.map((s) => (
                <tr key={s.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{s.simNumber}</td>
                  <td className="px-4 py-3 text-slate-600 text-sm font-mono">{s.iccid}</td>
                  <td className="px-4 py-3 text-slate-600">{operatorLabel[s.operator] || s.operator}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${s.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 'bg-slate-100 text-slate-600'}`}>
                      {s.status === 'ACTIVE' ? 'مستخدمة' : 'متوفرة'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-600">{s.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => toggleStatus(s)}
                      className={`rounded-2xl px-4 py-1.5 text-xs font-bold text-white ${s.status === 'ACTIVE' ? 'bg-red-500 hover:bg-red-600' : 'bg-green-500 hover:bg-green-600'}`}
                    >
                      {s.status === 'ACTIVE' ? 'تعطيل' : 'تفعيل'}
                    </button>
                  </td>
                </tr>
              ))}
              {sims.length === 0 && (
                <tr><td colSpan={6} className="p-8 text-center text-slate-400">لا يوجد شرائح</td></tr>
              )}
            </tbody>
          </table>
          </div>
        </div>
      )}
    </div>
  )
}
