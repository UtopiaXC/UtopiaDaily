import datetime
import inspect
import json
import os
import re
import shutil
import sys
import time

PRIORITY_WEIGHT = 0.005
KEYWORD_HIT_BONUS = 0.04
KEYWORD_BONUS_THRESHOLD = 0.40


def log_err(msg: str):
    sys.stderr.write(msg + "\n")
    sys.stderr.flush()


def clean_text(text: str) -> str:
    if not text:
        return ""
    text = text.replace("\r\n", "\n").replace("\r", "\n")
    text = re.sub(r"!\[.*?\]\(.*?\)", "", text)
    text = re.sub(r"\[(.*?)\]\(.*?\)", r"\1", text)
    text = re.sub(r"<[^>]+>", " ", text)
    text = re.sub(r"http[s]?://\S+", "", text)
    lines = []
    for raw_line in text.split("\n"):
        collapsed = re.sub(r"[ \t]+", " ", raw_line).strip()
        if collapsed:
            lines.append(collapsed)
    return "\n".join(lines)


def split_text_smart(text: str, max_chars: int = 800, max_segments: int = 256):
    if not text:
        return []
    raw_segments = re.split(r"(?<=[.?!。？！\n])\s*", text)
    merged = []
    chunk = ""
    for seg in raw_segments:
        seg = seg.strip()
        if not seg:
            continue
        if len(chunk) + len(seg) < max_chars:
            chunk = (chunk + " " + seg).strip() if chunk else seg
        else:
            if chunk:
                merged.append(chunk)
            chunk = seg
    if chunk:
        merged.append(chunk)
    if len(merged) > max_segments:
        half = max_segments // 2
        return merged[:half] + merged[-half:]
    return merged


def get_pooling_method(model_name: str) -> str:
    lower = model_name.lower()
    for kw in ("bge", "labse", "snowflake", "clip", "gte"):
        if kw in lower:
            return "cls"
    return "mean"


def normalize(np_mod, v):
    norm = np_mod.linalg.norm(v, axis=1, keepdims=True)
    return v / np_mod.clip(norm, a_min=1e-9, a_max=None)


def encode_texts(texts, session, tokenizer, pooling_method="mean"):
    import numpy as np

    if isinstance(texts, str):
        texts = [texts]
    if not texts:
        return np.array([])

    encoded = tokenizer.encode_batch(texts)
    input_ids = np.array([e.ids for e in encoded], dtype=np.int64)
    attention_mask = np.array([e.attention_mask for e in encoded], dtype=np.int64)

    input_names = [i.name for i in session.get_inputs()]
    inputs = {"input_ids": input_ids, "attention_mask": attention_mask}
    if "token_type_ids" in input_names:
        inputs["token_type_ids"] = np.array([e.type_ids for e in encoded], dtype=np.int64)

    outputs = session.run(None, inputs)
    if pooling_method == "cls":
        embeddings = outputs[0][:, 0]
    else:
        token_embeddings = outputs[0]
        mask_exp = np.broadcast_to(np.expand_dims(attention_mask, -1), token_embeddings.shape)
        sum_emb = np.sum(token_embeddings * mask_exp, axis=1)
        sum_mask = np.clip(np.sum(mask_exp, axis=1), a_min=1e-9, a_max=None)
        embeddings = sum_emb / sum_mask

    return normalize(np, embeddings)


def emit_progress(
    stage: str,
    message: str = "",
    current_item: str = "",
    completed_bytes: int = 0,
    total_bytes: int = 0,
    percent: float = -1.0,
    indeterminate: bool = False,
):
    payload = {
        "stage": stage,
        "message": message,
        "current_item": current_item,
        "completed_bytes": int(completed_bytes or 0),
        "total_bytes": int(total_bytes or 0),
        "percent": round(float(percent), 2) if percent >= 0 else -1,
        "indeterminate": bool(indeterminate),
    }
    sys.stderr.write("PROGRESS_JSON:" + json.dumps(payload, ensure_ascii=False) + "\n")
    sys.stderr.flush()


class _DummyLock:
    def __enter__(self):
        return self

    def __exit__(self, *args):
        return False


def _make_hf_tqdm_class(active_label_ref, last_bytes_ref):
    class _ProgressTracker:
        _lock = _DummyLock()

        def __init__(self, *args, **kwargs):
            self._iterable = args[0] if args else kwargs.get("iterable")
            self.total = int(kwargs.get("total") or 0)
            self.n = float(kwargs.get("initial") or 0)
            self.desc = str(kwargs.get("desc") or "")
            self.disable = False
            self.unit = kwargs.get("unit", "B")
            self._last_emit = 0.0
            if self.total > 0:
                last_bytes_ref["total"] = max(last_bytes_ref["total"], self.total)
                last_bytes_ref["completed"] = min(int(self.n), self.total)
                pct = min(95.0, (self.n / self.total) * 95.0)
                emit_progress(
                    stage="downloading",
                    message="downloading weight",
                    current_item=active_label_ref["name"] or self.desc,
                    completed_bytes=int(self.n),
                    total_bytes=self.total,
                    percent=pct,
                    indeterminate=False,
                )

        @classmethod
        def get_lock(cls):
            return cls._lock

        @classmethod
        def set_lock(cls, lock):
            cls._lock = lock

        def __iter__(self):
            if self._iterable is None:
                return iter(())
            for obj in self._iterable:
                yield obj
                self.update(1)

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            self.close()
            return False

        def update(self, n=1):
            try:
                delta = float(n or 0)
            except Exception:
                delta = 0.0
            self.n += delta
            now = time.monotonic()
            done = self.total > 0 and self.n >= self.total
            if not done and (now - self._last_emit) < 0.15:
                return
            self._last_emit = now
            if self.total > 0:
                comp = min(int(self.n), self.total)
                last_bytes_ref["completed"] = comp
                last_bytes_ref["total"] = self.total
                pct = min(95.0, (comp / self.total) * 95.0)
                emit_progress(
                    stage="downloading",
                    message="downloading weight",
                    current_item=active_label_ref["name"] or self.desc,
                    completed_bytes=comp,
                    total_bytes=self.total,
                    percent=pct,
                    indeterminate=False,
                )
            else:
                emit_progress(
                    stage="downloading",
                    message="downloading weight",
                    current_item=active_label_ref["name"] or self.desc,
                    completed_bytes=int(self.n),
                    total_bytes=0,
                    percent=-1.0,
                    indeterminate=True,
                )

        def set_description(self, desc=None, refresh=True):
            if desc:
                self.desc = str(desc)

        def set_description_str(self, desc=None, refresh=True):
            if desc:
                self.desc = str(desc)

        def set_postfix(self, ordered_dict=None, refresh=True, **kwargs):
            pass

        def set_postfix_str(self, s="", refresh=True):
            pass

        def refresh(self, *args, **kwargs):
            pass

        def clear(self, *args, **kwargs):
            pass

        def close(self):
            if self.total > 0:
                comp = min(int(self.n), self.total)
                last_bytes_ref["completed"] = comp
                last_bytes_ref["total"] = self.total
                pct = min(95.0, (comp / self.total) * 95.0)
                emit_progress(
                    stage="downloading",
                    message="downloading weight",
                    current_item=active_label_ref["name"] or self.desc,
                    completed_bytes=comp,
                    total_bytes=self.total,
                    percent=pct,
                    indeterminate=False,
                )

    return _ProgressTracker


def handle_download(payload: dict):
    os.environ.setdefault("HF_HUB_DISABLE_TELEMETRY", "1")
    os.environ.setdefault("TQDM_DISABLE", "1")
    from huggingface_hub import hf_hub_download

    repo_id = (payload.get("repo_id") or "").strip()
    models_dir = (payload.get("models_dir") or "").strip()
    presets_path = (payload.get("presets_path") or "").strip()
    if not repo_id or "/" not in repo_id:
        raise ValueError("Invalid HuggingFace repository ID")
    if not models_dir:
        raise ValueError("models_dir is required")

    os.makedirs(models_dir, exist_ok=True)
    folder_name = repo_id.split("/")[-1]
    local_dir = os.path.join(models_dir, folder_name)
    os.makedirs(local_dir, exist_ok=True)

    emit_progress("resolving", message="checking repository", current_item=repo_id, indeterminate=True)

    supports_tqdm_class = False
    try:
        supports_tqdm_class = "tqdm_class" in inspect.signature(hf_hub_download).parameters
    except Exception:
        supports_tqdm_class = False

    active_label = {"name": ""}
    last_bytes = {"completed": 0, "total": 0}
    tracker_cls = _make_hf_tqdm_class(active_label, last_bytes)

    def _hf_download(filename: str, track: bool = True):
        kwargs = {"repo_id": repo_id, "filename": filename, "local_dir": local_dir}
        if supports_tqdm_class and track:
            kwargs["tqdm_class"] = tracker_cls
        return hf_hub_download(**kwargs)

    weight_candidates = [
        ("onnx/model_quantized.onnx", "model_quantized.onnx"),
        ("onnx/model.onnx", "model.onnx"),
        ("model_quantized.onnx", "model_quantized.onnx"),
        ("model.onnx", "model.onnx"),
    ]
    weight_ok = False
    last_weight_err = None
    weight_size = 0
    for remote_file, local_target in weight_candidates:
        try:
            active_label["name"] = remote_file
            emit_progress("resolving", message="probing weights", current_item=remote_file, indeterminate=True)
            dl_path = _hf_download(remote_file, track=True)
            target_path = os.path.join(local_dir, local_target)
            if os.path.abspath(dl_path) != os.path.abspath(target_path):
                shutil.move(dl_path, target_path)
            weight_size = os.path.getsize(target_path) if os.path.exists(target_path) else max(last_bytes["total"], last_bytes["completed"])
            last_bytes["completed"] = weight_size
            last_bytes["total"] = weight_size
            emit_progress(
                "downloading",
                message="weights downloaded",
                current_item=remote_file,
                completed_bytes=weight_size,
                total_bytes=weight_size,
                percent=95.0,
                indeterminate=False,
            )
            weight_ok = True
            log_err(f"Downloaded ONNX weight: {remote_file}")
            break
        except Exception as err:
            last_weight_err = err
            continue

    if not weight_ok:
        raise RuntimeError(f"Failed to download ONNX weights from {repo_id}: {last_weight_err}")

    aux_steps = [
        ("tokenizer.json", True, 96.5),
        ("config.json", True, 98.0),
        ("special_tokens_map.json", False, 98.5),
        ("tokenizer_config.json", False, 99.0),
    ]
    for aux_file, required, step_pct in aux_steps:
        emit_progress(
            "downloading",
            message="downloading metadata",
            current_item=aux_file,
            completed_bytes=weight_size,
            total_bytes=weight_size,
            percent=step_pct,
            indeterminate=False,
        )
        if required:
            _hf_download(aux_file, track=False)
        else:
            try:
                _hf_download(aux_file, track=False)
            except Exception:
                pass

    onnx_subdir = os.path.join(local_dir, "onnx")
    if os.path.isdir(onnx_subdir) and not os.listdir(onnx_subdir):
        os.rmdir(onnx_subdir)

    emit_progress(
        "finalizing",
        message="writing metadata",
        current_item=folder_name,
        completed_bytes=weight_size,
        total_bytes=weight_size,
        percent=99.5,
        indeterminate=False,
    )

    model_info = {"name": folder_name, "hf_repo": repo_id}
    if presets_path and os.path.exists(presets_path):
        try:
            with open(presets_path, "r", encoding="utf-8") as f:
                preset_groups = json.load(f)
            for grp in preset_groups:
                for item in grp.get("models", []):
                    if item.get("hf_repo") == repo_id:
                        model_info = dict(item)
                        break
        except Exception:
            pass
    model_info["downloaded_at"] = datetime.datetime.utcnow().isoformat() + "Z"
    with open(os.path.join(local_dir, "info.json"), "w", encoding="utf-8") as f:
        json.dump(model_info, f, indent=2, ensure_ascii=False)

    emit_progress(
        "ready",
        message="completed",
        current_item=folder_name,
        completed_bytes=weight_size,
        total_bytes=weight_size,
        percent=100.0,
        indeterminate=False,
    )

    return {"status": "ok", "folder": folder_name, "info": model_info}


def handle_classify(payload: dict):
    import numpy as np
    import onnxruntime as ort
    from tokenizers import Tokenizer

    model_dir = (payload.get("model_dir") or "").strip()
    model_name = (payload.get("model_name") or "").strip()
    text = payload.get("text") or ""
    confidence = float(payload.get("confidence", 60))
    profile = payload.get("profile") or {}
    groups = profile.get("groups") or []

    if not model_dir or not os.path.isdir(model_dir):
        raise ValueError(f"ONNX model directory missing: {model_dir}")

    onnx_path = os.path.join(model_dir, "model_quantized.onnx")
    if not os.path.exists(onnx_path):
        onnx_path = os.path.join(model_dir, "model.onnx")
    if not os.path.exists(onnx_path):
        raise ValueError(f"ONNX model weights missing in {model_dir}")

    tokenizer_path = os.path.join(model_dir, "tokenizer.json")
    if not os.path.exists(tokenizer_path):
        raise ValueError(f"tokenizer.json missing in {model_dir}")

    cleaned_text = clean_text(text)
    if not cleaned_text or not groups:
        return {"status": "ok", "tags": []}

    segments = split_text_smart(cleaned_text)
    if not segments:
        return {"status": "ok", "tags": []}

    all_tags = []
    group_meta = {}
    for g in groups:
        g_name = g.get("name", "")
        mode = g.get("selection_mode", "limit") or "limit"
        max_res = int(g.get("max_results", 1) or 1)
        if mode == "single":
            limit = 1
        elif mode == "multi":
            limit = 1000
        else:
            limit = max(1, max_res)
        group_meta[g_name] = limit
        for t in g.get("tags") or []:
            t_copy = dict(t)
            t_copy["_group"] = g_name
            all_tags.append(t_copy)

    if not all_tags:
        return {"status": "ok", "tags": []}

    all_phrases = []
    phrase_meta = []
    for idx, t in enumerate(all_tags):
        name = (t.get("name") or "").strip()
        desc = (t.get("description") or "").strip()
        keywords = t.get("keywords") or []
        if name:
            all_phrases.append(name)
            phrase_meta.append(idx)
        if desc:
            all_phrases.append(desc)
            phrase_meta.append(idx)
        for kw in keywords:
            kw_s = (kw or "").strip()
            if kw_s:
                all_phrases.append(kw_s)
                phrase_meta.append(idx)

    if not all_phrases:
        return {"status": "ok", "tags": []}

    tokenizer = Tokenizer.from_file(tokenizer_path)
    tokenizer.enable_padding(pad_id=0, pad_token="[PAD]")
    tokenizer.enable_truncation(max_length=512)
    sess_opts = ort.SessionOptions()
    sess_opts.log_severity_level = 3
    session = ort.InferenceSession(onnx_path, sess_opts)

    pooling = get_pooling_method(model_name or os.path.basename(model_dir))
    phrase_emb = encode_texts(all_phrases, session, tokenizer, pooling_method=pooling)
    seg_emb = encode_texts(segments, session, tokenizer, pooling_method=pooling)

    sim_matrix = np.dot(seg_emb, phrase_emb.T)
    phrase_scores = np.max(sim_matrix, axis=0)

    tag_scores = {}
    for p_idx, score in enumerate(phrase_scores):
        t_idx = phrase_meta[p_idx]
        s_val = float(score)
        if t_idx not in tag_scores or s_val > tag_scores[t_idx]:
            tag_scores[t_idx] = s_val

    threshold = confidence / 100.0
    text_lower = cleaned_text.lower()
    candidates = []
    best_candidate = None
    for t_idx, raw_score in tag_scores.items():
        tag_info = all_tags[t_idx]
        priority = int(tag_info.get("priority", 0) or 0)
        keywords = tag_info.get("keywords") or []
        final_score = raw_score
        if raw_score > KEYWORD_BONUS_THRESHOLD:
            for kw in keywords:
                if kw and kw.lower() in text_lower:
                    final_score += KEYWORD_HIT_BONUS
                    break
        bounded_conf = round(min(max(final_score, 0.0), 1.0), 4)
        cand_obj = {
            "tag_id": int(tag_info.get("id", 0) or 0),
            "tag_name": tag_info.get("name", ""),
            "group_name": tag_info.get("_group", ""),
            "confidence": bounded_conf,
            "reason": "onnx_similarity",
        }
        if best_candidate is None or bounded_conf > best_candidate["confidence"]:
            best_candidate = cand_obj
        if final_score >= threshold:
            candidates.append({
                "obj": cand_obj,
                "priority": priority,
                "sort_score": final_score + priority * PRIORITY_WEIGHT,
            })

    candidates.sort(
        key=lambda x: (x["sort_score"], x["priority"], x["obj"]["confidence"]),
        reverse=True,
    )
    group_counts = {}
    selected = []
    for cand_entry in candidates:
        cand = cand_entry["obj"]
        g_name = cand["group_name"]
        limit = group_meta.get(g_name, 1)
        cur = group_counts.get(g_name, 0)
        if cur >= limit:
            continue
        group_counts[g_name] = cur + 1
        selected.append(cand)

    return {"status": "ok", "tags": selected, "best_candidate": best_candidate}


def main():
    raw_stdin = sys.stdin.read()
    payload = json.loads(raw_stdin) if raw_stdin.strip() else {}
    action = payload.get("action", "classify")
    if action == "download":
        res = handle_download(payload)
    elif action == "classify":
        res = handle_classify(payload)
    else:
        raise ValueError(f"Unsupported action: {action}")
    sys.stdout.write(json.dumps(res, ensure_ascii=False))
    sys.stdout.flush()


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        log_err(f"onnx_worker error: {exc}")
        sys.exit(1)

