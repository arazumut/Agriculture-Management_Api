# Tarım Yönetim Sistemi - API Endpoint'leri Gereksinimleri

Bu dokümanda, Flutter mobil uygulamanızın çalışabilmesi için ihtiyaç duyulan tüm API endpoint'leri detaylı bir şekilde açıklanmıştır.

## 📋 İçindekiler

1. [Kimlik Doğrulama ve Kullanıcı Yönetimi](#1-kimlik-doğrulama-ve-kullanıcı-yönetimi)
2. [Dashboard ve İstatistikler](#2-dashboard-ve-istatistikler)
3. [Arazi Yönetimi](#3-arazi-yönetimi)
4. [Hayvancılık Yönetimi](#4-hayvancılık-yönetimi)
5. [Üretim Yönetimi](#5-üretim-yönetimi)
6. [Finans Yönetimi](#6-finans-yönetimi)
7. [Raporlar ve Analizler](#7-raporlar-ve-analizler)
8. [Takvim ve Etkinlik Yönetimi](#8-takvim-ve-etkinlik-yönetimi)
9. [Bildirimler](#9-bildirimler)
10. [Ayarlar ve Konfigürasyon](#10-ayarlar-ve-konfigürasyon)
11. [Hava Durumu](#11-hava-durumu)

---

## 1. Kimlik Doğrulama ve Kullanıcı Yönetimi

### 1.1 Kullanıcı Girişi
```http
POST /api/auth/login
```
**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```
**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "string",
      "name": "string",
      "email": "string",
      "avatar": "string",
      "role": "string",
      "isVerified": boolean
    },
    "token": "string",
    "refreshToken": "string"
  }
}
```

### 1.2 Kullanıcı Kaydı
```http
POST /api/auth/register
```
**Request Body:**
```json
{
  "name": "string",
  "email": "string",
  "password": "string",
  "confirmPassword": "string",
  "farmName": "string",
  "location": "string"
}
```

### 1.3 Token Yenileme
```http
POST /api/auth/refresh
```
**Request Body:**
```json
{
  "refreshToken": "string"
}
```

### 1.4 Kullanıcı Profili Bilgileri
```http
GET /api/auth/profile
```
**Headers:** `Authorization: Bearer {token}`

### 1.5 Profil Güncelleme
```http
PUT /api/auth/profile
```
**Request Body:**
```json
{
  "name": "string",
  "email": "string",
  "avatar": "string",
  "farmName": "string",
  "location": "string"
}
```

### 1.6 Şifre Değiştirme
```http
PUT /api/auth/change-password
```
**Request Body:**
```json
{
  "currentPassword": "string",
  "newPassword": "string"
}
```

### 1.7 Çıkış Yapma
```http
POST /api/auth/logout
```

---

## 2. Dashboard ve İstatistikler

### 2.1 Dashboard Özet Verileri
```http
GET /api/dashboard/summary
```
**Response:**
```json
{
  "success": true,
  "data": {
    "totalAnimals": {
      "count": number,
      "trend": "string", // "+5", "-2"
      "percentage": number
    },
    "totalLands": {
      "area": number,
      "count": number,
      "productivity": number
    },
    "monthlyIncome": {
      "amount": number,
      "currency": "string",
      "trend": "string"
    },
    "monthlyExpense": {
      "amount": number,
      "currency": "string",
      "trend": "string"
    },
    "activeProducts": {
      "count": number,
      "categories": number
    }
  }
}
```

### 2.2 Son Aktiviteler
```http
GET /api/dashboard/recent-activities
```
**Query Parameters:**
- `limit`: number (default: 10)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "string",
      "type": "string", // "vaccination", "health_check", "harvest", etc.
      "title": "string",
      "description": "string",
      "date": "ISO date string",
      "category": "string", // "livestock", "land", "finance"
      "icon": "string"
    }
  ]
}
```

### 2.3 Grafik Verileri - Aylık Gelir/Gider
```http
GET /api/dashboard/charts/income-expense
```
**Query Parameters:**
- `period`: string ("month", "quarter", "year")

**Response:**
```json
{
  "success": true,
  "data": {
    "labels": ["string"],
    "income": [number],
    "expense": [number],
    "profit": [number]
  }
}
```

### 2.4 Grafik Verileri - Üretim Analizi
```http
GET /api/dashboard/charts/production
```
**Response:**
```json
{
  "success": true,
  "data": {
    "categories": ["string"],
    "values": [number],
    "colors": ["string"]
  }
}
```

---

## 3. Arazi Yönetimi

### 3.1 Arazi Listesi
```http
GET /api/lands
```
**Query Parameters:**
- `page`: number
- `limit`: number
- `status`: string ("active", "inactive", "all")

**Response:**
```json
{
  "success": true,
  "data": {
    "lands": [
      {
        "id": "string",
        "name": "string",
        "area": number,
        "unit": "string", // "hectare", "dönüm"
        "crop": "string",
        "status": "string",
        "lastActivity": "string",
        "productivity": number,
        "location": {
          "latitude": number,
          "longitude": number,
          "address": "string"
        },
        "createdAt": "ISO date string",
        "updatedAt": "ISO date string"
      }
    ],
    "pagination": {
      "page": number,
      "limit": number,
      "total": number,
      "totalPages": number
    }
  }
}
```

### 3.2 Arazi Detayları
```http
GET /api/lands/{id}
```

### 3.3 Yeni Arazi Ekleme
```http
POST /api/lands
```
**Request Body:**
```json
{
  "name": "string",
  "area": number,
  "unit": "string",
  "crop": "string",
  "location": {
    "latitude": number,
    "longitude": number,
    "address": "string"
  },
  "soilType": "string",
  "irrigationType": "string"
}
```

### 3.4 Arazi Güncelleme
```http
PUT /api/lands/{id}
```

### 3.5 Arazi Silme
```http
DELETE /api/lands/{id}
```

### 3.6 Arazi İstatistikleri
```http
GET /api/lands/statistics
```
**Response:**
```json
{
  "success": true,
  "data": {
    "totalArea": number,
    "totalLands": number,
    "averageProductivity": number,
    "activeCrops": number,
    "landsByStatus": {
      "active": number,
      "inactive": number,
      "maintenance": number
    }
  }
}
```

### 3.7 Arazi Aktiviteleri
```http
GET /api/lands/{id}/activities
```
**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "string",
      "type": "string", // "irrigation", "fertilization", "harvest"
      "description": "string",
      "date": "ISO date string",
      "notes": "string",
      "cost": number,
      "result": "string"
    }
  ]
}
```

### 3.8 Arazi Aktivitesi Ekleme
```http
POST /api/lands/{id}/activities
```
**Request Body:**
```json
{
  "type": "string",
  "description": "string",
  "scheduledDate": "ISO date string",
  "notes": "string",
  "estimatedCost": number
}
```

### 3.9 Verimlilik Analizi
```http
GET /api/lands/productivity-analysis
```
**Query Parameters:**
- `period`: string ("month", "quarter", "year")

---

## 4. Hayvancılık Yönetimi

### 4.1 Hayvan Listesi
```http
GET /api/livestock
```
**Query Parameters:**
- `page`: number
- `limit`: number
- `type`: string ("cattle", "sheep", "goat", "chicken")
- `status`: string ("healthy", "sick", "pregnant")

**Response:**
```json
{
  "success": true,
  "data": {
    "animals": [
      {
        "id": "string",
        "tagNumber": "string",
        "type": "string",
        "breed": "string",
        "gender": "string",
        "birthDate": "ISO date string",
        "weight": number,
        "healthStatus": "string",
        "location": "string",
        "mother": "string",
        "father": "string",
        "notes": "string",
        "createdAt": "ISO date string"
      }
    ],
    "pagination": {
      "page": number,
      "limit": number,
      "total": number,
      "totalPages": number
    }
  }
}
```

### 4.2 Hayvan Detayları
```http
GET /api/livestock/{id}
```

### 4.3 Yeni Hayvan Ekleme
```http
POST /api/livestock
```
**Request Body:**
```json
{
  "tagNumber": "string",
  "type": "string",
  "breed": "string",
  "gender": "string",
  "birthDate": "ISO date string",
  "weight": number,
  "healthStatus": "string",
  "location": "string",
  "mother": "string",
  "father": "string",
  "notes": "string"
}
```

### 4.4 Hayvan Güncelleme
```http
PUT /api/livestock/{id}
```

### 4.5 Hayvan Silme
```http
DELETE /api/livestock/{id}
```

### 4.6 Hayvancılık İstatistikleri
```http
GET /api/livestock/statistics
```
**Response:**
```json
{
  "success": true,
  "data": {
    "totalAnimals": number,
    "animalsByType": {
      "cattle": number,
      "sheep": number,
      "goat": number,
      "chicken": number
    },
    "healthStatistics": {
      "healthy": number,
      "sick": number,
      "pregnant": number,
      "vaccination_needed": number
    },
    "dailyMilkProduction": number,
    "vaccinationRate": number
  }
}
```

### 4.7 Sağlık Kayıtları
```http
GET /api/livestock/{id}/health-records
```
**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "string",
      "type": "string", // "vaccination", "treatment", "checkup"
      "description": "string",
      "date": "ISO date string",
      "veterinarian": "string",
      "cost": number,
      "notes": "string",
      "nextCheckup": "ISO date string"
    }
  ]
}
```

### 4.8 Sağlık Kaydı Ekleme
```http
POST /api/livestock/{id}/health-records
```
**Request Body:**
```json
{
  "type": "string",
  "description": "string",
  "date": "ISO date string",
  "veterinarian": "string",
  "cost": number,
  "notes": "string",
  "nextCheckup": "ISO date string"
}
```

### 4.9 Süt Üretim Kayıtları
```http
GET /api/livestock/milk-production
```
**Query Parameters:**
- `startDate`: ISO date string
- `endDate`: ISO date string
- `animalId`: string (optional)

### 4.10 Süt Üretim Kaydı Ekleme
```http
POST /api/livestock/milk-production
```
**Request Body:**
```json
{
  "animalId": "string",
  "date": "ISO date string",
  "amount": number,
  "quality": "string",
  "notes": "string"
}
```

### 4.11 Kategori Analizi
```http
GET /api/livestock/categories
```
**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "string",
      "count": number,
      "icon": "string",
      "color": "string"
    }
  ]
}
```

---

## 5. Üretim Yönetimi

### 5.1 Üretim Listesi
```http
GET /api/production
```
**Query Parameters:**
- `page`: number
- `limit`: number
- `category`: string
- `status`: string

**Response:**
```json
{
  "success": true,
  "data": {
    "productions": [
      {
        "id": "string",
        "name": "string",
        "category": "string",
        "amount": number,
        "unit": "string",
        "harvestDate": "ISO date string",
        "quality": "string",
        "storageLocation": "string",
        "status": "string",
        "price": number,
        "landId": "string",
        "notes": "string"
      }
    ],
    "pagination": {
      "page": number,
      "limit": number,
      "total": number,
      "totalPages": number
    }
  }
}
```

### 5.2 Üretim Detayları
```http
GET /api/production/{id}
```

### 5.3 Yeni Üretim Kaydı Ekleme
```http
POST /api/production
```
**Request Body:**
```json
{
  "name": "string",
  "category": "string",
  "amount": number,
  "unit": "string",
  "harvestDate": "ISO date string",
  "quality": "string",
  "storageLocation": "string",
  "price": number,
  "landId": "string",
  "notes": "string"
}
```

### 5.4 Üretim Güncelleme
```http
PUT /api/production/{id}
```

### 5.5 Üretim Silme
```http
DELETE /api/production/{id}
```

### 5.6 Üretim İstatistikleri
```http
GET /api/production/statistics
```
**Response:**
```json
{
  "success": true,
  "data": {
    "activeProducts": number,
    "totalProduction": number,
    "averageProductivity": number,
    "qualityDistribution": {
      "A+": number,
      "A": number,
      "B": number,
      "C": number
    },
    "categoryBreakdown": [
      {
        "name": "string",
        "count": number,
        "amount": number,
        "percentage": number
      }
    ]
  }
}
```

### 5.7 Ürün Kategorileri
```http
GET /api/production/categories
```
**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "string",
      "count": number,
      "icon": "string",
      "color": "string"
    }
  ]
}
```

---

## 6. Finans Yönetimi

### 6.1 Finansal Özet
```http
GET /api/finance/summary
```
**Query Parameters:**
- `period`: string ("month", "quarter", "year")

**Response:**
```json
{
  "success": true,
  "data": {
    "totalIncome": number,
    "totalExpense": number,
    "netProfit": number,
    "pendingPayments": number,
    "trends": {
      "income": number,
      "expense": number,
      "profit": number
    }
  }
}
```

### 6.2 İşlem Listesi
```http
GET /api/finance/transactions
```
**Query Parameters:**
- `page`: number
- `limit`: number
- `type`: string ("income", "expense")
- `category`: string
- `startDate`: ISO date string
- `endDate`: ISO date string

**Response:**
```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": "string",
        "type": "string", // "income", "expense"
        "category": "string",
        "description": "string",
        "amount": number,
        "currency": "string",
        "date": "ISO date string",
        "status": "string", // "completed", "pending", "cancelled"
        "paymentMethod": "string",
        "receipt": "string",
        "notes": "string"
      }
    ],
    "pagination": {
      "page": number,
      "limit": number,
      "total": number,
      "totalPages": number
    }
  }
}
```

### 6.3 Yeni İşlem Ekleme
```http
POST /api/finance/transactions
```
**Request Body:**
```json
{
  "type": "string",
  "category": "string",
  "description": "string",
  "amount": number,
  "currency": "string",
  "date": "ISO date string",
  "paymentMethod": "string",
  "receipt": "string",
  "notes": "string"
}
```

### 6.4 İşlem Güncelleme
```http
PUT /api/finance/transactions/{id}
```

### 6.5 İşlem Silme
```http
DELETE /api/finance/transactions/{id}
```

### 6.6 Kategori Listesi
```http
GET /api/finance/categories
```
**Response:**
```json
{
  "success": true,
  "data": {
    "income": ["string"],
    "expense": ["string"]
  }
}
```

### 6.7 Gelir-Gider Analizi
```http
GET /api/finance/analysis
```
**Query Parameters:**
- `period`: string
- `startDate`: ISO date string
- `endDate`: ISO date string

**Response:**
```json
{
  "success": true,
  "data": {
    "monthly": [
      {
        "month": "string",
        "income": number,
        "expense": number,
        "profit": number
      }
    ],
    "byCategory": [
      {
        "category": "string",
        "amount": number,
        "percentage": number
      }
    ]
  }
}
```

---

## 7. Raporlar ve Analizler

### 7.1 Rapor Listesi
```http
GET /api/reports
```
**Query Parameters:**
- `type`: string ("financial", "production", "livestock", "land")
- `period`: string

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "string",
      "title": "string",
      "type": "string",
      "description": "string",
      "generatedDate": "ISO date string",
      "period": "string",
      "format": "string", // "pdf", "excel", "csv"
      "downloadUrl": "string"
    }
  ]
}
```

### 7.2 Rapor Oluşturma
```http
POST /api/reports/generate
```
**Request Body:**
```json
{
  "type": "string",
  "period": "string",
  "startDate": "ISO date string",
  "endDate": "ISO date string",
  "format": "string",
  "includeCharts": boolean,
  "categories": ["string"]
}
```

### 7.3 Rapor İndirme
```http
GET /api/reports/{id}/download
```

### 7.4 Performans Metrikleri
```http
GET /api/reports/performance-metrics
```
**Query Parameters:**
- `period`: string

**Response:**
```json
{
  "success": true,
  "data": {
    "efficiency": number,
    "productivity": number,
    "profitability": number,
    "sustainability": number,
    "trends": [
      {
        "metric": "string",
        "value": number,
        "change": number,
        "trend": "string"
      }
    ]
  }
}
```

### 7.5 Karşılaştırma Analizi
```http
GET /api/reports/comparison
```
**Query Parameters:**
- `period1`: string
- `period2`: string
- `metrics`: string (comma-separated)

---

## 8. Takvim ve Etkinlik Yönetimi

### 8.1 Etkinlik Listesi
```http
GET /api/calendar/events
```
**Query Parameters:**
- `startDate`: ISO date string
- `endDate`: ISO date string
- `type`: string
- `status`: string

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "string",
      "title": "string",
      "description": "string",
      "type": "string", // "vaccination", "harvest", "maintenance"
      "startDate": "ISO date string",
      "endDate": "ISO date string",
      "isAllDay": boolean,
      "status": "string", // "pending", "completed", "cancelled"
      "priority": "string", // "low", "medium", "high"
      "location": "string",
      "relatedEntity": {
        "type": "string", // "land", "animal", "general"
        "id": "string",
        "name": "string"
      },
      "reminders": [
        {
          "time": number, // minutes before event
          "method": "string" // "notification", "email"
        }
      ]
    }
  ]
}
```

### 8.2 Etkinlik Detayları
```http
GET /api/calendar/events/{id}
```

### 8.3 Yeni Etkinlik Ekleme
```http
POST /api/calendar/events
```
**Request Body:**
```json
{
  "title": "string",
  "description": "string",
  "type": "string",
  "startDate": "ISO date string",
  "endDate": "ISO date string",
  "isAllDay": boolean,
  "priority": "string",
  "location": "string",
  "relatedEntity": {
    "type": "string",
    "id": "string"
  },
  "reminders": [
    {
      "time": number,
      "method": "string"
    }
  ]
}
```

### 8.4 Etkinlik Güncelleme
```http
PUT /api/calendar/events/{id}
```

### 8.5 Etkinlik Silme
```http
DELETE /api/calendar/events/{id}
```

### 8.6 Etkinlik Durumu Güncelleme
```http
PATCH /api/calendar/events/{id}/status
```
**Request Body:**
```json
{
  "status": "string",
  "notes": "string"
}
```

### 8.7 Takvim İstatistikleri
```http
GET /api/calendar/statistics
```
**Query Parameters:**
- `period`: string

**Response:**
```json
{
  "success": true,
  "data": {
    "totalEvents": number,
    "completedEvents": number,
    "pendingEvents": number,
    "todayEvents": number,
    "upcomingEvents": number,
    "eventsByType": [
      {
        "type": "string",
        "count": number
      }
    ]
  }
}
```

---

## 9. Bildirimler

### 9.1 Bildirim Listesi
```http
GET /api/notifications
```
**Query Parameters:**
- `page`: number
- `limit`: number
- `type`: string
- `read`: boolean

**Response:**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "string",
        "title": "string",
        "message": "string",
        "type": "string", // "reminder", "alert", "info", "warning"
        "priority": "string",
        "isRead": boolean,
        "createdAt": "ISO date string",
        "relatedEntity": {
          "type": "string",
          "id": "string",
          "name": "string"
        },
        "actions": [
          {
            "label": "string",
            "action": "string",
            "url": "string"
          }
        ]
      }
    ],
    "pagination": {
      "page": number,
      "limit": number,
      "total": number,
      "totalPages": number
    },
    "unreadCount": number
  }
}
```

### 9.2 Bildirim Okundu İşaretleme
```http
PATCH /api/notifications/{id}/read
```

### 9.3 Tüm Bildirimleri Okundu İşaretleme
```http
PATCH /api/notifications/mark-all-read
```

### 9.4 Bildirim Silme
```http
DELETE /api/notifications/{id}
```

### 9.5 Bildirim Ayarları
```http
GET /api/notifications/settings
```
**Response:**
```json
{
  "success": true,
  "data": {
    "pushNotifications": boolean,
    "emailNotifications": boolean,
    "smsNotifications": boolean,
    "notificationTypes": {
      "reminders": boolean,
      "alerts": boolean,
      "updates": boolean,
      "marketing": boolean
    },
    "quietHours": {
      "enabled": boolean,
      "startTime": "string",
      "endTime": "string"
    }
  }
}
```

### 9.6 Bildirim Ayarları Güncelleme
```http
PUT /api/notifications/settings
```

---

## 10. Ayarlar ve Konfigürasyon

### 10.1 Uygulama Ayarları
```http
GET /api/settings
```
**Response:**
```json
{
  "success": true,
  "data": {
    "general": {
      "language": "string",
      "currency": "string",
      "dateFormat": "string",
      "timeFormat": "string",
      "units": {
        "area": "string", // "hectare", "dönüm"
        "weight": "string", // "kg", "ton"
        "volume": "string" // "liter", "gallon"
      }
    },
    "notifications": {
      "push": boolean,
      "email": boolean,
      "sms": boolean
    },
    "privacy": {
      "locationSharing": boolean,
      "dataAnalytics": boolean,
      "personalizedAds": boolean
    },
    "backup": {
      "autoBackup": boolean,
      "backupFrequency": "string",
      "cloudStorage": boolean
    }
  }
}
```

### 10.2 Ayarları Güncelleme
```http
PUT /api/settings
```

### 10.3 Sistem Bilgileri
```http
GET /api/settings/system-info
```
**Response:**
```json
{
  "success": true,
  "data": {
    "appVersion": "string",
    "apiVersion": "string",
    "lastBackup": "ISO date string",
    "storageUsed": number,
    "storageLimit": number,
    "features": ["string"],
    "supportContact": "string"
  }
}
```

### 10.4 Veri Yedekleme
```http
POST /api/settings/backup
```

### 10.5 Veri Geri Yükleme
```http
POST /api/settings/restore
```
**Request Body:**
```json
{
  "backupFile": "string",
  "restoreOptions": {
    "includeFinance": boolean,
    "includeLivestock": boolean,
    "includeLands": boolean,
    "includeProduction": boolean
  }
}
```

---

## 11. Hava Durumu

### 11.1 Güncel Hava Durumu
```http
GET /api/weather/current
```
**Query Parameters:**
- `lat`: number
- `lon`: number

**Response:**
```json
{
  "success": true,
  "data": {
    "location": "string",
    "temperature": number,
    "humidity": number,
    "windSpeed": number,
    "windDirection": "string",
    "pressure": number,
    "visibility": number,
    "uvIndex": number,
    "condition": "string",
    "icon": "string",
    "lastUpdated": "ISO date string"
  }
}
```

### 11.2 Hava Durumu Tahmini (7 günlük)
```http
GET /api/weather/forecast
```
**Query Parameters:**
- `lat`: number
- `lon`: number
- `days`: number (default: 7)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "date": "ISO date string",
      "minTemp": number,
      "maxTemp": number,
      "condition": "string",
      "icon": "string",
      "humidity": number,
      "rainChance": number,
      "windSpeed": number
    }
  ]
}
```

### 11.3 Tarımsal Uyarılar
```http
GET /api/weather/agricultural-alerts
```
**Query Parameters:**
- `lat`: number
- `lon`: number

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "type": "string", // "frost", "drought", "flood", "storm"
      "severity": "string", // "low", "medium", "high"
      "title": "string",
      "description": "string",
      "startDate": "ISO date string",
      "endDate": "ISO date string",
      "recommendations": ["string"]
    }
  ]
}
```

---

## 📊 Veri Modelleri ve İlişkiler

### Temel Veri Yapıları:

1. **User (Kullanıcı)**
   - Profil bilgileri
   - Çiftlik bilgileri
   - Yetkilendirme

2. **Land (Arazi)**
   - Arazi özellikleri
   - Konum bilgileri
   - Üretim geçmişi

3. **Livestock (Hayvan)**
   - Hayvan bilgileri
   - Sağlık kayıtları
   - Üretim kayıtları

4. **Production (Üretim)**
   - Ürün bilgileri
   - Hasat verileri
   - Kalite metrikleri

5. **Transaction (Finansal İşlem)**
   - Gelir/gider kayıtları
   - Kategorizasyon
   - Ödeme detayları

6. **Event (Etkinlik)**
   - Takvim etkinlikleri
   - Hatırlatıcılar
   - Durum takibi

7. **Notification (Bildirim)**
   - Sistem bildirimleri
   - Kullanıcı tercihleri
   - Aksiyon öğeleri

---

## 🔐 Güvenlik Gereksinimleri

1. **JWT Token Authentication**
2. **Role-based Access Control**
3. **Data Encryption**
4. **API Rate Limiting**
5. **Input Validation**
6. **CORS Configuration**
7. **HTTPS Only**

---

## 📱 Response Format Standardı

Tüm API yanıtları aşağıdaki formatta olmalıdır:

```json
{
  "success": boolean,
  "data": object|array,
  "message": "string",
  "error": {
    "code": "string",
    "message": "string",
    "details": object
  },
  "meta": {
    "timestamp": "ISO date string",
    "version": "string",
    "requestId": "string"
  }
}
```

---

## 🚀 Öncelik Sırası

### Yüksek Öncelik:
1. Kimlik doğrulama ve kullanıcı yönetimi
2. Dashboard özet verileri
3. Temel CRUD işlemleri (Arazi, Hayvan, Finans)

### Orta Öncelik:
4. Raporlar ve analizler
5. Takvim ve etkinlik yönetimi
6. Bildirimler

### Düşük Öncelik:
7. Hava durumu entegrasyonu
8. Gelişmiş analizler
9. Yedekleme/geri yükleme

Bu API dokümantasyonu, Flutter uygulamanızın tüm özelliklerini destekleyecek şekilde tasarlanmıştır. Her endpoint'in detaylı açıklaması, request/response formatları ve gerekli parametreler belirtilmiştir.
