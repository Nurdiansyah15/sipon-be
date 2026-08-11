-- Tambah program_id ke pendaftar: referensi ke programs.id di akademik.
-- program (string) tetap dipertahankan sebagai cache display, sedangkan
-- program_id menjadi sumber kebenaran program yang dipilih pendaftar PSB.
-- Lihat docs/plan/santri-program-mapping.md.

ALTER TABLE pendaftar
    ADD COLUMN IF NOT EXISTS program_id UUID REFERENCES programs(id);
