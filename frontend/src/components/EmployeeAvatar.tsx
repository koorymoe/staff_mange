import { useRef, useState } from 'react'
import { api, fileUrl } from '../api'

// ═══ صورة الموظف ═══
//
// «هنا ضيفلي خاصية أكدر أضيف صورة بدل الحرف — أضيف صورة، ومن أضغط
// عليها تنفتح. أريد أضيف صور للموظفين».
//
// الحرف الأول ينفع كبديل، بس ما يعرّف: ثلاث موظفين اسمهم يبدي بـ«ع»
// ياخذون نفس الحرف ونفس اللون.
//
// ⚠️ مكوّن واحد لكل الشاشات: البطاقة الجانبية، قوائم الموظفين،
// الكادر بالحجز. لو كل شاشة رسمت الصورة بطريقتها، تصير وحدة دائرية
// ووحدة مربّعة ووحدة ما تفتح — ونفس الموظف يبان بثلاث أشكال.

/** لون ثابت للاسم — نفس الاسم يعطي نفس اللون كل مرة (بديل الصورة). */
const COLORS = [
  'bg-violet-500', 'bg-sky-500', 'bg-emerald-500', 'bg-amber-500',
  'bg-rose-500', 'bg-indigo-500', 'bg-teal-500', 'bg-orange-500',
]
function colorOf(seed: string) {
  let h = 0
  for (let i = 0; i < seed.length; i++) h = (h * 31 + seed.charCodeAt(i)) >>> 0
  return COLORS[h % COLORS.length]
}

const SIZES = {
  sm: 'h-8 w-8 text-sm',
  md: 'h-11 w-11 text-lg',
  lg: 'h-14 w-14 text-2xl',
  xl: 'h-20 w-20 text-3xl',
}

export default function EmployeeAvatar({
  name, photoUrl, size = 'md', rounded = 'full', canEdit = false, onPhotoChange, className = '',
}: {
  name: string
  photoUrl?: string | null
  size?: keyof typeof SIZES
  /** الشكل: دائرة (الافتراضي) أو مربّع بحواف — حسب مكان العرض */
  rounded?: 'full' | 'xl'
  /** يطلع زر الرفع — للمخوّل بس */
  canEdit?: boolean
  onPhotoChange?: (url: string | null) => void | Promise<void>
  className?: string
}) {
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const src = photoUrl ? fileUrl(photoUrl) : ''
  const shape = rounded === 'full' ? 'rounded-full' : 'rounded-2xl'
  const letter = (name || '؟').trim().charAt(0)

  const pick = async (file: File | undefined) => {
    if (!file) return
    // ⚠️ الفحص هنا **زيادة** على فحص السيرفر مو بديل عنه: هذا يوفّر
    // على الموظف رفعة كاملة تنرفض بالآخر، والسيرفر هو الي يحمي فعلاً.
    if (!file.type.startsWith('image/')) { setError('لازم تكون صورة'); return }
    setBusy(true)
    setError(null)
    try {
      const up = await api.uploadFile(file, 'employees')
      await onPhotoChange?.(up.url)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر رفع الصورة')
    } finally {
      setBusy(false)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  return (
    <>
      <div className={`relative inline-block ${className}`}>
        <button
          type="button"
          // الضغط يفتح الصورة بس لمن تكون موجودة — الحرف ما إله شنو
          // ينفتح، وزر يضغط وما يصير شي يخلّي الواحد يظن الشاشة علّقت.
          onClick={() => src && setOpen(true)}
          className={`flex ${SIZES[size]} ${shape} items-center justify-center overflow-hidden font-black text-white shadow-sm transition ${
            src ? 'cursor-zoom-in hover:opacity-90' : `${colorOf(name || '?')} cursor-default`
          }`}
          title={src ? `صورة ${name} — اضغط للتكبير` : name}
        >
          {src ? <img src={src} alt={name} className="h-full w-full object-cover" /> : letter}
        </button>

        {canEdit && (
          <>
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              disabled={busy}
              className="absolute -bottom-1 -left-1 flex h-6 w-6 items-center justify-center rounded-full border-2 border-white bg-brand-600 text-[10px] text-white shadow hover:bg-brand-700 disabled:opacity-50"
              title={src ? 'تبديل الصورة' : 'إضافة صورة'}
            >
              {busy ? '…' : '📷'}
            </button>
            <input
              ref={inputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => pick(e.target.files?.[0])}
            />
          </>
        )}
      </div>

      {error && <p className="mt-1 text-[10px] font-bold text-red-600">{error}</p>}

      {/* ═══ العارض ═══
          ⚠️ الضغط برّا الصورة يسدّه، والصورة نفسها ما تسدّه: الواحد
          يضغط على الصورة حتى يشوفها أوضح، وسدّها بالغلط يعصّب. */}
      {open && src && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 p-4"
          onClick={() => setOpen(false)}
        >
          <div className="max-h-full max-w-lg" onClick={(e) => e.stopPropagation()}>
            <div className="mb-2 flex items-center justify-between gap-3">
              <span className="rounded-lg bg-white/10 px-3 py-1 text-sm font-bold text-white">{name}</span>
              <div className="flex gap-2">
                {canEdit && (
                  <button
                    onClick={() => { void onPhotoChange?.(null); setOpen(false) }}
                    className="rounded-lg bg-red-500/90 px-3 py-1 text-xs font-bold text-white hover:bg-red-600"
                  >
                    🗑️ شيل الصورة
                  </button>
                )}
                <button
                  onClick={() => setOpen(false)}
                  className="rounded-lg bg-white/15 px-3 py-1 text-xs font-bold text-white hover:bg-white/25"
                >
                  ✕ إغلاق
                </button>
              </div>
            </div>
            <img src={src} alt={name} className="max-h-[75vh] w-auto rounded-2xl object-contain shadow-2xl" />
          </div>
        </div>
      )}
    </>
  )
}
