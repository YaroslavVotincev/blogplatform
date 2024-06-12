create table user_follows
(
    id      uuid primary key,
    user_id uuid not null,
    blog_id uuid not null,
    created timestamp default current_timestamp,
    updated timestamp default current_timestamp
);