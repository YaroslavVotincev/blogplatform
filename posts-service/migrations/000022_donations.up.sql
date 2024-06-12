create table donations
(
    id                uuid primary key,
    user_id           uuid             not null,
    blog_id           uuid             not null,
    value             double precision not null,
    currency          text             not null,
    user_comment      text             not null,
    status            text             not null,
    payment_confirmed boolean          not null,
    created           timestamp        not null default current_timestamp,
    updated           timestamp        not null default current_timestamp
);