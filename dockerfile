FROM golang:1.24.3-alpine AS builder

# Instalação de dependências do sistema
RUN apk add --no-cache git

# Configurar diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Fazer download das dependências
RUN go mod download

# Copiar o código-fonte
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Segunda fase - imagem mínima para execução
FROM alpine:latest

# Instalar certificados e timezone
RUN apk --no-cache add ca-certificates tzdata

# Copiar o binário compilado
COPY --from=builder /app/app /app

# Copiar o arquivo .env.example para dentro da imagem
COPY .env.example /.env

# Expor a porta da aplicação
EXPOSE 8080

# Executar a aplicação
CMD ["/app"]