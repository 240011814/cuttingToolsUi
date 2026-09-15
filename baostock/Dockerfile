FROM python:3.12-slim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV TZ=Asia/Shanghai

RUN apt-get update \
    && apt-get install -y --no-install-recommends tzdata \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone \
    && rm -rf /var/lib/apt/lists/*

ARG PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple
ARG PIP_TRUSTED_HOST=pypi.tuna.tsinghua.edu.cn

WORKDIR /app

COPY requirements.txt ./

RUN python -m pip install --no-cache-dir --upgrade pip \
    && python -m pip install --no-cache-dir \
      --index-url ${PIP_INDEX_URL} \
      --trusted-host ${PIP_TRUSTED_HOST} \
      -r requirements.txt

COPY pyproject.toml ./
COPY main.py ./
COPY baostock_api ./baostock_api

RUN mkdir -p /app/data

EXPOSE 3002

# /health 在锁卡死(持锁超 LOCK_STUCK_SECONDS)时返回 503, 探测失败 -> unhealthy
# 注意: docker 自身不会重启 unhealthy 容器, 自动拉起需 autoheal sidecar 或外部编排
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD python -c "import os,urllib.request; urllib.request.urlopen('http://127.0.0.1:' + os.environ.get('BAOSTOCK_API_PORT', '3002') + '/health', timeout=8)"

CMD ["python", "main.py"]
