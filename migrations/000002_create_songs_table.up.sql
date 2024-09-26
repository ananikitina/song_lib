CREATE TABLE songs (
    id SERIAL PRIMARY KEY,               
    group_name VARCHAR(255) NOT NULL,    
    song_name VARCHAR(255) NOT NULL,     
    release_date DATE,                   
    lyrics TEXT,                         
    link VARCHAR(255),                   
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW() 
);
