import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

type DashboardData = Awaited<ReturnType<typeof api.getSecurityDashboard>>
type LoginRow = DashboardData['recentLogins'][number]

// أسماء الأحداث الأمنية بالعربي
const EVENT_LABELS: Record<string, string> = {
  LOGIN_FAILED: '❌ كلمة مرور خاطئة',
  LOGIN_WHILE_LOCKED: '🔒 دخول لحساب محظور',
  ACCOUNT_LOCKED: '🚫 حظر تلقائي',
  ACCOUNT_UNLOCKED: '✅ فك حظر',
  AUTHZ_VIOLATION: '⚠️ وصول غير مخوّل',
  PERMISSIONS_CHANGED: '🔑 تغيير صلاحيات',
}

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}

// يفكك User-Agent البسيط لمتصفح/نظام تشغيل مقروء — تحليل نصي عادي على
// الجهاز، ما يحتاج مكتبة خارجية ولا يرسل بيانات لأي مكان.
function parseUserAgent(ua: string | null): { browser: string; os: string } {
  if (!ua) return { browser: 'غير معروف', os: 'غير معروف' }
  let browser = 'غير معروف'
  if (ua.includes('Edg/')) browser = 'Edge'
  else if (ua.includes('Chrome/') && !ua.includes('OPR/')) browser = 'Chrome'
  else if (ua.includes('Firefox/')) browser = 'Firefox'
  else if (ua.includes('Safari/') && !ua.includes('Chrome')) browser = 'Safari'
  else if (ua.includes('OPR/')) browser = 'Opera'
  else if (ua.includes('curl/')) browser = 'curl (أداة سطر أوامر)'

  let os = 'غير معروف'
  if (ua.includes('Windows NT 10.0')) os = 'Windows 10/11'
  else if (ua.includes('Windows')) os = 'Windows'
  else if (ua.includes('Mac OS X')) os = 'macOS'
  else if (ua.includes('Android')) os = 'Android'
  else if (ua.includes('iPhone') || ua.includes('iPad')) os = 'iOS'
  else if (ua.includes('Linux')) os = 'Linux'

  return { browser, os }
}

function StatCard({ label, value, sub, danger }: { label: string; value: string; sub?: string; danger?: boolean }) {
  return (
    <div className={`rounded-xl border p-5 text-center font-mono ${danger ? 'border-red-500/40 bg-red-950/40' : 'border-emerald-500/20 bg-black/40'}`}>
      <p className={`text-2xl font-bold ${danger ? 'text-red-400' : 'text-emerald-400'}`} dir="ltr">{value}</p>
      <p className="mt-1 text-xs text-emerald-200/50">{label}</p>
      {sub && <p className="mt-0.5 text-[10px] text-emerald-200/30" dir="ltr">{sub}</p>}
    </div>
  )
}

export default function SecurityDashboardPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [data, setData] = useState<DashboardData | null>(null)
  const [history, setHistory] = useState<number[]>([])
  const [error, setError] = useState<string | null>(null)
  const [freeing, setFreeing] = useState(false)
  const [selectedLogin, setSelectedLogin] = useState<LoginRow | null>(null)

  useEffect(() => {
    if (!isOwner) return
    const fetchData = () => {
      api.getSecurityDashboard()
        .then((d) => {
          setData(d)
          setHistory((prev) => [...prev.slice(-19), d.requestsLastMinute])
        })
        .catch((e) => setError(e.message))
    }
    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [isOwner])

  const handleFreeMemory = async () => {
    setFreeing(true)
    try {
      const res = await api.freeServerMemory()
      setData((prev) => (prev ? { ...prev, memoryUsedMB: res.memoryUsedMB } : prev))
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تحرير الذاكرة')
    } finally {
      setFreeing(false)
    }
  }

  if (!isOwner) {
    return (
      <div className="rounded-xl border border-white bg-white p-10 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        هذي الصفحة حصرية لمالك النظام.
      </div>
    )
  }

  const diskPct = data && data.diskTotalGB > 0 ? (data.diskUsedGB / data.diskTotalGB) * 100 : 0

  return (
    <div className="-m-3 min-h-screen rounded-2xl bg-black p-4 font-mono sm:-m-5 sm:p-6 lg:-m-8 lg:p-8">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-emerald-400">root@al-amani-server:~# monitor</h2>
          <p className="mt-1 text-sm text-emerald-200/40">
            صحة السيرفر وسجل تسجيل الدخول — حصري لحسابك، يتحدث تلقائياً كل 5 ثواني.
          </p>
        </div>
        <button
          onClick={handleFreeMemory}
          disabled={freeing}
          className="rounded-lg border border-emerald-500/30 bg-emerald-950/40 px-4 py-2 text-xs font-bold text-emerald-300 transition-colors hover:bg-emerald-900/40 disabled:opacity-50"
        >
          {freeing ? '...جاري' : '🧹 تحرير الذاكرة (Clear Cache)'}
        </button>
      </div>

      {error && (
        <p className="mt-6 rounded-lg border border-red-500/30 bg-red-950/40 p-4 text-red-400">
          تعذر جلب البيانات: {error}
        </p>
      )}

      {data && (
        <>
          <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
            <StatCard label="UPTIME" value={formatUptime(data.serverUptimeSeconds)} />
            <StatCard label="MEMORY" value={`${data.memoryUsedMB.toFixed(1)} MB`} />
            <StatCard label="GOROUTINES" value={`${data.goroutineCount}`} sub={`${data.cpuCount} CPU cores`} />
            <StatCard
              label="FAILED LOGINS (1H)"
              value={`${data.failedLoginsLastHour}`}
              danger={data.failedLoginsLastHour > 5}
            />
            <StatCard label="ONLINE NOW" value={`${data.onlineEmployees}`} sub="last 15 min" />
            <StatCard label="DB SIZE" value={`${data.dbSizeMB.toFixed(1)} MB`} />
            <StatCard label="DB CONNECTIONS" value={`${data.dbConnectionsInUse}/${data.dbConnectionsOpen}`} sub="in use / open" />
            <StatCard label="TOTAL REQUESTS" value={`${data.totalRequests}`} />
          </div>

          {/* مساحة القرص */}
          <div className="mt-4 rounded-xl border border-emerald-500/20 bg-black/40 p-5">
            <div className="flex items-center justify-between text-xs text-emerald-200/60">
              <span>DISK USAGE</span>
              <span dir="ltr">
                {data.diskUsedGB.toFixed(1)} GB / {data.diskTotalGB.toFixed(1)} GB used ({data.diskFreeGB.toFixed(1)} GB free)
              </span>
            </div>
            <div className="mt-2 h-3 w-full overflow-hidden rounded-full bg-emerald-950/60">
              <div
                className={`h-full rounded-full transition-all ${diskPct > 85 ? 'bg-red-500' : diskPct > 60 ? 'bg-amber-500' : 'bg-emerald-500'}`}
                style={{ width: `${Math.min(diskPct, 100)}%` }}
              />
            </div>
          </div>

          {/* شريط الضغط الحي */}
          <div className="mt-4 rounded-xl border border-emerald-500/20 bg-black/40 p-5">
            <div className="flex items-center justify-between">
              <h3 className="text-xs text-emerald-200/60">LIVE REQUEST LOAD</h3>
              <span className="flex items-center gap-1.5 text-xs text-emerald-400">
                <span className="relative flex h-2 w-2">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
                </span>
                LIVE
              </span>
            </div>
            <p className="mt-2 text-2xl font-bold text-emerald-400" dir="ltr">{data.requestsLastMinute} req/min</p>
            {history.length > 1 && (
              <div className="mt-3 flex h-14 items-end gap-1">
                {history.map((v, i) => {
                  const max = Math.max(...history, 1)
                  return (
                    <div
                      key={i}
                      className="flex-1 rounded-t bg-gradient-to-t from-emerald-600 to-emerald-400"
                      style={{ height: `${Math.max((v / max) * 100, 4)}%` }}
                      title={`${v} req`}
                    />
                  )
                })}
              </div>
            )}
          </div>

          {/* الحسابات المحظورة تلقائياً — المالك وحده يفك الحظر */}
          <div className="mt-4 overflow-hidden rounded-xl border border-red-500/30 bg-black/40">
            <h3 className="border-b border-red-500/20 px-4 py-3 text-xs text-red-300/80">
              🔒 حسابات محظورة تلقائياً ({data.lockedEmployees?.length || 0}) — ما تنفتح إلا بيدك
            </h3>
            {(data.lockedEmployees?.length || 0) === 0 ? (
              <p className="px-4 py-6 text-center text-xs text-emerald-200/40">ماكو حساب محظور حالياً</p>
            ) : (
              <div className="divide-y divide-red-500/10">
                {data.lockedEmployees!.map((e) => (
                  <div key={e.id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
                    <div className="text-xs text-emerald-100">
                      <span className="font-bold text-red-300">{e.name}</span>
                      <span className="text-emerald-200/40"> ({e.username || '—'} · {e.role})</span>
                      <div className="mt-0.5 text-emerald-200/60">
                        {e.lockedReason === 'FAILED_LOGINS'
                          ? `❌ ${e.failedLoginStreak} محاولات كلمة مرور خاطئة`
                          : e.lockedReason === 'AUTHZ_ABUSE'
                            ? `🚫 ${e.lockedDetail || 'حاول يوصل لعملية مو مخوّل لها'}`
                            : e.lockedDetail || e.lockedReason}
                        {e.lockedAt && (
                          <span className="text-emerald-200/30"> · {new Date(e.lockedAt).toLocaleString('ar-IQ')}</span>
                        )}
                      </div>
                    </div>
                    <button
                      onClick={async () => {
                        if (!confirm(`فك الحظر عن ${e.name}؟`)) return
                        try {
                          await api.unlockEmployee(e.id)
                          setData(await api.getSecurityDashboard())
                        } catch (err) {
                          alert(err instanceof Error ? err.message : 'تعذر فك الحظر')
                        }
                      }}
                      className="rounded-lg border border-emerald-500/40 px-3 py-1.5 text-xs font-medium text-emerald-300 hover:bg-emerald-500/10"
                    >
                      ✅ إعادة التفعيل
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* سجل الأحداث الأمنية */}
          <div className="mt-4 overflow-hidden rounded-xl border border-amber-500/20 bg-black/40">
            <h3 className="border-b border-amber-500/20 px-4 py-3 text-xs text-amber-200/70">
              ⚠️ SECURITY EVENTS — محاولات الدخول الفاشلة، الوصول غير المخوّل، الحظر، وتغيير الصلاحيات
            </h3>
            <div className="max-h-[420px] overflow-y-auto">
              {(data.securityEvents?.length || 0) === 0 ? (
                <p className="px-4 py-6 text-center text-xs text-emerald-200/40">ماكو أحداث مسجّلة</p>
              ) : (
                <table className="min-w-full divide-y divide-amber-500/10 text-right text-xs text-emerald-100">
                  <thead className="sticky top-0 bg-black">
                    <tr className="text-emerald-200/50">
                      <th className="px-4 py-2 font-normal">الحدث</th>
                      <th className="px-4 py-2 font-normal">الموظف</th>
                      <th className="px-4 py-2 font-normal">التفاصيل</th>
                      <th className="px-4 py-2 font-normal">IP</th>
                      <th className="px-4 py-2 font-normal">الوقت</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-amber-500/10">
                    {data.securityEvents!.map((ev) => (
                      <tr key={ev.id}>
                        <td className="px-4 py-2 whitespace-nowrap">{EVENT_LABELS[ev.kind] || ev.kind}</td>
                        <td className="px-4 py-2">{ev.employeeName || ev.employeeId || '—'}</td>
                        <td className="px-4 py-2 text-emerald-200/60">{ev.detail || '—'}</td>
                        <td className="px-4 py-2 text-emerald-200/40" dir="ltr">{ev.ip || '—'}</td>
                        <td className="px-4 py-2 text-emerald-200/40 whitespace-nowrap">
                          {new Date(ev.createdAt).toLocaleString('ar-IQ')}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </div>

          {/* سجل الدخول */}
          <div className="mt-4 overflow-hidden rounded-xl border border-emerald-500/20 bg-black/40">
            <h3 className="border-b border-emerald-500/20 px-4 py-3 text-xs text-emerald-200/60">
              LOGIN AUDIT LOG (LAST 100) — اضغط على أي سطر لتفاصيل الجهاز
            </h3>
            <div className="max-h-[500px] overflow-y-auto overflow-x-auto">
              <table className="min-w-full divide-y divide-emerald-500/10 text-right text-xs text-emerald-100">
                <thead className="sticky top-0 bg-black">
                  <tr className="text-emerald-200/50">
                    <th className="px-4 py-2 font-normal">USER</th>
                    <th className="px-4 py-2 font-normal">RESULT</th>
                    <th className="px-4 py-2 font-normal">IP</th>
                    <th className="px-4 py-2 font-normal">DEVICE</th>
                    <th className="px-4 py-2 font-normal">TIME</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-emerald-500/10">
                  {data.recentLogins.map((l) => {
                    const { browser, os } = parseUserAgent(l.userAgent)
                    return (
                      <tr
                        key={l.id}
                        onClick={() => setSelectedLogin(l)}
                        className="cursor-pointer transition-colors hover:bg-emerald-950/40"
                      >
                        <td className="px-4 py-2 font-medium">{l.employee?.name || l.username}</td>
                        <td className="px-4 py-2">
                          {l.success ? (
                            <span className="rounded-full bg-emerald-900/60 px-2 py-0.5 font-bold text-emerald-300">OK</span>
                          ) : (
                            <span className="rounded-full bg-red-900/60 px-2 py-0.5 font-bold text-red-300">FAIL</span>
                          )}
                        </td>
                        <td className="px-4 py-2 text-emerald-200/60">{l.ipAddress || '—'}</td>
                        <td className="px-4 py-2 text-emerald-200/60">{browser} / {os}</td>
                        <td className="px-4 py-2 text-emerald-200/40">
                          {new Date(l.createdAt).toISOString().replace('T', ' ').slice(0, 19)}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* تفاصيل الجهاز */}
      {selectedLogin && (
        <div
          className="fixed inset-0 z-[70] flex items-center justify-center bg-black/70 p-4"
          onClick={() => setSelectedLogin(null)}
        >
          <div
            onClick={(e) => e.stopPropagation()}
            className="w-full max-w-lg rounded-xl border border-emerald-500/30 bg-black p-6 font-mono text-emerald-100"
          >
            <h3 className="text-sm font-bold text-emerald-400">DEVICE DETAILS</h3>
            <div className="mt-4 space-y-2 text-xs">
              <p><span className="text-emerald-200/40">User:</span> {selectedLogin.employee?.name || selectedLogin.username}</p>
              <p><span className="text-emerald-200/40">Result:</span> {selectedLogin.success ? 'SUCCESS' : 'FAILED'}</p>
              <p><span className="text-emerald-200/40">IP Address:</span> {selectedLogin.ipAddress || '—'}</p>
              <p><span className="text-emerald-200/40">Browser:</span> {parseUserAgent(selectedLogin.userAgent).browser}</p>
              <p><span className="text-emerald-200/40">OS:</span> {parseUserAgent(selectedLogin.userAgent).os}</p>
              <p><span className="text-emerald-200/40">Time:</span> {new Date(selectedLogin.createdAt).toISOString().replace('T', ' ').slice(0, 19)}</p>
              <p className="break-all"><span className="text-emerald-200/40">Full User-Agent:</span> {selectedLogin.userAgent || '—'}</p>
              <p className="mt-3 text-[10px] text-emerald-200/30">
                ملاحظة: عنوان MAC الحقيقي للجهاز غير متاح تقنياً عبر المتصفح — قيد أمان بكل المتصفحات الحديثة.
              </p>
            </div>
            <button
              onClick={() => setSelectedLogin(null)}
              className="mt-5 w-full rounded-lg border border-emerald-500/30 py-2 text-xs text-emerald-300 hover:bg-emerald-950/40"
            >
              إغلاق
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
