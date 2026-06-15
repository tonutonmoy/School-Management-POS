-- +goose Up
ALTER TABLE departments ADD COLUMN IF NOT EXISTS dept_type VARCHAR(20) NOT NULL DEFAULT 'employee';
UPDATE departments SET dept_type = 'academic' WHERE slug IN ('science', 'commerce', 'arts');

CREATE TABLE designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_designations_slug_active ON designations (slug) WHERE deleted_at IS NULL;

CREATE TABLE employee_sequences (
    entity_type VARCHAR(20) NOT NULL,
    year INT NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    PRIMARY KEY (entity_type, year)
);

CREATE TABLE teachers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id VARCHAR(30) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    photo_url TEXT,
    gender VARCHAR(20) NOT NULL,
    date_of_birth DATE,
    blood_group VARCHAR(10),
    religion VARCHAR(50),
    nationality VARCHAR(50) DEFAULT 'Bangladeshi',
    phone VARCHAR(30),
    email VARCHAR(255),
    address TEXT,
    national_id VARCHAR(50),
    joining_date DATE NOT NULL DEFAULT CURRENT_DATE,
    department_id UUID REFERENCES departments(id),
    designation_id UUID REFERENCES designations(id),
    qualification TEXT,
    experience TEXT,
    salary NUMERIC(12,2),
    employment_type VARCHAR(20) NOT NULL DEFAULT 'full_time',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_teachers_status CHECK (status IN ('active', 'inactive', 'resigned')),
    CONSTRAINT chk_teachers_employment CHECK (employment_type IN ('full_time', 'part_time', 'contract'))
);

CREATE UNIQUE INDEX idx_teachers_employee_active ON teachers (employee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_teachers_department ON teachers (department_id);
CREATE INDEX idx_teachers_designation ON teachers (designation_id);
CREATE INDEX idx_teachers_status ON teachers (status);
CREATE INDEX idx_teachers_email ON teachers (LOWER(email)) WHERE deleted_at IS NULL;

CREATE TABLE teacher_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES subjects(id) ON DELETE CASCADE,
    class_id UUID REFERENCES classes(id) ON DELETE CASCADE,
    section_id UUID REFERENCES sections(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_teacher_assignment CHECK (subject_id IS NOT NULL OR class_id IS NOT NULL OR section_id IS NOT NULL)
);

CREATE INDEX idx_teacher_assignments_teacher ON teacher_assignments (teacher_id);
CREATE UNIQUE INDEX idx_teacher_assignments_unique ON teacher_assignments (
    teacher_id, COALESCE(subject_id, '00000000-0000-0000-0000-000000000000'::uuid),
    COALESCE(class_id, '00000000-0000-0000-0000-000000000000'::uuid),
    COALESCE(section_id, '00000000-0000-0000-0000-000000000000'::uuid)
);

CREATE TABLE teacher_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES subjects(id) ON DELETE SET NULL,
    class_id UUID REFERENCES classes(id) ON DELETE SET NULL,
    section_id UUID REFERENCES sections(id) ON DELETE SET NULL,
    day_of_week INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    room VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_teacher_schedules_teacher_day ON teacher_schedules (teacher_id, day_of_week);

CREATE TABLE teacher_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    doc_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_teacher_documents_type CHECK (doc_type IN ('nid', 'certificate', 'cv', 'appointment_letter', 'other'))
);

CREATE INDEX idx_teacher_documents_teacher ON teacher_documents (teacher_id);

CREATE TABLE staffs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id VARCHAR(30) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    photo_url TEXT,
    phone VARCHAR(30),
    email VARCHAR(255),
    address TEXT,
    department_id UUID REFERENCES departments(id),
    designation_id UUID REFERENCES designations(id),
    salary NUMERIC(12,2),
    joining_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_staffs_status CHECK (status IN ('active', 'inactive', 'resigned'))
);

CREATE UNIQUE INDEX idx_staffs_employee_active ON staffs (employee_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_staffs_department ON staffs (department_id);
CREATE INDEX idx_staffs_status ON staffs (status);

CREATE TABLE staff_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    doc_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_staff_documents_type CHECK (doc_type IN ('nid', 'certificate', 'cv', 'appointment_letter', 'other'))
);

CREATE INDEX idx_staff_documents_staff ON staff_documents (staff_id);

INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Administration', 'administration', 'School administration', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'administration');
INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Accounts', 'accounts', 'Finance and accounts', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'accounts');
INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Library', 'library', 'Library services', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'library');
INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Transport', 'transport', 'Transport department', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'transport');
INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Security', 'security', 'Security services', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'security');
INSERT INTO departments (name, slug, description, dept_type)
SELECT 'Maintenance', 'maintenance', 'Facilities maintenance', 'employee'
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE slug = 'maintenance');

INSERT INTO designations (name, slug, category) VALUES
    ('Principal', 'principal', 'leadership'),
    ('Vice Principal', 'vice-principal', 'leadership'),
    ('Senior Teacher', 'senior-teacher', 'teaching'),
    ('Assistant Teacher', 'assistant-teacher', 'teaching'),
    ('Accountant', 'accountant', 'administrative'),
    ('Librarian', 'librarian', 'support'),
    ('Driver', 'driver', 'support'),
    ('Office Assistant', 'office-assistant', 'support');

-- +goose Down
DROP TABLE IF EXISTS staff_documents;
DROP TABLE IF EXISTS staffs;
DROP TABLE IF EXISTS teacher_documents;
DROP TABLE IF EXISTS teacher_schedules;
DROP TABLE IF EXISTS teacher_assignments;
DROP TABLE IF EXISTS teachers;
DROP TABLE IF EXISTS employee_sequences;
DROP TABLE IF EXISTS designations;
ALTER TABLE departments DROP COLUMN IF EXISTS dept_type;
