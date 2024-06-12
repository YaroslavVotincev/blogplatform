create table content_files
(
    id         uuid primary key,
    content_id uuid      not null,
    type       text      not null,
    size       int    not null,
    created    timestamp not null default current_timestamp,
    updated    timestamp not null default current_timestamp
)