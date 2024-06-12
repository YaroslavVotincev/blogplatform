alter table posts
    add column if not exists likes_count int default 0,
    add column if not exists comments_count int default 0