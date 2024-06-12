create table history_logs
(
    id      uuid      not null,
    level   text      not null,
    service text      not null,
    user_id uuid,
    data    text      not null,
    created timestamp not null default current_timestamp,

    primary key (id)
);

