# syntax=docker/dockerfile:1
FROM node:18-alpine AS deps
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install

FROM node:18-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY frontend/. .
RUN npm run build

FROM node:18-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.next ./.next
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./package.json
EXPOSE 3000
CMD ["npm","start"]

# # syntax=docker/dockerfile:1
# FROM node:18-alpine AS build
# WORKDIR /app
# COPY frontend/package.json frontend/package-lock.json ./
# RUN npm ci
# COPY frontend/. .
# RUN npm run build

# FROM node:18-alpine
# WORKDIR /app
# COPY --from=build /app/.next ./.next
# COPY --from=build /app/public ./public
# COPY --from=build /app/node_modules ./node_modules
# COPY --from=build /app/package.json ./
# EXPOSE 3000
# CMD ["npm","start"]
