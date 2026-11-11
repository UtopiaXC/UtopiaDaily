# External Plugin Guide

Utopia Daily discovers external plugins at startup and on dashboard reload from two subdirectories:

- `plugins/scrapers/<plugin_id>/` for scraper plugins where `kind` is `"scraper"`
- `plugins/push/<plugin_id>/` for push plugins where `kind` is `"push"`

## Required Directory Layout

```text
plugins/scrapers/<plugin_id>/
├── manifest.json
├── main.py
├── requirements.txt
└── locales/
    ├── en_US.json
    └── zh_CN.json
```

Only `manifest.json` and the entrypoint referenced by `command` are mandatory.

## Manifest Schema

```json
{
  "id": "example_plugin",
  "kind": "scraper",
  "version": "0.2.0",
  "author": "Author",
  "command": ["python3", "main.py"],
  "configs": [
    {
      "key": "limit",
      "type": "int",
      "default_value": "10",
      "sort_order": 1
    }
  ],
  "tasks": [
    {
      "key": "fetch_top",
      "default_cron": "0 */2 * * *",
      "default_enabled": false
    }
  ]
}
```

Supported config field types are `string`, `int`, `bool`, `select`, `password`, and `array`.

## Stdio Execution Contract

- **stdin**: Single JSON object `{ "task_id": "<task_key>", "config": { ... } }`.
  - During connection testing from the dashboard, `task_id` is `"test"`.
  - For push plugins, `config` includes `recipient`, `title`, and `articles`.
- **stdout**: Single JSON object `{ "status": "success" | "failed", "data": [...], "message": "..." }`.
- **stderr**: Forwarded to core structured logs without affecting JSON output.
- **exit code**: `0` indicates process success; non-zero marks execution as failed.

## Python Dependency Isolation

When `requirements.txt` exists inside a Python plugin directory, triggering environment setup in the dashboard or enabling/running the plugin runs `uv venv .venv` and `uv pip install --python .venv/bin/python -r requirements.txt`. Subsequent executions automatically resolve `python3` commands through `.venv/bin/python`.
