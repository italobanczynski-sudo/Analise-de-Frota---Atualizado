FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/servidor .

FROM alpine:3.20
# ca-certificates: a API Gobrax é HTTPS; tzdata: datas de Brasília nos logs
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H app
ENV TZ=America/Sao_Paulo
USER app
COPY --from=build /out/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
