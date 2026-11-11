import type { Discipline } from '../apiTypes'

// (ل) 11-11 دازاي: «التأخر بالعمل بدون تبرير بدون عذر هاي مشكلة» — تفصيل الانضباط بالدوام بأيامه.
const ST: Record<string, { label: string; cls: string }> = {
  LATE: { label: 'متأخر بلا عذر', cls: 'bg-red-100 text-red-700' },
  LATE_EXCUSED: { label: 'متأخر بعذر مقبول', cls: 'bg-emerald-100 text-emerald-700' },
  ABSENT: { label: 'غياب بلا إجازة', cls: 'bg-rose-200 text-rose-800' },
}

export default function DisciplineBox({ d, title = '⏰ الانضباط بالدوام (١٥٪ من النهائي)' }: { d: Discipline; title?: string }) {
  return (
    <div className="rounded-xl border border-rose-200 bg-rose-50/40 p-2">
      <p className="mb-1 flex items-center justify-between text-xs font-extrabold text-rose-900">
        <span>{title}</span>
        <span className="rounded-full bg-white px-2 py-0.5 text-sm">{d.score ?? '—'}</span>
      </p>
      <p className="text-xs text-slate-600">
        {d.days} يوم دوام · بوقته {d.onTime} · <b className="text-red-700">بلا عذر {d.late}</b> ({d.lateMinutes} دقيقة) · بعذر {d.lateExcused} · <b className="text-rose-800">غياب {d.absent}</b>
      </p>
      <p className="text-[11px] text-slate-500">الدرجة = ١٠٠ − ٢ لكل تأخير بلا عذر − ٥ لكل غياب. السماح ١٠ دقايق، والعذر المقبول والإجازة ما يضرّون.</p>
      {d.items.length > 0 && (
        <ul className="mt-1 space-y-0.5 text-xs">
          {d.items.map((x) => (
            <li key={x.date} className="flex flex-wrap items-center gap-2">
              <span className="tabular-nums text-slate-600">{x.date}</span>
              <span className={`rounded-full px-2 py-0.5 font-bold ${ST[x.status].cls}`}>{ST[x.status].label}</span>
              {x.lateMinutes ? <span className="text-slate-600">حضر {x.checkIn} والدوام {x.shiftStart} — {x.lateMinutes} دقيقة</span> : null}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
