package pushers

import (
	"context"
	"errors"
	"strconv"

	"github.com/UtopiaXC/UtopiaDaily/internal/module"
	"github.com/UtopiaXC/UtopiaDaily/internal/pkg/mailer"
)

type SMTPPushModule struct{}

func NewSMTPPushModule() *SMTPPushModule {
	return &SMTPPushModule{}
}

func (s *SMTPPushModule) Meta() module.Metadata {
	return module.Metadata{
		ID:      "smtp_pusher",
		Kind:    module.KindPush,
		Source:  module.SourceBuiltIn,
		Version: module.DefaultVersion,
		Author:  "UtopiaXC",
	}
}

func (s *SMTPPushModule) ConfigSchema() []module.ConfigFieldSpec {
	return []module.ConfigFieldSpec{
		{Key: "override_smtp", Type: "bool", DefaultValue: "false", OverrideSystem: true, SortOrder: 1},
		{Key: "smtp_host", Type: "string", DefaultValue: "", SortOrder: 2},
		{Key: "smtp_port", Type: "int", DefaultValue: "587", SortOrder: 3},
		{Key: "smtp_encryption", Type: "select", DefaultValue: "tls", Options: []string{"tls", "ssl", "none"}, SortOrder: 4},
		{Key: "smtp_user", Type: "string", DefaultValue: "", SortOrder: 5},
		{Key: "smtp_password", Type: "password", DefaultValue: "", Secret: true, SortOrder: 6},
		{Key: "smtp_from", Type: "string", DefaultValue: "", SortOrder: 7},
		{Key: "smtp_from_name", Type: "string", DefaultValue: "Utopia Daily", SortOrder: 8},
		{Key: "default_recipient", Type: "string", DefaultValue: "", SortOrder: 9},
		{Key: "max_items", Type: "int", DefaultValue: "20", SortOrder: 10},
		{Key: "digest_title", Type: "string", DefaultValue: "Daily Digest", SortOrder: 11},
	}
}

func (s *SMTPPushModule) Tasks() []module.TaskSpec {
	return []module.TaskSpec{
		{Key: "send_digest", DefaultCron: "0 8 * * *", DefaultEnabled: false},
	}
}

func (s *SMTPPushModule) Locales() map[string]map[string]string {
	return map[string]map[string]string{
		"zh_CN": {
			"module.smtp_pusher.name":                          "SMTP 邮件推送",
			"module.smtp_pusher.description":                   "通过 SMTP 将最新新闻摘要发送至订阅邮箱",
			"module.smtp_pusher.config.override_smtp.desc":     "覆盖系统 SMTP 配置",
			"module.smtp_pusher.config.override_smtp.hint":     "开启后使用下方独立的 SMTP 服务器参数",
			"module.smtp_pusher.config.smtp_host.desc":         "SMTP 服务器",
			"module.smtp_pusher.config.smtp_port.desc":         "SMTP 端口",
			"module.smtp_pusher.config.smtp_encryption.desc":   "加密方式",
			"module.smtp_pusher.config.smtp_user.desc":         "SMTP 用户名",
			"module.smtp_pusher.config.smtp_password.desc":     "SMTP 密码",
			"module.smtp_pusher.config.smtp_from.desc":         "发件人邮箱",
			"module.smtp_pusher.config.smtp_from_name.desc":    "发件人名称",
			"module.smtp_pusher.config.default_recipient.desc": "默认收件人",
			"module.smtp_pusher.config.default_recipient.hint": "未配置个人订阅规则时使用的默认推送地址",
			"module.smtp_pusher.config.max_items.desc":         "单次最多条目数",
			"module.smtp_pusher.config.digest_title.desc":      "邮件标题",
			"module.smtp_pusher.task.send_digest.name":         "定时发送摘要",
			"module.smtp_pusher.task.send_digest.desc":         "按计划向订阅用户推送最新资讯",
		},
		"en_US": {
			"module.smtp_pusher.name":                          "SMTP Email Push",
			"module.smtp_pusher.description":                   "Dispatches news digests via SMTP email",
			"module.smtp_pusher.config.override_smtp.desc":     "Override System SMTP",
			"module.smtp_pusher.config.override_smtp.hint":     "Use the module SMTP settings below instead of system SMTP",
			"module.smtp_pusher.config.smtp_host.desc":         "SMTP Host",
			"module.smtp_pusher.config.smtp_port.desc":         "SMTP Port",
			"module.smtp_pusher.config.smtp_encryption.desc":   "Encryption",
			"module.smtp_pusher.config.smtp_user.desc":         "SMTP Username",
			"module.smtp_pusher.config.smtp_password.desc":     "SMTP Password",
			"module.smtp_pusher.config.smtp_from.desc":         "Sender Email",
			"module.smtp_pusher.config.smtp_from_name.desc":    "Sender Name",
			"module.smtp_pusher.config.default_recipient.desc": "Default Recipient",
			"module.smtp_pusher.config.default_recipient.hint": "Fallback recipient used when no personal rule matches",
			"module.smtp_pusher.config.max_items.desc":         "Max Items",
			"module.smtp_pusher.config.digest_title.desc":      "Digest Title",
			"module.smtp_pusher.task.send_digest.name":         "Send Scheduled Digest",
			"module.smtp_pusher.task.send_digest.desc":         "Deliver latest news digests to subscribers",
		},
	}
}

func (s *SMTPPushModule) resolveMailer(cfg module.ConfigAccessor) *mailer.Mailer {
	if cfg.GetBool("override_smtp") {
		return mailer.NewMailer(mailer.Config{
			Enabled:    true,
			Host:       cfg.GetString("smtp_host"),
			Port:       cfg.GetInt("smtp_port", 587),
			Encryption: cfg.GetString("smtp_encryption"),
			User:       cfg.GetString("smtp_user"),
			Password:   cfg.GetString("smtp_password"),
			From:       cfg.GetString("smtp_from"),
			FromName:   cfg.GetString("smtp_from_name"),
		})
	}

	port, _ := strconv.Atoi(cfg.GetSystemConfig("smtp_port"))
	if port <= 0 {
		port = 587
	}
	return mailer.NewMailer(mailer.Config{
		Enabled:    cfg.GetSystemConfig("smtp_enabled") == "true",
		Host:       cfg.GetSystemConfig("smtp_host"),
		Port:       port,
		Encryption: cfg.GetSystemConfig("smtp_encryption"),
		User:       cfg.GetSystemConfig("smtp_user"),
		Password:   cfg.GetSystemConfig("smtp_password"),
		From:       cfg.GetSystemConfig("smtp_from"),
		FromName:   cfg.GetSystemConfig("smtp_from_name"),
	})
}

func (s *SMTPPushModule) Push(ctx context.Context, taskKey string, cfg module.ConfigAccessor, payload module.PushPayload) error {
	recipient := payload.Recipient
	if recipient == "" {
		recipient = cfg.GetString("default_recipient")
	}
	if recipient == "" {
		return errors.New("recipient address is empty")
	}
	title := payload.Title
	if title == "" {
		title = cfg.GetString("digest_title")
	}
	siteName := cfg.GetSystemConfig("server_name")
	m := s.resolveMailer(cfg)
	subject, htmlBody, textBody := mailer.RenderNewsDigestEmail(siteName, title, payload.Articles)
	return m.SendHTML(recipient, subject, htmlBody, textBody)
}

func (s *SMTPPushModule) TestConfig(ctx context.Context, cfg module.ConfigAccessor) module.TestResult {
	m := s.resolveMailer(cfg)
	if err := m.TestConnection(); err != nil {
		return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
	}
	recipient := cfg.GetString("default_recipient")
	if recipient != "" {
		siteName := cfg.GetSystemConfig("server_name")
		subj, htmlBody, textBody := mailer.RenderTestEmail(siteName)
		if err := m.SendHTML(recipient, subj, htmlBody, textBody); err != nil {
			return module.TestResult{Success: false, MessageKey: "module.test.failed", Detail: err.Error()}
		}
	}
	return module.TestResult{Success: true, MessageKey: "module.test.passed"}
}
