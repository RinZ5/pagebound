FROM node:22-alpine AS frontend
WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/dist internal/webui/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pagebound ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /pagebound /usr/local/bin/pagebound
WORKDIR /data
EXPOSE 8080
ENV BOOKS_DIR=/data/books
ENV DB_PATH=/data/library.db
ENTRYPOINT ["pagebound"]
