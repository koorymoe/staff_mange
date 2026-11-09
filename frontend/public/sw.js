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

// ═══ 🔔 (ع) 11-08: إشعارات حقيقية على الموبايل ═══
// الخادم يبعث كل إشعار جديد مشفّر، وهنا يطلع على الشاشة حتى والموقع مسدود.
self.addEventListener('push', (e) => {
  let d = {}
  try { d = e.data ? e.data.json() : {} } catch { d = { body: e.data ? e.data.text() : '' } }
  e.waitUntil(self.registration.showNotification(d.title || 'نظام الأماني', {
    body: d.body || '',
    tag: d.tag,
    dir: 'rtl',
    lang: 'ar',
    icon: 'secretary-icon-192.png',
    badge: 'favicon.png',
  }))
})
self.addEventListener('notificationclick', (e) => {
  e.notification.close()
  const scope = self.registration.scope
  e.waitUntil(self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((list) => {
    const open = list.find((c) => c.url.startsWith(scope))
    return open ? open.focus() : self.clients.openWindow(scope)
  }))
})
