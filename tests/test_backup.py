import pytest
import responses
import requests
from backup import upload_to_cloud
from utils import Config, BackupConfig

@pytest.fixture
def config():
    return Config(backup=BackupConfig(
        host="https://cloud-api.yandex.net",
        authKey="OAuth secret"
    ))

@responses.activate
def test_upload_to_cloud_success(config):
    # Mock GET for upload URL
    responses.add(
        responses.GET,
        "https://cloud-api.yandex.net/v1/disk/resources/upload/",
        json={"href": "https://upload-link.com", "method": "PUT"},
        status=200
    )
    # Mock PUT for actual upload
    responses.add(
        responses.PUT,
        "https://upload-link.com",
        status=201
    )

    upload_to_cloud(config, "test.txt", b"content")

    assert len(responses.calls) == 2
    assert responses.calls[0].request.params["path"] == "/test.txt"

@responses.activate
def test_upload_to_cloud_retry_success(config):
    # Mock GET for upload URL
    responses.add(
        responses.GET,
        "https://cloud-api.yandex.net/v1/disk/resources/upload/",
        json={"href": "https://upload-link.com", "method": "PUT"},
        status=200
    )
    # Mock PUT for actual upload - fail twice, then succeed
    responses.add(responses.PUT, "https://upload-link.com", status=500)
    responses.add(responses.PUT, "https://upload-link.com", status=500)
    responses.add(responses.PUT, "https://upload-link.com", status=201)

    # We need to monkeypatch time.sleep to avoid waiting during tests
    import time
    from unittest.mock import patch
    with patch("time.sleep", return_value=None):
        upload_to_cloud(config, "test.txt", b"content")

    # 1 GET + 3 PUT attempts
    assert len(responses.calls) == 4

@responses.activate
def test_upload_to_cloud_all_retries_fail(config):
    responses.add(
        responses.GET,
        "https://cloud-api.yandex.net/v1/disk/resources/upload/",
        json={"href": "https://upload-link.com", "method": "PUT"},
        status=200
    )
    responses.add(responses.PUT, "https://upload-link.com", status=500)

    import time
    from unittest.mock import patch
    with patch("time.sleep", return_value=None):
        with pytest.raises(Exception, match="upload failed after 3 retries"):
            upload_to_cloud(config, "test.txt", b"content")

def test_upload_to_cloud_empty_content(config):
    with pytest.raises(ValueError, match="file content is empty"):
        upload_to_cloud(config, "test.txt", b"")
