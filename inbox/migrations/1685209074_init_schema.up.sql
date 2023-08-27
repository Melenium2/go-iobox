create table if not exists __inbox_table
(
	id varchar(36) not null,
	status varchar(12),
	event_type varchar(255) not null,
	handler_key varchar(255) not null,
	payload bytea not null default '{}'::bytea,
 	created_at timestamp not null default (now() at time zone 'utc'),
	updated_at timestamp not null default (now() at time zone 'utc')
);

create unique index if not exists __inbox_uniq_id_handler_key_idx on __inbox_table (id, handler_key);
