-- +goose Up
INSERT INTO roles (name, slug, description, is_system) VALUES
    ('Admin', 'admin', 'Full system administrator', TRUE),
    ('Principal', 'principal', 'School principal', TRUE),
    ('Accountant', 'accountant', 'Finance and fees management', TRUE),
    ('Teacher', 'teacher', 'Teaching staff', TRUE),
    ('Staff', 'staff', 'General school staff', TRUE);

INSERT INTO permissions (name, slug, description, module) VALUES
    ('Create Student', 'student.create', 'Create student records', 'student'),
    ('Update Student', 'student.update', 'Update student records', 'student'),
    ('Delete Student', 'student.delete', 'Delete student records', 'student'),
    ('Create Teacher', 'teacher.create', 'Create teacher records', 'teacher'),
    ('Update Teacher', 'teacher.update', 'Update teacher records', 'teacher'),
    ('Collect Fees', 'fees.collect', 'Collect school fees', 'fees'),
    ('Refund Fees', 'fees.refund', 'Process fee refunds', 'fees'),
    ('Manage Users', 'user.manage', 'Create, update, delete users', 'user'),
    ('Manage Roles', 'role.manage', 'Manage roles and permissions', 'role'),
    ('Manage School', 'school.manage', 'Manage school setup', 'school'),
    ('Manage Sessions', 'session.manage', 'Manage academic sessions', 'session'),
    ('View Dashboard', 'dashboard.view', 'View dashboard', 'dashboard'),
    ('View Audit Logs', 'audit.view', 'View audit logs', 'audit');

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'admin';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.slug IN (
    'student.create', 'student.update', 'student.delete',
    'teacher.create', 'teacher.update',
    'dashboard.view'
)
WHERE r.slug = 'principal';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.slug IN ('fees.collect', 'fees.refund', 'dashboard.view')
WHERE r.slug = 'accountant';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.slug IN ('student.create', 'student.update', 'dashboard.view')
WHERE r.slug = 'teacher';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.slug IN ('dashboard.view')
WHERE r.slug = 'staff';

-- +goose Down
DELETE FROM role_permissions;
DELETE FROM permissions;
DELETE FROM roles;
