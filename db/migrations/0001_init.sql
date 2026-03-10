-- +goose Up
CREATE TABLE users (
	id SERIAL 	PRIMARY KEY,
	username 		VARCHAR(100) 	NOT NULL UNIQUE,
	password 		VARCHAR(255) 	NOT NULL,
	email 			VARCHAR(100) 	UNIQUE,
	role 				VARCHAR(100) 	NOT NULL CHECK (role IN ('student', 'teacher', 'admin')),
	created_at 	TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 	TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE teachers (
	user_id 				INT 					PRIMARY KEY,
	name 						VARCHAR(255) 	NOT NULL,
	subject_id 			INT 					NOT NULL,
	date_of_birth 	DATE 					NOT NULL,
	gender 					VARCHAR(10) 	NOT NULL CHECK (gender IN ('male', 'female', 'other')),
	hire_date 			DATE 					NOT NULL,
	working_status 	VARCHAR(20) 	NOT NULL DEFAULT 'active' CHECK (working_status IN ('active', 'inactive', 'on_leave')),
	created_at 			TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 			TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE students (
	user_id 			INT 					PRIMARY KEY,
	name 					VARCHAR(100) 	NOT NULL,
	date_of_birth DATE 					NOT NULL,
	gender 				VARCHAR(10) 	NOT NULL CHECK (gender IN ('male', 'female', 'other')),
	graduated 		BOOLEAN  			NOT NULL DEFAULT FALSE,
	grade 				INT 					NOT NULL,
	created_at 		TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 		TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE subjects (
	id 					SERIAL 				PRIMARY KEY,
	name 				VARCHAR(100) 	NOT NULL UNIQUE,
	created_at 	TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at 	TIMESTAMPTZ 	NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE teachers ADD CONSTRAINT fk_teachers_user_id FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE teachers ADD CONSTRAINT fk_teachers_subject_id FOREIGN KEY (subject_id) REFERENCES subjects(id);
ALTER TABLE students ADD CONSTRAINT fk_students_user_id FOREIGN KEY (user_id) REFERENCES users(id);

-- +goose Down
ALTER TABLE students DROP CONSTRAINT fk_students_user_id;
ALTER TABLE teachers DROP CONSTRAINT fk_teachers_subject_id;
ALTER TABLE teachers DROP CONSTRAINT fk_teachers_user_id;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS teachers;
DROP TABLE IF EXISTS users;