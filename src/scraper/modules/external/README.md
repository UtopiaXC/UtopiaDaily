# 外部模组开发指南

**警告：本文档由AI（Gemini）生成，只用作暂时参考。正式文档稍后提供。**  
本文档介绍如何为 Utopia-Daily 开发外部抓取模组。

## 1. 目录结构

在 `src/scraper/modules/external/` 目录下创建一个新的文件夹，文件夹名称即为**模组ID**。
在该文件夹内，必须包含一个 `api.py` 文件作为入口。

示例结构：
```text
src/scraper/modules/external/
└── my_news_scraper/       <-- 模组ID
    ├── api.py             <-- 核心入口文件 (必须)
    ├── scraper.py         <-- 业务逻辑 (推荐分离)
    └── requirements.txt   <-- 依赖说明
```

## 2. 开发规范 (api.py)

`api.py` 需要包含以下三个部分：

### 2.1 模组元数据

定义一个名为 `MODULE_META` 的字典，包含模组的基本信息。

```python
MODULE_META = {
    "name": "我的新闻抓取器",
    "description": "抓取某某网站的新闻",
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

### 3.1 配置管理

#### `set_module_config`
注册或更新模块的配置项。通常在 `enable_module` 中调用。

```python
def set_module_config(self, key: str, description: str, value: str, force_init: bool = False, hint: str = "", regular: str = "")
```
*   **key**: 配置项的唯一标识符。
*   **description**: 在后台显示的配置描述。
*   **value**: 默认值。
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

### 3.2 任务调度

#### `set_module_schedule_task`
注册一个定时任务。通常在 `enable_module` 中调用。

```python
def set_module_schedule_task(self, key: str, description: str, cron: str, force_init: bool = False, hint: str = "", regular: str = "")
```
*   **key**: 任务的唯一标识符。
*   **description**: 任务描述。
*   **cron**: Cron 表达式 (格式: `分 时 日 月 周`)。例如 `0 8 * * *` 表示每天 08:00。
*   **force_init**: 是否强制更新任务描述。

#### `get_module_schedule_task`
获取任务的 Cron 表达式。

```python
def get_module_schedule_task(self, key: str) -> str
```

### 3.3 数据处理

#### `save_structured_results`
将抓取结果保存到系统数据库。

```python
def save_structured_results(self, value: Any) -> Dict[str, Any]
```
*   **value**: 要保存的数据字典。建议包含 `title`, `content`, `date`, `source` 等字段。
*   **返回**: 保存结果的状态字典。

#### `mark_message_tag`
调用系统 AI 模型为文本生成标签。

```python
def mark_message_tag(self, message: str) -> List[Dict[str, Any]]
```
*   **message**: 需要打标的文本内容。
*   **返回**: 标签列表，每个元素包含 `tag` 和 `confidence`。

### 3.4 生命周期回调 (需实现)

#### `enable_module`
模组启用时被调用。请在此处注册配置和任务。

```python
def enable_module(self) -> bool
```
*   **返回**: True 表示启用成功，False 表示失败。

#### `disable_module`
模组禁用时被调用。用于清理资源。

```python
def disable_module(self) -> bool
```

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

## 4. 完整代码示例

```python
from src.utils.logger.logger import Log
from src.scraper.modules.base_module import BaseModule
from datetime import datetime

MODULE_META = {
    "name": "示例模组",
    "description": "这是一个演示模组",
    "version": "1.0.0",
    "author": "DemoUser"
}

class MyModule(BaseModule):
    
    def enable_module(self) -> bool:
        self.set_module_config(
            key="target_url",
            description="目标网站地址",
            value="https://example.com",
            hint="请输入完整的URL"
        )
        self.set_module_schedule_task(
            key="daily_task",
            description="每日抓取任务",
            cron="0 8 * * *"
        )
        return True

    def execute_schedule_task(self, cron: str, task_key: str, timestamp: datetime) -> bool:
        if task_key == "daily_task":
            url = self.get_module_config("target_url")
            Log.i("MyModule", f"开始抓取: {url}")
            
            # 模拟抓取数据
            result = {
                "title": "示例新闻",
                "content": "这是内容...",
                "source": "Example",
                "date": str(timestamp)
            }
            
            # 自动打标
            tags = self.mark_message_tag(result["content"])
            result["tags"] = tags
            
            # 保存
            self.save_structured_results(result)
            return True
        return False

def create_module(context):
    return MyModule(context)
```
