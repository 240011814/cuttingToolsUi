FROM nginx:stable-alpine

# 设置端口，可在 docker run 时映射
ARG APP_PORT=80
ENV APP_PORT=${APP_PORT}
EXPOSE ${APP_PORT}

# 直接复制 Actions 构建好的 dist 文件夹
COPY dist /usr/share/nginx/html

# 修改 nginx 默认监听端口
RUN sed -i "s/listen 80;/listen ${APP_PORT};/" /etc/nginx/conf.d/default.conf

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]
