alter table posts drop column if exists content;
alter table blogs drop column if exists content;

create table contents (
    id uuid primary key ,
    data_json text not null,
    data_html text not null,
    created timestamp not null default current_timestamp,
    updated timestamp not null default current_timestamp
)