CREATE TABLE IF NOT EXISTS reviews (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    movie_id INT NOT NULL REFERENCES movies (id) ON DELETE CASCADE,
    user_id INT NOT NULL,
    comment TEXT,
    rating INT NOT NULL DEFAULT 5 CHECK (rating > 0 AND rating < 6),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE reviews ADD CONSTRAINT reviews_movie_id_user_id_uniqueness UNIQUE (movie_id, user_id);

-- Функция для обновления поля updated_at
CREATE OR REPLACE FUNCTION update_timestamp_t() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now(); 
    RETURN NEW; 
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER reviews_update_timestamp
AFTER UPDATE ON reviews
FOR EACH ROW EXECUTE FUNCTION update_timestamp_t();