import { useEffect, useState } from 'react'
import { api, parseQuotationSnapshot, type Quotation, type QuotationVersion } from '../api'

// ═══ 🕘 هيستوري عرض السعر (قرار (ع) 10-09) ═══
// العرض المعدّل هو الي يظهر، ومن هالزر يختار أي نسخة يعرضها ويطبعها.
// ⚠️ للعرض والطباعة بس — ما ترجّع النسخة القديمة بدل الحالية.
export default function QuotationHistoryPicker({ quotationId, selected, onPick }: {
  quotationId: string
  selected: number | null // null = الحالية
  onPick: (snap: Quotation | null, version: number | null) => void
}) {
  const [rows, setRows] = useState<QuotationVersion[]>([])
  const [open, setOpen] = useState(false)

  useEffect(() => {
    let alive = true
    api.getQuotationVersions(quotationId).then((r) => { if (alive) setRows(r ?? []) }).catch(() => {})
    return () => { alive = false }
  }, [quotationId])

  if (rows.length === 0) return null
  const fmt = (n: number) => n.toLocaleString('en-US', { maximumFractionDigits: 0 })
  return (
    <span style={{ position: 'relative', display: 'inline-block' }}>
      <button type="button" onClick={() => setOpen((o) => !o)} style={{
        background: selected == null ? '#455a64' : '#6a1b9a', color: 'white', border: 'none', padding: '10px 18px',
        borderRadius: '8px', cursor: 'pointer', fontWeight: 700, fontFamily: 'inherit',
      }}>🕘 {selected == null ? `النسخ (${rows.length + 1})` : `نسخة ${selected}`}</button>
      {open && (
        <div dir="rtl" style={{
          position: 'absolute', top: '110%', right: 0, zIndex: 3100, minWidth: '260px', background: 'var(--sf-card, #fff)',
          border: '1px solid #cfd8dc', borderRadius: '10px', boxShadow: '0 8px 24px rgba(0,0,0,.18)', padding: '6px',
        }}>
          <button type="button" onClick={() => { setOpen(false); onPick(null, null) }} style={rowStyle(selected == null)}>
            <b>✅ النسخة الحالية</b>
          </button>
          {rows.map((v) => {
            const snap = parseQuotationSnapshot(v.snapshot)
            return (
              <button key={v.id} type="button" disabled={!snap} onClick={() => { setOpen(false); onPick(snap, v.version) }} style={rowStyle(selected === v.version)}>
                <b>نسخة {v.version}</b>
                <span style={{ fontSize: '11px', opacity: 0.75 }}>
                  {snap ? `${fmt(snap.netTotal)} د.ع · ` : ''}{new Date(v.archivedAt).toLocaleDateString('ar-IQ')}{v.archivedByName ? ` · ${v.archivedByName}` : ''}
                </span>
              </button>
            )
          })}
        </div>
      )}
    </span>
  )
}

const rowStyle = (active: boolean): React.CSSProperties => ({
  display: 'flex', flexDirection: 'column', alignItems: 'flex-start', gap: '2px', width: '100%', textAlign: 'right',
  padding: '8px 10px', borderRadius: '8px', border: 'none', cursor: 'pointer', fontFamily: 'inherit',
  background: active ? '#ede7f6' : 'transparent', color: '#263238', fontSize: '13px',
})
