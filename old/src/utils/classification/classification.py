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

_loaded_models = {}
_tag_embeddings_cache = {}

PRIORITY_WEIGHT = 0.0025
KEYWORD_HIT_BONUS = 0.03
KEYWORD_BONUS_THRESHOLD = 0.75


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
        tokenizer.enable_padding(pad_id=0, pad_token="[PAD]")
        tokenizer.enable_truncation(max_length=512)
        sess_options = ort.SessionOptions()
        sess_options.log_severity_level = 3
        session = ort.InferenceSession(onnx_path, sess_options)
        _loaded_models[model_name] = (session, tokenizer)
        Log.i(TAG, f"Model {model_name} loaded successfully.")
        return session, tokenizer
    except Exception as e:
        Log.e(TAG, f"Failed to load model {model_name}", error=e)
        raise


def get_pooling_method(model_name):
    name = model_name.lower()
    cls_keywords = ["bge", "labse", "snowflake", "clip", "gte"]
    for kw in cls_keywords:
        if kw in name:
            return "cls"
    return "mean"


def cls_pooling(model_output):
    return model_output[0][:, 0]


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


def encode(texts, session, tokenizer, pooling_method='mean'):
    if isinstance(texts, str):
        texts = [texts]
    if not texts:
        return np.array([])

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
    if pooling_method == 'cls':
        embeddings = cls_pooling(outputs)
    else:
        embeddings = mean_pooling(outputs, attention_mask)

    embeddings = normalize(embeddings)
    return embeddings


def cos_sim(a, b):
    return np.dot(a, b.T)


def clean_text(text):
    if not text: return ""
    text = re.sub(r'!\[.*?\]\(.*?\)', '', text)
    text = re.sub(r'\[(.*?)\]\(.*?\)', r'\1', text)
    text = re.sub(r'<[^>]+>', '', text)
    text = re.sub(r'http[s]?://\S+', '', text)
    text = re.sub(r'\s+', ' ', text).strip()
    return text


def split_text_smart(text, max_chars=800, max_segments=6):
    if not text: return []
    raw_segments = re.split(r'(?<=[.?!。？！\n])\s*', text)

    merged_segments = []
    current_chunk = ""

    for seg in raw_segments:
        seg = seg.strip()
        if not seg: continue
        if len(current_chunk) + len(seg) < max_chars:
            current_chunk += " " + seg
        else:
            if current_chunk: merged_segments.append(current_chunk.strip())
            current_chunk = seg

    if current_chunk: merged_segments.append(current_chunk.strip())

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

    if not text: return []
    if not model_name or model_name == "None":
        Log.w(TAG, "No model selected.")
        return []

    cleaned_text = clean_text(text)
    if not cleaned_text: return []

    # Use Smart Split
    segments = split_text_smart(cleaned_text)
    if not segments: return []

    text_lower = cleaned_text.lower()  # For keyword matching later

    try:
        session, tokenizer = get_model(model_name)
    except Exception as e:
        Log.e(TAG, f"Could not get model {model_name}: {e}")
        return []

    pooling_method = get_pooling_method(model_name)
    Log.d(TAG, f"Using pooling method: {pooling_method} for model: {model_name}")

    config_hash = get_config_hash(tags_config)
    cache_key = (model_name, config_hash)

    phrase_embeddings = None
    phrase_meta = []

    if config_hash and cache_key in _tag_embeddings_cache:
        phrase_embeddings, phrase_meta = _tag_embeddings_cache[cache_key]
    else:
        if "tags" not in tags_config: return []
        all_phrases = []

        for i, tag in enumerate(tags_config["tags"]):
            name = tag.get("name")
            keywords = tag.get("keywords", [])

            if name:
                all_phrases.append(name)
                phrase_meta.append({"tag_idx": i})

            for kw in keywords:
                if kw:
                    all_phrases.append(kw)
                    phrase_meta.append({"tag_idx": i})

        if not all_phrases: return []

        try:
            phrase_embeddings = encode(all_phrases, session, tokenizer, pooling_method=pooling_method)
            if config_hash:
                _tag_embeddings_cache[cache_key] = (phrase_embeddings, phrase_meta)
        except Exception as e:
            Log.e(TAG, "Encoding phrases failed", error=e)
            return []

    try:
        segment_embeddings = encode(segments, session, tokenizer, pooling_method=pooling_method)
    except Exception as e:
        Log.e(TAG, "Encoding text segments failed", error=e)
        return []

    similarity_matrix = cos_sim(segment_embeddings, phrase_embeddings)
    phrase_scores = np.max(similarity_matrix, axis=0)

    tag_scores = {}
    for idx, score in enumerate(phrase_scores):
        tag_idx = phrase_meta[idx]["tag_idx"]
        if tag_idx not in tag_scores:
            tag_scores[tag_idx] = score
        else:
            tag_scores[tag_idx] = max(tag_scores[tag_idx], score)

    candidates = []
    threshold = confidence_threshold / 100.0

    for tag_idx, score in tag_scores.items():
        tag_info = tags_config["tags"][tag_idx]
        priority = tag_info.get("priority", 0)
        keywords = tag_info.get("keywords", [])
        final_score = score
        hit_keyword = False
        if score > KEYWORD_BONUS_THRESHOLD:
            for kw in keywords:
                if kw and kw.lower() in text_lower:
                    hit_keyword = True
                    break

        if hit_keyword:
            final_score += KEYWORD_HIT_BONUS

        priority_boost = priority * PRIORITY_WEIGHT
        final_score += priority_boost

        if final_score >= threshold:
            display_conf = min(float(final_score), 1.0)

            candidates.append({
                "score": display_conf,
                "raw_score": float(score),
                "tag": tag_info.get("name"),
                "group": tag_info.get("group"),
                "priority": priority
            })

    candidates.sort(key=lambda x: x["score"], reverse=True)

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

    Log.i(TAG, f"Classified: {len(final_results)} tags found.")
    return final_results