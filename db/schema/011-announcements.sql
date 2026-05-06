create table announcements
(
    id         int auto_increment
        primary key,
    course_id  int                                not null,
    title      varchar(255)                       not null,
    content    text                               not null,
    created_at datetime default CURRENT_TIMESTAMP not null,
    updated_at datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,
    constraint announcements_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_announcements_course_id
    on announcements (course_id);

create index idx_announcements_created_at
    on announcements (created_at);

