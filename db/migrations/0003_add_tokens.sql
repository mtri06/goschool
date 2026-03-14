-- +goose Up
CREATE TABLE tokens (
	id 							SERIAL 			PRIMARY KEY,
	body 						TEXT 				NOT NULL,
	user_id 				INT 				NOT NULL,
	type 						VARCHAR(50) NOT NULL CHECK (type IN ('refresh_token', 'password_update_token', 'email_verification_token')),
	expires_at 			TIMESTAMPTZ NOT NULL,
	is_revoked 			BOOLEAN 		NOT NULL DEFAULT FALSE,
	is_used 				BOOLEAN 		NOT NULL DEFAULT FALSE,
	is_blacklisted 	BOOLEAN 		NOT NULL DEFAULT FALSE,
	created_at 			TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 			TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tokens ADD CONSTRAINT fk_tokens_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX idx_tokens_body ON tokens (body);

-- +goose Down
DROP INDEX IF EXISTS idx_tokens_body;
ALTER TABLE tokens DROP CONSTRAINT fk_tokens_user_id;
DROP TABLE IF EXISTS tokens;
