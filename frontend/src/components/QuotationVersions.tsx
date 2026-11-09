import { useEffect, useState } from 'react'
import { api, parseQuotationSnapshot, type QuotationVersion, type Quotation } from '../api'

// ═══ النسخ القديمة لعرض السعر ═══
//
// (ع): «أكو عروض أسعار نسويهن وبعد فترة نرجع نريد نعدّل عليهن —
// أريد ينحفظ بدل القديم بس القديم يضل مؤرشف بغير مكان».
//
// ⚠️ **للقراءة وبس، وما ينرجّع**: النسخة القديمة وثيقة انرسلت للزبون
// بتاريخ معيّن. زر «رجّعني لهاي النسخة» يخلي الأرشيف مصدر تعديل، وأول
// غلطة ضغطة يمحي شهر شغل. اللي يريد يرجع يقرا الأرقام ويكتبها هو.
//
// ⚠️ ويُجلب **عند الفتح وبس**: أغلب العروض ما إلها نسخ، فتحميلها مع
// كل صف بالقائمة يسوي نداء لكل عرض بلا فائدة.

const fmt = (n: number) => n.toLocaleString('en-US', { maximumFractionDigits: 2 })

export default function QuotationVersions({ quotationId }: { quotationId: string }) {
  const [rows, setRows] = useState<QuotationVersion[] | null>(null)
  const [failed, setFailed] = useState(false)
  const [openId, setOpenId] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getQuotationVersions(quotationId)
      .then((r) => { if (alive) setRows(r) })
      .catch(() => { if (alive) setFailed(true) })
    return () => { alive = false }
  }, [quotationId])

  if (failed) return <p className="text-xs text-red-600">تعذّر جلب النسخ القديمة</p>
  if (rows === null) return <p className="text-xs text-slate-400">جاري التحميل…</p>
  if (rows.length === 0) {
    return (
      <p className="text-xs text-slate-500">
        ماكو نسخ قديمة — هذا العرض ما انتعدّل بعد.
      </p>
    )
  }

  return (
    <div dir="rtl" className="space-y-2">
      <p className="text-xs font-bold text-slate-600">
        📚 {rows.length} نسخة مؤرشفة — كل تعديل يحفظ الي كان قبله
      </p>
      {rows.map((v) => {
        const snap: Quotation | null = parseQuotationSnapshot(v.snapshot)
        const open = openId === v.id
        return (
          <div key={v.id} className="rounded-lg border border-slate-200 bg-white">
            <button
              type="button"
              onClick={() => setOpenId(open ? null : v.id)}
              className="flex w-full flex-wrap items-center justify-between gap-2 px-3 py-2 text-right"
            >
              <span className="text-xs font-extrabold text-[#0f2040]">
                نسخة {v.version}
                {snap && <span className="ms-2 font-bold text-slate-500">الصافي وقتها: {fmt(snap.netTotal)} د.ع</span>}
              </span>
              <span className="text-[11px] font-bold text-slate-500">
                {/* ⚠️ «—» لمّا الموظف ما انسجّل أو انمسح: ما ننسب
                    التعديل لواحد غلط. */}
                {v.archivedByName || '—'} · {new Date(v.archivedAt).toLocaleString('ar-IQ')}
                <span className="ms-2 text-brand-600">{open ? '▲' : '▼'}</span>
              </span>
            </button>
            {open && (
              <div className="border-t border-slate-100 px-3 py-2">
                {!snap ? (
                  <p className="text-xs text-red-600">تعذّر قراءة هاي النسخة</p>
                ) : (
                  <>
                    <p className="mb-1 text-[11px] font-bold text-slate-500">
                      المجموع {fmt(snap.grandTotal)} · خصم {snap.discountPercent}% ({fmt(snap.discountValue)}) · الصافي {fmt(snap.netTotal)}
                    </p>
                    <table className="w-full text-[11px]">
                      <thead>
                        <tr className="text-slate-500">
                          <th className="py-1 text-right font-bold">م</th>
                          <th className="py-1 text-right font-bold">البيان</th>
                          <th className="py-1 text-center font-bold">العدد</th>
                          <th className="py-1 text-center font-bold">السعر</th>
                          <th className="py-1 text-center font-bold">الإجمالي</th>
                        </tr>
                      </thead>
                      <tbody>
                        {/* الترتيب مثل ما كان بذاك اليوم — اللقطة محفوظة
                            مرتّبة، فما نعيد فرزها. */}
                        {snap.items.map((it, i) => (
                          <tr key={i} className="border-t border-slate-100">
                            <td className="py-1 text-slate-400">{i + 1}</td>
                            <td className="py-1 font-bold text-[#0f2040]">{it.productName}</td>
                            <td className="py-1 text-center">{it.quantity}</td>
                            <td className="py-1 text-center">{fmt(it.unitPrice)}</td>
                            <td className="py-1 text-center font-bold">{fmt(it.totalPrice)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </>
                )}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
