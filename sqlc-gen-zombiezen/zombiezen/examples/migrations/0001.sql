CREATE TABLE nullableTestTable (id integer PRIMARY KEY, myBool boolean) strict;

CREATE TABLE NAMES (
    id integer PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE authors (
    id integer PRIMARY KEY,
    first_name_id integer NOT NULL,
    last_name_id integer NOT NULL,
    FOREIGN KEY (first_name_id) REFERENCES NAMES(id),
    FOREIGN KEY (last_name_id) REFERENCES NAMES(id)
);