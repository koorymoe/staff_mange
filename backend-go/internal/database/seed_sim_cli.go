package database

import "github.com/jmoiron/sqlx"

// ═══ محتوى الـCLI: سويچ تدريبي ═══
//
// «محاكي للشبكات سويچات بأنواعهن هواوي تيبي لينك سيسكو الراوترات
// شلون يتبرمجن».
//
// ═══════════════════════════════════════════════════════════════
// ⚠️⚠️ اقرا هذا قبل أي شي ⚠️⚠️
//
// هذا **مو IOS مالت سيسكو ولا نسخة منه ولا صورة مالتهم**. هذا نحو
// تدريبي مكتوب من الأوامر الشائعة المنشورة الي أي منهج شبكات يعلّمها.
// صور أنظمة المصنّعين مرخّصة وما تنشغّل برّا شروطها — والمخطط الرئيسي
// يحذّر من هذا صراحةً (٤٠) ويكول «لا نبني IOS clone» (٤٢).
//
// الهدف: الفني يتعوّد على **الأسلوب** — الأنماط والاختصار و`?` وشكل
// الأخطاء. وأي فرق بينه وبين الجهاز الحقيقي يبقى فرقاً، ولهذا
// `verified = FALSE`: ما يوصل متدرّباً حتى يجرّبه فني على سويچ حقيقي.
// ═══════════════════════════════════════════════════════════════
//
// ⚠️ الشجرة أدناه **مولّدة** من `frontend/src/cli/ciscoLike.ts` — هي
// مصدر التأليف الوحيد. لا تعدّلها هنا بالإيد: التعديل يضيع بأول
// توليد، والنسختان تفترقان بصمت.

func seedSimCli(db *sqlx.DB) error {
	const (
		catID  = "simcat_networking"
		gramID = "simcli_ios_like_switch"
		devID  = "simdev_switch_24p"
		exID   = "simex_vlan_access_port"
	)

	if _, err := db.Exec(`
		INSERT INTO "SimCategory" (id, "serviceId", name, description, "sortOrder")
		SELECT $1,
		       (SELECT id FROM "Service" WHERE name ILIKE '%شبك%' LIMIT 1),
		       'الشبكات — السويچات والراوترات',
		       'تهيئة السويچات بسطر الأوامر: الأنماط، الـVLAN، منافذ الوصول والترنك.',
		       20
		ON CONFLICT (id) DO NOTHING`, catID); err != nil {
		return err
	}

	// ═══ النحو ═══
	if _, err := db.Exec(`
		INSERT INTO "SimCliGrammar" (id, name, os, tree, status, "sourceRef")
		VALUES ($1, 'Training Switch OS', '15.2-TRAINING', $2, 'DRAFT', $3)
		ON CONFLICT (id) DO NOTHING`,
		gramID,
		cliGrammarIOSLike,
		"أوامر شائعة منشورة بمناهج الشبكات — مو كتالوگ ولا صورة مصنّع. غير محقّق ميدانياً.",
	); err != nil {
		return err
	}

	// ═══ الجهاز ═══ سويچ ٢٤ منفذ. الأطراف هنا **منافذ شبكة** مو براغي،
	// فمحرّك التوصيل يرفض سنّها ببلوك براغي لحاله (توافق الموصّلات).
	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, geometry, status, "sourceRef", "localPractice", verified)
		VALUES ($1, $2, 'نموذج تدريبي', 'SW-24P-TRAINING',
		        'سويچ إدارة ٢٤ منفذ',
		        'سويچ طبقة ٢ قابل للإدارة: تهيئة بسطر الأوامر، VLAN، منافذ وصول وترنك.',
		        'CLI', $3, $4, $5, $6, 'DRAFT', $7, $8, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		devID, catID,
		`{"ports":24,"uplinks":2,"layer":2,"cliGrammarId":"`+gramID+`"}`,
		`[]`,
		`{"kind":"CLI","grammarId":"`+gramID+`"}`,
		`{"shape":"rack_1u","sizeM":{"w":0.44,"h":0.044,"d":0.20},"bodyColorHex":"#2b3440","faceColorHex":"#1b222c"}`,
		"أوامر شائعة منشورة — مو كتالوگ موديل بعينه.",
		"عدنا بالعراق الغالب سويچات مدارة بمشاريع الشركات، والمنازل بسويچات غير مدارة.",
	); err != nil {
		return err
	}

	// ═══ التمرين ═══
	//
	// ⚠️ التقييم على **الحالة** مو على الحروف المكتوبة: الفني ممكن
	// يوصل لنفس النتيجة بمسارات مختلفة (`vlan 10` قبل أو بعد تهيئة
	// المنفذ)، وأي تقييم يطابق نص الأمر يعاقبه على طريق صحيح.
	if _, err := db.Exec(`
		INSERT INTO "SimExercise" (id, "categoryId", title, brief, "engineKind", difficulty,
		                           "passScore", scene, steps, status, "sourceRef", verified, "sortOrder")
		VALUES ($1, $2, 'تهيئة VLAN ومنفذ وصول',
		        'أنشئ VLAN وسمّه، وحوّل منفذاً لوضع الوصول واربطه بالـVLAN — كله بسطر الأوامر.',
		        'CLI', 2, 80, $3, $4, 'DRAFT', $5, FALSE, 20)
		ON CONFLICT (id) DO NOTHING`,
		exID, catID,
		`{"devices":[{"ref":"sw1","deviceId":"`+devID+`","x":0.50,"y":0.50}],"cliDeviceRef":"sw1"}`,
		cliStepsVlan,
		"أوامر شائعة منشورة بمناهج الشبكات. غير محقّق ميدانياً.",
	); err != nil {
		return err
	}

	// ═══ نحو VRP وجهازه ═══
	//
	// ⚠️ هذا **الإثبات** إن المعمار بيانات مو كود: نظام ثاني كامل
	// بأنماطه ومحثه المحيط وصيغة عرضه — ومحرّك الأوامر ما انلمس.
	if _, err := db.Exec(`
		INSERT INTO "SimCliGrammar" (id, name, os, tree, status, "sourceRef")
		VALUES ($1, 'Training VRP-style OS', 'V200R-TRAINING', $2, 'DRAFT', $3)
		ON CONFLICT (id) DO NOTHING`,
		"simcli_vrp_like_switch",
		cliGrammarVrpLike,
		"أوامر شائعة منشورة بمناهج الشبكات — مو كتالوگ ولا صورة مصنّع. غير محقّق ميدانياً.",
	); err != nil {
		return err
	}

	if _, err := db.Exec(`
		INSERT INTO "SimDevice" (id, "categoryId", brand, model, name, summary, "engineKind",
		                         spec, terminals, ui, geometry, status, "sourceRef", "localPractice", verified)
		VALUES ($1, $2, 'نموذج تدريبي', 'SW-24P-VRP',
		        'سويچ إدارة ٢٤ منفذ — نمط VRP',
		        'نفس السويچ بنظام تشغيل ثاني: system-view وport link-type وdisplay.',
		        'CLI', $3, $4, $5, $6, 'DRAFT', $7, $8, FALSE)
		ON CONFLICT (id) DO NOTHING`,
		"simdev_switch_24p_vrp", catID,
		`{"ports":24,"uplinks":2,"layer":2,"cliGrammarId":"simcli_vrp_like_switch"}`,
		`[]`,
		`{"kind":"CLI","grammarId":"simcli_vrp_like_switch"}`,
		`{"shape":"rack_1u","sizeM":{"w":0.44,"h":0.044,"d":0.20},"bodyColorHex":"#2b3440","faceColorHex":"#1b222c"}`,
		"أوامر شائعة منشورة — مو كتالوگ موديل بعينه.",
		"عدنا بالعراق هواوي شائعة بمشاريع المقاولات والحكومة، وسيسكو بالبنوك والشركات.",
	); err != nil {
		return err
	}

	// ⚠️ نفس هدف التمرين السابق بالضبط، ونظام ثاني: الفني الي يتعلّم
	// **المفهوم** ينقل بين الأنظمة، والي يحفظ **الأوامر** ما ينقل.
	if _, err := db.Exec(`
		INSERT INTO "SimExercise" (id, "categoryId", title, brief, "engineKind", difficulty,
		                           "passScore", scene, steps, status, "sourceRef", verified, "sortOrder")
		VALUES ($1, $2, 'تهيئة VLAN ومنفذ وصول — نمط VRP',
		        'نفس الهدف بنظام ثاني: system-view وport link-type access وport default vlan.',
		        'CLI', 3, 80, $3, $4, 'DRAFT', $5, FALSE, 21)
		ON CONFLICT (id) DO NOTHING`,
		"simex_vlan_access_port_vrp", catID,
		`{"devices":[{"ref":"sw1","deviceId":"simdev_switch_24p_vrp","x":0.50,"y":0.50}],"cliDeviceRef":"sw1"}`,
		cliStepsVlanVrp,
		"أوامر شائعة منشورة بمناهج الشبكات. غير محقّق ميدانياً.",
	); err != nil {
		return err
	}

	return nil
}

// ⚠️ مولّد — شوف التعليق فوگ.
const cliGrammarIOSLike = `{
 "id": "simcli_ios_like_switch",
 "name": "Training Switch OS",
 "os": "15.2-TRAINING",
 "startMode": "exec",
 "banner": [
  "  ⚠️  نظام تدريبي — مو صورة مصنّع. الأوامر شائعة ومنشورة.",
  "",
  "User Access Verification",
  ""
 ],
 "modes": [
  {
   "id": "exec",
   "promptSuffix": ">",
   "root": [
    {
     "t": "enable",
     "help": "Turn on privileged commands",
     "enter": {
      "mode": "priv"
     }
    },
    {
     "t": "show",
     "help": "Show running system information",
     "children": [
      {
       "t": "version",
       "help": "System hardware and software status",
       "show": "version"
      },
      {
       "t": "clock",
       "help": "Display the system clock",
       "say": "*04:12:37.115 UTC Mon Mar 1 1993"
      }
     ]
    },
    {
     "t": "exit",
     "help": "Exit from the EXEC",
     "exit": true
    },
    {
     "t": "?",
     "help": "List available commands",
     "say": ""
    }
   ]
  },
  {
   "id": "priv",
   "promptSuffix": "#",
   "root": [
    {
     "t": "configure",
     "help": "Enter configuration mode",
     "children": [
      {
       "t": "terminal",
       "help": "Configure from the terminal",
       "enter": {
        "mode": "config"
       }
      }
     ]
    },
    {
     "t": "show",
     "help": "Show running system information",
     "children": [
      {
       "t": "running-config",
       "help": "Current operating configuration",
       "show": "running-config"
      },
      {
       "t": "version",
       "help": "System hardware and software status",
       "show": "version"
      },
      {
       "t": "vlan",
       "help": "VTP VLAN status",
       "children": [
        {
         "t": "brief",
         "help": "VTP all VLAN status in brief",
         "show": "vlan-brief"
        }
       ]
      },
      {
       "t": "interfaces",
       "help": "Interface status and configuration",
       "children": [
        {
         "t": "status",
         "help": "Interface line status",
         "show": "interfaces-status"
        }
       ]
      }
     ]
    },
    {
     "t": "write",
     "help": "Write running configuration to memory",
     "children": [
      {
       "t": "memory",
       "help": "Write to NV memory",
       "say": "Building configuration...\n[OK]"
      }
     ]
    },
    {
     "t": "disable",
     "help": "Turn off privileged commands",
     "exit": true
    },
    {
     "t": "exit",
     "help": "Exit from the EXEC",
     "exit": true
    }
   ]
  },
  {
   "id": "config",
   "promptSuffix": "(config)#",
   "root": [
    {
     "t": "hostname",
     "help": "Set system's network name",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "This system's network name",
       "set": "hostname",
       "val": "$1"
      }
     ]
    },
    {
     "t": "vlan",
     "help": "VLAN commands",
     "children": [
      {
       "t": "<arg>",
       "arg": "num",
       "help": "ISL VLAN IDs 1-4094",
       "set": "vlans.$1.exists",
       "val": "true",
       "enter": {
        "mode": "config-vlan",
        "ctx": "$1"
       }
      }
     ]
    },
    {
     "t": "interface",
     "help": "Select an interface to configure",
     "children": [
      {
       "t": "<arg>",
       "arg": "ifname",
       "help": "Interface name, e.g. GigabitEthernet0/1",
       "set": "interfaces.$1.exists",
       "val": "true",
       "enter": {
        "mode": "config-if",
        "ctx": "$1"
       }
      }
     ]
    },
    {
     "t": "exit",
     "help": "Exit from configure mode",
     "exit": true
    },
    {
     "t": "end",
     "help": "Exit to privileged EXEC mode",
     "endTo": "priv"
    }
   ]
  },
  {
   "id": "config-vlan",
   "promptSuffix": "(config-vlan)#",
   "root": [
    {
     "t": "name",
     "help": "Ascii name of the VLAN",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "The ascii name for the VLAN",
       "set": "vlans.$ctx.name",
       "val": "$1"
      }
     ]
    },
    {
     "t": "exit",
     "help": "Exit from VLAN configuration",
     "exit": true
    },
    {
     "t": "end",
     "help": "Exit to privileged EXEC mode",
     "endTo": "priv"
    }
   ]
  },
  {
   "id": "config-if",
   "promptSuffix": "(config-if)#",
   "root": [
    {
     "t": "switchport",
     "help": "Set switching mode characteristics",
     "children": [
      {
       "t": "mode",
       "help": "Set trunking mode of the interface",
       "children": [
        {
         "t": "access",
         "help": "Set trunking mode to ACCESS unconditionally",
         "set": "interfaces.$ctx.mode",
         "val": "access"
        },
        {
         "t": "trunk",
         "help": "Set trunking mode to TRUNK unconditionally",
         "set": "interfaces.$ctx.mode",
         "val": "trunk"
        }
       ]
      },
      {
       "t": "access",
       "help": "Set access mode characteristics of the interface",
       "children": [
        {
         "t": "vlan",
         "help": "Set VLAN when interface is in access mode",
         "children": [
          {
           "t": "<arg>",
           "arg": "num",
           "help": "VLAN ID of the VLAN when this port is in access mode",
           "set": "interfaces.$ctx.accessVlan",
           "val": "$1"
          }
         ]
        }
       ]
      },
      {
       "t": "trunk",
       "help": "Set trunking characteristics of the interface",
       "children": [
        {
         "t": "allowed",
         "help": "Set allowed VLAN characteristics",
         "children": [
          {
           "t": "vlan",
           "help": "Set allowed VLANs on the trunk",
           "children": [
            {
             "t": "<arg>",
             "arg": "vlanlist",
             "help": "VLAN IDs of the allowed VLANs",
             "set": "interfaces.$ctx.trunkVlans",
             "val": "$1"
            }
           ]
          }
         ]
        }
       ]
      }
     ]
    },
    {
     "t": "description",
     "help": "Interface specific description",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "Up to 200 characters describing this interface",
       "set": "interfaces.$ctx.description",
       "val": "$1"
      }
     ]
    },
    {
     "t": "shutdown",
     "help": "Shutdown the selected interface",
     "set": "interfaces.$ctx.shutdown",
     "val": "true"
    },
    {
     "t": "exit",
     "help": "Exit from interface configuration mode",
     "exit": true
    },
    {
     "t": "end",
     "help": "Exit to privileged EXEC mode",
     "endTo": "priv"
    }
   ]
  }
 ]
}`

const cliStepsVlan = `[
  {"index":1,"title":"النمط المميّز","instruction":"ادخل النمط المميّز (privileged EXEC). المحث لازم يصير #.",
   "expect":{"op":"STATE_EQ","path":"__mode","value":"priv"},
   "hint":"الأمر enable — وتقدر تختصره en.","weight":10},

  {"index":2,"title":"اسم الجهاز","instruction":"ادخل نمط التهيئة وغيّر اسم الجهاز إلى SW-ALAMANI.",
   "expect":{"op":"STATE_EQ","path":"hostname","value":"SW-ALAMANI"},
   "hint":"configure terminal ثم hostname متبوعاً بالاسم.","weight":20},

  {"index":3,"title":"إنشاء الـVLAN","instruction":"أنشئ VLAN رقم 10 وسمّه IT-DEPT.",
   "expect":{"op":"STATE_EQ","path":"vlans.10.name","value":"IT-DEPT"},
   "hint":"vlan 10 يدخّلك نمط الـVLAN، وبعدها name IT-DEPT.","weight":30,
   "wrong":[{"matchAny":true,"say":"الـVLAN ينكتب رقمه أول ثم اسمه بنمط (config-vlan). لو كتبت الاسم بنمط (config) الجهاز ما يعرفه.","penalty":3}]},

  {"index":4,"title":"منفذ الوصول","instruction":"حوّل المنفذ GigabitEthernet0/1 لوضع access واربطه بـVLAN 10.",
   "expect":{"op":"STATE_EQ","path":"interfaces.GigabitEthernet0/1.accessVlan","value":"10"},
   "hint":"interface gi0/1 ثم switchport mode access ثم switchport access vlan 10.","weight":40,
   "wrong":[{"matchAny":true,"say":"⚠️ ربط الـVLAN بمنفذ بلا ما تحوّله access يشتغل بالمحاكي بس بالميدان الجهاز يتجاهله لمن يشتغل الـtrunk التلقائي. حوّله access أول.","penalty":5}]}
]`

// ⚠️ مولّد من `frontend/src/cli/huaweiVrp.ts` — لا تعدّله بالإيد.
const cliGrammarVrpLike = `{
 "id": "simcli_vrp_like_switch",
 "name": "Training VRP-style OS",
 "os": "V200R-TRAINING",
 "startMode": "user",
 "showStyle": "vrp",
 "banner": [
  "  ⚠️  نظام تدريبي — مو صورة مصنّع. الأوامر شائعة ومنشورة.",
  "",
  "Info: The max number of VTY users is 10.",
  ""
 ],
 "modes": [
  {
   "id": "user",
   "promptSuffix": ">",
   "promptTemplate": "<$host>",
   "root": [
    {
     "t": "system-view",
     "help": "Enter system view",
     "enter": {
      "mode": "system"
     }
    },
    {
     "t": "display",
     "help": "Display information",
     "children": [
      {
       "t": "version",
       "help": "Display version information",
       "show": "version"
      },
      {
       "t": "current-configuration",
       "help": "Display current configuration",
       "show": "running-config"
      },
      {
       "t": "vlan",
       "help": "Display VLAN information",
       "children": [
        {
         "t": "brief",
         "help": "Brief VLAN information",
         "show": "vlan-brief"
        }
       ]
      }
     ]
    },
    {
     "t": "quit",
     "help": "Exit from current mode",
     "exit": true
    }
   ]
  },
  {
   "id": "system",
   "promptSuffix": "]",
   "promptTemplate": "[$host]",
   "root": [
    {
     "t": "sysname",
     "help": "Set the host name",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "Host name",
       "set": "hostname",
       "val": "$1"
      }
     ]
    },
    {
     "t": "vlan",
     "help": "Create VLAN or enter VLAN view",
     "children": [
      {
       "t": "<arg>",
       "arg": "num",
       "help": "VLAN ID <1-4094>",
       "set": "vlans.$1.exists",
       "val": "true",
       "enter": {
        "mode": "vlan-view",
        "ctx": "$1"
       }
      }
     ]
    },
    {
     "t": "interface",
     "help": "Enter interface view",
     "children": [
      {
       "t": "<arg>",
       "arg": "ifname",
       "help": "Interface name, e.g. GigabitEthernet0/0/1",
       "set": "interfaces.$1.exists",
       "val": "true",
       "enter": {
        "mode": "if-view",
        "ctx": "$1"
       }
      }
     ]
    },
    {
     "t": "display",
     "help": "Display information",
     "children": [
      {
       "t": "current-configuration",
       "help": "Display current configuration",
       "show": "running-config"
      },
      {
       "t": "this",
       "help": "Display configuration of current view",
       "show": "running-config"
      },
      {
       "t": "vlan",
       "help": "Display VLAN information",
       "children": [
        {
         "t": "brief",
         "help": "Brief VLAN information",
         "show": "vlan-brief"
        }
       ]
      }
     ]
    },
    {
     "t": "save",
     "help": "Save current configuration",
     "say": "Are you sure to continue? [Y/N]:y\nIt will take several minutes to save configuration file, please wait...\nConfiguration file had been saved successfully"
    },
    {
     "t": "quit",
     "help": "Return to user view",
     "exit": true
    },
    {
     "t": "return",
     "help": "Return to user view",
     "endTo": "user"
    }
   ]
  },
  {
   "id": "vlan-view",
   "promptSuffix": "]",
   "promptTemplate": "[$host-vlan$ctx]",
   "root": [
    {
     "t": "description",
     "help": "Specify VLAN description",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "Description",
       "set": "vlans.$ctx.name",
       "val": "$1"
      }
     ]
    },
    {
     "t": "name",
     "help": "Specify VLAN name",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "VLAN name",
       "set": "vlans.$ctx.name",
       "val": "$1"
      }
     ]
    },
    {
     "t": "quit",
     "help": "Return to system view",
     "exit": true
    },
    {
     "t": "return",
     "help": "Return to user view",
     "endTo": "user"
    }
   ]
  },
  {
   "id": "if-view",
   "promptSuffix": "]",
   "promptTemplate": "[$host-$ctx]",
   "root": [
    {
     "t": "port",
     "help": "Port configuration",
     "children": [
      {
       "t": "link-type",
       "help": "Set the link type of the port",
       "children": [
        {
         "t": "access",
         "help": "Access port",
         "set": "interfaces.$ctx.mode",
         "val": "access"
        },
        {
         "t": "trunk",
         "help": "Trunk port",
         "set": "interfaces.$ctx.mode",
         "val": "trunk"
        },
        {
         "t": "hybrid",
         "help": "Hybrid port",
         "set": "interfaces.$ctx.mode",
         "val": "hybrid"
        }
       ]
      },
      {
       "t": "default",
       "help": "Set the default VLAN of the port",
       "children": [
        {
         "t": "vlan",
         "help": "Default VLAN of the access port",
         "children": [
          {
           "t": "<arg>",
           "arg": "num",
           "help": "VLAN ID <1-4094>",
           "set": "interfaces.$ctx.accessVlan",
           "val": "$1"
          }
         ]
        }
       ]
      },
      {
       "t": "trunk",
       "help": "Trunk port configuration",
       "children": [
        {
         "t": "allow-pass",
         "help": "Allow VLANs to pass",
         "children": [
          {
           "t": "vlan",
           "help": "VLANs allowed on the trunk",
           "children": [
            {
             "t": "<arg>",
             "arg": "vlanlist",
             "help": "VLAN IDs",
             "set": "interfaces.$ctx.trunkVlans",
             "val": "$1"
            }
           ]
          }
         ]
        }
       ]
      }
     ]
    },
    {
     "t": "description",
     "help": "Specify the description of the interface",
     "children": [
      {
       "t": "<arg>",
       "arg": "word",
       "help": "Description",
       "set": "interfaces.$ctx.description",
       "val": "$1"
      }
     ]
    },
    {
     "t": "shutdown",
     "help": "Shut down the interface",
     "set": "interfaces.$ctx.shutdown",
     "val": "true"
    },
    {
     "t": "display",
     "help": "Display information",
     "children": [
      {
       "t": "this",
       "help": "Configuration of current view",
       "show": "running-config"
      }
     ]
    },
    {
     "t": "quit",
     "help": "Return to system view",
     "exit": true
    },
    {
     "t": "return",
     "help": "Return to user view",
     "endTo": "user"
    }
   ]
  }
 ]
}`

const cliStepsVlanVrp = `[
  {"index":1,"title":"نمط النظام","instruction":"ادخل نمط النظام (system view). المحث لازم يصير بين قوسين مربّعين.",
   "expect":{"op":"STATE_EQ","path":"__mode","value":"system"},
   "hint":"الأمر system-view — مو configure terminal، هذا نظام ثاني.","weight":10},

  {"index":2,"title":"اسم الجهاز","instruction":"غيّر اسم الجهاز إلى SW-ALAMANI.",
   "expect":{"op":"STATE_EQ","path":"hostname","value":"SW-ALAMANI"},
   "hint":"sysname متبوعاً بالاسم — مو hostname.","weight":20},

  {"index":3,"title":"إنشاء الـVLAN","instruction":"أنشئ VLAN رقم 20 وسمّه GUEST.",
   "expect":{"op":"STATE_EQ","path":"vlans.20.name","value":"GUEST"},
   "hint":"vlan 20 يدخّلك نمط الـVLAN، وبعدها description GUEST.","weight":30},

  {"index":4,"title":"منفذ الوصول","instruction":"حوّل المنفذ GigabitEthernet0/0/2 لنوع access واربطه بـVLAN 20.",
   "expect":{"op":"STATE_EQ","path":"interfaces.GigabitEthernet0/0/2.accessVlan","value":"20"},
   "hint":"interface g0/0/2 ثم port link-type access ثم port default vlan 20 — خطوتان منفصلتان.","weight":40,
   "wrong":[{"matchAny":true,"say":"⚠️ بهذا النظام التهيئة **خطوتان**: نوع المنفذ أول (port link-type access) وبعدها الـVLAN (port default vlan 20). أمر واحد ما يكفي.","penalty":4}]}
]`
