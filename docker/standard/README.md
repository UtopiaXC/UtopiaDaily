# Standard Image

This image contains all libs and requirements without transformer models.

**Features:**

1. No need to wait for installing environment(System libs and pip wheels) at start.
2. No transformer models included, you need to download them from HuggingFace via the dashboard.
3. Image size is about 450mb.

> NOTE: MUST check your network for accessing HuggingFace before using this image.

```bash
# How to build
cd ${base_path}/UtopiaDaily
docker build -f docker/standard/Dockerfile -t utopia-daily .

# how to use
docker run -itd \
-p 8000:8000    \
-v ${base_path}/utopia-daily/database:/app/database    \
-v ${base_path}/utopia-daily/config:/app/config    \
--name utopiaxc-daily      \
utopia-daily:latest
```