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

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة شرائح SIM</h2>

      <div className="mt-4">
        <button onClick={() => setShowForm(!showForm)} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md transition-all hover:shadow-lg">
          {showForm ? 'إلغاء' : 'إضافة شريحة'}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit} className="mt-4 grid grid-cols-1 gap-4 rounded-xl bg-white p-6 shadow-md sm:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">رقم الشريحة</label>
            <input required value={simNumber} onChange={(e) => setSimNumber(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">ICCID</label>
            <input required value={iccid} onChange={(e) => setIccid(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-600">المشغل</label>
            <select value={operator} onChange={(e) => setOperator(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500">
              <option value="ZAIN">زين</option>
              <option value="ASIACELL">آسياسيل</option>
              <option value="KOREK">كورك</option>
              <option value="OTHER">أخرى</option>
            </select>
          </div>
          <div className="sm:col-span-3">
            <button type="submit" disabled={submitting} className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md disabled:opacity-50">
              {submitting ? 'جاري الحفظ...' : 'حفظ الشريحة'}
            </button>
          </div>
        </form>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl bg-white shadow-md">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">رقم الشريحة</th>
                <th className="px-4 py-3 text-sm font-semibold">ICCID</th>
                <th className="px-4 py-3 text-sm font-semibold">المشغل</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {sims.map((s) => (
                <tr key={s.id} className="hover:bg-slate-50">
                  <td className="px-4 py-3 font-medium">{s.simNumber}</td>
                  <td className="px-4 py-3 text-slate-600 text-sm font-mono">{s.iccid}</td>
                  <td className="px-4 py-3 text-slate-600">{operatorLabel[s.operator] || s.operator}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded-full px-3 py-1 text-xs font-medium ${s.status === 'ACTIVE' ? 'bg-green-100 text-green-800' : 'bg-slate-100 text-slate-800'}`}>
                      {s.status === 'ACTIVE' ? 'فعالة' : 'غير فعالة'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-600">{s.customer?.fullName || '-'}</td>
                  <td className="px-4 py-3">
                    <button onClick={() => toggleStatus(s)} className="rounded-lg bg-slate-200 px-3 py-1 text-xs text-slate-700 hover:bg-slate-300">
                      {s.status === 'ACTIVE' ? 'تعطيل' : 'تفعيل'}
                    </button>
                  </td>
                </tr>
              ))}
              {sims.length === 0 && (
                <tr><td colSpan={6} className="px-4 py-6 text-center text-slate-400">لا يوجد شرائح</td></tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
