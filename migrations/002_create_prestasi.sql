CREATE TABLE IF NOT EXISTS prestasi (
    id                  SERIAL          PRIMARY KEY,
    student_id          SERIAL          NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    name_prestation     VARCHAR(20)     NOT NULL,
    juara               VARCHAR(20)     NOT NULL
);

CREATE INDEX IF NOT EXISTS prestation_student_id_idx ON prestasi (student_id);