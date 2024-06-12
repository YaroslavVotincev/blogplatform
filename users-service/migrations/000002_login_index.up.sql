drop index users_unique_login_idx;

create unique index users_unique_login_idx on users (lower(login));