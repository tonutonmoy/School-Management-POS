package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/notify"
	"github.com/school-management/pos/internal/repository"
)

type NotificationService struct {
	repos *repository.Repositories
	sms   notify.SMSProvider
	email notify.EmailProvider
	audit *AuditService
}

func NewNotificationService(repos *repository.Repositories, sms notify.SMSProvider, email notify.EmailProvider, audit *AuditService) *NotificationService {
	return &NotificationService{repos: repos, sms: sms, email: email, audit: audit}
}

func (s *NotificationService) ListForParent(ctx context.Context, parentID uuid.UUID, f dto.NotificationFilter) (*dto.PaginatedNotifications, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
	total, err := s.repos.Notifications.CountForParent(ctx, parentID, f.UnreadOnly)
	if err != nil {
		return nil, err
	}
	recs, err := s.repos.Notifications.ListForParent(ctx, parentID, f.UnreadOnly, int32(f.PageSize), int32((f.Page-1)*f.PageSize))
	if err != nil {
		return nil, err
	}
	items := make([]dto.NotificationResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, mapNotification(&r))
	}
	totalPages := int(total) / f.PageSize
	if int(total)%f.PageSize > 0 {
		totalPages++
	}
	return &dto.PaginatedNotifications{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize, TotalPages: totalPages}, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, id, parentID uuid.UUID) error {
	return s.repos.Notifications.MarkRead(ctx, id, parentID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, parentID uuid.UUID) error {
	return s.repos.Notifications.MarkAllRead(ctx, parentID)
}

func (s *NotificationService) CommunicationDashboard(ctx context.Context) (*dto.CommunicationDashboardStats, error) {
	stats, err := s.repos.Notifications.CommunicationStats(ctx)
	if err != nil {
		return nil, err
	}
	sms, _ := s.repos.Notifications.ListRecentSMS(ctx, 10)
	emails, _ := s.repos.Notifications.ListRecentEmails(ctx, 10)
	return &dto.CommunicationDashboardStats{
		SMSSentToday: stats.SMSToday, EmailsSentToday: stats.EmailsToday,
		NotificationCount: stats.Notifications, DeliverySuccessRate: stats.DeliveryRate,
		RecentSMS: mapSMSLogs(sms), RecentEmails: mapEmailLogs(emails),
	}, nil
}

func (s *NotificationService) SendSMS(ctx context.Context, phone, message, eventType string, parentID *uuid.UUID, refType string, refID *uuid.UUID) error {
	logRec, err := s.repos.Notifications.CreateSMSLog(ctx, repository.SMSLogParams{
		RecipientPhone: phone, Message: message, EventType: eventType,
		Provider: s.sms.Name(), Status: model.DeliveryPending,
		ReferenceType: refType, ReferenceID: refID, ParentID: parentID,
	})
	if err != nil {
		return err
	}
	_ = s.repos.Notifications.Enqueue(ctx, "sms", map[string]any{
		"log_id": logRec.ID.String(), "phone": phone, "message": message,
	})
	sendErr := s.sms.Send(ctx, phone, message)
	now := time.Now()
	status := model.DeliverySent
	errMsg := ""
	if sendErr != nil {
		status = model.DeliveryFailed
		errMsg = sendErr.Error()
	}
	return s.repos.Notifications.UpdateSMSStatus(ctx, logRec.ID, status, &now, errMsg)
}

func (s *NotificationService) SendEmail(ctx context.Context, to, subject, body, eventType string, parentID *uuid.UUID, refType string, refID *uuid.UUID) error {
	logRec, err := s.repos.Notifications.CreateEmailLog(ctx, repository.EmailLogParams{
		RecipientEmail: to, Subject: subject, Body: body, EventType: eventType,
		Provider: s.email.Name(), Status: model.DeliveryPending,
		ReferenceType: refType, ReferenceID: refID, ParentID: parentID,
	})
	if err != nil {
		return err
	}
	_ = s.repos.Notifications.Enqueue(ctx, "email", map[string]any{
		"log_id": logRec.ID.String(), "to": to, "subject": subject,
	})
	sendErr := s.email.Send(ctx, to, subject, body)
	now := time.Now()
	status := model.DeliverySent
	errMsg := ""
	if sendErr != nil {
		status = model.DeliveryFailed
		errMsg = sendErr.Error()
	}
	return s.repos.Notifications.UpdateEmailStatus(ctx, logRec.ID, status, &now, errMsg)
}

func (s *NotificationService) createInApp(ctx context.Context, parentID uuid.UUID, title, body, category, refType string, refID *uuid.UUID) {
	pid := parentID
	_, _ = s.repos.Notifications.CreateInApp(ctx, repository.InAppNotificationParams{
		ParentID: &pid, Title: title, Body: body, Category: category,
		ReferenceType: refType, ReferenceID: refID,
	})
}

func (s *NotificationService) notifyParentsOfStudent(ctx context.Context, studentID uuid.UUID, title, body, category, smsEvent, emailEvent, refType string, refID *uuid.UUID) {
	parents, err := s.repos.Parents.ListParentsForStudent(ctx, studentID)
	if err != nil {
		return
	}
	for _, p := range parents {
		s.createInApp(ctx, p.ID, title, body, category, refType, refID)
		user, _ := s.repos.Users.GetByID(ctx, p.UserID)
		if user == nil {
			continue
		}
		pid := p.ID
		phone := p.Phone
		if phone == "" {
			phone = user.Phone
		}
		if phone != "" {
			_ = s.SendSMS(ctx, phone, body, smsEvent, &pid, refType, refID)
		}
		if user.Email != "" {
			_ = s.SendEmail(ctx, user.Email, title, body, emailEvent, &pid, refType, refID)
		}
	}
}

func (s *NotificationService) OnAbsent(ctx context.Context, studentID uuid.UUID, date time.Time) {
	st, err := s.repos.Students.GetByID(ctx, studentID)
	if err != nil || st == nil {
		return
	}
	title := "Attendance Alert"
	body := fmt.Sprintf("%s %s was marked absent on %s.", st.FirstName, st.LastName, date.Format("2006-01-02"))
	sid := studentID
	s.notifyParentsOfStudent(ctx, studentID, title, body, model.NotifyCategoryAttendance, model.SMSEventAbsent, "", model.EntityStudent, &sid)
}

func (s *NotificationService) OnPaymentReceived(ctx context.Context, studentID uuid.UUID, paymentID uuid.UUID, amount float64, receiptNo string) {
	title := "Payment Received"
	body := fmt.Sprintf("Payment of ৳%.2f received. Receipt: %s", amount, receiptNo)
	pid := paymentID
	s.notifyParentsOfStudent(ctx, studentID, title, body, model.NotifyCategoryPayment, model.SMSEventPayment, model.EmailEventReceipt, model.EntityPayment, &pid)
}

func (s *NotificationService) OnResultPublished(ctx context.Context, examID, studentID uuid.UUID, examName string, gpa float64) {
	title := "Result Published"
	body := fmt.Sprintf("Results for %s published. GPA: %.2f", examName, gpa)
	eid := examID
	s.notifyParentsOfStudent(ctx, studentID, title, body, model.NotifyCategoryResult, model.SMSEventResult, model.EmailEventResult, model.EntityExamResult, &eid)
}

func (s *NotificationService) OnNewNotice(ctx context.Context, noticeID uuid.UUID, title, body string) {
	nid := noticeID
	parents, _ := s.repos.Parents.List(ctx, 10000, 0)
	for _, p := range parents {
		s.createInApp(ctx, p.ID, title, body, model.NotifyCategoryNotice, model.EntityNotice, &nid)
		user, _ := s.repos.Users.GetByID(ctx, p.UserID)
		if user == nil {
			continue
		}
		pid := p.ID
		phone := p.Phone
		if phone == "" {
			phone = user.Phone
		}
		if phone != "" {
			_ = s.SendSMS(ctx, phone, title+": "+body, model.SMSEventNotice, &pid, model.EntityNotice, &nid)
		}
		if user.Email != "" {
			_ = s.SendEmail(ctx, user.Email, title, body, model.EmailEventNotice, &pid, model.EntityNotice, &nid)
		}
	}
}

func (s *NotificationService) OnPaymentFailed(ctx context.Context, studentID uuid.UUID, amount float64, ref string) {
	title := "Payment Failed"
	body := fmt.Sprintf("Online payment of ৳%.2f failed. Reference: %s", amount, ref)
	sid := studentID
	s.notifyParentsOfStudent(ctx, studentID, title, body, model.NotifyCategoryPayment, model.SMSEventPaymentFailed, model.EmailEventPaymentFailed, model.EntityGatewayTransaction, &sid)
}

func (s *NotificationService) OnRefundProcessed(ctx context.Context, studentID uuid.UUID, amount float64, refundID uuid.UUID) {
	title := "Refund Processed"
	body := fmt.Sprintf("Refund of ৳%.2f has been processed.", amount)
	rid := refundID
	s.notifyParentsOfStudent(ctx, studentID, title, body, model.NotifyCategoryPayment, model.SMSEventRefund, model.EmailEventRefund, model.EntityPaymentRefund, &rid)
}

func (s *NotificationService) OnAdmissionPaymentReceived(ctx context.Context, email, appNo string, amount float64, receipt string) {
	subject := "Admission Fee Payment Received"
	body := fmt.Sprintf("Payment of ৳%.2f received for application %s. Receipt: %s", amount, appNo, receipt)
	_ = s.SendEmail(ctx, email, subject, body, model.EmailEventReceipt, nil, model.EntityAdmissionApplication, nil)
}
func (s *NotificationService) OnPasswordReset(ctx context.Context, email, resetURL string) {
	subject := "Password Reset"
	body := fmt.Sprintf("Reset your password: %s", resetURL)
	_ = s.SendEmail(ctx, email, subject, body, model.EmailEventPasswordReset, nil, model.EntityUser, nil)
}

func mapNotification(r *repository.NotificationRecord) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID: r.ID, Title: r.Title, Body: r.Body, Category: r.Category,
		ReferenceType: r.ReferenceType, ReferenceID: r.ReferenceID,
		IsRead: r.IsRead, ReadAt: r.ReadAt, CreatedAt: r.CreatedAt,
	}
}

func mapSMSLogs(recs []repository.SMSLogRecord) []dto.SMSLogResponse {
	items := make([]dto.SMSLogResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.SMSLogResponse{
			ID: r.ID, RecipientPhone: r.RecipientPhone, Message: r.Message,
			EventType: r.EventType, Provider: r.Provider, Status: r.Status,
			SentAt: r.SentAt, CreatedAt: r.CreatedAt,
		})
	}
	return items
}

func mapEmailLogs(recs []repository.EmailLogRecord) []dto.EmailLogResponse {
	items := make([]dto.EmailLogResponse, 0, len(recs))
	for _, r := range recs {
		items = append(items, dto.EmailLogResponse{
			ID: r.ID, RecipientEmail: r.RecipientEmail, Subject: r.Subject,
			EventType: r.EventType, Provider: r.Provider, Status: r.Status,
			SentAt: r.SentAt, CreatedAt: r.CreatedAt,
		})
	}
	return items
}

func (s *NotificationService) ExportSMSLogs(ctx context.Context, from, to time.Time) ([]repository.SMSLogRecord, error) {
	return s.repos.Notifications.ListSMSForExport(ctx, from, to)
}
