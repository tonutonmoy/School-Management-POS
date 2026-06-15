package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/school-management/pos/internal/dto"
)

func TeachersCSV(teachers []dto.TeacherResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Employee ID", "Name", "Email", "Phone", "Department", "Designation", "Employment", "Status", "Joining Date"})
	for _, t := range teachers {
		_ = w.Write([]string{
			t.EmployeeID, t.FullName, t.Email, t.Phone, t.DepartmentName, t.DesignationName,
			t.EmploymentType, t.Status, t.JoiningDate.Format("2006-01-02"),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func StaffCSV(staff []dto.StaffResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Employee ID", "Name", "Email", "Phone", "Department", "Designation", "Status", "Joining Date"})
	for _, s := range staff {
		_ = w.Write([]string{
			s.EmployeeID, s.FullName, s.Email, s.Phone, s.DepartmentName, s.DesignationName,
			s.Status, s.JoiningDate.Format("2006-01-02"),
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func TeachersExcel(teachers []dto.TeacherResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Teachers"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Employee ID", "Name", "Email", "Phone", "Department", "Designation", "Employment", "Status", "Joining Date"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, t := range teachers {
		vals := []any{t.EmployeeID, t.FullName, t.Email, t.Phone, t.DepartmentName, t.DesignationName, t.EmploymentType, t.Status, t.JoiningDate.Format("2006-01-02")}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func StaffExcel(staff []dto.StaffResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Staff"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Employee ID", "Name", "Email", "Phone", "Department", "Designation", "Status", "Joining Date"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, s := range staff {
		vals := []any{s.EmployeeID, s.FullName, s.Email, s.Phone, s.DepartmentName, s.DesignationName, s.Status, s.JoiningDate.Format("2006-01-02")}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func AssignmentReportCSV(teachers []dto.TeacherResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Teacher", "Employee ID", "Subject", "Class", "Section"})
	for _, t := range teachers {
		if len(t.Assignments) == 0 {
			_ = w.Write([]string{t.FullName, t.EmployeeID, "", "", ""})
			continue
		}
		for _, a := range t.Assignments {
			_ = w.Write([]string{t.FullName, t.EmployeeID, a.SubjectName, a.ClassName, a.SectionName})
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func DepartmentReportText(dept string, teachers []dto.TeacherResponse, staff []dto.StaffResponse) string {
	return fmt.Sprintf("Department: %s\nTeachers: %d\nStaff: %d\n", dept, len(teachers), len(staff))
}

func StudentAttendanceCSV(records []dto.StudentAttendanceResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Date", "Student", "Admission #", "Roll #", "Class", "Section", "Status", "Remarks"})
	for _, r := range records {
		_ = w.Write([]string{
			r.AttendanceDate.Format("2006-01-02"), r.StudentName, r.AdmissionNumber, r.RollNumber,
			r.ClassName, r.SectionName, r.Status, r.Remarks,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func EmployeeAttendanceCSV(records []dto.StudentAttendanceResponse) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"Date", "Employee", "ID", "Department", "Status", "Remarks"})
	for _, r := range records {
		_ = w.Write([]string{
			r.AttendanceDate.Format("2006-01-02"), r.StudentName, r.AdmissionNumber,
			r.ClassName, r.Status, r.Remarks,
		})
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func StudentAttendanceExcel(records []dto.StudentAttendanceResponse) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Attendance"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"Date", "Student", "Admission #", "Roll #", "Class", "Section", "Status", "Remarks"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for row, r := range records {
		vals := []any{r.AttendanceDate.Format("2006-01-02"), r.StudentName, r.AdmissionNumber, r.RollNumber, r.ClassName, r.SectionName, r.Status, r.Remarks}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
