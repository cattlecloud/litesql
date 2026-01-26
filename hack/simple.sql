pragma encoding     = 'UTF-8';
pragma foreign_keys = ON;
pragma temp_store   = MEMORY;
pragma max_length   = 1048576; -- 1 MiB
pragma journal_mode = OFF;
pragma synchronous  = NORMAL;
pragma busy_timeout = 1000;
pragma cache_size   = -4000; -- 4 MB
pragma auto_vacuum  = INCREMENTAL;
pragma mmap_size    = 4194304; -- 4 MiB

BEGIN;

create table if not exists oauth (
    id integer primary key,
    -- enum

    provider text check(length(provider) < 16) not null unique
    -- name of the oauth identity provider
) strict;

insert or ignore into oauth(id, provider) values
    (1, 'Testing'),
    (2, 'Apple'),
    (3, 'Google'),
    (4, 'Microsoft');

create table if not exists standing (
    id integer primary key,
    -- enum

    label text check(length(label) < 16) not null unique
    -- readable name of the user's account standing
) strict;

insert or ignore into standing(id, label) values
    (1, 'admin'),
    (2, 'normal'),
    (3, 'moderator'),
    (4, 'banned');

create table if not exists users (
    id integer primary key autoincrement,
    -- auto

    oauth_id integer not null,
    -- foreign key into oauth provider

    username text check(length(username) > 1 and length(username) < 20) not null unique,
    -- chosen username

    email text check(length(email) < 128) not null,
    -- chosen email address

    standing_id integer not null default 2,
    -- foreign key into standing table

    creation integer default (cast(unixepoch('now') as int)) not null,
    -- creation timestamp

    foreign key (oauth_id)    references oauth(id)
    foreign key (standing_id) references standing(id)
) strict;

insert or ignore into users (oauth_id, username, email, standing_id) values
    (3, 'Admin', 'admin@example.org', 1),
    (1, 'Beth',  'beth@example.org',  3),
    (2, 'Carl',  'carl@example.org',  2),
    (2, 'David', 'dave@example.org',  2),
    (4, 'Eve',   'eve@example.org',   4);

COMMIT;
