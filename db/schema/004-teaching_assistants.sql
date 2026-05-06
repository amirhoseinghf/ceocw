create table teaching_assistants
(
    id         int auto_increment
        primary key,
    first_name varchar(100)                        not null,
    last_name  varchar(100)                        not null,
    image_url  text                                null,
    linkedin   varchar(255)                        null,
    telegram   varchar(255)                        null,
    instagram  varchar(255)                        null,
    website    varchar(255)                        null,
    github     varchar(255)                        null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    updated_at timestamp default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP
);

