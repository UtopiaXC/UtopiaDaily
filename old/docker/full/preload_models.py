# docker/preload_models.py
import os
import shutil
import json
from huggingface_hub import hf_hub_download

# 目标下载目录（构建时的临时目录）
BASE_DIR = "/app/models"

# 只需要这 4 个特定模型
TARGET_MODELS = [
    {
        "name": "E5-Small Multilingual",
        "repo": "Xenova/multilingual-e5-small"
    },
    {
        "name": "BGE-Small ZH v1.5",
        "repo": "Xenova/bge-small-zh-v1.5"
    },
    {
        "name": "MiniLM-L6 v2",
        "repo": "Xenova/all-MiniLM-L6-v2"
    },
    {
        "name": "MiniLM-L3 v2",
        "repo": "Xenova/paraphrase-MiniLM-L3-v2"
    }
]

FILES_TO_DOWNLOAD = [
    "onnx/model_quantized.onnx",
    "tokenizer.json",
    "config.json",
    "special_tokens_map.json",
    "tokenizer_config.json"
]


def preload():
    if not os.path.exists(BASE_DIR):
        os.makedirs(BASE_DIR)

    print(f"--- Starting Preload of {len(TARGET_MODELS)} Models ---")

    for model in TARGET_MODELS:
        repo_id = model["repo"]
        folder_name = repo_id.split("/")[-1]
        local_dir = os.path.join(BASE_DIR, folder_name)

        print(f"Downloading: {model['name']} ({repo_id}) -> {local_dir}")

        for filename in FILES_TO_DOWNLOAD:
            try:
                file_path = hf_hub_download(
                    repo_id=repo_id,
                    filename=filename,
                    local_dir=local_dir
                )

                if "model_quantized.onnx" in filename:
                    target_path = os.path.join(local_dir, "model_quantized.onnx")
                    if file_path != target_path:
                        shutil.move(file_path, target_path)
                        print(f"  Moved onnx to root")
            except Exception as e:
                print(f"  Warning: Failed to download {filename}: {e}")

        onnx_dir = os.path.join(local_dir, "onnx")
        if os.path.exists(onnx_dir) and not os.listdir(onnx_dir):
            os.rmdir(onnx_dir)

        info_path = os.path.join(local_dir, "info.json")
        info_data = {
            "name": model["name"],
            "hf_repo": repo_id
        }
        with open(info_path, 'w', encoding='utf-8') as f:
            json.dump(info_data, f, indent=2, ensure_ascii=False)
        print(f"  Generated info.json")

    print("--- Preload Complete ---")


if __name__ == "__main__":
    preload()