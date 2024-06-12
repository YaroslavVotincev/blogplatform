create table users
(
    id uuid not null primary key,
    login varchar(30) not null,
    email varchar(254) not null,
    hashed_password text not null,
    role text not null,
    deleted bool not null default false,
    enabled bool not null default false,
    email_confirmed_at timestamp default null,
    erase_at timestamp default null,
    created timestamp not null default current_timestamp,
    updated timestamp not null default current_timestamp
);

create unique index users_unique_login_idx on users(login);
create unique index users_unique_email_idx on users(email);

create table profiles
(
    id uuid primary key,
    first_name varchar(30) not null,
    last_name varchar(30) not null,
    middle_name varchar(30) not null,

    foreign key (id) references users(id) on delete cascade
);