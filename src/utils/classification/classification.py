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

# Cache loaded models to avoid reloading
_loaded_models = {}

# Cache tag embeddings: {(model_name, config_hash): (all_phrases, phrase_embeddings, phrase_to_tag_map)}
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
    models = [{"label": "None", "value": "None"}] # Default option
    if not os.path.exists(MODELS_DIR):
        return models
        
    for item in os.listdir(MODELS_DIR):
        item_path = os.path.join(MODELS_DIR, item)
        if os.path.isdir(item_path):
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
        tokenizer.enable_padding(pad_id=0, pad_token="[PAD]", length=512)
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
    return np.dot(a, b.T)

def clean_text(text):
    if not text:
        return ""
    
    # Remove Markdown images
    text = re.sub(r'!\[.*?\]\(.*?\)', '', text)
    # Replace Markdown links
    text = re.sub(r'\[(.*?)\]\(.*?\)', r'\1', text)
    # Remove HTML tags
    text = re.sub(r'<[^>]+>', '', text)
    # Remove URLs
    text = re.sub(r'http[s]?://\S+', '', text)
    # Remove excessive whitespace
    text = re.sub(r'\s+', ' ', text).strip()
    # Remove common source attributions at the end
    text = re.sub(r'（.*?）$', '', text)
    text = re.sub(r'\(.*?\)$', '', text)
    
    return text

def split_text(text, max_segments=10):
    # Split by common sentence delimiters
    segments = re.split(r'(?<=[.?!。？！\n])\s*', text)
    valid_segments = [s.strip() for s in segments if len(s.strip()) > 10]
    
    if not valid_segments:
        valid_segments = [text]
        
    return valid_segments[:max_segments]

def get_config_hash(tags_config):
    """Generate a hash for the tags configuration to use as a cache key."""
    try:
        # Sort keys to ensure consistent JSON string
        config_str = json.dumps(tags_config, sort_keys=True)
        return hashlib.md5(config_str.encode('utf-8')).hexdigest()
    except Exception:
        return None

def classify_text(text, model_name, tags_config, confidence_threshold=60):
    # Log the input text as requested
    Log.d(TAG, f"Classifying text: {text[:100]}...") 
    
    if not text:
        Log.w(TAG, "Input text is empty.")
        return []
        
    if not model_name or model_name == "None":
        Log.w(TAG, "No model selected.")
        return []

    # 1. Clean Text
    cleaned_text = clean_text(text)
    if not cleaned_text:
        Log.w(TAG, "Text became empty after cleaning.")
        return []
        
    # 2. Split Text into Segments
    segments = split_text(cleaned_text)
    Log.d(TAG, f"Text split into {len(segments)} segments for matching.")

    try:
        session, tokenizer = get_model(model_name)
    except Exception as e:
        Log.e(TAG, f"Could not get model {model_name}: {e}")
        return []

    # 3. Prepare Labels (Use Cache)
    config_hash = get_config_hash(tags_config)
    cache_key = (model_name, config_hash)
    
    all_phrases = []
    phrase_embeddings = None
    phrase_to_tag_map = []

    if config_hash and cache_key in _tag_embeddings_cache:
        Log.d(TAG, "Using cached tag embeddings.")
        all_phrases, phrase_embeddings, phrase_to_tag_map = _tag_embeddings_cache[cache_key]
    else:
        Log.d(TAG, "Computing new tag embeddings...")
        if "tags" not in tags_config:
            return []

        for i, tag in enumerate(tags_config["tags"]):
            name = tag.get("name")
            keywords = tag.get("keywords", [])
            
            if name:
                all_phrases.append(name)
                phrase_to_tag_map.append(i)
                
            for kw in keywords:
                if kw:
                    all_phrases.append(kw)
                    phrase_to_tag_map.append(i)

        if not all_phrases:
            Log.w(TAG, "No phrases generated from config.")
            return []

        try:
            phrase_embeddings = encode(all_phrases, session, tokenizer)
            # Cache the result
            if config_hash:
                _tag_embeddings_cache[cache_key] = (all_phrases, phrase_embeddings, phrase_to_tag_map)
                # Optional: Limit cache size if needed, but config changes are rare
        except Exception as e:
            Log.e(TAG, "Encoding phrases failed", error=e)
            return []

    # 4. Encode Text Segments
    try:
        segment_embeddings = encode(segments, session, tokenizer)
    except Exception as e:
        Log.e(TAG, "Encoding text segments failed", error=e)
        return []

    # 5. Compute Similarity Matrix
    # (num_segments, num_phrases)
    similarity_matrix = cos_sim(segment_embeddings, phrase_embeddings)
    
    # 6. Find Best Match for Each Phrase
    # max_sim_per_phrase: (num_phrases,)
    max_sim_per_phrase = np.max(similarity_matrix, axis=0)

    # 7. Aggregate Scores per Tag
    tag_scores = {} # tag_index -> max_score
    
    for idx, score in enumerate(max_sim_per_phrase):
        tag_idx = phrase_to_tag_map[idx]
        if tag_idx not in tag_scores:
            tag_scores[tag_idx] = score
        else:
            tag_scores[tag_idx] = max(tag_scores[tag_idx], score)

    # 8. Filter Results
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
            Log.d(TAG, f"Candidate: {tag_info.get('name')} (Score: {score:.4f})")

    candidates.sort(key=lambda x: x["score"], reverse=True)

    # 9. Apply Limits
    group_limits = {g["name"]: g.get("max_results", 1) for g in tags_config.get("groups", [])}
    group_counts = {g_name: 0 for g_name in group_limits}

    final_results = []
    for cand in candidates:
        group = cand["group"]
        if group in group_limits:
            if group_counts[group] < group_limits[group]:
                final_results.append({
                    "tag": cand["tag"],
                    "confidence": round(cand["score"], 4)
                })
                group_counts[group] += 1
        else:
            final_results.append({
                "tag": cand["tag"],
                "confidence": round(cand["score"], 4)
            })
            
    Log.i(TAG, f"Classification finished. Found {len(final_results)} tags.")
    return final_results
