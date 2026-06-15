package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/school-management/pos/internal/dto"
	"github.com/school-management/pos/internal/model"
	"github.com/school-management/pos/internal/repository"
)

type WebsiteService struct {
	repos *repository.Repositories
	audit *AuditService
}

func NewWebsiteService(repos *repository.Repositories, audit *AuditService) *WebsiteService {
	return &WebsiteService{repos: repos, audit: audit}
}

func (s *WebsiteService) RecordVisit(ctx context.Context, path string) {
	_ = s.repos.Website.RecordVisit(ctx, path)
}

func (s *WebsiteService) Dashboard(ctx context.Context) (*dto.WebsiteDashboardStats, error) {
	st, err := s.repos.Website.DashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.WebsiteDashboardStats{
		AdmissionApplications: st.AdmissionApps, WebsiteVisitors: st.VisitorsToday,
		ContactInquiries: st.ContactTotal, NewContacts: st.ContactNew, PendingAdmissions: st.PendingAdmissions,
	}, nil
}

func (s *WebsiteService) GetSettings(ctx context.Context) (*dto.WebsiteSettingsResponse, error) {
	rec, err := s.repos.Website.GetSettings(ctx)
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapSettings(rec)
	return &resp, nil
}

func (s *WebsiteService) SaveSettings(ctx context.Context, req dto.WebsiteSettingsRequest, logoURL, faviconURL string, actorID uuid.UUID, ip string) (*dto.WebsiteSettingsResponse, error) {
	rec, err := s.repos.Website.UpdateSettings(ctx, repository.WebsiteSettingsParams{
		SiteName: req.SiteName, Tagline: req.Tagline, PrimaryColor: req.PrimaryColor, SecondaryColor: req.SecondaryColor,
		FacebookURL: req.FacebookURL, TwitterURL: req.TwitterURL, InstagramURL: req.InstagramURL, YoutubeURL: req.YoutubeURL,
		ContactEmail: req.ContactEmail, ContactPhone: req.ContactPhone, ContactAddress: req.ContactAddress,
		DefaultMetaTitle: req.DefaultMetaTitle, DefaultMetaDescription: req.DefaultMetaDescription,
	})
	if err != nil {
		return nil, err
	}
	if logoURL != "" || faviconURL != "" {
		_ = s.repos.Website.UpdateSettingsMedia(ctx, logoURL, faviconURL)
		rec, _ = s.repos.Website.GetSettings(ctx)
	}
	resp := mapSettings(rec)
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityWebsitePage, nil, ip, map[string]any{"type": "settings"})
	return &resp, nil
}

func (s *WebsiteService) PublicHome(ctx context.Context) (*dto.PublicSiteData, error) {
	settings, _ := s.GetSettings(ctx)
	banners, _ := s.repos.Website.ListBanners(ctx, true)
	news, _ := s.repos.Website.SearchNews(ctx, true, 5, 0)
	events, _ := s.repos.Website.SearchEvents(ctx, true, 5, 0)
	data := &dto.PublicSiteData{}
	if settings != nil {
		data.Settings = settings
	}
	for _, b := range banners {
		data.Banners = append(data.Banners, mapBanner(&b))
	}
	for _, n := range news {
		data.News = append(data.News, mapNews(&n))
	}
	for _, e := range events {
		data.Events = append(data.Events, mapEvent(&e))
	}
	return data, nil
}

func (s *WebsiteService) GetPageBySlug(ctx context.Context, slug string) (*dto.WebsitePageResponse, error) {
	rec, err := s.repos.Website.GetPageBySlug(ctx, slug)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	blocks, _ := s.repos.Website.ListBlocks(ctx, &rec.ID)
	resp := mapPage(rec)
	for _, b := range blocks {
		if b.IsActive {
			resp.Blocks = append(resp.Blocks, mapBlock(&b))
		}
	}
	return &resp, nil
}

func (s *WebsiteService) ListPages(ctx context.Context) ([]dto.WebsitePageResponse, error) {
	recs, err := s.repos.Website.ListPages(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]dto.WebsitePageResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapPage(&recs[i]))
	}
	return items, nil
}

func (s *WebsiteService) SavePage(ctx context.Context, id *uuid.UUID, req dto.WebsitePageRequest, actorID uuid.UUID, ip string) (*dto.WebsitePageResponse, error) {
	req.Slug = slugify(req.Slug)
	params := repository.WebsitePageParams{
		Slug: req.Slug, Title: req.Title, PageType: req.PageType, Content: req.Content,
		MetaTitle: req.MetaTitle, MetaDescription: req.MetaDescription, IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	}
	var rec *repository.WebsitePageRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreatePage(ctx, params)
		if err != nil || rec == nil {
			return nil, err
		}
		s.audit.Log(ctx, &actorID, model.ActionCreate, model.EntityWebsitePage, &rec.ID, ip, map[string]any{"slug": req.Slug})
	} else {
		rec, err = s.repos.Website.UpdatePage(ctx, *id, params)
		s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityWebsitePage, id, ip, map[string]any{"slug": req.Slug})
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapPage(rec)
	return &resp, nil
}

func (s *WebsiteService) DeletePage(ctx context.Context, id uuid.UUID, actorID uuid.UUID, ip string) error {
	if err := s.repos.Website.SoftDeletePage(ctx, id); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionDelete, model.EntityWebsitePage, &id, ip, nil)
	return nil
}

func (s *WebsiteService) ListBanners(ctx context.Context) ([]dto.WebsiteBannerResponse, error) {
	recs, _ := s.repos.Website.ListBanners(ctx, false)
	items := make([]dto.WebsiteBannerResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapBanner(&recs[i]))
	}
	return items, nil
}

func (s *WebsiteService) SaveBanner(ctx context.Context, id *uuid.UUID, req dto.WebsiteBannerRequest, imageURL string, actorID uuid.UUID, ip string) (*dto.WebsiteBannerResponse, error) {
	params := repository.WebsiteBannerParams{
		Title: req.Title, Subtitle: req.Subtitle, ImageURL: imageURL, LinkURL: req.LinkURL,
		SortOrder: req.SortOrder, IsActive: req.IsActive,
	}
	var rec *repository.WebsiteBannerRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreateBanner(ctx, params)
	} else {
		if imageURL == "" {
			banners, _ := s.repos.Website.ListBanners(ctx, false)
			for _, b := range banners {
				if b.ID == *id {
					params.ImageURL = b.ImageURL
					break
				}
			}
		}
		rec, err = s.repos.Website.UpdateBanner(ctx, *id, params)
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapBanner(rec)
	return &resp, nil
}

func (s *WebsiteService) ListGallery(ctx context.Context) ([]dto.WebsiteGalleryResponse, error) {
	recs, _ := s.repos.Website.ListGallery(ctx, false)
	items := make([]dto.WebsiteGalleryResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapGallery(&recs[i]))
	}
	return items, nil
}

func (s *WebsiteService) SaveGallery(ctx context.Context, id *uuid.UUID, req dto.WebsiteGalleryRequest, imageURL string) (*dto.WebsiteGalleryResponse, error) {
	params := repository.WebsiteGalleryParams{Title: req.Title, Caption: req.Caption, ImageURL: imageURL, SortOrder: req.SortOrder, IsActive: req.IsActive}
	var rec *repository.WebsiteGalleryRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreateGallery(ctx, params)
	} else {
		if imageURL == "" {
			items, _ := s.repos.Website.ListGallery(ctx, false)
			for _, g := range items {
				if g.ID == *id {
					params.ImageURL = g.ImageURL
					break
				}
			}
		}
		rec, err = s.repos.Website.UpdateGallery(ctx, *id, params)
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapGallery(rec)
	return &resp, nil
}

func (s *WebsiteService) ListNews(ctx context.Context, page int, publishedOnly bool) (*dto.PaginatedNews, error) {
	if page < 1 {
		page = 1
	}
	limit := int32(20)
	total, _ := s.repos.Website.CountNews(ctx, publishedOnly)
	recs, err := s.repos.Website.SearchNews(ctx, publishedOnly, limit, int32((page-1)*int(limit)))
	if err != nil {
		return nil, err
	}
	items := make([]dto.NewsResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapNews(&recs[i]))
	}
	return &dto.PaginatedNews{Items: items, Total: total, Page: page, PageSize: int(limit)}, nil
}

func (s *WebsiteService) GetNewsBySlug(ctx context.Context, slug string) (*dto.NewsResponse, error) {
	rec, err := s.repos.Website.GetNewsBySlug(ctx, slug)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	resp := mapNews(rec)
	return &resp, nil
}

func (s *WebsiteService) SaveNews(ctx context.Context, id *uuid.UUID, req dto.NewsRequest, imageURL string, actorID uuid.UUID, ip string) (*dto.NewsResponse, error) {
	req.Slug = slugify(req.Slug)
	if req.IsPublished && req.PublishedAt == nil {
		now := time.Now()
		req.PublishedAt = &now
	}
	params := repository.NewsParams{
		Slug: req.Slug, Title: req.Title, Excerpt: req.Excerpt, Body: req.Body, ImageURL: imageURL,
		IsPublished: req.IsPublished, PublishedAt: req.PublishedAt, MetaTitle: req.MetaTitle, MetaDescription: req.MetaDescription,
	}
	var rec *repository.NewsRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreateNews(ctx, params)
	} else {
		if imageURL == "" {
			if existing, _ := s.repos.Website.GetNewsByID(ctx, *id); existing != nil {
				params.ImageURL = existing.ImageURL
			}
		}
		rec, err = s.repos.Website.UpdateNews(ctx, *id, params)
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapNews(rec)
	return &resp, nil
}

func (s *WebsiteService) ListEvents(ctx context.Context, page int, publishedOnly bool) (*dto.PaginatedEvents, error) {
	if page < 1 {
		page = 1
	}
	limit := int32(20)
	total, _ := s.repos.Website.CountEvents(ctx, publishedOnly)
	recs, err := s.repos.Website.SearchEvents(ctx, publishedOnly, limit, int32((page-1)*int(limit)))
	if err != nil {
		return nil, err
	}
	items := make([]dto.EventResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapEvent(&recs[i]))
	}
	return &dto.PaginatedEvents{Items: items, Total: total, Page: page, PageSize: int(limit)}, nil
}

func (s *WebsiteService) SaveEvent(ctx context.Context, id *uuid.UUID, req dto.EventRequest, imageURL string, actorID uuid.UUID, ip string) (*dto.EventResponse, error) {
	req.Slug = slugify(req.Slug)
	params := repository.EventParams{
		Slug: req.Slug, Title: req.Title, Description: req.Description, Location: req.Location,
		EventDate: req.EventDate, EndDate: req.EndDate, ImageURL: imageURL,
		IsPublished: req.IsPublished, MetaTitle: req.MetaTitle, MetaDescription: req.MetaDescription,
	}
	var rec *repository.EventRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreateEvent(ctx, params)
	} else {
		if imageURL == "" {
			if existing, _ := s.repos.Website.GetEventByID(ctx, *id); existing != nil {
				params.ImageURL = existing.ImageURL
			}
		}
		rec, err = s.repos.Website.UpdateEvent(ctx, *id, params)
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapEvent(rec)
	return &resp, nil
}

func (s *WebsiteService) ListDownloads(ctx context.Context, publishedOnly bool) ([]dto.DownloadResponse, error) {
	recs, err := s.repos.Website.ListDownloads(ctx, publishedOnly)
	if err != nil {
		return nil, err
	}
	items := make([]dto.DownloadResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapDownload(&recs[i]))
	}
	return items, nil
}

func (s *WebsiteService) SaveDownload(ctx context.Context, id *uuid.UUID, req dto.DownloadRequest, fileURL, fileName string, actorID uuid.UUID, ip string) (*dto.DownloadResponse, error) {
	req.Slug = slugify(req.Slug)
	params := repository.DownloadParams{
		Slug: req.Slug, Title: req.Title, Description: req.Description, Category: req.Category,
		FileURL: fileURL, FileName: fileName, IsPublished: req.IsPublished,
	}
	var rec *repository.DownloadRecord
	var err error
	if id == nil {
		rec, err = s.repos.Website.CreateDownload(ctx, params)
	} else {
		if fileURL == "" {
			if existing, _ := s.repos.Website.GetDownloadByID(ctx, *id); existing != nil {
				params.FileURL = existing.FileURL
				params.FileName = existing.FileName
			}
		}
		rec, err = s.repos.Website.UpdateDownload(ctx, *id, params)
	}
	if err != nil || rec == nil {
		return nil, err
	}
	resp := mapDownload(rec)
	return &resp, nil
}

func (s *WebsiteService) DownloadFile(ctx context.Context, slug string) (*dto.DownloadResponse, error) {
	rec, err := s.repos.Website.GetDownloadBySlug(ctx, slug)
	if err != nil || rec == nil {
		return nil, ErrNotFound
	}
	_ = s.repos.Website.IncrementDownload(ctx, rec.ID)
	resp := mapDownload(rec)
	return &resp, nil
}

func (s *WebsiteService) SubmitContact(ctx context.Context, req dto.ContactFormRequest) (*dto.ContactMessageResponse, error) {
	rec, err := s.repos.Website.CreateContact(ctx, repository.ContactParams{
		Name: req.Name, Email: req.Email, Phone: req.Phone, Subject: req.Subject, Message: req.Message,
	})
	if err != nil {
		return nil, err
	}
	resp := mapContact(rec)
	return &resp, nil
}

func (s *WebsiteService) ListContacts(ctx context.Context, status string, page int) (*dto.PaginatedContactMessages, error) {
	if page < 1 {
		page = 1
	}
	limit := int32(20)
	total, _ := s.repos.Website.CountContacts(ctx, status)
	recs, err := s.repos.Website.SearchContacts(ctx, status, limit, int32((page-1)*int(limit)))
	if err != nil {
		return nil, err
	}
	items := make([]dto.ContactMessageResponse, 0, len(recs))
	for i := range recs {
		items = append(items, mapContact(&recs[i]))
	}
	totalPages := int(total) / int(limit)
	if int(total)%int(limit) > 0 {
		totalPages++
	}
	return &dto.PaginatedContactMessages{Items: items, Total: total, Page: page, PageSize: int(limit), TotalPages: totalPages}, nil
}

func (s *WebsiteService) UpdateContactStatus(ctx context.Context, id uuid.UUID, status string, actorID uuid.UUID, ip string) error {
	if err := s.repos.Website.UpdateContactStatus(ctx, id, status); err != nil {
		return err
	}
	s.audit.Log(ctx, &actorID, model.ActionUpdate, model.EntityContactMessage, &id, ip, map[string]any{"status": status})
	return nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func mapSettings(r *repository.WebsiteSettingsRecord) dto.WebsiteSettingsResponse {
	return dto.WebsiteSettingsResponse{
		ID: r.ID, SiteName: r.SiteName, Tagline: r.Tagline, LogoURL: r.LogoURL, FaviconURL: r.FaviconURL,
		PrimaryColor: r.PrimaryColor, SecondaryColor: r.SecondaryColor,
		FacebookURL: r.FacebookURL, TwitterURL: r.TwitterURL, InstagramURL: r.InstagramURL, YoutubeURL: r.YoutubeURL,
		ContactEmail: r.ContactEmail, ContactPhone: r.ContactPhone, ContactAddress: r.ContactAddress,
		DefaultMetaTitle: r.DefaultMetaTitle, DefaultMetaDescription: r.DefaultMetaDescription,
	}
}

func mapPage(r *repository.WebsitePageRecord) dto.WebsitePageResponse {
	return dto.WebsitePageResponse{
		ID: r.ID, Slug: r.Slug, Title: r.Title, PageType: r.PageType, Content: r.Content,
		MetaTitle: r.MetaTitle, MetaDescription: r.MetaDescription, OGImage: r.OGImage,
		IsPublished: r.IsPublished, SortOrder: r.SortOrder,
	}
}

func mapBlock(r *repository.WebsiteBlockRecord) dto.WebsiteBlockResponse {
	return dto.WebsiteBlockResponse{
		ID: r.ID, PageID: r.PageID, BlockType: r.BlockType, Title: r.Title, Content: r.Content,
		ImageURL: r.ImageURL, SortOrder: r.SortOrder, IsActive: r.IsActive,
	}
}

func mapBanner(r *repository.WebsiteBannerRecord) dto.WebsiteBannerResponse {
	return dto.WebsiteBannerResponse{
		ID: r.ID, Title: r.Title, Subtitle: r.Subtitle, ImageURL: r.ImageURL, LinkURL: r.LinkURL,
		SortOrder: r.SortOrder, IsActive: r.IsActive,
	}
}

func mapGallery(r *repository.WebsiteGalleryRecord) dto.WebsiteGalleryResponse {
	return dto.WebsiteGalleryResponse{
		ID: r.ID, Title: r.Title, Caption: r.Caption, ImageURL: r.ImageURL, SortOrder: r.SortOrder, IsActive: r.IsActive,
	}
}

func mapNews(r *repository.NewsRecord) dto.NewsResponse {
	return dto.NewsResponse{
		ID: r.ID, Slug: r.Slug, Title: r.Title, Excerpt: r.Excerpt, Body: r.Body, ImageURL: r.ImageURL,
		IsPublished: r.IsPublished, PublishedAt: r.PublishedAt, MetaTitle: r.MetaTitle, MetaDescription: r.MetaDescription, CreatedAt: r.CreatedAt,
	}
}

func mapEvent(r *repository.EventRecord) dto.EventResponse {
	return dto.EventResponse{
		ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, Location: r.Location,
		EventDate: r.EventDate, EndDate: r.EndDate, ImageURL: r.ImageURL, IsPublished: r.IsPublished,
		MetaTitle: r.MetaTitle, MetaDescription: r.MetaDescription, CreatedAt: r.CreatedAt,
	}
}

func mapDownload(r *repository.DownloadRecord) dto.DownloadResponse {
	return dto.DownloadResponse{
		ID: r.ID, Slug: r.Slug, Title: r.Title, Description: r.Description, Category: r.Category,
		FileURL: r.FileURL, FileName: r.FileName, DownloadCount: r.DownloadCount, IsPublished: r.IsPublished, CreatedAt: r.CreatedAt,
	}
}

func mapContact(r *repository.ContactRecord) dto.ContactMessageResponse {
	return dto.ContactMessageResponse{
		ID: r.ID, Name: r.Name, Email: r.Email, Phone: r.Phone, Subject: r.Subject, Message: r.Message,
		Status: r.Status, CreatedAt: r.CreatedAt,
	}
}
