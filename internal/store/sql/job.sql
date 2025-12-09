CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,          -- pending, running, success, failed
    tf_module TEXT NOT NULL,       -- nama module/template yang digunakan
    tf_vars JSONB,                 -- variabel terraform (bentuk JSON)
    logs TEXT,                     -- log output (opsional: nanti bisa split ke log table)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
