import fcntl
import os
import re
import shutil
import sys
from contextlib import ExitStack
from getpass import getpass
from pathlib import Path
from tempfile import TemporaryDirectory

from dotenv import set_key
from eth_account import Account
from polymarket import ApiKeyCreds, SecureClient
from polymarket.errors import PolymarketError


def save_env_files(root: Path, credentials: ApiKeyCreds) -> None:
    values = {
        "POLY_API_KEY": credentials.key,
        "POLY_SECRET": credentials.secret,
        "POLY_PASSPHRASE": credentials.passphrase,
    }
    if not all(values.values()):
        raise ValueError("Polymarket вернул неполный набор CLOB-ключей.")

    dev, prod = root / ".env.dev", root / ".env.prod"
    with ExitStack() as stack:
        directory = os.open(root, os.O_RDONLY | os.O_DIRECTORY)
        stack.callback(os.close, directory)
        fcntl.flock(directory, fcntl.LOCK_EX)
        staging = Path(stack.enter_context(TemporaryDirectory(dir=root, prefix=".clob-keys-")))

        for target in (dev, prod):
            staged = staging / target.name
            if target.exists():
                shutil.copyfile(target, staged)
            else:
                staged.touch(mode=0o600)
            for name, value in values.items():
                escaped = re.sub(r'(["\\$`])', r'\\\1', value)
                set_key(staged, name, f'"{escaped}"', quote_mode="never")
            staged.chmod(0o600)

        backup = staging / "original-dev"
        if dev.exists():
            shutil.copy2(dev, backup)
        os.replace(staging / dev.name, dev)
        try:
            os.replace(staging / prod.name, prod)
        except OSError:
            if backup.exists():
                os.replace(backup, dev)
            else:
                dev.unlink()
            raise


def main() -> None:
    if not sys.stdin.isatty():
        raise SystemExit("Запусти make get-api-keys в интерактивном терминале.")
    private_key = getpass("Приватный ключ кошелька Polymarket (ввод скрыт): ").strip()
    try:
        signer = Account.from_key(private_key)
    except (ValueError, TypeError):
        raise SystemExit("Некорректный приватный ключ кошелька.") from None

    try:
        with SecureClient.create(private_key=private_key, wallet=signer.address) as client:
            credentials = client.credentials
    except PolymarketError as exc:
        status = getattr(exc, "status", None)
        detail = f"HTTP {status}" if status is not None else type(exc).__name__
        raise SystemExit(f"Не удалось получить CLOB-ключи ({detail}).") from None

    save_env_files(Path.cwd(), credentials)
    print("CLOB-ключи записаны в .env.dev и .env.prod.")


if __name__ == "__main__":
    main()
