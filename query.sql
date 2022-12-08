-- ALTER TABLE todo
-- ADD COLUMN remind_at TIMESTAMP NOT NULL DEFAULT NOW()

-- ALTER TABLE todo
-- ADD COLUMN is_reminded NOT NULL  BOOL DEFAULT FALSE

-- ALTER TABLE users
-- ADD COLUMN telegram_bot_token   text


-- INSERT INTO users (name,lim,telegram_chat_id,telegram_bot_token) VALUES ('Ali',1000,'1048346950','5596315438:AAHHHkxalUMl20KFEFZs-3lXNM4jMwn4MKs')
-- INSERT INTO todo (user_id,content,remind_at,is_reminded) VALUES (1,'hi',current_timestamp,true)
-- CREATE TABLE todo(
-- 	id SERIAL PRIMARY KEY,
-- 	user_id integer required,
-- 	content TEXT NOT NULL,
-- 	completed BOOL DEFAULT FALSE,
-- 	created_at TIMESTAMP DEFAULT NOW(),
-- 	remind_at TIMESTAMP NOT NULL,
-- 	is_reminded BOOL DEFAULT FALSE
-- )
-- CREATE TABLE users(
-- 	id serial primary key,
-- 	name text unique NOT NULL,
-- 	lim INTEGER DEFAULT 15,
-- 	telegram_chat_id TEXT DEFAULT '',
-- 	telegram_bot_token TEXT DEFAULT ''
-- )