# Schema: Persuratan

Dokumen ini merangkum skema DB persuratan (sub-modul kesantrian).
Migration: `20260810130000_create_persuratan_tables`.

---

## `master_tipe_surat` — master jenis surat

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| nama | VARCHAR(200) NOT NULL | Nama tipe surat |
| kode | VARCHAR(20) UNIQUE NOT NULL | Kode singkat (SKA, SKM, dst) |
| created_by | UUID | Admin yang membuat |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |

Index:
- `idx_master_tipe_surat_kode` UNIQUE pada `kode` (implicit dari `UNIQUE` constraint)

---

## `surat` — catatan surat keluar

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| nomor | VARCHAR(100) UNIQUE NOT NULL | Auto-generated: `{SEQ}/{KODE}/{ORG}/{ROMAWI}/{YYYY}` |
| seq | INT NOT NULL | Sequence per bulan |
| tipe_surat_id | UUID NOT NULL FK → master_tipe_surat | |
| keterangan | TEXT | Deskripsi surat |
| tanggal | DATE NOT NULL | Tanggal surat dikeluarkan |
| created_by | UUID NOT NULL | User ID admin pembuat |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |
| updated_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |

Index:
- `idx_surat_nomor` UNIQUE pada `nomor` (implicit)
- `idx_surat_tipe` pada `tipe_surat_id`
- `idx_surat_tanggal` pada `(EXTRACT(MONTH FROM tanggal), EXTRACT(YEAR FROM tanggal))` — untuk seq lookup

**Tidak ada kolom `deleted_at`** — surat yang dihapus di-hard-delete (nomor tidak bisa dipakai ulang, cukup hilangkan baris).

---

## `surat_dokumen_aset` — junction surat ↔ dokumen_aset

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | `DEFAULT gen_random_uuid()` |
| surat_id | UUID NOT NULL FK → surat ON DELETE CASCADE | |
| dokumen_aset_id | UUID NOT NULL | No FK (cross-module ke dokumen_aset) |
| created_at | TIMESTAMPTZ NOT NULL | `DEFAULT NOW()` |

Constraints:
- `UNIQUE (surat_id, dokumen_aset_id)` — tidak boleh link duplikat

Index:
- `idx_surat_dokumen_surat` pada `surat_id`
- `idx_surat_dokumen_aset` pada `dokumen_aset_id`

---

## DDL

```sql
CREATE TABLE IF NOT EXISTS master_tipe_surat (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama        VARCHAR(200) NOT NULL,
    kode        VARCHAR(20) UNIQUE NOT NULL,
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS surat (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nomor           VARCHAR(100) UNIQUE NOT NULL,
    seq             INT NOT NULL,
    tipe_surat_id   UUID NOT NULL REFERENCES master_tipe_surat(id),
    keterangan      TEXT,
    tanggal         DATE NOT NULL,
    created_by      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_surat_tipe ON surat(tipe_surat_id);
CREATE INDEX idx_surat_tanggal ON surat(EXTRACT(MONTH FROM tanggal), EXTRACT(YEAR FROM tanggal));

CREATE TABLE IF NOT EXISTS surat_dokumen_aset (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    surat_id         UUID NOT NULL REFERENCES surat(id) ON DELETE CASCADE,
    dokumen_aset_id  UUID NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (surat_id, dokumen_aset_id)
);
CREATE INDEX idx_surat_dokumen_surat ON surat_dokumen_aset(surat_id);
CREATE INDEX idx_surat_dokumen_aset ON surat_dokumen_aset(dokumen_aset_id);
```
