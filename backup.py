import requests
import time
import io
from typing import Optional
from utils import Config

def upload_to_cloud(cfg: Config, file_path: str, file_content: bytes) -> None:
    if not file_content:
        raise ValueError("file content is empty")

    upload_url_api = f"{cfg.backup.host}/v1/disk/resources/upload/"
    headers = {"Authorization": cfg.backup.auth_key if hasattr(cfg.backup, 'auth_key') else cfg.backup.authKey}

    # Handle both camelCase and snake_case for authKey if it was loaded differently
    auth_key = getattr(cfg.backup, 'authKey', getattr(cfg.backup, 'auth_key', ""))

    params = {
        "path": "/" + file_path,
        "overwrite": "true"
    }

    response = requests.get(upload_url_api, headers={"Authorization": auth_key}, params=params)
    if response.status_code != 200:
        raise Exception(f"failed to get upload URL: {response.status_code}, body: {response.text}")

    data = response.json()
    href = data.get("href")
    method = data.get("method", "PUT")

    max_retries = 3
    last_err = None

    for attempt in range(max_retries):
        try:
            upload_resp = requests.request(method, href, data=file_content)
            if upload_resp.status_code in [200, 201]:
                return

            last_err = Exception(f"failed to upload file: {upload_resp.status_code}, response body: {upload_resp.text}")
        except Exception as e:
            last_err = e

        time.sleep(2)

    raise Exception(f"upload failed after {max_retries} retries: {last_err}")
