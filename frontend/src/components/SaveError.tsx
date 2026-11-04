// ═══ شريط خطأ الحفظ ═══
//
// لاصق بأعلى الشاشة (sticky) بقصد: الفشل يصير غالباً وأنت بنص قائمة
// طويلة، ورسالة تطلع فوگ برّا مجال النظر ما تنقرا — يعني نرجع لنفس
// الفشل الصامت الي نحاول نعالجه.
//
// وتگول صراحةً «التغيير ما انحفظ»: بدونها الموظف يشوف رسالة حمرا
// ويحتار — هل انحفظ وطلع تحذير؟ ولا ما انحفظ؟
export default function SaveError({ message, onClose }: {
  message: string | null
  onClose: () => void
}) {
  if (!message) return null
  return (
    <div
      dir="rtl"
      className="sticky top-2 z-30 mb-3 flex items-start justify-between gap-3 rounded-xl border-2 border-red-300 bg-red-50 px-4 py-3 shadow-lg"
    >
      <div className="min-w-0">
        <p className="text-sm font-extrabold text-red-800">⚠️ {message}</p>
        <p className="mt-0.5 text-xs text-red-700">
          التغيير <b>ما انحفظ</b> — جرّب مرة ثانية. إذا تكرر، دز صورة الرسالة للدعم.
        </p>
      </div>
      <button
        onClick={onClose}
        className="shrink-0 rounded-lg border border-red-300 bg-white px-2.5 py-1 text-xs font-bold text-red-700"
      >
        إخفاء
      </button>
    </div>
  )
}
