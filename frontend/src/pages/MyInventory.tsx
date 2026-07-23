import { useEffect, useState } from 'react'
import { api, type PersonalTool, type VehicleTool, type OnDemandTool, type ToolRequest } from '../api'
import { useSession } from '../session'

type TabKey = 'checklist' | 'vehicle' | 'request' | 'my-requests'

const tabs: { key: TabKey; label: string; icon: string }[] = [
  { key: 'checklist', label: 'جرد أدواتي', icon: '📋' },
  { key: 'vehicle', label: 'أدوات السيارة', icon: '🚗' },
  { key: 'request', label: 'طلب أداة', icon: '📦' },
  { key: 'my-requests', label: 'طلباتي', icon: '📝' },
]

const statusLabels: Record<string, string> = {
  AVAILABLE: 'متوفرة',
  CHECKED_OUT: 'مستخدمة',
  DAMAGED: 'تالفة',
  PENDING: 'بانتظار الموافقة',
  APPROVED: 'تمت الموافقة',
  REJECTED: 'مرفوض',
  RETURNED: 'تم الإرجاع',
}

const statusColors: Record<string, string> = {
  AVAILABLE: 'bg-green-100 text-green-700',
  CHECKED_OUT: 'bg-yellow-100 text-yellow-700',
  DAMAGED: 'bg-red-100 text-red-700',
  PENDING: 'bg-yellow-100 text-yellow-700',
  APPROVED: 'bg-green-100 text-green-700',
  REJECTED: 'bg-red-100 text-red-700',
  RETURNED: 'bg-gray-100 text-gray-600',
}

export default function MyInventory() {
  const { employee } = useSession()
  const [activeTab, setActiveTab] = useState<TabKey>('checklist')
  const [personalTools, setPersonalTools] = useState<PersonalTool[]>([])
  const [vehicleTools, setVehicleTools] = useState<VehicleTool[]>([])
  const [onDemandTools, setOnDemandTools] = useState<OnDemandTool[]>([])
  const [myRequests, setMyRequests] = useState<ToolRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [checkMap, setCheckMap] = useState<Record<string, boolean>>({})
  const [checkDone, setCheckDone] = useState(false)
  const [requesting, setRequesting] = useState(false)

  const load = async () => {
    if (!employee) return
    setLoading(true)
    try {
      const [pt, vt, od, tr] = await Promise.all([
        api.getPersonalTools(employee.id),
        api.getVehicleTools(),
        api.getOnDemandTools(),
        api.getToolRequests(employee.id),
      ])
      setPersonalTools(pt)
      setVehicleTools(vt)
      setOnDemandTools(od)
      setMyRequests(tr)
    } catch (e: any) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [employee?.id])

  const [submittingCheck, setSubmittingCheck] = useState(false)

  const toggleCheck = (id: string) => {
    setCheckMap((prev) => ({ ...prev, [id]: !prev[id] }))
  }

  const allChecked = personalTools.length > 0 && personalTools.every((t) => checkMap[t.id])

  // تسجل نتيجة الجرد للإداري حتى لو اكو نقص، بدل ما تكون كتلة على تحديد كل الأدوات —
  // الهدف إنه الإداري يعرف بالنقص قبل الخروج للحجز ويوفر البديل، مو يمنع الفني من الإرسال
  const confirmChecklist = async () => {
    const missing = personalTools.filter((t) => !checkMap[t.id]).map((t) => t.name)
    setSubmittingCheck(true)
    try {
      await api.createInventoryCheck({
        complete: missing.length === 0,
        missingItems: missing.length > 0 ? missing.join('، ') : undefined,
      })
      setCheckDone(true)
    } catch (e: any) {
      alert(e.message || 'تعذر تسجيل الجرد')
    } finally {
      setSubmittingCheck(false)
    }
  }

  const handleRequestTool = async (toolId: string) => {
    if (!employee) return
    setRequesting(true)
    try {
      await api.createToolRequest({ employeeId: employee.id, toolId })
      await load()
      setActiveTab('my-requests')
    } catch (e: any) {
      alert(e.message || 'حدث خطأ')
    } finally {
      setRequesting(false)
    }
  }

  const checkedCount = Object.values(checkMap).filter(Boolean).length

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">جرد الأدوات والمعدات</h2>
      <p className="mt-1 text-slate-500">تأكد من أدواتك قبل الخروج واطلب ما تحتاجه</p>

      {/* Tabs */}
      <div className="mt-6 flex gap-1 rounded-xl bg-slate-100 p-1">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-all ${
              activeTab === tab.key
                ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow-md'
                : 'text-slate-600 hover:text-brand-700'
            }`}
          >
            {tab.icon} {tab.label}
          </button>
        ))}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6">

          {/* ===== جرد أدواتي ===== */}
          {activeTab === 'checklist' && (
            <div>
              {/* Progress bar */}
              <div className="mb-4 rounded-xl bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="mb-2 flex items-center justify-between text-sm">
                  <span className="font-semibold text-brand-900">تقدم الجرد</span>
                  <span className="font-bold text-brand-700">{checkedCount} / {personalTools.length}</span>
                </div>
                <div className="h-3 overflow-hidden rounded-full bg-slate-100">
                  <div
                    className="h-full rounded-full bg-gradient-to-l from-brand-500 to-brand-800 transition-all duration-500"
                    style={{ width: personalTools.length ? `${(checkedCount / personalTools.length) * 100}%` : '0%' }}
                  />
                </div>
              </div>

              {personalTools.length === 0 ? (
                <div className="rounded-xl bg-white p-12 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="text-4xl mb-3">🔧</div>
                  <p className="text-slate-400 text-lg">لا توجد أدوات خاصة مسجلة باسمك</p>
                  <p className="text-slate-400 text-sm mt-1">تواصل مع الإدارة لإضافة أدواتك</p>
                </div>
              ) : (
                <>
                  <div className="space-y-2">
                    {personalTools.map((tool) => (
                      <div
                        key={tool.id}
                        onClick={() => !checkDone && toggleCheck(tool.id)}
                        className={`flex cursor-pointer items-center gap-4 rounded-xl bg-white p-4 shadow-[0_2px_10px_rgba(15,32,64,0.05)] transition-all hover:shadow-md ${
                          checkMap[tool.id] ? 'border-2 border-green-400 bg-green-50/50' : 'border-2 border-transparent'
                        }`}
                      >
                        <div className={`flex h-10 w-10 items-center justify-center rounded-full text-lg transition-all ${
                          checkMap[tool.id] ? 'bg-green-500 text-white' : 'bg-slate-100 text-slate-400'
                        }`}>
                          {checkMap[tool.id] ? '✓' : '○'}
                        </div>
                        <div className="flex-1">
                          <div className="font-bold text-brand-900">{tool.name}</div>
                          {tool.barcode && (
                            <div className="text-xs font-mono text-slate-400 mt-0.5">باركود: {tool.barcode}</div>
                          )}
                        </div>
                        <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[tool.status] || 'bg-slate-100 text-slate-600'}`}>
                          {statusLabels[tool.status] || tool.status}
                        </span>
                      </div>
                    ))}
                  </div>

                  {!checkDone ? (
                    <button
                      onClick={confirmChecklist}
                      disabled={submittingCheck || personalTools.length === 0}
                      className={`mt-6 w-full rounded-xl py-4 text-lg font-bold shadow-lg transition-all disabled:opacity-50 ${
                        allChecked
                          ? 'bg-gradient-to-l from-green-500 to-green-700 text-white hover:shadow-xl cursor-pointer'
                          : 'bg-gradient-to-l from-amber-500 to-amber-600 text-white hover:shadow-xl cursor-pointer'
                      }`}
                    >
                      {submittingCheck
                        ? 'جارٍ الإرسال...'
                        : allChecked
                          ? '✅ تأكيد الجرد - جاهز للخروج'
                          : `⚠️ إرسال الجرد بوجود نقص (${personalTools.length - checkedCount} أداة ناقصة)`}
                    </button>
                  ) : (
                    <div className={`mt-6 rounded-xl border-2 p-6 text-center ${allChecked ? 'bg-green-50 border-green-300' : 'bg-amber-50 border-amber-300'}`}>
                      <div className="text-3xl mb-2">{allChecked ? '✅' : '⚠️'}</div>
                      <p className={`font-bold text-lg ${allChecked ? 'text-green-800' : 'text-amber-800'}`}>
                        {allChecked ? 'تم تأكيد الجرد بنجاح' : 'تم إرسال الجرد — فيه نقص'}
                      </p>
                      <p className={`text-sm mt-1 ${allChecked ? 'text-green-600' : 'text-amber-700'}`}>
                        {allChecked ? 'جميع الأدوات متوفرة - يمكنك الخروج' : 'تم إبلاغ الإدارة بالنقص لتوفير البديل قبل الخروج'}
                      </p>
                    </div>
                  )}
                </>
              )}
            </div>
          )}

          {/* ===== أدوات السيارة ===== */}
          {activeTab === 'vehicle' && (
            <div>
              {vehicleTools.length === 0 ? (
                <div className="rounded-xl bg-white p-12 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="text-4xl mb-3">🚗</div>
                  <p className="text-slate-400 text-lg">لا توجد أدوات مركبات مسجلة</p>
                </div>
              ) : (
                <div className="overflow-hidden rounded-xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">لوحة المركبة</th>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        <th className="px-4 py-3 text-sm font-semibold">الباركود</th>
                        <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {vehicleTools.map((t) => (
                        <tr key={t.id} className="hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">{t.vehicleId}</td>
                          <td className="px-4 py-3">{t.name}</td>
                          <td className="px-4 py-3 font-mono text-sm text-slate-500">{t.barcode}</td>
                          <td className="px-4 py-3">
                            <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[t.status] || 'bg-slate-100'}`}>
                              {statusLabels[t.status] || t.status}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  </div>
                </div>
              )}
            </div>
          )}

          {/* ===== طلب أداة ===== */}
          {activeTab === 'request' && (
            <div>
              {onDemandTools.length === 0 ? (
                <div className="rounded-xl bg-white p-12 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="text-4xl mb-3">📦</div>
                  <p className="text-slate-400 text-lg">لا توجد أدوات متاحة للطلب</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                  {onDemandTools.map((tool) => {
                    const available = tool.availableQuantity > 0
                    const alreadyRequested = myRequests.some(
                      (r) => r.toolId === tool.id && (r.status === 'PENDING' || r.status === 'APPROVED')
                    )
                    return (
                      <div key={tool.id} className="rounded-xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        <div className="mb-3 flex items-center gap-3">
                          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-50 text-2xl">🔧</div>
                          <div>
                            <div className="font-bold text-brand-900">{tool.name}</div>
                            {tool.barcode && <div className="text-xs font-mono text-slate-400">{tool.barcode}</div>}
                          </div>
                        </div>
                        <div className="mb-3 flex items-center justify-between text-sm">
                          <span className="text-slate-500">الكمية المتاحة</span>
                          <span className={`font-bold ${available ? 'text-green-600' : 'text-red-600'}`}>
                            {tool.availableQuantity} / {tool.totalQuantity}
                          </span>
                        </div>
                        {alreadyRequested ? (
                          <div className="rounded-lg bg-yellow-50 px-4 py-2.5 text-center text-sm font-medium text-yellow-700">
                            لديك طلب نشط لهذه الأداة
                          </div>
                        ) : (
                          <button
                            onClick={() => handleRequestTool(tool.id)}
                            disabled={!available || requesting}
                            className={`w-full rounded-lg py-2.5 text-sm font-bold transition-all ${
                              available
                                ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white hover:shadow-md cursor-pointer'
                                : 'bg-slate-200 text-slate-400 cursor-not-allowed'
                            }`}
                          >
                            {available ? '📋 طلب هذه الأداة' : 'غير متاحة'}
                          </button>
                        )}
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          )}

          {/* ===== طلباتي ===== */}
          {activeTab === 'my-requests' && (
            <div>
              {myRequests.length === 0 ? (
                <div className="rounded-xl bg-white p-12 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="text-4xl mb-3">📝</div>
                  <p className="text-slate-400 text-lg">لا توجد طلبات سابقة</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {myRequests.map((req) => (
                    <div key={req.id} className="flex items-center gap-4 rounded-xl bg-white p-4 shadow-[0_2px_10px_rgba(15,32,64,0.05)]">
                      <div className={`flex h-10 w-10 items-center justify-center rounded-full text-lg ${
                        req.status === 'APPROVED' ? 'bg-green-100 text-green-600' :
                        req.status === 'PENDING' ? 'bg-yellow-100 text-yellow-600' :
                        req.status === 'REJECTED' ? 'bg-red-100 text-red-600' :
                        'bg-gray-100 text-gray-500'
                      }`}>
                        {req.status === 'APPROVED' ? '✓' : req.status === 'PENDING' ? '⏳' : req.status === 'REJECTED' ? '✕' : '↩'}
                      </div>
                      <div className="flex-1">
                        <div className="font-bold text-brand-900">{req.tool?.name || '-'}</div>
                        <div className="text-xs text-slate-400 mt-0.5">
                          {new Date(req.requestedAt).toLocaleDateString('ar-IQ')}
                        </div>
                      </div>
                      <span className={`rounded-full px-4 py-1.5 text-xs font-bold ${statusColors[req.status]}`}>
                        {statusLabels[req.status]}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
