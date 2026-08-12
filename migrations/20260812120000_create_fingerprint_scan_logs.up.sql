CREATE TABLE IF NOT EXISTS fingerprint_scan_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sn          VARCHAR(50) NOT NULL,
    scan_date   TIMESTAMPTZ NOT NULL,
    pin         VARCHAR(50) NOT NULL,
    verifymode  INT,
    inoutmode   INT,
    deviceip    VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fingerprint_scan_logs_pin_scandate ON fingerprint_scan_logs(pin, scan_date);
CREATE INDEX idx_fingerprint_scan_logs_scandate ON fingerprint_scan_logs(scan_date);
