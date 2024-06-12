alter table users add column banned_until timestamp default null;
alter table users add column banned_reason text default null;