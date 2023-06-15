	alter table if exists __inbox_table
		add column attempt smallint not null default 0,
	 	add column error_message text,
	 	add column next_attempt timestamp;
