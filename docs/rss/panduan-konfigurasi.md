# Panduan Konfigurasi RSS Scraping

Dokumentasi lengkap untuk konfigurasi RSS scraping di sistem SIPON.

## Daftar Isi

1. [Pendahuluan](#pendahuluan)
2. [Struktur RSS Feed](#struktur-rss-feed)
3. [Struktur HTML Artikel](#struktur-html-artikel)
4. [Panduan Konfigurasi](#panduan-konfigurasi)
5. [Field Mapping Otomatis](#field-mapping-otomatis)
6. [Tips dan Best Practices](#tips-dan-best-practices)
7. [Troubleshooting](#troubleshooting)
8. [Studi Kasus](#studi-kasus)

---

## Pendahuluan

Sistem scraping RSS memungkinkan kita untuk mengambil artikel dari website lain secara otomatis. Proses scraping terdiri dari 2 tahap:

1. **Fetch RSS Feed** - Mengambil daftar artikel dari URL RSS
2. **Scrape Detail** - Mengunjungi setiap URL artikel untuk mengambil konten lengkap

### Flow Scraping

```
POST /api/v1/web/article-sources/{source_id}/scrape
    ↓
Load source + categories aktif
    ↓
Untuk setiap kategori:
    1. Build URL RSS (base_url + url_suffix atau url_override)
    2. Fetch RSS feed → dapatkan daftar item
    3. Filter by keywords (jika ada)
    4. Deduplication by original_url
    5. Untuk setiap item baru:
       - Fetch detail page
       - Extract content (readability atau CSS selector)
       - Extract author, tags
       - Truncate content (max 100 kata)
       - Save article ke database
    6. Update last_scraped_at
    ↓
Return hasil scraping
```

---

## Struktur RSS Feed

RSS feed biasanya mengikuti standar RSS 2.0 atau Atom. Field yang umum ditemukan:

### Field Standar RSS

| Field | Deskripsi | Contoh |
|-------|-----------|--------|
| `title` | Judul artikel | "Menapaktilasi Jejak Dakwah Ulama" |
| `link` | URL lengkap artikel | `https://example.com/artikel/123` |
| `description` | Ringkasan/excerpt | "Semarang, KGS Media..." |
| `pubDate` | Tanggal publish | `Tue, 07 Jul 2026 13:25:58 +0000` |
| `category` | Kategori artikel | `<category>KGS-News</category>` |

### Field WordPress (dc namespace)

| Field | Deskripsi | Contoh |
|-------|-----------|--------|
| `dc:creator` | Nama penulis | `<dc:creator>Tim Redaksi</dc:creator>` |
| `content:encoded` | Konten HTML lengkap | `<content:encoded><![CDATA[...]]></content:encoded>` |

### Field Gambar/Thumbnail

Thumbnail bisa ditemukan di beberapa tempat:
1. **Di RSS feed** - `<enclosure url="...">` atau `<media:content url="...">`
2. **Di detail page** - `<img>` pertama di content
3. **Di meta tag** - `<meta property="og:image" content="...">`

Sistem akan otomatis detect dan menggunakan thumbnail yang tersedia.

---

## Struktur HTML Artikel

Ketika scraping detail page, sistem perlu mengekstrak:
1. **Konten utama** - Teks artikel
2. **Author** - Nama penulis
3. **Tags** - Tag/label artikel (opsional)
4. **Thumbnail** - URL gambar utama

### Selector Umum

Berikut selector yang umum bekerja di berbagai website:

#### Konten
- `.entry-content` (WordPress standar)
- `article .content`
- `.post-content`
- `.article-body`
- `[itemprop=articleBody]`

#### Author
- `.author`
- `.byline`
- `.post-author`
- `meta[name=author]`
- `.vcard .fn`

#### Tags
- `.tags a`
- `.tag-links a`
- `.post-tags a`
- `[href*='/tag/']`

### Readability Mode

Jika CSS selector tidak diisi atau tidak bekerja, sistem akan menggunakan **readability** untuk extract konten secara otomatis. Readability sangat efektif untuk:
- Website berita standar
- WordPress
- Blog platform populer

**Kapan perlu custom selector:**
- Website dengan struktur HTML unik
- Konten tersembunyi di dalam modal/popup
- Website dengan banyak sidebar/ads yang mengganggu

---

## Panduan Konfigurasi

### 1. Buat Source Baru

Buka menu **Sumber RSS** → klik **Tambah Sumber**

#### Field Wajib

| Field | Deskripsi | Contoh |
|-------|-----------|--------|
| **Nama** | Nama tampilan source | "Pondok Pesantren Kyai Galang Sewu" |
| **Key** | Identifier unik (lowercase + underscore) | `kgs` |
| **Base URL** | URL dasar website (tanpa path RSS) | `https://kyaigalangsewu.net` |

#### Field Opsional

| Field | Deskripsi | Default |
|-------|-----------|---------|
| **Auto Publish** | Otomatis publish artikel hasil scrape | `false` |
| **Status** | Aktif/nonaktif | `true` |

#### CSS Selector (Opsional)

| Field | Deskripsi | Kapan Digunakan |
|-------|-----------|-----------------|
| **Content Selector** | Selector untuk konten utama | Jika readability tidak bekerja |
| **Author Selector** | Selector untuk nama penulis | Jika ingin extract author spesifik |
| **Tags Selector** | Selector untuk tag/label | Jika ingin extract tags |

**Catatan:** 
- Kosongkan jika tidak yakin → sistem akan gunakan readability
- Test dulu dengan browser DevTools untuk memastikan selector bekerja

---

### 2. Tambah Kategori Source

Setelah source dibuat, klik **Tambah** di bagian Kategori

#### Field Wajib

| Field | Deskripsi | Contoh |
|-------|-----------|--------|
| **Key** | Identifier kategori di source ini | `all`, `nasional`, `olahraga` |

#### Field URL

Pilih salah satu:

| Field | Deskripsi | Contoh |
|-------|-----------|--------|
| **URL Suffix** | Ditambahkan ke base URL | `/index.php/feed/` |
| **URL Override** | URL full (menggantikan base URL) | `https://example.com/rss/kategori/sport` |

**Hasil URL:**
```
Jika pakai URL Suffix:
  Base URL: https://kyaigalangsewu.net
  + URL Suffix: /index.php/feed/
  = https://kyaigalangsewu.net/index.php/feed/

Jika pakai URL Override:
  URL Override: https://example.com/custom/feed.xml
  = https://example.com/custom/feed.xml
```

#### Field Konfigurasi

| Field | Deskripsi | Default | Contoh |
|-------|-----------|---------|--------|
| **Article Limit** | Maksimal artikel per scrape | `10` | `50` |
| **Map Kategori** | Mapping ke kategori artikel di sistem | (kosong) | Pilih "Berita Pondok" |
| **Keywords** | Filter kata kunci (comma separated) | (kosong) | `ziarah, haul, santri` |
| **Status** | Aktif/nonaktif | `true` | `false` |

---

### 3. Test Scrape

Setelah konfigurasi selesai:

1. Klik tombol **Scrape Sekarang** pada source
2. Tunggu proses selesai (biasanya 10-60 detik tergantung jumlah artikel)
3. Cek hasil di console log atau response API
4. Cek di menu **Kelola Artikel** → artikel baru seharusnya muncul

**Contoh Response:**
```json
{
  "status": "success",
  "message": "Scrape selesai",
  "data": {
    "source_key": "kgs",
    "categories": [
      {
        "category_key": "all",
        "total_items": 15,
        "new_items": 10,
        "duplicates": 5
      }
    ]
  }
}
```

---

## Field Mapping Otomatis

Sistem akan otomatis mapping field dari RSS ke database:

| RSS/HTML Field | Database Field | Sumber |
|----------------|----------------|--------|
| `title` | `articles.title` | RSS feed |
| `link` | `articles.original_url` | RSS feed |
| `content:encoded` | `articles.content` | Detail page (readability/selector) |
| `description` | `articles.summary` | RSS feed |
| `dc:creator` / author selector | `articles.author` | RSS feed atau detail page |
| `pubDate` | `articles.published_at` | RSS feed |
| Thumbnail (img pertama) | `articles.thumbnail_url` | Detail page |
| Category mapping | `articles.category_id` | Konfigurasi source category |

### Deduplication

Sistem menggunakan `original_url` untuk deduplication:
- Artikel dengan URL yang sama tidak akan di-scrape 2x
- Jika artikel sudah ada, akan di-skip otomatis
- Counter `duplicates` di response menunjukkan jumlah artikel yang di-skip

---

## Tips dan Best Practices

### 1. Filter Keywords

Gunakan filter keywords jika hanya ingin scrape artikel tertentu:

**Contoh:**
```
Keywords: "ziarah, haul, santri"
```

**Behavior:**
- Hanya artikel yang mengandung salah satu keyword di title atau description yang akan di-scrape
- Case-insensitive
- Pisahkan dengan koma

**Kapan gunakan:**
- Website dengan banyak kategori artikel
- Hanya tertarik pada topik tertentu
- Menghindari artikel yang tidak relevan

### 2. Article Limit

**Rekomendasi:**
- **Test/Development:** `10` → Cepat, cocok untuk testing
- **Daily Update:** `20-50` → Cukup untuk update harian
- **Initial Import:** `100-500` → Untuk import artikel lama

**Catatan:**
- Limit lebih besar = scraping lebih lama
- Limit lebih besar = load server lebih tinggi
- Mulai dari limit kecil, naikkan jika perlu

### 3. Auto Publish

**Aktif (true):**
- Artikel langsung published
- Muncul di halaman publik
- Cocok untuk source terpercaya

**Nonaktif (false):**
- Artikel sebagai draft
- Perlu review manual sebelum publish
- Cocok untuk source yang perlu kurasi

### 4. Multi-Feed Configuration

Jika website memiliki multiple feeds (berdasarkan kategori), buat multiple source category:

**Contoh:**
```
Source: CNN Indonesia
├── Category: nasional
│   URL Suffix: /feed/nasional
│   Map: Berita Nasional
│
├── Category: internasional
│   URL Suffix: /feed/internasional
│   Map: Berita Internasional
│
└── Category: olahraga
    URL Suffix: /feed/olahraga
    Map: Berita Olahraga
```

**Cara temukan URL feed per kategori:**
1. Buka halaman kategori di website
2. Cari link RSS/Feed (biasanya di footer atau sidebar)
3. Atau coba pattern umum:
   - `/category/{nama}/feed/`
   - `/feed/{nama}`
   - `?feed=rss&cat={id}`

### 5. CSS Selector Best Practices

**Test selector dulu:**
```javascript
// Di browser console
document.querySelector('.entry-content')
document.querySelectorAll('.tags a')
```

**Prioritas:**
1. Kosongkan (gunakan readability) → paling mudah
2. Gunakan selector umum (`.entry-content`) → biasanya bekerja
3. Custom selector spesifik → hanya jika perlu

**Hindari:**
- Selector yang terlalu spesifik (bisa berubah)
- Selector yang bergantung pada ID dinamis
- Selector yang mencakup sidebar/ads

---

## Troubleshooting

### ❌ RSS Feed Tidak Bisa Diakses

**Gejala:**
```
Error: fetch feed: connection timeout
Error: feed returned status 403
```

**Solusi:**
1. Cek URL dengan browser → pastikan bisa diakses
2. Cek apakah website memblokir IP server
3. Coba akses dengan curl: `curl https://example.com/feed/`
4. Jika 403, website mungkin butuh User-Agent header (sudah otomatis di-set)

---

### ❌ Konten Kosong atau Tidak Lengkap

**Gejala:**
- Artikel berhasil di-scrape tapi content kosong
- Content hanya berisi 1-2 kalimat

**Solusi:**
1. **Cek readability** → buka detail page, lihat apakah konten bisa di-extract
2. **Tambah content selector** → coba `.entry-content` atau selector lain
3. **Cek struktur HTML** → buka DevTools, inspect element konten
4. **Test dengan curl:**
   ```bash
   curl https://example.com/artikel/123 | grep -E '<article|<div class="content"'
   ```

---

### ❌ Author Tidak Terdeteksi

**Gejala:**
- Field author kosong atau "Unknown"

**Solusi:**
1. **Cek RSS feed** → apakah ada `<dc:creator>` atau `<author>`
2. **Cek detail page** → cari elemen author dengan DevTools
3. **Tambah author selector:**
   - `meta[name=author]` → ambil dari meta tag
   - `.author` → class umum untuk author
   - `.byline` → class umum untuk byline
   - `.vcard .fn` → microformat untuk author

---

### ❌ Thumbnail Tidak Muncul

**Gejala:**
- Artikel tidak punya thumbnail

**Solusi:**
1. **Cek apakah ada gambar** di content → harus ada minimal 1 `<img>`
2. **Cek URL gambar** → harus absolute URL (bukan relative)
3. **Cek format** → harus jpg, png, gif, atau webp
4. **Manual fix** → edit artikel, tambahkan thumbnail URL

---

### ❌ Duplicate Articles

**Gejala:**
- Artikel yang sama muncul beberapa kali

**Penyebab:**
- URL berbeda untuk artikel yang sama (misal: dengan/without trailing slash)
- RSS feed mengembalikan artikel yang sudah ada

**Solusi:**
1. **Check deduplication** → sistem sudah cek `original_url`
2. **Normalize URL** → pastikan URL konsisten (tanpa query params yang tidak perlu)
3. **Manual cleanup** → hapus duplicate di database

---

### ❌ Scraping Sangat Lambat

**Gejala:**
- Scrape memakan waktu > 5 menit

**Penyebab:**
- Article limit terlalu besar
- Website target lambat
- Banyak artikel baru yang perlu di-scrape

**Solusi:**
1. **Kurangi article limit** → mulai dari 10-20
2. **Gunakan keywords filter** → hanya scrape artikel relevan
3. **Scrape per kategori** → buat multiple category dengan limit kecil
4. **Check website** → apakah website target memang lambat

---

### ❌ Content Terlalu Panjang

**Gejala:**
- Artikel hasil scrape sangat panjang (ribuan kata)

**Penyebab:**
- Website target punya artikel panjang
- Readability tidak memfilter dengan baik

**Solusi:**
1. **System sudah truncate** → max 100 kata (sudah di-set di backend)
2. **Custom selector** → gunakan selector yang lebih spesifik
3. **Manual edit** → edit artikel hasil scrape, potong bagian yang tidak perlu

---

## Studi Kasus

### Studi Kasus 1: Pondok Pesantren Kyai Galang Sewu

**Website:** `https://kyaigalangsewu.net`

**Analisis RSS:**
- URL Feed: `https://kyaigalangsewu.net/index.php/feed/`
- Format: WordPress RSS 2.0
- Author: `<dc:creator>Tim Redaksi</dc:creator>`
- Categories: `<category>KGS-News</category>`, `<category>kyai galang sewu</category>`
- Content: `<content:encoded>` dengan HTML lengkap + gambar
- Thumbnail: `<img>` pertama di content

**Konfigurasi:**

**Source:**
```
Nama: Pondok Pesantren Kyai Galang Sewu
Key: kgs
Base URL: https://kyaigalangsewu.net
Auto Publish: true
```

**Source Category:**
```
Key: all
URL Suffix: /index.php/feed/
Article Limit: 10
Map Kategori: Berita Pondok
Keywords: (kosong)
```

**CSS Selector:**
```
Content Selector: .entry-content (opsional, readability sudah cukup)
Author Selector: .author.vcard a (opsional, dc:creator sudah ada)
Tags Selector: (kosong)
```

**Hasil:**
- Berhasil scrape 10 artikel terbaru
- Author terdeteksi: "Tim Redaksi"
- Thumbnail terdeteksi dari gambar pertama
- Content lengkap dengan formatting HTML

---

### Studi Kasus 2: CNN Indonesia

**Website:** `https://www.cnnindonesia.com`

**Analisis RSS:**
- Multiple feeds per kategori
- URL Feed Nasional: `https://www.cnnindonesia.com/nasional/feed`
- URL Feed Internasional: `https://www.cnnindonesia.com/internasional/feed`
- Author: Tidak ada di RSS, harus scrape dari detail page
- Thumbnail: Ada di RSS sebagai `<media:content>`

**Konfigurasi:**

**Source:**
```
Nama: CNN Indonesia
Key: cnn_id
Base URL: https://www.cnnindonesia.com
Auto Publish: false (perlu review)
```

**Source Categories:**
```
Category 1:
  Key: nasional
  URL Suffix: /nasional/feed
  Article Limit: 20
  Map Kategori: Berita Nasional
  Author Selector: .author_name

Category 2:
  Key: internasional
  URL Suffix: /internasional/feed
  Article Limit: 20
  Map Kategori: Berita Internasional
  Author Selector: .author_name
```

---

### Studi Kasus 3: Blog WordPress Sederhana

**Website:** `https://blog.example.com`

**Analisis RSS:**
- Standard WordPress RSS
- URL Feed: `https://blog.example.com/feed/`
- Author: Ada di RSS sebagai `<dc:creator>`
- Content: Lengkap di `<content:encoded>`
- Thumbnail: Tidak ada di RSS, harus scrape dari detail page

**Konfigurasi:**

**Source:**
```
Nama: Example Blog
Key: example_blog
Base URL: https://blog.example.com
Auto Publish: true
```

**Source Category:**
```
Key: all
URL Suffix: /feed/
Article Limit: 50
Map Kategori: Blog
```

**CSS Selector:**
```
(kosongkan semua, gunakan readability)
```

---

## API Reference

### Endpoints

#### List Sources
```
GET /api/v1/web/article-sources
```

#### Create Source
```
POST /api/v1/web/article-sources
Body: {
  "name": "string",
  "key": "string",
  "base_url": "string",
  "auto_publish": boolean,
  "is_active": boolean
}
```

#### Update Source
```
PUT /api/v1/web/article-sources/{id}
Body: {
  "name": "string",
  "base_url": "string",
  "auto_publish": boolean,
  "is_active": boolean
}
```

#### Delete Source
```
DELETE /api/v1/web/article-sources/{id}
```

#### Create Source Category
```
POST /api/v1/web/article-sources/{source_id}/categories
Body: {
  "category_key": "string",
  "url_suffix": "string",
  "url_override": "string",
  "article_limit": number,
  "article_category_id": "string",
  "keywords": ["string"],
  "is_active": boolean
}
```

#### Trigger Scrape
```
POST /api/v1/web/article-sources/{source_id}/scrape
Response: {
  "source_key": "string",
  "categories": [
    {
      "category_key": "string",
      "total_items": number,
      "new_items": number,
      "duplicates": number
    }
  ]
}
```

---

## FAQ

### Q: Berapa sering harus scrape?
**A:** Tergantung kebutuhan:
- Website berita aktif: setiap 1-6 jam
- Website standar: setiap 12-24 jam
- Website jarang update: setiap 2-3 hari

### Q: Apakah ada limit jumlah source?
**A:** Tidak ada limit teknis, tapi:
- Setiap source = 1 HTTP request untuk RSS + N request untuk detail
- Pastikan server punya resource yang cukup
- Rekomendasi: max 50-100 source aktif

### Q: Apakah bisa scrape website yang butuh login?
**A:** Tidak, sistem saat ini hanya mendukung website publik.

### Q: Apakah bisa scrape website yang pakai JavaScript rendering?
**A:** Tidak, sistem hanya bisa scrape website yang content-nya sudah ada di HTML awal. Website yang butuh JavaScript rendering (SPA) tidak didukung.

### Q: Bagaimana jika website berubah struktur HTML?
**A:** Update CSS selector di source configuration. Jika tidak yakin, kosongkan selector untuk gunakan readability.

### Q: Apakah artikel yang sudah di-scrape bisa di-scrape ulang?
**A:** Tidak, sistem menggunakan `original_url` untuk deduplication. Artikel dengan URL yang sama akan di-skip.

---

## Changelog

### 2026-08-04
- Initial documentation
- Added configuration guide
- Added troubleshooting section
- Added case studies
