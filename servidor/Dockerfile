# Construído a partir da raiz do repositório (o painel index.html fica lá):
#   docker build -f servidor/Dockerfile .      (o docker-compose.yml já faz assim)
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY servidor/go.mod servidor/go.sum ./
RUN go mod download
COPY servidor/ .
# o painel servido em /painel é o mesmo index.html do GitHub Pages
COPY index.html api/web/painel/index.html
COPY vendor/xlsx.full.min.js api/web/painel/xlsx.full.min.js
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/servidor .

FROM alpine:3.22
# ca-certificates: a API Gobrax é HTTPS; tzdata: datas de Brasília nos logs
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H app
ENV TZ=America/Sao_Paulo
USER app
COPY --from=build /out/servidor /servidor
EXPOSE 8080
ENTRYPOINT ["/servidor"]
