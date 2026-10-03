// ═══════════════════════════════════════════════════════════════════
// تتبّع الوجه والإيد من كاميرا الجهاز — **كلّه داخل الجهاز**
// ═══════════════════════════════════════════════════════════════════
//
// طلب (ع): «يطلب النظام أذونات حته الي يستخدم النظام من الموبايل
// يعتمد بي ع الكامرا الاماميه مال الجهاز ويتتبع عيون والمكان الي
// يضغط بي».
//
// 🔒 **والقيد الي ما ينكسر**: صور الموظفين **ما تطلع لأي مزوّد
//    خارجي** قبل قرار (ع) المكتوب. فهذا الملف:
//    ① يحمّل النموذج والـWASM **من نطاقنا** (`/mp/…`) مو من CDN گوگل.
//    ② **ما يرفع ولا بايت** — الاستنتاج كله بالـWASM محلياً.
//    ③ **ما يخزّن ولا إطار** — كل إطار يُقرأ ويُنسى.
//   وهاي مو وعود: `MODEL_BASE` تحت يثبت ①، وماكو أي `fetch` لرفع
//   يثبت ②، وماكو أي مصفوفة إطارات يثبت ③.
//
// ⚠️ **وهذا الملف ما يُستورد مباشرة إطلاقاً** — يُجاب بـ`import()`
//    مثل محرّك الأفتار بالضبط. السبب مقاس: المكتبة **٨٦٠ ك.ب**
//    والـWASM **١١.٧ م.ب**، فلو انستورد مباشرة كل موظف يفتح النظام
//    يحمّلها حتى لو ما شغّل الكاميرا ولا مرة.

/** أصابع الإيد بترتيب ميدياپايپ — الإبهام أول والخنصر آخر. */
export const FINGERS = ['thumb', 'index', 'middle', 'ring', 'pinky'] as const
export type FingerName = typeof FINGERS[number]

/** زوايا الرأس بالراديان، مستخرجة من مصفوفة وضع الرأس. */
export interface HeadPose {
  yaw: number
  pitch: number
  roll: number
}

export interface HandReading {
  /** `true` للإيد اليمنى **مال الشخص**، مو يمين الصورة. */
  right: boolean
  /**
   * انحناء كل إصبع بين **٠ (ممدود)** و**١ (مقبوض)**.
   * ⚠️ مشتقّة من زوايا المفاصل مو من المسافة: المسافة تتغيّر مع
   * بُعد الإيد من الكاميرا فتنطي انحناءً وهمياً، والزاوية ما تتأثر.
   */
  curls: Record<FingerName, number>
}

export interface TrackReading {
  head: HeadPose | null
  hands: HandReading[]
  /**
   * أوزان التعابير (رمشة/ابتسامة/فتح فم…) بأسماء ARKit — ٥٢ قناة.
   * ⚠️ **وهاي تُقرأ ولا تُستهلك على الشخصية الحالية**: قِستها —
   * `amani-tech-v4.glb` فيه **صفر `morphTarget`**، فماكو شي تسوقه.
   * تُعرض بالتجربة كدليل إن الإشارة موجودة، وتنربط لمن يجي مجسّم
   * فيه تعابير. **وما أسمّي توجّه الرأس «تعابير»** — شيئان مختلفان.
   */
  expressions: Record<string, number>
}

export interface TrackerOptions {
  /** يُنادى كل إطار فيه قراءة — الشاشة تسوق الشخصية منه. */
  onReading: (r: TrackReading) => void
  /** يُنادى لو فشل شي — بنص عربي للعرض، بلا إخفاء. */
  onError: (message: string) => void
  /** تتبّع الإيدين كلفته أعلى؛ نخليه اختيارياً للقياس. */
  hands?: boolean
}

export interface Tracker {
  stop(): void
  /** إطارات/ثانية مقاسة فعلاً — مو الي نتمناه. */
  fps(): number
  /** متوسط زمن الاستنتاج بالملي ثانية — يفسّر الـfps. */
  inferenceMs(): number
  /** عنصر الفيديو، حتى الشاشة تعرض معاينة للمالك. */
  video: HTMLVideoElement
}

/**
 * 🔒 **النطاق الوحيد الي يُحمَّل منه** — نطاقنا، بقاعدة التطبيق.
 * تغيير هذا لـCDN **يكسر قيد (ع) على الخصوصية**، فلا يُغيَّر.
 */
const MODEL_BASE = `${import.meta.env.BASE_URL}mp`

/** الإصبع ← أرقام نقاطه بميدياپايپ (القاعدة ← الطرف). */
const FINGER_POINTS: Record<FingerName, [number, number, number, number]> = {
  thumb: [1, 2, 3, 4],
  index: [5, 6, 7, 8],
  middle: [9, 10, 11, 12],
  ring: [13, 14, 15, 16],
  pinky: [17, 18, 19, 20],
}

type Pt = { x: number; y: number; z: number }

/** زاوية المفصل الوسطي بالراديان — ١٨٠° يعني ممدوداً تماماً. */
function jointAngle(a: Pt, b: Pt, c: Pt): number {
  const u = { x: a.x - b.x, y: a.y - b.y, z: a.z - b.z }
  const v = { x: c.x - b.x, y: c.y - b.y, z: c.z - b.z }
  const lu = Math.hypot(u.x, u.y, u.z)
  const lv = Math.hypot(v.x, v.y, v.z)
  if (lu < 1e-6 || lv < 1e-6) return Math.PI
  const dot = (u.x * v.x + u.y * v.y + u.z * v.z) / (lu * lv)
  return Math.acos(Math.min(1, Math.max(-1, dot)))
}

/**
 * انحناء إصبع من نقاطه: متوسط انحراف مفصليه عن الاستقامة.
 * ⚠️ **الإبهام له مدى أصغر تشريحياً** فينقاس بنفس الطريقة لكن
 * تطبيعه أضيق — بلا هاي يطلع الإبهام «مقبوضاً» وهو مفتوح.
 */
function curlOf(pts: Pt[], finger: FingerName): number {
  const [p0, p1, p2, p3] = FINGER_POINTS[finger].map((i) => pts[i])
  if (!p0 || !p1 || !p2 || !p3) return 0
  const bend = (Math.PI - jointAngle(p0, p1, p2)) + (Math.PI - jointAngle(p1, p2, p3))
  const span = finger === 'thumb' ? Math.PI * 0.55 : Math.PI * 1.15
  return Math.min(1, Math.max(0, bend / span))
}

/**
 * زوايا الرأس من مصفوفة الوضع (صف-أول، ١٦ عنصراً).
 * ⚠️ الترتيب مهم: `atan2` على العناصر الغلط تنطي زوايا تبين
 * «شغّالة» وهي مقلوبة المحاور — فالاستخراج قياسي هنا مو مخمّن.
 */
function poseFromMatrix(m: number[]): HeadPose | null {
  if (m.length < 16) return null
  const r02 = m[2], r12 = m[6], r22 = m[10], r00 = m[0], r01 = m[1]
  return {
    yaw: Math.atan2(r02, r22),
    pitch: Math.asin(Math.min(1, Math.max(-1, -r12))),
    roll: Math.atan2(r01, r00),
  }
}

/**
 * يشغّل الكاميرا والتتبّع. **يرمي** لو الإذن مرفوض أو الجهاز بلا
 * كاميرا — والمتصل يعرض السبب بلا ما يخفيه.
 */
export async function startTracking(options: TrackerOptions): Promise<Tracker> {
  const { FilesetResolver, FaceLandmarker, HandLandmarker } =
    await import('@mediapipe/tasks-vision')

  // ⚠️ الإذن أول شي: لو انرفض ما نحمّل ١١.٧ م.ب بلا فايدة.
  let stream: MediaStream
  try {
    stream = await navigator.mediaDevices.getUserMedia({
      video: { facingMode: 'user', width: { ideal: 640 }, height: { ideal: 480 } },
      audio: false,
    })
  } catch {
    // ⚠️ **ما نبلع السبب**: «الإذن مرفوض» و«ماكو كاميرا» مشكلتان
    // مختلفتان تماماً، والمالك لازم يعرف أيهما.
    throw new Error('الكاميرا ما اشتغلت — يا الإذن مرفوض يا ماكو كاميرا بالجهاز')
  }

  const video = document.createElement('video')
  video.autoplay = true
  video.muted = true
  video.playsInline = true
  video.srcObject = stream
  await video.play()

  const files = await FilesetResolver.forVisionTasks(`${MODEL_BASE}/wasm`)
  const face = await FaceLandmarker.createFromOptions(files, {
    baseOptions: { modelAssetPath: `${MODEL_BASE}/models/face_landmarker.task` },
    runningMode: 'VIDEO',
    numFaces: 1,
    // هاي الاثنتان هن الي ينطون التعابير ووضع الرأس — بلاهن
    // نرجع بنقاط خام وحدها وما نگدر نسوق الشخصية.
    outputFaceBlendshapes: true,
    outputFacialTransformationMatrixes: true,
  })
  const hand = options.hands
    ? await HandLandmarker.createFromOptions(files, {
      baseOptions: { modelAssetPath: `${MODEL_BASE}/models/hand_landmarker.task` },
      runningMode: 'VIDEO',
      numHands: 2,
    })
    : null

  let running = true
  let frames = 0
  let fpsValue = 0
  let inferAvg = 0
  let fpsAt = performance.now()
  let lastTs = -1

  const loop = () => {
    if (!running) return
    // ⚠️ **الطابع الزمني لازم يتقدّم**: ميدياپايپ يرمي لو ناداه
    // بنفس الطابع مرتين (يصير لو الفيديو ما جدّد إطاره بعد).
    const ts = performance.now()
    if (video.readyState >= 2 && ts > lastTs) {
      lastTs = ts
      const t0 = performance.now()
      try {
        const fr = face.detectForVideo(video, ts)
        const matrix = fr.facialTransformationMatrixes?.[0]?.data
        const head = matrix ? poseFromMatrix(Array.from(matrix)) : null

        const expressions: Record<string, number> = {}
        for (const c of fr.faceBlendshapes?.[0]?.categories ?? []) {
          if (c.categoryName) expressions[c.categoryName] = c.score
        }

        const hands: HandReading[] = []
        if (hand) {
          const hr = hand.detectForVideo(video, ts)
          hr.landmarks.forEach((pts, i) => {
            // ⚠️ **الكاميرا الأمامية مقلوبة مثل المراية**: ميدياپايپ
            // يرجّع يد الشخص الحقيقية بـ`handedness`، فنعتمد عليها
            // مو على موقع الإيد بالصورة — وإلا اليمين تصير شمال.
            const label = hr.handedness?.[i]?.[0]?.categoryName ?? ''
            const curls = {} as Record<FingerName, number>
            for (const f of FINGERS) curls[f] = curlOf(pts as Pt[], f)
            hands.push({ right: label === 'Right', curls })
          })
        }

        inferAvg = inferAvg === 0
          ? performance.now() - t0
          : inferAvg * 0.9 + (performance.now() - t0) * 0.1
        options.onReading({ head, hands, expressions })
      } catch (e) {
        // إطار واحد يفشل ما يوقف التتبّع، بس **ما ينسكت عنه**.
        options.onError(e instanceof Error ? e.message : 'فشل إطار')
      }
      frames++
      if (ts - fpsAt >= 1000) {
        fpsValue = Math.round((frames * 1000) / (ts - fpsAt))
        frames = 0
        fpsAt = ts
      }
    }
    requestAnimationFrame(loop)
  }
  requestAnimationFrame(loop)

  return {
    video,
    fps: () => fpsValue,
    inferenceMs: () => Math.round(inferAvg * 10) / 10,
    stop() {
      running = false
      // 🔒 **إطفاء الكاميرا فعلياً** مو إخفاءها: بلا `stop()` على
      // كل مسار يبقى ضوء الكاميرا شاعلاً والجهاز يصوّر — وهذا
      // **خرق ثقة** حتى لو الصورة ما تطلع من الجهاز.
      for (const t of stream.getTracks()) t.stop()
      video.srcObject = null
      face.close()
      hand?.close()
    },
  }
}
