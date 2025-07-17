FROM reg.undercloud.cri.epita.fr/docker/golang:1.21-alpine AS builder
WORKDIR /app
LABEL maintainer="heman.narayanan@epita.fr"
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /comments-engine

FROM reg.undercloud.cri.epita.fr/docker/alpine:3.19
RUN apk add --no-cache curl bash && \
    adduser -D appuser
USER appuser
WORKDIR /app
COPY --from=builder /comments-engine /app/
ENV HOST="0.0.0.0"
EXPOSE 8000
CMD ["./comments-engine"]