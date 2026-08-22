FROM golang:1.25.0-alpine AS builder

WORKDIR /login_rate

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o /login_rate/login_rate ./iternal/cmd

FROM alpine:3.22

RUN addgroup -S login_rate && adduser -S -G login_rate login_rate

WORKDIR /login_rate

COPY --from=builder /login_rate/login_rate ./iternal/cmd/login_rate

USER login_rate

EXPOSE 8080

ENTRYPOINT ["./iternal/cmd/login_rate"]