-- +goose Up
CREATE TABLE application_sequences (
    year INT PRIMARY KEY,
    last_number INT NOT NULL DEFAULT 0
);

CREATE TABLE admission_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_number VARCHAR(30) NOT NULL,
    tracking_token VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10) NOT NULL,
    blood_group VARCHAR(10),
    religion VARCHAR(50),
    nationality VARCHAR(50) DEFAULT 'Bangladeshi',
    phone VARCHAR(30),
    email VARCHAR(255),
    address TEXT,
    father_name VARCHAR(150),
    father_phone VARCHAR(30),
    father_occupation VARCHAR(100),
    mother_name VARCHAR(150),
    mother_phone VARCHAR(30),
    mother_occupation VARCHAR(100),
    guardian_name VARCHAR(150),
    guardian_phone VARCHAR(30),
    previous_school VARCHAR(200),
    previous_class VARCHAR(50),
    previous_board VARCHAR(100),
    session_id UUID REFERENCES academic_sessions(id) ON DELETE SET NULL,
    class_id UUID REFERENCES classes(id) ON DELETE SET NULL,
    section_id UUID REFERENCES sections(id) ON DELETE SET NULL,
    admission_fee_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    payment_reference VARCHAR(100),
    receipt_number VARCHAR(30),
    review_notes TEXT,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    student_id UUID REFERENCES students(id) ON DELETE SET NULL,
    parent_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_admission_status CHECK (status IN ('pending', 'under_review', 'approved', 'rejected', 'admitted')),
    CONSTRAINT chk_admission_payment CHECK (payment_status IN ('unpaid', 'pending', 'paid', 'waived'))
);

CREATE UNIQUE INDEX idx_admission_app_number ON admission_applications (application_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_admission_status ON admission_applications (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_admission_created ON admission_applications (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_admission_session ON admission_applications (session_id) WHERE deleted_at IS NULL;

CREATE TABLE admission_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES admission_applications(id) ON DELETE CASCADE,
    doc_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_admission_doc_type CHECK (doc_type IN ('birth_certificate', 'previous_marksheet', 'passport_photo', 'transfer_certificate', 'other'))
);

CREATE INDEX idx_admission_documents_app ON admission_documents (application_id) WHERE deleted_at IS NULL;

CREATE TABLE website_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_name VARCHAR(200) NOT NULL DEFAULT 'School',
    tagline VARCHAR(300),
    logo_url TEXT,
    favicon_url TEXT,
    primary_color VARCHAR(20) DEFAULT '#4f46e5',
    secondary_color VARCHAR(20) DEFAULT '#0f172a',
    facebook_url TEXT,
    twitter_url TEXT,
    instagram_url TEXT,
    youtube_url TEXT,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(30),
    contact_address TEXT,
    default_meta_title VARCHAR(200),
    default_meta_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE website_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(150) NOT NULL,
    title VARCHAR(200) NOT NULL,
    page_type VARCHAR(30) NOT NULL DEFAULT 'custom',
    content TEXT NOT NULL DEFAULT '',
    meta_title VARCHAR(200),
    meta_description TEXT,
    og_image TEXT,
    is_published BOOLEAN NOT NULL DEFAULT true,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_page_type CHECK (page_type IN ('home', 'about', 'principal', 'admission', 'teachers', 'contact', 'custom'))
);

CREATE UNIQUE INDEX idx_website_pages_slug ON website_pages (slug) WHERE deleted_at IS NULL;

CREATE TABLE website_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID REFERENCES website_pages(id) ON DELETE CASCADE,
    block_type VARCHAR(30) NOT NULL DEFAULT 'text',
    title VARCHAR(200),
    content TEXT NOT NULL DEFAULT '',
    image_url TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_block_type CHECK (block_type IN ('text', 'hero', 'feature', 'cta', 'html', 'image'))
);

CREATE INDEX idx_website_blocks_page ON website_blocks (page_id, sort_order) WHERE deleted_at IS NULL;

CREATE TABLE website_banners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    subtitle TEXT,
    image_url TEXT NOT NULL,
    link_url TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE website_gallery (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    caption TEXT,
    image_url TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE news (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(200) NOT NULL,
    title VARCHAR(300) NOT NULL,
    excerpt TEXT,
    body TEXT NOT NULL,
    image_url TEXT,
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMPTZ,
    meta_title VARCHAR(200),
    meta_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_news_slug ON news (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_news_published ON news (is_published, published_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(200) NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(300),
    event_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ,
    image_url TEXT,
    is_published BOOLEAN NOT NULL DEFAULT false,
    meta_title VARCHAR(200),
    meta_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_events_slug ON events (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_events_date ON events (event_date) WHERE deleted_at IS NULL;

CREATE TABLE downloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(200) NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT,
    category VARCHAR(30) NOT NULL DEFAULT 'other',
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    download_count INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_download_category CHECK (category IN ('prospectus', 'circular', 'form', 'document', 'other'))
);

CREATE UNIQUE INDEX idx_downloads_slug ON downloads (slug) WHERE deleted_at IS NULL;

CREATE TABLE contact_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(30),
    subject VARCHAR(300) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    replied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_contact_status CHECK (status IN ('new', 'read', 'replied', 'closed'))
);

CREATE INDEX idx_contact_status ON contact_messages (status, created_at DESC) WHERE deleted_at IS NULL;

CREATE TABLE website_visits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path VARCHAR(500) NOT NULL,
    visited_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_website_visits_date ON website_visits (visited_at);

INSERT INTO website_settings (site_name, tagline, default_meta_title, default_meta_description) VALUES
    ('School Management POS', 'Excellence in Education', 'Welcome to Our School', 'Quality education for every student');

INSERT INTO website_pages (slug, title, page_type, content, is_published, sort_order) VALUES
    ('home', 'Home', 'home', '<p>Welcome to our school.</p>', true, 1),
    ('about', 'About Us', 'about', '<p>About our institution.</p>', true, 2),
    ('principal-message', 'Principal Message', 'principal', '<p>Message from the principal.</p>', true, 3),
    ('admission-info', 'Admission Information', 'admission', '<p>How to apply for admission.</p>', true, 4),
    ('teachers', 'Our Teachers', 'teachers', '<p>Meet our faculty.</p>', true, 5),
    ('contact', 'Contact Us', 'contact', '<p>Get in touch with us.</p>', true, 6);

-- +goose Down
DROP TABLE IF EXISTS website_visits;
DROP TABLE IF EXISTS contact_messages;
DROP TABLE IF EXISTS downloads;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS news;
DROP TABLE IF EXISTS website_gallery;
DROP TABLE IF EXISTS website_banners;
DROP TABLE IF EXISTS website_blocks;
DROP TABLE IF EXISTS website_pages;
DROP TABLE IF EXISTS website_settings;
DROP TABLE IF EXISTS admission_documents;
DROP TABLE IF EXISTS admission_applications;
DROP TABLE IF EXISTS application_sequences;
