# --- Stage 1: Builder ---
FROM golang:1.25-alpine AS builder

# Gerekli sistem kütüphanelerini yükle (Alpine için)
RUN apk add --no-cache git

# Çalışma dizinini ayarla
WORKDIR /app

# Bağımlılık dosyalarını kopyala
COPY go.mod go.sum ./

# Bağımlılıkları indir
RUN go mod download

# Kaynak kodun geri kalanını kopyala
COPY . .

# Uygulamayı derle
# CGO_ENABLED=0 -> Statik binary üretir (daha taşınabilir)
# GOOS=linux -> Linux için derle
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- Stage 2: Runner ---
FROM alpine:latest

# Güvenlik sertifikalarını yükle (HTTPS istekleri için gerekli)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Builder aşamasından sadece derlenmiş dosyayı (binary) kopyala
COPY --from=builder /app/main .

# .env dosyasını production'da container içine kopyalamayız,
# environment variable olarak veririz. Ama iskelet olduğu için opsiyonel:
# COPY .env .

# Uygulamanın portunu dışarı aç
EXPOSE 3000

# Uygulamayı başlat
CMD ["./main"]