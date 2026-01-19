from curl_cffi import requests
import time
from hashlib import md5
import urllib.parse
from functools import reduce
from datetime import datetime
from src.utils.logger.logger import Log

TAG = "BILIBILI_SCRAPER"

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3',
    'Referer': 'https://www.bilibili.com/'
}

mixinKeyEncTab = [
    46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
    33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
    61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
    36, 20, 34, 44, 52
]

def get_mixin_key(orig: str):
    return reduce(lambda s, i: s + orig[i], mixinKeyEncTab, '')[:32]

def enc_wbi(params: dict, img_key: str, sub_key: str):
    mixin_key = get_mixin_key(img_key + sub_key)
    curr_time = round(time.time())
    params['wts'] = curr_time
    params = dict(sorted(params.items()))
    params = {
        k : ''.join(filter(lambda chr: chr not in "!'()*", str(v)))
        for k, v
        in params.items()
    }
    query = urllib.parse.urlencode(params)
    wbi_sign = md5((query + mixin_key).encode()).hexdigest()
    params['w_rid'] = wbi_sign
    return params

def get_wbi_keys(session: requests.Session) -> tuple[str, str]:
    try:
        resp = session.get('https://api.bilibili.com/x/web-interface/nav', headers=headers)
        resp.raise_for_status()
        json_content = resp.json()
        img_url: str = json_content['data']['wbi_img']['img_url']
        sub_url: str = json_content['data']['wbi_img']['sub_url']
        img_key = img_url.rsplit('/', 1)[1].split('.')[0]
        sub_key = sub_url.rsplit('/', 1)[1].split('.')[0]
        return img_key, sub_key
    except Exception as e:
        Log.e(TAG, f"Failed to get Wbi keys: {e}")
        return None, None

def get_ranking():
    url = "https://api.bilibili.com/x/web-interface/ranking/v2?rid=0&type=all"
    try:
        res = requests.get(url, headers=headers, timeout=10, impersonate="chrome110")
        data = res.json()
        if data["code"] != 0:
            Log.e(TAG, f"Failed to fetch ranking: {data['message']}")
            return []
        
        items = []
        for item in data["data"]["list"][:20]: # Top 20
            video = {
                "title": item["title"],
                "summary": item["desc"],
                "source": f"Bilibili Ranking #{item.get('rank', '?')}",
                "from_url": item["short_link"] if "short_link" in item else f"https://www.bilibili.com/video/{item['bvid']}",
                "content": f"<iframe src=\"//player.bilibili.com/player.html?bvid={item['bvid']}&page=1\" scrolling=\"no\" border=\"0\" frameborder=\"no\" framespacing=\"0\" allowfullscreen=\"true\" style=\"width: 100%; height: 500px;\"></iframe><br><br>{item['desc']}",
                "content_type": "html", # Use HTML to embed player
                "datetime_released": datetime.fromtimestamp(item["pubdate"]).isoformat(),
                "attachments": [{
                    "original_url": item["pic"],
                    "converted_url": "",
                    "type": "image",
                    "store_policy": "store_and_replace"
                }],
                "tags": [{"tag": item["tname"], "confidence": 1.0}],
                "metadata": {
                    "bvid": item["bvid"],
                    "author": item["owner"]["name"],
                    "view": item["stat"]["view"],
                    "danmaku": item["stat"]["danmaku"]
                }
            }
            items.append(video)
        return items
    except Exception as e:
        Log.e(TAG, f"Error fetching ranking: {e}")
        return []

def get_uploader_videos(uid):
    session = requests.Session()
    # 1. Get keys and cookies
    img_key, sub_key = get_wbi_keys(session)
    if not img_key or not sub_key:
        return []

    # 2. Prepare params
    params = {
        'mid': uid,
        'ps': 10, # Page size
        'tid': 0, # All partitions
        'pn': 1,  # Page number
        'keyword': '',
        'order': 'pubdate', # Sort by publish date
        'platform': 'web',
        'web_location': 1550101,
        'order_avoided': 'true'
    }
    
    # 3. Sign params
    signed_params = enc_wbi(params, img_key, sub_key)
    
    # 4. Request API
    url = "https://api.bilibili.com/x/space/wbi/arc/search"
    
    # DEBUG: Print full URL
    full_url = f"{url}?{urllib.parse.urlencode(signed_params)}"
    Log.w(TAG, f"DEBUG URL: {full_url}")
    
    try:
        res = session.get(url, params=signed_params, headers=headers, timeout=10, impersonate="chrome110")
        data = res.json()
        
        if data["code"] != 0:
            Log.w(TAG, f"Failed to fetch videos for uid {uid}: {data['message']} (Code: {data['code']})")
            return []
            
        items = []
        if "data" in data and "list" in data["data"] and "vlist" in data["data"]["list"]:
            vlist = data["data"]["list"]["vlist"]
            for item in vlist:
                video = {
                    "title": item["title"],
                    "summary": item["description"],
                    "source": f"Bilibili UP: {item['author']}",
                    "from_url": f"https://www.bilibili.com/video/{item['bvid']}",
                    "content": f"<iframe src=\"//player.bilibili.com/player.html?bvid={item['bvid']}&page=1\" scrolling=\"no\" border=\"0\" frameborder=\"no\" framespacing=\"0\" allowfullscreen=\"true\" style=\"width: 100%; height: 500px;\"></iframe><br><br>{item['description']}",
                    "content_type": "html",
                    "datetime_released": datetime.fromtimestamp(item["created"]).isoformat(),
                    "attachments": [{
                        "original_url": item["pic"],
                        "converted_url": "",
                        "type": "image",
                        "store_policy": "store_and_replace"
                    }],
                    "tags": [], 
                    "metadata": {
                        "bvid": item["bvid"],
                        "author": item["author"],
                        "play": item["play"],
                        "comment": item["comment"]
                    }
                }
                items.append(video)
        return items
    except Exception as e:
        Log.e(TAG, f"Error fetching videos for uid {uid}: {e}")
        return []