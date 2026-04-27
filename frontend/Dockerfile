# Build stage
FROM node:24-alpine AS builder

# 安装 pnpm
RUN npm install -g pnpm@latest

WORKDIR /app

# 复制依赖定义并安装
COPY package.json pnpm-lock.yaml ./
# 如果有 workspace 相关的配置，可能需要特殊处理
# 但这里我们假设 frontend 是自包含的
RUN pnpm install --frozen-lockfile

# 复制源码并构建
COPY . .
RUN pnpm build

# Production stage
FROM nginx:stable-alpine

# 复制构建产物到 Nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
