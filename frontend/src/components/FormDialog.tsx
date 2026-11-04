import { useState } from 'react'

// نافذة أسئلة من النظام — تنستعمل من utils/dialog (askForm).
export type Field =
  | { kind: 'text'; key: string; label: string; placeholder?: string; required?: boolean }
  | { kind: 'choice'; key: string; label: string; options: [string, string][]; required?: boolean }

export default function FormDialog({ title, fields, ok, onDone }: { title: string; fields: Field[]; ok: string; onDone: (v: Record<string, string> | null) => void }) {
  const [vals, setVals] = useState<Record<string, string>>({})
  const missing = fields.some((f) => f.required !== false && !(vals[f.key] ?? '').trim())
  return (
    <div dir="rtl" className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 p-4" onClick={() => onDone(null)}>
      <div className="max-h-[90vh] w-full max-w-md overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold text-[#0f2040]">{title}</h3>
        <div className="mt-3 space-y-4">
          {fields.map((f) => (
            <div key={f.key}>
              <p className="mb-1 text-sm font-bold text-slate-700">{f.label}</p>
              {f.kind === 'text' ? (
                <textarea autoFocus rows={2} value={vals[f.key] ?? ''} placeholder={f.placeholder}
                  onChange={(e) => setVals((v) => ({ ...v, [f.key]: e.target.value }))}
                  className="w-full rounded-lg border border-slate-300 p-2 text-sm" />
              ) : (
                <div className="flex flex-col gap-1.5">
                  {f.options.map(([k, label]) => (
                    <button key={k} type="button" onClick={() => setVals((v) => ({ ...v, [f.key]: k }))}
                      className={`rounded-lg border px-3 py-2 text-right text-sm font-bold transition ${vals[f.key] === k ? 'border-brand-600 bg-brand-600 text-white' : 'border-slate-200 bg-white text-slate-700 hover:bg-slate-50'}`}>
                      {label}
                    </button>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
        <div className="mt-5 flex gap-2">
          <button type="button" disabled={missing} onClick={() => onDone(vals)}
            className="flex-1 rounded-lg bg-brand-700 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">{ok}</button>
          <button type="button" onClick={() => onDone(null)} className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">إلغاء</button>
        </div>
      </div>
    </div>
  )
}
