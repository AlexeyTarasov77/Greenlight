CREATE OR REPLACE PROCEDURE create_movie_constraints() AS $$
    BEGIN
        ALTER TABLE movies ADD CONSTRAINT movies_runtime_check CHECK (runtime >= 0);
        ALTER TABLE movies ADD CONSTRAINT movies_year_check CHECK (year BETWEEN 1888 AND date_part('year', now()));
        ALTER TABLE movies ADD CONSTRAINT genres_length_check CHECK (array_length(genres, 1) BETWEEN 1 AND 5);
        ALTER TABLE movies ADD CONSTRAINT movies_unique_check UNIQUE (title, version, year);
        EXCEPTION WHEN others THEN RAISE NOTICE 'That constratint already exists';
    END;
$$ LANGUAGE plpgsql;

CALL create_movie_constraints();