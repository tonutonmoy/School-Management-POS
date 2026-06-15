-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Manage System', 'system.manage', 'System settings, health, and white label', 'system'),
    ('Manage Backups', 'system.backup', 'Create, restore, and download backups', 'system'),
    ('View Audit Center', 'system.audit', 'Search and export audit logs', 'system'),
    ('Manage License', 'system.license', 'Activate and manage license', 'system'),
    ('Security Dashboard', 'system.security', 'View security monitoring', 'system')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('system.manage', 'system.backup', 'system.audit', 'system.license', 'system.security')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('system.audit', 'system.security')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('system.manage', 'system.backup', 'system.audit', 'system.license', 'system.security')
);
DELETE FROM permissions WHERE slug IN ('system.manage', 'system.backup', 'system.audit', 'system.license', 'system.security');
