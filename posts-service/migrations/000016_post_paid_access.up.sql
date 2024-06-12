create table post_paid_access
(
    id      uuid primary key,
    user_id uuid      not null,
    post_id uuid      not null,
    created timestamp not null default current_timestamp
);