-- Lightings

DROP TABLE lights_telemetry;
DROP TABLE lights;

CREATE TABLE IF NOT EXISTS lights (
    light_id INTEGER  PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    serial_  BIGINT  NOT NULL,
    model    VARCHAR NOT NULL
);

ALTER TABLE lights ADD CONSTRAINT lights_uk UNIQUE (serial_);

CREATE TABLE IF NOT EXISTS lights_telemetry (
    light_telemetry_id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    light_id           INTEGER NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL,
    current_brightness NUMERIC NOT NULL,
    target_brightness  NUMERIC NOT NULL
);

CREATE INDEX lights_telemetry_hvacs_fk ON lights_telemetry(light_id, created_at);
ALTER TABLE lights_telemetry ADD CONSTRAINT fk_lights_telemetry_lights FOREIGN KEY(light_id) REFERENCES lights(light_id);

INSERT INTO lights(model, serial_) VALUES ('BrightHome 1.2', 123);
INSERT INTO lights(model, serial_) VALUES ('BrightHome 1.3', 456);
COMMIT;
