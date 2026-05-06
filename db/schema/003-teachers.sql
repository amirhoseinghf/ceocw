create table teachers
(
    id                 bigint unsigned auto_increment
        primary key,
    image_url          text null,
    first_name         text not null,
    last_name          text not null,
    first_name_english text null,
    last_name_english  text null,
    page_url           text null,
    constraint id
        unique (id)
);

