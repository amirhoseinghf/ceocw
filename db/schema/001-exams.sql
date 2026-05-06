create table exams
(
    id            bigint unsigned auto_increment
        primary key,
    course_id     int                  not null,
    semester_id   int                  null,
    exam_type     varchar(255)         not null,
    file_name     text                 not null,
    this_semester tinyint(1) default 0 not null,
    constraint id
        unique (id)
);

create index idx_exams_course_id
    on exams (course_id);

create index idx_exams_exam_type
    on exams (exam_type);

create index idx_exams_semester_id
    on exams (semester_id);

create index idx_exams_this_semester
    on exams (this_semester);

