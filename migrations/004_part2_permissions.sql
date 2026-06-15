-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('View Student', 'student.view', 'View student records and profiles', 'student'),
    ('Create Class', 'class.create', 'Create classes', 'class'),
    ('Update Class', 'class.update', 'Update classes', 'class'),
    ('Delete Class', 'class.delete', 'Delete classes', 'class'),
    ('Create Subject', 'subject.create', 'Create subjects', 'subject'),
    ('Update Subject', 'subject.update', 'Update subjects', 'subject'),
    ('Delete Subject', 'subject.delete', 'Delete subjects', 'subject')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('student.view', 'class.create', 'class.update', 'class.delete', 'subject.create', 'subject.update', 'subject.delete')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('student.view', 'student.create', 'student.update', 'student.delete', 'class.create', 'class.update', 'subject.create', 'subject.update')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('student.view', 'student.create', 'student.update')
WHERE r.slug = 'teacher'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN (
        'student.view', 'class.create', 'class.update', 'class.delete',
        'subject.create', 'subject.update', 'subject.delete'
    )
);
DELETE FROM permissions WHERE slug IN (
    'student.view', 'class.create', 'class.update', 'class.delete',
    'subject.create', 'subject.update', 'subject.delete'
);
