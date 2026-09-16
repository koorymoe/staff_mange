// parseCoordsFromUrl يستخرج الإحداثيات من رابط خرائط (جوجل/OSM) بأشكاله
// الشائعة. الروابط المختصرة (maps.app.goo.gl) ما بيها إحداثيات أصلاً — هذي
// تنحل بالسيرفر عبر /geo/resolve-map-link.
export function parseCoordsFromUrl(url: string): { lat: number; lng: number } | null {
  const patterns = [
    /@(-?\d+\.\d+),(-?\d+\.\d+)/,
    /[?&](?:q|query|ll|daddr|destination)=(-?\d+\.\d+)%2C(-?\d+\.\d+)/i,
    /[?&](?:q|query|ll|daddr|destination)=(-?\d+\.\d+),(-?\d+\.\d+)/i,
    /!3d(-?\d+\.\d+)!4d(-?\d+\.\d+)/,
    /[?&]mlat=(-?\d+\.\d+)&mlon=(-?\d+\.\d+)/i,
    /#map=\d+\/(-?\d+\.\d+)\/(-?\d+\.\d+)/,
    /\/(-?\d+\.\d+),(-?\d+\.\d+)/,
  ]
  for (const re of patterns) {
    const m = url.match(re)
    if (m) {
      const lat = parseFloat(m[1])
      const lng = parseFloat(m[2])
      if (Math.abs(lat) <= 90 && Math.abs(lng) <= 180) return { lat, lng }
    }
  }
  return null
}
