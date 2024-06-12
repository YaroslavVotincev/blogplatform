create table comments
(
    id        uuid primary key not null,
    parent_id uuid             not null,
    author_id uuid             not null,
    content   text             not null,
    created   timestamp        not null default current_timestamp,
    updated   timestamp        not null default current_timestamp
);