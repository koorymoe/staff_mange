import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type GpsDeviceRequest } from '../../api'
import { useSession } from '../../session'
import { matches as searchMatches } from '../../utils/search'

const subLabel = (t: string) => t === 'THREE_MONTHS' ? '3 أشهر' : t === 'SIX_MONTHS' ? '6 أشهر' : 'سنوي'
const remainingDays = (end: string | null) => end ? Math.ceil((new Date(end).getTime() - Date.now()) / 86400000) : null

export default function GpsRenewal() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [results, setResults] = useState<GpsDeviceRequest[]>([])
  const [selected, setSelected] = useState<GpsDeviceRequest | null>(null)
  const [searching, setSearching] = useState(false)
  const [notFound, setNotFound] = useState(false)
  const [newSub, setNewSub] = useState<'THREE_MONTHS' | 'SIX_MONTHS' | 'YEARLY'>('THREE_MONTHS')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  const handleSearch = async () => {
    if (!search.trim()) return
    setSearching(true); setNotFound(false); setResults([]); setSelected(null); setError('')
    const all = await api.getGpsDevices()
    const matches = all.filter(r =>
      (r.status === 'APPROVED' || r.status === 'DELIVERED') &&
      searchMatches([r.customer.fullName, r.customer.phone], search)
    )
    if (matches.length > 0) {
      setResults(matches)
      if (matches.length === 1) setSelected(matches[0])
    } else {
      setNotFound(true)
    }
    setSearching(false)
  }

  const handleSubmit = async () => {
    if (!selected || !employee) return
    setLoading(true); setError('')
    try {
      await api.createGpsRenewal({
        customerId: selected.customerId,
        deviceRequestId: selected.id,
        employeeId: employee.id,
        subscriptionType: newSub,
      })
      setSuccess(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ أثناء إرسال الطلب')
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <div className="mx-auto mt-20 max-w-lg rounded-2xl bg-white p-8 text-center shadow-sm">
        <div className="mb-4 text-6xl">✅</div>
        <h2 className="mb-2 text-2xl font-bold text-brand-900">تم إرسال طلب التجديد للإداري</h2>
        <p className="mb-6 text-slate-500">سيتم مراجعته والموافقة عليه من قبل الإداري</p>
        <div className="flex justify-center gap-3">
          <button onClick={() => navigate('/')} className="rounded-lg bg-slate-100 px-5 py-2.5 text-sm font-bold text-slate-600 hover:bg-slate-200">الرئيسية</button>
          <button onClick={() => { setSuccess(false); setSelected(null); setResults([]); setSearch('') }} className="rounded-lg bg-brand-500 px-5 py-2.5 text-sm font-bold text-white hover:bg-brand-600">طلب آخر</button>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-2xl">
      <h2 className="mb-6 text-2xl font-bold text-brand-900">طلب تجديد اشتراك 🔄</h2>

      <div className="mb-4 rounded-2xl bg-white p-5 shadow-sm">
        <h3 className="mb-3 font-bold text-brand-900">البحث عن الزبون</h3>
        <div className="flex gap-2">
          <input value={search} onChange={e => setSearch(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSearch()}
            placeholder="اسم الزبون أو رقم الهاتف..." className="flex-1 rounded-lg border border-slate-200 px-4 py-2.5 text-sm" />
          <button onClick={handleSearch} disabled={searching} className="rounded-lg bg-brand-500 px-6 py-2.5 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {searching ? '...' : 'بحث'}
          </button>
        </div>
        {notFound && <p className="mt-3 text-sm text-red-500">لم يتم العثور على الزبون</p>}
      </div>

      {results.length > 1 && !selected && (
        <div className="mb-4 rounded-2xl bg-white p-5 shadow-sm">
          <h3 className="mb-3 font-bold text-brand-900">نتائج البحث ({results.length})</h3>
          <div className="flex flex-col gap-2">
            {results.map(r => {
              const days = remainingDays(r.subscriptionEnd)
              return (
                <button key={r.id} onClick={() => setSelected(r)}
                  className="rounded-xl border-2 border-slate-200 bg-slate-50 p-4 text-right transition-all hover:border-brand-400">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-bold text-brand-900">{r.customer.fullName} {r.customer.fatherName} {r.customer.grandfatherName}</p>
                      <p className="mt-0.5 text-sm text-slate-500">📞 {r.customer.phone}</p>
                    </div>
                    {days !== null && (
                      <span className={`font-bold text-sm ${days < 0 ? 'text-red-600' : days < 30 ? 'text-amber-600' : 'text-emerald-600'}`}>
                        {days < 0 ? `منتهي ${Math.abs(days)} يوم` : `${days} يوم`}
                      </span>
                    )}
                  </div>
                </button>
              )
            })}
          </div>
        </div>
      )}

      {selected && (
        <div className="flex flex-col gap-5 rounded-2xl bg-white p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-brand-900">بيانات الزبون</h3>
            {results.length > 1 && <button onClick={() => setSelected(null)} className="text-sm text-brand-600 hover:underline">← تغيير الزبون</button>}
          </div>
          <div className="grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-4 text-sm">
            <p><span className="text-slate-500">الاسم: </span><b>{selected.customer.fullName} {selected.customer.fatherName} {selected.customer.grandfatherName}</b></p>
            <p><span className="text-slate-500">الهاتف: </span><b>{selected.customer.phone}</b></p>
            <p><span className="text-slate-500">الاشتراك الحالي: </span><b>{subLabel(selected.subscriptionType)}</b></p>
            <p><span className="text-slate-500">تاريخ الانتهاء: </span><b>{selected.subscriptionEnd ? new Date(selected.subscriptionEnd).toLocaleDateString('ar-IQ') : '-'}</b></p>
          </div>
          <div>
            <label className="mb-2 block text-xs font-bold text-slate-500">نوع الاشتراك الجديد</label>
            <div className="grid grid-cols-3 gap-3">
              {([['THREE_MONTHS', '3 أشهر'], ['SIX_MONTHS', '6 أشهر'], ['YEARLY', 'سنوي']] as const).map(([val, label]) => (
                <button key={val} type="button" onClick={() => setNewSub(val)}
                  className={`rounded-xl border-2 p-3 text-center text-sm font-bold transition-all ${newSub === val ? 'border-brand-500 bg-brand-500 text-white' : 'border-slate-200 text-slate-600'}`}>
                  {label}
                </button>
              ))}
            </div>
          </div>
          {error && <p className="text-sm text-red-500">{error}</p>}
          <button onClick={handleSubmit} disabled={loading} className="rounded-xl bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {loading ? 'جاري الإرسال...' : '📤 إرسال طلب التجديد للإداري'}
          </button>
        </div>
      )}
    </div>
  )
}
