package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/middleware"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/validator"
	"github.com/school-management/pos/internal/web"
)

func (h *Handler) registerWebsiteRoutes(app, auth fiber.Router, mw *middleware.Middleware) {
	site := app.Group("/site")
	site.Get("/", h.PublicSiteHome)
	site.Get("/page/:slug", h.PublicSitePage)
	site.Get("/news", h.PublicNewsList)
	site.Get("/news/:slug", h.PublicNewsDetail)
	site.Get("/events", h.PublicEventsList)
	site.Get("/downloads", h.PublicDownloads)
	site.Get("/downloads/:slug/file", h.PublicDownloadFile)
	site.Get("/contact", h.PublicContactPage)
	site.Post("/contact", mw.CSRFProtect(), h.PublicContactSubmit)
	site.Get("/gallery", h.PublicGallery)

	admin := auth.Group("/website", mw.RequirePermission(model.PermWebsiteManage))
	admin.Get("/dashboard", h.WebsiteDashboard)
	admin.Get("/settings", h.WebsiteSettingsPage)
	admin.Post("/settings", mw.CSRFProtect(), h.WebsiteSettingsSave)
	admin.Get("/pages", h.WebsitePagesList)
	admin.Get("/pages/new", h.WebsitePageForm)
	admin.Get("/pages/:id/edit", h.WebsitePageEdit)
	admin.Post("/pages", mw.CSRFProtect(), h.WebsitePageCreate)
	admin.Post("/pages/:id", mw.CSRFProtect(), h.WebsitePageUpdate)
	admin.Post("/pages/:id/delete", mw.CSRFProtect(), h.WebsitePageDelete)
	admin.Get("/banners", h.WebsiteBannersList)
	admin.Post("/banners", mw.CSRFProtect(), h.WebsiteBannerCreate)
	admin.Post("/banners/:id", mw.CSRFProtect(), h.WebsiteBannerUpdate)
	admin.Get("/gallery", h.WebsiteGalleryAdmin)
	admin.Post("/gallery", mw.CSRFProtect(), h.WebsiteGalleryCreate)
	admin.Get("/news", h.WebsiteNewsAdmin)
	admin.Get("/news/new", h.WebsiteNewsForm)
	admin.Post("/news", mw.CSRFProtect(), h.WebsiteNewsCreate)
	admin.Get("/events", h.WebsiteEventsAdmin)
	admin.Post("/events", mw.CSRFProtect(), h.WebsiteEventCreate)
	admin.Get("/downloads", h.WebsiteDownloadsAdmin)
	admin.Post("/downloads", mw.CSRFProtect(), h.WebsiteDownloadCreate)
	admin.Get("/contacts", h.WebsiteContactsList)
	admin.Post("/contacts/:id/status", mw.CSRFProtect(), h.WebsiteContactUpdateStatus)
}

func (h *Handler) trackVisit(c fiber.Ctx) {
	h.services.Website.RecordVisit(c.Context(), c.Path())
}

func (h *Handler) PublicSiteHome(c fiber.Ctx) error {
	h.trackVisit(c)
	data, _ := h.services.Website.PublicHome(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicSiteHomePage{Data: data})
}

func (h *Handler) PublicSitePage(c fiber.Ctx) error {
	h.trackVisit(c)
	slug := c.Params("slug")
	page, err := h.services.Website.GetPageBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(404).SendString("Page not found")
	}
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicSitePageView{Page: page, Settings: settings})
}

func (h *Handler) PublicNewsList(c fiber.Ctx) error {
	h.trackVisit(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Website.ListNews(c.Context(), page, true)
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicNewsListPage{Data: data, Settings: settings})
}

func (h *Handler) PublicNewsDetail(c fiber.Ctx) error {
	h.trackVisit(c)
	news, err := h.services.Website.GetNewsBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(404).SendString("Not found")
	}
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicNewsDetailPage{News: news, Settings: settings})
}

func (h *Handler) PublicEventsList(c fiber.Ctx) error {
	h.trackVisit(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Website.ListEvents(c.Context(), page, true)
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicEventsListPage{Data: data, Settings: settings})
}

func (h *Handler) PublicDownloads(c fiber.Ctx) error {
	h.trackVisit(c)
	items, _ := h.services.Website.ListDownloads(c.Context(), true)
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicDownloadsPage{Downloads: items, Settings: settings})
}

func (h *Handler) PublicDownloadFile(c fiber.Ctx) error {
	dl, err := h.services.Website.DownloadFile(c.Context(), c.Params("slug"))
	if err != nil || dl.FileURL == "" {
		return c.Status(404).SendString("Not found")
	}
	return c.Redirect().To(dl.FileURL)
}

func (h *Handler) PublicGallery(c fiber.Ctx) error {
	h.trackVisit(c)
	items, _ := h.services.Website.ListGallery(c.Context())
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicGalleryPage{Gallery: items, Settings: settings})
}

func (h *Handler) PublicContactPage(c fiber.Ctx) error {
	h.trackVisit(c)
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.PublicContactPage{Settings: settings, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) PublicContactSubmit(c fiber.Ctx) error {
	req := dto.ContactFormRequest{
		Name: c.FormValue("name"), Email: c.FormValue("email"), Phone: c.FormValue("phone"),
		Subject: c.FormValue("subject"), Message: c.FormValue("message"),
	}
	if errs := h.validate.Validate(req); len(errs) > 0 {
		h.flash(c, validator.FirstError(errs), true)
		return c.Redirect().To("/site/contact")
	}
	if _, err := h.services.Website.SubmitContact(c.Context(), req); err != nil {
		h.flash(c, "Unable to send message", true)
		return c.Redirect().To("/site/contact")
	}
	h.flash(c, "Message sent successfully", false)
	return c.Redirect().To("/site/contact")
}

func (h *Handler) WebsiteDashboard(c fiber.Ctx) error {
	stats, _ := h.services.Website.Dashboard(c.Context())
	return h.render(c, fiber.StatusOK, web.WebsiteDashboardPage{Stats: stats})
}

func (h *Handler) WebsiteSettingsPage(c fiber.Ctx) error {
	settings, _ := h.services.Website.GetSettings(c.Context())
	return h.render(c, fiber.StatusOK, web.WebsiteSettingsPage{Settings: settings, Flash: c.Cookies("flash"), FlashType: c.Cookies("flash_type")})
}

func (h *Handler) WebsiteSettingsSave(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.WebsiteSettingsRequest{
		SiteName: c.FormValue("site_name"), Tagline: c.FormValue("tagline"),
		PrimaryColor: c.FormValue("primary_color"), SecondaryColor: c.FormValue("secondary_color"),
		FacebookURL: c.FormValue("facebook_url"), TwitterURL: c.FormValue("twitter_url"),
		InstagramURL: c.FormValue("instagram_url"), YoutubeURL: c.FormValue("youtube_url"),
		ContactEmail: c.FormValue("contact_email"), ContactPhone: c.FormValue("contact_phone"),
		ContactAddress: c.FormValue("contact_address"),
		DefaultMetaTitle: c.FormValue("default_meta_title"), DefaultMetaDescription: c.FormValue("default_meta_description"),
	}
	logoURL, faviconURL := h.uploadOptional(c, "logo"), h.uploadOptional(c, "favicon")
	if _, err := h.services.Website.SaveSettings(c.Context(), req, logoURL, faviconURL, user.ID, c.IP()); err != nil {
		h.flash(c, "Save failed", true)
	} else {
		h.flash(c, "Settings saved", false)
	}
	return c.Redirect().To("/website/settings")
}

func (h *Handler) uploadOptional(c fiber.Ctx, field string) string {
	file, err := c.FormFile(field)
	if err != nil || file == nil {
		return ""
	}
	f, _ := file.Open()
	if f == nil {
		return ""
	}
	defer f.Close()
	url, _ := h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
	return url
}

func (h *Handler) WebsitePagesList(c fiber.Ctx) error {
	pages, _ := h.services.Website.ListPages(c.Context())
	return h.render(c, fiber.StatusOK, web.WebsitePagesListPage{Pages: pages})
}

func (h *Handler) WebsitePageForm(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.WebsitePageFormPage{Title: "Create Page"})
}

func (h *Handler) WebsitePageEdit(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	pages, _ := h.services.Website.ListPages(c.Context())
	var page *dto.WebsitePageResponse
	for i := range pages {
		if pages[i].ID == id {
			page = &pages[i]
			break
		}
	}
	if page == nil {
		return c.Status(404).SendString("Not found")
	}
	return h.render(c, fiber.StatusOK, web.WebsitePageFormPage{Title: "Edit Page", Page: page})
}

func (h *Handler) parsePageRequest(c fiber.Ctx) dto.WebsitePageRequest {
	return dto.WebsitePageRequest{
		Slug: c.FormValue("slug"), Title: c.FormValue("title"), PageType: c.FormValue("page_type"),
		Content: c.FormValue("content"), MetaTitle: c.FormValue("meta_title"), MetaDescription: c.FormValue("meta_description"),
		IsPublished: c.FormValue("is_published") == "on" || c.FormValue("is_published") == "true",
	}
}

func (h *Handler) WebsitePageCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := h.parsePageRequest(c)
	if _, err := h.services.Website.SavePage(c.Context(), nil, req, user.ID, c.IP()); err != nil {
		h.flash(c, "Create failed", true)
		return c.Redirect().To("/website/pages/new")
	}
	h.flash(c, "Page created", false)
	return c.Redirect().To("/website/pages")
}

func (h *Handler) WebsitePageUpdate(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	req := h.parsePageRequest(c)
	if _, err := h.services.Website.SavePage(c.Context(), &id, req, user.ID, c.IP()); err != nil {
		h.flash(c, "Update failed", true)
	} else {
		h.flash(c, "Page updated", false)
	}
	return c.Redirect().To("/website/pages")
}

func (h *Handler) WebsitePageDelete(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	_ = h.services.Website.DeletePage(c.Context(), id, user.ID, c.IP())
	h.flash(c, "Page deleted", false)
	return c.Redirect().To("/website/pages")
}

func (h *Handler) WebsiteBannersList(c fiber.Ctx) error {
	banners, _ := h.services.Website.ListBanners(c.Context())
	return h.render(c, fiber.StatusOK, web.WebsiteBannersPage{Banners: banners})
}

func (h *Handler) WebsiteBannerCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.WebsiteBannerRequest{
		Title: c.FormValue("title"), Subtitle: c.FormValue("subtitle"), LinkURL: c.FormValue("link_url"),
		IsActive: c.FormValue("is_active") == "on",
	}
	img := h.uploadOptional(c, "image")
	if img == "" {
		h.flash(c, "Image required", true)
		return c.Redirect().To("/website/banners")
	}
	if _, err := h.services.Website.SaveBanner(c.Context(), nil, req, img, user.ID, c.IP()); err != nil {
		h.flash(c, "Create failed", true)
	} else {
		h.flash(c, "Banner created", false)
	}
	return c.Redirect().To("/website/banners")
}

func (h *Handler) WebsiteBannerUpdate(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	req := dto.WebsiteBannerRequest{
		Title: c.FormValue("title"), Subtitle: c.FormValue("subtitle"), LinkURL: c.FormValue("link_url"),
		IsActive: c.FormValue("is_active") == "on",
	}
	img := h.uploadOptional(c, "image")
	_, _ = h.services.Website.SaveBanner(c.Context(), &id, req, img, user.ID, c.IP())
	return c.Redirect().To("/website/banners")
}

func (h *Handler) WebsiteGalleryAdmin(c fiber.Ctx) error {
	items, _ := h.services.Website.ListGallery(c.Context())
	return h.render(c, fiber.StatusOK, web.WebsiteGalleryAdminPage{Gallery: items})
}

func (h *Handler) WebsiteGalleryCreate(c fiber.Ctx) error {
	req := dto.WebsiteGalleryRequest{Title: c.FormValue("title"), Caption: c.FormValue("caption"), IsActive: true}
	img := h.uploadOptional(c, "image")
	if img != "" {
		_, _ = h.services.Website.SaveGallery(c.Context(), nil, req, img)
	}
	return c.Redirect().To("/website/gallery")
}

func (h *Handler) WebsiteNewsAdmin(c fiber.Ctx) error {
	data, _ := h.services.Website.ListNews(c.Context(), 1, false)
	return h.render(c, fiber.StatusOK, web.WebsiteNewsAdminPage{Data: data})
}

func (h *Handler) WebsiteNewsForm(c fiber.Ctx) error {
	return h.render(c, fiber.StatusOK, web.WebsiteNewsFormPage{})
}

func (h *Handler) WebsiteNewsCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.NewsRequest{
		Slug: c.FormValue("slug"), Title: c.FormValue("title"), Excerpt: c.FormValue("excerpt"), Body: c.FormValue("body"),
		IsPublished: c.FormValue("is_published") == "on",
		MetaTitle: c.FormValue("meta_title"), MetaDescription: c.FormValue("meta_description"),
	}
	img := h.uploadOptional(c, "image")
	if _, err := h.services.Website.SaveNews(c.Context(), nil, req, img, user.ID, c.IP()); err != nil {
		h.flash(c, "Create failed", true)
		return c.Redirect().To("/website/news/new")
	}
	return c.Redirect().To("/website/news")
}

func (h *Handler) WebsiteEventsAdmin(c fiber.Ctx) error {
	data, _ := h.services.Website.ListEvents(c.Context(), 1, false)
	return h.render(c, fiber.StatusOK, web.WebsiteEventsAdminPage{Data: data})
}

func (h *Handler) WebsiteEventCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.EventRequest{
		Slug: c.FormValue("slug"), Title: c.FormValue("title"), Description: c.FormValue("description"),
		Location: c.FormValue("location"), IsPublished: c.FormValue("is_published") == "on",
		MetaTitle: c.FormValue("meta_title"), MetaDescription: c.FormValue("meta_description"),
	}
	if t, err := parseDate(c.FormValue("event_date")); err == nil {
		req.EventDate = t
	}
	img := h.uploadOptional(c, "image")
	if _, err := h.services.Website.SaveEvent(c.Context(), nil, req, img, user.ID, c.IP()); err != nil {
		h.flash(c, "Create failed", true)
	} else {
		h.flash(c, "Event created", false)
	}
	return c.Redirect().To("/website/events")
}

func (h *Handler) WebsiteDownloadsAdmin(c fiber.Ctx) error {
	items, _ := h.services.Website.ListDownloads(c.Context(), false)
	return h.render(c, fiber.StatusOK, web.WebsiteDownloadsAdminPage{Downloads: items})
}

func (h *Handler) WebsiteDownloadCreate(c fiber.Ctx) error {
	user := middleware.GetUser(c)
	req := dto.DownloadRequest{
		Slug: c.FormValue("slug"), Title: c.FormValue("title"), Description: c.FormValue("description"),
		Category: c.FormValue("category"), IsPublished: c.FormValue("is_published") == "on",
	}
	file, _ := c.FormFile("file")
	var fileURL, fileName string
	if file != nil {
		f, _ := file.Open()
		if f != nil {
			fileURL, _ = h.storage.Upload(c.Context(), file.Filename, f, file.Header.Get("Content-Type"))
			fileName = file.Filename
			f.Close()
		}
	}
	if _, err := h.services.Website.SaveDownload(c.Context(), nil, req, fileURL, fileName, user.ID, c.IP()); err != nil {
		h.flash(c, "Upload failed", true)
	} else {
		h.flash(c, "Download added", false)
	}
	return c.Redirect().To("/website/downloads")
}

func (h *Handler) WebsiteContactsList(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	data, _ := h.services.Website.ListContacts(c.Context(), c.Query("status"), page)
	return h.render(c, fiber.StatusOK, web.WebsiteContactsPage{Data: data})
}

func (h *Handler) WebsiteContactUpdateStatus(c fiber.Ctx) error {
	id, _ := parseUUIDParam(c, "id")
	user := middleware.GetUser(c)
	status := c.FormValue("status")
	_ = h.services.Website.UpdateContactStatus(c.Context(), id, status, user.ID, c.IP())
	return c.Redirect().To("/website/contacts")
}
