import uvicorn
import logging
from src.utils.logger.logger import Log

TAG = "WEB_SERVER"

class TagInjectionFilter(logging.Filter):
    def __init__(self, tag_prefix):
        super().__init__()
        self.tag_prefix = tag_prefix
    
    def filter(self, record):
        if not record.msg.startswith(f"[{self.tag_prefix}]"):
            record.msg = f"[{self.tag_prefix}] {record.msg}"
        return True

def run_web_server(host="0.0.0.0", port=8000):
    Log.i(TAG, f"Starting Web Server on {host}:{port}")
    our_logger = Log.get_logger()
    WEB_SERVER_TAG = "WEB_SERVER_CORE"

    for logger_name in ["uvicorn", "uvicorn.error", "uvicorn.access"]:
        logger = logging.getLogger(logger_name)
        logger.filters = []
        logger.addFilter(TagInjectionFilter(WEB_SERVER_TAG))
        logger.handlers = our_logger.handlers
        logger.propagate = False
        logger.setLevel(logging.INFO)

    try:
        uvicorn.run(
            "src.web.entry:app",
            host=host,
            port=port,
            log_level="info",
            reload=False,
            log_config=None
        )
    except Exception as e:
        Log.e(TAG, "Server crashed", error=e)
