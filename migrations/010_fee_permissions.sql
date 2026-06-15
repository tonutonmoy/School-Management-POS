-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Create Fee', 'fee.create', 'Create fee types and structures', 'fees'),
    ('Update Fee', 'fee.update', 'Update fee types and structures', 'fees'),
    ('Delete Fee', 'fee.delete', 'Delete fee types and structures', 'fees'),
    ('Collect Fee', 'fee.collect', 'Collect fee payments', 'fees'),
    ('Refund Fee', 'fee.refund', 'Refund fee payments', 'fees'),
    ('View Fee Reports', 'fee.report.view', 'View financial reports', 'fees')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('fee.create', 'fee.update', 'fee.delete', 'fee.collect', 'fee.refund', 'fee.report.view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('fee.create', 'fee.update', 'fee.collect', 'fee.refund', 'fee.report.view')
WHERE r.slug = 'accountant'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('fee.collect', 'fee.report.view')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('fee.create', 'fee.update', 'fee.delete', 'fee.collect', 'fee.refund', 'fee.report.view')
);
DELETE FROM permissions WHERE slug IN ('fee.create', 'fee.update', 'fee.delete', 'fee.collect', 'fee.refund', 'fee.report.view');
