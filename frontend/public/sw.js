// (ع) 10-14: النظام يفتح بلا نت — حتى الليدر يدوس «وصلنا/بدأنا/تم» عند الزبون.
// الشبكة أولاً دائماً (حتى ما يعلك إصدار قديم)، والكاش احتياط وقت ماكو نت.
// الـAPI ما ينحفظ هنا أبداً — بيانات «مهامي» تنحفظ بالتطبيق نفسه.
const CACHE = 'amani-staff-v1'
self.addEventListener('install', () => self.skipWaiting())
self.addEventListener('activate', (e) => {
  e.waitUntil(caches.keys().then((ks) => Promise.all(ks.filter((k) => k !== CACHE).map((k) => caches.delete(k)))).then(() => self.clients.claim()))
})
self.addEventListener('fetch', (e) => {
  const req = e.request
  const url = new URL(req.url)
  if (req.method !== 'GET' || url.origin !== location.origin || url.pathname.startsWith('/api/')) return
  e.respondWith(fetch(req).then((res) => {
    if (res.ok && (req.mode === 'navigate' || url.pathname.startsWith('/assets/') || /\.(png|svg|webp|ico|woff2?)$/.test(url.pathname))) {
      const copy = res.clone()
      caches.open(CACHE).then((c) => c.put(req.mode === 'navigate' ? '/' : req, copy))
    }
    return res
  }).catch(() => caches.match(req.mode === 'navigate' ? '/' : req).then((hit) => hit || Response.error())))
})
