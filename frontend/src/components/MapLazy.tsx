import { Suspense, lazy } from 'react'

// ═══ الخرائط تنزل لمن تنعرض، مو مع كل فتحة للنظام ═══
//
// مكتبة الخرائط (leaflet) ١٤٩ كيلوبايت + ١٥ ستايل. جانت تنستورد
// استيراد عادي من MapViewer/LocationPicker، وهذولا منستوردين
// بالرئيسية والحجوزات والتنسيق وشاشة الحجز والموردين — يعني كل موظف
// يسجّل دخول ينزّل المكتبة كاملة حتى لو ما فتح ولا خريطة بيومه.
//
// هنا نلفهن بتحميل مؤجل: المكتبة ما تنزل إلا لمن تنرسم خريطة فعلاً.
// نفس الخصائص ونفس السلوك — بس بوقت التنزيل الصحيح.

const MapViewerInner = lazy(() => import('./MapViewer'))
const LocationPickerInner = lazy(() => import('./LocationPicker'))

// مربّع رمادي بنفس ارتفاع الخريطة حتى الصفحة ما تنقز لمن توصل المكتبة
const Placeholder = ({ height }: { height: number | string }) => (
  <div
    style={{ height: typeof height === 'number' ? `${height}px` : height }}
    className="flex w-full items-center justify-center rounded-xl bg-slate-100 text-xs text-slate-400"
  >
    جاري تحميل الخريطة...
  </div>
)

type MapViewerProps = React.ComponentProps<typeof MapViewerInner>
type LocationPickerProps = React.ComponentProps<typeof LocationPickerInner>

export function MapViewer(props: MapViewerProps) {
  return (
    <Suspense fallback={<Placeholder height={props.height ?? 260} />}>
      <MapViewerInner {...props} />
    </Suspense>
  )
}

export function LocationPicker(props: LocationPickerProps) {
  return (
    <Suspense fallback={<Placeholder height={300} />}>
      <LocationPickerInner {...props} />
    </Suspense>
  )
}
