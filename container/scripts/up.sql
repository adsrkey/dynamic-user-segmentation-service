CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users 
(
	id 		uuid unique not null primary key
);

CREATE TABLE IF NOT EXISTS segments (
	id 	uuid default gen_random_uuid() primary key,
    slug    	text unique not null
);

CREATE TABLE IF NOT EXISTS segments_users 
(
	id 	uuid default gen_random_uuid() primary key,
    segment_id	uuid	references	public.segments (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_id		uuid	references	users (id) ON UPDATE CASCADE ON DELETE cascade, 
    UNIQUE(user_id,segment_id)
); 
