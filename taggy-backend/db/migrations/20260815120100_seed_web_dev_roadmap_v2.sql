-- +goose Up
-- Add a second published roadmap version for Web Development so learners can
-- compare / switch versions. v1 stays ACTIVE (current). v2 is ARCHIVED.

INSERT INTO roadmap_version (roadmap_id, version_number, status, generated_by, published_at)
SELECT r.id, 2, 'ARCHIVED', 'ADMIN', NOW()
FROM roadmaps r
INNER JOIN skills s ON s.id = r.skill_id
WHERE s.slug = 'web-development'
  AND NOT EXISTS (
      SELECT 1
      FROM roadmap_version rv
      WHERE rv.roadmap_id = r.id
        AND rv.version_number = 2
  );

INSERT INTO milestones (
    roadmap_version_id,
    title,
    slug,
    description,
    estimated_hours,
    order_index,
    difficulty
)
SELECT
    rv.id,
    m.title,
    m.slug,
    m.description,
    m.estimated_hours,
    m.order_index,
    m.difficulty
FROM roadmap_version rv
INNER JOIN roadmaps r ON r.id = rv.roadmap_id
INNER JOIN skills s ON s.id = r.skill_id
CROSS JOIN (
    VALUES
        ('HTML & CSS Basics', 'html-css-basics', 'Semantic HTML, CSS layout, and responsive fundamentals.', 40, 1, 'BEGINNER'),
        ('JavaScript Fundamentals', 'javascript-fundamentals', 'Variables, functions, DOM, and async basics.', 60, 2, 'BEGINNER'),
        ('TypeScript Fundamentals', 'typescript-fundamentals', 'Types, interfaces, and safer frontend code.', 50, 3, 'INTERMEDIATE'),
        ('React Essentials', 'react-essentials', 'Components, hooks, and state management.', 80, 4, 'INTERMEDIATE'),
        ('Backend with Go', 'backend-with-go', 'APIs, databases, and authentication.', 100, 5, 'INTERMEDIATE'),
        ('Capstone Project', 'capstone-project', 'Build and ship a full-stack portfolio project.', 120, 6, 'ADVANCED')
) AS m(title, slug, description, estimated_hours, order_index, difficulty)
WHERE s.slug = 'web-development'
  AND rv.version_number = 2
  AND NOT EXISTS (
      SELECT 1 FROM milestones mil WHERE mil.roadmap_version_id = rv.id
  );

-- +goose Down
DELETE FROM milestones
WHERE roadmap_version_id IN (
    SELECT rv.id
    FROM roadmap_version rv
    INNER JOIN roadmaps r ON r.id = rv.roadmap_id
    INNER JOIN skills s ON s.id = r.skill_id
    WHERE s.slug = 'web-development'
      AND rv.version_number = 2
);

DELETE FROM roadmap_version
WHERE roadmap_id IN (
    SELECT r.id
    FROM roadmaps r
    INNER JOIN skills s ON s.id = r.skill_id
    WHERE s.slug = 'web-development'
)
AND version_number = 2;
