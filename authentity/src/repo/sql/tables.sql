CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY, -- pretty obvious

    server      TEXT, -- Server id that requested
    service     TEXT, -- Service where change was made

    table_name  TEXT, -- Table to who it was written to

    prev_state  TEXT, -- Previous data change, If empty: new insert.
    new_state   TEXT, -- New Data Change

    ip          VARCHAR(45), -- Len considering ipv4/ipv6 and ipv4-mapped-ipv6
    mac         VARCHAR(17), -- Mac addr: 00:1A:2B:3C:4D:5E or 00-1A-2B-3C-4D-5E

    message     TEXT, -- Message / Description
    timestamp   TEXT NOT NULL , -- When did we get the log
    user_id     TEXT NOT NULL   -- Who's Identity session specifically was it?
);

CREATE TABLE IF NOT EXISTS identity (
    identity_id     VARCHAR(16) PRIMARY KEY, -- uid len 16

    username        VARCHAR(255) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,

    phone_number    VARCHAR(50) UNIQUE,

    ip              VARCHAR(45),
    mac             VARCHAR(17),

    status          BOOLEAN DEFAULT TRUE, -- active, banned, deleting

    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME

);


CREATE TABLE IF NOT EXISTS account (
    account_id              VARCHAR(16) PRIMARY KEY, -- uid len 16
    identity_id             VARCHAR(16), -- uid len 16

    password_hash           VARCHAR(255) NOT NULL,
    password_salt           VARCHAR(255),

    status                  VARCHAR(50) DEFAULT 'active', -- e.g., active, suspended, deleted

    last_login              DATETIME,
    failed_login_attempts   INT DEFAULT 0,
    lockout_end             DATETIME,

    created_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME,

    FOREIGN KEY (identity_id) REFERENCES identity(identity_id)
);


CREATE TABLE IF NOT EXISTS profile (
    profile_id          VARCHAR(16) PRIMARY KEY, -- uid len 16
    identity_id         VARCHAR(16), -- uid len 16

    first_name          VARCHAR(100),
    last_name           VARCHAR(100),

    date_of_birth       DATE,

    address             TEXT,
    sex                 VARCHAR(6), -- ...
    alias               VARCHAR(50),

    preferences         TEXT, -- JSON or serialized text
    profile_picture_url VARCHAR(255),
    bio                 TEXT,

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME,


    FOREIGN  KEY (identity_id) REFERENCES identity(identity_id)
);


CREATE TABLE IF NOT EXISTS resource (
    resource_id     VARCHAR(16) PRIMARY KEY, -- uid len 16
    identity_id     VARCHAR(16), -- uid len 16

    resource_name   VARCHAR(255),
    resource_type   VARCHAR(50), -- e.g., file, service, document

    permissions     VARCHAR(255), -- e.g., read, write, delete

    resource_url    VARCHAR(255),

    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,

    FOREIGN KEY (identity_id) REFERENCES identity(identity_id)
);

