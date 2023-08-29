CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS segments (
	id		uuid default gen_random_uuid() primary key,
    slug	text unique not null
);

CREATE TABLE IF NOT EXISTS segments_users 
(
	id 			uuid default gen_random_uuid() primary key,
    segment_id	uuid references	public.segments (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id		uuid not null,
    
    -- CONSTRAINT	segments_users_pkey PRIMARY KEY (segment_id, user_id)
    UNIQUE(user_id,segment_id) -- ограничение на уникальность id при котором данные не перезатираются
);

create type segment_operation as enum ('create', 'delete');

CREATE TABLE IF NOT EXISTS operations_outbox
(
	id         		uuid default gen_random_uuid() primary key,
	user_id    		uuid not null,
	segment    		text not null,
	operation  		segment_operation not null,
	operation_at 	timestamp not null
);
