# 外部模组开发指南

**警告：本文档由AI（Gemini）生成，只用作暂时参考。正式文档稍后提供。**  
本文档介绍如何为 Utopia-Daily 开发外部抓取模组。

## 1. 目录结构

在 `src/scraper/modules/external/` 目录下创建一个新的文件夹，文件夹名称即为**模组ID**。
在该文件夹内，必须包含一个 `controller.py` 文件作为入口（ModuleManager 会直接加载此文件）。

示例结构：
```text
src/scraper/modules/external/
└── my_news_scraper/       <-- 模组ID
    ├── controller.py      <-- 核心入口文件 (必须)
    ├── scraper.py         <-- 业务逻辑 (非必须，建议将业务逻辑从controller中剥离)
    ├── locales/           <-- 国际化资源目录 (可选，推荐使用)
    │   ├── en_US.json
    │   └── zh_CN.json
    └── requirements.txt   <-- 依赖文件 (可选，用于定义模组特有的第三方包，尽量不要使用)
```

## 2. 开发规范 (controller.py)

`controller.py` 需要包含以下三个部分：

### 2.1 模组元数据

定义一个名为 `MODULE_META` 的字典，包含模组的基本信息。

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

创建一个类，继承自 `src.scraper.modules.base_module.BaseModule`，并实现必要的方法。

### 2.3 入口函数

定义 `create_module(context)` 函数，返回模组类的实例。

```python
def create_module(context):
    return MyModule(context)
```

## 3. API 参考

模组类继承自 `BaseModule`，可以使用以下方法与系统交互。

### 3.1 依赖管理 (新增)

#### `install_requirements`
如果模组依赖于系统未提供的第三方 Python 包，请在 `enable_module` 或 `test_module` 中调用此方法。系统会自动读取模组目录下的 `requirements.txt` 并将依赖安装到模组专属的 `libs` 目录中，避免污染全局环境。

```python
def install_requirements(self, requirements_file: str = "requirements.txt") -> bool
```
*   **requirements_file**: 依赖文件名，默认为 "requirements.txt"。
*   **返回**: True 表示安装成功，False 表示失败。

### 3.2 配置管理

#### `set_module_config`
注册或更新模块的配置项。通常在 `enable_module` 中调用。

```python
def set_module_config(self, key: str, description: str, value: str, value_type: str = "string", force_init: bool = False, hint: str = "", regular: str = "")
```
*   **key**: 配置项的唯一标识符。
*   **description**: 在后台显示的配置描述。**推荐使用 I18n Key**。
*   **value**: 默认值。
*   **value_type**: 值类型（如 string / password），默认 "string"。
*   **force_init**: 是否强制覆盖已存在的配置描述（通常为 False）。
*   **hint**: 输入框的提示信息。
*   **regular**: 用于验证输入的正则表达式（可选）。

#### `get_module_config`
获取当前配置的值。

```python
def get_module_config(self, key: str) -> str
```
*   **key**: 配置项标识符。
*   **返回**: 配置值的字符串，如果不存在则返回 None。

#### `drop_module_config`
删除一个配置项。

```python
def drop_module_config(self, key: str)
```

### 3.3 任务调度

#### `set_module_schedule_task`
注册一个定时任务。通常在 `enable_module` 中调用。

```python
def set_module_schedule_task(self, key: str, description: str, cron: str = "", force_init: bool = False)
```
*   **key**: 任务的唯一标识符。
*   **description**: 任务描述。**推荐使用 I18n Key**。
*   **cron**: Cron 表达式 (格式: `分 时 日 月 周`)。例如 `0 8 * * *` 表示每天 08:00。
*   **force_init**: 是否强制更新任务描述。

#### `drop_module_schedule_task`
删除一个定时任务。

```python
def drop_module_schedule_task(self, key: str)
```

#### `get_module_schedule_task`
获取任务的 Cron 表达式。

```python
def get_module_schedule_task(self, key: str) -> str
```

### 3.4 数据处理

#### `save_structured_results`
将抓取结果保存到系统数据库。

```python
def save_structured_results(self, value: Any, fingerprint: str = "") -> Dict[str, Any]
```
*   **value**: 要保存的数据字典。推荐结构如下（请按实际业务填充，尽量使用来源提供的发布时间，而非抓取时间）：

```json
{
    "title": "",
    "summary": "",
    "source": "",
    "from_url": "",
    "content": "",
    "content_type": "",
    "datetime_released": "",
    "quotation": {
        "content": "",
        "content_type": "",
        "from": "",
        "datetime_released": ""
    },
    "tags": [],
    "metadata": {}
}

```

*   **fingerprint**: 可选去重指纹（如唯一 URL、摘要哈希）。未提供 fingerprint 时，请避免把“抓取时间”写入 `value`，否则易产生重复写入；系统会自动记录 `created_at`，无需自行维护抓取时间。
*   **返回**: 保存结果的状态字典。

#### `mark_message_tag`
调用系统 AI 模型为文本生成标签。

```python
def mark_message_tag(self, message: str) -> List[Dict[str, Any]]
```
*   **message**: 需要打标的文本内容。
*   **返回**: 标签列表，每个元素包含 `tag` 和 `confidence`。

### 3.5 生命周期回调 (需实现)

#### `enable_module`
模组启用时被调用。请在此处注册配置和任务。**如果模组有第三方依赖，建议在此处调用 `install_requirements`。**

```python
def enable_module(self) -> bool
```
*   **返回**: True 表示启用成功，False 表示失败。

#### `disable_module`
模组禁用时被调用。用于清理资源。

```python
def disable_module(self) -> bool
```

#### `test_module`
模组自检接口。用于验证模组可用性（如网络连接、API 密钥有效性）。必须实现，启用模组会调用此方法检查模组可用性，如果没有或检查失败则不允许启用模组。**建议在此处也调用 `install_requirements` 以确保测试环境完整。**

```python
def test_module(self) -> Tuple[bool, str]
```
*   **返回**: 一个元组 `(success, message)`。
    *   `success`: True 表示测试通过，False 表示失败。
    *   `message`: 成功或失败的详细描述信息。

#### `execute_schedule_task`
定时任务触发时被调用。

```python
def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool
```
*   **cron**: 触发该任务的 Cron 表达式。
*   **task_key**: 触发的任务标识符。
*   **timestamp**: 任务触发的时间戳。
*   **返回**: True 表示执行成功。

#### `generate_html` / `generate_markdown` (可选)
自定义数据的渲染格式。如果不实现，系统将使用默认的 JSON 渲染。

```python
def generate_html(self, value: Any) -> Optional[str]
def generate_markdown(self, value: Any) -> Optional[str]
```

## 4. 国际化 (I18n) 支持

系统支持模块化的国际化配置。

### 4.1 目录结构
在模组目录下创建 `locales` 文件夹，并放置 JSON 格式的语言文件。
```text
my_module/
├── locales/
│   ├── en_US.json
│   └── zh_CN.json
```

### 4.2 Key 命名规范
为了避免冲突，建议使用以下命名规范：
`module.<module_id>.<category>.<key>`

示例 (`en_US.json`):
```json
{
    "module.my_news.desc": "My News Scraper",
    "module.my_news.config.url.desc": "Target Website URL",
    "module.my_news.task.daily.desc": "Daily Fetch Task"
}
```

### 4.3 使用方法
在 `MODULE_META`、`set_module_config` 和 `set_module_schedule_task` 的描述字段中，直接填写 Key 即可。系统会在显示时自动翻译。

## 5. 完整代码示例

```python
from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule
from datetime import datetime
from typing import Tuple

MODULE_META = {
    "id": "my_news_scraper",
    "name": "示例模组",
    "description": "module.example.desc",  # 使用 I18n Key
    "version": "1.0.0",
    "author": "DemoUser"
}


class MyModule(BaseModule):

    def enable_module(self) -> bool:
        # 1. 安装依赖 (如果存在 requirements.txt)
        if not self._context.install_requirements("requirements.txt"):
            Log.e("MyModule", "Failed to install dependencies")
            return False

        # 2. 注册配置
        self.set_module_config(
            key="target_url",
            description="module.example.config.url.desc",  # 使用 I18n Key
            value="https://example.com",
            hint="请输入完整的URL"
        )
        self.set_module_schedule_task(
            key="daily_task",
            description="module.example.task.daily.desc",  # 使用 I18n Key
            cron="0 8 * * *"
        )
        return True

    def test_module(self) -> Tuple[bool, str]:
        # 在此处测试抓取是否正常。请不要在此处调用get config，get task等方法，未enable的情况下数据不会写入，请只测试功能访问性
        return False, f"Fail to get connection to xxx"

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        if task_key == "daily_task":
            url = self.get_module_config("target_url")
            Log.i("MyModule", f"开始抓取: {url}")

            result_from = "test.com/test/12345"
            # 模拟抓取数据
            result = {
                "title": "示例新闻",
                "source": "Example",
                "from_url": result_from,
                "content": "这是内容...",
                "content_type": "text", # 可选text，html，markdown
                "datetime_released": "2024-01-01T08:00:00Z",  # 使用来源时间，避免用抓取时间
                "quotation": {
                    "content": "引用片段...",
                    "content_type": "text",
                    "from": "Example",
                    "datetime_released": "2024-01-01T08:00:00Z"
                },
                "tags": [],
                "metadata": {}
            }

            # 自动打标
            tags = self.mark_message_tag(result["content"])
            result["tags"] = tags

            # 保存（使用 URL 作为指纹防止重复，或可使用摘要哈希等）
            self.save_structured_results(result, fingerprint=result_from)
            return True
        return False


def create_module(context):
    return MyModule(context)
```
