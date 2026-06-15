package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/export"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/payment"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerPaymentRoutes(app, auth fiber.Router, mw *middleware.Middleware) {
	app.Post("/webhooks/payment/:provider", h.PaymentWebhook)
	app.Get("/payments/success", h.PaymentSuccess)
	app.Get("/payments/failed", h.PaymentFailed)

	pub := app.Group("/admission")
	pub.Post("/pay", mw.CSRFProtect(), h.PublicAdmissionPay)

	parent := auth.Group("/parent", mw.RequireParent())
	parent.Get("/students/:id/pay", h.ParentPayNowPage)
	parent.Post("/students/:id/pay", mw.CSRFProtect(), h.ParentInitiatePayment)

	admin := auth.Group("/payments", mw.RequirePermission(model.PermPaymentManage))
	admin.Get("/dashboard", h.PaymentDashboard)
	admin.Get("/gateways", h.PaymentGatewaysPage)
	admin.Post("/gateways/:id", mw.CSRFProtect(), h.PaymentGatewayUpdate)
	admin.Get("/transactions", mw.RequirePermission(model.PermPaymentReportView), h.PaymentTransactionsPage)
	admin.Get("/transactions/export.csv", mw.RequirePermission(model.PermPaymentReportView), h.PaymentTransactionsExport)
	admin.Get("/reports", mw.RequirePermission(model.PermPaymentReportView), h.PaymentReportsPage)
	admin.Get("/refunds", mw.RequirePermission(model.PermPaymentRefund), h.PaymentRefundsPage)
	admin.Post("/refunds", mw.CSRFProtect(), mw.RequirePermission(model.PermPaymentRefund), h.PaymentRefundRequest)
	admin.Post("/refunds/:id/approve", mw.CSRFProtect(), mw.RequirePermission(model.PermPaymentRefund), h.PaymentRefundApprove)
	admin.Post("/refunds/:id/reject", mw.CSRFProtect(), mw.RequirePermission(model.PermPaymentRefund), h.PaymentRefundReject)
}

func (h *Handler) PaymentWebhook(c fiber.Ctx) error {
	slug := c.Params("provider")
	ref := c.FormValue("transaction_ref")
	if ref == "" {
		ref = c.FormValue("ref")
	}
	gwRef := c.FormValue("gateway_ref")
	if gwRef == "" {
		gwRef = c.FormValue("trxID")
	}
	status := c.FormValue("status")
	if status == "" {
		status = "success"
	}
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	payload := dto.WebhookPayload{
		TransactionRef: ref, GatewayRef: gwRef, Status: status, Amount: amount,
		Signature: c.FormValue("signature"), RawBody: map[string]any{"provider": slug},
	}
	if err := h.services.Payment.HandleWebhook(c.Context(), slug, payload, c.IP()); err != nil {
		h.logger.Warn("payment webhook failed", "provider", slug, "ref", ref, "error", err)
		return c.Status(fiber.StatusBadRequest).SendString("invalid")
	}
	return c.SendString("OK")
}

func (h *Handler) PaymentSuccess(c fiber.Ctx) error {
	ref := c.Query("ref")
	if ref != "" {
		sig := payment.SignPayload("", ref, 0)
		_ = h.services.Payment.CompleteTransaction(c.Context(), ref, c.Query("gateway"), true, map[string]any{"sandbox": true, "signature": sig}, c.IP())
	}
	return h.render(c, fiber.StatusOK, web.PaymentResultPage{Success: true, Ref: ref})
}

func (h *Handler) PaymentFailed(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.PaymentResultPage{Success: false, Ref: c.Query("ref")})
}

func (h *Handler) PublicAdmissionPay(c fiber.Ctx) error {
	appNo := c.FormValue("application_number")
	token := c.FormValue("tracking_token")
	app, err := h.services.Admission.Track(c.Context(), appNo, token)
	if err != nil {
		h.flash(c, "Invalid application", true)
		return c.Redirect().To("/admission/track")
	}
	req := dto.InitiatePaymentRequest{
		GatewaySlug: c.FormValue("gateway"), PaymentType: model.GwPaymentAdmission,
		ApplicationID: app.ID, Amount: app.AdmissionFeeAmount,
	}
	resp, err := h.services.Payment.InitiateAdmissionPayment(c.Context(), app.ID, req, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/admission/track")
	}
	return c.Redirect().To(resp.RedirectURL)
}

func (h *Handler) ParentPayNowPage(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid student")
	}
	if err := h.requireStudentAccess(c, studentID); err != nil {
		return err
	}
	data, err := h.services.Payment.ParentPayNowData(c.Context(), studentID)
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.ParentPayNowPage{Data: data})
}

func (h *Handler) ParentInitiatePayment(c fiber.Ctx) error {
	studentID, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid student")
	}
	if err := h.requireStudentAccess(c, studentID); err != nil {
		return err
	}
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	var parentID *uuid.UUID
	if user := h.parentUser(c); user != nil && user.RoleSlug == model.RoleParent {
		if p, _ := h.services.Parent.GetProfile(c.Context(), user.ID); p != nil {
			parentID = &p.ID
		}
	}
	req := dto.InitiatePaymentRequest{
		GatewaySlug: c.FormValue("gateway"), PaymentType: model.GwPaymentStudentFee,
		StudentID: studentID, Amount: amount, IdempotencyKey: c.FormValue("idempotency_key"),
	}
	resp, err := h.services.Payment.InitiateStudentFeePayment(c.Context(), studentID, parentID, req, c.IP())
	if err != nil {
		h.flash(c, err.Error(), true)
		return c.Redirect().To("/parent/students/" + studentID.String() + "/fees")
	}
	return c.Redirect().To(resp.RedirectURL)
}

func (h *Handler) PaymentDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Payment.DashboardStats(c.Context())
	return h.render(c, fiber.StatusOK, web.PaymentDashboardPage{Stats: stats})
}

func (h *Handler) PaymentGatewaysPage(c fiber.Ctx) error {
	gateways, _ := h.services.Payment.ListGateways(c.Context())
	return h.render(c, fiber.StatusOK, web.PaymentGatewaysPage{Gateways: gateways, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) PaymentGatewayUpdate(c fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid gateway")
	}
	req := dto.PaymentGatewayRequest{
		Name: c.FormValue("name"), Slug: c.FormValue("slug"),
		IsActive: c.FormValue("is_active") == "on" || c.FormValue("is_active") == "true",
		IsSandbox: c.FormValue("is_sandbox") == "on" || c.FormValue("is_sandbox") == "true",
		APIKey: c.FormValue("api_key"), APISecret: c.FormValue("api_secret"),
		MerchantID: c.FormValue("merchant_id"), StoreID: c.FormValue("store_id"),
		CallbackURL: c.FormValue("callback_url"), SuccessURL: c.FormValue("success_url"), FailURL: c.FormValue("fail_url"),
	}
	user := middleware.GetUser(c)
	if _, err := h.services.Payment.UpdateGateway(c.Context(), id, req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Gateway updated", false)
	}
	return c.Redirect().To("/payments/gateways")
}

func (h *Handler) parsePaymentReportFilter(c fiber.Ctx) dto.PaymentReportFilter {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	return dto.PaymentReportFilter{
		Query: c.Query("q"), Status: c.Query("status"), GatewaySlug: c.Query("gateway"),
		PaymentType: c.Query("type"), Page: page, PageSize: 25,
		From: parseOptionalDate(c.Query("from")), To: parseOptionalDateEnd(c.Query("to")),
	}
}

func parseOptionalDateEnd(s string) time.Time {
	t, err := parseDate(s)
	if err != nil {
		return time.Time{}
	}
	return t.Add(24*time.Hour - time.Second)
}

func (h *Handler) PaymentTransactionsPage(c fiber.Ctx) error {
	data, _ := h.services.Payment.SearchTransactions(c.Context(), h.parsePaymentReportFilter(c))
	return h.render(c, fiber.StatusOK, web.PaymentTransactionsPage{Data: data, Filter: h.parsePaymentReportFilter(c)})
}

func (h *Handler) PaymentTransactionsExport(c fiber.Ctx) error {
	f := h.parsePaymentReportFilter(c)
	f.Page, f.PageSize = 1, 10000
	data, err := h.services.Payment.SearchTransactions(c.Context(), f)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	csv, err := export.GatewayTransactionsCSV(data.Items)
	if err != nil {
		return c.Status(500).SendString("Export failed")
	}
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", `attachment; filename="payment-transactions.csv"`)
	return c.Send(csv)
}

func (h *Handler) PaymentReportsPage(c fiber.Ctx) error {
	from := parseOptionalDate(c.Query("from"))
	to := parseOptionalDateEnd(c.Query("to"))
	report, _ := h.services.Payment.GatewayCollectionReport(c.Context(), from, to)
	failed, _ := h.services.Payment.SearchTransactions(c.Context(), dto.PaymentReportFilter{Status: model.GwStatusFailed, Page: 1, PageSize: 50})
	return h.render(c, fiber.StatusOK, web.PaymentReportsPage{Report: report, Failed: failed, From: c.Query("from"), To: c.Query("to")})
}

func (h *Handler) PaymentRefundsPage(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Payment.ListRefunds(c.Context(), c.Query("status"), page, 25)
	return h.render(c, fiber.StatusOK, web.PaymentRefundsPage{Data: data, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) PaymentRefundRequest(c fiber.Ctx) error {
	paymentID, err := uuid.Parse(c.FormValue("payment_id"))
	if err != nil {
		h.flash(c, "Invalid payment", true)
		return c.Redirect().To("/payments/refunds")
	}
	amount, _ := strconv.ParseFloat(c.FormValue("amount"), 64)
	req := dto.RefundRequest{PaymentID: paymentID, Amount: amount, Reason: c.FormValue("reason")}
	user := middleware.GetUser(c)
	if _, err := h.services.Payment.RequestRefund(c.Context(), req, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else {
		h.flash(c, "Refund requested", false)
	}
	return c.Redirect().To("/payments/refunds")
}

func (h *Handler) PaymentRefundApprove(c fiber.Ctx) error {
	return h.paymentRefundAction(c, true)
}

func (h *Handler) PaymentRefundReject(c fiber.Ctx) error {
	return h.paymentRefundAction(c, false)
}

func (h *Handler) paymentRefundAction(c fiber.Ctx, approve bool) error {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return c.Status(400).SendString("Invalid refund")
	}
	user := middleware.GetUser(c)
	if err := h.services.Payment.ApproveRefund(c.Context(), id, approve, user.ID, c.IP()); err != nil {
		h.flash(c, err.Error(), true)
	} else if approve {
		h.flash(c, "Refund processed", false)
	} else {
		h.flash(c, "Refund rejected", false)
	}
	return c.Redirect().To("/payments/refunds")
}

func parseOptionalDate(s string) time.Time {
	t, _ := parseDate(s)
	return t
}
