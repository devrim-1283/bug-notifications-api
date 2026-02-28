# Bug Notifications API

Sitelerden gelen bug bildirimlerini toplayan, Redis queue uzerinden isleyip PostgreSQL'e kaydeden minimal ve guvenli bir mikro servis.

## Mimari

```
Browser -> Sitenizin Backend'i (BFF) -> Bug Notifications API -> Redis Queue -> Worker -> PostgreSQL
```

| Katman | Gorev |
|--------|-------|
| **API** | Istegi alir, dogrular, kuyruga yazar |
| **Redis Queue** | Mesajlari tamponlar, ani trafik yukunu emer |
| **Worker** | Kuyruktan okur, PostgreSQL'e yazar |
| **DLQ** | Basarisiz mesajlar dead-letter queue'ya alinir |

## Ozellikler

- **Tek endpoint:** `POST /v1/reports` (202 Accepted, asenkron isleme)
- **Health check:** `GET /health`
- **Coklu resim yukleme:** Opsiyonel, max 5 resim (jpg/png/webp/gif, her biri max 5MB)
- **5 katmanli guvenlik:** CORS + API Key + Origin eslesmesi + Browser-only + Rate limit
- **Retry & DLQ:** Basarisiz DB insertlerde 5 deneme, sonra DLQ
- **Idempotency:** `event_id` uzerinden duplicate kayit onleme
- **Yatay olceklenebilir:** Worker sayisi `WORKER_CONCURRENCY` ile ayarlanabilir

## Guvenlik

| Katman | Koruma |
|--------|--------|
| CORS | Sadece kayitli site domain'lerine izin verir |
| API Key | `X-API-Key` header ile dogrulama |
| Origin Match | API key hangi domain'e aitse istek o domain'den gelmeli |
| Browser-Only | `Sec-Fetch-Site` + `User-Agent` kontrolu (Postman/curl engellenir) |
| Rate Limit | IP bazli token bucket (varsayilan: 10 req/s) |

> **API Key Guvenligi:** Her API key bir domain'e kilitlidir. Dogrudan frontend'den kullanilsa bile baska bir domain'den ayni key ile istek atilamaz (Origin + CORS + Sec-Fetch kontrolleri). Ekstra gizlilik icin BFF (Backend For Frontend) deseni kullanilabilir — bu durumda API key sadece sunucunuzda kalir ve kullaniciya asla acilmaz.

## Gereksinimler

- Go 1.25+
- Redis
- PostgreSQL

## Kurulum

```bash
# Repoyu klonla
git clone <repo-url>
cd bug-notifications-api

# Env dosyasini olustur
cp .env.example .env
# .env dosyasini duzenle (DATABASE_URL, REDIS_URL, SITE_KEYS)

# API key olustur (her site icin)
openssl rand -hex 32

# Veritabanini hazirla
psql $DATABASE_URL -f migrations/001_create_bug_reports.sql

# Bagimliklar
go mod download
```

## Calistirma

### Lokal

```bash
# API sunucusu
go run ./cmd/api

# Worker (ayri terminal)
go run ./cmd/worker
```

### Docker

```bash
docker build -t bug-notifications-api .
docker run -p 3000:3000 --env-file .env bug-notifications-api
```

### Coolify ile Deploy

1. Coolify'da yeni bir proje olusturun
2. **Git Repository** olarak bu repoyu baglayin
3. Build Pack: **Dockerfile**
4. Port: **3000** (otomatik algilanir)
5. Ortam degiskenlerini Coolify UI uzerinden ayarlayin:
   - `DATABASE_URL` — PostgreSQL baglanti adresi
   - `REDIS_URL` — Redis baglanti adresi
   - `SITE_KEYS` — domain:key ciftleri
6. Deploy!

**MODE degiskeni ile calistirma modu:**

| Deger | Aciklama |
|-------|----------|
| `all` (varsayilan) | API + Worker ayni container'da calisir |
| `api` | Sadece API sunucusu |
| `worker` | Sadece Worker |

> **Tavsiye:** Kucuk/orta olcekte `MODE=all` yeterlidir. Yuksek trafik icin API ve Worker'i ayri Coolify servisleri olarak deploy edin (`MODE=api` ve `MODE=worker`).

## Ortam Degiskenleri

| Degisken | Varsayilan | Aciklama |
|----------|-----------|----------|
| `PORT` | `3000` | API sunucu portu |
| `REDIS_URL` | `redis://localhost:6379` | Redis baglanti adresi |
| `DATABASE_URL` | _(zorunlu)_ | PostgreSQL baglanti adresi |
| `SITE_KEYS` | _(zorunlu)_ | `domain:key` ciftleri, virgul ile ayrilmis |
| `RATE_LIMIT_RPS` | `10` | IP basina saniyede max istek |
| `WORKER_CONCURRENCY` | `10` | Paralel worker sayisi |
| `MODE` | `all` | `all` / `api` / `worker` |
| `TLS_CERT_FILE` | _(opsiyonel)_ | TLS sertifika dosyasi |
| `TLS_KEY_FILE` | _(opsiyonel)_ | TLS private key dosyasi |
| `TRUSTED_PROXIES` | _(opsiyonel)_ | Guvenilir proxy CIDR araliklari |
| `IMAGE_API_URL` | _(opsiyonel)_ | Resim API base URL'i |
| `IMAGE_API_KEY` | _(opsiyonel)_ | Resim API anahtari |

**SITE_KEYS ornegi:**
```
SITE_KEYS=example.com:a1b2c3d4e5f6,mysite.io:x9y8z7w6v5u4
```

## Proje Yapisi

```
cmd/
  api/           API sunucu entrypoint
  worker/        Worker entrypoint
internal/
  api/           HTTP handler'lar
  config/        Konfigurason yukleyici
  db/            PostgreSQL baglanti ve repository
  middleware/    CORS, auth, rate limit, browser-only
  model/         Veri modelleri
  queue/         Redis producer/consumer
  validate/      Input dogrulama
  worker/        Worker isleme mantigi
migrations/      SQL migration dosyalari
```

## API Kullanimi

### Bug Bildirimi Gonder

```
POST /v1/reports
```

**Header'lar:**
```
Content-Type: application/json
X-API-Key: <site-api-key>
```

**Body:**
```json
{
  "site_id": "example.com",
  "title": "Login butonu calismiyor",
  "description": "Safari'de tiklandiginda hicbir sey olmuyor",
  "category": "functionality",
  "page_url": "https://example.com/login",
  "contact_type": "email",
  "contact_value": "ali@example.com",
  "first_name": "Ali",
  "last_name": "Yilmaz"
}
```

**Zorunlu alanlar:** `site_id`, `title`, `description`, `category`

**Gecerli kategoriler:** `design`, `functionality`, `performance`, `content`, `mobile`, `security`, `other`

**Basarili yanit (202):**
```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "queued": true
}
```

### Resimli Bug Bildirimi (multipart/form-data)

Resim eklemek icin `multipart/form-data` kullanin. `images` field'ina birden fazla dosya eklenebilir.

```javascript
const form = new FormData();
form.append('title', 'Login butonu calismiyor');
form.append('description', 'Safari de tiklandiginda beyaz ekran geliyor');
form.append('category', 'functionality');
form.append('page_url', 'https://example.com/login');
form.append('images', dosya1); // 1. resim
form.append('images', dosya2); // 2. resim

await fetch('https://bug.devrimsoft.com/v1/reports', {
  method: 'POST',
  headers: { 'X-API-Key': 'site-api-key' },
  body: form
});
```

**Resim kurallari:**

| Kural | Deger |
|-------|-------|
| Max dosya sayisi | 5 |
| Max dosya boyutu | 5MB (her biri) |
| Gecerli formatlar | jpg, png, webp, gif |
| Dogrulama | Uzanti + magic bytes (icerik dogrulama) |

### Saglik Kontrolu

```
GET /health
```

```json
{
  "status": "ok"
}
```

## Resim Depolama (Image API)

Resimler DevrimSoft'un kendi **R2 Image Processor API**'si uzerinden islenir ve depolanir (`view.devrimsoft.com`). Bu, bu projeye ozel gelistirilmis ayri bir mikro servistir.

| Endpoint | Aciklama |
|----------|----------|
| `POST /upload` | Resim yukler, otomatik WebP'ye donusturur, R2'ye kaydeder |
| `DELETE /delete` | Yuklenmis resmi siler |
| `GET /health` | API ve R2 baglanti durumu |

**Ozellikler:**
- Otomatik WebP donusumu (kalite: 85, max 1920px genislik)
- EXIF/GPS metadata temizleme (gizlilik)
- API key bazli klasor izolasyonu
- UUID dosya adlari (cakisma onleme)

Resim API'si bu servis tarafindan dahili olarak kullanilir. Kullanici dogrudan erisemez — resimler bug report istegi icinde dosya olarak gonderilir, API otomatik olarak R2'ye yukler ve URL'leri DB'de saklar.

```
IMAGE_API_URL=https://view.devrimsoft.com
IMAGE_API_KEY=your-r2-api-key
```

> Bu degiskenler ayarlanmazsa resim yukleme devre disi kalir, diger tum ozellikler calismaya devam eder.

## Yazar

**Devrim Tuncer** tarafindan gelistirilmistir.

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Devrim%20Tun%C3%A7er-blue?logo=linkedin)](https://www.linkedin.com/in/devrim-tun%C3%A7er-218a55320/)

Bu proje [devrimsoft.com](https://devrimsoft.com) urunudur ve canli projelerde kullanilmaktadir.

Bu projenin gelistirilmesinde AI araclari kullanilmistir.

## Lisans

MIT

---

# Bug Notifications API (English)

A minimal and secure microservice that collects bug reports from websites, processes them through a Redis queue, and persists them to PostgreSQL.

## Architecture

```
Browser -> Your Site's Backend (BFF) -> Bug Notifications API -> Redis Queue -> Worker -> PostgreSQL
```

| Layer | Role |
|-------|------|
| **API** | Receives requests, validates, writes to queue |
| **Redis Queue** | Buffers messages, absorbs traffic spikes |
| **Worker** | Reads from queue, writes to PostgreSQL |
| **DLQ** | Failed messages are moved to dead-letter queue |

## Features

- **Single endpoint:** `POST /v1/reports` (202 Accepted, async processing)
- **Health check:** `GET /health`
- **Multiple image upload:** Optional, up to 5 images (jpg/png/webp/gif, max 5MB each)
- **5-layer security:** CORS + API Key + Origin matching + Browser-only + Rate limit
- **Retry & DLQ:** 5 retry attempts on failed DB inserts, then DLQ
- **Idempotency:** Duplicate prevention via `event_id`
- **Horizontally scalable:** Worker count adjustable via `WORKER_CONCURRENCY`

## Security

| Layer | Protection |
|-------|------------|
| CORS | Only allows registered site domains |
| API Key | Validation via `X-API-Key` header |
| Origin Match | API key must match the requesting domain |
| Browser-Only | `Sec-Fetch-Site` + `User-Agent` checks (blocks Postman/curl) |
| Rate Limit | IP-based token bucket (default: 10 req/s) |

> **API Key Security:** Each API key is locked to a domain. Even when used directly from the frontend, the same key cannot be used from a different domain (Origin + CORS + Sec-Fetch checks). For extra privacy, you can use the BFF (Backend For Frontend) pattern — in this case, the API key stays only on your server and is never exposed to the user.

## Requirements

- Go 1.25+
- Redis
- PostgreSQL

## Setup

```bash
# Clone the repo
git clone <repo-url>
cd bug-notifications-api

# Create env file
cp .env.example .env
# Edit .env (DATABASE_URL, REDIS_URL, SITE_KEYS)

# Generate API key (for each site)
openssl rand -hex 32

# Prepare database
psql $DATABASE_URL -f migrations/001_create_bug_reports.sql

# Dependencies
go mod download
```

## Running

### Local

```bash
# API server
go run ./cmd/api

# Worker (separate terminal)
go run ./cmd/worker
```

### Docker

```bash
docker build -t bug-notifications-api .
docker run -p 3000:3000 --env-file .env bug-notifications-api
```

### Deploy with Coolify

1. Create a new project in Coolify
2. Connect this repo as **Git Repository**
3. Build Pack: **Dockerfile**
4. Port: **3000** (auto-detected)
5. Set environment variables via Coolify UI:
   - `DATABASE_URL` — PostgreSQL connection string
   - `REDIS_URL` — Redis connection string
   - `SITE_KEYS` — domain:key pairs
6. Deploy!

**Running mode via MODE variable:**

| Value | Description |
|-------|-------------|
| `all` (default) | API + Worker run in the same container |
| `api` | API server only |
| `worker` | Worker only |

> **Recommendation:** For small/medium scale, `MODE=all` is sufficient. For high traffic, deploy API and Worker as separate Coolify services (`MODE=api` and `MODE=worker`).

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | API server port |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `DATABASE_URL` | _(required)_ | PostgreSQL connection string |
| `SITE_KEYS` | _(required)_ | `domain:key` pairs, comma separated |
| `RATE_LIMIT_RPS` | `10` | Max requests per second per IP |
| `WORKER_CONCURRENCY` | `10` | Number of parallel workers |
| `MODE` | `all` | `all` / `api` / `worker` |
| `TLS_CERT_FILE` | _(optional)_ | TLS certificate file |
| `TLS_KEY_FILE` | _(optional)_ | TLS private key file |
| `TRUSTED_PROXIES` | _(optional)_ | Trusted proxy CIDR ranges |
| `IMAGE_API_URL` | _(optional)_ | Image API base URL |
| `IMAGE_API_KEY` | _(optional)_ | Image API key |

**SITE_KEYS example:**
```
SITE_KEYS=example.com:a1b2c3d4e5f6,mysite.io:x9y8z7w6v5u4
```

## Project Structure

```
cmd/
  api/           API server entrypoint
  worker/        Worker entrypoint
internal/
  api/           HTTP handlers
  config/        Configuration loader
  db/            PostgreSQL connection and repository
  middleware/    CORS, auth, rate limit, browser-only
  model/         Data models
  queue/         Redis producer/consumer
  validate/      Input validation
  worker/        Worker processing logic
migrations/      SQL migration files
```

## API Usage

### Submit Bug Report

```
POST /v1/reports
```

**Headers:**
```
Content-Type: application/json
X-API-Key: <site-api-key>
```

**Body:**
```json
{
  "site_id": "example.com",
  "title": "Login button not working",
  "description": "Nothing happens when clicking the login button on Safari",
  "category": "functionality",
  "page_url": "https://example.com/login",
  "contact_type": "email",
  "contact_value": "ali@example.com",
  "first_name": "Ali",
  "last_name": "Yilmaz"
}
```

**Required fields:** `site_id`, `title`, `description`, `category`

**Valid categories:** `design`, `functionality`, `performance`, `content`, `mobile`, `security`, `other`

**Successful response (202):**
```json
{
  "event_id": "550e8400-e29b-41d4-a716-446655440000",
  "queued": true
}
```

### Bug Report with Images (multipart/form-data)

To attach images, use `multipart/form-data`. Multiple files can be added to the `images` field.

```javascript
const form = new FormData();
form.append('title', 'Login button not working');
form.append('description', 'White screen appears when clicking login on Safari');
form.append('category', 'functionality');
form.append('page_url', 'https://example.com/login');
form.append('images', file1); // 1st image
form.append('images', file2); // 2nd image

await fetch('https://bug.devrimsoft.com/v1/reports', {
  method: 'POST',
  headers: { 'X-API-Key': 'site-api-key' },
  body: form
});
```

**Image rules:**

| Rule | Value |
|------|-------|
| Max file count | 5 |
| Max file size | 5MB (each) |
| Allowed formats | jpg, png, webp, gif |
| Validation | Extension + magic bytes (content verification) |

### Health Check

```
GET /health
```

```json
{
  "status": "ok"
}
```

## Image Storage (Image API)

Images are processed and stored via DevrimSoft's own **R2 Image Processor API** (`view.devrimsoft.com`). This is a separate microservice built specifically for this ecosystem.

| Endpoint | Description |
|----------|-------------|
| `POST /upload` | Uploads image, auto-converts to WebP, stores in R2 |
| `DELETE /delete` | Deletes an uploaded image |
| `GET /health` | API and R2 connection status |

**Features:**
- Automatic WebP conversion (quality: 85, max 1920px width)
- EXIF/GPS metadata stripping (privacy)
- API key-based folder isolation
- UUID filenames (collision prevention)

The Image API is used internally by this service. Users cannot access it directly — images are sent as files within the bug report request, and the API automatically uploads them to R2 and stores the URLs in the database.

```
IMAGE_API_URL=https://view.devrimsoft.com
IMAGE_API_KEY=your-r2-api-key
```

> If these variables are not set, image upload is disabled. All other features continue to work normally.

## Author

Developed by **Devrim Tuncer**.

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Devrim%20Tun%C3%A7er-blue?logo=linkedin)](https://www.linkedin.com/in/devrim-tun%C3%A7er-218a55320/)

This project is a product of [devrimsoft.com](https://devrimsoft.com) and is used in production.

AI tools were used in the development of this project.

## License

MIT
