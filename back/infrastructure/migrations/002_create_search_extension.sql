CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE products
ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;


UPDATE products p
SET search_vector =
    setweight(
        to_tsvector(
            'simple',
            coalesce(p.title, '')
        ),
        'A'
    )
    ||
    setweight(
        to_tsvector(
            'russian',
            coalesce(p.description, '')
        ),
        'B'
    )
    ||
    setweight(
        to_tsvector(
            'russian',
            coalesce(c.name, '')
        ),
        'C'
    )
FROM categories c
WHERE c.category_id = p.category_id;


UPDATE products p
SET search_vector =
    setweight(
        to_tsvector(
            'simple',
            coalesce(p.title, '')
        ),
        'A'
    )
    ||
    setweight(
        to_tsvector(
            'russian',
            coalesce(p.description, '')
        ),
        'B'
    )
WHERE p.category_id IS NULL;


CREATE INDEX IF NOT EXISTS products_search_vector_idx
ON products
USING GIN(search_vector);


CREATE INDEX IF NOT EXISTS products_title_trgm_idx
ON products
USING GIN(title gin_trgm_ops);


CREATE INDEX IF NOT EXISTS products_description_trgm_idx
ON products
USING GIN(description gin_trgm_ops);

CREATE OR REPLACE FUNCTION products_search_update()
RETURNS TRIGGER AS
$$
DECLARE
    category_name TEXT;
BEGIN

    SELECT name
    INTO category_name
    FROM categories
    WHERE category_id = NEW.category_id;

    NEW.search_vector :=
        setweight(
            to_tsvector(
                'simple',
                coalesce(NEW.title, '')
            ),
            'A'
        )
        ||
        setweight(
            to_tsvector(
                'russian',
                coalesce(NEW.description, '')
            ),
            'B'
        )
        ||
        setweight(
            to_tsvector(
                'russian',
                coalesce(category_name, '')
            ),
            'C'
        );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS products_search_trigger
ON products;


CREATE TRIGGER products_search_trigger
BEFORE INSERT OR UPDATE
ON products
FOR EACH ROW
EXECUTE FUNCTION products_search_update();

ALTER TABLE categories
ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;


UPDATE categories
SET search_vector =
    setweight(
        to_tsvector(
            'russian',
            coalesce(name, '')
        ),
        'A'
    );


CREATE INDEX IF NOT EXISTS categories_search_vector_idx
ON categories
USING GIN(search_vector);

CREATE OR REPLACE FUNCTION categories_search_update()
RETURNS TRIGGER AS
$$
BEGIN

    NEW.search_vector :=
        setweight(
            to_tsvector(
                'russian',
                coalesce(NEW.name, '')
            ),
            'A'
        );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS categories_search_trigger
ON categories;


CREATE TRIGGER categories_search_trigger
BEFORE INSERT OR UPDATE
ON categories
FOR EACH ROW
EXECUTE FUNCTION categories_search_update();