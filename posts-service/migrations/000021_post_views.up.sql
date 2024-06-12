create table post_views
(
    id          uuid primary key,
    post_id     uuid      not null,
    user_id     uuid               default null,
    fingerprint text               default null,
    created     timestamp not null default current_timestamp
);