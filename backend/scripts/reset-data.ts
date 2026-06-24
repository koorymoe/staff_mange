import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()

async function main() {
  console.log('🗑️ حذف البيانات التجريبية...')

  // حذف بالترتيب (العلاقات أولاً)
  const tables = [
    ['MissionEvent', prisma.missionEvent.deleteMany()],
    ['Mission', prisma.mission.deleteMany()],
    ['BookingAssignment', prisma.bookingAssignment.deleteMany()],
    ['CartItem', prisma.cartItem.deleteMany()],
    ['Booking', prisma.booking.deleteMany()],
    ['Customer', prisma.customer.deleteMany()],
    ['Expense', prisma.expense.deleteMany()],
    ['Complaint', prisma.complaint.deleteMany()],
    ['WorkReport', prisma.workReport.deleteMany()],
    ['Attendance', prisma.attendance.deleteMany()],
    ['EmployeeSkill', prisma.employeeSkill.deleteMany()],
    ['EmployeePermission', prisma.employeePermission.deleteMany()],
    ['Project', prisma.project.deleteMany()],
    ['Quotation', prisma.quotation.deleteMany()],
    ['Product', prisma.product.deleteMany()],
    ['Supplier', prisma.supplier.deleteMany()],
  ] as const

  for (const [name, promise] of tables) {
    const result = await promise
    if (result.count > 0) console.log(`  ✅ ${name}: ${result.count} deleted`)
  }

  // حذف كل الموظفين ماعدا الأدمن
  const admin = await prisma.employee.findFirst({ where: { role: 'ADMIN' } })
  if (admin) {
    const emp = await prisma.employee.deleteMany({ where: { id: { not: admin.id } } })
    if (emp.count > 0) console.log(`  ✅ Employees: ${emp.count} deleted`)
    console.log(`\n🔒 الأدمن محفوظ: ${admin.name} (${admin.username})`)
  }

  // تأكيد النتيجة
  const remaining = await prisma.employee.count()
  const services = await prisma.service.count()
  const skills = await prisma.skill.count()
  console.log(`\n📊 النتيجة:`)
  console.log(`  موظفين: ${remaining}`)
  console.log(`  خدمات: ${services} (محفوظة)`)
  console.log(`  مهارات: ${skills} (محفوظة)`)
  console.log('\n✨ تم! النظام جاهز للبيانات الحقيقية')
}

main().catch(console.error).finally(() => prisma.$disconnect())
