# Build stage
FROM node:24-alpine AS builder

# 安装 pnpm
RUN npm install -g pnpm@latest

WORKDIR /app

# 复制 workspace 相关文件
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY frontend ./frontend

# 安装依赖
RUN pnpm install --frozen-lockfile

# 构建
WORKDIR /app/frontend
RUN pnpm build

# Production stage
FROM nginx:stable-alpine

# 复制构建产物到 Nginx 目录
COPY --from=builder /app/frontend/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY frontend/nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
