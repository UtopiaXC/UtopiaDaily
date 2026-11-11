package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
)

type NewspaperQuotationDTO struct {
	Content string `json:"content"`
	From    string `json:"from"`
	URL     string `json:"url,omitempty"`
}

type NewspaperAttachmentDTO struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type NewspaperTagDTO struct {
	Tag        string  `json:"tag"`
	GroupName  string  `json:"group_name,omitempty"`
	Confidence float64 `json:"confidence"`
}

type NewspaperArticleDTO struct {
	ID               uint                     `json:"id"`
	ModuleID         string                   `json:"module_id"`
	Title            string                   `json:"title"`
	Summary          string                   `json:"summary"`
	Source           string                   `json:"source"`
	Author           string                   `json:"author,omitempty"`
	FromURL          string                   `json:"from_url"`
	Content          string                   `json:"content"`
	ContentType      string                   `json:"content_type"`
	RenderType       string                   `json:"render_type"`
	DatetimeReleased time.Time                `json:"datetime_released"`
	Quotation        *NewspaperQuotationDTO   `json:"quotation,omitempty"`
	Attachments      []NewspaperAttachmentDTO `json:"attachments,omitempty"`
	Tags             []NewspaperTagDTO        `json:"tags,omitempty"`
}

type NewspaperResponse struct {
	Config  NewspaperConfig       `json:"config"`
	Sources []string              `json:"sources"`
	News    []NewspaperArticleDTO `json:"news"`
}

type NewspaperService struct {
	scraperRepo *repository.ScraperRepository
	cfgService  *ConfigService
}

func NewNewspaperService(scraperRepo *repository.ScraperRepository, cfgService *ConfigService) *NewspaperService {
	return &NewspaperService{
		scraperRepo: scraperRepo,
		cfgService:  cfgService,
	}
}

func (s *NewspaperService) GetLatest(sourceFilter string) (*NewspaperResponse, error) {
	cfg := s.cfgService.GetNewspaperConfig()
	rows, err := s.scraperRepo.QueryRecent(cfg.LookbackDays, cfg.MaxItems, nil)
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	sourceSet := make(map[string]bool)
	sources := make([]string, 0)
	articles := make([]NewspaperArticleDTO, 0, len(rows))

	for _, r := range rows {
		if r.Source != "" && !sourceSet[r.Source] {
			sourceSet[r.Source] = true
			sources = append(sources, r.Source)
		}
		if sourceFilter != "" && !strings.EqualFold(r.Source, sourceFilter) && !strings.EqualFold(r.ModuleID, sourceFilter) {
			continue
		}

		dto := NewspaperArticleDTO{
			ID:               r.ID,
			ModuleID:         r.ModuleID,
			Title:            r.Title,
			Summary:          r.Summary,
			Source:           r.Source,
			Author:           r.Author,
			FromURL:          r.FromURL,
			Content:          r.Content,
			ContentType:      r.ContentType,
			RenderType:       r.RenderType,
			DatetimeReleased: r.DatetimeReleased,
		}
		if r.QuotationContent != "" {
			dto.Quotation = &NewspaperQuotationDTO{
				Content: r.QuotationContent,
				From:    r.QuotationFrom,
				URL:     r.QuotationURL,
			}
		}
		if len(r.Attachments) > 0 {
			atts := make([]NewspaperAttachmentDTO, 0, len(r.Attachments))
			for _, att := range r.Attachments {
				atts = append(atts, NewspaperAttachmentDTO{
					URL:  att.OriginalURL,
					Type: att.MediaType,
				})
			}
			dto.Attachments = atts
		}
		if len(r.ResultTags) > 0 {
			tags := make([]NewspaperTagDTO, 0, len(r.ResultTags))
			for _, rt := range r.ResultTags {
				if rt.Tag != nil {
					tags = append(tags, NewspaperTagDTO{
						Tag:        rt.Tag.Name,
						GroupName:  rt.Tag.GroupName,
						Confidence: rt.Confidence,
					})
				}
			}
			dto.Tags = tags
		}
		articles = append(articles, dto)
	}

	return &NewspaperResponse{
		Config:  cfg,
		Sources: sources,
		News:    articles,
	}, nil
}
