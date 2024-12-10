-- HVACs

DROP TABLE hvacs_telemetry;
DROP TABLE hvacs;

CREATE TABLE IF NOT EXISTS hvacs (
    hvac_id INTEGER  PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    serial_ BIGINT  NOT NULL,
    model   VARCHAR NOT NULL
);

ALTER TABLE hvacs ADD CONSTRAINT hvacs_uk UNIQUE (serial_);

CREATE TABLE IF NOT EXISTS hvacs_telemetry (
    hvac_telemetry_id   BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    hvac_id             INTEGER NOT NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    current_temperature NUMERIC NOT NULL,
    target_temperature  NUMERIC NOT NULL
);

CREATE INDEX hvacs_telemetry_hvacs_fk ON hvacs_telemetry(hvac_id, created_at);
ALTER TABLE hvacs_telemetry ADD CONSTRAINT fk_hvacs_telemetry_hvacs FOREIGN KEY(hvac_id) REFERENCES hvacs(hvac_id);

INSERT INTO hvacs(model, serial_) VALUES ('WarmHome 1.2', 123);
INSERT INTO hvacs(model, serial_) VALUES ('WarmHome 1.3', 456);
COMMIT;
