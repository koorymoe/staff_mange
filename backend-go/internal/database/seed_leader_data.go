package database

import "github.com/jmoiron/sqlx"

// catalogSeedRow يمثّل صف واحد من كتالوج أسعار المنظومات الثمانية (تركيب/تسليك/برمجة).
type catalogSeedRow struct {
	SystemName string
	ItemName   string
	Category   string  // install | wiring | programming
	Value      float64 // سعر ثابت لل install/programming، أو مضاعف للمتر لل wiring
}

var systemPriceCatalogSeed = []catalogSeedRow{
	{SystemName: "كاميرات انلوك", ItemName: "تقسيم العدد الكلي لكاميرات على شاشتين في dvr 16 كاميرا فما فوق", Category: "install", Value: 7000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ربط مايك خارجي الى جهاز الكاميرات", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ستندر تنصيب و تشغيل الكاميرات", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تنصيب كاميرا على الستاند", Category: "install", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تنصيب راك كاميرات", Category: "install", Value: 17000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تثبيت ستاند شاشة 43", Category: "install", Value: 13500.0},
	{SystemName: "كاميرات انلوك", ItemName: "تثبيت ستاند شاشة اكبر من 43", Category: "install", Value: 16000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ترتيب منظومة ستاند", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ربط متحكم كي في ام", Category: "install", Value: 20000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تحويلة  الكيبل من  ip الى انلوك من خلال استخدام بولنة تحويل", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ربط يو بي اس", Category: "install", Value: 7000.0},
	{SystemName: "كاميرات انلوك", ItemName: "توسعة منظومة كاميرات  الانلوك بمقدار كاميرا او كامرتين من خلال برمجة واضافه كاميرا ip", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ترتيب منظومة كاميرات في الراك 4 يو", Category: "install", Value: 50000.0},
	{SystemName: "كاميرات انلوك", ItemName: "ترتيب منظومة كاميرات في الراك 6 يو", Category: "install", Value: 65000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تنصيب p to p", Category: "install", Value: 25000.0},
	{SystemName: "انذار حريق", ItemName: "نصب بورد الرئيسي", Category: "install", Value: 52000.0},
	{SystemName: "انذار حريق", ItemName: "تثبيت حساس  العادي", Category: "install", Value: 14000.0},
	{SystemName: "انذار حريق", ItemName: "تثبيت حساس معنون", Category: "install", Value: 19000.0},
	{SystemName: "انذار حريق", ItemName: "تثبيت جرس انذار", Category: "install", Value: 14000.0},
	{SystemName: "انذار حريق", ItemName: "تثبيت كاسر زجاجي", Category: "install", Value: 10000.0},
	{SystemName: "انذار حريق", ItemName: "تثبيت حساس ليزري", Category: "install", Value: 19000.0},
	{SystemName: "انذار حريق", ItemName: "اضافة جهاز اتصال sim", Category: "install", Value: 15000.0},
	{SystemName: "انذار حريق", ItemName: "تسليك كيبل حريق سعر المتر", Category: "install", Value: 2000.0},
	{SystemName: "انذار حريق", ItemName: "تنضيف حساس وصيانة", Category: "install", Value: 7000.0},
	{SystemName: "اجهزة البصمة", ItemName: "تثبيت جهاز بصمة", Category: "install", Value: 25000.0},
	{SystemName: "كاميرات  IP", ItemName: "تنصيب كاميرات ptz", Category: "install", Value: 25000.0},
	{SystemName: "كاميرات  IP", ItemName: "صيانة وتعديل ip كاميرات", Category: "install", Value: 30000.0},
	{SystemName: "كاميرات  IP", ItemName: "اضافة باج بنل لترتيب راك", Category: "install", Value: 15000.0},
	{SystemName: "كاميرات  IP", ItemName: "ستندر تنصيب و تشغيل كاميرات اقل او تساوي 20 متر", Category: "install", Value: 13500.0},
	{SystemName: "كاميرات  IP", ItemName: "ستندر تنصيب و تشغيل كاميرات اقل او تساوي 20 متر   VIP", Category: "install", Value: 25000.0},
	{SystemName: "كاميرات  IP", ItemName: "ستندر تنصيب و تشغيل الكاميرات بدون تسليك", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "سعر تسليك المتر فوق 3 متر كارتفاع للعمل العادي", Category: "install", Value: 16500.0},
	{SystemName: "كاميرات  IP", ItemName: "تسليك المتر الواحد تسليك عادي", Category: "install", Value: 1000.0},
	{SystemName: "كاميرات  IP", ItemName: "تسليك المتر  الواحد تسليك   بوري و تري", Category: "install", Value: 1750.0},
	{SystemName: "كاميرات  IP", ItemName: "تنصيب كاميرا على الستاند", Category: "install", Value: 5000.0},
	{SystemName: "كاميرات  IP", ItemName: "تنصيب راك كاميرات", Category: "install", Value: 17000.0},
	{SystemName: "كاميرات  IP", ItemName: "تثبيت ستاند شاشة 43", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "تثبيت ستاند شاشة اكبر من 43", Category: "install", Value: 15000.0},
	{SystemName: "كاميرات  IP", ItemName: "ترتيب منظومة ستاند", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "ترتيب منظومة كاميرات في الراك 4 يو", Category: "install", Value: 50000.0},
	{SystemName: "كاميرات  IP", ItemName: "ترتيب منظومة كاميرات في الراك 6 يو", Category: "install", Value: 65000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط متحكم كي في ام", Category: "install", Value: 20000.0},
	{SystemName: "كاميرات  IP", ItemName: "تحويلة  الكيبل من  ip الى انلوك من خلال استخدام بولنة تحويل", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط يو بي اس", Category: "install", Value: 7000.0},
	{SystemName: "كاميرات  IP", ItemName: "توسعة منظومة كاميرات  الانلوك بمقدار كاميرا او كامرتين من خلال برمجة واضافه كاميرا ip", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "تقسيم العدد الكلي لكاميرات على شاشتين في dvr 16 كاميرا فما فوق", Category: "install", Value: 7000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط مايك خارجي الى جهاز الكاميرات", Category: "install", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "تنصيب p to p", Category: "install", Value: 25000.0},
	{SystemName: "الاقفال الالكترونية", ItemName: "قفل باب حديد", Category: "install", Value: 40000.0},
	{SystemName: "الاقفال الالكترونية", ItemName: "قفل باب خشبي", Category: "install", Value: 50000.0},
	{SystemName: "الاقفال الالكترونية", ItemName: "قفل باب المنيوم", Category: "install", Value: 50000.0},
	{SystemName: "الاقفال الالكترونية", ItemName: "تثبيت قفل + لوحة", Category: "install", Value: 25000.0},
	{SystemName: "منظومة الصوت", ItemName: "نصب سماعة", Category: "install", Value: 10000.0},
	{SystemName: "منظومة الصوت", ItemName: "ربط مايك", Category: "install", Value: 20000.0},
	{SystemName: "منظومة الصوت", ItemName: "تسليك المتر الواحد", Category: "install", Value: 2000.0},
	{SystemName: "منظومة الصوت", ItemName: "ربط جهاز مانع صدى", Category: "install", Value: 20000.0},
	{SystemName: "منظومة الصوت", ItemName: "عمل بورد استعداء منصفل", Category: "install", Value: 50000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تسليك كيبل كاميرا انلوك", Category: "wiring", Value: 1.0},
	{SystemName: "كاميرات انلوك", ItemName: "كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "كاميرات انلوك", ItemName: "تسليك كيبل كاميرا انلوك vip", Category: "wiring", Value: 1.8},
	{SystemName: "كاميرات انلوك", ItemName: "كيبل كهرباء vip", Category: "wiring", Value: 1.7},
	{SystemName: "انذار حريق", ItemName: "تسليك كيبل حريق", Category: "wiring", Value: 1.2},
	{SystemName: "انذار حريق", ItemName: "تسليك كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "انذار حريق", ItemName: "كيبل كهرباء vip", Category: "wiring", Value: 1.7},
	{SystemName: "انذار حريق", ItemName: "كيبل حريق vip", Category: "wiring", Value: 2.0},
	{SystemName: "اجهزة البصمة", ItemName: "كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "اجهزة البصمة", ItemName: "تسليك كيبل lan", Category: "wiring", Value: 1.1},
	{SystemName: "كاميرات  IP", ItemName: "تسليك كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "كاميرات  IP", ItemName: "تسليك كيبل كاميرات ip", Category: "wiring", Value: 1.1},
	{SystemName: "كاميرات  IP", ItemName: "كيبل كهرباء vip", Category: "wiring", Value: 1.7},
	{SystemName: "الاقفال الالكترونية", ItemName: "تسليك كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "الاقفال الالكترونية", ItemName: "كيبل كهرباء vip", Category: "wiring", Value: 1.7},
	{SystemName: "منظومة الصوت", ItemName: "تسليك كيبل كهرباء", Category: "wiring", Value: 0.9},
	{SystemName: "منظومة الصوت", ItemName: "تسليك كيبل صوت", Category: "wiring", Value: 1.15},
	{SystemName: "منظومة الصوت", ItemName: "كيبل كهرباء vip", Category: "wiring", Value: 1.7},
	{SystemName: "منظومة الصوت", ItemName: "تسليك كيبل صوت vip", Category: "wiring", Value: 1.9},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة المنظومة مع ربطها بالتطبيق اضافة هاتف", Category: "programming", Value: 2500.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة المنظومة مع حجب جزء من الصورة لخصوصية الزبون", Category: "programming", Value: 2500.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة و تفعيل التسجيل على الحركة", Category: "programming", Value: 3000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة و اعدادات الكاميرا لتعديل الوضوح و تعديل الانارة", Category: "programming", Value: 2500.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة و اعدادت الكاميرا الخاصة بالتصوير الليلي الملون", Category: "programming", Value: 3500.0},
	{SystemName: "كاميرات انلوك", ItemName: "تشخيص مشكلة انطفاء الكاميرات و حل مشكلة البور", Category: "programming", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تشخيص مشكلة في جهاز دي في ار و حل مشكلة البور", Category: "programming", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة و تنصيب موسع اشارة", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة نوع الادخال انلوك الى  ip", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تغير نوع الكاميرات ااكثر من مرتين", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة للكاميرات البوليسية", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "اجور اضافية برمجة و تنصيب كاميرات المخصصة لتفعيل الوجهة", Category: "programming", Value: 7000.0},
	{SystemName: "كاميرات انلوك", ItemName: "حل مشاكل التخزين تبديل او عمل تهيئة", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "عمل جدول تنقل تلقائي في الكاميرات", Category: "programming", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تفعيل خاصية التعرف على الانسان  او المركبات", Category: "programming", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة و ميزة تخطيي الحدود او الخطوط وتحديد منطقة معينه", Category: "programming", Value: 7000.0},
	{SystemName: "كاميرات انلوك", ItemName: "شرح كيف سحب التسجيلات على usb خارجي", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة وتسجيل فقط عند وجود حركة", Category: "programming", Value: 5000.0},
	{SystemName: "كاميرات انلوك", ItemName: "برمجة كاميرا تخصصية", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "تشخيص مشكلة وحلها لكاميرا واحدة", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات انلوك", ItemName: "p to p برمجة", Category: "programming", Value: 25000.0},
	{SystemName: "انذار حريق", ItemName: "برمجة المنظومة التقليدية", Category: "programming", Value: 10000.0},
	{SystemName: "انذار حريق", ItemName: "برمجة المنظومة  conventonal", Category: "programming", Value: 25000.0},
	{SystemName: "اجهزة البصمة", ItemName: "برمجة جهاز البصمة", Category: "programming", Value: 25000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة ورسم مناطق التسلل (Intrusion Detection)", Category: "programming", Value: 7000.0},
	{SystemName: "كاميرات  IP", ItemName: "رسم خطوط العبور (Line Crossing).", Category: "programming", Value: 7000.0},
	{SystemName: "كاميرات  IP", ItemName: "تفعيل إنذارات السرقة الذكية (Perimeter Protection).", Category: "programming", Value: 15000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة التعرف على الوجه (Face Recognition).", Category: "programming", Value: 20000.0},
	{SystemName: "كاميرات  IP", ItemName: "إنشاء مستخدمين محدودي الصلاحيات.", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط كاميرات IP مع أنظمة إنذار أو Access Control.", Category: "programming", Value: 25000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط كاميرات LPR للتعرف على لوحات السيارات.", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة كاميرات للتتبع الذكي Auto Tracking.", Category: "programming", Value: 12000.0},
	{SystemName: "كاميرات  IP", ItemName: "تقنيات DNR وWDR لتحسين الإضاءة.", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "ربط كاميرا wifi مع منظومة ip", Category: "programming", Value: 17000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة بوينت  تو بوينت 2قطعة", Category: "programming", Value: 50000.0},
	{SystemName: "كاميرات  IP", ItemName: "تحليل ذكي إضافي مثل “Smart Search” و “Smart Playback” — تسهيل للبحث في الفيديو حسب الأحداث المهمة (مثل الأشخاص أو المركبات).", Category: "programming", Value: 20000.0},
	{SystemName: "كاميرات  IP", ItemName: "عمل خطة خطة لخزن البيانات وعمل \nنسخة احتياطية", Category: "programming", Value: 50000.0},
	{SystemName: "كاميرات  IP", ItemName: "تفعيل خاصية عد الاشخاص", Category: "programming", Value: 50000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة القناة الصفرية", Category: "programming", Value: 20000.0},
	{SystemName: "كاميرات  IP", ItemName: "برمجة كاميرا تخصصية", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "تشخيص مشكلة وحلها لكاميرا واحدة", Category: "programming", Value: 10000.0},
	{SystemName: "كاميرات  IP", ItemName: "p to p برمجة", Category: "programming", Value: 25000.0},
	{SystemName: "الاقفال الالكترونية", ItemName: "ربط قفل على  الهاتف اكثر من جهاز مبايل", Category: "programming", Value: 10000.0},
	{SystemName: "منظومة الصوت", ItemName: "ضبط صوت امبلفاير", Category: "programming", Value: 25000.0},
}

// materialSeedRow يمثّل صف واحد من أرشيف مواد الشد (نسخة من نسخة من مواد الشد).
type materialSeedRow struct {
	Name          string
	Code          string
	SellPrice     float64
	ProfitPerUnit float64
}

var materialArchiveSeed = []materialSeedRow{
	{Name: "60-30 ستاند عدل", Code: "94028", SellPrice: 11000.0, ProfitPerUnit: 6000.0},
	{Name: "120-60 ستاند عدل", Code: "94029", SellPrice: 13000.0, ProfitPerUnit: 4500.0},
	{Name: "ستاند عمود", Code: "94141", SellPrice: 3500.0, ProfitPerUnit: 1800.0},
	{Name: "جهاز حماية اسوار بدون شاشة", Code: "94236", SellPrice: 9500.0, ProfitPerUnit: 3500.0},
	{Name: "ستاند عكس 30 -60", Code: "94389", SellPrice: 15000.0, ProfitPerUnit: 6900.0},
	{Name: "ستاند مسطرة", Code: "94390", SellPrice: 5000.0, ProfitPerUnit: 1000.0},
	{Name: "ستاند Q-09 DVR", Code: "19932004", SellPrice: 6000.0, ProfitPerUnit: 2500.0},
	{Name: "BNC اصلي", Code: "94682", SellPrice: 1000.0, ProfitPerUnit: 500.0},
	{Name: "USB TO USB M/M 1.5M", Code: "94879", SellPrice: 3250.0, ProfitPerUnit: 2000.0},
	{Name: "USB TO USB M/F 5M", Code: "95152", SellPrice: 6250.0, ProfitPerUnit: 4500.0},
	{Name: "بوكس استراحة تركي ابيض", Code: "95490", SellPrice: 1250.0, ProfitPerUnit: 450.0},
	{Name: "ستاند شاشة TOKYOSAT 26-55INCH TS102", Code: "129402", SellPrice: 8000.0, ProfitPerUnit: 3000.0},
	{Name: "كيبل SHIELD PREMIUM HDMI TO HDMI 2.0V 4K 2160P 1.5M", Code: "224111", SellPrice: 3000.0, ProfitPerUnit: 1950.0},
	{Name: "فيش ستلايت", Code: "199416", SellPrice: 500.0, ProfitPerUnit: 230.0},
	{Name: "فيشة كهرباء كبس DC FEMAL", Code: "2001836", SellPrice: 500.0, ProfitPerUnit: 250.0},
	{Name: "بورت 8*1", Code: "1993154", SellPrice: 3250.0, ProfitPerUnit: 2000.0},
	{Name: "فيشة كهرباء كبس DC MALE", Code: "2001835", SellPrice: 500.0, ProfitPerUnit: 250.0},
	{Name: "FEMALE", Code: "1993157", SellPrice: 500.0, ProfitPerUnit: 250.0},
	{Name: "بورت 4*1", Code: "1993172", SellPrice: 2500.0, ProfitPerUnit: 2000.0},
	{Name: "موسع اشارة TP-link", Code: "1994622", SellPrice: 25000.0, ProfitPerUnit: 8000.0},
	{Name: "واير بورت 2*1 DC FMALE TO 2MALE", Code: "94315", SellPrice: 1250.0, ProfitPerUnit: 750.0},
	{Name: "كيبل كتان HDMI TO HDMI 3M", Code: "95389", SellPrice: 3500.0, ProfitPerUnit: 2100.0},
	{Name: "كيبل AL-WISAM HDMI TO HDMI 3M 4K*2K 2160P 2.0V X5803", Code: "95161", SellPrice: 4500.0, ProfitPerUnit: 3000.0},
	{Name: "كيبل SHIELD PREMIUM HDMI TO HDMI 2.0V 4K 2160P 15M", Code: "2001453", SellPrice: 15000.0, ProfitPerUnit: 9000.0},
	{Name: "كيبل كتان HDMI CABLE 1.5M", Code: "95072", SellPrice: 2000.0, ProfitPerUnit: 1400.0},
	{Name: "ستاند كاميرا مراقبة مزدوج 3-1.5 متر حرف L", Code: "2001661", SellPrice: 31000.0, ProfitPerUnit: 10000.0},
	{Name: "5A 12V محولة عنكبوتي", Code: "19931950", SellPrice: 8000.0, ProfitPerUnit: 4500.0},
	{Name: "HDMI TO LAN 20 متر", Code: "19931965", SellPrice: 8000.0, ProfitPerUnit: 4000.0},
	{Name: "120-60 ستاند L", Code: "19931967", SellPrice: 15000.0, ProfitPerUnit: 8500.0},
	{Name: "60-30 ستاند L", Code: "19931991", SellPrice: 10000.0, ProfitPerUnit: 1900.0},
	{Name: "LAN TO LAN", Code: "19932002", SellPrice: 1000.0, ProfitPerUnit: 500.0},
	{Name: "2TB HARD بنفسجي", Code: "2001964", SellPrice: 95000.0, ProfitPerUnit: 0.0},
	{Name: "1TB HARD بنفسجي", Code: "2001432", SellPrice: 68000.0, ProfitPerUnit: 0.0},
	{Name: "ذاكرة تخزين Seagate 500GB", Code: "19932016", SellPrice: 12000.0, ProfitPerUnit: 0.0},
	{Name: "HDMI 4K 1.5M", Code: "19932027", SellPrice: 3000.0, ProfitPerUnit: 1250.0},
	{Name: "CAT7 فيش", Code: "19937154", SellPrice: 1000.0, ProfitPerUnit: 250.0},
	{Name: "كيبل LAN UTP 1M  ابيض", Code: "19937164", SellPrice: 1500.0, ProfitPerUnit: 1000.0},
	{Name: "كيبل SFTP CAT6 2M SHIELD AL-WISAM", Code: "2001774", SellPrice: 2250.0, ProfitPerUnit: 1000.0},
	{Name: "محولة SYMBOL OF QUALITY CCTV POWER SUPPLY 12V/10A", Code: "2001665", SellPrice: 21000.0, ProfitPerUnit: 5000.0},
	{Name: "محول طاقة POWER SUPPLY 298 12V/20A S-250-12", Code: "2001904", SellPrice: 17250.0, ProfitPerUnit: 7250.0},
	{Name: "فيشة BLUESTORM  CAT6 BST-RJ45-U6-EZ", Code: "94042", SellPrice: 250.0, ProfitPerUnit: 0.0},
	{Name: "تري كيبل 16*16 سم", Code: "1993218", SellPrice: 1500.0, ProfitPerUnit: 750.0},
	{Name: "تري كيبل 16*25 سم", Code: "94886", SellPrice: 2000.0, ProfitPerUnit: 1000.0},
	{Name: "تري كيبل 25*25 سم", Code: "94885", SellPrice: 2500.0, ProfitPerUnit: 1500.0},
	{Name: "كيبل كاميرات CAMSCAN RG59 250Y COAXIAL //فريق الشد", Code: "2001980", SellPrice: 375.0, ProfitPerUnit: 100.0},
	{Name: "CABLE CAT6 SFTP 305M FALCONE دبل جاكيت  //فريق الشد", Code: "1993165", SellPrice: 375.0, ProfitPerUnit: 100.0},
	{Name: "بوري تغليف اردني", Code: "94887", SellPrice: 2500.0, ProfitPerUnit: 1000.0},
	{Name: "Stainless Steel Ring 300", Code: "94143", SellPrice: 1000.0, ProfitPerUnit: 500.0},
	{Name: "بوكس استراحة معيني", Code: "1993174", SellPrice: 1500.0, ProfitPerUnit: 500.0},
	{Name: "كيبل SHIELD PREMIUM HDMI  4K  10", Code: "2001454", SellPrice: 11000.0, ProfitPerUnit: 4000.0},
}

// seedSystemPriceCatalog يزرع كتالوج أسعار المنظومات الثمانية (نفس بيانات شيت
// "تكاليف المشروع" بالضبط) — idempotent عبر UNIQUE(systemName,itemName,category).
func seedSystemPriceCatalog(db *sqlx.DB) error {
	for i, row := range systemPriceCatalogSeed {
		id := "spc_" + itoa(i)
		if _, err := db.Exec(`
			INSERT INTO "SystemPriceCatalog" (id, "systemName", "itemName", category, value)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT ("systemName", "itemName", category) DO UPDATE SET value = EXCLUDED.value
		`, id, row.SystemName, row.ItemName, row.Category, row.Value); err != nil {
			return err
		}
	}
	return nil
}

// seedMaterialArchive يزرع أرشيف مواد الشد (نسخة من نسخة من مواد الشد) بكود
// فريد — idempotent عبر UNIQUE(code)، ويحدّث السعر/الربح لو الكود موجود أصلاً.
func seedMaterialArchive(db *sqlx.DB) error {
	for i, row := range materialArchiveSeed {
		id := "mat_" + itoa(i)
		if _, err := db.Exec(`
			INSERT INTO "Material" (id, name, code, "sellPrice", "profitPerUnit")
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, "sellPrice" = EXCLUDED."sellPrice", "profitPerUnit" = EXCLUDED."profitPerUnit"
		`, id, row.Name, row.Code, row.SellPrice, row.ProfitPerUnit); err != nil {
			return err
		}
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
