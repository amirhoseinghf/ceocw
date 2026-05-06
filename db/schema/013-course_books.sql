create table course_books
(
    course_id int not null,
    book_id   int not null,
    primary key (course_id, book_id),
    constraint course_books_ibfk_1
        foreign key (course_id) references courses (id)
            on delete cascade,
    constraint course_books_ibfk_2
        foreign key (book_id) references books (id)
            on delete cascade
);

create index book_id
    on course_books (book_id);

