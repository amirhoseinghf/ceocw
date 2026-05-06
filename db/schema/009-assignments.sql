create table assignments
(
    id            int auto_increment
        primary key,
    course_id     int                  not null,
    title         varchar(255)         not null,
    file_name     varchar(255)         null,
    solution_name varchar(255)         null,
    description   text                 null,
    release_date  datetime             null,
    deadline_date datetime             null,
    is_extended   tinyint(1) default 0 not null,
    is_project    tinyint(1) default 0 not null,
    constraint assignments_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_assignments_course_id
    on assignments (course_id);

create index idx_assignments_deadline
    on assignments (deadline_date);

