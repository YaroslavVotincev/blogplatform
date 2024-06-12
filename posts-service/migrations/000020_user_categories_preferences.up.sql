create table user_categories_preferences
(
    user_id  uuid not null,
    category text not null,

    primary key (user_id, category)
);