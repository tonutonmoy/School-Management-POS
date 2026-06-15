-- +goose Up
CREATE TABLE grading_systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE grading_scales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES grading_systems(id) ON DELETE CASCADE,
    grade VARCHAR(5) NOT NULL,
    min_percentage NUMERIC(5,2) NOT NULL,
    max_percentage NUMERIC(5,2) NOT NULL,
    gpa_point NUMERIC(3,2) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_grading_scales_system ON grading_scales (system_id);

CREATE TABLE exams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    exam_type VARCHAR(50) NOT NULL,
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_marks NUMERIC(8,2) NOT NULL DEFAULT 100,
    passing_marks NUMERIC(8,2) NOT NULL DEFAULT 40,
    grading_system_id UUID REFERENCES grading_systems(id),
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    result_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_exam_status CHECK (status IN ('draft', 'active', 'published', 'archived')),
    CONSTRAINT chk_exam_result_status CHECK (result_status IN ('draft', 'published')),
    CONSTRAINT chk_exam_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_exams_session_class ON exams (session_id, class_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_exams_status ON exams (status) WHERE deleted_at IS NULL;

CREATE TABLE exam_subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id),
    full_marks NUMERIC(8,2) NOT NULL,
    pass_marks NUMERIC(8,2) NOT NULL,
    written_marks NUMERIC(8,2) NOT NULL DEFAULT 0,
    mcq_marks NUMERIC(8,2) NOT NULL DEFAULT 0,
    practical_marks NUMERIC(8,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_exam_subject_marks CHECK (full_marks >= pass_marks),
    CONSTRAINT chk_exam_subject_components CHECK (written_marks + mcq_marks + practical_marks <= full_marks)
);

CREATE UNIQUE INDEX idx_exam_subjects_unique ON exam_subjects (exam_id, subject_id);

CREATE TABLE student_marks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    exam_subject_id UUID NOT NULL REFERENCES exam_subjects(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    written_score NUMERIC(8,2) NOT NULL DEFAULT 0,
    mcq_score NUMERIC(8,2) NOT NULL DEFAULT 0,
    practical_score NUMERIC(8,2) NOT NULL DEFAULT 0,
    total_score NUMERIC(8,2) NOT NULL DEFAULT 0,
    is_absent BOOLEAN NOT NULL DEFAULT false,
    entered_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_student_marks_unique ON student_marks (exam_subject_id, student_id);
CREATE INDEX idx_student_marks_exam ON student_marks (exam_id);
CREATE INDEX idx_student_marks_student ON student_marks (student_id);

CREATE TABLE exam_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id UUID NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    section_id UUID NOT NULL REFERENCES sections(id),
    total_obtained NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_full NUMERIC(10,2) NOT NULL DEFAULT 0,
    percentage NUMERIC(5,2) NOT NULL DEFAULT 0,
    gpa NUMERIC(3,2) NOT NULL DEFAULT 0,
    cgpa NUMERIC(3,2) NOT NULL DEFAULT 0,
    grade VARCHAR(5),
    is_passed BOOLEAN NOT NULL DEFAULT false,
    class_position INT,
    section_position INT,
    merit_position INT,
    result_status VARCHAR(20) NOT NULL DEFAULT 'draft',
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_result_status CHECK (result_status IN ('draft', 'published'))
);

CREATE UNIQUE INDEX idx_exam_results_unique ON exam_results (exam_id, student_id);
CREATE INDEX idx_exam_results_exam ON exam_results (exam_id);
CREATE INDEX idx_exam_results_passed ON exam_results (exam_id, is_passed);

CREATE TABLE report_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_result_id UUID NOT NULL REFERENCES exam_results(id) ON DELETE CASCADE,
    exam_id UUID NOT NULL REFERENCES exams(id),
    student_id UUID NOT NULL REFERENCES students(id),
    card_token VARCHAR(64) NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_report_cards_token ON report_cards (card_token);
CREATE INDEX idx_report_cards_student ON report_cards (student_id);

-- Default GPA grading system
INSERT INTO grading_systems (id, name, is_default) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'Default GPA System', true);

INSERT INTO grading_scales (system_id, grade, min_percentage, max_percentage, gpa_point, sort_order) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'A+', 80, 100, 5.00, 1),
    ('a0000000-0000-0000-0000-000000000001', 'A',  70, 79.99, 4.00, 2),
    ('a0000000-0000-0000-0000-000000000001', 'A-', 60, 69.99, 3.50, 3),
    ('a0000000-0000-0000-0000-000000000001', 'B',  50, 59.99, 3.00, 4),
    ('a0000000-0000-0000-0000-000000000001', 'C',  40, 49.99, 2.00, 5),
    ('a0000000-0000-0000-0000-000000000001', 'D',  33, 39.99, 1.00, 6),
    ('a0000000-0000-0000-0000-000000000001', 'F',  0,  32.99, 0.00, 7);

-- +goose Down
DROP TABLE IF EXISTS report_cards;
DROP TABLE IF EXISTS exam_results;
DROP TABLE IF EXISTS student_marks;
DROP TABLE IF EXISTS exam_subjects;
DROP TABLE IF EXISTS exams;
DROP TABLE IF EXISTS grading_scales;
DROP TABLE IF EXISTS grading_systems;
