create table course_users
(
    course_id int not null,
    user_id   int not null,
    role      enum ('ta', 'head_ta') not null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    primary key (course_id, user_id),
    constraint course_users_course_fk
        foreign key (course_id) references courses (id)
            on delete cascade,
    constraint course_users_user_fk
        foreign key (user_id) references users (id)
            on delete cascade
);

create index idx_course_users_user_id
    on course_users (user_id);

create index idx_course_users_role
    on course_users (role);
