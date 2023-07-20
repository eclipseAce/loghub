FROM node:lts as webui

WORKDIR /app

COPY webui/ webui

RUN cd webui \
 && npm install \
 && npm run build --registry=https://registry.npmmirror.com

FROM golang as backend

WORKDIR /app

COPY main.go go.mod go.sum .
COPY msg/ msg
COPY web/ web
COPY --from=webui /app/webui/ webui

RUN env GOOS=linux GOARCH=amd64 go build

FROM debian:stable-slim

ENV LANG=en_US.utf8
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=backend /app/loghub .

EXPOSE 5044
EXPOSE 6060

ENTRYPOINT ./loghub
