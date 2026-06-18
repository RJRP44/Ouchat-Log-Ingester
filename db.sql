-- Create tables

CREATE TABLE cats
(
    cat VARCHAR(5) PRIMARY KEY
);

CREATE TABLE sessions
(
    cat       VARCHAR(5) REFERENCES cats,
    timestamp TIMESTAMP,
    PRIMARY KEY (cat, timestamp)
);

CREATE TABLE raw_data
(
    cat VARCHAR(5),
    timestamp TIMESTAMP,
    mcu_ms INTEGER CHECK ( mcu_ms > 0 ),
    data TEXT,
    PRIMARY KEY (cat, timestamp, mcu_ms),
    FOREIGN KEY (cat, timestamp) REFERENCES sessions(cat, timestamp)
);

CREATE TABLE calibration_data
(
    cat VARCHAR(5),
    timestamp TIMESTAMP,
    data TEXT,
    PRIMARY KEY (cat, timestamp),
    FOREIGN KEY (cat, timestamp) REFERENCES sessions(cat, timestamp)
);

CREATE TABLE logs
(
    id SERIAL PRIMARY KEY,
    cat VARCHAR(5),
    timestamp TIMESTAMP,
    mcu_ms INTEGER CHECK ( mcu_ms > 0 ),
    level INTEGER,
    log TEXT,
    FOREIGN KEY (cat, timestamp) REFERENCES sessions(cat, timestamp)
);