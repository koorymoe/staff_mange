import { useState } from 'react'
import { Link } from 'react-router-dom'

// ═══ 📖 «شنو مطلوب مني؟» — تعليمات كادر الديكور (دازاي 10-25) ═══
// جمل قصيرة بالعراقي وأيقونات كبيرة، لأن أغلبهم ما مكملين مدارس.
// ⚠️ المحتوى مسودة لوفي — دازاي يراجعه ويعدّله (هو أدرى بالشغل).

const DAY: { icon: string; text: string }[] = [
  { icon: '🕗', text: 'أول ما توصل الشغل أو تبدي يومك: سجّل حضورك.' },
  { icon: '🪚', text: 'افتح «شغلي اليوم» وشوف وين رايح.' },
  { icon: '🚗', text: 'من تطلع: اضغط «طلعت».' },
  { icon: '📍', text: 'من توصل البيت: اضغط «وصلت».' },
  { icon: '📸', text: 'صوّر المكان قبل ما تلمس شي.' },
  { icon: '▶️', text: 'اضغط «بديت» واشتغل.' },
  { icon: '📸', text: 'من تخلص: صوّر الشغل زين، من قريب ومن بعيد.' },
  { icon: '✅', text: 'اضغط «خلصت».' },
]

const RULES: { icon: string; text: string }[] = [
  { icon: '🙂', text: 'سلّم على صاحب البيت، وتكلم بهدوء واحترام.' },
  { icon: '🧹', text: 'نظّف مكانك قبل ما تطلع. لا تترك زبالة.' },
  { icon: '🛡️', text: 'غطّي الأثاث والأرض قبل الشغل.' },
  { icon: '🚭', text: 'لا تدخن داخل البيت.' },
  { icon: '💵', text: 'لا تاخذ فلوس من الزبون إلا إذا مسؤولك كلك.' },
  { icon: '📞', text: 'إذا صارت مشكلة أو نقصت مادة: خابر مسؤولك فوراً، لا تتصرف لحالك.' },
]

const TRADES: { key: string; icon: string; name: string; before: string[]; photo: string[]; never: string[] }[] = [
  { key: 'paint', icon: '🎨', name: 'الصبغ',
    before: ['تأكد من اللون ويا صاحب البيت قبل ما تفتح السطل.', 'غطّي الأرض والأثاث والسويچات بالنايلون والشريط.', 'صلّح الشقوق والحفر بالمعجون قبل الصبغ.'],
    photo: ['الحايط قبل.', 'الحايط بعد، والزوايا والسقف.'],
    never: ['لا تصبغ على حايط رطب أو بيه نزيز — خابر مسؤولك.', 'لا تخلط ألوان بدون موافقة.'] },
  { key: 'gypsum', icon: '🧱', name: 'الجبس والبورد',
    before: ['تأكد من القياسات والرسمة ويا مسؤولك.', 'شوف وين الأسلاك والإنارة قبل ما تثبّت.', 'تأكد الهيكل (الحديد) مستقيم بالميزان.'],
    photo: ['السقف أو الحايط قبل.', 'الهيكل قبل ما تسدّه.', 'الشغل بعد ما يخلص.'],
    never: ['لا تسد البورد قبل ما الكهربائي يخلص أسلاكه.', 'لا تقص سلك أو بوري.'] },
  { key: 'carpentry', icon: '🪵', name: 'النجارة',
    before: ['تأكد من القياس مرتين قبل ما تقص.', 'تأكد من نوع الخشب واللون ويا الطلب.', 'احمِ الأرض من النشارة.'],
    photo: ['المكان قبل.', 'القطعة بعد التركيب، وهي مفتوحة ومسدودة.'],
    never: ['لا تشغّل المنشار بدون نظارة وكفوف.', 'لا تخلي نشارة بالبيت.'] },
  { key: 'tiles', icon: '🔲', name: 'البلاط والسيراميك',
    before: ['تأكد من نوع البلاط واللون ويا صاحب البيت.', 'شوف الأرض مستوية بالميزان قبل ما تبدي.', 'اعرف اتجاه الميلان للمجاري بالحمام والمطبخ.'],
    photo: ['الأرض أو الحايط قبل.', 'البلاط بعد، والفواصل من قريب.'],
    never: ['لا تبلّط على أرض مو مستوية — خابر مسؤولك.', 'لا تسد المجاري بالسمنت.'] },
]

export default function DecorGuidePage() {
  const [open, setOpen] = useState<string | null>(null)
  return (
    <div dir="rtl" className="mx-auto max-w-lg space-y-4 pb-24">
      <div className="rounded-3xl bg-gradient-to-l from-amber-600 to-orange-500 p-5 text-white shadow">
        <h1 className="text-3xl font-black">📖 شنو مطلوب مني؟</h1>
        <p className="mt-1 text-lg">اقراها مرة، وارجعلها وكت ما تحتاج.</p>
      </div>

      <Section title="📅 كل يوم، بهالترتيب">
        <ol className="space-y-2">
          {DAY.map((s, i) => (
            <li key={i} className="flex items-center gap-3 rounded-2xl bg-slate-50 p-3">
              <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[#0f2040] text-lg font-black text-white">{i + 1}</span>
              <span className="text-3xl">{s.icon}</span>
              <span className="text-lg font-bold text-slate-800">{s.text}</span>
            </li>
          ))}
        </ol>
      </Section>

      <Section title="🤝 بكل بيت">
        <ul className="space-y-2">
          {RULES.map((r, i) => (
            <li key={i} className="flex items-center gap-3 text-lg font-bold text-slate-800"><span className="text-3xl">{r.icon}</span>{r.text}</li>
          ))}
        </ul>
      </Section>

      <Section title="🛠️ شغلك أنت">
        <div className="space-y-2">
          {TRADES.map((t) => (
            <div key={t.key} className="overflow-hidden rounded-2xl ring-1 ring-slate-200">
              <button type="button" onClick={() => setOpen(open === t.key ? null : t.key)}
                className="flex w-full items-center gap-3 bg-white p-4 text-right text-xl font-extrabold text-[#0f2040]">
                <span className="text-3xl">{t.icon}</span><span className="flex-1">{t.name}</span><span>{open === t.key ? '▲' : '▼'}</span>
              </button>
              {open === t.key && (
                <div className="space-y-3 bg-slate-50 p-4 text-lg">
                  <List title="✔️ قبل ما تبدي" items={t.before} />
                  <List title="📸 صوّر" items={t.photo} />
                  <List title="⛔ ممنوع" items={t.never} danger />
                </div>
              )}
            </div>
          ))}
        </div>
      </Section>

      <Link to="/decor-today" className="block rounded-2xl bg-[#0f2040] p-4 text-center text-xl font-extrabold text-white shadow">🪚 رجوع لشغلي اليوم</Link>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-3xl bg-white p-4 shadow ring-1 ring-slate-200">
      <h2 className="mb-3 text-2xl font-black text-[#0f2040]">{title}</h2>
      {children}
    </section>
  )
}

function List({ title, items, danger }: { title: string; items: string[]; danger?: boolean }) {
  return (
    <div>
      <p className={`font-black ${danger ? 'text-red-700' : 'text-slate-700'}`}>{title}</p>
      <ul className="mt-1 space-y-1">
        {items.map((x, i) => <li key={i} className={danger ? 'text-red-800' : 'text-slate-800'}>• {x}</li>)}
      </ul>
    </div>
  )
}
