# Full Image

This image contains everything.

**Features:**

1. Everything is done,just start it. 
2. Image size is about 1G.

```bash
# How to build
cd ${base_path}/UtopiaDaily
docker build -f docker/full/Dockerfile -t utopia-daily-full .

# how to use
docker run -itd \
-p 8000:8000    \
-v ${base_path}/utopia-daily/database:/app/database    \
-v ${base_path}/utopia-daily/config:/app/config    \
--name utopiaxc-daily      \
utopia-daily-full:latest
```