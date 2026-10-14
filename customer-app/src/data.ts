import { useEffect, useState } from 'react'
import { api, type CBooking, type Home, type SCard } from './api'

// كاش خفيف بالذاكرة: التنقل بين الشاشات يطلع فوري، والتحديث بالخلفية.
const cache: { services?: SCard[]; home?: Home; bookings?: CBooking[] } = {}

function useCached<K extends keyof typeof cache>(key: K, load: () => Promise<NonNullable<(typeof cache)[K]>>) {
  const [data, setData] = useState(cache[key])
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    load().then((d) => { cache[key] = d; if (alive) setData(d) }).catch((e) => { if (alive) setErr(e.message) })
    return () => { alive = false }
  }, [key, load])
  return { data, err, set: (d: NonNullable<(typeof cache)[K]>) => { cache[key] = d; setData(d) } }
}

export const useServices = () => useCached('services', api.services)
export const useHome = () => useCached('home', api.home)
export const useBookings = () => useCached('bookings', api.bookings)
export const clearCache = () => { delete cache.services; delete cache.home; delete cache.bookings }
