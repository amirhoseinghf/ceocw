alter table announcements
    add column created_by_user_id int null after course_id;

alter table announcements
    add constraint announcements_created_by_user_fk
        foreign key (created_by_user_id) references users (id)
        on delete set null;

create index idx_announcements_created_by_user_id
    on announcements (created_by_user_id);
