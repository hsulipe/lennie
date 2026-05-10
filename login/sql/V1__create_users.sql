CREATE TABLE users (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  cpf VARCHAR(14) NOT NULL UNIQUE,
  phone VARCHAR(20) NOT NULL UNIQUE,
  birthdate DATE NOT NULL,
  activated_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE,
  deleted BOOLEAN DEFAULT FALSE,
  phone_verified BOOLEAN DEFAULT FALSE,
  email_verified BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_cpf ON users (cpf);

--
CREATE TYPE identity_provider AS ENUM(
  'password',
  'google',
  'apple',
  'github',
  'facebook'
);

CREATE TABLE user_identities (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  provider identity_provider NOT NULL,
  provider_identifier VARCHAR(255) NOT NULL,
  credential_hash TEXT NOT NULL,
  scopes TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted BOOLEAN DEFAULT FALSE,
  UNIQUE (provider, provider_identifier),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_identities_user ON user_identities (user_id);