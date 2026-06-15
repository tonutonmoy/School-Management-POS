-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Mark Student Attendance', 'attendance.student.mark', 'Mark student attendance', 'attendance'),
    ('View Student Attendance', 'attendance.student.view', 'View student attendance', 'attendance'),
    ('Mark Teacher Attendance', 'attendance.teacher.mark', 'Mark teacher attendance', 'attendance'),
    ('View Teacher Attendance', 'attendance.teacher.view', 'View teacher attendance', 'attendance'),
    ('Mark Staff Attendance', 'attendance.staff.mark', 'Mark staff attendance', 'attendance'),
    ('View Staff Attendance', 'attendance.staff.view', 'View staff attendance', 'attendance'),
    ('Approve Leave', 'leave.approve', 'Approve leave requests', 'leave'),
    ('Reject Leave', 'leave.reject', 'Reject leave requests', 'leave')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN (
    'attendance.student.mark', 'attendance.student.view',
    'attendance.teacher.mark', 'attendance.teacher.view',
    'attendance.staff.mark', 'attendance.staff.view',
    'leave.approve', 'leave.reject'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN (
    'attendance.student.mark', 'attendance.student.view',
    'attendance.teacher.mark', 'attendance.teacher.view',
    'attendance.staff.mark', 'attendance.staff.view',
    'leave.approve', 'leave.reject'
)
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN (
    'attendance.student.mark', 'attendance.student.view',
    'attendance.teacher.mark', 'attendance.teacher.view'
)
WHERE r.slug = 'teacher'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN (
    'attendance.staff.mark', 'attendance.staff.view'
)
WHERE r.slug = 'staff'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN (
        'attendance.student.mark', 'attendance.student.view',
        'attendance.teacher.mark', 'attendance.teacher.view',
        'attendance.staff.mark', 'attendance.staff.view',
        'leave.approve', 'leave.reject'
    )
);
DELETE FROM permissions WHERE slug IN (
    'attendance.student.mark', 'attendance.student.view',
    'attendance.teacher.mark', 'attendance.teacher.view',
    'attendance.staff.mark', 'attendance.staff.view',
    'leave.approve', 'leave.reject'
);
