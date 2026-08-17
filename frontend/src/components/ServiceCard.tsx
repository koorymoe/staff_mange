import type { ServiceItem } from '../utils/loginServices'

// بطاقة خدمة بشاشة الدخول — بملف لحالها حتى يضل التحديث السريع
// (Fast Refresh) شغّالاً: الملف الي يصدّر مكوّناً وثوابت وياه يكسره.

export default function ServiceCard({ item, side }: { item: ServiceItem; side: 'right' | 'left' }) {
  return (
    <div className={`group flex items-center gap-3 ${side === 'right' ? 'flex-row' : 'flex-row-reverse'}`}>
      {/* الأيقونة بقاعدة متوهّجة — نفس فكرة المنصّات بالتصميم */}
      <div className="relative shrink-0">
        <div
          className="absolute inset-0 rounded-2xl blur-xl transition-opacity duration-500 group-hover:opacity-90"
          style={{ backgroundColor: item.color, opacity: 0.35 }}
        />
        <div
          className="relative flex h-14 w-14 items-center justify-center rounded-2xl border text-2xl backdrop-blur-sm"
          style={{ borderColor: `${item.color}66`, backgroundColor: `${item.color}1a` }}
        >
          {item.icon}
        </div>
      </div>

      <div className={side === 'right' ? 'text-right' : 'text-left'}>
        <p className="text-[13px] font-extrabold leading-tight" style={{ color: item.color }}>
          {item.title}
        </p>
        <p className="mt-0.5 text-[10px] leading-tight text-slate-400">{item.desc}</p>
      </div>
    </div>
  )
}
