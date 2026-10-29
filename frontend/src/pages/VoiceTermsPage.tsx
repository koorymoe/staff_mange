import { useEffect, useRef, useState } from 'react'
import { apiRequest } from '../api'
import { useSession } from '../session'
import './VoiceTermsPage.css'

type Source = 'SEED' | 'SYSTEM' | 'OWNER' | 'LEARNED'
type Term = { term: string; meaning: string; category: string; source: Source }
type Fix = { wrong: string; right: string; source: Source }
type Lexicon = { terms: Term[]; fixes: Fix[]; total: number }
type Filter = '' | 'taught' | 'SYSTEM' | 'SEED'
type Removal = { kind: 'term' | 'wrong'; value: string; source: Source }
const sources = {
  OWNER: { label: 'علّمتها انت', style: 'gold' },
  LEARNED: { label: 'علّمتها انت', style: 'gold' },
  SYSTEM: { label: 'من النظام', style: 'blue' },
  SEED: { label: 'الورقة', style: 'gray' },
}
function Badge({ source }: { source: Source }) {
  const info = sources[source] ?? { label: source, style: 'gray' }
  return <span className={`ear-source ${info.style}`}>{info.label}</span>
}

export default function VoiceTermsPage() {
  const { employee } = useSession()
  // Keep the request-bearing component unmounted for every other role.
  return employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN' ? (
    <VoiceTerms key={employee.id} />
  ) : (
    <p role="alert">قاموس الأذن للمالك ومدير النظام بس.</p>
  )
}

function VoiceTerms() {
  const [data, setData] = useState<(Lexicon & { key: string }) | null>(null)
  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [source, setSource] = useState<Filter>('')
  const [tab, setTab] = useState<'terms' | 'fixes'>('terms')
  const [reload, setReload] = useState(0)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [actionError, setActionError] = useState('')
  const [notice, setNotice] = useState('')
  const [term, setTerm] = useState('')
  const [heardAs, setHeardAs] = useState('')
  const [meaning, setMeaning] = useState('')
  const [busy, setBusy] = useState(false)
  const [removal, setRemoval] = useState<Removal | null>(null)
  const mutation = useRef(false)
  const termInput = useRef<HTMLInputElement>(null)
  const returnFocus = useRef<HTMLButtonElement | null>(null)
  const requestKey = JSON.stringify([query, source, reload])
  useEffect(() => {
    if (search.trim() === query) return
    const timer = window.setTimeout(() => {
      setQuery(search.trim())
    }, 300)
    return () => window.clearTimeout(timer)
  }, [search, query])
  useEffect(() => {
    let alive = true
    // The contract accepts ONE exact source. Fetch both taught sources; do not
    // filter the truncated all-source response and silently lose older words.
    const selected = source === 'taught' ? ['OWNER', 'LEARNED'] : [source]
    Promise.all(
      selected.map((s) =>
        apiRequest<Lexicon>(
          `/voice/terms?${new URLSearchParams({ q: query, source: s })}`,
        ),
      ),
    )
      .then((results) => {
        if (!alive) return
        setData({
          key: requestKey,
          terms: results.flatMap((r) => r.terms ?? []),
          fixes: results[0].fixes ?? [],
          total: results[0].total,
        })
        setLoadError('')
        setLoading(false)
      })
      .catch((e: Error) => {
        if (alive) {
          setLoadError(e.message)
          setLoading(false)
        }
      })
    return () => {
      alive = false
    }
  }, [query, source, reload, requestKey])
  function refresh() {
    setLoading(true)
    setLoadError('')
    setReload((n) => n + 1)
  }
  async function save() {
    if (mutation.current || removal || !term.trim()) return
    mutation.current = true
    setBusy(true)
    setActionError('')
    setNotice('')
    try {
      await apiRequest('/voice/terms', {
        method: 'POST',
        body: JSON.stringify({
          term: term.trim(),
          meaning: meaning.trim(),
          heardAs: heardAs.trim(),
        }),
      })
      setTerm('')
      setMeaning('')
      setHeardAs('')
      setNotice('انحفظت الكلمة. إذا الفلتر يخفيها، اختار «الكل» حتى تشوفها.')
      refresh()
      termInput.current?.focus()
    } catch (e) {
      setActionError((e as Error).message)
    } finally {
      mutation.current = false
      setBusy(false)
    }
  }
  function cancelRemoval() {
    setRemoval(null)
    setActionError('')
    window.requestAnimationFrame(() => returnFocus.current?.focus())
  }
  async function remove() {
    if (!removal || mutation.current) return
    mutation.current = true
    setBusy(true)
    setActionError('')
    setNotice('')
    try {
      await apiRequest(
        `/voice/terms?${new URLSearchParams({ [removal.kind]: removal.value })}`,
        { method: 'DELETE' },
      )
      setRemoval(null)
      setNotice(
        removal.kind === 'term'
          ? 'انشالت الكلمة. التصحيحات المرتبطة تبقى لحد ما تشيلها من تبويب التصحيحات.'
          : 'انشال التصحيح. الكلمة تبقى بالقاموس.',
      )
      refresh()
    } catch (e) {
      setActionError((e as Error).message)
    } finally {
      mutation.current = false
      setBusy(false)
    }
  }
  const waiting = loading || search.trim() !== query || data?.key !== requestKey
  // The API returns ALL fixes regardless of q/source. Filter these locally.
  const fixes = (data?.fixes ?? []).filter(
    (fix) =>
      (!source ||
        (source === 'taught'
          ? ['OWNER', 'LEARNED'].includes(fix.source)
          : fix.source === source)) &&
      `${fix.wrong} ${fix.right}`
        .toLocaleLowerCase('ar-IQ')
        .includes(query.toLocaleLowerCase('ar-IQ')),
  )
  const rows = tab === 'terms' ? (data?.terms ?? []) : fixes
  return (
    <main className="voice-terms" dir="rtl">
      <header className="ear-hero">
        <div>
          <span className="ear-eyebrow">كمال · ذاكرة الصوت</span>
          <h1>👂 قاموس الأذن</h1>
          <p>كمال يكتب هالكلمات صح، ويتعلم من تصحيحاتك</p>
        </div>
        <div className="ear-count" aria-live="polite">
          الأذن تعرف{' '}
          <strong>{data ? data.total.toLocaleString('ar-IQ') : '—'}</strong>{' '}
          كلمة
        </div>
      </header>

      <section className="ear-panel" aria-labelledby="ear-teach-title">
        <h2 id="ear-teach-title">علّم الأذن كلمة</h2>
        <p className="ear-muted">
          اكتب الكلمة الصح، وإذا انسمعت غلط ضيف شكلها حتى ما تتكرر.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void save()
          }}
        >
          <label>
            الكلمة الصح
            <input
              ref={termInput}
              value={term}
              onChange={(e) => setTerm(e.target.value)}
              required
              minLength={2}
              disabled={busy || !!removal}
              placeholder="مثلاً: داش كام"
            />
          </label>
          <label>
            انكتبت غلط هيچ (اختياري)
            <input
              value={heardAs}
              onChange={(e) => setHeardAs(e.target.value)}
              minLength={2}
              disabled={busy || !!removal}
              placeholder="مثلاً: دش كاب"
            />
          </label>
          <label>
            معناها
            <input
              value={meaning}
              onChange={(e) => setMeaning(e.target.value)}
              disabled={busy || !!removal}
              placeholder="اشرحها بكلمتين…"
            />
          </label>
          <button
            className="ear-primary"
            disabled={busy || !!removal || !term.trim()}
            type="submit"
          >
            {busy && !removal ? 'دا أحفظ…' : 'حفظ الكلمة ↵'}
          </button>
        </form>
      </section>
      {notice && (
        <p className="ear-notice" role="status">
          ✓ {notice}
        </p>
      )}
      {actionError && !removal && (
        <p className="ear-error" role="alert">
          {actionError}
        </p>
      )}

      <section className="ear-panel" aria-label="محتوى القاموس">
        <div className="ear-tabs" role="group" aria-label="أقسام القاموس">
          <button
            aria-pressed={tab === 'terms'}
            disabled={busy || !!removal}
            onClick={() => setTab('terms')}
          >
            الكلمات
          </button>
          <button
            aria-pressed={tab === 'fixes'}
            disabled={busy || !!removal}
            onClick={() => setTab('fixes')}
          >
            التصحيحات
          </button>
        </div>
        <label className="ear-search">
          دوّر بالقاموس
          <input
            type="search"
            value={search}
            disabled={busy || !!removal}
            onChange={(e) => {
              setSearch(e.target.value)
              setLoadError('')
            }}
            placeholder={
              tab === 'terms'
                ? 'ابحث بالكلمة أو معناها…'
                : 'ابحث بالخطأ أو تصحيحه…'
            }
          />
        </label>
        <div className="ear-chips" role="group" aria-label="مصدر الكلمات">
          {(
            [
              ['', 'الكل'],
              ['taught', 'علّمتها انت'],
              ['SYSTEM', 'النظام'],
              ['SEED', 'الورقة'],
            ] as const
          ).map(([key, label]) => (
            <button
              key={key}
              aria-pressed={source === key}
              disabled={busy || !!removal}
              onClick={() => {
                setSource(key)
                refresh()
              }}
            >
              {label}
            </button>
          ))}
        </div>
        {tab === 'terms' && (
          <p className="ear-muted ear-limit">
            {source === 'taught'
              ? 'يعرض لحد ٥٠٠ كلمة من كل مصدر تعلّم.'
              : 'يعرض أحدث ٥٠٠ كلمة مطابقة كحد أقصى.'}{' '}
            ضيّق البحث إذا ما لگيت كلمتك.
          </p>
        )}
        {loadError ? (
          <p role="alert" className="ear-error">
            {loadError} <button onClick={refresh}>أعد المحاولة</button>
          </p>
        ) : waiting ? (
          <p role="status" className="ear-empty">
            دا أجيب القاموس…
          </p>
        ) : rows.length === 0 ? (
          <div className="ear-empty">
            <span>🎧</span>
            <h3>
              {search || source
                ? 'ما لگينا شي بهالبحث'
                : 'الأذن جاهزة تتعلم منك'}
            </h3>
            <p>
              {search || source
                ? 'جرّب كلمة ثانية أو غيّر المصدر.'
                : 'ضيف أول كلمة من فوق، وكل تصحيح يساعد كمال يسمعك أحسن.'}
            </p>
          </div>
        ) : (
          <ul
            className="ear-list"
            aria-label={tab === 'terms' ? 'كلمات القاموس' : 'تصحيحات الأذن'}
          >
            {rows.map((row) => {
              const isTerm = 'term' in row
              const value = isTerm ? row.term : row.wrong
              const active =
                removal?.kind === (isTerm ? 'term' : 'wrong') &&
                removal.value === value
              return (
                <li key={`${tab}:${value}`}>
                  <div className="ear-row">
                    <div className="ear-word">
                      {isTerm ? (
                        <>
                          <h3>{row.term}</h3>
                          <p>{row.meaning || 'بعد ما مضاف معنى'}</p>
                          {row.category && <small>{row.category}</small>}
                        </>
                      ) : (
                        <div className="ear-fix">
                          <del>{row.wrong}</del>
                          <span aria-hidden="true">←</span>
                          <strong>{row.right}</strong>
                        </div>
                      )}
                    </div>
                    <div className="ear-row-actions">
                      <Badge source={row.source} />
                      <button
                        className="ear-delete"
                        disabled={busy || !!removal}
                        aria-label={`حذف ${isTerm ? 'الكلمة' : 'التصحيح'} ${value}`}
                        onClick={(e) => {
                          returnFocus.current = e.currentTarget
                          setRemoval({
                            kind: isTerm ? 'term' : 'wrong',
                            value,
                            source: row.source,
                          })
                          setActionError('')
                          setNotice('')
                        }}
                      >
                        حذف
                      </button>
                    </div>
                  </div>
                  {active && (
                    <div
                      className="ear-confirm"
                      role="group"
                      aria-label="تأكيد الحذف"
                      onKeyDown={(e) => {
                        if (e.key === 'Escape' && !busy) cancelRemoval()
                      }}
                    >
                      <p>
                        تشيل {isTerm ? 'الكلمة' : 'التصحيح'} «{value}»؟{' '}
                        {isTerm
                          ? 'هذا ما يشيل تصحيحاتها.'
                          : 'هذا ما يشيل الكلمة الصح من القاموس.'}
                      </p>
                      {['SEED', 'SYSTEM'].includes(row.source) && (
                        <p className="ear-muted">
                          كلمات الورقة والنظام ممكن ترجع عند تحديث القاموس
                          تلقائياً.
                        </p>
                      )}
                      {actionError && (
                        <p className="ear-error" role="alert">
                          {actionError}
                        </p>
                      )}
                      <button autoFocus disabled={busy} onClick={cancelRemoval}>
                        لا، خلّيها
                      </button>
                      <button
                        className="ear-delete"
                        disabled={busy}
                        onClick={() => void remove()}
                      >
                        {busy ? 'دا أشيل…' : 'إي، احذف'}
                      </button>
                    </div>
                  )}
                </li>
              )
            })}
          </ul>
        )}
      </section>
    </main>
  )
}
