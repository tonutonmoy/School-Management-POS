-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('View Teacher', 'teacher.view', 'View teacher records', 'teacher'),
    ('Delete Teacher', 'teacher.delete', 'Delete teacher records', 'teacher'),
    ('Create Staff', 'staff.create', 'Create staff records', 'staff'),
    ('Update Staff', 'staff.update', 'Update staff records', 'staff'),
    ('Delete Staff', 'staff.delete', 'Delete staff records', 'staff'),
    ('View Staff', 'staff.view', 'View staff records', 'staff'),
    ('Create Department', 'department.create', 'Create departments', 'department'),
    ('Update Department', 'department.update', 'Update departments', 'department'),
    ('Delete Department', 'department.delete', 'Delete departments', 'department'),
    ('Create Designation', 'designation.create', 'Create designations', 'designation'),
    ('Update Designation', 'designation.update', 'Update designations', 'designation'),
    ('Delete Designation', 'designation.delete', 'Delete designations', 'designation')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN (
    'teacher.view', 'teacher.delete', 'staff.create', 'staff.update', 'staff.delete', 'staff.view',
    'department.create', 'department.update', 'department.delete',
    'designation.create', 'designation.update', 'designation.delete'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN (
    'teacher.view', 'teacher.create', 'teacher.update', 'staff.view', 'staff.create', 'staff.update',
    'department.create', 'department.update', 'designation.create', 'designation.update'
)
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN (
        'teacher.view', 'teacher.delete', 'staff.create', 'staff.update', 'staff.delete', 'staff.view',
        'department.create', 'department.update', 'department.delete',
        'designation.create', 'designation.update', 'designation.delete'
    )
);
DELETE FROM permissions WHERE slug IN (
    'teacher.view', 'teacher.delete', 'staff.create', 'staff.update', 'staff.delete', 'staff.view',
    'department.create', 'department.update', 'department.delete',
    'designation.create', 'designation.update', 'designation.delete'
);
