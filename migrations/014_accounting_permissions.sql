-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('View Accounting', 'accounting.view', 'View chart of accounts, ledger, and books', 'accounting'),
    ('Manage Accounting', 'accounting.manage', 'Manage accounts, journal entries, and periods', 'accounting'),
    ('Create Expense', 'expense.create', 'Create and submit expenses', 'accounting'),
    ('Approve Expense', 'expense.approve', 'Approve and post expenses', 'accounting'),
    ('View Finance Reports', 'finance.report.view', 'View trial balance, income statement, balance sheet', 'accounting')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('accounting.view', 'accounting.manage', 'expense.create', 'expense.approve', 'finance.report.view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('accounting.view', 'accounting.manage', 'expense.create', 'expense.approve', 'finance.report.view', 'fee.collect', 'fee.report.view')
WHERE r.slug = 'accountant'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('accounting.view', 'finance.report.view', 'expense.approve')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('accounting.view', 'accounting.manage', 'expense.create', 'expense.approve', 'finance.report.view')
);
DELETE FROM permissions WHERE slug IN ('accounting.view', 'accounting.manage', 'expense.create', 'expense.approve', 'finance.report.view');
