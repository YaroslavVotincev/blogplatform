create table goals
(
    id          uuid      not null primary key,
    blog_id     uuid      not null,
    type        text      not null,
    description text      not null,
    target      int       not null,
    current     int       not null,

    created     timestamp not null,
    updated     timestamp not null
);