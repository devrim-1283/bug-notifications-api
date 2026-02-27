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

### Saglik Kontrolu

```
GET /health
```

```json
{
  "status": "ok"
}
```

## Yazar

**Devrim Tuncer** tarafindan gelistirilmistir.

[![LinkedIn](https://img.shields.io/badge/LinkedIn-Devrim%20Tun%C3%A7er-blue?logo=linkedin)](https://www.linkedin.com/in/devrim-tun%C3%A7er-218a55320/)

Bu proje [devrimsoft.com](https://devrimsoft.com) urunudur ve canli projelerde kullanilmaktadir.

Bu projenin gelistirilmesinde AI araclari kullanilmistir.

## Lisans

MIT
