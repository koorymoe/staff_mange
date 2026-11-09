import { useEffect, useState } from 'react'
import { api, type CustomerSignature } from '../api'

// (ع) 11-07: توقيع الزبون عند التسليم — يطلع بمراجعة التقرير إذا انوقّع
export default function CustomerSignatureView({ bookingId }: { bookingId: string }) {
  const [s, setS] = useState<CustomerSignature | null | undefined>(undefined)
  useEffect(() => {
    let alive = true
    api.getCustomerSignature(bookingId).then((r) => { if (alive) setS(r) }).catch(() => { if (alive) setS(null) })
    return () => { alive = false }
  }, [bookingId])
  if (s === undefined) return null
  if (!s) return <p className="text-xs text-slate-400">✍️ ماكو توقيع زبون على هذا الحجز</p>
  return (
    <div className="rounded-xl border border-emerald-200 bg-emerald-50/60 p-2">
      <p className="text-xs font-bold text-emerald-800">
        ✍️ وقّع الزبون: {s.signerName} · {new Date(s.signedAt).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}
        {s.collectedByName && <span className="font-normal text-slate-500"> · أخذه {s.collectedByName}</span>}
      </p>
      <img src={s.image} alt="توقيع الزبون" className="mt-1 h-24 rounded-lg bg-white" />
    </div>
  )
}
