FROM node:22-alpine AS build

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build


# Production environment
FROM nginxinc/nginx-unprivileged:alpine

# Copy built files
COPY --from=build /app/dist /usr/share/nginx/html/admin

# Copy entrypoint script
COPY --chmod=555 docker-entrypoint.sh /docker-entrypoint.sh

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

# Use entrypoint script to replace environment variables at container start
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"] 