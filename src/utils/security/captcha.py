import random
import uuid
import time
import io
import base64
import math
from PIL import Image, ImageDraw, ImageFont, ImageFilter

# Simple in-memory storage: {captcha_id: (answer, expires_at)}
_CAPTCHA_STORE = {}
MAX_STORE_SIZE = 100
CAPTCHA_TTL = 60 # 1 minute

def _generate_math_question():
    a = random.randint(1, 20)
    b = random.randint(1, 20)
    operator = random.choice(['+', '-'])
    
    if operator == '+':
        answer = a + b
    else:
        if a < b: a, b = b, a
        answer = a - b
    
    return f"{a} {operator} {b} = ?", str(answer)

def _apply_distortion(image):
    """
    Apply sine wave distortion to the image.
    """
    width, height = image.size
    new_image = Image.new("RGB", (width, height), (255, 255, 255))
    
    # Random distortion parameters
    period = random.randint(20, 40) # Wave width
    amplitude = random.randint(3, 6) # Wave height
    phase = random.randint(0, 100)
    
    for x in range(width):
        # Calculate y offset based on sine wave
        y_offset = int(amplitude * math.sin(2 * math.pi * (x + phase) / period))
        
        for y in range(height):
            # Source y coordinate
            src_y = y + y_offset
            
            # Check bounds
            if 0 <= src_y < height:
                pixel = image.getpixel((x, src_y))
                new_image.putpixel((x, y), pixel)
                
    return new_image

def generate_captcha():
    """
    Generates a distorted arithmetic captcha image (PNG Base64).
    Returns: (captcha_id, base64_image_src)
    """
    now = time.time()
    
    # Cleanup
    if len(_CAPTCHA_STORE) > MAX_STORE_SIZE:
        keys_to_del = [k for k, v in _CAPTCHA_STORE.items() if v[1] < now]
        for k in keys_to_del:
            del _CAPTCHA_STORE[k]
        if len(_CAPTCHA_STORE) > MAX_STORE_SIZE:
            keys = list(_CAPTCHA_STORE.keys())
            for i in range(100):
                if i < len(keys):
                    del _CAPTCHA_STORE[keys[i]]

    question, answer = _generate_math_question()
    captcha_id = str(uuid.uuid4())
    _CAPTCHA_STORE[captcha_id] = (answer, now + CAPTCHA_TTL)
    
    # Image Config
    width, height = 160, 60
    image = Image.new('RGB', (width, height), (255, 255, 255))
    draw = ImageDraw.Draw(image)
    
    # Draw Noise (Lines)
    for _ in range(5):
        x1 = random.randint(0, width)
        y1 = random.randint(0, height)
        x2 = random.randint(0, width)
        y2 = random.randint(0, height)
        draw.line([(x1, y1), (x2, y2)], fill=(200, 200, 200), width=1)
        
    # Draw Noise (Points)
    for _ in range(100):
        x = random.randint(0, width)
        y = random.randint(0, height)
        draw.point((x, y), fill=(180, 180, 180))

    # Draw Text
    # Try to load a default font, fallback to simple if not found
    try:
        # Use a larger font size if possible. 
        # On Linux/Docker, paths might differ. 'arial.ttf' is common on Windows.
        # For portability, we can use ImageFont.load_default() but it's very small.
        # Let's try to find a system font or use default scaled up? 
        # Pillow's default font doesn't scale well.
        # We'll use load_default() for safety, but it might be small.
        # Ideally, ship a .ttf file in assets.
        font = ImageFont.load_default()
        # Since default font is small, we might just draw it centered.
        # To make it bigger without TTF, we can draw on a small image and resize up, but that's blurry.
        # Let's assume we can use a basic font or just accept default for now.
        # BETTER: Use a simple TTF if available, or fallback.
        # For this demo, I'll use load_default() but draw it clearly.
    except:
        font = ImageFont.load_default()

    # Calculate text size (approximate for default font)
    # text_bbox = draw.textbbox((0, 0), question, font=font)
    # text_w = text_bbox[2] - text_bbox[0]
    # text_h = text_bbox[3] - text_bbox[1]
    
    # Since default font is tiny, let's try to draw it larger by drawing on a smaller canvas and scaling up?
    # No, that's ugly.
    # Let's try to use a common font path.
    font_path = None
    import os
    if os.name == 'nt': # Windows
        font_path = "arial.ttf"
    else: # Linux
        # Common paths
        possible_paths = [
            "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
            "/usr/share/fonts/TTF/DejaVuSans.ttf",
            "/usr/share/fonts/liberation/LiberationSans-Regular.ttf"
        ]
        for p in possible_paths:
            if os.path.exists(p):
                font_path = p
                break
    
    try:
        if font_path:
            font = ImageFont.truetype(font_path, 28)
        else:
            # Fallback: try to load without path (system path)
            font = ImageFont.truetype("DejaVuSans.ttf", 28)
    except:
        # Ultimate fallback
        font = ImageFont.load_default()

    # Draw text centered
    # Random color
    text_color = (random.randint(0, 100), random.randint(0, 100), random.randint(0, 100))
    
    # Position with some randomness
    text_x = 20 + random.randint(-5, 5)
    text_y = 15 + random.randint(-5, 5)
    
    draw.text((text_x, text_y), question, font=font, fill=text_color)

    # Apply Distortion
    image = _apply_distortion(image)
    
    # Save to buffer
    buffer = io.BytesIO()
    image.save(buffer, format="PNG")
    img_str = base64.b64encode(buffer.getvalue()).decode("utf-8")
    
    return captcha_id, f"data:image/png;base64,{img_str}"

def verify_captcha(captcha_id, user_input):
    if not captcha_id or not user_input:
        return False
    
    record = _CAPTCHA_STORE.get(captcha_id)
    if not record:
        return False
    
    answer, expires_at = record
    
    del _CAPTCHA_STORE[captcha_id]
    
    if time.time() > expires_at:
        return False
        
    return answer == user_input.strip()
