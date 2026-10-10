package database

import (
	"github.com/jmoiron/sqlx"
)

// ═══ محتوى المرحلة الأولى: لوحة تحكم دخول مستقلة ═══
//
// «القفل الإلكتروني بيه ١٥ واير ملوّنات كل لون شنو يعني وشلون
// يربطهنة بكهرباء وشلون يبرمجه وشنو الطريقة الصحيحة».
//
// ═══════════════════════════════════════════════════════════════
// ⚠️⚠️ اقرا هذا قبل أي شي ⚠️⚠️
//
// هذا المحتوى **غير محقّق** (`verified = FALSE`). مبني على الأعراف
// المنشورة والشائعة بلوحات التحكم بالدخول المستقلة — **مو على كتالوگ
// موديل بعينه**. ألوان الأسلاك تختلف فعلاً بين مصنّع وآخر.
//
// ولهذا السبب بالضبط ما يوصل أي متدرّب: الاستعلام بالمستودع يشترط
// `verified = TRUE`، فهذا الجهاز يشوفه **المالك وحده** لحد ما يجرّبه فني
// على جهاز حقيقي بالورشة ويعتمده.
//
// الهدف منه بهالمرحلة: **إثبات المحرّك** — إن التوصيل والتحقق وتفسير
// الغلط والدرجة كلها تشتغل. المحتوى الحقيقي يجي بعد الاعتماد.
// ═══════════════════════════════════════════════════════════════
//
// ⚠️ البذرة idempotent: تنضاف مرة وحدة وما تنكتب فوق أي تعديل يسويه
// المالك بعدين. المعرّفات ثابتة عمداً حتى تنعرف بأي تشغيل.
func seedSimLock(db *sqlx.DB) error {
	const (
		catID = "simcat_access_control"
		devID = "simdev_ac_keypad_15w"
		psuID = "simdev_psu_12v"
		exID  = "simex_ac_keypad_wiring"
		lesID = "simles_ac_keypad_colors"
	)

	// الفئة — تنربط بخدمة «الاقفال الالكترونية» إذا موجودة.
	if _, err := db.Exec(`
		INSERT INTO "SimCategory" (id, "serviceId", name, description, "sortOrder")
		SELECT $1,
		       (SELECT id FROM "Service" WHERE name ILIKE '%قفال%' LIMIT 1),
		       'الأقفال وأنظمة التحكم بالدخول',
		       'توصيل لوحات التحكم بالدخول وبرمجتها — الأسلاك، التغذية، الريلاي، وزر الخروج.',
		       10
		ON CONFLICT (id) DO NOTHING`, catID); err != nil {
		return err
	}

	// ═══ الجهاز: لوحة تحكم دخول بـ١٥ سلك ═══
	//
	// ⚠️ `x`/`y` نسب ٠..١ من صورة الجهاز، مو بكسلات — التعريف يبقى صالحاً
	// بأي حجم صورة أو شاشة.
	// ⚠️ `danger` على الأطراف الخطرة: تطلع لمن يمر عليها، وتطلع بتفسير
	// الغلط. المتدرّب لازم يعرف **العاقبة** مو بس إنه غلط.
	terminals := `[
	  {"id":"t_red","label":"أحمر","colorHex":"#dc2626","colorName":"RED","kind":"POWER_POS","signal":"+12V",
	   "x":0.12,"y":0.18,"description":"موجب التغذية — ١٢ فولت مستمر",
	   "danger":"عكسه مع الأسود يحرق اللوحة فوراً وما ينفع الضمان"},
	  {"id":"t_black","label":"أسود","colorHex":"#111827","colorName":"BLACK","kind":"POWER_NEG","signal":"GND",
	   "x":0.12,"y":0.26,"description":"سالب التغذية (الأرضي المشترك)"},
	  {"id":"t_blue","label":"أزرق","colorHex":"#2563eb","colorName":"BLUE","kind":"RELAY_NO","signal":"NO",
	   "x":0.12,"y":0.34,"description":"تماس الريلاي المفتوح — ينغلق لحظة الفتح"},
	  {"id":"t_purple","label":"بنفسجي","colorHex":"#7c3aed","colorName":"PURPLE","kind":"RELAY_COM","signal":"COM",
	   "x":0.12,"y":0.42,"description":"مشترك الريلاي — منه تمرّ تغذية القفل"},
	  {"id":"t_orange","label":"برتقالي","colorHex":"#ea580c","colorName":"ORANGE","kind":"RELAY_NC","signal":"NC",
	   "x":0.12,"y":0.50,"description":"تماس الريلاي المغلق — للأقفال المغناطيسية"},
	  {"id":"t_yellow","label":"أصفر","colorHex":"#ca8a04","colorName":"YELLOW","kind":"INPUT","signal":"EXIT",
	   "x":0.12,"y":0.58,"description":"مدخل زر الخروج — يفتح من داخل الغرفة"},
	  {"id":"t_brown","label":"بني","colorHex":"#92400e","colorName":"BROWN","kind":"INPUT","signal":"DOOR_SENSOR",
	   "x":0.12,"y":0.66,"description":"حسّاس الباب — يعرف إذا الباب بقى مفتوحاً"},
	  {"id":"t_green","label":"أخضر","colorHex":"#16a34a","colorName":"GREEN","kind":"DATA","signal":"WIEGAND_D0",
	   "x":0.12,"y":0.74,"description":"داتا ويگند D0 — لقارئ خارجي","pairId":"wiegand"},
	  {"id":"t_white","label":"أبيض","colorHex":"#e5e7eb","colorName":"WHITE","kind":"DATA","signal":"WIEGAND_D1",
	   "x":0.12,"y":0.82,"description":"داتا ويگند D1 — لقارئ خارجي","pairId":"wiegand"},
	  {"id":"t_gray","label":"رمادي","colorHex":"#6b7280","colorName":"GRAY","kind":"OUTPUT","signal":"BELL",
	   "x":0.12,"y":0.90,"description":"مخرج الجرس"}
	]`

	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, status, "sourceRef", "localPractice", verified)
		VALUES ($1, $2, 'نموذج تدريبي', 'AC-KEYPAD-15W',
		        'لوحة تحكم بالدخول — ١٥ سلك',
		        'لوحة كيباد مستقلة: تغذية ١٢ فولت، ريلاي فتح، زر خروج، حسّاس باب، وقارئ ويگند.',
		        'WIRING', $3, $4, $5, 'DRAFT', $6, $7, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		devID, catID,
		`{"power":{"voltage":"12VDC","currentA":0.5},"relay":{"maxA":2,"maxV":36}}`,
		terminals,
		`{"kind":"WIRING","hotspotRadius":0.022}`,
		"أعراف منشورة عامة للوحات التحكم بالدخول المستقلة — مو كتالوگ موديل بعينه. غير محقّق ميدانياً.",
		"عدنا بالعراق الشائع تغذية مستقلة للقفل عن اللوحة، والريلاي يقطع الموجب مو السالب.",
	); err != nil {
		return err
	}

	// مصدر تغذية بسيط بالمشهد.
	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, status, "sourceRef", verified)
		VALUES ($1, $2, 'عام', 'PSU-12V-3A', 'مصدر تغذية ١٢ فولت',
		        'محوّل ١٢ فولت مستمر ٣ أمبير.', 'WIRING', $3, $4, $5, 'DRAFT', $6, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		psuID, catID,
		`{"power":{"voltage":"12VDC","currentA":3}}`,
		`[
		  {"id":"out_pos","label":"موجب +","colorHex":"#dc2626","colorName":"RED","kind":"POWER_POS","signal":"+12V",
		   "x":0.80,"y":0.30,"description":"خرج الموجب ١٢ فولت"},
		  {"id":"out_neg","label":"سالب −","colorHex":"#111827","colorName":"BLACK","kind":"POWER_NEG","signal":"GND",
		   "x":0.80,"y":0.44,"description":"خرج السالب (الأرضي)"}
		]`,
		`{"kind":"WIRING","hotspotRadius":0.022}`,
		"جهاز مساعد بالمشهد — مواصفات عامة.",
	); err != nil {
		return err
	}

	// ═══ الدرس ═══
	if _, err := db.Exec(`
		INSERT INTO "SimLesson" (id, "categoryId", "deviceId", title, blocks, "sortOrder", status)
		VALUES ($1, $2, $3, 'ألوان أسلاك لوحة التحكم بالدخول', $4, 10, 'DRAFT')
		ON CONFLICT (id) DO NOTHING`,
		lesID, catID, devID, `[
		  {"t":"warn","md":"هذا المحتوى **غير محقّق ميدانياً**. ألوان الأسلاك تختلف بين مصنّع وآخر — ارجع لكتالوگ الجهاز الي بإيدك قبل أي توصيل حقيقي."},
		  {"t":"text","md":"لوحة التحكم بالدخول تاخذ تغذية ١٢ فولت مستمر، وتفتح القفل عن طريق **ريلاي** داخلها — يعني هي ما تغذّي القفل، بس تسكّر الدائرة له."},
		  {"t":"table","head":["اللون","الوظيفة","ملاحظة"],
		   "rows":[["أحمر","+12V","الموجب — عكسه يحرق اللوحة"],
		           ["أسود","GND","السالب المشترك"],
		           ["بنفسجي","COM","مشترك الريلاي"],
		           ["أزرق","NO","يسكّر لحظة الفتح — للأقفال الكهربائية"],
		           ["برتقالي","NC","مفتوح لحظة الفتح — للأقفال المغناطيسية"],
		           ["أصفر","زر الخروج","يفتح من داخل الغرفة"],
		           ["بني","حسّاس الباب","ينبّه إذا بقى الباب مفتوحاً"],
		           ["أخضر/أبيض","D0/D1","داتا ويگند لقارئ خارجي"],
		           ["رمادي","الجرس","مخرج جرس الباب"]]},
		  {"t":"warn","md":"**NO مو NC.** القفل الكهربائي العادي يحتاج NO (يشتغل لمن يوصله الكهرباء)، والمغناطيسي يحتاج NC (يفتح لمن تنقطع عنه). خلطهم يعني باب يبقى مفتوح دائماً أو مقفول دائماً."},
		  {"t":"quiz","q":"قفل مغناطيسي — أي تماس تستعمل؟","options":["NO (أزرق)","NC (برتقالي)"],"answer":1,
		   "why":"المغناطيسي يمسك الباب وهو مكهرب، فلازم ينقطع عنه التيار حتى يفتح — يعني NC."}
		]`); err != nil {
		return err
	}

	// ═══ التمرين ═══
	//
	// ⚠️ الخطوة الأولى فيها غلط **قاتل** مقصود: عكس القطبية. ما ينهي
	// المحاولة — يعرض شاشة «لو هذا كان بالميدان چان احترقت اللوحة»
	// ويخلّيه يعيد. هاي لحظة التعليم الحقيقية.
	if _, err := db.Exec(`
		INSERT INTO "SimExercise" (id, "categoryId", title, brief, "engineKind", difficulty,
		                           "passScore", scene, steps, status, "sourceRef", verified, "sortOrder")
		VALUES ($1, $2, 'توصيل لوحة التحكم بالدخول',
		        'وصّل اللوحة بالتغذية، وجهّز مخرج الريلاي للقفل الكهربائي، وركّب زر الخروج.',
		        'WIRING', 1, 80, $3, $4, 'DRAFT', $5, FALSE, 10)
		ON CONFLICT (id) DO NOTHING`,
		exID, catID,
		// ⚠️ القفل **يمين** والمغذّي **يسار**: أطراف القفل على حافّته
		// اليسرى وأطراف المغذّي على حافّته اليمنى، فبهالترتيب الأسلاك
		// تمشي بالفراغ بينهما بدل ما تلف حول الجهازين. وهذا يقرا صح
		// بواجهة عربية (الأساسي يمين).
		//
		// ⚠️⚠️ **الأرقام هنا مو بترحيل**: `runVersionedMigrations` تمشي
		// **قبل** البذور بـ`Migrate()`. فترحيل يعدّل صف تزرعه البذرة
		// بعده ما يلگي شي يعدّله على قاعدة جديدة — يمر بلا مفعول
		// وبلا غلط. ترحيل `0254` يصلّح القواعد **المزروعة سابقاً**
		// (مثل الإنتاج)، وهذي الأرقام تصلّح **الجديدة**. الاثنان لازمان.
		`{"devices":[{"ref":"lock1","deviceId":"`+devID+`","x":0.68,"y":0.50},
		             {"ref":"psu1","deviceId":"`+psuID+`","x":0.22,"y":0.40}]}`,
		`[
		  {"index":1,"title":"موجب التغذية","instruction":"وصّل موجب التغذية (+12V) للسلك الأحمر باللوحة.",
		   "expect":{"op":"CONNECT","from":"lock1:t_red","to":"psu1:out_pos"},
		   "hint":"الأحمر دائماً الموجب بهذا النوع من اللوحات.",
		   "wrong":[{"match":{"op":"CONNECT","from":"lock1:t_red","to":"psu1:out_neg"},
		             "say":"عكست القطبية — ربطت الأحمر (الموجب) بالسالب. لو هذا كان بالميدان چان احترقت اللوحة فوراً وما نفع الضمان.",
		             "penalty":25,"fatal":true},
		            {"matchAny":true,"say":"هذا مو الطرف الصحيح. راجع جدول الألوان بالدرس.","penalty":5}],
		   "weight":20,"hintPenalty":5,"wrongPenalty":5},

		  {"index":2,"title":"سالب التغذية","instruction":"وصّل سالب التغذية (GND) للسلك الأسود.",
		   "expect":{"op":"CONNECT","from":"lock1:t_black","to":"psu1:out_neg"},
		   "hint":"الأسود هو السالب المشترك.",
		   "wrong":[{"matchAny":true,"say":"السالب يروح للأسود بس.","penalty":5}],
		   "weight":20,"hintPenalty":5,"wrongPenalty":5},

		  {"index":3,"title":"مشترك الريلاي","instruction":"جهّز مخرج القفل: وصّل مشترك الريلاي (البنفسجي COM) بموجب التغذية.",
		   "expect":{"op":"CONNECT","from":"lock1:t_purple","to":"psu1:out_pos"},
		   "hint":"الـCOM يمرّ منه التيار للقفل، فياخذ الموجب.",
		   "wrong":[{"match":{"op":"CONNECT","from":"lock1:t_orange","to":"psu1:out_pos"},
		             "say":"هذا NC (البرتقالي) — يستعمل للأقفال المغناطيسية. إحنا نجهّز قفلاً كهربائياً عادياً.",
		             "penalty":10},
		            {"matchAny":true,"say":"دوّر على السلك البنفسجي — هو مشترك الريلاي.","penalty":5}],
		   "weight":30,"hintPenalty":5,"wrongPenalty":5},

		  {"index":4,"title":"زر الخروج","instruction":"وصّل زر الخروج: السلك الأصفر بسالب التغذية.",
		   "expect":{"op":"CONNECT","from":"lock1:t_yellow","to":"psu1:out_neg"},
		   "hint":"زر الخروج يشتغل بتقصير الطرف على الأرضي.",
		   "wrong":[{"matchAny":true,"say":"مدخل زر الخروج هو الأصفر، ويشتغل بالأرضي.","penalty":5}],
		   "weight":30,"hintPenalty":5,"wrongPenalty":5}
		]`,
		"تمرين تدريبي على أعراف عامة — غير محقّق على موديل بعينه.",
	); err != nil {
		return err
	}
	return nil
}
