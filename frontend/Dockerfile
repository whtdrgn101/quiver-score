FROM node:22-alpine AS build

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY nginx.conf.template /tmp/nginx.conf.template
ENV PORT=8080
EXPOSE 8080
CMD ["/bin/sh", "-c", "envsubst '${PORT} ${API_URL}' < /tmp/nginx.conf.template > /etc/nginx/conf.d/default.conf && nginx -g 'daemon off;'"]
