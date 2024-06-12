FROM golang:alpine AS build-stage

WORKDIR /build
COPY ["go.mod", "go.sum", "./"]
RUN go mod download
COPY . .
RUN GOOS=linux go build -o /app cmd/main.go

FROM alpine AS release

WORKDIR /
COPY --from=build-stage /build/migrations /migrations
COPY --from=build-stage /app /app
ENTRYPOINT [ "/app" ]