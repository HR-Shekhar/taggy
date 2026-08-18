-- +goose Up
-- Seed one skill with roadmap, milestones, and community for local development / testing.

INSERT INTO skills (name, slug, description, is_active)
VALUES (
    'Web Development',
    'web-development',
    'Learn frontend and backend web development from fundamentals to job-ready projects.',
    TRUE
);

INSERT INTO roadmaps (skill_id)
VALUES ((SELECT id FROM skills WHERE slug = 'web-development'));

INSERT INTO roadmap_version (roadmap_id, version_number, status, generated_by, published_at)
VALUES (
    (SELECT id FROM roadmaps WHERE skill_id = (SELECT id FROM skills WHERE slug = 'web-development')),
    1,
    'ACTIVE',
    'ADMIN',
    NOW()
);

UPDATE roadmaps
SET current_version_id = (
    SELECT rv.id
    FROM roadmap_version rv
    INNER JOIN roadmaps r ON r.id = rv.roadmap_id
    INNER JOIN skills s ON s.id = r.skill_id
    WHERE s.slug = 'web-development'
    LIMIT 1
)
WHERE skill_id = (SELECT id FROM skills WHERE slug = 'web-development');

INSERT INTO milestones (roadmap_version_id, title, description, estimated_hours, order_index, difficulty)
SELECT rv.id, m.title, m.description, m.estimated_hours, m.order_index, m.difficulty
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
INNER JOIN skills s ON s.id = r.skill_id
CROSS JOIN (
    VALUES
        ('HTML & CSS Basics', 'Learn semantic HTML and CSS layout fundamentals.', 40, 1, 'BEGINNER'),
        ('JavaScript Fundamentals', 'Variables, functions, DOM, and async basics.', 60, 2, 'BEGINNER'),
        ('React Essentials', 'Components, hooks, and state management.', 80, 3, 'INTERMEDIATE'),
        ('Backend with Go', 'APIs, databases, and authentication.', 100, 4, 'INTERMEDIATE'),
        ('Capstone Project', 'Build and ship a full-stack portfolio project.', 120, 5, 'ADVANCED')
) AS m(title, description, estimated_hours, order_index, difficulty)
WHERE s.slug = 'web-development' AND rv.status = 'ACTIVE';

INSERT INTO community (skill_id, name, description)
VALUES (
    (SELECT id FROM skills WHERE slug = 'web-development'),
    'Web Development Community',
    'Official community for web development learners on Taggy.'
);

-- +goose Down
DELETE FROM community WHERE skill_id = (SELECT id FROM skills WHERE slug = 'web-development');
DELETE FROM milestones WHERE roadmap_version_id IN (
    SELECT rv.id FROM roadmap_version rv
    INNER JOIN roadmaps r ON r.id = rv.roadmap_id
    INNER JOIN skills s ON s.id = r.skill_id
    WHERE s.slug = 'web-development'
);
DELETE FROM roadmap_version WHERE roadmap_id IN (
    SELECT id FROM roadmaps WHERE skill_id = (SELECT id FROM skills WHERE slug = 'web-development')
);
DELETE FROM roadmaps WHERE skill_id = (SELECT id FROM skills WHERE slug = 'web-development');
DELETE FROM skills WHERE slug = 'web-development';
