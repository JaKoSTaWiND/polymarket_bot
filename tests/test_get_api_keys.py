import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from dotenv import dotenv_values
from polymarket import ApiKeyCreds

from scripts.get_api_keys import save_env_files


class SaveEnvFilesTest(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.root = Path(temp.name)
        self.dev = self.root / ".env.dev"
        self.prod = self.root / ".env.prod"
        self.credentials = ApiKeyCreds(key="api-key", secret="secret==", passphrase="passphrase")

    def test_updates_both_files_and_preserves_environment_settings(self):
        for path, stage in [(self.dev, "DEV"), (self.prod, "PROD")]:
            path.write_text(f"# settings\nSTAGE={stage}\nPOLY_SECRET=old\n")

        save_env_files(self.root, self.credentials)
        save_env_files(self.root, self.credentials)

        for path, stage in [(self.dev, "DEV"), (self.prod, "PROD")]:
            self.assertEqual(dotenv_values(path), {
                "STAGE": stage,
                "POLY_API_KEY": "api-key",
                "POLY_SECRET": "secret==",
                "POLY_PASSPHRASE": "passphrase",
            })
            self.assertIn("# settings\n", path.read_text())
            self.assertEqual(path.read_text().count("POLY_SECRET="), 1)
            self.assertEqual(path.stat().st_mode & 0o777, 0o600)

    def test_creates_missing_production_file(self):
        self.dev.write_text("STAGE=DEV\n")
        save_env_files(self.root, self.credentials)
        self.assertEqual(set(dotenv_values(self.prod)), {
            "POLY_API_KEY", "POLY_SECRET", "POLY_PASSPHRASE",
        })

    def test_make_shell_reads_passphrase_literally(self):
        credentials = ApiKeyCreds(key="key", secret="secret", passphrase="bot's $account")
        save_env_files(self.root, credentials)
        value = subprocess.check_output(
            ["sh", "-c", '. ./.env.dev && printf %s "$POLY_PASSPHRASE"'],
            cwd=self.root, text=True, stderr=subprocess.DEVNULL,
        )
        self.assertEqual(value, credentials.passphrase)

    def test_restores_development_file_if_production_replace_fails(self):
        self.dev.write_text("POLY_API_KEY=old-dev\n")
        self.prod.write_text("POLY_API_KEY=old-prod\n")
        replace = os.replace

        def fail_prod(source, target):
            if target == self.prod:
                raise PermissionError("test failure")
            return replace(source, target)

        with patch("scripts.get_api_keys.os.replace", side_effect=fail_prod):
            with self.assertRaises(PermissionError):
                save_env_files(self.root, self.credentials)

        self.assertEqual(self.dev.read_text(), "POLY_API_KEY=old-dev\n")
        self.assertEqual(self.prod.read_text(), "POLY_API_KEY=old-prod\n")
