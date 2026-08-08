-- Shared, atomic number-sequence counter for all document numbers issued by
-- the keuangan module (invoice, payment, ...). One compact table keyed by
-- (doc_type, year) instead of one counter table per document type — new doc
-- types just use a new doc_type value, no new table/migration needed.
CREATE TABLE IF NOT EXISTS finance_number_sequences (
    doc_type VARCHAR(20) NOT NULL,
    year     INTEGER     NOT NULL,
    seq      INTEGER     NOT NULL DEFAULT 0,
    PRIMARY KEY (doc_type, year)
);
