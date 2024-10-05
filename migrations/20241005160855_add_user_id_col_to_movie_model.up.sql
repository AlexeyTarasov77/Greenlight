ALTER TABLE movies ADD COLUMN user_id INT NOT NULL CHECK (user_id > 0);

ALTER TABLE reviews ADD CONSTRAINT user_id_constraint CHECK (user_id > 0);