create table posts
(
    id                uuid primary key,
    blog_id           uuid      not null,
    title             text      not null,
    short_description text      not null,
    content           text      not null,
    created           timestamp not null default current_timestamp,
    updated           timestamp not null default current_timestamp
);