create table courses
(
    id                         int auto_increment
        primary key,
    title                      varchar(255)    not null,
    short_name                 varchar(255)    not null,
    image_url                  text            null,
    telegram_link              varchar(512)    null,
    bale_link                  varchar(512)    null,
    quera_link                 varchar(512)    null,
    active_section             varchar(255)    null,
    teacher_id                 bigint unsigned not null,
    semester_id                bigint unsigned not null,
    description                text            null,
    class_schedule_day_of_week varchar(50)     null,
    class_schedule_start_time  varchar(20)     null,
    class_schedule_end_time    varchar(20)     null,
    class_schedule_location    varchar(255)    null,
    constraint courses_ibfk_1
        foreign key (teacher_id) references teachers (id),
    constraint courses_ibfk_2
        foreign key (semester_id) references semesters (id)
);

create index idx_courses_semester_id
    on courses (semester_id);

create index idx_courses_short_name
    on courses (short_name);

create index idx_courses_teacher_id
    on courses (teacher_id);

