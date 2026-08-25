DELETE TABLE items;
DELETE TABLE tags;
DELETE TABLE items_tags;

DELETE INDEX idx_items_tags_tag_id;
DELETE INDEX idx_items_attributes;