-- +goose Up
-- Insert high school subjects
INSERT INTO subjects (name) VALUES
('Mathematics'),
('English'),
('Physics'),
('Chemistry'),
('Biology'),
('History'),
('Geography'),
('Art'),
('Music');

-- +goose Down
DELETE FROM subjects WHERE name IN (
    'Mathematics', 'English', 'Physics', 'Chemistry', 
    'Biology', 'History', 'Geography', 'Art', 'Music'
);