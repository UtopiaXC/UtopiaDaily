import logging
import os
import sys
import re
import signal
from datetime import datetime


current_file_path = os.path.abspath(__file__)
project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(current_file_path))))
LOG_DIR = os.path.join(project_root, "logs")

# 50M
MAX_LOG_SIZE = 50 * 1024 * 1024

class VersionedFileHandler(logging.FileHandler):
    def __init__(self, log_dir, prefix, max_bytes=MAX_LOG_SIZE):
        self.log_dir = log_dir
        self.prefix = prefix
        self.max_bytes = max_bytes
        self.current_date = datetime.now().strftime("%Y%m%d")
        self.current_index = self._determine_startup_index(self.current_date)
        filename = self._fmt_filename(self.current_date, self.current_index)
        super().__init__(filename, mode='a', encoding='utf-8', delay=False)

    def _determine_startup_index(self, date_str):
        if not os.path.exists(self.log_dir):
            return 1
        pattern = re.compile(rf"^{re.escape(self.prefix)}\.{date_str}\.(\d{{2}})\.log$")
        max_idx = 0
        latest_file = None
        for log_file_name in os.listdir(self.log_dir):
            match = pattern.match(log_file_name)
            if match:
                idx = int(match.group(1))
                if idx > max_idx:
                    max_idx = idx
                    latest_file = os.path.join(self.log_dir, log_file_name)
        if max_idx == 0 or latest_file is None:
            return 1

        try:
            size = os.path.getsize(latest_file)
            if size >= self.max_bytes:
                return max_idx + 1
            else:
                return max_idx
        except OSError:
            return max_idx + 1

    def _fmt_filename(self, date_str, index):
        return os.path.join(self.log_dir, f"{self.prefix}.{date_str}.{index:02d}.log")

    def emit(self, record):
        try:
            if self.stream is None:
                self.stream = self._open()
                
            now_date = datetime.now().strftime("%Y%m%d")
            rollover = False
            if now_date != self.current_date:
                self.current_date = now_date
                self.current_index = 1
                rollover = True
            elif self.max_bytes > 0:
                if self.stream and hasattr(self.stream, 'tell'):
                    try:
                        if self.stream.tell() >= self.max_bytes:
                            self.current_index += 1
                            rollover = True
                    except ValueError:
                        # Stream might be closed
                        self.stream = self._open()
                        
            if rollover:
                self.close()
                self.baseFilename = self._fmt_filename(self.current_date, self.current_index)
                self.stream = self._open()
            super().emit(record)
            self.flush()
        except Exception:
            self.handleError(record)


class Log:
    _logger = None
    _console_handler = None
    _sys_handler = None
    
    @classmethod
    def _ensure_initialized(cls):
        if cls._logger is not None:
            return
        if not os.path.exists(LOG_DIR):
            try:
                os.makedirs(LOG_DIR)
            except Exception:
                pass
        cls._logger = logging.getLogger("UtopiaDaily")
        # Default to DEBUG initially, controlled by handlers
        cls._logger.setLevel(logging.DEBUG)
        
        if cls._logger.hasHandlers():
            cls._logger.handlers.clear()
            
        formatter = logging.Formatter(
            '[%(asctime)s] [%(levelname)s] [PID:%(process)d] %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )

        cls._console_handler = logging.StreamHandler(sys.stdout)
        cls._console_handler.setLevel(logging.DEBUG)
        cls._console_handler.setFormatter(formatter)
        cls._logger.addHandler(cls._console_handler)
        
        try:
            cls._sys_handler = VersionedFileHandler(LOG_DIR, "system", max_bytes=MAX_LOG_SIZE)
            cls._sys_handler.setLevel(logging.INFO)
            cls._sys_handler.setFormatter(formatter)
            cls._logger.addHandler(cls._sys_handler)
            
            error_log_file = os.path.join(LOG_DIR, "error.log")
            err_handler = logging.FileHandler(error_log_file, encoding='utf-8')
            err_handler.setLevel(logging.ERROR)
            err_handler.setFormatter(formatter)
            cls._logger.addHandler(err_handler)

        except Exception as e:
            print(f"Warning: Failed to setup log handlers: {e}")

    @classmethod
    def set_level(cls, level_str):
        """
        Dynamically sets the logging level.
        :param level_str: "DEBUG", "INFO", "WARNING", "ERROR"
        """
        cls._ensure_initialized()
        level_map = {
            "DEBUG": logging.DEBUG,
            "INFO": logging.INFO,
            "WARNING": logging.WARNING,
            "ERROR": logging.ERROR
        }
        level = level_map.get(level_str.upper(), logging.INFO)
        if cls._console_handler:
            cls._console_handler.setLevel(level)
        if cls._sys_handler:
            cls._sys_handler.setLevel(level)
        cls.i("LOGGER", f"Log level changed to {level_str.upper()}")

    @classmethod
    def setup_global_exception_handler(cls):
        """
        Sets up a global exception handler to log uncaught exceptions before exit.
        """
        def handle_exception(exc_type, exc_value, exc_traceback):
            if issubclass(exc_type, KeyboardInterrupt):
                sys.__excepthook__(exc_type, exc_value, exc_traceback)
                return

            cls.e("FATAL", "Uncaught exception caused program exit", error=(exc_type, exc_value, exc_traceback))
            sys.__excepthook__(exc_type, exc_value, exc_traceback)

        sys.excepthook = handle_exception

    @classmethod
    def get_logger(cls):
        cls._ensure_initialized()
        return cls._logger

    @classmethod
    def d(cls, tag, msg):
        cls._ensure_initialized()
        cls._logger.debug(f"[{tag}] {msg}")

    @classmethod
    def i(cls, tag, msg):
        cls._ensure_initialized()
        cls._logger.info(f"[{tag}] {msg}")

    @classmethod
    def w(cls, tag, msg):
        cls._ensure_initialized()
        cls._logger.warning(f"[{tag}] {msg}")

    @classmethod
    def e(cls, tag, msg, error=None, stack_trace=True):
        cls._ensure_initialized()
        if error:
            cls._logger.error(f"[{tag}] {msg}", exc_info=error)
        else:
            cls._logger.error(f"[{tag}] {msg}", exc_info=stack_trace)

    @classmethod
    def fatal(cls, tag, msg, error=None):
        cls._ensure_initialized()
        cls.e(tag, f"FATAL ERROR: {msg} - TERMINATING PROCESS", error=error)
        try:
            for handler in cls._logger.handlers:
                handler.flush()
        except:
            pass
        os.kill(os.getpid(), signal.SIGTERM)
