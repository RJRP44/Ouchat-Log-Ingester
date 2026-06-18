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

CREATE TABLE logs
(
    cat VARCHAR(5),
    timestamp TIMESTAMP,
    mcu_ms INTEGER CHECK ( mcu_ms > 0 ),
    level INTEGER,
    log TEXT,
    PRIMARY KEY (cat, timestamp, mcu_ms),
    FOREIGN KEY (cat, timestamp) REFERENCES sessions(cat, timestamp)
);

