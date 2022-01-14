FROM node:17-stretch-slim as builder
EXPOSE 8080
WORKDIR /src

RUN apt-get update \
  && apt-get install -y --no-install-recommends make \
  && rm -rf /var/lib/apt/lists/*

ENV YARN_CACHE_FOLDER /cache/.yarn

COPY ./package.json ./yarn.lock ./
COPY ./Makefile ./
COPY . .
