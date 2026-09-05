import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

type ConversationRow = Awaited<ReturnType<typeof api.getAssistantConversations>>['conversations'][number]

const PAGE_SIZE = 30

export default function AssistantConversationsPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'

  const [employees, setEmployees] = useState<{ id: string; name: string; position: string | null }[]>([])
  const [employeeId, setEmployeeId] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')

  const [rows, setRows] = useState<ConversationRow[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  useEffect(() => {
    if (!isOwner) return
    api.getAssistantConversationEmployees()
      .then((r) => setEmployees(r.employees))
      .catch(() => {})
  }, [isOwner])

  const fetchPage = async (offset: number, replace: boolean) => {
    // Filters (isOwner/employeeId/from/to) can change after mount and re-trigger this
    // effect; re-arm loading via a microtask instead of synchronously in the effect body.
    queueMicrotask(() => { setLoading(true); setError(null) })
    try {
      const res = await api.getAssistantConversations({
        employeeId: employeeId || undefined,
        from: from || undefined,
        to: to || undefined,
        limit: PAGE_SIZE,
        offset,
      })
      setTotal(res.total)
      setRows((prev) => (replace ? res.conversations : [...prev, ...res.conversations]))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر جلب المحادثات')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (!isOwner) return
    // fetchPage re-arms loading/error via queueMicrotask (see above), not synchronously.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchPage(0, true)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOwner, employeeId, from, to])

  if (!isOwner) {
    return (
      <div className="rounded-xl border border-white bg-white p-10 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        هذي الصفحة حصرية لمالك النظام.
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-xl font-bold text-slate-800">محادثات الموظفين مع المساعد الذكي</h2>
        <p className="mt-1 text-sm text-slate-400">مراجعة كل الأسئلة والأجوبة بين الموظفين والمساعد الذكي — حصري لحسابك.</p>
      </div>

      <div className="flex flex-wrap items-end gap-3 rounded-xl border border-slate-100 bg-white p-4 shadow-sm">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">الموظف</label>
          <select
            value={employeeId}
            onChange={(e) => setEmployeeId(e.target.value)}
            className="rounded-lg border border-slate-200 px-3 py-2 text-sm"
          >
            <option value="">كل الموظفين</option>
            {employees.map((emp) => (
              <option key={emp.id} value={emp.id}>{emp.name}</option>
            ))}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">من تاريخ</label>
          <input type="date" value={from} onChange={(e) => setFrom(e.target.value)} className="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-500">إلى تاريخ</label>
          <input type="date" value={to} onChange={(e) => setTo(e.target.value)} className="rounded-lg border border-slate-200 px-3 py-2 text-sm" />
        </div>
        <div className="text-xs text-slate-400">
          {total > 0 && `إجمالي: ${total} محادثة`}
        </div>
      </div>

      {error && (
        <p className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-600">{error}</p>
      )}

      <div className="space-y-3">
        {rows.map((row) => {
          const expanded = expandedId === row.id
          const replyShort = row.reply.length > 220 && !expanded ? row.reply.slice(0, 220) + '…' : row.reply
          return (
            <div key={row.id} className="rounded-xl border border-slate-100 bg-white p-4 shadow-sm">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span className="font-bold text-slate-600">{row.employee?.name || 'موظف محذوف'}</span>
                <span dir="ltr">{new Date(row.createdAt).toLocaleString('ar-IQ')}</span>
              </div>
              <p className="mt-2 text-sm text-slate-800"><span className="font-medium text-slate-500">سؤال: </span>{row.message}</p>
              <p className="mt-1 text-sm text-slate-600">
                <span className="font-medium text-slate-500">جواب: </span>
                {replyShort}
              </p>
              {row.reply.length > 220 && (
                <button
                  onClick={() => setExpandedId(expanded ? null : row.id)}
                  className="mt-1 text-xs font-medium text-blue-600 hover:underline"
                >
                  {expanded ? 'عرض أقل' : 'عرض المزيد'}
                </button>
              )}
            </div>
          )
        })}
        {rows.length === 0 && !loading && (
          <p className="rounded-xl border border-slate-100 bg-white p-6 text-center text-sm text-slate-400">لا توجد محادثات مطابقة.</p>
        )}
      </div>

      {rows.length < total && (
        <div className="text-center">
          <button
            onClick={() => fetchPage(rows.length, false)}
            disabled={loading}
            className="rounded-lg border border-slate-200 bg-white px-5 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-50"
          >
            {loading ? '...جاري التحميل' : 'تحميل المزيد'}
          </button>
        </div>
      )}
    </div>
  )
}
