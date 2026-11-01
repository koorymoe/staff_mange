import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// 🏆 تصنيف المدراء (دازاي 11-01): مدير إدارة مقابل مدير إدارة — على الهيكل الجديد
type Criterion = { key: string; label: string; weight: number; score: number | null; detail: string }
type Row = {
  rank: number
  divisionName: string
  employeeName: string
  acting: boolean
  people: number
  total: number | null
  criteria: Criterion[]
}
type Ranking = { period: string; rows: Row[] }

const thisMonth = () => new Date().toISOString().slice(0, 7)
const color = (v: number | null) =>
  v === null ? '#94a3b8' : v >= 75 ? '#16a34a' : v >= 50 ? '#d97706' : '#dc2626'
const medal = (r: number) => (r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : `#${r}`)

export default function DirectorsRankingPage() {
  const [period, setPeriod] = useState(thisMonth())
  const [data, setData] = useState<Ranking | null>(null)
  const [error, setError] = useState('')
  useEffect(() => {
    let alive = true
    apiRequest<Ranking>(`/org/directors-ranking?period=${period}`)
      .then((d) => alive && setData(d))
      .catch((e: Error) => alive && setError(e.message))
    return () => {
      alive = false
    }
  }, [period])
  return (
    <div className="page" style={{ padding: 16, maxWidth: 1100, margin: '0 auto' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
        <h1 style={{ margin: 0 }}>🏆 تصنيف المدراء</h1>
        <input type="month" value={period} onChange={(e) => {
          if (!e.target.value) return
          setData(null)
          setError('')
          setPeriod(e.target.value)
        }} />
      </div>
      <p style={{ color: '#64748b', fontSize: 13 }}>
        مدير كل إدارة مقابل الباقين، حسب: المخرجات ٣٠٪ · إنجاز الخطة ٢٠٪ · السرعة ١٥٪ · حل المشاكل ١٥٪ ·
        قلة المشاكل ١٠٪ · الشكاوي والتأخير ١٠٪. المعيار الي ماله بيانات ما ينحسب عليه. الأدمن والمالك ما ينقيّمون.
      </p>
      {error && <div className="error">{error}</div>}
      {!data && !error && <div>جاري التحميل…</div>}
      {data && data.rows.length === 0 && <div>ماكو مدراء إدارات معيّنين بالهيكل بعد.</div>}
      <div style={{ display: 'grid', gap: 12 }}>
        {data?.rows.map((r) => (
          <div key={r.divisionName} className="card" style={{ padding: 14, borderRadius: 12, border: '1px solid #e2e8f0', background: 'var(--card-bg, #fff)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
              <span style={{ fontSize: 22, minWidth: 40 }}>{r.rank ? medal(r.rank) : '—'}</span>
              <div style={{ flex: 1, minWidth: 160 }}>
                <b>{r.employeeName}</b>
                {r.acting && <span style={{ fontSize: 11, marginInlineStart: 6, color: '#d97706' }}>بالوكالة</span>}
                <div style={{ fontSize: 12, color: '#64748b' }}>
                  مدير {r.divisionName} · {r.people} موظف
                </div>
              </div>
              <span style={{ fontSize: 26, fontWeight: 700, color: color(r.total) }}>
                {r.total === null ? 'ماكو بيانات' : `${r.total}`}
              </span>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: 8, marginTop: 10 }}>
              {r.criteria.map((c) => (
                <div key={c.key} style={{ fontSize: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span>{c.label} <span style={{ color: '#94a3b8' }}>{c.weight}٪</span></span>
                    <b style={{ color: color(c.score) }}>{c.score === null ? '—' : c.score}</b>
                  </div>
                  <div style={{ height: 6, background: '#e2e8f0', borderRadius: 3, overflow: 'hidden' }}>
                    <div style={{ width: `${c.score ?? 0}%`, height: '100%', background: color(c.score) }} />
                  </div>
                  <div style={{ color: '#94a3b8', marginTop: 2 }}>{c.detail}</div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
