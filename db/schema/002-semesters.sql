create table semesters
(
    id     bigint unsigned auto_increment
        primary key,
    season varchar(20) not null,
    year   int         not null,
    constraint id
        unique (id),
    constraint season
        unique (season, year),
    constraint chk_season
        check (`season` in ('spring', 'fall'))
);

