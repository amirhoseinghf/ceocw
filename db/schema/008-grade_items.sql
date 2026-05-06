create table grade_items
(
    id         int auto_increment
        primary key,
    course_id  int          not null,
    name       varchar(255) not null,
    percentage varchar(50)  not null,
    constraint grade_items_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_grade_items_course_id
    on grade_items (course_id);

