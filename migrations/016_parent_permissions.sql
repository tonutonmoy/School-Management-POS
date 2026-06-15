-- +goose Up
INSERT INTO roles (name, slug, description, is_system) VALUES
    ('Parent', 'parent', 'Parent portal access', TRUE);

INSERT INTO permissions (name, slug, description, module) VALUES
    ('Parent View', 'parent.view', 'Access parent portal and child data', 'parent'),
    ('Create Notice', 'notice.create', 'Create school notices', 'communication'),
    ('Update Notice', 'notice.update', 'Update school notices', 'communication'),
    ('Delete Notice', 'notice.delete', 'Delete school notices', 'communication'),
    ('Send Notification', 'notification.send', 'Send SMS and email notifications', 'communication')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('parent.view', 'notice.create', 'notice.update', 'notice.delete', 'notification.send')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug = 'parent.view'
WHERE r.slug = 'parent'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('notice.create', 'notice.update', 'notice.delete', 'notification.send')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('parent.view', 'notice.create', 'notice.update', 'notice.delete', 'notification.send')
);
DELETE FROM permissions WHERE slug IN ('parent.view', 'notice.create', 'notice.update', 'notice.delete', 'notification.send');
DELETE FROM roles WHERE slug = 'parent';
