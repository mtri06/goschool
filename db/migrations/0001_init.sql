-- +goose Up
CREATE TABLE users (
  id SERIAL     PRIMARY KEY,
  username      VARCHAR(100) NOT NULL UNIQUE,
  password      VARCHAR(255) NOT NULL,
  email         VARCHAR(100),
  role          VARCHAR(100) NOT NULL CHECK (role IN ('student', 'teacher', 'admin')),
  name          VARCHAR(255) NOT NULL,
  date_of_birth DATE         NOT NULL,
  gender        VARCHAR(10)  NOT NULL CHECK (gender IN ('male', 'female', 'other')),
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_teachers (
  user_id        INT         PRIMARY KEY,
  subject_id     INT         NOT NULL,
  hire_date      DATE        NOT NULL,
  working_status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (working_status IN ('active', 'inactive', 'on_leave')),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_students (
  user_id         INT         PRIMARY KEY,
  admission_date  DATE        NOT NULL,
  graduated       BOOLEAN     NOT NULL DEFAULT FALSE,
  graduated_date  DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE subjects (
  id         SERIAL       PRIMARY KEY,
  name       VARCHAR(100) NOT NULL UNIQUE,
  status     VARCHAR(20)  NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE class (
  id          SERIAL        PRIMARY KEY,
  name        VARCHAR(100)  NOT NULL UNIQUE,
  grade       INT,
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE class_teachers (
  id                  SERIAL      PRIMARY KEY,
  teacher_id          INT         NOT NULL,
  class_id            INT         NOT NULL,
  subject_id          INT         NOT NULL,
  year                INT         NOT NULL,
  start_date          DATE        NOT NULL,
  end_date            DATE        NOT NULL,
  is_homeroom_teacher BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE enrollments (
  id         SERIAL      PRIMARY KEY,
  student_id INT         NOT NULL,
  class_id   INT         NOT NULL,
  year       INT         NOT NULL, 
  start_date DATE        NOT NULL,
  end_date   DATE        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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

ALTER TABLE user_teachers ADD CONSTRAINT fk_user_teachers_user_id FOREIGN KEY (user_id)
REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE user_teachers ADD CONSTRAINT fk_user_teachers_subject_id FOREIGN KEY (subject_id)
REFERENCES subjects(id) ON DELETE NO ACTION;
ALTER TABLE user_students ADD CONSTRAINT fk_user_students_user_id FOREIGN KEY (user_id)
REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE class_teachers ADD CONSTRAINT fk_class_teachers_teacher_id FOREIGN KEY (teacher_id)
REFERENCES user_teachers(user_id) ON DELETE NO ACTION;
ALTER TABLE class_teachers ADD CONSTRAINT fk_class_teachers_class_id FOREIGN KEY (class_id)
REFERENCES class(id) ON DELETE NO ACTION;
ALTER TABLE class_teachers ADD CONSTRAINT fk_class_teachers_subject_id FOREIGN KEY (subject_id)
REFERENCES subjects(id) ON DELETE NO ACTION;
ALTER TABLE tokens ADD CONSTRAINT fk_tokens_user_id FOREIGN KEY (user_id) 
REFERENCES users(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE INDEX idx_tokens_body ON tokens (body);
CREATE INDEX idx_class_teachers_class_id_year ON class_teachers (class_id, year);
CREATE INDEX idx_enrollments_class_id_year ON enrollments (class_id, year);
CREATE INDEX idx_class_teachers_teacher_id ON class_teachers (teacher_id);
CREATE UNIQUE INDEX idx_enrollments_student_class_year ON enrollments (student_id, class_id, year);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_tokens_body;
DROP INDEX IF EXISTS idx_class_teachers_class_id_year;
DROP INDEX IF EXISTS idx_enrollments_class_id_year;
DROP INDEX IF EXISTS idx_class_teachers_teacher_id;
DROP INDEX IF EXISTS idx_enrollments_student_class_year;

ALTER TABLE class_teachers DROP CONSTRAINT fk_class_teachers_subject_id;
ALTER TABLE class_teachers DROP CONSTRAINT fk_class_teachers_class_id;
ALTER TABLE class_teachers DROP CONSTRAINT fk_class_teachers_teacher_id;
ALTER TABLE user_students DROP CONSTRAINT fk_user_students_user_id;
ALTER TABLE user_teachers DROP CONSTRAINT fk_user_teachers_subject_id;
ALTER TABLE user_teachers DROP CONSTRAINT fk_user_teachers_user_id;
ALTER TABLE tokens DROP CONSTRAINT fk_tokens_user_id;

DROP TABLE IF EXISTS class_teachers;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS user_students;
DROP TABLE IF EXISTS user_teachers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tokens;
