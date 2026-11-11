package service

import (
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/UtopiaXC/UtopiaDaily/internal/classification"
	"github.com/UtopiaXC/UtopiaDaily/internal/common"
	"github.com/UtopiaXC/UtopiaDaily/internal/config"
	"github.com/UtopiaXC/UtopiaDaily/internal/event"
	"github.com/UtopiaXC/UtopiaDaily/internal/logger"
	"github.com/UtopiaXC/UtopiaDaily/internal/model"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/i18n"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/mailer"
	"github.com/UtopiaXC/UtopiaDaily/internal/repository"
)

type SystemConfigView struct {
	Key            string              `json:"key"`
	Tab            string              `json:"tab"`
	Group          string              `json:"group"`
	Type           string              `json:"type"`
	Value          string              `json:"value"`
	DescriptionKey string              `json:"description"`
	HintKey        string              `json:"hint,omitempty"`
	Options        []config.OptionItem `json:"options,omitempty"`
	IsEditable     bool                `json:"is_editable"`
	IsSecret       bool                `json:"is_secret"`
	SortOrder      int                 `json:"sort_order"`
}

type PublicSystemInfo struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Version             string `json:"version"`
	DefaultLocale       string `json:"default_locale"`
	AllowRegistration   bool   `json:"allow_registration"`
	SetupRequired       bool   `json:"setup_required"`
	NewspaperVisibility string `json:"newspaper_visibility"`
}

type SiteInfoDTO struct {
	Initialized         bool   `json:"initialized"`
	SetupRequired       bool   `json:"setup_required"`
	SiteName            string `json:"site_name"`
	Description         string `json:"description"`
	Version             string `json:"version"`
	DefaultLocale       string `json:"default_locale"`
	AllowRegistration   bool   `json:"allow_registration"`
	SMTPEnabled         bool   `json:"smtp_enabled"`
	NewspaperTitle      string `json:"newspaper_title"`
	NewspaperSubtitle   string `json:"newspaper_subtitle"`
	NewspaperVisibility string `json:"newspaper_visibility"`
}

type SMTPTestInput struct {
	Recipient  string `json:"recipient"`
	Host       string `json:"smtp_host"`
	Port       string `json:"smtp_port"`
	Encryption string `json:"smtp_encryption"`
	User       string `json:"smtp_user"`
	Password   string `json:"smtp_password"`
	From       string `json:"smtp_from"`
	FromName   string `json:"smtp_from_name"`
}

type NewspaperConfig struct {
	Title            string `json:"title"`
	Subtitle         string `json:"subtitle"`
	LookbackDays     int    `json:"lookback_days"`
	MaxItems         int    `json:"max_items"`
	Theme            string `json:"theme"`
	Sort             string `json:"sort"`
	ShowTags         bool   `json:"show_tags"`
	ShowSummary      bool   `json:"show_summary"`
	ShowAuthor       bool   `json:"show_author"`
	ShowSourceFilter bool   `json:"show_source_filter"`
	AllowThemeSwitch bool   `json:"allow_theme_switch"`
}

type ConfigService struct {
	cfgRepo     *repository.ConfigRepository
	userRepo    *repository.UserRepository
	scraperRepo *repository.ScraperRepository
	i18nMgr     *i18n.Assembler
	classifier  classification.Classifier
	clsSvc      *ClassificationService
	bus         *event.Bus
	baseCfg     *config.Config
}

func NewConfigService(
	cfgRepo *repository.ConfigRepository,
	userRepo *repository.UserRepository,
	scraperRepo *repository.ScraperRepository,
	i18nMgr *i18n.Assembler,
	classifier classification.Classifier,
	bus *event.Bus,
	baseCfg *config.Config,
) *ConfigService {
	return &ConfigService{
		cfgRepo:     cfgRepo,
		userRepo:    userRepo,
		scraperRepo: scraperRepo,
		i18nMgr:     i18nMgr,
		classifier:  classifier,
		bus:         bus,
		baseCfg:     baseCfg,
	}
}

func (s *ConfigService) GetDefaultLocale() string {
	defLocale, err := s.cfgRepo.Get("default_locale")
	if err != nil || strings.TrimSpace(defLocale) == "" {
		defLocale = s.baseCfg.Site.DefaultLocale
	}
	if strings.TrimSpace(defLocale) == "" {
		return "en_US"
	}
	return strings.TrimSpace(defLocale)
}

func (s *ConfigService) GetPublicSystemInfo() PublicSystemInfo {
	m, _ := s.cfgRepo.GetMap()
	hasAdmin, _ := s.userRepo.HasAdmin()
	name := m["server_name"]
	if name == "" {
		name = s.baseCfg.Site.Name
	}
	desc := m["site_description"]
	if desc == "" {
		desc = s.baseCfg.Site.Description
	}
	defLocale := m["default_locale"]
	if defLocale == "" {
		defLocale = s.baseCfg.Site.DefaultLocale
	}
	if defLocale == "" {
		defLocale = "en_US"
	}
	return PublicSystemInfo{
		Name:                name,
		Description:         desc,
		Version:             config.Version,
		DefaultLocale:       defLocale,
		AllowRegistration:   m["allow_registration"] == "true" && m["smtp_enabled"] == "true",
		SetupRequired:       !hasAdmin,
		NewspaperVisibility: normalizeVisibility(m["newspaper_visibility"]),
	}
}

func (s *ConfigService) AttachClassificationService(clsSvc *ClassificationService) {
	s.clsSvc = clsSvc
}

func (s *ConfigService) GetNewspaperVisibility() string {
	v, _ := s.cfgRepo.Get("newspaper_visibility")
	return normalizeVisibility(v)
}

func normalizeVisibility(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "registered") {
		return "registered"
	}
	return "public"
}

func (s *ConfigService) GetSiteInfo() SiteInfoDTO {
	m, _ := s.cfgRepo.GetMap()
	hasAdmin, _ := s.userRepo.HasAdmin()
	name := m["server_name"]
	if name == "" {
		name = s.baseCfg.Site.Name
	}
	desc := m["site_description"]
	if desc == "" {
		desc = s.baseCfg.Site.Description
	}
	defLocale := m["default_locale"]
	if defLocale == "" {
		defLocale = s.baseCfg.Site.DefaultLocale
	}
	if defLocale == "" {
		defLocale = "en_US"
	}
	newsCfg := s.GetNewspaperConfig()
	return SiteInfoDTO{
		Initialized:         hasAdmin,
		SetupRequired:       !hasAdmin,
		SiteName:            name,
		Description:         desc,
		Version:             config.Version,
		DefaultLocale:       defLocale,
		AllowRegistration:   m["allow_registration"] == "true" && m["smtp_enabled"] == "true",
		SMTPEnabled:         m["smtp_enabled"] == "true",
		NewspaperTitle:      newsCfg.Title,
		NewspaperSubtitle:   newsCfg.Subtitle,
		NewspaperVisibility: normalizeVisibility(m["newspaper_visibility"]),
	}
}

func parseConfigBoolDefault(m map[string]string, key string, def bool) bool {
	val, ok := m[key]
	if !ok || strings.TrimSpace(val) == "" {
		return def
	}
	return strings.EqualFold(strings.TrimSpace(val), "true")
}

func (s *ConfigService) GetNewspaperConfig() NewspaperConfig {
	m, _ := s.cfgRepo.GetMap()
	title := m["newspaper_title"]
	if title == "" {
		title = m["server_name"]
	}
	if title == "" {
		title = "Utopia Daily"
	}
	subtitle := m["newspaper_subtitle"]
	if subtitle == "" {
		subtitle = "Daily News Digest"
	}
	lookback, _ := strconv.Atoi(m["newspaper_lookback_days"])
	if lookback <= 0 {
		lookback = 3
	}
	maxItems, _ := strconv.Atoi(m["newspaper_max_items"])
	if maxItems <= 0 {
		maxItems = 100
	}
	theme := strings.ToLower(strings.TrimSpace(m["newspaper_theme"]))
	if theme != "classic" && theme != "modern" && theme != "feed" {
		theme = "classic"
	}
	sortMode := strings.ToLower(strings.TrimSpace(m["newspaper_sort"]))
	if sortMode != "newest" && sortMode != "source" {
		sortMode = "newest"
	}
	return NewspaperConfig{
		Title:            title,
		Subtitle:         subtitle,
		LookbackDays:     lookback,
		MaxItems:         maxItems,
		Theme:            theme,
		Sort:             sortMode,
		ShowTags:         parseConfigBoolDefault(m, "newspaper_show_tags", true),
		ShowSummary:      parseConfigBoolDefault(m, "newspaper_show_summary", false),
		ShowAuthor:       parseConfigBoolDefault(m, "newspaper_show_author", true),
		ShowSourceFilter: parseConfigBoolDefault(m, "newspaper_show_source_filter", true),
		AllowThemeSwitch: parseConfigBoolDefault(m, "newspaper_allow_theme_switch", true),
	}
}

func (s *ConfigService) ListSystemConfigs(allowedTabs map[string]bool) ([]SystemConfigView, error) {
	dbMap, err := s.cfgRepo.GetMap()
	if err != nil {
		return nil, common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	var langOptions []config.OptionItem
	if s.i18nMgr != nil {
		for _, lang := range s.i18nMgr.Languages() {
			langOptions = append(langOptions, config.OptionItem{
				Value: lang.Code,
				Label: lang.Name,
			})
		}
	}

	specs := config.SystemFieldRegistry()
	out := make([]SystemConfigView, 0, len(specs))
	for _, spec := range specs {
		if len(allowedTabs) > 0 && !allowedTabs[spec.Tab] {
			continue
		}
		val, ok := dbMap[spec.Key]
		if !ok {
			val = spec.DefaultValue
		}
		if spec.IsSecret && val != "" {
			val = "********"
		}
		opts := spec.Options
		if spec.Key == "default_locale" && len(langOptions) > 0 {
			opts = langOptions
		}
		out = append(out, SystemConfigView{
			Key:            spec.Key,
			Tab:            spec.Tab,
			Group:          spec.Group,
			Type:           spec.ValueType,
			Value:          val,
			DescriptionKey: spec.DescriptionKey,
			HintKey:        spec.HintKey,
			Options:        opts,
			IsEditable:     spec.IsEditable,
			IsSecret:       spec.IsSecret,
			SortOrder:      spec.SortOrder,
		})
	}
	return out, nil
}

func (s *ConfigService) UpdateConfig(key, value string, isFrontendTab bool) error {
	var target *config.SystemFieldSpec
	for _, spec := range config.SystemFieldRegistry() {
		if spec.Key == key {
			specCopy := spec
			target = &specCopy
			break
		}
	}
	if target == nil {
		return common.NewAPIError(http.StatusNotFound, common.ErrConfigNotFound)
	}
	if isFrontendTab && target.Tab != "frontend" {
		return common.NewAPIError(http.StatusForbidden, common.ErrAuthForbidden)
	}
	if !target.IsEditable {
		return common.NewAPIError(http.StatusBadRequest, common.ErrConfigNotEditable)
	}

	if target.IsSecret && value == "********" {
		return nil
	}

	value = strings.TrimSpace(value)
	if err := validateConfigFieldValue(key, value, *target); err != nil {
		return err
	}

	if key == "allow_registration" && value == "true" {
		smtpEnabled, _ := s.cfgRepo.Get("smtp_enabled")
		if smtpEnabled != "true" {
			return common.NewAPIError(http.StatusBadRequest, common.ErrRegistrationSMTPRequired)
		}
	}
	if key == "base_url" {
		value = strings.TrimRight(value, "/")
	}

	if err := s.cfgRepo.Set(key, value); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	switch key {
	case "log_level":
		logger.SetLevel(value)
	case "log_permanent_level":
		logger.SetPermanentLevel(value)
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.config.updated", map[string]any{
			"key": key,
		}, nil, "config", key)
	}
	return nil
}

func (s *ConfigService) UpdateBatchConfigs(configs map[string]any, isFrontendTab bool) error {
	specMap := make(map[string]config.SystemFieldSpec)
	for _, spec := range config.SystemFieldRegistry() {
		specMap[spec.Key] = spec
	}
	dbMap, err := s.cfgRepo.GetMap()
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	effectiveSMTP := dbMap["smtp_enabled"]
	if !isFrontendTab {
		if rawSMTP, ok := configs["smtp_enabled"]; ok {
			effectiveSMTP = strings.TrimSpace(stringifyConfigScalar(rawSMTP))
		}
	}

	toSave := make(map[string]string)
	for key, rawVal := range configs {
		spec, ok := specMap[key]
		if !ok {
			continue
		}
		if isFrontendTab && spec.Tab != "frontend" {
			return common.NewAPIError(http.StatusForbidden, common.ErrAuthForbidden)
		}
		if !spec.IsEditable {
			continue
		}
		val := strings.TrimSpace(stringifyConfigScalar(rawVal))
		if spec.IsSecret && val == "********" {
			continue
		}
		if err := validateConfigFieldValue(key, val, spec); err != nil {
			return err
		}
		if key == "allow_registration" && val == "true" && effectiveSMTP != "true" {
			return common.NewAPIError(http.StatusBadRequest, common.ErrRegistrationSMTPRequired)
		}
		if key == "base_url" {
			val = strings.TrimRight(val, "/")
		}
		existing, exists := dbMap[key]
		if !exists {
			existing = spec.DefaultValue
		}
		if val != existing || !exists {
			toSave[key] = val
		}
	}

	if len(toSave) == 0 {
		return nil
	}
	if err := s.cfgRepo.SetMany(toSave); err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}

	if lvl, ok := toSave["log_level"]; ok {
		logger.SetLevel(lvl)
	}
	if permLvl, ok := toSave["log_permanent_level"]; ok {
		logger.SetPermanentLevel(permLvl)
	}
	if s.bus != nil {
		for k := range toSave {
			s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.config.updated", map[string]any{
				"key": k,
			}, nil, "config", k)
		}
	}
	return nil
}

func (s *ConfigService) TestSMTP(input SMTPTestInput) error {
	m, err := s.cfgRepo.GetMap()
	if err != nil {
		return common.NewAPIError(http.StatusInternalServerError, common.ErrDatabaseError)
	}
	if v := strings.TrimSpace(input.Host); v != "" {
		m["smtp_host"] = v
	}
	if v := strings.TrimSpace(input.Port); v != "" {
		m["smtp_port"] = v
	}
	if v := strings.TrimSpace(input.Encryption); v != "" {
		m["smtp_encryption"] = v
	}
	if v := strings.TrimSpace(input.User); v != "" {
		m["smtp_user"] = v
	}
	if input.Password != "" && input.Password != "********" {
		m["smtp_password"] = input.Password
	}
	if v := strings.TrimSpace(input.From); v != "" {
		m["smtp_from"] = v
	}
	if v := strings.TrimSpace(input.FromName); v != "" {
		m["smtp_from_name"] = v
	}

	host := strings.TrimSpace(m["smtp_host"])
	from := strings.TrimSpace(m["smtp_from"])
	if host == "" || from == "" {
		return common.NewAPIError(http.StatusBadRequest, common.ErrSMTPNotConfigured)
	}
	port, _ := strconv.Atoi(m["smtp_port"])
	if port <= 0 {
		port = 465
	}

	ml := mailer.NewMailer(mailer.Config{
		Enabled:    true,
		Host:       host,
		Port:       port,
		Encryption: m["smtp_encryption"],
		From:       from,
		FromName:   m["smtp_from_name"],
		User:       m["smtp_user"],
		Password:   m["smtp_password"],
	})

	recipient := strings.TrimSpace(input.Recipient)
	if recipient == "" {
		if err := ml.TestConnection(); err != nil {
			return common.NewAPIError(http.StatusBadGateway, common.ErrSMTPTestFailed, map[string]any{"detail": err.Error()})
		}
		return nil
	}
	if _, err := mail.ParseAddress(recipient); err != nil {
		return common.NewAPIError(http.StatusBadRequest, common.ErrAuthEmailInvalid)
	}

	siteName := m["server_name"]
	subj, htmlBody, textBody := mailer.RenderTestEmail(siteName)
	if err := ml.SendHTML(recipient, subj, htmlBody, textBody); err != nil {
		return common.NewAPIError(http.StatusBadGateway, common.ErrSMTPTestFailed, map[string]any{"detail": err.Error()})
	}

	if s.bus != nil {
		s.bus.Emit(model.EventLevelNormal, model.EventCategorySystem, "event.smtp.tested", map[string]any{
			"recipient": recipient,
		}, nil, "smtp", recipient)
	}
	return nil
}

func validateConfigFieldValue(key, value string, spec config.SystemFieldSpec) error {
	switch spec.ValueType {
	case "int":
		n, err := strconv.Atoi(value)
		if err != nil {
			return common.NewAPIError(http.StatusBadRequest, common.ErrConfigValidationFailed)
		}
		if key == "classification_max_input_chars" && n < 0 {
			return common.NewAPIError(http.StatusBadRequest, common.ErrConfigValidationFailed)
		}
		if key == "classification_confidence" && (n < 1 || n > 100) {
			return common.NewAPIError(http.StatusBadRequest, common.ErrConfigValidationFailed)
		}
	case "boolean":
		if value != "true" && value != "false" {
			return common.NewAPIError(http.StatusBadRequest, common.ErrConfigValidationFailed)
		}
	case "select":
		if len(spec.Options) > 0 {
			valid := false
			for _, opt := range spec.Options {
				if opt.Value == value {
					valid = true
					break
				}
			}
			if !valid {
				return common.NewAPIError(http.StatusBadRequest, common.ErrConfigValidationFailed)
			}
		}
	}
	return nil
}

func stringifyConfigScalar(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (s *ConfigService) GetClassificationSlot() classification.SlotStatus {
	groups, tags := s.scraperRepo.CountClassificationSlot()
	provider, _ := s.cfgRepo.Get("classification_provider")
	if strings.TrimSpace(provider) == "" {
		if s.classifier != nil {
			provider = s.classifier.Provider()
		} else {
			provider = "none"
		}
	}
	workerState := "idle"
	var pendingTasks int64
	if s.clsSvc != nil {
		workerState, _ = s.clsSvc.WorkerState()
		if qStatus, err := s.clsSvc.GetQueueStatus(); err == nil {
			pendingTasks = qStatus.Stats.Pending + qStatus.Stats.Running
		}
	}
	return classification.SlotStatus{
		Provider:      provider,
		SlotReady:     true,
		TagGroupCount: groups,
		TagCount:      tags,
		WorkerState:   workerState,
		PendingTasks:  pendingTasks,
	}
}
