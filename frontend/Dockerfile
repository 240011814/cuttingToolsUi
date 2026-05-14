# 指定多架构支持
FROM --platform=$BUILDPLATFORM nginx:stable-alpine

# 设置构建参数，可选：不同平台使用不同基础镜像
ARG TARGETPLATFORM
ARG BACKEND_URL=http://backend:8080

# 复制前端打包文件
COPY dist /usr/share/nginx/html

# 复制 Nginx 模板配置
COPY nginx.conf.template /etc/nginx/templates/nginx.conf.template

# 设置环境变量，用于 envsubst 替换
ENV BACKEND_URL=${BACKEND_URL} \
    NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx

# 暴露端口
EXPOSE 80

# 使用官方 entrypoint，保证 envsubst 能正常工作
ENTRYPOINT ["/docker-entrypoint.sh"]

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]
