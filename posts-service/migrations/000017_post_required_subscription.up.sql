alter table posts
    drop column if exists required_subscriptions;
alter table posts
    add column subscription_id text default null;