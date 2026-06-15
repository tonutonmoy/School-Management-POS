-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Create Exam', 'exam.create', 'Create examinations', 'exam'),
    ('Update Exam', 'exam.update', 'Update examinations', 'exam'),
    ('Delete Exam', 'exam.delete', 'Delete examinations', 'exam'),
    ('Publish Exam', 'exam.publish', 'Publish examinations', 'exam'),
    ('Enter Marks', 'marks.entry', 'Enter student marks', 'exam'),
    ('Update Marks', 'marks.update', 'Update student marks', 'exam'),
    ('Process Results', 'result.process', 'Process exam results', 'exam'),
    ('Publish Results', 'result.publish', 'Publish exam results', 'exam')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('exam.create', 'exam.update', 'exam.delete', 'exam.publish', 'marks.entry', 'marks.update', 'result.process', 'result.publish')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('exam.create', 'exam.update', 'exam.publish', 'marks.entry', 'marks.update', 'result.process', 'result.publish')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('marks.entry', 'marks.update', 'result.process')
WHERE r.slug = 'teacher'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN (
        'exam.create', 'exam.update', 'exam.delete', 'exam.publish',
        'marks.entry', 'marks.update', 'result.process', 'result.publish'
    )
);
DELETE FROM permissions WHERE slug IN (
    'exam.create', 'exam.update', 'exam.delete', 'exam.publish',
    'marks.entry', 'marks.update', 'result.process', 'result.publish'
);
