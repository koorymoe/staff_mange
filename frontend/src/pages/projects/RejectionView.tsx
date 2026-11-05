import { useMemo, useState } from 'react'
import { matches } from '../../utils/search'
import { todayBaghdad } from '../../utils/format'
import { type Project, CATEGORIES, categoryColor, parseRejection, restoreRejected, useCanPurge, purgeRejected } from './shared'
import { RejStat } from './ui'

// ---------------------------------------------------------------------------
// Rejection analytics view
// ---------------------------------------------------------------------------
export function RejectionView({ projects, canManage, onRefresh }: { projects: Project[]; canManage: boolean; onRefresh: () => void }) {
  const canPurge = useCanPurge()
  const [category, setCategory] = useState('all')
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<'count' | 'date' | 'name'>('count')
  const [expanded, setExpanded] = useState<string | null>(null)

  const rejected = projects.filter(p => p.stage.includes('مرفوض'))
  const total = projects.length || 1

  const grouped = useMemo(() => {
    const map: Record<string, { reason: string; category: string; projects: Project[]; lastDate: string | null }> = {}
    for (const p of rejected) {
      const { reason, category: cat } = parseRejection(p.task)
      if (category !== 'all' && cat !== category) continue
      if (!map[reason]) map[reason] = { reason, category: cat, projects: [], lastDate: null }
      map[reason].projects.push(p)
      const d = p.createdAt
      if (d && (!map[reason].lastDate || d > map[reason].lastDate!)) map[reason].lastDate = d
    }
    let list = Object.values(map)
    if (search.trim()) list = list.filter(r => matches([r.reason], search) ||
      r.projects.some(p => matches([p.name, p.location], search)))
    if (sort === 'count') list.sort((a, b) => b.projects.length - a.projects.length)
    else if (sort === 'date') list.sort((a, b) => new Date(b.lastDate || 0).getTime() - new Date(a.lastDate || 0).getTime())
    else list.sort((a, b) => a.reason.localeCompare(b.reason, 'ar'))
    return list
  }, [rejected, category, search, sort])

  const exportCsv = () => {
    if (rejected.length === 0) { alert('لا توجد مشاريع مرفوضة لتصديرها!'); return }
    let csv = '﻿الفئة,السبب,اسم المشروع,الموقع,الهاتف,التاريخ,ملاحظات\n'
    for (const p of rejected) {
      const { reason, notes, category: cat } = parseRejection(p.task)
      if (category !== 'all' && cat !== category) continue
      const date = p.createdAt ? new Date(p.createdAt).toLocaleDateString('ar-IQ') : ''
      csv += `"${cat}","${reason}","${p.name}","${p.location || ''}","${p.phone || ''}","${date}","${notes}"\n`
    }
    const link = document.createElement('a')
    link.href = encodeURI('data:text/csv;charset=utf-8,' + csv)
    link.download = `تقرير_رفض_المشاريع_${todayBaghdad()}.csv`
    document.body.appendChild(link); link.click(); document.body.removeChild(link)
  }

  const lastDate = grouped[0]?.lastDate ? new Date(grouped[0].lastDate).toLocaleDateString('ar-IQ') : '-'

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-xl font-bold text-red-600">🚫 أسباب رفض المشاريع</h2>
        <button onClick={exportCsv} className="px-4 py-2 rounded-full bg-green-700 text-white font-bold text-sm hover:brightness-110">📊 سحب التقرير (Excel)</button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <RejStat label="إجمالي المرفوضات" value={rejected.length} color="from-red-500 to-red-700" />
        <RejStat label="نسبة الرفض" value={`${Math.round((rejected.length / total) * 100)}%`} color="from-amber-400 to-orange-500" />
        <RejStat label="الفئات المختلفة" value={new Set(rejected.map(p => parseRejection(p.task).category)).size} color="from-gray-500 to-gray-700" />
        <RejStat label="آخر رفض" value={lastDate} color="from-cyan-500 to-cyan-700" small />
      </div>

      {/* Category filter */}
      <div className="bg-white rounded-xl shadow-sm border p-4 space-y-3">
        <div className="flex flex-wrap gap-2">
          {CATEGORIES.map(c => (
            <button key={c.key} onClick={() => setCategory(c.key)}
              className="px-4 py-1.5 rounded-full text-sm font-bold border-2 transition"
              style={category === c.key
                ? { background: c.color, color: '#fff', borderColor: c.color }
                : { borderColor: 'var(--bd-line)', color: 'var(--t-muted)' }}>
              {c.label}
            </button>
          ))}
        </div>
        <div className="grid sm:grid-cols-2 gap-3">
          <input className="inp" placeholder="ابحث في الأسباب أو المشاريع..." value={search} onChange={e => setSearch(e.target.value)} />
          <select className="inp" value={sort} onChange={e => setSort(e.target.value as typeof sort)}>
            <option value="count">الأكثر تكراراً</option>
            <option value="date">الأحدث</option>
            <option value="name">أبجدياً</option>
          </select>
        </div>
      </div>

      {/* Reasons list */}
      {grouped.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <div className="text-5xl mb-3">✅</div>
          <p className="text-lg">لا توجد مشاريع مرفوضة بهذه الفئة</p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {grouped.map(item => {
            const color = categoryColor(item.category)
            const open = expanded === item.reason
            return (
              <div key={item.reason} className="bg-white rounded-2xl shadow-sm p-5 cursor-pointer"
                style={{ borderRight: `6px solid ${color}` }}
                onClick={() => setExpanded(open ? null : item.reason)}>
                <div className="flex items-start justify-between">
                  <span className="text-xs px-2 py-1 rounded-full text-white font-bold" style={{ background: color }}>{item.category}</span>
                  <span className="text-xs px-3 py-1 rounded-full text-white font-bold" style={{ background: color }}>{item.projects.length} مشروع</span>
                </div>
                <h3 className="font-bold text-gray-800 mt-3">{item.reason}</h3>
                <div className="text-xs text-gray-500 mt-2 flex gap-3">
                  <span>📅 {item.lastDate ? new Date(item.lastDate).toLocaleDateString('ar-IQ') : 'غير معروف'}</span>
                  <span>📂 {item.projects.length} حالة</span>
                </div>
                {open && (
                  <div className="bg-gray-50 rounded-lg p-3 mt-3 space-y-2" style={{ borderRight: `3px solid ${color}` }}>
                    <div className="text-xs font-bold text-red-600">المشاريع المرفوضة:</div>
                    {item.projects.map(p => (
                      <div key={p.id} className="flex justify-between items-center border-b last:border-0 pb-1 text-sm">
                        <div>
                          <div className="font-bold">{p.name}</div>
                          <div className="text-xs text-gray-500">{p.location || 'لا يوجد موقع'}</div>
                        </div>
                        {canManage && (
                          <button type="button" onClick={(e) => { e.stopPropagation(); void restoreRejected(p, onRefresh) }}
                            className="rounded-lg bg-emerald-600 px-2.5 py-1 text-xs font-bold text-white hover:brightness-110">↩️ إرجاع للمشاريع</button>
                        )}
                        {canPurge && (
                          <button type="button" onClick={(e) => { e.stopPropagation(); void purgeRejected(p, onRefresh) }}
                            className="rounded-lg bg-red-600 px-2.5 py-1 text-xs font-bold text-white hover:brightness-110">🗑️ حذف نهائي</button>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
