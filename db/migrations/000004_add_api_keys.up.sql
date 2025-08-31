CREATE TYPE http_method_enum AS ENUM ('GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD');

CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    client_id TEXT NOT NULL,
    api_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE endpoints (
    id SERIAL PRIMARY KEY,
    http_method http_method_enum NOT NULL,
    http_endpoint TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT uq_endpoints UNIQUE (http_method, http_endpoint)
);

CREATE TABLE api_key_endpoints (
    api_key_id INT REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint_id INT REFERENCES endpoints(id) ON DELETE CASCADE,
    is_active BOOL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (api_key_id, endpoint_id)
);

-- Indexes for faster lookup
CREATE INDEX idx_api_key_endpoints_key ON api_key_endpoints(api_key_id);
CREATE INDEX idx_api_key_endpoints_endpoint ON api_key_endpoints(endpoint_id);
CREATE INDEX idx_api_key_endpoints_lookup ON api_key_endpoints(api_key_id, endpoint_id);

-- Insert existing endpoints
INSERT INTO endpoints (http_method, http_endpoint)
VALUES
        ('POST', '/admin/apikeys'),
        ('GET', '/admin/endpoints'),
        ('GET', '/products'),
        ('GET', '/products/{id}'),
        ('POST', '/orders');

