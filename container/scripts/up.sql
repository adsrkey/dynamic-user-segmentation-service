CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
	id		uuid  primary key
);

CREATE TABLE IF NOT EXISTS segments (
	id		uuid default gen_random_uuid() primary key,
    slug	text unique not null
);

CREATE TABLE IF NOT EXISTS segments_users 
(
	id 			uuid default gen_random_uuid() primary key,
    segment_id	uuid references	public.segments (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id		uuid references	public.users (id) ON UPDATE CASCADE ON DELETE CASCADE,
    
    -- CONSTRAINT	segments_users_pkey PRIMARY KEY (segment_id, user_id)
    UNIQUE(user_id,segment_id) -- ограничение на уникальность id при котором данные не перезатираются
);

create type segment_operation as enum ('create', 'delete');

CREATE TABLE IF NOT EXISTS operations_outbox
(
	id         		uuid default gen_random_uuid() primary key,
	user_id    		uuid references	public.users (id) ON UPDATE CASCADE ON DELETE CASCADE,
	segment    		text not null,
	operation  		segment_operation not null,
	operation_at 	timestamp not null,
	UNIQUE(operation_at,segment,user_id)
);

CREATE INDEX segments_slug_index
    ON public.segments (slug);
