package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebsiteRepository interface {
	GetSettings(ctx context.Context) (*WebsiteSettingsRecord, error)
	UpdateSettings(ctx context.Context, params WebsiteSettingsParams) (*WebsiteSettingsRecord, error)
	UpdateSettingsMedia(ctx context.Context, logoURL, faviconURL string) error

	ListPages(ctx context.Context) ([]WebsitePageRecord, error)
	GetPageByID(ctx context.Context, id uuid.UUID) (*WebsitePageRecord, error)
	GetPageBySlug(ctx context.Context, slug string) (*WebsitePageRecord, error)
	CreatePage(ctx context.Context, params WebsitePageParams) (*WebsitePageRecord, error)
	UpdatePage(ctx context.Context, id uuid.UUID, params WebsitePageParams) (*WebsitePageRecord, error)
	SoftDeletePage(ctx context.Context, id uuid.UUID) error

	ListBlocks(ctx context.Context, pageID *uuid.UUID) ([]WebsiteBlockRecord, error)
	CreateBlock(ctx context.Context, params WebsiteBlockParams) (*WebsiteBlockRecord, error)
	UpdateBlock(ctx context.Context, id uuid.UUID, params WebsiteBlockParams) (*WebsiteBlockRecord, error)
	SoftDeleteBlock(ctx context.Context, id uuid.UUID) error

	ListBanners(ctx context.Context, activeOnly bool) ([]WebsiteBannerRecord, error)
	CreateBanner(ctx context.Context, params WebsiteBannerParams) (*WebsiteBannerRecord, error)
	UpdateBanner(ctx context.Context, id uuid.UUID, params WebsiteBannerParams) (*WebsiteBannerRecord, error)
	SoftDeleteBanner(ctx context.Context, id uuid.UUID) error

	ListGallery(ctx context.Context, activeOnly bool) ([]WebsiteGalleryRecord, error)
	CreateGallery(ctx context.Context, params WebsiteGalleryParams) (*WebsiteGalleryRecord, error)
	UpdateGallery(ctx context.Context, id uuid.UUID, params WebsiteGalleryParams) (*WebsiteGalleryRecord, error)
	SoftDeleteGallery(ctx context.Context, id uuid.UUID) error

	SearchNews(ctx context.Context, publishedOnly bool, limit, offset int32) ([]NewsRecord, error)
	CountNews(ctx context.Context, publishedOnly bool) (int64, error)
	GetNewsByID(ctx context.Context, id uuid.UUID) (*NewsRecord, error)
	GetNewsBySlug(ctx context.Context, slug string) (*NewsRecord, error)
	CreateNews(ctx context.Context, params NewsParams) (*NewsRecord, error)
	UpdateNews(ctx context.Context, id uuid.UUID, params NewsParams) (*NewsRecord, error)
	SoftDeleteNews(ctx context.Context, id uuid.UUID) error

	SearchEvents(ctx context.Context, publishedOnly bool, limit, offset int32) ([]EventRecord, error)
	CountEvents(ctx context.Context, publishedOnly bool) (int64, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*EventRecord, error)
	GetEventBySlug(ctx context.Context, slug string) (*EventRecord, error)
	CreateEvent(ctx context.Context, params EventParams) (*EventRecord, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, params EventParams) (*EventRecord, error)
	SoftDeleteEvent(ctx context.Context, id uuid.UUID) error

	ListDownloads(ctx context.Context, publishedOnly bool) ([]DownloadRecord, error)
	GetDownloadByID(ctx context.Context, id uuid.UUID) (*DownloadRecord, error)
	GetDownloadBySlug(ctx context.Context, slug string) (*DownloadRecord, error)
	CreateDownload(ctx context.Context, params DownloadParams) (*DownloadRecord, error)
	UpdateDownload(ctx context.Context, id uuid.UUID, params DownloadParams) (*DownloadRecord, error)
	IncrementDownload(ctx context.Context, id uuid.UUID) error
	SoftDeleteDownload(ctx context.Context, id uuid.UUID) error

	CreateContact(ctx context.Context, params ContactParams) (*ContactRecord, error)
	SearchContacts(ctx context.Context, status string, limit, offset int32) ([]ContactRecord, error)
	CountContacts(ctx context.Context, status string) (int64, error)
	UpdateContactStatus(ctx context.Context, id uuid.UUID, status string) error
	GetContactByID(ctx context.Context, id uuid.UUID) (*ContactRecord, error)

	RecordVisit(ctx context.Context, path string) error
	CountVisitsToday(ctx context.Context) (int64, error)
	CountVisitsPeriod(ctx context.Context, from, to time.Time) (int64, error)
	DashboardStats(ctx context.Context) (*WebsiteDashboardRecord, error)
}

type WebsiteSettingsRecord struct {
	ID                     uuid.UUID
	SiteName, Tagline      string
	LogoURL, FaviconURL    string
	PrimaryColor, SecondaryColor string
	FacebookURL, TwitterURL, InstagramURL, YoutubeURL string
	ContactEmail, ContactPhone, ContactAddress string
	DefaultMetaTitle, DefaultMetaDescription string
}

type WebsiteSettingsParams struct {
	SiteName, Tagline string
	PrimaryColor, SecondaryColor string
	FacebookURL, TwitterURL, InstagramURL, YoutubeURL string
	ContactEmail, ContactPhone, ContactAddress string
	DefaultMetaTitle, DefaultMetaDescription string
}

type WebsitePageRecord struct {
	ID          uuid.UUID
	Slug, Title string
	PageType, Content string
	MetaTitle, MetaDescription, OGImage string
	IsPublished bool
	SortOrder   int
}

type WebsitePageParams struct {
	Slug, Title, PageType, Content string
	MetaTitle, MetaDescription, OGImage string
	IsPublished bool
	SortOrder   int
}

type WebsiteBlockRecord struct {
	ID        uuid.UUID
	PageID    *uuid.UUID
	BlockType, Title, Content, ImageURL string
	SortOrder int
	IsActive  bool
}

type WebsiteBlockParams struct {
	PageID    *uuid.UUID
	BlockType, Title, Content, ImageURL string
	SortOrder int
	IsActive  bool
}

type WebsiteBannerRecord struct {
	ID        uuid.UUID
	Title, Subtitle, ImageURL, LinkURL string
	SortOrder int
	IsActive  bool
}

type WebsiteBannerParams struct {
	Title, Subtitle, ImageURL, LinkURL string
	SortOrder int
	IsActive  bool
}

type WebsiteGalleryRecord struct {
	ID        uuid.UUID
	Title, Caption, ImageURL string
	SortOrder int
	IsActive  bool
}

type WebsiteGalleryParams struct {
	Title, Caption, ImageURL string
	SortOrder int
	IsActive  bool
}

type NewsRecord struct {
	ID          uuid.UUID
	Slug, Title, Excerpt, Body, ImageURL string
	IsPublished bool
	PublishedAt *time.Time
	MetaTitle, MetaDescription string
	CreatedAt   time.Time
}

type NewsParams struct {
	Slug, Title, Excerpt, Body, ImageURL string
	IsPublished bool
	PublishedAt *time.Time
	MetaTitle, MetaDescription string
}

type EventRecord struct {
	ID          uuid.UUID
	Slug, Title, Description, Location, ImageURL string
	EventDate   time.Time
	EndDate     *time.Time
	IsPublished bool
	MetaTitle, MetaDescription string
	CreatedAt   time.Time
}

type EventParams struct {
	Slug, Title, Description, Location, ImageURL string
	EventDate   time.Time
	EndDate     *time.Time
	IsPublished bool
	MetaTitle, MetaDescription string
}

type DownloadRecord struct {
	ID          uuid.UUID
	Slug, Title, Description, Category, FileURL, FileName string
	DownloadCount int
	IsPublished bool
	CreatedAt   time.Time
}

type DownloadParams struct {
	Slug, Title, Description, Category, FileURL, FileName string
	IsPublished bool
}

type ContactRecord struct {
	ID        uuid.UUID
	Name, Email, Phone, Subject, Message, Status string
	CreatedAt time.Time
}

type ContactParams struct {
	Name, Email, Phone, Subject, Message string
}

type WebsiteDashboardRecord struct {
	AdmissionApps, VisitorsToday, ContactTotal, ContactNew, PendingAdmissions int64
}

type websiteRepo struct{ pool *pgxpool.Pool }

func NewWebsiteRepository(pool *pgxpool.Pool) WebsiteRepository {
	return &websiteRepo{pool: pool}
}

func (r *websiteRepo) GetSettings(ctx context.Context) (*WebsiteSettingsRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, site_name, COALESCE(tagline,''), COALESCE(logo_url,''), COALESCE(favicon_url,''),
			COALESCE(primary_color,''), COALESCE(secondary_color,''),
			COALESCE(facebook_url,''), COALESCE(twitter_url,''), COALESCE(instagram_url,''), COALESCE(youtube_url,''),
			COALESCE(contact_email,''), COALESCE(contact_phone,''), COALESCE(contact_address,''),
			COALESCE(default_meta_title,''), COALESCE(default_meta_description,'')
		FROM website_settings LIMIT 1`)
	var rec WebsiteSettingsRecord
	if err := row.Scan(&rec.ID, &rec.SiteName, &rec.Tagline, &rec.LogoURL, &rec.FaviconURL,
		&rec.PrimaryColor, &rec.SecondaryColor, &rec.FacebookURL, &rec.TwitterURL, &rec.InstagramURL, &rec.YoutubeURL,
		&rec.ContactEmail, &rec.ContactPhone, &rec.ContactAddress, &rec.DefaultMetaTitle, &rec.DefaultMetaDescription); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) UpdateSettings(ctx context.Context, params WebsiteSettingsParams) (*WebsiteSettingsRecord, error) {
	existing, _ := r.GetSettings(ctx)
	if existing == nil {
		row := r.pool.QueryRow(ctx, `
			INSERT INTO website_settings (site_name, tagline, primary_color, secondary_color,
				facebook_url, twitter_url, instagram_url, youtube_url,
				contact_email, contact_phone, contact_address, default_meta_title, default_meta_description)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`,
			params.SiteName, params.Tagline, params.PrimaryColor, params.SecondaryColor,
			params.FacebookURL, params.TwitterURL, params.InstagramURL, params.YoutubeURL,
			params.ContactEmail, params.ContactPhone, params.ContactAddress, params.DefaultMetaTitle, params.DefaultMetaDescription)
		var id uuid.UUID
		if err := row.Scan(&id); err != nil {
			return nil, err
		}
	} else {
		_, err := r.pool.Exec(ctx, `
			UPDATE website_settings SET site_name=$2, tagline=$3, primary_color=$4, secondary_color=$5,
				facebook_url=$6, twitter_url=$7, instagram_url=$8, youtube_url=$9,
				contact_email=$10, contact_phone=$11, contact_address=$12,
				default_meta_title=$13, default_meta_description=$14, updated_at=NOW() WHERE id=$1`,
			existing.ID, params.SiteName, params.Tagline, params.PrimaryColor, params.SecondaryColor,
			params.FacebookURL, params.TwitterURL, params.InstagramURL, params.YoutubeURL,
			params.ContactEmail, params.ContactPhone, params.ContactAddress, params.DefaultMetaTitle, params.DefaultMetaDescription)
		if err != nil {
			return nil, err
		}
	}
	return r.GetSettings(ctx)
}

func (r *websiteRepo) UpdateSettingsMedia(ctx context.Context, logoURL, faviconURL string) error {
	s, _ := r.GetSettings(ctx)
	if s == nil {
		return nil
	}
	q := `UPDATE website_settings SET updated_at=NOW()`
	args := []any{s.ID}
	n := 2
	if logoURL != "" {
		q += fmt.Sprintf(", logo_url=$%d", n)
		args = append(args, logoURL)
		n++
	}
	if faviconURL != "" {
		q += fmt.Sprintf(", favicon_url=$%d", n)
		args = append(args, faviconURL)
		n++
	}
	q += " WHERE id=$1"
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *websiteRepo) ListPages(ctx context.Context) ([]WebsitePageRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, title, page_type, content, COALESCE(meta_title,''), COALESCE(meta_description,''),
			COALESCE(og_image,''), is_published, sort_order
		FROM website_pages WHERE deleted_at IS NULL ORDER BY sort_order, title`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPages(rows)
}

func scanPages(rows pgx.Rows) ([]WebsitePageRecord, error) {
	var items []WebsitePageRecord
	for rows.Next() {
		var rec WebsitePageRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.PageType, &rec.Content,
			&rec.MetaTitle, &rec.MetaDescription, &rec.OGImage, &rec.IsPublished, &rec.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) GetPageByID(ctx context.Context, id uuid.UUID) (*WebsitePageRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, title, page_type, content, COALESCE(meta_title,''), COALESCE(meta_description,''),
			COALESCE(og_image,''), is_published, sort_order
		FROM website_pages WHERE id=$1 AND deleted_at IS NULL`, id)
	var rec WebsitePageRecord
	if err := row.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.PageType, &rec.Content,
		&rec.MetaTitle, &rec.MetaDescription, &rec.OGImage, &rec.IsPublished, &rec.SortOrder); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) GetPageBySlug(ctx context.Context, slug string) (*WebsitePageRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, title, page_type, content, COALESCE(meta_title,''), COALESCE(meta_description,''),
			COALESCE(og_image,''), is_published, sort_order
		FROM website_pages WHERE slug=$1 AND deleted_at IS NULL AND is_published=true`, slug)
	var rec WebsitePageRecord
	if err := row.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.PageType, &rec.Content,
		&rec.MetaTitle, &rec.MetaDescription, &rec.OGImage, &rec.IsPublished, &rec.SortOrder); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) CreatePage(ctx context.Context, params WebsitePageParams) (*WebsitePageRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO website_pages (slug, title, page_type, content, meta_title, meta_description, og_image, is_published, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, params.Slug, params.Title, params.PageType, params.Content,
		params.MetaTitle, params.MetaDescription, params.OGImage, params.IsPublished, params.SortOrder)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetPageByID(ctx, id)
}

func (r *websiteRepo) UpdatePage(ctx context.Context, id uuid.UUID, params WebsitePageParams) (*WebsitePageRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE website_pages SET slug=$2, title=$3, page_type=$4, content=$5, meta_title=$6, meta_description=$7,
			og_image=$8, is_published=$9, sort_order=$10, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`,
		id, params.Slug, params.Title, params.PageType, params.Content, params.MetaTitle, params.MetaDescription,
		params.OGImage, params.IsPublished, params.SortOrder)
	if err != nil {
		return nil, err
	}
	return r.GetPageByID(ctx, id)
}

func (r *websiteRepo) SoftDeletePage(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE website_pages SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) ListBlocks(ctx context.Context, pageID *uuid.UUID) ([]WebsiteBlockRecord, error) {
	q := `SELECT id, page_id, block_type, COALESCE(title,''), content, COALESCE(image_url,''), sort_order, is_active
		FROM website_blocks WHERE deleted_at IS NULL`
	args := []any{}
	if pageID != nil {
		q += ` AND page_id = $1`
		args = append(args, *pageID)
	}
	q += ` ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebsiteBlockRecord
	for rows.Next() {
		var rec WebsiteBlockRecord
		if err := rows.Scan(&rec.ID, &rec.PageID, &rec.BlockType, &rec.Title, &rec.Content, &rec.ImageURL, &rec.SortOrder, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CreateBlock(ctx context.Context, params WebsiteBlockParams) (*WebsiteBlockRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO website_blocks (page_id, block_type, title, content, image_url, sort_order, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, params.PageID, params.BlockType, params.Title, params.Content, params.ImageURL, params.SortOrder, params.IsActive)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	blocks, _ := r.ListBlocks(ctx, params.PageID)
	for _, b := range blocks {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) UpdateBlock(ctx context.Context, id uuid.UUID, params WebsiteBlockParams) (*WebsiteBlockRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE website_blocks SET page_id=$2, block_type=$3, title=$4, content=$5, image_url=$6, sort_order=$7, is_active=$8, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id, params.PageID, params.BlockType, params.Title, params.Content, params.ImageURL, params.SortOrder, params.IsActive)
	if err != nil {
		return nil, err
	}
	blocks, _ := r.ListBlocks(ctx, params.PageID)
	for _, b := range blocks {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) SoftDeleteBlock(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE website_blocks SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) ListBanners(ctx context.Context, activeOnly bool) ([]WebsiteBannerRecord, error) {
	q := `SELECT id, title, COALESCE(subtitle,''), image_url, COALESCE(link_url,''), sort_order, is_active FROM website_banners WHERE deleted_at IS NULL`
	if activeOnly {
		q += ` AND is_active=true`
	}
	q += ` ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebsiteBannerRecord
	for rows.Next() {
		var rec WebsiteBannerRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Subtitle, &rec.ImageURL, &rec.LinkURL, &rec.SortOrder, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CreateBanner(ctx context.Context, params WebsiteBannerParams) (*WebsiteBannerRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO website_banners (title, subtitle, image_url, link_url, sort_order, is_active)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, params.Title, params.Subtitle, params.ImageURL, params.LinkURL, params.SortOrder, params.IsActive)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	banners, _ := r.ListBanners(ctx, false)
	for _, b := range banners {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) UpdateBanner(ctx context.Context, id uuid.UUID, params WebsiteBannerParams) (*WebsiteBannerRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE website_banners SET title=$2, subtitle=$3, image_url=$4, link_url=$5, sort_order=$6, is_active=$7, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id, params.Title, params.Subtitle, params.ImageURL, params.LinkURL, params.SortOrder, params.IsActive)
	if err != nil {
		return nil, err
	}
	banners, _ := r.ListBanners(ctx, false)
	for _, b := range banners {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) SoftDeleteBanner(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE website_banners SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) ListGallery(ctx context.Context, activeOnly bool) ([]WebsiteGalleryRecord, error) {
	q := `SELECT id, title, COALESCE(caption,''), image_url, sort_order, is_active FROM website_gallery WHERE deleted_at IS NULL`
	if activeOnly {
		q += ` AND is_active=true`
	}
	q += ` ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebsiteGalleryRecord
	for rows.Next() {
		var rec WebsiteGalleryRecord
		if err := rows.Scan(&rec.ID, &rec.Title, &rec.Caption, &rec.ImageURL, &rec.SortOrder, &rec.IsActive); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CreateGallery(ctx context.Context, params WebsiteGalleryParams) (*WebsiteGalleryRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO website_gallery (title, caption, image_url, sort_order, is_active) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		params.Title, params.Caption, params.ImageURL, params.SortOrder, params.IsActive)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	items, _ := r.ListGallery(ctx, false)
	for _, g := range items {
		if g.ID == id {
			return &g, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) UpdateGallery(ctx context.Context, id uuid.UUID, params WebsiteGalleryParams) (*WebsiteGalleryRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE website_gallery SET title=$2, caption=$3, image_url=$4, sort_order=$5, is_active=$6, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id, params.Title, params.Caption, params.ImageURL, params.SortOrder, params.IsActive)
	if err != nil {
		return nil, err
	}
	items, _ := r.ListGallery(ctx, false)
	for _, g := range items {
		if g.ID == id {
			return &g, nil
		}
	}
	return nil, nil
}

func (r *websiteRepo) SoftDeleteGallery(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE website_gallery SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) SearchNews(ctx context.Context, publishedOnly bool, limit, offset int32) ([]NewsRecord, error) {
	q := `SELECT id, slug, title, COALESCE(excerpt,''), body, COALESCE(image_url,''), is_published, published_at,
		COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM news WHERE deleted_at IS NULL`
	if publishedOnly {
		q += ` AND is_published=true`
	}
	q += ` ORDER BY published_at DESC NULLS LAST, created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNews(rows)
}

func scanNews(rows pgx.Rows) ([]NewsRecord, error) {
	var items []NewsRecord
	for rows.Next() {
		var rec NewsRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.Excerpt, &rec.Body, &rec.ImageURL, &rec.IsPublished,
			&rec.PublishedAt, &rec.MetaTitle, &rec.MetaDescription, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CountNews(ctx context.Context, publishedOnly bool) (int64, error) {
	q := `SELECT COUNT(*) FROM news WHERE deleted_at IS NULL`
	if publishedOnly {
		q += ` AND is_published=true`
	}
	var n int64
	err := r.pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}

func (r *websiteRepo) GetNewsByID(ctx context.Context, id uuid.UUID) (*NewsRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, title, COALESCE(excerpt,''), body, COALESCE(image_url,''), is_published, published_at,
		COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM news WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanNews(rows)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *websiteRepo) GetNewsBySlug(ctx context.Context, slug string) (*NewsRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, title, COALESCE(excerpt,''), body, COALESCE(image_url,''), is_published, published_at,
		COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM news WHERE slug=$1 AND deleted_at IS NULL AND is_published=true`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanNews(rows)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *websiteRepo) CreateNews(ctx context.Context, params NewsParams) (*NewsRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO news (slug, title, excerpt, body, image_url, is_published, published_at, meta_title, meta_description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, params.Slug, params.Title, params.Excerpt, params.Body, params.ImageURL,
		params.IsPublished, params.PublishedAt, params.MetaTitle, params.MetaDescription)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetNewsByID(ctx, id)
}

func (r *websiteRepo) UpdateNews(ctx context.Context, id uuid.UUID, params NewsParams) (*NewsRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE news SET slug=$2, title=$3, excerpt=$4, body=$5, image_url=$6, is_published=$7, published_at=$8,
			meta_title=$9, meta_description=$10, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`,
		id, params.Slug, params.Title, params.Excerpt, params.Body, params.ImageURL, params.IsPublished, params.PublishedAt, params.MetaTitle, params.MetaDescription)
	if err != nil {
		return nil, err
	}
	return r.GetNewsByID(ctx, id)
}

func (r *websiteRepo) SoftDeleteNews(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE news SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) SearchEvents(ctx context.Context, publishedOnly bool, limit, offset int32) ([]EventRecord, error) {
	q := `SELECT id, slug, title, description, COALESCE(location,''), event_date, end_date, COALESCE(image_url,''),
		is_published, COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM events WHERE deleted_at IS NULL`
	if publishedOnly {
		q += ` AND is_published=true`
	}
	q += ` ORDER BY event_date DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func scanEvents(rows pgx.Rows) ([]EventRecord, error) {
	var items []EventRecord
	for rows.Next() {
		var rec EventRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.Description, &rec.Location, &rec.EventDate, &rec.EndDate,
			&rec.ImageURL, &rec.IsPublished, &rec.MetaTitle, &rec.MetaDescription, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CountEvents(ctx context.Context, publishedOnly bool) (int64, error) {
	q := `SELECT COUNT(*) FROM events WHERE deleted_at IS NULL`
	if publishedOnly {
		q += ` AND is_published=true`
	}
	var n int64
	err := r.pool.QueryRow(ctx, q).Scan(&n)
	return n, err
}

func (r *websiteRepo) GetEventByID(ctx context.Context, id uuid.UUID) (*EventRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, title, description, COALESCE(location,''), event_date, end_date, COALESCE(image_url,''),
		is_published, COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM events WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanEvents(rows)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *websiteRepo) GetEventBySlug(ctx context.Context, slug string) (*EventRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, title, description, COALESCE(location,''), event_date, end_date, COALESCE(image_url,''),
		is_published, COALESCE(meta_title,''), COALESCE(meta_description,''), created_at FROM events WHERE slug=$1 AND deleted_at IS NULL AND is_published=true`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanEvents(rows)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *websiteRepo) CreateEvent(ctx context.Context, params EventParams) (*EventRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO events (slug, title, description, location, event_date, end_date, image_url, is_published, meta_title, meta_description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`, params.Slug, params.Title, params.Description, params.Location,
		params.EventDate, params.EndDate, params.ImageURL, params.IsPublished, params.MetaTitle, params.MetaDescription)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetEventByID(ctx, id)
}

func (r *websiteRepo) UpdateEvent(ctx context.Context, id uuid.UUID, params EventParams) (*EventRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE events SET slug=$2, title=$3, description=$4, location=$5, event_date=$6, end_date=$7, image_url=$8,
			is_published=$9, meta_title=$10, meta_description=$11, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`,
		id, params.Slug, params.Title, params.Description, params.Location, params.EventDate, params.EndDate, params.ImageURL,
		params.IsPublished, params.MetaTitle, params.MetaDescription)
	if err != nil {
		return nil, err
	}
	return r.GetEventByID(ctx, id)
}

func (r *websiteRepo) SoftDeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE events SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) ListDownloads(ctx context.Context, publishedOnly bool) ([]DownloadRecord, error) {
	q := `SELECT id, slug, title, COALESCE(description,''), category, file_url, file_name, download_count, is_published, created_at
		FROM downloads WHERE deleted_at IS NULL`
	if publishedOnly {
		q += ` AND is_published=true`
	}
	q += ` ORDER BY category, title`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DownloadRecord
	for rows.Next() {
		var rec DownloadRecord
		if err := rows.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.Description, &rec.Category, &rec.FileURL, &rec.FileName,
			&rec.DownloadCount, &rec.IsPublished, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) GetDownloadByID(ctx context.Context, id uuid.UUID) (*DownloadRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, slug, title, COALESCE(description,''), category, file_url, file_name, download_count, is_published, created_at
		FROM downloads WHERE id=$1 AND deleted_at IS NULL`, id)
	var rec DownloadRecord
	if err := row.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.Description, &rec.Category, &rec.FileURL, &rec.FileName,
		&rec.DownloadCount, &rec.IsPublished, &rec.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) GetDownloadBySlug(ctx context.Context, slug string) (*DownloadRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, slug, title, COALESCE(description,''), category, file_url, file_name, download_count, is_published, created_at
		FROM downloads WHERE slug=$1 AND deleted_at IS NULL AND is_published=true`, slug)
	var rec DownloadRecord
	if err := row.Scan(&rec.ID, &rec.Slug, &rec.Title, &rec.Description, &rec.Category, &rec.FileURL, &rec.FileName,
		&rec.DownloadCount, &rec.IsPublished, &rec.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) CreateDownload(ctx context.Context, params DownloadParams) (*DownloadRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO downloads (slug, title, description, category, file_url, file_name, is_published)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, params.Slug, params.Title, params.Description, params.Category, params.FileURL, params.FileName, params.IsPublished)
	var id uuid.UUID
	if err := row.Scan(&id); err != nil {
		return nil, err
	}
	return r.GetDownloadByID(ctx, id)
}

func (r *websiteRepo) UpdateDownload(ctx context.Context, id uuid.UUID, params DownloadParams) (*DownloadRecord, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE downloads SET slug=$2, title=$3, description=$4, category=$5, file_url=$6, file_name=$7, is_published=$8, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id, params.Slug, params.Title, params.Description, params.Category, params.FileURL, params.FileName, params.IsPublished)
	if err != nil {
		return nil, err
	}
	return r.GetDownloadByID(ctx, id)
}

func (r *websiteRepo) IncrementDownload(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE downloads SET download_count = download_count + 1 WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) SoftDeleteDownload(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE downloads SET deleted_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *websiteRepo) CreateContact(ctx context.Context, params ContactParams) (*ContactRecord, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO contact_messages (name, email, phone, subject, message) VALUES ($1,$2,$3,$4,$5)
		RETURNING id, name, email, COALESCE(phone,''), subject, message, status, created_at`,
		params.Name, params.Email, params.Phone, params.Subject, params.Message)
	var rec ContactRecord
	if err := row.Scan(&rec.ID, &rec.Name, &rec.Email, &rec.Phone, &rec.Subject, &rec.Message, &rec.Status, &rec.CreatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) SearchContacts(ctx context.Context, status string, limit, offset int32) ([]ContactRecord, error) {
	q := `SELECT id, name, email, COALESCE(phone,''), subject, message, status, created_at FROM contact_messages WHERE deleted_at IS NULL`
	args := []any{}
	n := 1
	if status != "" {
		q += fmt.Sprintf(" AND status=$%d", n)
		args = append(args, status)
		n++
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ContactRecord
	for rows.Next() {
		var rec ContactRecord
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.Email, &rec.Phone, &rec.Subject, &rec.Message, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *websiteRepo) CountContacts(ctx context.Context, status string) (int64, error) {
	q := `SELECT COUNT(*) FROM contact_messages WHERE deleted_at IS NULL`
	args := []any{}
	if status != "" {
		q += ` AND status=$1`
		args = append(args, status)
	}
	var n int64
	err := r.pool.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *websiteRepo) UpdateContactStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE contact_messages SET status=$2, updated_at=NOW(), replied_at=CASE WHEN $2='replied' THEN NOW() ELSE replied_at END WHERE id=$1`, id, status)
	return err
}

func (r *websiteRepo) GetContactByID(ctx context.Context, id uuid.UUID) (*ContactRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, name, email, COALESCE(phone,''), subject, message, status, created_at FROM contact_messages WHERE id=$1 AND deleted_at IS NULL`, id)
	var rec ContactRecord
	if err := row.Scan(&rec.ID, &rec.Name, &rec.Email, &rec.Phone, &rec.Subject, &rec.Message, &rec.Status, &rec.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *websiteRepo) RecordVisit(ctx context.Context, path string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO website_visits (path) VALUES ($1)`, path)
	return err
}

func (r *websiteRepo) CountVisitsToday(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM website_visits WHERE visited_at >= CURRENT_DATE`).Scan(&n)
	return n, err
}

func (r *websiteRepo) CountVisitsPeriod(ctx context.Context, from, to time.Time) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM website_visits WHERE visited_at >= $1 AND visited_at < $2`, from, to).Scan(&n)
	return n, err
}

func (r *websiteRepo) DashboardStats(ctx context.Context) (*WebsiteDashboardRecord, error) {
	var s WebsiteDashboardRecord
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM website_visits WHERE visited_at >= CURRENT_DATE),
			(SELECT COUNT(*) FROM contact_messages WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM contact_messages WHERE deleted_at IS NULL AND status='new'),
			(SELECT COUNT(*) FROM admission_applications WHERE deleted_at IS NULL AND status IN ('pending','under_review'))`).
		Scan(&s.AdmissionApps, &s.VisitorsToday, &s.ContactTotal, &s.ContactNew, &s.PendingAdmissions)
	return &s, err
}
