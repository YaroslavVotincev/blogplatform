create table user_subscriptions
(
    id              uuid primary key,
    user_id         uuid      not null,
    subscription_id uuid      not null,
    blog_id         uuid      not null,
    status          text      not null,
    is_active          boolean   not null,
    expires_at      timestamp default null,
    created         timestamp not null,
    updated         timestamp not null
);