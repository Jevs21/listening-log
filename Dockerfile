# Stage 1: Build client
FROM node:22-alpine AS client-build
WORKDIR /app/client
COPY client/package.json client/package-lock.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

# Stage 2: Build server
FROM golang:1.25-alpine AS server-build
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=client-build /app/client/dist ../client/dist
RUN CGO_ENABLED=0 go build -o /server .
RUN CGO_ENABLED=0 go build -o /dbtools ./cmd/dbtools

# Stage 3: Runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=server-build /server /server
COPY --from=server-build /dbtools /usr/local/bin/dbtools
COPY --from=client-build /app/client/dist /client/dist
EXPOSE 8080
CMD ["/server"]
