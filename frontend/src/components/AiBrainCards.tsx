import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import './AiBrains.css'

type Brain = 'internal' | 'external'
type BrainStatus = {
  model?: string
  chatModel?: string
  spentTodayUSD?: number
  dailyLimitUSD?: number
}
const titles: Record<Brain, string> = {
  internal: 'الدماغ الداخلي — ايريس وماتركس',
  external: 'الدماغ الخارجي — الجودة والحجوزات',
}

export default function AiBrainCards({
  owner,
  onChanged,
}: {
  owner: boolean
  onChanged?: () => void
}) {
  const [brains, setBrains] = useState<Partial<Record<Brain, BrainStatus>>>({})
  const [drafts, setDrafts] = useState<Partial<Record<Brain, string>>>({})
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<Brain | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  useEffect(() => {
    // Luffy confirmed both budget status and editing remain owner-only.
    if (!owner) return
    let alive = true
    apiRequest<{ brains?: Partial<Record<Brain, BrainStatus>> }>('/ai/status')
      .then((value) => {
        if (!alive) return
        if (!value.brains?.internal || !value.brains.external)
          throw new Error('حالة الدماغين بعد ما متوفرة من الخادم.')
        setBrains(value.brains)
        setDrafts((old) => ({
          internal:
            old.internal ?? String(value.brains?.internal?.dailyLimitUSD ?? ''),
          external:
            old.external ?? String(value.brains?.external?.dailyLimitUSD ?? ''),
        }))
        setError('')
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [owner, reload])
  const save = async (brain: Brain) => {
    if (!owner || busy) return
    const dailyUSD = Number(drafts[brain])
    if (!Number.isFinite(dailyUSD) || dailyUSD < 0.1 || dailyUSD > 50) {
      setError('السقف لازم من ٠.١ إلى ٥٠ دولار باليوم.')
      return
    }
    setBusy(brain)
    setError('')
    setSuccess('')
    try {
      const value = await apiRequest<BrainStatus>(`/ai/budgets/${brain}`, {
        method: 'PUT',
        body: JSON.stringify({ dailyUSD }),
      })
      setBrains((old) => ({ ...old, [brain]: { ...old[brain], ...value } }))
      setSuccess('انحفظ السقف اليومي.')
      setReload((n) => n + 1)
      onChanged?.()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(null)
    }
  }
  if (!owner) return null
  return (
    <section className="ai-brain-section" aria-label="الدماغين">
      {error && (
        <p role="alert" className="ai-brain-notice">
          {error}{' '}
          {owner && (
            <button onClick={() => setReload((n) => n + 1)}>
              أعد جلب الحالة
            </button>
          )}
        </p>
      )}
      {success && <p role="status">{success}</p>}
      <div className="ai-brain-grid">
        {(['internal', 'external'] as const).map((brain) => {
          const state = brains[brain]
          const spent = state?.spentTodayUSD
          const limit = state?.dailyLimitUSD
          const known =
            typeof spent === 'number' &&
            Number.isFinite(spent) &&
            typeof limit === 'number' &&
            Number.isFinite(limit) &&
            limit > 0
          const percent = known
            ? Math.max(0, Math.min(100, (spent / limit) * 100))
            : 0
          return (
            <article key={brain} className="ai-brain-card">
              <span>
                🧠 {brain === 'internal' ? 'داخل الشركة' : 'خدمة الزبائن'}
              </span>
              <h2>{titles[brain]}</h2>
              <p className="ai-brain-spend">
                <bdi dir="ltr">
                  {known
                    ? `$${spent.toFixed(3)} / $${limit.toFixed(2)}`
                    : '— / —'}
                </bdi>{' '}
                <small>صرف اليوم / السقف</small>
              </p>
              {known ? (
                <div
                  className="ai-brain-progress"
                  role="progressbar"
                  aria-label={`استخدام ${titles[brain]}`}
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(percent)}
                >
                  <span style={{ width: `${percent}%` }} />
                </div>
              ) : (
                <p className="ai-brain-muted">
                  {error
                    ? 'بيانات الدماغ غير متوفرة حالياً.'
                    : 'دا أجيب بيانات الدماغ…'}
                </p>
              )}
              <p className="ai-brain-model">
                الموديل:{' '}
                <b dir="ltr">
                  {brain === 'internal'
                    ? state?.chatModel || state?.model || '—'
                    : state?.model || '—'}
                </b>
              </p>
              {brain === 'internal' &&
                state?.chatModel &&
                state.model !== state.chatModel && (
                  <small>بقية المهام: {state.model || '—'}</small>
                )}
              {owner && (
                <form
                  onSubmit={(e) => {
                    e.preventDefault()
                    void save(brain)
                  }}
                >
                  <label>
                    السقف اليومي بالدولار —{' '}
                    {brain === 'internal' ? 'الداخلي' : 'الخارجي'}
                    <input
                      required
                      type="number"
                      min={0.1}
                      max={50}
                      step="any"
                      value={drafts[brain] ?? ''}
                      onChange={(e) =>
                        setDrafts((old) => ({
                          ...old,
                          [brain]: e.target.value,
                        }))
                      }
                    />
                  </label>
                  <button disabled={!!busy || !state}>
                    {busy === brain ? 'دا أحفظ…' : 'حفظ السقف'}
                  </button>
                </form>
              )}
            </article>
          )
        })}
      </div>
    </section>
  )
}
