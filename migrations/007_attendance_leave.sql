-- +goose Up
CREATE TABLE student_attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES academic_sessions(id),
    class_id UUID NOT NULL REFERENCES classes(id),
    section_id UUID NOT NULL REFERENCES sections(id),
    attendance_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL,
    marked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_student_attendance_status CHECK (status IN ('present', 'absent', 'late', 'leave'))
);

CREATE UNIQUE INDEX idx_student_attendance_unique ON student_attendance (student_id, attendance_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_student_attendance_date ON student_attendance (attendance_date);
CREATE INDEX idx_student_attendance_class_section ON student_attendance (class_id, section_id, attendance_date);
CREATE INDEX idx_student_attendance_session ON student_attendance (session_id, attendance_date);

CREATE TABLE teacher_attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    attendance_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL,
    marked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_teacher_attendance_status CHECK (status IN ('present', 'absent', 'late', 'leave'))
);

CREATE UNIQUE INDEX idx_teacher_attendance_unique ON teacher_attendance (teacher_id, attendance_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_teacher_attendance_date ON teacher_attendance (attendance_date);

CREATE TABLE staff_attendance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL REFERENCES staffs(id) ON DELETE CASCADE,
    attendance_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL,
    marked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_staff_attendance_status CHECK (status IN ('present', 'absent', 'late', 'leave'))
);

CREATE UNIQUE INDEX idx_staff_attendance_unique ON staff_attendance (staff_id, attendance_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_staff_attendance_date ON staff_attendance (attendance_date);

CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(20) NOT NULL,
    teacher_id UUID REFERENCES teachers(id) ON DELETE CASCADE,
    staff_id UUID REFERENCES staffs(id) ON DELETE CASCADE,
    leave_type VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    applied_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    review_remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_leave_entity CHECK (
        (entity_type = 'teacher' AND teacher_id IS NOT NULL AND staff_id IS NULL) OR
        (entity_type = 'staff' AND staff_id IS NOT NULL AND teacher_id IS NULL)
    ),
    CONSTRAINT chk_leave_type CHECK (leave_type IN ('casual', 'sick', 'annual', 'emergency')),
    CONSTRAINT chk_leave_status CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_leave_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_leave_requests_status ON leave_requests (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_teacher ON leave_requests (teacher_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_leave_requests_staff ON leave_requests (staff_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS staff_attendance;
DROP TABLE IF EXISTS teacher_attendance;
DROP TABLE IF EXISTS student_attendance;
