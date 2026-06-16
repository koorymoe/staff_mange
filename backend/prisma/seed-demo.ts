import { PrismaClient } from '@prisma/client'
import bcrypt from 'bcryptjs'

const prisma = new PrismaClient()

async function main() {
  const hash = (pw: string) => bcrypt.hashSync(pw, 10)

  // ── 1. خدمات رئيسية للبيانات الوهمية ──────────────────────────────
  const svcNames = ['كاميرات IP', 'انذار حريق', 'الطاقة الشمسية', 'شبكات', 'GPS', 'بيوت ذكية', 'انذار سرقة']
  const svcs: Record<string, { id: string; skills: { id: string; name: string }[] }> = {}
  for (const name of svcNames) {
    const s = await prisma.service.findUnique({ where: { name }, include: { skills: true } })
    if (s) svcs[name] = { id: s.id, skills: s.skills }
  }

  // ── 2. الموظفون ──────────────────────────────────────────────────────
  const adminEmp = await prisma.employee.upsert({
    where: { username: 'admin' },
    update: {},
    create: { name: 'مدير النظام', username: 'admin', password: hash('admin123'), role: 'ADMIN', position: 'مدير عام', phone: '07701000001' },
  })

  const monitor = await prisma.employee.upsert({
    where: { username: 'monitor1' },
    update: {},
    create: { name: 'حسام المراقب', username: 'monitor1', password: hash('monitor123'), role: 'MONITOR', position: 'مراقب عمليات', phone: '07701000002' },
  })

  const finance = await prisma.employee.upsert({
    where: { username: 'finance1' },
    update: {},
    create: { name: 'كرار المحاسب', username: 'finance1', password: hash('finance123'), role: 'FINANCE', position: 'محاسب', phone: '07701000003' },
  })

  const coord1 = await prisma.employee.upsert({
    where: { username: 'coord1' },
    update: {},
    create: { name: 'يوسف عبيد', username: 'coord1', password: hash('coord123'), role: 'HR_COORDINATOR', position: 'إداري كوادر', phone: '07701000004' },
  })

  const coord2 = await prisma.employee.upsert({
    where: { username: 'coord2' },
    update: {},
    create: { name: 'علي حسين', username: 'coord2', password: hash('coord123'), role: 'HR_COORDINATOR', position: 'إداري كوادر', phone: '07701000005' },
  })

  const sales1 = await prisma.employee.upsert({
    where: { username: 'sales1' },
    update: {},
    create: { name: 'أحمد مبيعات', username: 'sales1', password: hash('sales123'), role: 'SALES', position: 'موظف مبيعات', phone: '07701000006' },
  })

  const sales2 = await prisma.employee.upsert({
    where: { username: 'sales2' },
    update: {},
    create: { name: 'زينب مبيعات', username: 'sales2', password: hash('sales123'), role: 'SALES', position: 'موظف مبيعات', phone: '07701000007' },
  })

  const tech1 = await prisma.employee.upsert({
    where: { username: 'tech1' },
    update: {},
    create: { name: 'مصطفى الفني', username: 'tech1', password: hash('tech123'), role: 'TECHNICIAN', position: 'فني أول', phone: '07701000008', hasDrivingLicense: true, hasSafetyCertificate: true },
  })

  const tech2 = await prisma.employee.upsert({
    where: { username: 'tech2' },
    update: {},
    create: { name: 'حيدر الفني', username: 'tech2', password: hash('tech123'), role: 'TECHNICIAN', position: 'فني ثاني', phone: '07701000009', hasDrivingLicense: true },
  })

  const tech3 = await prisma.employee.upsert({
    where: { username: 'tech3' },
    update: {},
    create: { name: 'علاء الفني', username: 'tech3', password: hash('tech123'), role: 'TECHNICIAN', position: 'فني', phone: '07701000010', hasSafetyCertificate: true },
  })

  const tech4 = await prisma.employee.upsert({
    where: { username: 'tech4' },
    update: {},
    create: { name: 'كريم الفني', username: 'tech4', password: hash('tech123'), role: 'TECHNICIAN', position: 'فني', phone: '07701000011' },
  })

  const pm1 = await prisma.employee.upsert({
    where: { username: 'pm1' },
    update: {},
    create: { name: 'سامر المشرف', username: 'pm1', password: hash('pm123'), role: 'PROJECT_MANAGER', position: 'مشرف مشاريع', phone: '07701000012' },
  })

  // ── 3. مهارات الفنيين ────────────────────────────────────────────────
  const assignSkills = async (empId: string, serviceNames: string[]) => {
    for (const sn of serviceNames) {
      const svc = svcs[sn]
      if (!svc) continue
      for (const skill of svc.skills.slice(0, 3)) {
        await prisma.employeeSkill.upsert({
          where: { employeeId_skillId: { employeeId: empId, skillId: skill.id } },
          update: {},
          create: { employeeId: empId, skillId: skill.id, canPerform: true },
        })
      }
    }
  }

  await assignSkills(tech1.id, ['كاميرات IP', 'شبكات', 'بيوت ذكية'])
  await assignSkills(tech2.id, ['انذار حريق', 'انذار سرقة', 'GPS'])
  await assignSkills(tech3.id, ['الطاقة الشمسية', 'كاميرات IP'])
  await assignSkills(tech4.id, ['GPS', 'انذار سرقة', 'شبكات'])

  // ── 4. الزبائن ──────────────────────────────────────────────────────
  const custData = [
    { name: 'شركة النور للتجارة', phone: '07810000001', location: 'بغداد - الكرادة' },
    { name: 'مطعم الزيتونة', phone: '07810000002', location: 'بغداد - المنصور' },
    { name: 'فندق دجلة', phone: '07810000003', location: 'بغداد - الجادرية' },
    { name: 'مدرسة الأمل الأهلية', phone: '07810000004', location: 'بغداد - الدورة' },
    { name: 'مستوصف الرعاية', phone: '07810000005', location: 'بغداد - الشعب' },
    { name: 'محلات السعد الإلكترونية', phone: '07810000006', location: 'بغداد - البياع' },
    { name: 'مصنع الصناعات الخفيفة', phone: '07810000007', location: 'بغداد - الزعفرانية' },
    { name: 'بنك التجارة الأهلي', phone: '07810000008', location: 'بغداد - الوزيرية' },
    { name: 'أبو حسن للمقاولات', phone: '07810000009', location: 'بغداد - الحرية' },
    { name: 'كافيه ماروني', phone: '07810000010', location: 'بغداد - الحارثية' },
    { name: 'مجمع عمر التجاري', phone: '07810000011', location: 'بغداد - الكاظمية' },
    { name: 'شركة البريد السريع', phone: '07810000012', location: 'بغداد - الأعظمية' },
  ]

  const customers: Record<string, string> = {} // name → id
  for (const c of custData) {
    const cust = await prisma.customer.upsert({
      where: { phone: c.phone },
      update: {},
      create: c,
    })
    customers[c.name] = cust.id
  }

  // ── 5. الحجوزات ──────────────────────────────────────────────────────
  // helper
  const daysAgo = (n: number) => new Date(Date.now() - n * 86400000)
  const daysFromNow = (n: number) => new Date(Date.now() + n * 86400000)

  const bookingDefs = [
    // ═══ مكتملة (COMPLETED) ═══
    {
      custName: 'شركة النور للتجارة',
      svc: 'كاميرات IP',
      sales: sales1,
      coord: coord1,
      status: 'COMPLETED' as const,
      priority: 'NORMAL' as const,
      quotedPrice: 850000,
      address: 'بغداد - الكرادة، شارع فلسطين',
      assignedVehicle: 'كيا سول - ١٢٣٤',
      techs: [tech1, tech2],
      amountCollected: 850000,
      advancePaid: 300000,
      amountVerified: true,
      completionNotes: 'تم تركيب 8 كاميرات IP مع NVR وتفعيل المشاهدة عن بعد',
      createdAt: daysAgo(20),
      scheduledAt: daysAgo(15),
      completedAt: daysAgo(14),
    },
    {
      custName: 'مطعم الزيتونة',
      svc: 'انذار حريق',
      sales: sales2,
      coord: coord1,
      status: 'COMPLETED' as const,
      priority: 'URGENT' as const,
      quotedPrice: 650000,
      address: 'بغداد - المنصور، شارع ١٤ رمضان',
      assignedVehicle: 'بيك آب - ٥٦٧٨',
      techs: [tech2, tech3],
      amountCollected: 650000,
      advancePaid: 200000,
      amountVerified: true,
      completionNotes: 'تم تركيب منظومة إنذار حريق كاملة مع 12 كاشف دخان',
      createdAt: daysAgo(18),
      scheduledAt: daysAgo(12),
      completedAt: daysAgo(11),
    },
    {
      custName: 'مدرسة الأمل الأهلية',
      svc: 'الطاقة الشمسية',
      sales: sales1,
      coord: coord2,
      status: 'COMPLETED' as const,
      priority: 'NORMAL' as const,
      quotedPrice: 3500000,
      address: 'بغداد - الدورة، الحي الصناعي',
      assignedVehicle: 'كيا سول - ١٢٣٤',
      techs: [tech1, tech3, tech4],
      amountCollected: 3200000,
      advancePaid: 1000000,
      amountVerified: false,
      completionNotes: 'تم تركيب 20 لوح شمسي مع إنفرتر 10kW وبطاريات',
      createdAt: daysAgo(30),
      scheduledAt: daysAgo(22),
      completedAt: daysAgo(20),
    },
    {
      custName: 'بنك التجارة الأهلي',
      svc: 'شبكات',
      sales: sales2,
      coord: coord1,
      status: 'COMPLETED' as const,
      priority: 'URGENT' as const,
      quotedPrice: 1200000,
      address: 'بغداد - الوزيرية، شارع الصناعة',
      assignedVehicle: 'بيك آب - ٥٦٧٨',
      techs: [tech1, tech4],
      amountCollected: 1200000,
      advancePaid: 500000,
      amountVerified: true,
      completionNotes: 'تم تمديد شبكة كاملة للطابق الأرضي والأول',
      createdAt: daysAgo(25),
      scheduledAt: daysAgo(18),
      completedAt: daysAgo(17),
    },
    {
      custName: 'كافيه ماروني',
      svc: 'بيوت ذكية',
      sales: sales1,
      coord: coord2,
      status: 'COMPLETED' as const,
      priority: 'NORMAL' as const,
      quotedPrice: 750000,
      address: 'بغداد - الحارثية، جانب دوار الرصافة',
      assignedVehicle: 'كيا سول - ١٢٣٤',
      techs: [tech1, tech2],
      amountCollected: 700000,
      advancePaid: 250000,
      amountVerified: false,
      completionNotes: 'تم ربط الإضاءة والتكييف بنظام تحكم ذكي',
      createdAt: daysAgo(15),
      scheduledAt: daysAgo(8),
      completedAt: daysAgo(7),
    },
    // ═══ جاري التنفيذ (IN_PROGRESS) ═══
    {
      custName: 'فندق دجلة',
      svc: 'كاميرات IP',
      sales: sales1,
      coord: coord1,
      status: 'IN_PROGRESS' as const,
      priority: 'URGENT' as const,
      quotedPrice: 4500000,
      address: 'بغداد - الجادرية، كورنيش دجلة',
      assignedVehicle: 'بيك آب - ٥٦٧٨',
      techs: [tech1, tech2, tech3],
      createdAt: daysAgo(10),
      scheduledAt: daysAgo(3),
    },
    {
      custName: 'مصنع الصناعات الخفيفة',
      svc: 'الطاقة الشمسية',
      sales: sales2,
      coord: coord2,
      status: 'IN_PROGRESS' as const,
      priority: 'NORMAL' as const,
      quotedPrice: 6000000,
      address: 'بغداد - الزعفرانية، المنطقة الصناعية',
      assignedVehicle: 'كيا سول - ١٢٣٤',
      techs: [tech3, tech4],
      createdAt: daysAgo(7),
      scheduledAt: daysAgo(2),
    },
    // ═══ مثبتة (CONFIRMED) ═══
    {
      custName: 'مستوصف الرعاية',
      svc: 'انذار حريق',
      sales: sales1,
      coord: coord1,
      status: 'CONFIRMED' as const,
      priority: 'URGENT' as const,
      quotedPrice: 900000,
      address: 'بغداد - الشعب، خلف المستشفى',
      assignedVehicle: 'بيك آب - ٥٦٧٨',
      techs: [tech2],
      createdAt: daysAgo(5),
      scheduledAt: daysFromNow(1),
    },
    {
      custName: 'محلات السعد الإلكترونية',
      svc: 'GPS',
      sales: sales2,
      coord: coord2,
      status: 'CONFIRMED' as const,
      priority: 'NORMAL' as const,
      quotedPrice: 350000,
      address: 'بغداد - البياع، الشارع العام',
      techs: [tech4],
      createdAt: daysAgo(3),
      scheduledAt: daysFromNow(2),
    },
    {
      custName: 'مجمع عمر التجاري',
      svc: 'شبكات',
      sales: sales1,
      coord: coord1,
      status: 'CONFIRMED' as const,
      priority: 'URGENT' as const,
      quotedPrice: 2200000,
      address: 'بغداد - الكاظمية، شارع الإمام',
      assignedVehicle: 'كيا سول - ١٢٣٤',
      techs: [tech1, tech4],
      createdAt: daysAgo(2),
      scheduledAt: daysFromNow(3),
    },
    // ═══ بانتظار التثبيت (PENDING) ═══
    {
      custName: 'أبو حسن للمقاولات',
      svc: 'انذار سرقة',
      sales: sales2,
      coord: null,
      status: 'PENDING' as const,
      priority: 'NORMAL' as const,
      createdAt: daysAgo(2),
    },
    {
      custName: 'شركة البريد السريع',
      svc: 'كاميرات IP',
      sales: sales1,
      coord: null,
      status: 'PENDING' as const,
      priority: 'URGENT' as const,
      createdAt: daysAgo(1),
    },
    {
      custName: 'كافيه ماروني',
      svc: 'GPS',
      sales: sales2,
      coord: null,
      status: 'PENDING' as const,
      priority: 'NORMAL' as const,
      createdAt: daysAgo(0),
    },
    // ═══ ملغاة (CANCELLED) ═══
    {
      custName: 'أبو حسن للمقاولات',
      svc: 'بيوت ذكية',
      sales: sales1,
      coord: null,
      status: 'CANCELLED' as const,
      priority: 'NORMAL' as const,
      createdAt: daysAgo(40),
    },
  ]

  let seqNum = await prisma.booking.count()

  for (const def of bookingDefs) {
    seqNum++
    const custId = customers[def.custName]
    if (!custId) { console.log('Missing customer:', def.custName); continue }

    const svcData = svcs[def.svc]
    const isConfirmed = def.status !== 'PENDING' && def.status !== 'CANCELLED'

    const booking = await prisma.booking.create({
      data: {
        code: `B${seqNum}`,
        sequenceNumber: seqNum,
        customerId: custId,
        serviceId: svcData?.id,
        status: def.status,
        priority: def.priority,
        transferEmployeeId: def.sales.id,
        confirmedByEmployeeId: isConfirmed && def.coord ? def.coord.id : undefined,
        confirmedByName: isConfirmed && def.coord ? def.coord.name : undefined,
        transferToProjects: false,
        quotedPrice: (def as any).quotedPrice ?? null,
        address: (def as any).address ?? null,
        assignedVehicle: (def as any).assignedVehicle ?? null,
        scheduledAt: (def as any).scheduledAt ?? null,
        completedAt: (def as any).completedAt ?? null,
        completionNotes: (def as any).completionNotes ?? null,
        amountCollected: (def as any).amountCollected ?? null,
        advancePaid: (def as any).advancePaid ?? null,
        amountVerified: (def as any).amountVerified ?? false,
        adminNotes: null,
        createdAt: def.createdAt,
      },
    })

    // تعيين الفنيين
    if (isConfirmed && (def as any).techs) {
      const techRoles = ['TECH_1', 'TECH_2', 'TECH_3'] as const
      for (let i = 0; i < (def as any).techs.length; i++) {
        await prisma.bookingAssignment.upsert({
          where: { bookingId_role: { bookingId: booking.id, role: techRoles[i] } },
          update: {},
          create: { bookingId: booking.id, employeeId: (def as any).techs[i].id, role: techRoles[i] },
        })
      }
    }

    console.log(`✓ Created booking ${booking.code} - ${def.custName} - ${def.status}`)
  }

  // ── 6. المصاريف ──────────────────────────────────────────────────────
  const expensesData = [
    { emp: tech1, amount: 25000, desc: 'شراء كيبلات CAT6 إضافية للموقع', status: 'APPROVED' as const, daysAgo: 10 },
    { emp: tech2, amount: 15000, desc: 'بنزين لرحلة موقع المستوصف', status: 'APPROVED' as const, daysAgo: 8 },
    { emp: tech3, amount: 45000, desc: 'شراء مواد توصيل للطاقة الشمسية', status: 'PENDING' as const, daysAgo: 3 },
    { emp: tech4, amount: 12000, desc: 'وجبات فريق العمل في الموقع', status: 'PENDING' as const, daysAgo: 2 },
    { emp: tech1, amount: 80000, desc: 'قطع غيار NVR - طارئ', status: 'APPROVED' as const, daysAgo: 15 },
    { emp: tech2, amount: 35000, desc: 'أدوات تفييش احتياطية', status: 'REJECTED' as const, daysAgo: 20 },
    { emp: tech3, amount: 20000, desc: 'بنزين موقع الزعفرانية', status: 'PENDING' as const, daysAgo: 1 },
    { emp: sales1, amount: 10000, desc: 'مواصلات زيارة زبون - الكاظمية', status: 'APPROVED' as const, daysAgo: 5 },
    { emp: sales2, amount: 8000, desc: 'هدية بسيطة لزبون - ترحيب', status: 'REJECTED' as const, daysAgo: 7 },
    { emp: tech4, amount: 18000, desc: 'كيبلات شبكة إضافية', status: 'APPROVED' as const, daysAgo: 12 },
  ]

  for (const e of expensesData) {
    await prisma.expense.create({
      data: {
        employeeId: e.emp.id,
        amount: e.amount,
        description: e.desc,
        status: e.status,
        createdAt: new Date(Date.now() - e.daysAgo * 86400000),
      },
    })
    console.log(`✓ Expense: ${e.emp.name} - ${e.amount} - ${e.status}`)
  }

  console.log('\n✅ Demo seed complete!')
  console.log('\nبيانات تسجيل الدخول:')
  console.log('مدير النظام:  admin / admin123')
  console.log('المراقب:      monitor1 / monitor123')
  console.log('المحاسب:      finance1 / finance123')
  console.log('الإداري 1:    coord1 / coord123')
  console.log('الإداري 2:    coord2 / coord123')
  console.log('مبيعات 1:     sales1 / sales123')
  console.log('مبيعات 2:     sales2 / sales123')
  console.log('فني 1:        tech1 / tech123')
  console.log('فني 2:        tech2 / tech123')
  console.log('فني 3:        tech3 / tech123')
  console.log('فني 4:        tech4 / tech123')
}

main()
  .catch((e) => { console.error(e); process.exit(1) })
  .finally(() => prisma.$disconnect())
