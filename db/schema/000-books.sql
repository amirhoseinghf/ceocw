create table books
(
    id           int auto_increment
        primary key,
    title        varchar(255) not null,
    image_url    text         null,
    download_url text         null
);

