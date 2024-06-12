CREATE TABLE IF NOT EXISTS settings_services
(
    service VARCHAR(64) PRIMARY KEY,
    created TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS settings_items
(
    key     VARCHAR(64)                                                                            NOT NULl,
    value   TEXT                                                                                   NOT NULL,
    service VARCHAR(64) REFERENCES settings_services (service) ON UPDATE CASCADE ON DELETE CASCADE NOT NULL,
    created TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_settings_items PRIMARY KEY (service, key)
);
