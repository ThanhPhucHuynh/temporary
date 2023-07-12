FROM node:16-alpine as deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i; \
  else echo "Lockfile not found." && exit 1; \
  fi

# rebuild
FROM node:16-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG REACT_APP_API_KEY
ENV REACT_APP_API_KEY=$REACT_APP_API_KEY
RUN yarn build 

# production stage
FROM nginx:1.17-alpine as production-stage
COPY --from=builder /app/build /usr/share/nginx/html
COPY /deployment/nginx.conf /etc/nginx/conf.d/default.conf
# RUN rm -rf ./*

EXPOSE 80
ENV PORT 80
CMD ["nginx", "-g", "daemon off;"]


# docker build -t phimdeee --build-arg REACT_APP_API_KEY=ad391402ee734399f6e59f3316372e90 . 
# docker run -p 3000:80 phimdeee