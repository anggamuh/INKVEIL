FROM golang:1.25-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /inkveil .
FROM alpine:3.22
RUN addgroup -S inkveil && adduser -S -G inkveil inkveil && mkdir -p /app/data && chown -R inkveil:inkveil /app
WORKDIR /app
COPY --from=build /inkveil /app/inkveil
USER inkveil
ENV ADDR=0.0.0.0:8080
EXPOSE 8080
VOLUME ["/app/data"]
ENTRYPOINT ["/app/inkveil"]
