// ═══ واجهة إعدادات الجهاز ═══
//
// ⚠️ **الحقول مسوّدة حتى تضغط «تطبيق»** — مثل الجهاز الحقيقي بالضبط.
// الكتابة المباشرة على الحالة تخلّي المحاكاة تعيد الحساب مع كل حرف،
// وتخفي أهم درس بالواجهة: **الإعداد ما ينفّذ إلا بالتطبيق**. والفني
// الي ينسى يضغط يظن الجهاز خربان — وهذا يصير بالميدان يومياً.
//
// ⚠️ والشكل مقصود إنه يشبه صفحة راوتر: شريط علوي بالموديل والعنوان،
// تبويبات أفقية، أقسام بعناوين، وزرّان بالأسفل. المتدرّب الي يتعوّد
// على هالتخطيط يلگا نفسه بأي واجهة راوتر حقيقية.

import { useMemo, useState } from 'react'
import { applyDraft, draftFrom, fieldVisible, type PanelDraft, type PanelField, type PanelSchema } from './schema'

interface Props {
  schema: PanelSchema
  /** خصائص الجهاز الحالية — المصدر والهدف. */
  params: Record<string, unknown>
  /** القيم المحسوبة للحقول القرائية (قدرة الاستقبال، حالة التسجيل…). */
  computed?: Record<string, { text: string; tone?: 'ok' | 'warn' | 'bad' }>
  onApply: (next: Record<string, unknown>) => void
  onClose: () => void
}

export default function PanelUI({ schema, params, computed, onApply, onClose }: Props) {
  const [tab, setTab] = useState(schema.tabs[0]?.id ?? '')
  const initial = useMemo(() => draftFrom(schema, params), [schema, params])
  const [draft, setDraft] = useState<PanelDraft>(initial)
  const [saved, setSaved] = useState(false)

  // ⚠️ المقارنة على المسوّدة مقابل الابتدائية: هذا الي يخلّي زر
  // «تطبيق» ينوّر لمن اكو تغيير فعلاً — ومنه يعرف المتدرّب إنه لسه
  // ما طبّق.
  const dirty = useMemo(
    () => Object.keys(draft).some((k) => String(draft[k] ?? '') !== String(initial[k] ?? '')),
    [draft, initial],
  )

  const set = (k: string, v: string | number | boolean) => {
    setDraft((d) => ({ ...d, [k]: v }))
    setSaved(false)
  }

  const apply = () => {
    onApply(applyDraft(params, draft))
    setSaved(true)
  }

  const active = schema.tabs.find((t) => t.id === tab) ?? schema.tabs[0]

  return (
    <div dir="rtl" className="overflow-hidden rounded-xl bg-[#0f1420] ring-1 ring-slate-700">
      {/* ═══ الشريط العلوي ═══ */}
      <div className="flex flex-wrap items-center justify-between gap-2 bg-gradient-to-l from-[#132038] to-[#0f1726] px-4 py-2.5">
        <button onClick={onClose} className="text-[11px] font-bold text-slate-400 hover:text-slate-200">✕ سكّر</button>
        <div className="text-left">
          <p className="text-[13px] font-bold text-slate-100">{schema.name}</p>
          <p className="font-mono text-[10px] text-slate-500" dir="ltr">{schema.address}</p>
        </div>
      </div>

      {schema.warn && (
        <p className="bg-amber-500/10 px-4 py-1.5 text-[10.5px] text-amber-200">⚠️ {schema.warn}</p>
      )}

      {/* ═══ التبويبات ═══ */}
      <div className="flex flex-wrap gap-0.5 border-b border-slate-800 bg-[#0b1120] px-2 pt-1.5">
        {schema.tabs.map((t) => (
          <button key={t.id} onClick={() => setTab(t.id)}
            className={`rounded-t-lg px-3.5 py-1.5 text-[11.5px] font-bold transition ${
              t.id === active?.id ? 'bg-[#0f1420] text-sky-300' : 'text-slate-500 hover:text-slate-300'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {/* ═══ المحتوى ═══ */}
      <div className="max-h-[340px] space-y-4 overflow-y-auto p-4">
        {active?.sections.map((sec, si) => (
          <div key={si}>
            <h4 className="mb-1 text-[12px] font-bold text-slate-200">{sec.title}</h4>
            {sec.note && <p className="mb-2 text-[10.5px] leading-relaxed text-slate-500">{sec.note}</p>}
            <div className="grid gap-2.5 sm:grid-cols-2">
              {sec.fields.filter((f) => fieldVisible(f, draft)).map((f) => (
                <FieldRow key={f.id} f={f} draft={draft} computed={computed} onSet={set} />
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* ═══ الأزرار ═══ */}
      <div className="flex flex-wrap items-center gap-2 border-t border-slate-800 bg-[#0b1120] px-4 py-2.5">
        <button
          onClick={apply}
          disabled={!dirty}
          className={`rounded-lg px-4 py-1.5 text-[12px] font-bold transition ${
            dirty ? 'bg-sky-600 text-white hover:bg-sky-500' : 'bg-slate-800 text-slate-500'}`}
        >
          تطبيق
        </button>
        <button
          onClick={() => { setDraft(initial); setSaved(false) }}
          disabled={!dirty}
          className="rounded-lg bg-slate-800 px-3 py-1.5 text-[12px] font-bold text-slate-300 disabled:opacity-40"
        >
          إلغاء
        </button>

        {/* ⚠️ التنبيه الي يعلّم الدرس: تعديل بلا تطبيق = ما صار شي. */}
        {dirty && (
          <span className="text-[11px] font-bold text-amber-300">
            ⚠️ عدّلت ولا طبّقت — الجهاز لسه على إعداداته القديمة.
          </span>
        )}
        {saved && !dirty && <span className="text-[11px] font-bold text-emerald-400">✅ انطبّقت الإعدادات</span>}
      </div>
    </div>
  )
}

function FieldRow({ f, draft, computed, onSet }: {
  f: PanelField
  draft: PanelDraft
  computed?: Record<string, { text: string; tone?: 'ok' | 'warn' | 'bad' }>
  onSet: (k: string, v: string | number | boolean) => void
}) {
  const label = (
    <span className="mb-1 flex items-baseline gap-1 text-[11px] font-bold text-slate-400">
      {f.label}{f.unit && <span className="text-slate-600">({f.unit})</span>}
    </span>
  )

  // ═══ حقل قرائي محسوب ═══
  //
  // ⚠️ يقرا من نتيجة المحرّك مو من قيمة مخزونة: قدرة الاستقبال
  // **تُقاس** — لو انخزنت كخاصية، الفني يگدر «يعدّلها» وهذا مستحيل
  // بالجهاز الحقيقي.
  if (f.kind === 'readonly') {
    const c = f.computed ? computed?.[f.computed] : undefined
    return (
      <label className="block">
        {label}
        <div className={`rounded-lg bg-[#0b1120] px-2.5 py-1.5 font-mono text-[12.5px] font-bold ring-1 ring-slate-800 ${
          c?.tone === 'bad' ? 'text-red-400' : c?.tone === 'warn' ? 'text-amber-300'
            : c ? 'text-emerald-400' : 'text-slate-600'}`}>
          {c?.text ?? '—'}
        </div>
        {f.help && <span className="mt-0.5 block text-[10px] leading-relaxed text-slate-600">{f.help}</span>}
      </label>
    )
  }

  const key = f.path ?? f.id
  const val = draft[key]

  return (
    <label className="block">
      {label}
      {f.kind === 'bool' ? (
        <button
          type="button"
          onClick={() => onSet(key, !(val === true || val === 'true'))}
          className={`w-full rounded-lg px-3 py-1.5 text-[12px] font-bold transition ${
            val === true || val === 'true' ? 'bg-emerald-600 text-white' : 'bg-slate-800 text-slate-400'}`}
        >
          {val === true || val === 'true' ? 'مفعّل ✓' : 'معطّل'}
        </button>
      ) : f.kind === 'select' ? (
        <select
          value={String(val ?? '')} onChange={(e) => onSet(key, e.target.value)}
          className="w-full rounded-lg border border-slate-700 bg-[#0b1120] px-2 py-1.5 text-[12px] text-slate-200"
        >
          {f.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      ) : (
        <input
          type={f.kind === 'password' ? 'password' : f.kind === 'number' ? 'number' : 'text'}
          value={String(val ?? '')}
          placeholder={f.placeholder}
          dir={f.kind === 'number' || /ip|mask|ssid|user/i.test(key) ? 'ltr' : undefined}
          onChange={(e) => onSet(key, f.kind === 'number' ? Number(e.target.value) : e.target.value)}
          className="w-full rounded-lg border border-slate-700 bg-[#0b1120] px-2 py-1.5 text-[12px] text-slate-200 placeholder:text-slate-600"
        />
      )}
      {f.help && <span className="mt-0.5 block text-[10px] leading-relaxed text-slate-600">{f.help}</span>}
    </label>
  )
}
