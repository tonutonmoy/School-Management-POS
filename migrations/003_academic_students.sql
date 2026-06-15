-- +goose Up
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_departments_slug_active ON departments (slug) WHERE deleted_at IS NULL;

CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL,
    description TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_classes_code_active ON classes (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_classes_sort_order ON classes (sort_order);

CREATE TABLE sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    class_id UUID NOT NULL REFERENCES classes(id),
    name VARCHAR(50) NOT NULL,
    capacity INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_sections_class_name_active ON sections (class_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_sections_class_id ON sections (class_id);

CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    code VARCHAR(30) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_subjects_code_active ON subjects (code) WHERE deleted_at IS NULL;

CREATE TABLE class_subjects (
    class_id UUID NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (class_id, subject_id)
);

CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admission_number VARCHAR(30) NOT NULL,
    roll_number VARCHAR(30),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(20) NOT NULL,
    blood_group VARCHAR(10),
    religion VARCHAR(50),
    nationality VARCHAR(50) DEFAULT 'Bangladeshi',
    photo_url TEXT,
    phone VARCHAR(30),
    email VARCHAR(255),
    address TEXT,
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    section_id UUID NOT NULL REFERENCES sections(id),
    department_id UUID REFERENCES departments(id),
    admission_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_students_status CHECK (status IN ('active', 'inactive', 'graduated', 'transferred'))
);

CREATE UNIQUE INDEX idx_students_admission_active ON students (admission_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_students_session_id ON students (session_id);
CREATE INDEX idx_students_class_id ON students (class_id);
CREATE INDEX idx_students_section_id ON students (section_id);
CREATE INDEX idx_students_status ON students (status);
CREATE INDEX idx_students_admission_date ON students (admission_date);
CREATE INDEX idx_students_name ON students (last_name, first_name);

CREATE TABLE student_parents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    father_name VARCHAR(150),
    father_phone VARCHAR(30),
    father_occupation VARCHAR(100),
    mother_name VARCHAR(150),
    mother_phone VARCHAR(30),
    mother_occupation VARCHAR(100),
    guardian_name VARCHAR(150),
    guardian_phone VARCHAR(30),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE student_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    doc_type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_student_documents_type CHECK (doc_type IN ('birth_certificate', 'previous_marksheet', 'passport_photo', 'other'))
);

CREATE INDEX idx_student_documents_student_id ON student_documents (student_id);

CREATE TABLE student_promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    promotion_type VARCHAR(20) NOT NULL,
    from_session_id UUID REFERENCES academic_sessions(id),
    to_session_id UUID REFERENCES academic_sessions(id),
    from_class_id UUID REFERENCES classes(id),
    to_class_id UUID REFERENCES classes(id),
    from_section_id UUID REFERENCES sections(id),
    to_section_id UUID REFERENCES sections(id),
    promotion_date DATE NOT NULL DEFAULT CURRENT_DATE,
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_student_promotions_type CHECK (promotion_type IN ('promote', 'transfer'))
);

CREATE INDEX idx_student_promotions_student_id ON student_promotions (student_id);

CREATE TABLE admission_sequences (
    year INT PRIMARY KEY,
    last_number INT NOT NULL DEFAULT 0
);

INSERT INTO departments (name, slug, description) VALUES
    ('Science', 'science', 'Science department'),
    ('Commerce', 'commerce', 'Commerce department'),
    ('Arts', 'arts', 'Arts department');

-- +goose Down
DROP TABLE IF EXISTS student_promotions;
DROP TABLE IF EXISTS student_documents;
DROP TABLE IF EXISTS student_parents;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS admission_sequences;
DROP TABLE IF EXISTS class_subjects;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS sections;
DROP TABLE IF EXISTS classes;
DROP TABLE IF EXISTS departments;
