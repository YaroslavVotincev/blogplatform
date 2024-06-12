alter table users
    drop column if exists wallet;

create table wallets
(
    id        uuid primary key,
    address   text      not null,
    publicKey text      not null,
    secretKey text      not null,
    mnemonic  text[]    not null,
    created   timestamp not null default current_timestamp
)