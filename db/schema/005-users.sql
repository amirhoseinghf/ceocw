create table users
(
    id            int auto_increment
        primary key,
    first_name    varchar(100)                                                        not null,
    last_name     varchar(100)                                                        not null,
    email         varchar(255)                                                        not null,
    password_hash char(60)                                                            not null,
    user_type     enum ('normal', 'ta', 'head_ta', 'admin') default 'normal'          not null,
    image_path    varchar(255)                                                        null,
    is_active     tinyint(1)                                default 1                 not null,
    created_at    timestamp                                 default CURRENT_TIMESTAMP null,
    updated_at    timestamp                                 default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
    constraint email
        unique (email)
);

