create table blog_incomes
(
    id                  uuid primary key,
    blog_id             uuid             not null,
    user_id             uuid             not null,
    value               double precision not null,
    currency            text             not null,
    item_id             uuid             not null,
    item_type                text             not null,
    sent_to_user_wallet boolean          not null,
    created             timestamp        not null default current_timestamp
);