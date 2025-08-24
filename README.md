# 🌾 Tarım Yönetim Sistemi API

Flutter mobil uygulaması için geliştirilmiş kapsamlı Tarım Yönetim Sistemi REST API'si.

## 🚀 Özellikler

- **🔐 JWT Tabanlı Kimlik Doğrulama** - Güvenli kullanıcı yönetimi
- **🌱 Arazi Yönetimi** - Arazi kayıtları, aktiviteler ve verimlilik analizi
- **🐄 Hayvancılık Yönetimi** - Hayvan kayıtları, sağlık takibi ve süt üretimi
- **🌾 Üretim Yönetimi** - Ürün takibi, hasat kayıtları ve kalite kontrolü
- **💰 Finans Yönetimi** - Gelir-gider takibi ve finansal analizler
- **📅 Takvim ve Etkinlikler** - Planlama ve hatırlatıcılar
- **🔔 Bildirim Sistemi** - Akıllı bildirimler ve uyarılar
- **📊 Dashboard ve Raporlar** - Detaylı istatistikler ve grafikler
- **🌤️ Hava Durumu Entegrasyonu** - Tarımsal uyarılar ve tahminler
- **📚 Swagger Dokümantasyonu** - Kapsamlı API dokümantasyonu

## 🛠️ Teknolojiler

- **Backend**: Go (Golang)
- **Veritabanı**: SQLite
- **Web Framework**: Gin
- **Authentication**: JWT
- **API Documentation**: Swagger
- **Password Hashing**: bcrypt

## 📋 Gereksinimler

- Go 1.21+
- SQLite3
- Git

## 🚀 Kurulum

### 1. Projeyi Klonlayın

```bash
git clone <repository-url>
cd Agri-Management_Api
```

### 2. Bağımlılıkları Yükleyin

```bash
go mod tidy
```

### 3. Environment Dosyasını Ayarlayın

```bash
cp config.env.example config.env
# config.env dosyasını düzenleyin
```

### 4. Swagger Dokümantasyonunu Oluşturun

```bash
swag init -g cmd/api/main.go
```

### 5. Uygulamayı Çalıştırın

```bash
go run cmd/api/main.go
```

API varsayılan olarak `http://localhost:8080` adresinde çalışacaktır.

## 📚 API Dokümantasyonu

Swagger dokümantasyonuna `http://localhost:8080/swagger/index.html` adresinden erişebilirsiniz.

## 🔐 API Endpoints

### Kimlik Doğrulama
- `POST /api/v1/auth/register` - Kullanıcı kaydı
- `POST /api/v1/auth/login` - Kullanıcı girişi
- `POST /api/v1/auth/refresh` - Token yenileme
- `GET /api/v1/auth/profile` - Profil bilgileri
- `PUT /api/v1/auth/profile` - Profil güncelleme
- `PUT /api/v1/auth/change-password` - Şifre değiştirme
- `POST /api/v1/auth/logout` - Çıkış yapma

### Dashboard
- `GET /api/v1/dashboard/summary` - Dashboard özeti
- `GET /api/v1/dashboard/recent-activities` - Son aktiviteler
- `GET /api/v1/dashboard/charts/income-expense` - Gelir-gider grafik
- `GET /api/v1/dashboard/charts/production` - Üretim grafik

### Arazi Yönetimi
- `GET /api/v1/lands` - Arazi listesi
- `POST /api/v1/lands` - Yeni arazi oluşturma
- `GET /api/v1/lands/{id}` - Arazi detayları
- `PUT /api/v1/lands/{id}` - Arazi güncelleme
- `DELETE /api/v1/lands/{id}` - Arazi silme
- `GET /api/v1/lands/statistics` - Arazi istatistikleri
- `GET /api/v1/lands/{id}/activities` - Arazi aktiviteleri
- `POST /api/v1/lands/{id}/activities` - Aktivite oluşturma

### Hayvancılık Yönetimi
- `GET /api/v1/livestock` - Hayvan listesi
- `POST /api/v1/livestock` - Yeni hayvan ekleme
- `GET /api/v1/livestock/{id}` - Hayvan detayları
- `PUT /api/v1/livestock/{id}` - Hayvan güncelleme
- `DELETE /api/v1/livestock/{id}` - Hayvan silme
- `GET /api/v1/livestock/statistics` - Hayvancılık istatistikleri
- `GET /api/v1/livestock/{id}/health-records` - Sağlık kayıtları
- `POST /api/v1/livestock/{id}/health-records` - Sağlık kaydı ekleme

### Üretim Yönetimi
- `GET /api/v1/production` - Üretim listesi
- `POST /api/v1/production` - Yeni üretim kaydı
- `GET /api/v1/production/{id}` - Üretim detayları
- `PUT /api/v1/production/{id}` - Üretim güncelleme
- `DELETE /api/v1/production/{id}` - Üretim silme
- `GET /api/v1/production/statistics` - Üretim istatistikleri

### Finans Yönetimi
- `GET /api/v1/finance/summary` - Finansal özet
- `GET /api/v1/finance/transactions` - İşlem listesi
- `POST /api/v1/finance/transactions` - Yeni işlem
- `GET /api/v1/finance/transactions/{id}` - İşlem detayları
- `PUT /api/v1/finance/transactions/{id}` - İşlem güncelleme
- `DELETE /api/v1/finance/transactions/{id}` - İşlem silme
- `GET /api/v1/finance/analysis` - Finansal analiz

### Takvim ve Etkinlikler
- `GET /api/v1/calendar/events` - Etkinlik listesi
- `POST /api/v1/calendar/events` - Yeni etkinlik
- `GET /api/v1/calendar/events/{id}` - Etkinlik detayları
- `PUT /api/v1/calendar/events/{id}` - Etkinlik güncelleme
- `DELETE /api/v1/calendar/events/{id}` - Etkinlik silme
- `PATCH /api/v1/calendar/events/{id}/status` - Durum güncelleme

### Bildirimler
- `GET /api/v1/notifications` - Bildirim listesi
- `PATCH /api/v1/notifications/{id}/read` - Okundu işaretleme
- `PATCH /api/v1/notifications/mark-all-read` - Tümünü okundu işaretleme
- `DELETE /api/v1/notifications/{id}` - Bildirim silme
- `GET /api/v1/notifications/settings` - Bildirim ayarları
- `PUT /api/v1/notifications/settings` - Ayarları güncelleme

### Ayarlar
- `GET /api/v1/settings` - Uygulama ayarları
- `PUT /api/v1/settings` - Ayarları güncelleme
- `GET /api/v1/settings/system-info` - Sistem bilgileri
- `POST /api/v1/settings/backup` - Veri yedekleme
- `POST /api/v1/settings/restore` - Veri geri yükleme

### Hava Durumu
- `GET /api/v1/weather/current` - Güncel hava durumu
- `GET /api/v1/weather/forecast` - Hava durumu tahmini
- `GET /api/v1/weather/agricultural-alerts` - Tarımsal uyarılar

## 🗄️ Veritabanı Şeması

### Ana Tablolar
- **users** - Kullanıcı bilgileri
- **lands** - Arazi kayıtları
- **livestock** - Hayvan kayıtları
- **production** - Üretim kayıtları
- **transactions** - Finansal işlemler
- **events** - Takvim etkinlikleri
- **notifications** - Bildirimler
- **health_records** - Sağlık kayıtları
- **milk_production** - Süt üretim kayıtları
- **land_activities** - Arazi aktiviteleri

## 🔒 Güvenlik

- JWT token tabanlı kimlik doğrulama
- Şifre hash'leme (bcrypt)
- Role-based access control
- CORS yapılandırması
- Input validation
- SQL injection koruması

## 📱 Flutter Entegrasyonu

Bu API, Flutter mobil uygulaması ile tam uyumlu olarak tasarlanmıştır:

- RESTful API standartları
- JSON response formatı
- Hata yönetimi
- Sayfalama desteği
- Filtreleme ve arama
- Real-time bildirimler

## 🧪 Test

```bash
# Unit testleri çalıştır
go test ./...

# Coverage raporu
go test -cover ./...
```

## 📦 Deployment

### Docker ile

```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main cmd/api/main.go
EXPOSE 8080
CMD ["./main"]
```

### Production için

```bash
# Binary oluştur
go build -o agri-api cmd/api/main.go

# Çalıştır
./agri-api
```

## 🤝 Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit yapın (`git commit -m 'Add amazing feature'`)
4. Push yapın (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır. Detaylar için `LICENSE` dosyasına bakın.

## 📞 İletişim

- **Proje Sahibi**: [Adınız]
- **Email**: [email@example.com]
- **Proje Linki**: [https://github.com/username/Agri-Management_Api](https://github.com/username/Agri-Management_Api)

## 🙏 Teşekkürler

- [Gin Framework](https://github.com/gin-gonic/gin)
- [SQLite](https://www.sqlite.org/)
- [Swagger](https://swagger.io/)
- [JWT](https://jwt.io/)

---

⭐ Bu projeyi beğendiyseniz yıldız vermeyi unutmayın!
