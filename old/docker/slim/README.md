# Slim Image

This image only contains python source code and environment.

**Features:**

1. No system environment, image size is small. Will run install sh script by container at start.
2. No transformer models included, you need to download them from HuggingFace via the dashboard.
3. Image size is about 200mb.

> NOTE: MUST check your network for accessing HuggingFace, pip before using this image.

```bash
# How to build
cd ${base_path}/UtopiaDaily
docker build -f docker/slim/Dockerfile -t utopia-daily-slim .

# how to use
docker run -itd \
-p 8000:8000    \
-v ${base_path}/utopia-daily/database:/app/database    \
-v ${base_path}/utopia-daily/config:/app/config    \
# Option: You can change your pip mirror here
-e PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple \
# Option: You can change your apt mirror here
-e DEBIAN_MIRROR=http://mirrors.tuna.tsinghua.edu.cn \
--name utopiaxc-daily-slim      \
utopia-daily-slim:latest
```