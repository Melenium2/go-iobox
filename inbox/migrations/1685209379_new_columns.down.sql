alter table if exists __inbox_table
	drop column if exists attempt,
	drop column if exists error_message,
	drop column if exists next_attempt;
