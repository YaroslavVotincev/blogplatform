drop table blogs;
truncate posts;
truncate tags;
truncate posts_tags;

create table blogs
(
    id                uuid primary key,
    author_id         uuid      not null,
    type              text      not null,
    url               text      not null,
    title             text      not null,
    content           text               default null,
    short_description text      not null,
    status            text      not null,
    accept_donations  boolean   not null,
    avatar            text               default null,
    cover             text               default null,
    created           timestamp not null default current_timestamp,
    updated           timestamp not null default current_timestamp
);

create unique index blogs_url_uidx on blogs (url);