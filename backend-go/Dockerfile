# مرحلة البناء
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/api

# مرحلة التشغيل — صورة نهائية صغيرة بدون أدوات البناء
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
# مستخدم غير root مخصص — الصورة كانت تشتغل root افتراضياً بدون هذا السطر،
# وهذا يوسّع أي ثغرة اختراق بالسيرفر لاختراق كامل بصلاحيات root داخل الحاوية.
# UID/GID ثابتين (1000) حتى نقدر نمنح صلاحية قراءة صريحة على bind-mount
# ./tutorials من docker-compose.yml (يُقرأ فقط عبر os.ReadDir/os.ReadFile،
# ما يحتاج كتابة) — الصورة لا تكتب لأي ملف محلي وقت التشغيل (لا سجلات ولا
# قاعدة بيانات محلية)، فما نحتاج أي مجلد إضافي بصلاحية كتابة لـappuser.
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build /server /app/server
# مجلد الملفات المرفوعة: ينمسك من فوليوم دائم بالـcompose، بس ننشئه
# هنا بملكية appuser حتى يكدر يكتب بيه — بدونه الحاوية كانت تموت عند
# الإقلاع لأن /app ملك root والمستخدم مو root.
RUN mkdir -p /app/uploads && chown appuser:appuser /app/uploads
USER appuser
EXPOSE 4000
ENTRYPOINT ["/app/server"]
