create table post_likes
(
    id       uuid primary key,
    post_id  uuid      not null,
    user_id  uuid      not null,
    positive boolean   not null,
    created  timestamp not null default current_timestamp
);