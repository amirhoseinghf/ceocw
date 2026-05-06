create table slides
(
    id        int auto_increment
        primary key,
    course_id int          not null,
    title     varchar(255) not null,
    file_name varchar(255) not null,
    constraint slides_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_slides_course_id
    on slides (course_id);

