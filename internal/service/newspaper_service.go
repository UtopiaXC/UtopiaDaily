package service

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
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
	Provider   string  `json:"provider,omitempty"`
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
	newsRepo    *repository.NewspaperRepository
	modRepo     *repository.ModuleRepository
	cfgService  *ConfigService
}

func NewNewspaperService(
	scraperRepo *repository.ScraperRepository,
	newsRepo *repository.NewspaperRepository,
	modRepo *repository.ModuleRepository,
	cfgService *ConfigService,
) *NewspaperService {
	return &NewspaperService{
		scraperRepo: scraperRepo,
		newsRepo:    newsRepo,
		modRepo:     modRepo,
		cfgService:  cfgService,
	}
}

func (s *NewspaperService) GetLatest(sourceFilter string) (*NewspaperResponse, error) {
	cfg := s.cfgService.GetNewspaperConfig()

	var rows []model.ScraperResult
	if s.newsRepo != nil && s.modRepo != nil {
		scrapers, _ := s.modRepo.ListByKind(model.ModuleKindScraper)
		cfgMap, _ := s.newsRepo.GetModuleConfigMap()

		seenIDs := make(map[uint]bool)
		hasExplicitScraper := len(scrapers) > 0

		if hasExplicitScraper {
			for _, m := range scrapers {
				mc, hasRow := cfgMap[m.ID]
				if hasRow && !mc.Visible {
					continue
				}
				lookback := cfg.LookbackDays
				if hasRow && mc.LookbackDaysOverride > 0 {
					lookback = mc.LookbackDaysOverride
				}
				modLimit := cfg.MaxItems
				if hasRow && mc.MaxItemsOverride > 0 {
					modLimit = mc.MaxItemsOverride
				}
				modRows, err := s.scraperRepo.QueryByModule(m.ID, lookback, modLimit)
				if err != nil {
					return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
				}
				for _, r := range modRows {
					if !seenIDs[r.ID] {
						seenIDs[r.ID] = true
						rows = append(rows, r)
					}
				}
			}
		} else {
			var err error
			rows, err = s.scraperRepo.QueryRecent(cfg.LookbackDays, cfg.MaxItems, nil)
			if err != nil {
				return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
			}
		}
	} else {
		var err error
		rows, err = s.scraperRepo.QueryRecent(cfg.LookbackDays, cfg.MaxItems, nil)
		if err != nil {
			return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
		}
	}

	if cfg.Sort == "source" {
		sort.SliceStable(rows, func(i, j int) bool {
			si := strings.ToLower(rows[i].Source)
			sj := strings.ToLower(rows[j].Source)
			if si != sj {
				return si < sj
			}
			if !rows[i].DatetimeReleased.Equal(rows[j].DatetimeReleased) {
				return rows[i].DatetimeReleased.After(rows[j].DatetimeReleased)
			}
			return rows[i].ID > rows[j].ID
		})
	} else {
		sort.SliceStable(rows, func(i, j int) bool {
			if !rows[i].DatetimeReleased.Equal(rows[j].DatetimeReleased) {
				return rows[i].DatetimeReleased.After(rows[j].DatetimeReleased)
			}
			return rows[i].ID > rows[j].ID
		})
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
		if len(articles) >= cfg.MaxItems {
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
					gName := ""
					if rt.Tag.Group != nil {
						gName = rt.Tag.Group.Name
					}
					tags = append(tags, NewspaperTagDTO{
						Tag:        rt.Tag.Name,
						GroupName:  gName,
						Confidence: rt.Confidence,
						Provider:   rt.Provider,
					})
				}
			}
			dto.Tags = tags
		}
		articles = append(articles, dto)
	}

	sort.Strings(sources)

	return &NewspaperResponse{
		Config:  cfg,
		Sources: sources,
		News:    articles,
	}, nil
}

func (s *NewspaperService) ListModuleConfigs() ([]repository.NewspaperModuleConfigDTO, error) {
	items, err := s.newsRepo.ListModuleConfigs()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return items, nil
}

func (s *NewspaperService) SaveModuleConfigs(items []repository.NewspaperModuleConfigInput) error {
	if err := s.newsRepo.SaveModuleConfigs(items); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}
