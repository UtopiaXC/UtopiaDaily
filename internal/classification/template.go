package classification

import "strings"

func DefaultTemplate(locale string) ProfileSpec {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(locale)), "zh") {
		return ProfileSpec{
			Name:        "默认新闻分类",
			Description: "适用于综合新闻聚合的体裁、领域与区域标签模板",
			Groups: []GroupSpec{
				{
					Name:          "体裁",
					Description:   "判断文章体裁类型，单选或精选最贴合的一类",
					SelectionMode: "single",
					MaxResults:    1,
					Tags: []TagSpec{
						{Name: "快讯/消息", Priority: 10, Keywords: []string{"据悉", "消息", "通报", "电讯", "报道", "今日", "快讯", "突发", "最新", "实录", "宣布"}},
						{Name: "深度/特稿", Priority: 9, Keywords: []string{"调查", "专访", "深度", "侧记", "观察", "回眸", "前瞻", "起底", "全景", "始末"}},
						{Name: "观点/评论", Priority: 6, Keywords: []string{"社论", "社评", "评论", "专栏", "笔者", "看法", "立场", "锐评", "分析"}},
						{Name: "官方文件", Priority: 8, Keywords: []string{"公告", "白皮书", "纲要", "决议", "草案", "意见", "声明", "条例", "办法", "报告"}},
					},
				},
				{
					Name:          "领域",
					Description:   "新闻所属核心行业或公共议题领域",
					SelectionMode: "limit",
					MaxResults:    3,
					Tags: []TagSpec{
						{Name: "军事防务", Priority: 12, Keywords: []string{"美军", "解放军", "司令部", "部队", "袭击", "空袭", "开火", "击毙", "恐怖分子", "基地组织", "伊斯兰国", "武装", "航母", "导弹", "战机", "反恐"}},
						{Name: "国际政治", Priority: 11, Keywords: []string{"外交", "双边", "制裁", "谈判", "协议", "峰会", "大使", "领事", "盟友", "政权", "边境", "主权", "干涉", "联合国"}},
						{Name: "国内时政", Priority: 10, Keywords: []string{"党派", "选举", "内阁", "议会", "国会", "元首", "总理", "主席", "巡视", "反腐", "行政", "施政", "干部", "任免"}},
						{Name: "宏观经济", Priority: 10, Keywords: []string{"GDP", "CPI", "通胀", "利率", "央行", "外汇", "财政", "债务", "税收", "贸易", "关税", "经济增长", "衰退"}},
						{Name: "商业市场", Priority: 10, Keywords: []string{"股市", "A股", "收盘", "财报", "营收", "利润", "融资", "IPO", "收购", "股份", "上市公司", "投资者", "市值"}},
						{Name: "硬科技", Priority: 11, Keywords: []string{"芯片", "半导体", "航天", "卫星", "火箭", "量子", "核能", "新材料", "实验室", "科研", "技术突破", "专利"}},
						{Name: "数字互联网", Priority: 10, Keywords: []string{"人工智能", "大模型", "算法", "App", "软件", "平台", "服务器", "云服务", "数据", "网络安全", "程序员", "开源"}},
						{Name: "能源工业", Priority: 11, Keywords: []string{"石油", "原油", "天然气", "煤炭", "电力", "电网", "新能源", "光伏", "矿产", "供应链", "制造业", "工厂", "产能"}},
						{Name: "社会民生", Priority: 6, Keywords: []string{"就业", "招聘", "养老", "社保", "住房", "房价", "物价", "交通", "旅游", "人口", "社区", "居民", "福利"}},
						{Name: "法治案件", Priority: 10, Keywords: []string{"法院", "检察院", "公安", "警方", "逮捕", "判决", "诉讼", "立案", "刑拘", "犯罪", "嫌疑人", "司法", "律师"}},
						{Name: "突发灾难", Priority: 12, Keywords: []string{"地震", "台风", "洪水", "火灾", "坍塌", "爆炸", "预警", "山体滑坡", "海啸", "极端天气", "救援", "搜救"}},
						{Name: "文体娱乐", Priority: 9, Keywords: []string{"电影", "票房", "音乐", "明星", "综艺", "游戏", "电竞", "冠军", "决赛", "联赛", "奥运", "俱乐部", "转会"}},
						{Name: "医疗健康", Priority: 10, Keywords: []string{"医院", "医生", "患者", "病毒", "疫苗", "感染", "临床", "药物", "医保", "治疗", "确诊", "病例", "养生"}},
					},
				},
				{
					Name:          "区域",
					Description:   "新闻涉及的主要国家或地理区域",
					SelectionMode: "limit",
					MaxResults:    2,
					Tags: []TagSpec{
						{Name: "中国", Priority: 10, Keywords: []string{"CN", "中华人民共和国", "中共", "中央", "两岸", "内地", "港澳", "大陆", "省委", "市委", "国务院", "北京", "上海"}},
						{Name: "美国", Priority: 10, Keywords: []string{"US", "USA", "美方", "美军", "白宫", "华盛顿", "五角大楼", "联邦", "民主党", "共和党", "国会山", "纽约"}},
						{Name: "欧洲", Priority: 9, Keywords: []string{"EU", "欧盟", "英国", "法国", "德国", "俄罗斯", "乌克兰", "北约", "克里姆林宫", "伦敦", "巴黎", "柏林"}},
						{Name: "亚太", Priority: 9, Keywords: []string{"日本", "韩国", "印度", "东盟", "澳大利亚", "新加坡", "东京", "首尔", "新德里", "南海", "半岛"}},
						{Name: "中东非", Priority: 8, Keywords: []string{"以色列", "沙特", "伊朗", "埃及", "土耳其", "叙利亚", "伊拉克", "巴勒斯坦", "迪拜", "南非", "非洲", "阿拉伯"}},
						{Name: "拉美", Priority: 9, Keywords: []string{"巴西", "阿根廷", "墨西哥", "委内瑞拉", "智利", "南美", "拉丁美洲", "马杜罗", "加拉加斯"}},
						{Name: "全球", Priority: 2, Keywords: []string{"联合国", "世卫组织", "IMF", "跨国", "多边", "境外", "海外", "国际社会", "全球化"}},
					},
				},
			},
		}
	}

	return ProfileSpec{
		Name:        "Default News Taxonomy",
		Description: "Standard genre, domain, and region taxonomy for news aggregation",
		Groups: []GroupSpec{
			{
				Name:          "Genre",
				Description:   "Primary format or editorial style of the article",
				SelectionMode: "single",
				MaxResults:    1,
				Tags: []TagSpec{
					{Name: "News/Flash", Priority: 10, Keywords: []string{"breaking", "update", "report", "news", "announced", "statement", "latest", "flash", "briefing"}},
					{Name: "In-Depth", Priority: 9, Keywords: []string{"investigation", "exclusive", "analysis", "deep dive", "feature", "profile", "insight", "background", "review"}},
					{Name: "Opinion", Priority: 6, Keywords: []string{"editorial", "opinion", "column", "commentary", "view", "perspective", "analysis", "argument"}},
					{Name: "Official Doc", Priority: 8, Keywords: []string{"statement", "announcement", "white paper", "resolution", "draft", "guidelines", "regulation", "report", "notice"}},
				},
			},
			{
				Name:          "Domain",
				Description:   "Topical sectors and subject domains covered by the article",
				SelectionMode: "limit",
				MaxResults:    3,
				Tags: []TagSpec{
					{Name: "Military", Priority: 12, Keywords: []string{"military", "army", "forces", "attack", "strike", "missile", "defense", "troops", "terrorist", "navy", "air force", "combat", "Pentagon"}},
					{Name: "Int'l Politics", Priority: 11, Keywords: []string{"diplomacy", "sanctions", "treaty", "summit", "ambassador", "bilateral", "sovereignty", "border", "conflict", "UN", "foreign ministry"}},
					{Name: "Politics", Priority: 10, Keywords: []string{"election", "parliament", "congress", "president", "prime minister", "cabinet", "campaign", "party", "administration", "policy", "vote"}},
					{Name: "Macro Economy", Priority: 10, Keywords: []string{"GDP", "inflation", "interest rate", "central bank", "debt", "currency", "trade", "tariff", "tax", "recession", "economy", "fiscal"}},
					{Name: "Business", Priority: 10, Keywords: []string{"stocks", "shares", "revenue", "profit", "IPO", "merger", "acquisition", "investor", "market cap", "earnings", "Wall Street"}},
					{Name: "Hard Tech", Priority: 11, Keywords: []string{"semiconductor", "chip", "aerospace", "satellite", "quantum", "nuclear", "laboratory", "patent", "innovation", "research", "science"}},
					{Name: "Digital/IT", Priority: 10, Keywords: []string{"AI", "artificial intelligence", "LLM", "algorithm", "software", "app", "cloud", "cybersecurity", "server", "data", "open source"}},
					{Name: "Energy/Ind", Priority: 11, Keywords: []string{"oil", "gas", "petroleum", "coal", "electricity", "grid", "renewable", "solar", "battery", "manufacturing", "supply chain", "factory"}},
					{Name: "Social", Priority: 6, Keywords: []string{"employment", "jobs", "housing", "pension", "population", "traffic", "welfare", "community", "social security"}},
					{Name: "Legal/Crime", Priority: 10, Keywords: []string{"court", "police", "arrest", "lawsuit", "sentence", "prosecutor", "crime", "investigation", "suspect", "attorney", "justice"}},
					{Name: "Disaster", Priority: 12, Keywords: []string{"earthquake", "flood", "fire", "typhoon", "hurricane", "explosion", "collapse", "rescue", "evacuation", "emergency"}},
					{Name: "Ent/Sports", Priority: 9, Keywords: []string{"movie", "music", "concert", "celebrity", "game", "esports", "olympics", "championship", "league", "match", "athlete"}},
					{Name: "Health", Priority: 10, Keywords: []string{"hospital", "doctor", "virus", "vaccine", "clinical", "infection", "pandemic", "drug", "treatment", "healthcare"}},
				},
			},
			{
				Name:          "Region",
				Description:   "Main countries or geopolitical regions mentioned in the article",
				SelectionMode: "limit",
				MaxResults:    2,
				Tags: []TagSpec{
					{Name: "China", Priority: 10, Keywords: []string{"China", "Beijing", "Shanghai", "PRC", "Mainland", "Hong Kong", "Taiwan", "Shenzhen"}},
					{Name: "USA", Priority: 10, Keywords: []string{"USA", "US", "America", "Washington", "White House", "Congress", "Federal", "Republican", "Democrat"}},
					{Name: "Europe", Priority: 9, Keywords: []string{"Europe", "EU", "UK", "France", "Germany", "Russia", "Ukraine", "NATO", "London", "Paris", "Berlin", "Kremlin"}},
					{Name: "Asia-Pacific", Priority: 9, Keywords: []string{"Japan", "Korea", "India", "ASEAN", "Australia", "Tokyo", "Seoul", "Delhi", "South China Sea"}},
					{Name: "Middle East/Africa", Priority: 8, Keywords: []string{"Israel", "Iran", "Saudi Arabia", "Egypt", "Turkey", "Syria", "Gaza", "Africa", "South Africa", "Dubai"}},
					{Name: "LatAm", Priority: 9, Keywords: []string{"Brazil", "Argentina", "Mexico", "Venezuela", "Chile", "Latin America", "South America"}},
					{Name: "Global", Priority: 2, Keywords: []string{"UN", "WHO", "IMF", "global", "international", "world", "multilateral", "cross-border"}},
				},
			},
		},
	}
}
