CREATE TABLE news_tags (
  news_id BIGINT NOT NULL,
  tag_id INT NOT NULL,
  PRIMARY KEY (news_id, tag_id)
);