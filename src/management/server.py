import uvicorn
import logging
from src.utils.logger.logger import Log

TAG = "MANAGEMENT_SERVER"

class TagInjectionFilter(logging.Filter):
    def __init__(self, tag_prefix):
        super().__init__()
        self.tag_prefix = tag_prefix
    
    def filter(self, record):
        # Avoid double tagging if it's already tagged
        if not record.msg.startswith(f"[{self.tag_prefix}]"):
            record.msg = f"[{self.tag_prefix}] {record.msg}"
        return True

def run_management_server(host="0.0.0.0", port=8000):
    """
    Starts the FastAPI server using Uvicorn.
    This function is intended to be run in a separate process.
    """
    Log.i(TAG, f"Starting Management Server on {host}:{port}")
    
    our_logger = Log.get_logger()
    
    # Unified TAG for all uvicorn logs
    WEB_SERVER_TAG = "MANAGEMENT_WEB_SERVER"

    for logger_name in ["uvicorn", "uvicorn.error", "uvicorn.access"]:
        logger = logging.getLogger(logger_name)
        
        # Clear existing filters to be safe and add our injection filter
        logger.filters = []
        logger.addFilter(TagInjectionFilter(WEB_SERVER_TAG))
        
        # Replace handlers with our system logger's handlers
        logger.handlers = our_logger.handlers
        logger.propagate = False
        
        # Ensure levels are appropriate.
        logger.setLevel(logging.INFO)

    try:
        uvicorn.run(
            "src.management.backend_entry:app",
            host=host,
            port=port,
            log_level="info",
            reload=False,
            log_config=None
        )
    except Exception as e:
        Log.e(TAG, "Server crashed", error=e)
