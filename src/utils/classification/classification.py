import os
import json
from src.utils.logger.logger import Log
from src.database.connection import system_session_scope
from src.database.models import SystemConfig

TAG = "CLASSIFICATION_MANAGER"

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
MODELS_DIR = os.path.join(PROJECT_ROOT, "models")

def ensure_models_dir():
    if not os.path.exists(MODELS_DIR):
        try:
            os.makedirs(MODELS_DIR)
            Log.i(TAG, f"Created models directory at: {MODELS_DIR}")
        except Exception as e:
            Log.e(TAG, f"Failed to create models directory: {e}")

def list_local_models():
    ensure_models_dir()
    models = [{"label": "None", "value": "None"}] # Default option
    if not os.path.exists(MODELS_DIR):
        return models
        
    for item in os.listdir(MODELS_DIR):
        item_path = os.path.join(MODELS_DIR, item)
        if os.path.isdir(item_path):
            # Check if it looks like a model directory (e.g., has config.json or .onnx files)
            # For now, we just list all directories
            
            # Try to read info.json for better display name
            info_path = os.path.join(item_path, "info.json")
            label = item
            if os.path.exists(info_path):
                try:
                    with open(info_path, 'r', encoding='utf-8') as f:
                        info = json.load(f)
                        if "name" in info:
                            label = info["name"]
                except Exception:
                    pass
            
            models.append({"label": label, "value": item})
    return models

def get_models_dir():
    return MODELS_DIR

def check_current_model():
    """
    Check if the currently configured model exists. If not, revert to 'None'.
    """
    ensure_models_dir()
    with system_session_scope() as session:
        config = session.query(SystemConfig).filter_by(key="ai_model").first()
        if config and config.value and config.value != "None":
            model_path = os.path.join(MODELS_DIR, config.value)
            if not os.path.exists(model_path):
                Log.w(TAG, f"Configured model '{config.value}' not found at {model_path}. Reverting to 'None'.")
                config.value = "None"
                # We don't need to commit explicitly as the context manager handles it, 
                # but for clarity or if exception happens:
                # session.commit() is handled by system_session_scope on exit
