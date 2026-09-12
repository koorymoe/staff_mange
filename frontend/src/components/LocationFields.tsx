import { useRef, useState } from 'react'
import { LocationPicker } from './MapLazy'
import { parseCoordsFromUrl } from '../lib/mapUrl'

interface Props {
  address: string
  onAddressChange: (v: string) => void
  point: { lat: number; lng: number } | null
  onPointChange: (p: { lat: number; lng: number } | null) => void
  locationUrl: string
  onLocationUrlChange: (v: string) => void
  /** دالة تطلب من السيرفر يفكّ الرابط (الروابط المختصرة) */
  resolveUrl: (url: string) => Promise<{ lat: number; lng: number }>
  addressLabel?: string
}

/**
 * LocationFields الطريقة الموحّدة لتحديد الموقع بكل النظام (الموردون،
 * المشاريع، الحجوزات): عنوان كلامي + رابط موقع + تحديد على الخريطة.
 *
 * الرابط يغني عن الخريطة: إذا انكتب رابط تنختفي الخريطة وتنحدد النقطة منه
 * تلقائياً — ما يصير يحط الاثنين سوه حتى ما تتضارب المعلومة.
 */
export default function LocationFields({
  address, onAddressChange,
  point, onPointChange,
  locationUrl, onLocationUrlChange,
  resolveUrl,
  addressLabel = 'العنوان (كلامي)',
}: Props) {
  const [urlError, setUrlError] = useState('')
  const [resolving, setResolving] = useState(false)
  // نتجاهل رد سيرفر متأخر لرابط قديم
  const urlRef = useRef(locationUrl)

  const handleUrl = (url: string) => {
    urlRef.current = url
    onLocationUrlChange(url)
    if (!url.trim()) { setUrlError(''); onPointChange(null); setResolving(false); return }
    const coords = parseCoordsFromUrl(url)
    if (coords) { onPointChange(coords); setUrlError(''); setResolving(false); return }
    onPointChange(null)
    setUrlError('')
    setResolving(true)
    const attempted = url
    resolveUrl(url)
      .then((p) => {
        if (attempted !== urlRef.current) return
        onPointChange(p); setUrlError('')
      })
      .catch((err: unknown) => {
        if (attempted !== urlRef.current) return
        setUrlError(err instanceof Error ? err.message : 'ما كدرنا نطلع الإحداثيات من هذا الرابط — حدد على الخريطة.')
      })
      .finally(() => { if (attempted === urlRef.current) setResolving(false) })
  }

  return (
    <div className="space-y-3">
      <div>
        <label className="mb-1 block text-sm font-medium text-slate-600">{addressLabel}</label>
        <input
          value={address}
          onChange={(e) => onAddressChange(e.target.value)}
          placeholder="مثال: كربلاء - شارع الجمهورية - قرب..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
        />
      </div>

      <div>
        <label className="mb-1 block text-sm font-medium text-slate-600">رابط الموقع (بديل عن الخريطة)</label>
        <input
          value={locationUrl}
          onChange={(e) => handleUrl(e.target.value)}
          placeholder="الصق رابط الموقع من خرائط جوجل..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          dir="ltr"
        />
        {resolving && <p className="mt-1 text-xs font-bold text-brand-600">جاري فتح الرابط واستخراج الموقع...</p>}
        {urlError && <p className="mt-1 text-xs font-bold text-red-600">{urlError}</p>}
        {locationUrl.trim() && point && (
          <p className="mt-1 text-xs font-bold text-emerald-700">
            ✓ انحددت النقطة تلقائياً من الرابط ({point.lat.toFixed(5)}, {point.lng.toFixed(5)})
          </p>
        )}
      </div>

      {/* الخريطة تختفي لما يكون اكو رابط — الاثنين سوه يتضاربون */}
      {!locationUrl.trim() && (
        <LocationPicker value={point} onChange={onPointChange} />
      )}
    </div>
  )
}
