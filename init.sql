-- Drop all created tables

DROP TABLE todo_items;
DROP TABLE project_items;

-- Create all tables
CREATE TABLE project_items
(
    id          serial not null unique,
    summary     varchar(255) not null,
    description varchar(255),
    is_done     boolean not null default false
);

CREATE TABLE todos_projects (
  todo_id integer,
  project_id integer,
  PRIMARY KEY (todo_id, project_id)
  CONSTRAINT fk_todo FOREIGN KEY(todo_id) REFERENCES todo_items(id)
  CONSTRAINT fk_project FOREIGN KEY(project_id) REFERENCES project_items(id)
);

INSERT INTO project_items (summary, description, is_done)
VALUES
    ('Project 1', 'Project 1 Hello fkn world', false),
    ('Project 2', 'Project 2 Hello fkn world', false),

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
