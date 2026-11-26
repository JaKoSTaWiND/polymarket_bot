.PHONY: run get-api-keys

run:
	@set -ea; \
	if [ -f .env.dev ]; then . ./.env.dev; fi; \
	GOWORK=off exec go run ./cmd/bot

get-api-keys: .venv/.clob-deps
	@.venv/bin/python scripts/get_api_keys.py

.venv/.clob-deps: Makefile
	python3 -m venv .venv
	.venv/bin/python -m pip install --disable-pip-version-check polymarket-client==0.20.1 python-dotenv==1.2.2
	@touch $@
