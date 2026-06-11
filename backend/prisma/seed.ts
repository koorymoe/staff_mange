import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()

const services = [
  'كاميرات انلوك',
  'كاميرات IP',
  'انذار حريق',
  'GPS',
  'DASH CAM',
  'الاقفال الالكترونية',
  'الحاكيات',
  'اجهزة البصمة',
  'الطاقة الشمسية',
  'انذار سرقة',
  'شبكات',
  'ستلايت',
  'شاشات',
  'بيوت ذكية',
  'كاميرات تخصصية',
  'انظمة الصوت',
  'التليفونات البدالة',
  'كشوفات المشاريع',
  'خدمة تصميم الشبكات حسب الموقع والحاجة السلكية واللاسلكية',
  'تنفيذ شبكات للمؤسسات Wi-Fi والفنادق والمطاعم والمولات',
  'التليفونات البدالة IP + Analogue',
  'انترنيت المنازل (تنصيب وصيانة)',
  'خدمة ربط المواقع باستخدام VPN',
  'تنصيب وربط الطابعات الشبكية',
  'ادارة الحسابات (Active Directory)',
  'خدمة Hotspot لشبكات الواي فاي',
  'تنصيب برامج مراقبة الشبكة وادارة الشبكة عن بعد',
  'خدمة تصميم وتنفيذ ATS',
]

async function main() {
  for (const name of services) {
    await prisma.service.upsert({
      where: { name },
      update: {},
      create: { name },
    })
  }
  console.log(`Seeded ${services.length} services`)
}

main()
  .catch((e) => {
    console.error(e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
