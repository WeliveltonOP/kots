FROM node:17-stretch-slim as builder

WORKDIR /src

ARG OKTETO_NAMESPACE

RUN apt-get update \
  && apt-get install -y --no-install-recommends make \
  && rm -rf /var/lib/apt/lists/*

COPY package.json yarn.lock Makefile ./
RUN --mount=type=cache,target=/root/.yarn YARN_CACHE_FOLDER=/root/.yarn make deps

COPY . .
RUN --mount=type=cache,target=./node_modules/.cache/webpack make build-dev


FROM nginx:1.21.4-alpine
COPY --from=builder /src/dist /usr/share/nginx/html
COPY deploy/nginx.conf.dev /etc/nginx/conf.d/default.conf
EXPOSE 8080
