import os
import json
import re
import hashlib
import numpy as np
from tokenizers import Tokenizer
import onnxruntime as ort
from src.utils.logger.logger import Log
from src.database.connection import system_session_scope
from src.database.models import SystemConfig

TAG = "CLASSIFICATION_MANAGER"

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
MODELS_DIR = os.path.join(PROJECT_ROOT, "models")

# Cache loaded models to avoid reloading: {model_name: (session, tokenizer)}
_loaded_models = {}

# Cache tag embeddings: {(model_name, config_hash): (phrase_embeddings, phrase_meta)}
# phrase_meta stores list of dicts: {'tag_index': int, 'is_keyword': bool}
_tag_embeddings_cache = {}


def ensure_models_dir():
    if not os.path.exists(MODELS_DIR):
        try:
            os.makedirs(MODELS_DIR)
            Log.i(TAG, f"Created models directory at: {MODELS_DIR}")
        except Exception as e:
            Log.e(TAG, f"Failed to create models directory: {e}")


def list_local_models():
    ensure_models_dir()
    models = [{"label": "None", "value": "None"}]
    if not os.path.exists(MODELS_DIR):
        return models

    for item in os.listdir(MODELS_DIR):
        item_path = os.path.join(MODELS_DIR, item)
        if os.path.isdir(item_path):
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
    ensure_models_dir()
    with system_session_scope() as session:
        config = session.query(SystemConfig).filter_by(key="ai_model").first()
        if config and config.value and config.value != "None":
            model_path = os.path.join(MODELS_DIR, config.value)
            if not os.path.exists(model_path):
                Log.w(TAG, f"Configured model '{config.value}' not found at {model_path}. Reverting to 'None'.")
                config.value = "None"


def get_model(model_name):
    if model_name in _loaded_models:
        return _loaded_models[model_name]

    model_dir = os.path.join(MODELS_DIR, model_name)
    if not os.path.exists(model_dir):
        raise ValueError(f"Model {model_name} not found at {model_dir}")

    onnx_path = os.path.join(model_dir, "model_quantized.onnx")
    if not os.path.exists(onnx_path):
        onnx_path = os.path.join(model_dir, "model.onnx")
        if not os.path.exists(onnx_path):
            raise ValueError(f"ONNX model file not found in {model_dir}")

    tokenizer_path = os.path.join(model_dir, "tokenizer.json")
    if not os.path.exists(tokenizer_path):
        raise ValueError(f"Tokenizer file not found in {model_dir}")

    Log.i(TAG, f"Loading model: {model_name}...")
    try:
        tokenizer = Tokenizer.from_file(tokenizer_path)

        # [OPTIMIZATION] Dynamic Padding
        # Removed length=512. Now it pads to the longest sequence in the batch.
        # This significantly speeds up inference for short texts.
        tokenizer.enable_padding(pad_id=0, pad_token="[PAD]")
        tokenizer.enable_truncation(max_length=512)

        sess_options = ort.SessionOptions()
        sess_options.log_severity_level = 3
        # Optional: Set intra_op_num_threads to 1-4 depending on CPU to avoid blocking
        # sess_options.intra_op_num_threads = 2

        session = ort.InferenceSession(onnx_path, sess_options)

        _loaded_models[model_name] = (session, tokenizer)
        Log.i(TAG, f"Model {model_name} loaded successfully.")
        return session, tokenizer
    except Exception as e:
        Log.e(TAG, f"Failed to load model {model_name}", error=e)
        raise


def mean_pooling(model_output, attention_mask):
    token_embeddings = model_output[0]
    input_mask_expanded = np.expand_dims(attention_mask, -1)
    input_mask_expanded = np.broadcast_to(input_mask_expanded, token_embeddings.shape)

    sum_embeddings = np.sum(token_embeddings * input_mask_expanded, axis=1)
    sum_mask = np.clip(np.sum(input_mask_expanded, axis=1), a_min=1e-9, a_max=None)

    return sum_embeddings / sum_mask


def normalize(v):
    norm = np.linalg.norm(v, axis=1, keepdims=True)
    return v / np.clip(norm, a_min=1e-9, a_max=None)


def encode(texts, session, tokenizer):
    if isinstance(texts, str):
        texts = [texts]
    if not texts:
        return np.array([])

    # With dynamic padding enabled in get_model, this is much faster
    encoded = tokenizer.encode_batch(texts)

    input_ids = np.array([e.ids for e in encoded], dtype=np.int64)
    attention_mask = np.array([e.attention_mask for e in encoded], dtype=np.int64)

    input_names = [i.name for i in session.get_inputs()]
    inputs = {
        'input_ids': input_ids,
        'attention_mask': attention_mask
    }

    if 'token_type_ids' in input_names:
        token_type_ids = np.array([e.type_ids for e in encoded], dtype=np.int64)
        inputs['token_type_ids'] = token_type_ids

    outputs = session.run(None, inputs)

    embeddings = mean_pooling(outputs, attention_mask)
    embeddings = normalize(embeddings)

    return embeddings


def cos_sim(a, b):
    # a: (n_segments, dim), b: (n_tags, dim) -> (n_segments, n_tags)
    return np.dot(a, b.T)


def clean_text(text):
    if not text:
        return ""
    text = re.sub(r'!\[.*?\]\(.*?\)', '', text)
    text = re.sub(r'\[(.*?)\]\(.*?\)', r'\1', text)
    text = re.sub(r'<[^>]+>', '', text)
    text = re.sub(r'http[s]?://\S+', '', text)
    text = re.sub(r'\s+', ' ', text).strip()
    text = re.sub(r'（.*?）$', '', text)
    text = re.sub(r'\(.*?\)$', '', text)
    return text


def split_text_smart(text, max_chars=800, max_segments=6):
    """
    [OPTIMIZATION] Smart Splitting
    Instead of splitting by every sentence (which kills context), this merges sentences
    into larger chunks up to max_chars to utilize the full transformer window.
    """
    if not text:
        return []

    # 1. Split by strong delimiters to get raw sentences
    raw_segments = re.split(r'(?<=[.?!。？！\n])\s*', text)

    merged_segments = []
    current_chunk = ""

    for seg in raw_segments:
        seg = seg.strip()
        if not seg: continue

        # Check if adding this segment exceeds our soft limit
        if len(current_chunk) + len(seg) < max_chars:
            current_chunk += " " + seg
        else:
            if current_chunk:
                merged_segments.append(current_chunk.strip())
            current_chunk = seg

    if current_chunk:
        merged_segments.append(current_chunk.strip())

    # Limit total segments to avoid processing extremely large documents forever
    # We take the start, middle, and end if there are too many
    if len(merged_segments) > max_segments:
        half = max_segments // 2
        return merged_segments[:half] + merged_segments[-half:]

    return merged_segments


def get_config_hash(tags_config):
    try:
        config_str = json.dumps(tags_config, sort_keys=True)
        return hashlib.md5(config_str.encode('utf-8')).hexdigest()
    except Exception:
        return None


def classify_text(text, model_name, tags_config, confidence_threshold=60):
    Log.d(TAG, f"Classifying text length: {len(text) if text else 0}")

    if not text:
        return []
    if not model_name or model_name == "None":
        Log.w(TAG, "No model selected.")
        return []

    cleaned_text = clean_text(text)
    if not cleaned_text:
        return []

    # Use Smart Split
    segments = split_text_smart(cleaned_text)
    if not segments:
        return []

    Log.d(TAG, f"Text split into {len(segments)} contextual chunks.")

    try:
        session, tokenizer = get_model(model_name)
    except Exception as e:
        Log.e(TAG, f"Could not get model {model_name}: {e}")
        return []

    # --- Tag Embedding & Caching Strategy ---
    config_hash = get_config_hash(tags_config)
    cache_key = (model_name, config_hash)

    phrase_embeddings = None
    phrase_meta = []  # Stores: [{"tag_idx": 0, "name": "..."}]

    if config_hash and cache_key in _tag_embeddings_cache:
        # Hit Cache
        phrase_embeddings, phrase_meta = _tag_embeddings_cache[cache_key]
    else:
        # Miss Cache - Build Embeddings
        Log.d(TAG, "Computing new tag embeddings...")
        if "tags" not in tags_config:
            return []

        all_phrases = []

        for i, tag in enumerate(tags_config["tags"]):
            name = tag.get("name")
            keywords = tag.get("keywords", [])

            if name:
                all_phrases.append(name)
                phrase_meta.append({"tag_idx": i, "is_name": True})

            for kw in keywords:
                if kw:
                    all_phrases.append(kw)
                    phrase_meta.append({"tag_idx": i, "is_name": False})

        if not all_phrases:
            return []

        try:
            phrase_embeddings = encode(all_phrases, session, tokenizer)
            if config_hash:
                _tag_embeddings_cache[cache_key] = (phrase_embeddings, phrase_meta)
        except Exception as e:
            Log.e(TAG, "Encoding phrases failed", error=e)
            return []

    # --- Inference ---
    try:
        segment_embeddings = encode(segments, session, tokenizer)
    except Exception as e:
        Log.e(TAG, "Encoding text segments failed", error=e)
        return []

    # Similarity: (num_segments, num_phrases)
    similarity_matrix = cos_sim(segment_embeddings, phrase_embeddings)

    # Max similarity per phrase across all text segments
    # phrase_scores: (num_phrases,)
    phrase_scores = np.max(similarity_matrix, axis=0)

    # --- Aggregation ---
    tag_scores = {}  # tag_index -> max_score

    for idx, score in enumerate(phrase_scores):
        meta = phrase_meta[idx]
        tag_idx = meta["tag_idx"]

        # Here you could add logic: if meta["is_name"] is True, boost score slightly
        # For now, we use raw max similarity

        if tag_idx not in tag_scores:
            tag_scores[tag_idx] = score
        else:
            tag_scores[tag_idx] = max(tag_scores[tag_idx], score)

    # --- Filtering ---
    threshold = confidence_threshold / 100.0
    candidates = []

    for tag_idx, score in tag_scores.items():
        if score >= threshold:
            tag_info = tags_config["tags"][tag_idx]
            candidates.append({
                "score": float(score),
                "tag": tag_info.get("name"),
                "group": tag_info.get("group"),
                "priority": tag_info.get("priority", 0)
            })

    # Sort by score desc
    candidates.sort(key=lambda x: x["score"], reverse=True)

    # --- Group Limits ---
    group_limits = {g["name"]: g.get("max_results", 1) for g in tags_config.get("groups", [])}
    group_counts = {}

    final_results = []
    for cand in candidates:
        group = cand["group"]

        if group:
            limit = group_limits.get(group, 1)
            current_count = group_counts.get(group, 0)
            if current_count >= limit:
                continue
            group_counts[group] = current_count + 1

        final_results.append({
            "tag": cand["tag"],
            "confidence": round(cand["score"], 4)
        })

    Log.i(TAG, f"Classification finished. Found {len(final_results)} tags.")
    return final_results