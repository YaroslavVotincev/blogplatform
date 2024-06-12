create table tags
(
    id   uuid primary key,
    name text default null,
    slug text not null
);

create unique index tags_slug_idx on tags (slug);

create table posts_tags
(
    post_id uuid not null,
    tag_id  text not null,
    primary key (post_id, tag_id)
);

alter table posts
    add column tags_string text not null default '';