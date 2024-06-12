create table subscriptions
(
    id                uuid primary key,
    blog_id           uuid           not null,
    title             text           not null,
    short_description text           not null,
    cover             text                    default null,
    is_free           boolean        not null,
    price_rub         numeric(10, 2) not null,
    is_active         boolean        not null,
    created           timestamp      not null default current_timestamp,
    updated           timestamp      not null default current_timestamp
);