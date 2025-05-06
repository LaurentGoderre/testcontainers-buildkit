FROM golang:1.24-alpine AS build
WORKDIR /go/src/example.com/frontend/test
RUN apk add --no-cache file
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/build ./cmd/build && file /out/build | grep "statically linked"
  
FROM alpine:3.21
RUN apk add --no-cache alpine-sdk
COPY --from=build /out/build /bin/build
ENTRYPOINT ["/bin/build"]
