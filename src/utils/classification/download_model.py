import os
import shutil
import json
from huggingface_hub import hf_hub_download
from src.utils.logger.logger import Log
from src.utils.classification.classification import get_models_dir

TAG = "MODEL_DOWNLOADER"

FILES_TO_DOWNLOAD = [
    "onnx/model_quantized.onnx",
    "tokenizer.json",
    "config.json",
    "special_tokens_map.json",
    "tokenizer_config.json"
]

def get_default_models_json():
    """
    Reads and returns the content of default_models.json
    """
    try:
        current_dir = os.path.dirname(os.path.abspath(__file__))
        json_path = os.path.join(current_dir, "default_models.json")
        
        if not os.path.exists(json_path):
            Log.e(TAG, f"default_models.json not found at {json_path}")
            return []
            
        with open(json_path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        Log.e(TAG, "Failed to read default_models.json", error=e)
        return []

def save_model_info(local_dir: str, repo_id: str):
    """
    Saves model information to info.json in the model directory.
    """
    try:
        # Find model info from default_models.json
        default_models = get_default_models_json()
        model_info = None
        
        for group in default_models:
            for model in group.get("models", []):
                if model.get("hf_repo") == repo_id:
                    model_info = model
                    break
            if model_info:
                break
        
        if not model_info:
            # Fallback if not found in default list
            model_info = {
                "name": repo_id.split("/")[-1],
                "hf_repo": repo_id
            }

        info_path = os.path.join(local_dir, "info.json")
        with open(info_path, 'w', encoding='utf-8') as f:
            json.dump(model_info, f, indent=2, ensure_ascii=False)
            
        Log.i(TAG, f"Saved model info to {info_path}")
    except Exception as e:
        Log.e(TAG, f"Failed to save model info for {repo_id}", error=e)

def download_model_from_hf(repo_id: str):
    """
    Download a model from Hugging Face Hub to the local modules directory.
    """
    try:
        models_dir = get_models_dir()
        # Use the last part of repo_id as folder name, or full name replacing / with _
        # Usually just the name is fine if unique enough, but let's use full name to avoid conflicts
        # Actually, user wants to select from a list, let's use the repo name part.
        model_name = repo_id.split("/")[-1]
        local_dir = os.path.join(models_dir, model_name)
        
        if not os.path.exists(local_dir):
            os.makedirs(local_dir)
        
        Log.i(TAG, f"Starting download for {repo_id} to {local_dir}...")
        
        for filename in FILES_TO_DOWNLOAD:
            Log.i(TAG, f"Downloading {filename}...")
            # Download file
            file_path = hf_hub_download(
                repo_id=repo_id,
                filename=filename,
                local_dir=local_dir,
                local_dir_use_symlinks=False
            )
            
            # Move model_quantized.onnx to root of model dir
            if "model_quantized.onnx" in filename:
                target_path = os.path.join(local_dir, "model_quantized.onnx")
                # hf_hub_download with local_dir preserves directory structure (onnx/model_quantized.onnx)
                # So file_path is .../onnx/model_quantized.onnx
                if file_path != target_path:
                    shutil.move(file_path, target_path)
                    Log.i(TAG, f"Moved {filename} to root")

        # Cleanup empty onnx folder
        onnx_dir = os.path.join(local_dir, "onnx")
        if os.path.exists(onnx_dir) and not os.listdir(onnx_dir):
            os.rmdir(onnx_dir)
        
        # Save info.json
        save_model_info(local_dir, repo_id)
        
        Log.i(TAG, f"Successfully downloaded {repo_id}")
        return True, f"Successfully downloaded {model_name}"
    except Exception as e:
        Log.e(TAG, f"Failed to download model {repo_id}", error=e)
        # Cleanup on failure? Maybe not, partial download might be useful or retriable.
        return False, str(e)
