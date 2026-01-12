import requests
from bs4 import BeautifulSoup
import re
from markdownify import markdownify as md
from datetime import datetime, timedelta, timezone
import time
from src.utils.logger.logger import Log

TAG = "TELEGRAM_CHANNEL_MODULE_SCRAPER"

class TelegramScraper:
    def __init__(self, config):
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
        }
        self.target_tz_offset = 9  # JST
        self.lookback_days = int(config.get("lookback_days", 3))
        self.max_pages = int(config.get("max_pages", 5))

    def _get_message_date(self, msg_node):
        time_tag = msg_node.find('time', class_='time')
        if time_tag:
            raw_time_str = time_tag.get('datetime')
            try:
                return datetime.fromisoformat(raw_time_str)
            except ValueError:
                return None
        return None

    def _parse_single_message(self, msg, channel_name):
        # 1. Metadata
        post_id_full = msg.get('data-post')
        post_link = f"https://t.me/{post_id_full}" if post_id_full else ""

        dt_utc = self._get_message_date(msg)
        if not dt_utc: return None

        target_tz = timezone(timedelta(hours=self.target_tz_offset))
        dt_target = dt_utc.astimezone(target_tz)
        formatted_time = dt_target.strftime('%Y-%m-%d %H:%M')

        # 2. Quote
        reply_md = ""
        reply_block = msg.find('a', class_='tgme_widget_message_reply')
        if reply_block:
            author_tag = reply_block.find('span', class_='tgme_widget_message_author_name')
            text_tag = reply_block.find('div', class_='tgme_widget_message_metatext') or \
                       reply_block.find('div', class_='tgme_widget_message_text')
            author = author_tag.get_text().strip() if author_tag else "Original"
            text = text_tag.get_text().strip() if text_tag else ""
            if text:
                reply_md = f"> **Quote {author}**: {text}"

        # 3. Main Content
        all_text_divs = msg.find_all('div', class_='tgme_widget_message_text')
        main_text_raw = ""
        plain_text_content = ""

        for div in all_text_divs:
            if reply_block and any(p == reply_block for p in div.parents):
                continue
            main_text_raw = md(str(div), strip=['div', 'span', 'img']).strip()
            plain_text_content = div.get_text(strip=True)
            break

        # 4. Link Preview
        preview_block = msg.find(class_='tgme_widget_message_link_preview')
        p_title, p_desc, p_site_name, p_image_url = "", "", "", ""
        if preview_block:
            t_div = preview_block.find(class_='link_preview_title')
            d_div = preview_block.find(class_='link_preview_description')
            s_div = preview_block.find(class_='link_preview_site_name')
            if t_div: p_title = t_div.get_text().strip()
            if d_div: p_desc = d_div.get_text().strip()
            if s_div: p_site_name = s_div.get_text().strip()

            img_div = preview_block.find(class_='link_preview_right_image') or \
                      preview_block.find(class_='link_preview_image')
            if img_div:
                style = img_div.get('style', '')
                match = re.search(r"url\(['\"]?(.*?)['\"]?\)", style)
                if match: p_image_url = match.group(1)

        # 5. Pure Link Check
        raw_content = main_text_raw.replace('\\n', '\n')
        clean_check_str = re.sub(r'[<>]', '', raw_content).strip()
        is_pure_link = bool(
            re.match(r'^https?://\S+$', plain_text_content) or re.match(r'^https?://\S+$', clean_check_str))

        # 6. Title Generation
        title_text = ""
        final_content = ""
        if is_pure_link:
            title_text = p_title if p_title else (f"From {p_site_name}" if p_site_name else "Link Share")
            final_content = p_desc if p_desc else ""
        else:
            title_text = p_title if p_title else (raw_content[:30] if raw_content else "Untitled")
            final_content = f"{raw_content}\n\n> {p_desc}" if (
                    p_title and p_desc and p_desc not in raw_content) else raw_content

        # Fallback
        if is_pure_link and title_text == "Link Share" and not final_content:
            final_content = f"<{post_link}>"

        clean_title = title_text.replace('\n', ' ').strip()[:57]

        # 7. Assemble Markdown
        # Note: Tags are now handled by the system API, so we don't add them here.
        md_lines = [f"### {clean_title}", f"*{formatted_time} (JST)*"]
        if reply_md: md_lines.append(reply_md)
        if final_content: md_lines.append(final_content)
        if p_image_url: md_lines.append(f"![image]({p_image_url})")
        if post_link: md_lines.append(f"[🔗 Source]({post_link})")

        full_md = "\n\n".join(md_lines) + "\n\n---\n"

        return {
            "source": channel_name,
            "title": clean_title,
            "date": dt_utc.isoformat(),
            "markdown": full_md,
            "raw_data": {"link": post_link, "summary": final_content[:50], "text": plain_text_content}
        }

    def run(self, channel_urls, on_item_found):
        """
        Main entry point for scraping
        """
        now_utc = datetime.now(timezone.utc)
        cutoff_date = now_utc - timedelta(days=self.lookback_days)

        Log.i(TAG, f"Start scraping. Cutoff: {cutoff_date.strftime('%Y-%m-%d')}")

        for url in channel_urls:
            try:
                count = self._scrape_single_channel(url, cutoff_date, on_item_found)
                Log.i(TAG, f"[{url.split('/')[-1]}] Scraped: {count} items")
            except Exception as e:
                Log.e(TAG, f"Failed to scrape {url}", error=e)

    def _scrape_single_channel(self, channel_url, cutoff_date, callback):
        base_url = channel_url.replace("t.me/", "t.me/s/") if "/s/" not in channel_url else channel_url
        current_url = base_url
        page_count = 0
        items_count = 0

        while page_count < self.max_pages:
            page_count += 1
            # Log.d(TAG, f"Requesting {current_url} (Page {page_count})")

            res = requests.get(current_url, headers=self.headers, timeout=15)
            soup = BeautifulSoup(res.text, 'html.parser')
            messages = soup.find_all('div', class_='tgme_widget_message')

            if not messages: break

            first_msg_date = self._get_message_date(messages[0])

            for msg in messages:
                msg_date = self._get_message_date(msg)
                if not msg_date or msg_date < cutoff_date:
                    continue

                item = self._parse_single_message(msg, channel_url.split('/')[-1])
                if item:
                    callback(item)
                    items_count += 1

            # Pagination check
            if first_msg_date and first_msg_date > cutoff_date:
                post_data = messages[0].get('data-post')
                if post_data and '/' in post_data:
                    current_url = f"{base_url}?before={post_data.split('/')[-1]}"
                    time.sleep(1)
                    continue
            break

        return items_count

    def test_connection(self, channel_url):
        """
        Test connection to a single channel.
        Returns: (success: bool, message: str)
        """
        try:
            target_url = channel_url.replace("t.me/", "t.me/s/") if "/s/" not in channel_url else channel_url
            res = requests.get(target_url, headers=self.headers, timeout=10)
            if res.status_code == 200:
                # Basic validation that we got a Telegram page
                if "tgme_widget_message" in res.text or "tgme_channel_info" in res.text:
                    return True, f"Successfully connected to {channel_url}"
                else:
                    return False, f"Connected to {channel_url} but content doesn't look like a Telegram channel"
            else:
                return False, f"Failed to connect to {channel_url}, status: {res.status_code}"
        except Exception as e:
            return False, f"Connection test failed: {str(e)}"
