-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Manage Payments', 'payment.manage', 'Configure gateways and process online payments', 'payment'),
    ('Refund Payments', 'payment.refund', 'Approve and process payment refunds', 'payment'),
    ('View Payment Reports', 'payment.report.view', 'View payment transaction reports', 'payment')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('payment.manage', 'payment.refund', 'payment.report.view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('payment.manage', 'payment.report.view', 'payment.refund')
WHERE r.slug = 'accountant'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug = 'payment.report.view'
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('payment.manage', 'payment.refund', 'payment.report.view')
);
DELETE FROM permissions WHERE slug IN ('payment.manage', 'payment.refund', 'payment.report.view');
