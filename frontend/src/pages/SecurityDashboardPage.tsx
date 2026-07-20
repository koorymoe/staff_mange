import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h} ساعة و${m} دقيقة`
}

export default function SecurityDashboardPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [data, setData] = useState<Awaited<ReturnType<typeof api.getSecurityDashboard>> | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!isOwner) return
    api.getSecurityDashboard().then(setData).catch((e) => setError(e.message))
  }, [isOwner])

  if (!isOwner) {
    return (
      <div className="rounded-xl border border-white bg-white p-10 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        هذي الصفحة حصرية لمالك النظام.
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">لوحة المراقبة الخلفية 👁️</h2>
      <p className="mt-1 text-slate-500">
        صحة السيرفر وسجل محاولات تسجيل الدخول — حصرية لحسابك.
      </p>

      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر جلب البيانات: {error}</p>
      )}

      {data && (
        <>
          <div className="mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div className="rounded-xl border border-white bg-white p-5 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-xl font-bold text-brand-900">{formatUptime(data.serverUptimeSeconds)}</p>
              <p className="mt-1 text-xs text-slate-500">مدة تشغيل السيرفر</p>
            </div>
            <div className="rounded-xl border border-white bg-white p-5 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-xl font-bold text-brand-900">{data.memoryUsedMB.toFixed(1)} MB</p>
              <p className="mt-1 text-xs text-slate-500">الذاكرة المستخدمة</p>
            </div>
            <div className="rounded-xl border border-white bg-white p-5 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <p className="text-xl font-bold text-brand-900">{data.goroutineCount}</p>
              <p className="mt-1 text-xs text-slate-500">مهام السيرفر النشطة</p>
            </div>
            <div className={`rounded-xl border border-white p-5 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)] ${data.failedLoginsLastHour > 5 ? 'bg-red-50' : 'bg-white'}`}>
              <p className={`text-xl font-bold ${data.failedLoginsLastHour > 5 ? 'text-red-600' : 'text-brand-900'}`}>
                {data.failedLoginsLastHour}
              </p>
              <p className="mt-1 text-xs text-slate-500">محاولات دخول فاشلة (آخر ساعة)</p>
            </div>
          </div>

          <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="border-b border-slate-100 px-4 py-3 text-sm font-bold text-brand-800">
              سجل محاولات تسجيل الدخول (آخر 100)
            </h3>
            <div className="max-h-[500px] overflow-y-auto overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right text-sm">
                <thead className="sticky top-0 bg-slate-50">
                  <tr>
                    <th className="px-4 py-2 font-semibold text-slate-600">المستخدم</th>
                    <th className="px-4 py-2 font-semibold text-slate-600">النتيجة</th>
                    <th className="px-4 py-2 font-semibold text-slate-600">عنوان IP</th>
                    <th className="px-4 py-2 font-semibold text-slate-600">الجهاز/المتصفح</th>
                    <th className="px-4 py-2 font-semibold text-slate-600">الوقت</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {data.recentLogins.map((l) => (
                    <tr key={l.id}>
                      <td className="px-4 py-2 font-medium">{l.employee?.name || l.username}</td>
                      <td className="px-4 py-2">
                        {l.success ? (
                          <span className="rounded-full bg-green-100 px-2 py-0.5 text-xs font-bold text-green-800">نجح</span>
                        ) : (
                          <span className="rounded-full bg-red-100 px-2 py-0.5 text-xs font-bold text-red-800">فشل</span>
                        )}
                      </td>
                      <td className="px-4 py-2 text-slate-500">{l.ipAddress || '—'}</td>
                      <td className="max-w-xs truncate px-4 py-2 text-xs text-slate-400">{l.userAgent || '—'}</td>
                      <td className="px-4 py-2 text-slate-500">{new Date(l.createdAt).toLocaleString('ar-IQ')}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
