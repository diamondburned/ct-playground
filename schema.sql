CREATE TYPE log_type AS ENUM ('prod', 'test');

-- Mapped from Go:
-- https://pkg.go.dev/github.com/google/certificate-transparency-go@v1.1.6/loglist3#Log
CREATE TABLE logs (
	log_id blob not null primary key,
	pub_key blob not null,
	api_url text not null,
	dns_url text not null,
	mmd integer not null,
	status smallint not null, -- loglist3.logstates
	type log_type not null,
	state jsonb not null,
	previous_operators jsonb not null
);

CREATE TPYE log_leaf_type AS ENUM ('timestamped_entry');

CREATE TABLE log_leaves (
	parent_log_id blob not null references logs(log_id),
	

);
