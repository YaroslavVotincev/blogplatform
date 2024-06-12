alter table posts
    add column access_mode text default '1';
alter table posts
    drop column temp_type;
alter table posts
    add column price numeric(10, 2) default null;
alter table posts
    add column required_subscriptions text[] default null;