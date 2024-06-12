alter table wallets
    add column if not exists balance_rub numeric(10, 2) not null default 0.00;