# Build stage
FROM golang:1.24 AS build
WORKDIR /src
ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod .
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/app ./cmd/app &&     CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/producer ./cmd/producer

# Final image
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/app /app
COPY --from=build /out/producer /producer
COPY web /web
ENV APP_NAME=wb-order-service HTTP_PORT=8081
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/app"]
