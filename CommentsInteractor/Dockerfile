FROM reg.undercloud.cri.epita.fr/docker/golang:1.21-alpine AS builder
LABEL maintainer="heman.narayanan@epita.fr"
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /comments-interactor

FROM reg.undercloud.cri.epita.fr/docker/alpine:3.19
RUN apk add --no-cache curl bash && \
    adduser -D appuser
USER appuser
WORKDIR /app
COPY --from=builder /comments-interactor /app/
ENV HOST="0.0.0.0"
EXPOSE 9000
CMD ["./comments-interactor"]