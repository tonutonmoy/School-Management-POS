-- +goose Up
INSERT INTO permissions (name, slug, description, module) VALUES
    ('Review Admissions', 'admission.review', 'View and review online admission applications', 'admission'),
    ('Approve Admissions', 'admission.approve', 'Approve admission applications', 'admission'),
    ('Reject Admissions', 'admission.reject', 'Reject admission applications', 'admission'),
    ('Manage Website', 'website.manage', 'Manage school website CMS', 'website')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.slug = 'admin'
AND p.slug IN ('admission.review', 'admission.approve', 'admission.reject', 'website.manage')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('admission.review', 'admission.approve', 'admission.reject')
WHERE r.slug = 'principal'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r
JOIN permissions p ON p.slug IN ('admission.review', 'admission.approve', 'admission.reject')
WHERE r.slug IN ('staff', 'accountant')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE slug IN ('admission.review', 'admission.approve', 'admission.reject', 'website.manage')
);
DELETE FROM permissions WHERE slug IN ('admission.review', 'admission.approve', 'admission.reject', 'website.manage');
