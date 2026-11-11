package service

import (
	"net/http"
	"time"

	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
)

type DataArticleTagDTO struct {
	TagID      uint    `json:"tag_id"`
	Tag        string  `json:"tag"`
	GroupName  string  `json:"group_name"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
	Reason     string  `json:"reason,omitempty"`
}

type DataArticleSummaryDTO struct {
	ID               uint                `json:"id"`
	ModuleID         string              `json:"module_id"`
	TaskKey          string              `json:"task_key,omitempty"`
	Title            string              `json:"title"`
	Summary          string              `json:"summary"`
	Source           string              `json:"source"`
	Author           string              `json:"author,omitempty"`
	FromURL          string              `json:"from_url"`
	ContentType      string              `json:"content_type"`
	RenderType       string              `json:"render_type"`
	DatetimeReleased time.Time           `json:"datetime_released"`
	CreatedAt        time.Time           `json:"created_at"`
	Tags             []DataArticleTagDTO `json:"tags"`
}

type DataArticleDetailDTO struct {
	DataArticleSummaryDTO
	Fingerprint string                   `json:"fingerprint"`
	Content     string                   `json:"content"`
	Quotation   *NewspaperQuotationDTO   `json:"quotation,omitempty"`
	Attachments []NewspaperAttachmentDTO `json:"attachments,omitempty"`
}

type DataService struct {
	scraperRepo *repository.ScraperRepository
	clsRepo     *repository.ClassificationRepository
	clsSvc      *ClassificationService
}

func NewDataService(
	scraperRepo *repository.ScraperRepository,
	clsRepo *repository.ClassificationRepository,
	clsSvc *ClassificationService,
) *DataService {
	return &DataService{
		scraperRepo: scraperRepo,
		clsRepo:     clsRepo,
		clsSvc:      clsSvc,
	}
}

func (s *DataService) ListArticles(params repository.DataQueryParams) ([]DataArticleSummaryDTO, int64, []string, error) {
	rows, total, err := s.scraperRepo.QueryPaginated(params)
	if err != nil {
		return nil, 0, nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	mods, _ := s.scraperRepo.ListDistinctModuleIDs()
	out := make([]DataArticleSummaryDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDataSummaryDTO(r))
	}
	return out, total, mods, nil
}

func (s *DataService) GetArticleDetail(id uint) (*DataArticleDetailDTO, error) {
	rec, err := s.scraperRepo.GetDetailByID(id)
	if err != nil {
		return nil, common.NewAPIError(http.StatusNotFound, common.ErrRecordNotFound)
	}
	dto := &DataArticleDetailDTO{
		DataArticleSummaryDTO: toDataSummaryDTO(*rec),
		Fingerprint:           rec.Fingerprint,
		Content:               rec.Content,
	}
	if rec.QuotationContent != "" {
		dto.Quotation = &NewspaperQuotationDTO{
			Content: rec.QuotationContent,
			From:    rec.QuotationFrom,
			URL:     rec.QuotationURL,
		}
	}
	for _, att := range rec.Attachments {
		dto.Attachments = append(dto.Attachments, NewspaperAttachmentDTO{
			URL:  att.OriginalURL,
			Type: att.MediaType,
		})
	}
	return dto, nil
}

func (s *DataService) DeleteArticles(ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	cnt, err := s.scraperRepo.DeleteByIDs(ids)
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return cnt, nil
}

func (s *DataService) AddTag(resultID, tagID uint) error {
	if resultID == 0 || tagID == 0 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	if _, err := s.scraperRepo.GetDetailByID(resultID); err != nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrRecordNotFound)
	}
	if err := s.clsRepo.AddManualResultTag(resultID, tagID); err != nil {
		return common.SanitizeDBError(err)
	}
	return nil
}

func (s *DataService) RemoveTag(resultID, tagID uint) error {
	if resultID == 0 || tagID == 0 {
		return common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	if err := s.clsRepo.RemoveResultTag(resultID, tagID); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return nil
}

func (s *DataService) EnqueueReclassify(ids []uint) (int, error) {
	if len(ids) == 0 {
		return 0, common.NewAPIError(http.StatusBadRequest, common.ErrValidationFailed)
	}
	pairs, err := s.scraperRepo.GetBatchModuleMap(ids)
	if err != nil {
		return 0, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	return s.clsSvc.EnqueueBatch(pairs)
}

func toDataSummaryDTO(r model.ScraperResult) DataArticleSummaryDTO {
	tags := make([]DataArticleTagDTO, 0, len(r.ResultTags))
	for _, rt := range r.ResultTags {
		if rt.Tag != nil {
			gName := ""
			if rt.Tag.Group != nil {
				gName = rt.Tag.Group.Name
			}
			tags = append(tags, DataArticleTagDTO{
				TagID:      rt.TagID,
				Tag:        rt.Tag.Name,
				GroupName:  gName,
				Confidence: rt.Confidence,
				Provider:   rt.Provider,
				Reason:     rt.Reason,
			})
		}
	}
	return DataArticleSummaryDTO{
		ID:               r.ID,
		ModuleID:         r.ModuleID,
		TaskKey:          r.TaskKey,
		Title:            r.Title,
		Summary:          r.Summary,
		Source:           r.Source,
		Author:           r.Author,
		FromURL:          r.FromURL,
		ContentType:      r.ContentType,
		RenderType:       r.RenderType,
		DatetimeReleased: r.DatetimeReleased,
		CreatedAt:        r.CreatedAt,
		Tags:             tags,
	}
}
