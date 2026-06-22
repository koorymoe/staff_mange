import { Router } from 'express'
import { prisma } from '../prisma'

const router = Router()

// GET /specialties
router.get('/specialties', async (_req, res) => {
  const specs = await prisma.supplierSpecialty.findMany({ orderBy: { order: 'asc' } })
  res.json(specs)
})

// POST /specialties
router.post('/specialties', async (req, res) => {
  const { name } = req.body
  const existing = await prisma.supplierSpecialty.findUnique({ where: { name } })
  if (existing) return res.status(400).json({ error: 'التخصص موجود مسبقاً' })
  const count = await prisma.supplierSpecialty.count()
  const spec = await prisma.supplierSpecialty.create({ data: { name, order: count } })
  res.json(spec)
})

// DELETE /specialties/:id
router.delete('/specialties/:id', async (req, res) => {
  await prisma.supplierSpecialty.delete({ where: { id: req.params.id } })
  res.json({ ok: true })
})

// GET / - list all suppliers with specialties and avg rating
router.get('/', async (_req, res) => {
  const suppliers = await prisma.supplier.findMany({
    include: {
      specialties: { include: { specialty: true } },
      ratings: { orderBy: { createdAt: 'desc' } },
    },
    orderBy: { createdAt: 'desc' },
  })
  res.json(suppliers.map(s => ({
    ...s,
    specialties: s.specialties.map(sl => sl.specialty),
    avgRating: s.ratings.length > 0
      ? s.ratings.reduce((sum, r) => sum + r.value, 0) / s.ratings.length
      : 0,
    ratingCount: s.ratings.length,
  })))
})

// POST / - create supplier
router.post('/', async (req, res) => {
  const { companyName, ownerName, phone, lat, lng, isMaterialSupplier, isContractor, traderTypes, notes, specialtyIds } = req.body
  const supplier = await prisma.supplier.create({
    data: {
      companyName, ownerName, phone,
      lat: lat || null, lng: lng || null,
      isMaterialSupplier: !!isMaterialSupplier,
      isContractor: !!isContractor,
      traderTypes: traderTypes || [],
      notes: notes || null,
      specialties: {
        create: (specialtyIds || []).map((id: string) => ({ specialtyId: id })),
      },
    },
    include: {
      specialties: { include: { specialty: true } },
      ratings: true,
    },
  })
  res.json({ ...supplier, specialties: supplier.specialties.map(sl => sl.specialty), avgRating: 0, ratingCount: 0 })
})

// PUT /:id - update supplier
router.put('/:id', async (req, res) => {
  const { companyName, ownerName, phone, lat, lng, isMaterialSupplier, isContractor, traderTypes, notes, specialtyIds } = req.body
  await prisma.supplierSpecialtyLink.deleteMany({ where: { supplierId: req.params.id } })
  const supplier = await prisma.supplier.update({
    where: { id: req.params.id },
    data: {
      companyName, ownerName, phone,
      lat: lat || null, lng: lng || null,
      isMaterialSupplier: !!isMaterialSupplier,
      isContractor: !!isContractor,
      traderTypes: traderTypes || [],
      notes: notes || null,
      specialties: {
        create: (specialtyIds || []).map((id: string) => ({ specialtyId: id })),
      },
    },
    include: {
      specialties: { include: { specialty: true } },
      ratings: { orderBy: { createdAt: 'desc' } },
    },
  })
  res.json({
    ...supplier,
    specialties: supplier.specialties.map(sl => sl.specialty),
    avgRating: supplier.ratings.length > 0 ? supplier.ratings.reduce((sum, r) => sum + r.value, 0) / supplier.ratings.length : 0,
    ratingCount: supplier.ratings.length,
  })
})

// DELETE /:id
router.delete('/:id', async (req, res) => {
  await prisma.supplier.delete({ where: { id: req.params.id } })
  res.json({ ok: true })
})

// POST /:id/rate - add a rating
router.post('/:id/rate', async (req, res) => {
  const { value, note, ratedById, ratedByName } = req.body
  if (!value || value < 1 || value > 5) return res.status(400).json({ error: 'التقييم يجب أن يكون بين 1 و 5' })
  const rating = await prisma.supplierRating.create({
    data: {
      supplierId: req.params.id,
      value, note: note || null,
      ratedById, ratedByName,
    },
  })
  res.json(rating)
})

export default router
