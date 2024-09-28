CREATE TABLE songs (
    id SERIAL PRIMARY KEY,               
    group_name VARCHAR(255) NOT NULL,    
    song_name VARCHAR(255) NOT NULL,     
    release_date VARCHAR(255),                   
    text TEXT,                         
    link TEXT,                   
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW() 
);

-- Поиск по группе
CREATE INDEX idx_group_name ON songs (group_name);

-- Поиск по названию песни
CREATE INDEX idx_song_name ON songs (song_name);

-- Поиск по группе и названию песни одновременно
CREATE INDEX idx_group_and_song ON songs (group_name, song_name);
