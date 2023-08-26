create table if not exists __outbox_table (
    id varchar(36) not null primary key,
    status varchar(12),
    event_type varchar(255) not null,
    payload jsonb not null,
    created_at timestamp not null default (now() at time zone 'utc'),
    updated_at timestamp not null default (now() at time zone 'utc')
);
