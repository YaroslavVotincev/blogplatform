create table notifications
(
    id         uuid primary key,
    event_code text      not null,
    user_id    uuid      not null,
    seen       boolean   not null default false,
    data       json      not null,
    created    timestamp not null,
    updated    timestamp not null
);