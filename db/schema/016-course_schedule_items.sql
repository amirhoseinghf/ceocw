create table if not exists course_schedule_items
(
    id int not null auto_increment,
    course_id int not null,
    day_of_week varchar(50) not null,
    start_time varchar(20) null,
    end_time varchar(20) null,
    location varchar(255) null,
    sort_order int not null default 0,
    created_at timestamp default CURRENT_TIMESTAMP null,
    primary key (id),
    constraint course_schedule_items_course_fk
        foreign key (course_id) references courses (id)
            on delete cascade
);

create index idx_course_schedule_items_course_id
    on course_schedule_items (course_id, sort_order);
