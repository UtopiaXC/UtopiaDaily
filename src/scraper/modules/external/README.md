# 外部模组开发指南
**警告：本文档由AI（Gemini）生成，只用作暂时参考。正式文档稍后提供。**  
本文档介绍如何为 Utopia-Daily 开发外部抓取模组。

## 1. 目录结构

在 `src/scraper/modules/external/` 目录下创建一个新的文件夹，文件夹名称即为**模组ID**。
在该文件夹内，必须包含一个 `controller.py` 文件作为入口。

示例结构：
```text
src/scraper/modules/external/
└── my_news_scraper/       <-- 模组ID
    ├── controller.py      <-- 核心入口文件 (必须)
    ├── scraper.py         <-- 业务逻辑 (推荐分离)
    ├── locales/           <-- 国际化资源目录 (推荐)
    │   ├── en_US.json
    │   └── zh_CN.json
    └── requirements.txt   <-- 依赖文件 (可选)
```

## 2. 开发规范 (controller.py)

`controller.py` 需要包含以下三个部分：

### 2.1 模组元数据

定义 `MODULE_META` 字典：

```python
MODULE_META = {
    "id" : "my_news_scraper",
    "name": "module.my_news.name", # 推荐使用 I18n Key
    "description": "module.my_news.desc", # 推荐使用 I18n Key
    "version": "1.0.0",
    "author": "开发者名称"
}
```

### 2.2 模组类

创建一个类，继承自 `src.scraper.modules.base_module.BaseModule`。

### 2.3 入口函数

定义 `create_module(context)` 函数：

```python
def create_module(context):
    return MyModule(context)
```

## 3. API 参考

### 3.1 依赖管理

#### `install_requirements`
安装 `requirements.txt` 中的依赖到模组专属目录。建议在 `enable_module` 和 `test_module` 中调用。

```python
def install_requirements(self, requirements_file: str = "requirements.txt") -> bool
```

### 3.2 配置管理

#### `set_module_config`
注册配置项。通常在 `enable_module` 中调用。

```python
def set_module_config(self, key: str, description: str, value: str, value_type: str = "string", options: Optional[Union[List, Dict]] = None, force_init: bool = False, hint: str = "", regular: str = "")
```
*   **value_type**: 支持 `text`, `int`,  `double`, `switch`, `select`, `password`, `array`, `date`, `datetime`。
*   **options**: 
    *   用于 `select`: 可选的选项，支持key-value `{"A": "Label A", "B": "Label B"}`。
    *   用于 `switch`: 字典 `{"true": "开启", "false": "关闭"}`。Switch 最好仅用于表示开/关状态。
*   **regular**: 正则验证表达式。

#### `get_module_config`
获取配置值。

```python
def get_module_config(self, key: str) -> Any
```

### 3.3 任务调度

#### `set_module_schedule_task`
注册定时任务。通常在 `enable_module` 中调用。

```python
def set_module_schedule_task(self, key: str, description: str, cron: str = "", force_init: bool = False)
```

### 3.4 数据处理

#### `save_structured_results`
保存抓取结果。

```python
def save_structured_results(self, value: Dict[str, Any], fingerprint: str = "") -> Dict[str, Any]
```
*   **value**: 数据字典，推荐结构如下：
    ```json
    {
        "title": "标题",
        "summary": "摘要",
        "source": "来源名称",
        "from_url": "原始链接",
        "content": "正文内容",
        "content_type": "text/html/markdown",
        "datetime_released": "发布时间 (ISO格式)",
        "quotation": {
            "content": "引用内容",
            "content_type": "text",
            "from": "引用来源",
            "datetime_released": "引用时间"
        },
        "tags": [{"tag": "标签名", "confidence": 0.9}],
        "metadata": {}
    }
    ```
*   **fingerprint**: 去重指纹（如 URL）。未提供时，系统会自动处理，但建议提供以防止重复抓取。

#### `mark_message_tag`
调用系统 AI 模型为文本生成标签。

```python
def mark_message_tag(self, message: str) -> List[Dict[str, Any]]
```
*   **message**: 需要打标的文本内容。
*   **返回**: 标签列表，每个元素包含 `tag` 和 `confidence`。

### 3.5 生命周期回调 (需实现)

#### `enable_module`
模组启用时调用。在此处注册配置和任务，安装依赖。

```python
def enable_module(self) -> bool
```

#### `disable_module`
模组禁用时调用。清理资源。

```python
def disable_module(self) -> bool
```

#### `test_module`
模组自检接口（如测试网络连接）。**必须实现**。

```python
def test_module(self) -> Tuple[bool, str]
```

#### `test_config`
测试配置有效性。在用户保存配置时调用。

```python
def test_config(self, config: Dict[str, Any]) -> Tuple[bool, str]
```
*   **config**: 包含待测试配置的字典。
*   **返回**: `(True, "OK")` 或 `(False, "Error message")`。

#### `execute_schedule_task`
定时任务触发时调用。

```python
def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool
```

## 4. 国际化 (I18n)

在模组目录下创建 `locales` 文件夹，放置 `en_US.json`, `zh_CN.json` 等文件。
Key 命名建议：`module.<module_id>.<category>.<key>`。

## 5. 完整代码示例

```python
from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule
from datetime import datetime
from typing import Tuple, Dict, Any

MODULE_META = {
    "id": "my_news_scraper",
    "name": "示例模组",
    "description": "module.example.desc",
    "version": "1.0.0",
    "author": "DemoUser"
}

class MyModule(BaseModule):

    def enable_module(self) -> bool:
        if not self.install_requirements():
            return False

        self.set_module_config(
            key="target_url",
            description="Target URL",
            value="https://example.com",
            value_type="text",
            regular="^https?://.*"
        )
        self.set_module_schedule_task(
            key="daily_task",
            description="Daily Task",
            cron="0 8 * * *"
        )
        return True

    def test_module(self) -> Tuple[bool, str]:
        # 测试基本连通性
        return True, "Connection OK"

    def test_config(self, config: Dict[str, Any]) -> Tuple[bool, str]:
        # 测试配置有效性
        url = config.get("target_url")
        if not url or "example" not in url:
            return False, "Invalid URL"
        return True, "Config Valid"

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        if task_key == "daily_task":
            url = self.get_module_config("target_url")
            Log.i("MyModule", f"Scraping {url}")
            
            result = {
                "title": "Example News",
                "source": "Example",
                "from_url": "https://example.com/1",
                "content": "Content...",
                "datetime_released": datetime.now().isoformat()
            }
            
            # 自动打标
            tags = self.mark_message_tag(result["content"])
            result["tags"] = tags

            self.save_structured_results(result, fingerprint="https://example.com/1")
            return True
        return False

def create_module(context):
    return MyModule(context)
```
