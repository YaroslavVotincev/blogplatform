create table blogs
(
    id                uuid primary key,
    author_id         uuid      not null,
    url               text      not null,
    title             text      not null,
    description       text      not null,
    short_description text      not null,
    status            text      not null,
    created           timestamp not null default current_timestamp,
    updated           timestamp not null default current_timestamp
);