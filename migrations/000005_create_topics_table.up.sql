CREATE TABLE IF NOT EXISTS topics (
    id         UUID PRIMARY KEY,
    name_cn    VARCHAR(100) NOT NULL,
    name_vi    VARCHAR(100) NOT NULL,
    name_en    VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vocabulary_topics (
    vocabulary_id UUID NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    topic_id      UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (vocabulary_id, topic_id)
);

CREATE INDEX IF NOT EXISTS idx_vocabulary_topics_topic_id ON vocabulary_topics(topic_id);

-- Seed 10 system topics
INSERT INTO topics (id, name_cn, name_vi, name_en, slug, sort_order) VALUES
    ('a1000000-0000-0000-0000-000000000001', '日常生活', 'Cuộc sống hằng ngày', 'Daily Life', 'daily-life', 1),
    ('a1000000-0000-0000-0000-000000000002', '饮食', 'Ẩm thực', 'Food & Drink', 'food-drink', 2),
    ('a1000000-0000-0000-0000-000000000003', '交通出行', 'Giao thông', 'Transportation', 'transportation', 3),
    ('a1000000-0000-0000-0000-000000000004', '购物', 'Mua sắm', 'Shopping', 'shopping', 4),
    ('a1000000-0000-0000-0000-000000000005', '工作', 'Công việc', 'Work & Career', 'work-career', 5),
    ('a1000000-0000-0000-0000-000000000006', '教育', 'Giáo dục', 'Education', 'education', 6),
    ('a1000000-0000-0000-0000-000000000007', '健康', 'Sức khỏe', 'Health', 'health', 7),
    ('a1000000-0000-0000-0000-000000000008', '旅游', 'Du lịch', 'Travel', 'travel', 8),
    ('a1000000-0000-0000-0000-000000000009', '文化', 'Văn hóa', 'Culture', 'culture', 9),
    ('a1000000-0000-0000-0000-000000000010', '科技', 'Khoa học công nghệ', 'Technology', 'technology', 10)
ON CONFLICT (slug) DO NOTHING;
