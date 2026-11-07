import { beforeEach, describe, expect, it, vi } from 'vitest'

// (مراجعة 11-07) طابور المراحل بلا نت: نداءين بنفس اللحظة ما يرسلون مرتين،
// «تم» ينتظر صور الحجز، وضغطة موظف ثاني ما تنرسل بجلسة غيره.

const calls: string[] = []
let pendingFor: Record<string, number> = {}

vi.mock('../api', () => ({
  api: {
    completeBooking: vi.fn(async (id: string) => { calls.push('COMPLETE:' + id) }),
    markArrived: vi.fn(async (id: string) => { calls.push('ARRIVED:' + id) }),
    stageProof: vi.fn(async () => {}),
  },
}))
vi.mock('./offlinePhotos', () => ({
  flushPhotos: vi.fn(async () => { await new Promise((r) => setTimeout(r, 5)); return 0 }),
  pendingPhotos: vi.fn(async (b: string) => Array(pendingFor[b] ?? 0).fill({})),
}))

const store = new Map<string, string>()
vi.stubGlobal('localStorage', {
  getItem: (k: string) => store.get(k) ?? null,
  setItem: (k: string, v: string) => { store.set(k, v) },
  removeItem: (k: string) => { store.delete(k) },
})
vi.stubGlobal('window', { dispatchEvent: () => true, addEventListener: () => {} })
vi.stubGlobal('navigator', { onLine: true })

const { flushStages, pendingStages } = await import('./offlineStages')

const queue = (items: object[]) => store.set('amani_stage_queue_v1', JSON.stringify(items))
const me = (id: string) => store.set('currentEmployee', JSON.stringify({ id }))

describe('طابور المراحل بلا نت', () => {
  beforeEach(() => { store.clear(); calls.length = 0; pendingFor = {}; me('e1') })

  it('نداءين بنفس اللحظة يرسلون كل ضغطة مرة وحدة', async () => {
    queue([{ id: 'a', employeeId: 'e1', bookingId: 'b1', code: 'B1', stage: 'ARRIVED', pressedAt: '' }])
    await Promise.all([flushStages(), flushStages()])
    expect(calls).toEqual(['ARRIVED:b1'])
    expect(pendingStages()).toEqual([])
  })

  it('«تم» ما ينرسل وصور الحجز بعدها بالطابور، ويبقى محفوظ', async () => {
    pendingFor = { b1: 1 }
    queue([{ id: 'c', employeeId: 'e1', bookingId: 'b1', code: 'B1', stage: 'COMPLETED', pressedAt: '' }])
    await flushStages()
    expect(calls).toEqual([])
    expect(pendingStages()).toHaveLength(1)
    pendingFor = {}
    await flushStages()
    expect(calls).toEqual(['COMPLETE:b1'])
    expect(pendingStages()).toEqual([])
  })

  it('ضغطة موظف ثاني تبقى لحد ما يدخل صاحبها', async () => {
    queue([{ id: 'x', employeeId: 'e2', bookingId: 'b2', code: 'B2', stage: 'ARRIVED', pressedAt: '' }])
    await flushStages()
    expect(calls).toEqual([])
    expect(pendingStages()).toHaveLength(1)
    me('e2')
    await flushStages()
    expect(calls).toEqual(['ARRIVED:b2'])
  })
})
