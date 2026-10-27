import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiRequest } from '../api'
import { useSession } from '../session'
import { startAgentRebrand } from '../utils/agentBrand'

type Agent = {
  agent: 'MATRIX' | 'IRIS' | 'QUALITY' | 'BOOKING'
  name: string
  editable: boolean
  default?: string
  enabled?: boolean
  built?: boolean
}
type Status = {
  on: boolean
  ownKey?: boolean
  built?: boolean
  model?: string
  dailyCap?: number
  spentTodayUSD?: number
  dailyLimitUSD?: number
}
const statusKeys = {
  MATRIX: 'matrix',
  IRIS: 'secretary',
  QUALITY: 'quality',
  BOOKING: 'booking',
}
export default function AiAgentsPage() {
  const { employee } = useSession()
  return employee?.actualRole === 'OWNER' ? (
    <OwnerAgents />
  ) : (
    <p className="p-8 text-center text-slate-500">
      هاي الشاشة لمالك النظام فقط.
    </p>
  )
}
function OwnerAgents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [status, setStatus] = useState<Record<string, Status>>({})
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState('')
  const [error, setError] = useState('')
  const [saved, setSaved] = useState('')
  const [pending, setPending] = useState<Agent | null>(null)
  useEffect(() => {
    let alive = true
    apiRequest<Agent[]>('/ai/agents')
      .then((x) => {
        if (alive) setAgents(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    apiRequest<Record<string, Status>>('/ai/status')
      .then((x) => {
        if (alive) setStatus(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload])
  const rename = async (agent: Agent) => {
    const name = (drafts[agent.agent] ?? agent.name).trim()
    if (Array.from(name).length < 2 || Array.from(name).length > 30) {
      setError('الاسم من حرفين إلى ٣٠ حرف.')
      return
    }
    setBusy(agent.agent)
    setError('')
    setSaved('')
    try {
      await apiRequest(`/ai/agents/${agent.agent}`, {
        method: 'PUT',
        body: JSON.stringify({ name }),
      })
      setSaved('انحفظ الاسم.')
      startAgentRebrand()
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy('')
    }
  }
  const toggle = async () => {
    if (!pending || busy) return
    setBusy(pending.agent)
    setError('')
    setSaved('')
    try {
      await apiRequest(`/ai/agents/${pending.agent}/enabled`, {
        method: 'PUT',
        body: JSON.stringify({ enabled: !pending.enabled }),
      })
      setSaved(pending.enabled ? 'انطفى الوكيل.' : 'انشغل الوكيل.')
      setPending(null)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy('')
    }
  }
  return (
    <div className="mx-auto max-w-5xl space-y-5">
      <header>
        <h1 className="text-2xl font-extrabold text-slate-800">
          🤖 وكلاء الذكاء
        </h1>
        <p className="mt-2 text-sm text-slate-500">
          أسماء الوكلاء وحالتهم وحدود الاستخدام. المفاتيح السرية ما تنعرض هنا.
        </p>
      </header>
      {error && (
        <p role="alert" className="rounded-xl bg-red-50 p-3 text-red-700">
          {error}
        </p>
      )}
      {saved && (
        <p role="status" className="text-emerald-700">
          {saved}
        </p>
      )}
      <div className="grid gap-4 sm:grid-cols-2">
        {agents.map((agent) => {
          const s = status[statusKeys[agent.agent]]
          return (
            <article
              className="rounded-2xl border border-slate-200 bg-white p-5"
              key={agent.agent}
            >
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-bold text-slate-800">
                  {agent.name}
                </h2>
                <span
                  className={`rounded-full px-3 py-1 text-xs font-bold ${s?.on ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}
                >
                  {s ? (s.on ? 'مفتاح متوفر' : 'مطفي') : 'دا أجيب الحالة…'}
                </span>
              </div>
              {agent.editable ? (
                <form
                  className="mt-4 flex flex-wrap gap-2"
                  onSubmit={(e) => {
                    e.preventDefault()
                    void rename(agent)
                  }}
                >
                  <label className="min-w-0 flex-1 text-sm text-slate-600">
                    اسم الوكيل
                    <input
                      required
                      minLength={2}
                      maxLength={30}
                      className="inp mt-1 text-base"
                      value={drafts[agent.agent] ?? agent.name}
                      onChange={(e) =>
                        setDrafts((v) => ({
                          ...v,
                          [agent.agent]: e.target.value,
                        }))
                      }
                    />
                  </label>
                  <button
                    disabled={!!busy}
                    className="self-end rounded-xl bg-brand-600 px-4 py-2 font-bold text-white disabled:opacity-50"
                  >
                    {busy === agent.agent ? 'دا أحفظ…' : 'حفظ'}
                  </button>
                </form>
              ) : null}
              {agent.editable ? (
                <p className="mt-2 text-xs text-slate-500">
                  {agent.agent === 'QUALITY' || agent.agent === 'BOOKING'
                    ? 'هذا الاسم يطلع للزبون بالواتساب.'
                    : 'اسم داخلي — يشوفه الكادر بالإشعارات والشاشات، والوكيل يعرّف نفسه بيه.'}
                  {agent.default && agent.name !== agent.default && ` (الأصلي: ${agent.default})`}
                </p>
              ) : (
                <p className="mt-3 text-sm text-slate-500">الاسم ثابت.</p>
              )}
              <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div>
                  <dt className="text-slate-500">الموديل</dt>
                  <dd className="break-all text-slate-800">
                    {s?.model || 'غير محدد'}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">المفتاح</dt>
                  <dd className="text-slate-800">
                    {s ? (s.ownKey ? 'مستقل' : 'مشترك / غير متوفر') : '—'}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">السقف اليومي</dt>
                  <dd className="text-slate-800">
                    {s?.dailyLimitUSD != null
                      ? `$${s.dailyLimitUSD}`
                      : s?.dailyCap != null
                        ? `${s.dailyCap} طلب`
                        : 'غير محدد'}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">صرف اليوم</dt>
                  <dd className="text-slate-800">
                    {s?.spentTodayUSD != null
                      ? `$${s.spentTodayUSD.toFixed(3)}`
                      : 'غير متوفر'}
                  </dd>
                </div>
              </dl>
              <div className="mt-4 flex items-center justify-between border-t border-slate-100 pt-3">
                <span className="text-sm text-slate-600">تشغيل الوكيل</span>
                <button
                  role="switch"
                  aria-label={`تشغيل ${agent.name}`}
                  aria-checked={
                    agent.editable ? (agent.enabled ?? false) : (s?.on ?? false)
                  }
                  disabled={
                    !!busy ||
                    !agent.editable ||
                    agent.built === false ||
                    typeof agent.enabled !== 'boolean'
                  }
                  onClick={() => {
                    setPending(agent)
                    setError('')
                  }}
                  className={`h-7 w-12 rounded-full p-1 disabled:opacity-50 ${(agent.editable ? agent.enabled : s?.on) ? 'bg-emerald-600' : 'bg-slate-200'}`}
                >
                  <span className="block h-5 w-5 rounded-full bg-white" />
                </button>
              </div>
              <p className="mt-2 text-xs text-slate-500">
                {agent.built === false || s?.built === false
                  ? 'ما مشتغل بعد — ينتظر ربط الواتساب وقرار الخصوصية'
                  : !agent.editable
                    ? 'تشغيل هذا الوكيل من إعدادات الخادم.'
                    : typeof agent.enabled !== 'boolean'
                      ? 'حالة التشغيل بعد ما متوفرة من الخادم.'
                      : agent.enabled
                        ? 'مفعّل من إعدادات المالك.'
                        : 'مطفي من إعدادات المالك.'}
              </p>
            </article>
          )
        })}
      </div>
      <div className="flex flex-wrap gap-3">
        <Link
          className="rounded-xl border border-slate-200 bg-white px-4 py-3 font-bold text-slate-700"
          to="/whatsapp-conversations"
        >
          محادثات الواتساب ←
        </Link>
        <Link
          className="rounded-xl bg-brand-600 px-4 py-3 font-bold text-white"
          to="/whatsapp-sandbox"
        >
          افتح وضع التجربة ←
        </Link>
      </div>
      {pending && (
        <div className="fixed inset-0 z-[70] overflow-y-auto bg-black/40 p-4">
          <div
            role="dialog"
            aria-modal="true"
            aria-label="تغيير تشغيل الوكيل"
            className="mx-auto mt-12 max-w-md rounded-2xl bg-white p-5"
          >
            <h2 className="text-lg font-bold text-slate-800">
              {pending.enabled ? 'تطفي' : 'تشغّل'} {pending.name}؟
            </h2>
            <p className="mt-3 text-sm text-slate-500">
              التشغيل الحي يخضع لربط الواتساب وموافقات الخصوصية بالخادم.
            </p>
            {error && (
              <p role="alert" className="mt-3 text-red-700">
                {error}
              </p>
            )}
            <div className="mt-4 flex gap-2">
              <button
                disabled={!!busy}
                onClick={() => void toggle()}
                className="rounded-xl bg-brand-600 px-4 py-2 font-bold text-white"
              >
                تأكيد
              </button>
              <button
                disabled={!!busy}
                onClick={() => setPending(null)}
                className="rounded-xl border border-slate-200 px-4 py-2 text-slate-700"
              >
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
