alter table posts
    add column status text not null default 'draft';
alter table posts
    add column url text not null default substr(md5(random()::text), 1, 25);
ALTER TABLE posts
    ALTER COLUMN url DROP DEFAULT;
alter table posts
    add column cover text default null;
alter table posts
    drop column content;
alter table posts
    add column content text default null;