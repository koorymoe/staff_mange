import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Empty, Skel, TopBar } from '../components/ui'

// ═══ الخدمات (شاشة 4) — شبكة عمودين بصورة كل خدمة ═══
export default function ServicesPage() {
  const [params] = useSearchParams()
  const [q, setQ] = useState(params.get('q') ?? '')
  const { data, err } = useServices()
  const list = (data ?? []).filter((s) => !q.trim() || s.name.includes(q.trim()) || (s.category ?? '').includes(q.trim()))

  return (
    <div className="page-in">
      <TopBar title="الخدمات" back={false} />
      <div className="px-4 pt-1">
        <div className="flex h-12 items-center gap-2 rounded-xl bg-white px-4 shadow-[0_1px_2px_rgba(15,29,58,.05)] ring-1 ring-line">
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث عن خدمة ..." className="min-w-0 flex-1 bg-transparent text-[14px] outline-none placeholder:text-[#a3acbf]" />
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#0d3b8c" strokeWidth="2.2" strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></svg>
        </div>
        {err && <p className="mt-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-600">{err}</p>}
        <div className="mt-4 grid grid-cols-2 gap-3 pb-4">
          {!data && [0, 1, 2, 3, 4, 5].map((i) => <Skel key={i} className="h-40" />)}
          {list.map((s, i) => (
            <Link key={s.id} to={`/services/${s.id}`} className="touch-card card rise group overflow-hidden transition" style={{ animationDelay: `${Math.min(i, 8) * 35}ms` }}>
              <div className="relative overflow-hidden">
                <ServiceArt id={s.id} name={s.name} hasImage={s.hasImage} className="h-32 w-full" />
                {s.rating && <span className="absolute left-2 top-2 rounded-full bg-white/92 px-2 py-1 text-[10px] font-extrabold text-[#8a5a12] shadow-sm">★ {s.rating}</span>}
              </div>
              <div className="flex items-center gap-1.5 px-3 py-3">
                <span className="min-w-0 flex-1 truncate text-right text-[14px] font-extrabold text-ink">{s.name}</span>
                <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-[#edf3ff] text-navy transition group-hover:-translate-x-0.5">‹</span>
              </div>
            </Link>
          ))}
        </div>
        {data && list.length === 0 && <Empty icon="🔍" title="ما لگينا خدمة بهذا الاسم" />}
      </div>
    </div>
  )
}
