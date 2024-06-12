create table categories
(
    code    text primary key,
    name    text      not null,
    created timestamp not null default current_timestamp,
    updated timestamp not null default current_timestamp
);

insert into categories
values ('auto', 'Авто, мото'),
       ('business', 'Бизнес'),
       ('gaming', 'Видеоигры'),
       ('travel', 'Города, страны, туризм и путешествия'),
       ('home', 'Дом, ремонт и строительство'),
       ('pets', 'Животные'),
       ('animals', 'Искусство'),
       ('tech', 'Компьютер и интернет, сайты и программирование'),
       ('health', 'Красота и здоровье'),
       ('cooking', 'Кулинария и рецепты'),
       ('medicine', 'Медицина'),
       ('military', 'Милитари'),
       ('music', 'Музыка'),
       ('property', 'Недвижимость'),
       ('education', 'Образование'),
       ('ads', 'Объявления'),
       ('unions', 'Объединения, группы людей'),
       ('relationships', 'Отношения, семья и знакомства'),
       ('work', 'Работа'),
       ('entertainment', 'Развлечения'),
       ('sport', 'Спорт'),
       ('insurance', 'Страхование'),
       ('hobbies', 'Увлечения и хобби'),
       ('finance', 'Финансы');

create table blog_categories
(
    blog_id  uuid not null,
    category text not null,

    primary key (blog_id, category)
);