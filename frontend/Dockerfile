# syntax=docker/dockerfile:1
# مرحلة البناء
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
# كاش npm دائم (10-08): ما ينزّل الحزم من النت كل مرة.
RUN --mount=type=cache,target=/root/.npm npm ci --prefer-offline --no-audit --no-fund
COPY . .
# بالسيرفر الفعلي الموقع يكون على جذر الدومين مباشرة، والباك إند وراء نفس
# الدومين على /api (يتعامل معه nginx كـ reverse proxy)
ENV VITE_BASE_PATH=/
ENV VITE_API_URL=/api
RUN npm run build

# مرحلة التشغيل — nginx يصدّر الملفات الثابتة ويحوّل /api للباك إند
FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
