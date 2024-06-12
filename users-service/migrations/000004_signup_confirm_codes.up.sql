create table signup_confirm_codes
(
    code       text primary key,
    user_id    uuid      not null,
    expires_at timestamp not null,
    created_at timestamp not null
);