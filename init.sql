-- Drop all created tables

DROP TABLE todo_items;

-- Create all tables

CREATE TABLE todo_items
(
    id          serial       not null unique,
    summary     varchar(255) not null,
    description varchar(255),
    is_done     boolean      not null default false
);

INSERT INTO todo_items (summary, description, is_done)
VALUES
    ('Todo 1 Hello world', 'Todo 1 Hello world from planet earth', false),
    ('Todo 2 Hello world', 'Todo 2 Hello world from planet earth', false),
    ('Todo 3 Hello world', 'Todo 3 Hello world from planet earth', false),
    ('Todo 4 Hello world', 'Todo 4 Hello world from planet earth', false);
