CREATE TABLE tags (
  id SERIAL PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE -- เช่น "AI", "Startup", "ฟุตบอล"
);