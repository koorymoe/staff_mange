package database

import "github.com/jmoiron/sqlx"

// ═══ محتوى GPS: تركيب جهاز تتبّع بالسيارة ═══
//
// آخر خدمة من ثمان خدمات الشركة ما چان بيها ولا تمرين.
//
// ═══════════════════════════════════════════════════════════════
// ⚠️⚠️ اقرا هذا قبل أي شي ⚠️⚠️
//
// هذا المحتوى **غير محقّق** (`verified = FALSE`). مبني على الأعراف
// المنشورة الشائعة بأجهزة التتبّع — **مو على كتالوگ موديل بعينه**.
// ألوان الأسلاك تختلف بين مصنّع وآخر، وبعض الأجهزة تعكس وظيفة
// الأبيض والأزرق. ارجع لكتالوگ الجهاز الي بإيدك.
// ═══════════════════════════════════════════════════════════════
//
// ⚠️ ليش هالجهاز **بمحرّك التوصيل ثلاثي الأبعاد** مو بمساحة العمل؟
// لأنه جهاز بأسلاك ملوّنة تنربط بسيارة — مو توبولوجي فيه عقد
// ووصلات. فينعاد استعمال نفس المحرّك الي انبنى للقفل بلا سطر جديد:
// `sim3d/deviceGeometry.ts` و`Workbench3D.tsx` و`evaluate.ts`.
//
// ═══ لحظة التعليم ═══
//
// الغلط القاتل هنا **مو حرق جهاز** — هو أخبث من هيچ: **الأصفر
// (ACC) على الموجب الدائم**. الجهاز يشتغل، والأضوية تضوي، والتقارير
// توصل — فالفني يسلّم ويمشي. وبعد أيام:
//   • الجهاز يظن السيارة شغّالة على طول فيبقى يرسل بلا نوم
//   • بطارية السيارة تفرغ والزبون يرجع غاضباً
//   • وكل تقارير «ساعات التشغيل» و«التوقف» غلط من أساسها
// وهذا **أكثر غلط يصير عدنا** لأن الجهاز «يشتغل» — والفني ما عنده
// سبب يشك.

func seedSimGps(db *sqlx.DB) error {
	const (
		catID = "simcat_gps"
		devID = "simdev_gps_tracker"
		harID = "simdev_car_harness"
		exID  = "simex_gps_wiring"
		lesID = "simles_gps_colors"
	)

	if _, err := db.Exec(`
		INSERT INTO "SimCategory" (id, "serviceId", name, description, "sortOrder")
		SELECT $1,
		       (SELECT id FROM "Service" WHERE name ILIKE '%gps%' OR name ILIKE '%تتبع%' OR name ILIKE '%تتبّع%' LIMIT 1),
		       'أجهزة التتبّع GPS',
		       'تركيب أجهزة التتبّع بالسيارات: التغذية، الكونتاكت، قطع الوقود، والهوائيات.',
		       30
		ON CONFLICT (id) DO NOTHING`, catID); err != nil {
		return err
	}

	// ═══ الجهاز ═══
	//
	// ⚠️ `danger` على كل طرف خطر: يطلع بالمرور بالمشهد الثلاثي وبتفسير
	// الغلط. المتدرّب لازم يعرف **العاقبة** مو بس إنه غلط.
	trackerTerminals := `[
	  {"id":"g_red","label":"أحمر","colorHex":"#dc2626","colorName":"RED","kind":"POWER_POS","signal":"+12V دائم",
	   "x":0.12,"y":0.14,"description":"الموجب الدائم — من بطارية السيارة عبر فيوز",
	   "danger":"بلا فيوز: أي قصر بالسلك يحرق ضفيرة السيارة كلها، مو الجهاز بس"},
	  {"id":"g_black","label":"أسود","colorHex":"#111827","colorName":"BLACK","kind":"POWER_NEG","signal":"GND",
	   "x":0.12,"y":0.30,"description":"الأرضي — على هيكل السيارة أو سالب البطارية"},
	  {"id":"g_yellow","label":"أصفر","colorHex":"#eab308","colorName":"YELLOW","kind":"INPUT","signal":"ACC",
	   "x":0.12,"y":0.46,"description":"الكونتاكت — يجي جهد بس لمن يدير السائق مفتاح التشغيل",
	   "danger":"لو انربط بالموجب الدائم: الجهاز يظن السيارة شغّالة دائماً، يفرّغ البطارية، وكل تقارير التشغيل تطلع غلط"},
	  {"id":"g_white","label":"أبيض","colorHex":"#e5e7eb","colorName":"WHITE","kind":"RELAY_COM","signal":"ريلاي COM",
	   "x":0.12,"y":0.62,"description":"مشترك ريلاي قطع الوقود"},
	  {"id":"g_blue","label":"أزرق","colorHex":"#2563eb","colorName":"BLUE","kind":"RELAY_NO","signal":"ريلاي NO",
	   "x":0.12,"y":0.78,"description":"تماس مفتوح — ينسكّر لمن يأمر السيرفر بقطع الوقود",
	   "danger":"القطع لازم يكون على **مضخّة الوقود** مو على الكونتاكت — قطع الكونتاكت بسيارة ماشية يوگف الدركسون والفرامل"},
	  {"id":"g_gnss","label":"هوائي GNSS","colorHex":"#059669","colorName":"GREEN","kind":"ANTENNA","signal":"GNSS",
	   "x":0.12,"y":0.92,"description":"هوائي الموقع — وجهه للسما تحت التابلو",
	   "danger":"جوّا هيكل معدني ما يستقبل — الجهاز يشتغل بس بلا موقع"}
	]`

	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, geometry, status, "sourceRef", "localPractice", verified)
		VALUES ($1, $2, 'نموذج تدريبي', 'GT-06-TRAINING',
		        'جهاز تتبّع GPS',
		        'جهاز تتبّع بستة أطراف: تغذية دائمة، أرضي، كونتاكت، ريلاي قطع وقود، وهوائي.',
		        'WIRING', $3, $4, $5, $6, 'DRAFT', $7, $8, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		devID, catID,
		`{"power":{"voltage":"9-36VDC","currentA":0.06},"relay":{"maxA":10},"backup":{"mah":450}}`,
		trackerTerminals,
		`{"kind":"WIRING","hotspotRadius":0.022}`,
		`{"shape":"wall_box","sizeM":{"w":0.085,"h":0.055,"d":0.020},
		  "bodyColorHex":"#334155","faceColorHex":"#1e293b",
		  "terminalPost":{"radiusM":0.0028,"heightM":0.0045},
		  "features":[{"kind":"statusLed","x":0.62,"y":0.22,"channel":"status_led"},
		              {"kind":"statusLed","x":0.62,"y":0.5,"channel":"gsm_led"},
		              {"kind":"terminalPlate","x0":0.03,"y0":0.06,"x1":0.24,"y1":0.98}]}`,
		"أعراف منشورة شائعة لأجهزة التتبّع — مو كتالوگ موديل بعينه. ألوان الأسلاك تختلف بين مصنّع وآخر.",
		"عدنا بالعراق الشائع تركيب الجهاز تحت التابلو، والقطع على مضخّة الوقود مو على الكونتاكت.",
	); err != nil {
		return err
	}

	// ═══ ضفيرة السيارة ═══ الطرف الثاني للتوصيل.
	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, geometry, status, "sourceRef", verified)
		VALUES ($1, $2, 'عام', 'CAR-HARNESS', 'ضفيرة السيارة',
		        'نقاط الربط بالسيارة: بطارية مع فيوز، أرضي الهيكل، الكونتاكت، ومضخّة الوقود.',
		        'WIRING', $3, $4, $5, $6, 'DRAFT', $7, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		harID, catID,
		`{"system":"12V"}`,
		`[
		  {"id":"h_bat_fused","label":"بطارية + (بفيوز)","colorHex":"#dc2626","colorName":"RED","kind":"POWER_POS","signal":"+12V دائم",
		   "x":0.86,"y":0.14,"description":"موجب البطارية بعد فيوز ٢ أمبير — الطريق الصحيح للتغذية"},
		  {"id":"h_bat_raw","label":"بطارية + (بلا فيوز)","colorHex":"#f97316","colorName":"ORANGE","kind":"POWER_POS","signal":"+12V دائم",
		   "x":0.86,"y":0.30,"description":"موجب مباشر بلا حماية",
		   "danger":"بلا فيوز أي قصر يحرق ضفيرة السيارة كلها"},
		  {"id":"h_gnd","label":"أرضي الهيكل","colorHex":"#111827","colorName":"BLACK","kind":"POWER_NEG","signal":"GND",
		   "x":0.86,"y":0.46,"description":"برغي أرضي على هيكل السيارة"},
		  {"id":"h_acc","label":"الكونتاكت ACC","colorHex":"#eab308","colorName":"YELLOW","kind":"INPUT","signal":"ACC",
		   "x":0.86,"y":0.62,"description":"يجي جهد بس لمن يدير السائق المفتاح"},
		  {"id":"h_fuel","label":"مضخّة الوقود","colorHex":"#2563eb","colorName":"BLUE","kind":"RELAY_NO","signal":"FUEL",
		   "x":0.86,"y":0.78,"description":"خط مضخّة الوقود — محل القطع الصحيح"},
		  {"id":"h_sky","label":"تحت التابلو (سما مكشوفة)","colorHex":"#059669","colorName":"GREEN","kind":"ANTENNA","signal":"SKY",
		   "x":0.86,"y":0.94,"description":"محل تركيب الهوائي — بلا معدن فوگه"}
		]`,
		`{"kind":"WIRING","hotspotRadius":0.022}`,
		`{"shape":"psu_brick","sizeM":{"w":0.120,"h":0.090,"d":0.035},
		  "bodyColorHex":"#4b5563","faceColorHex":"#2b3242",
		  "terminalPost":{"radiusM":0.0028,"heightM":0.0045},
		  "features":[{"kind":"terminalPlate","x0":0.72,"y0":0.06,"x1":0.97,"y1":0.98}]}`,
		"نقاط ربط عامة بالسيارة — تختلف بين موديل وآخر.",
	); err != nil {
		return err
	}

	// ═══ الدرس ═══
	if _, err := db.Exec(`
		INSERT INTO "SimLesson" (id, "categoryId", "deviceId", title, blocks, "sortOrder", status)
		VALUES ($1, $2, $3, 'أسلاك جهاز التتبّع وأخطاء التركيب', $4, 10, 'DRAFT')
		ON CONFLICT (id) DO NOTHING`,
		lesID, catID, devID, `[
		  {"t":"warn","md":"هذا المحتوى **غير محقّق ميدانياً**. ألوان الأسلاك تختلف بين مصنّع وآخر، وبعض الأجهزة تعكس وظيفة الأبيض والأزرق — ارجع لكتالوگ الجهاز الي بإيدك."},
		  {"t":"text","md":"جهاز التتبّع يحتاج **ثلاثة أشياء** حتى يشتغل صح: تغذية دائمة (حتى يبقى شغّالاً والسيارة مطفية)، أرضي، و**إشارة كونتاكت** حتى يعرف متى السيارة شغّالة ومتى مطفية."},
		  {"t":"table","head":["اللون","الوظيفة","محل الربط الصحيح"],
		   "rows":[["أحمر","+12V دائم","موجب البطارية **بعد فيوز**"],
		           ["أسود","GND","برغي أرضي على الهيكل"],
		           ["أصفر","ACC (الكونتاكت)","خط يجيه جهد بس لمن يدير المفتاح"],
		           ["أبيض","ريلاي COM","مشترك ريلاي القطع"],
		           ["أزرق","ريلاي NO","**مضخّة الوقود** — مو الكونتاكت"],
		           ["أخضر","هوائي GNSS","تحت التابلو ووجهه للسما"]]},
		  {"t":"warn","md":"**الأصفر مو أحمر ثانٍ.** ربط الـACC بالموجب الدائم يخلّي الجهاز يظن السيارة شغّالة على طول: يبقى يرسل بلا نوم، يفرّغ بطارية السيارة بأيام، وكل تقارير ساعات التشغيل والتوقف تطلع غلط. والأخبث إنه **يشتغل** — فالفني يسلّم ويمشي وما عنده سبب يشك."},
		  {"t":"warn","md":"**القطع على مضخّة الوقود مو على الكونتاكت.** قطع الكونتاكت بسيارة ماشية يطفّي المحرك فيوگف الدركسون الهيدروليكي والفرامل المساعدة — هذا خطر على الأرواح مو عطل."},
		  {"t":"quiz","q":"وين تربط السلك الأصفر (ACC)؟","options":["على موجب البطارية الدائم","على خط يجيه جهد لمن يدير المفتاح بس"],"answer":1,
		   "why":"الـACC وظيفته يخبر الجهاز متى السيارة شغّالة. على الموجب الدائم يصير دائماً «شغّالة» — والجهاز ما ينام والبطارية تفرغ."},
		  {"t":"quiz","q":"ليش الفيوز على السلك الأحمر؟","options":["حماية الجهاز","حماية ضفيرة السيارة"],"answer":1,
		   "why":"الجهاز يسحب ملّي أمبيرات. الفيوز يحمي **السلك** — أي قصر بينه وبين البطارية يمرّر عشرات الأمبيرات ويحرق الضفيرة."}
		]`); err != nil {
		return err
	}

	// ═══ التمرين ═══
	//
	// ⚠️ ترتيب المشهد: الجهاز **يمين** والضفيرة **يسار** — أطراف
	// الجهاز على حافّته اليسرى وأطراف الضفيرة على حافّتها اليمنى،
	// فالأسلاك تمشي بالفراغ بينهما. نفس درس القفل.
	if _, err := db.Exec(`
		INSERT INTO "SimExercise" (id, "categoryId", title, brief, "engineKind", difficulty,
		                           "passScore", scene, steps, status, "sourceRef", verified, "sortOrder")
		VALUES ($1, $2, 'تركيب جهاز تتبّع بالسيارة',
		        'وصّل الجهاز بالتغذية والأرضي، وميّز خط الكونتاكت، وجهّز قطع الوقود، وركّب الهوائي.',
		        'WIRING', 2, 80, $3, $4, 'DRAFT', $5, FALSE, 40)
		ON CONFLICT (id) DO NOTHING`,
		exID, catID,
		`{"devices":[{"ref":"gps1","deviceId":"`+devID+`","x":0.70,"y":0.50},
		             {"ref":"car1","deviceId":"`+harID+`","x":0.24,"y":0.50}]}`,
		`[
		  {"index":1,"title":"التغذية الدائمة","instruction":"وصّل الأحمر (+12V دائم) بموجب البطارية **المحمي بفيوز**.",
		   "expect":{"op":"CONNECT","from":"gps1:g_red","to":"car1:h_bat_fused"},
		   "hint":"اكو طرفان للبطارية — واحد بفيوز وواحد بلا. الفيوز يحمي السلك مو الجهاز.",
		   "wrong":[{"match":{"op":"CONNECT","from":"gps1:g_red","to":"car1:h_bat_raw"},
		             "say":"⚠️ ربطت على موجب **بلا فيوز**. الجهاز يسحب ملّي أمبيرات فما راح يحترق — بس أي قصر بالسلك بينه وبين البطارية يمرّر عشرات الأمبيرات ويحرق **ضفيرة السيارة كلها**. الفيوز يحمي السلك مو الجهاز.",
		             "penalty":15},
		            {"match":{"op":"CONNECT","from":"gps1:g_red","to":"car1:h_gnd"},
		             "say":"⚠️ عكست القطبية — الأحمر على الأرضي. أغلب الأجهزة عدها حماية عكس، بس مو كلها، والي ما عنده يحترق فوراً.",
		             "penalty":20,"fatal":true},
		            {"matchAny":true,"say":"الأحمر يروح لموجب البطارية بعد الفيوز.","penalty":5}],
		   "weight":20,"hintPenalty":5,"wrongPenalty":5},

		  {"index":2,"title":"الأرضي","instruction":"وصّل الأسود بأرضي هيكل السيارة.",
		   "expect":{"op":"CONNECT","from":"gps1:g_black","to":"car1:h_gnd"},
		   "hint":"برغي أرضي نظيف على الهيكل — مو على قطعة مدهونة.",
		   "wrong":[{"matchAny":true,"say":"الأسود هو الأرضي، ويروح لبرغي الهيكل.","penalty":5}],
		   "weight":15,"hintPenalty":5,"wrongPenalty":5},

		  {"index":3,"title":"إشارة الكونتاكت","instruction":"وصّل الأصفر (ACC) بخط الكونتاكت — الخط الي يجيه جهد بس لمن يدير السائق المفتاح.",
		   "expect":{"op":"CONNECT","from":"gps1:g_yellow","to":"car1:h_acc"},
		   "hint":"جرّب الخط بأڤوميتر: لازم يعطي صفراً والمفتاح مطفي، و١٢ فولت لمن تديره.",
		   "wrong":[{"match":{"op":"CONNECT","from":"gps1:g_yellow","to":"car1:h_bat_fused"},
		             "say":"🔥 هذا **أخبث غلط بتركيب أجهزة التتبّع**. ربطت الأصفر (ACC) بالموجب الدائم — والجهاز راح **يشتغل**، والأضوية تضوي، والتقارير توصل، فتسلّم وتمشي. وبعد أيام: الجهاز يظن السيارة شغّالة على طول فما ينام أبداً، بطارية السيارة تفرغ، وكل تقارير ساعات التشغيل والتوقف غلط من أساسها. والزبون يرجع غاضباً وأنت ما عندك سبب تشك بالتوصيل لأنه «چان شغّال».",
		             "penalty":25,"fatal":true},
		            {"match":{"op":"CONNECT","from":"gps1:g_yellow","to":"car1:h_bat_raw"},
		             "say":"🔥 نفس الغلط القاتل — موجب دائم بلا فيوز هم. الأصفر لازم يشوف الجهد **بس لمن يدير المفتاح**.",
		             "penalty":25,"fatal":true},
		            {"matchAny":true,"say":"الأصفر يروح لخط الكونتاكت — الي يتغيّر جهده مع مفتاح التشغيل.","penalty":5}],
		   "weight":30,"hintPenalty":5,"wrongPenalty":8},

		  {"index":4,"title":"قطع الوقود","instruction":"جهّز القطع: وصّل الأزرق (ريلاي NO) بخط **مضخّة الوقود**.",
		   "expect":{"op":"CONNECT","from":"gps1:g_blue","to":"car1:h_fuel"},
		   "hint":"القطع يصير على المضخّة — الكونتاكت أخطر بكثير.",
		   "wrong":[{"match":{"op":"CONNECT","from":"gps1:g_blue","to":"car1:h_acc"},
		             "say":"🚨 **خطر على الأرواح.** ربطت القطع على الكونتاكت — قطعه بسيارة ماشية يطفّي المحرك، فيوگف الدركسون الهيدروليكي وتضعف الفرامل المساعدة، والسائق ما يگدر يسيطر. القطع يصير على **مضخّة الوقود**: السيارة تفصل تدريجياً وتبقى مسيطراً عليها.",
		             "penalty":30,"fatal":true},
		            {"matchAny":true,"say":"الأزرق (NO) يروح لخط مضخّة الوقود.","penalty":5}],
		   "weight":20,"hintPenalty":5,"wrongPenalty":8},

		  {"index":5,"title":"الهوائي","instruction":"ركّب هوائي GNSS تحت التابلو بمحل سماه مكشوفة.",
		   "expect":{"op":"CONNECT","from":"gps1:g_gnss","to":"car1:h_sky"},
		   "hint":"وجه الهوائي للسما، وبلا معدن فوگه.",
		   "wrong":[{"matchAny":true,"say":"الهوائي يحتاج سما مكشوفة — جوّا هيكل معدني الجهاز يشتغل بس بلا موقع.","penalty":5}],
		   "weight":15,"hintPenalty":5,"wrongPenalty":5}
		]`,
		"أعراف تركيب منشورة شائعة — غير محقّقة على موديل بعينه.",
	); err != nil {
		return err
	}

	return nil
}
