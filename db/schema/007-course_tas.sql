create table course_tas
(
    course_id int not null,
    ta_id     int not null,
    primary key (course_id, ta_id),
    constraint course_tas_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade,
    constraint course_tas_ibfk_2
        foreign key (ta_id) references teaching_assistants (id)
            on delete cascade
);

create index ta_id
    on course_tas (ta_id);

