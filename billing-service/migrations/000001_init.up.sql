create table robokassa_invoices
(
    id           serial primary key not null,
    out_sum      numeric(10, 2)     not null,
    item_id      uuid               not null,
    item_type    text               not null,
    user_id      uuid               not null,
    expires_at   timestamp          not null,
    status       text               not null,
    payment_link text               not null,
    created      timestamp          not null default current_timestamp,
    updated      timestamp          not null default current_timestamp
);