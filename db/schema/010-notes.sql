create table notes
(
    id         int auto_increment
        primary key,
    course_id  int                  not null,
    title      varchar(255)         not null,
    file_name  varchar(255)         not null,
    is_updated tinyint(1) default 0 not null,
    constraint notes_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_notes_course_id
    on notes (course_id);

