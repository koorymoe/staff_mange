// ═══ ترقيم الصفحات ═══
//
// «هاي الفلترة هنانه؟ أريدك تسويلي مثلها للحجوزات».
//
// كان الترقيم مكتوب جوّا شاشة الزبائن وحدها. ونقله بالنسخ لشاشة
// الحجوزات يعني نسختين تفترقن أول تعديل — وحدة تعرض «…» والثانية لا،
// وحدة تبدي بـ١٠ والثانية بـ٢٥. فانفصل مكوّناً واحداً يخدم الاثنين.

function PageBtn({ children, onClick, disabled }: {
  children: React.ReactNode; onClick: () => void; disabled?: boolean
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="h-7 min-w-7 rounded-lg border border-slate-300 px-2 text-[11px] font-bold text-slate-600 disabled:opacity-40"
    >
      {children}
    </button>
  )
}

/** نافذة أرقام الصفحات — الصفر يعني «…»
 *  ⚠️ ٦٥ زر صفحة ما ينقرا ولا ينضغط بالموبايل، فنعرض جيران الحالية بس. */
function pageWindow(current: number, total: number): number[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const out: number[] = [1]
  const from = Math.max(2, current - 1)
  const to = Math.min(total - 1, current + 1)
  if (from > 2) out.push(0)
  for (let i = from; i <= to; i++) out.push(i)
  if (to < total - 1) out.push(0)
  out.push(total)
  return out
}

export default function Pager({
  page, perPage, total, unit = 'سجل', onPage, onPerPage, perPageOptions = [10, 25, 50],
}: {
  page: number
  perPage: number
  /** العدد الكلي — يجي من السيرفر لمن يكون الترقيم بالسيرفر */
  total: number
  unit?: string
  onPage: (p: number) => void
  onPerPage: (n: number) => void
  perPageOptions?: number[]
}) {
  const pageCount = Math.max(1, Math.ceil(total / perPage))
  const safePage = Math.min(Math.max(1, page), pageCount)

  return (
    <div className="flex flex-wrap items-center justify-between gap-2">
      <label className="flex items-center gap-1.5 text-[11px] text-slate-500">
        عرض
        <select
          value={perPage}
          onChange={(e) => { onPerPage(Number(e.target.value)); onPage(1) }}
          className="rounded-lg border border-slate-300 px-2 py-1 text-[11px]"
        >
          {perPageOptions.map((n) => <option key={n} value={n}>{n}</option>)}
        </select>
        من {total} {unit}
      </label>

      {pageCount > 1 && (
        <div className="flex flex-wrap items-center gap-1">
          <PageBtn onClick={() => onPage(1)} disabled={safePage === 1}>«</PageBtn>
          <PageBtn onClick={() => onPage(Math.max(1, safePage - 1))} disabled={safePage === 1}>‹</PageBtn>
          {pageWindow(safePage, pageCount).map((n, i) =>
            n === 0
              ? <span key={`gap${i}`} className="px-1 text-slate-400">…</span>
              : (
                <button
                  key={n}
                  onClick={() => onPage(n)}
                  className={`h-7 min-w-7 rounded-lg px-2 text-[11px] font-bold ${
                    n === safePage ? 'bg-[#2c5aad] text-white' : 'border border-slate-300 text-slate-600 hover:bg-slate-50'
                  }`}
                >
                  {n}
                </button>
              ))}
          <PageBtn onClick={() => onPage(Math.min(pageCount, safePage + 1))} disabled={safePage === pageCount}>›</PageBtn>
          <PageBtn onClick={() => onPage(pageCount)} disabled={safePage === pageCount}>»</PageBtn>
        </div>
      )}
    </div>
  )
}
