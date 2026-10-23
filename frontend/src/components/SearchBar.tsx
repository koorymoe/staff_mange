/**
 * شريط البحث — شكل «إدارة الكوادر» (Employees.tsx).
 *
 * ⚠️ الترشيح نفسه لازم يمرّ بـ`matches`/`makeFilter` من utils/search.ts
 * مو بمقارنة نصية عادية: هذاك يعالج الهمزة والتاء المربوطة والألف
 * المقصورة والأرقام الهندية. البحث بـ«احمد» لازم يلگي «أحمد».
 */
export default function SearchBar({
  value, onChange, placeholder = 'بحث...', children,
}: {
  value: string
  onChange: (v: string) => void
  placeholder?: string
  /** مرشّحات تنعرض جنب البحث. */
  children?: React.ReactNode
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div
        className="flex flex-1 items-center gap-2 rounded-xl border px-4 py-2.5"
        style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          strokeWidth="2" className="shrink-0" style={{ color: 'var(--t-faint)' }}>
          <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="w-full bg-transparent text-sm outline-none"
          style={{ color: 'var(--t-body)' }}
        />
      </div>
      {children}
    </div>
  )
}
