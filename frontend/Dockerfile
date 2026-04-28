FROM nginx:stable-alpine
COPY dist /usr/share/nginx/html
COPY nginx.conf.template /etc/nginx/templates/nginx.conf.template
ENV BACKEND_URL=http://backend:8080
ENV NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
