APP_NAME := stundenzettel-generator
APP_ID := io.github.matousjobanek.stundenzettel
FYNE_CROSS_ENGINE ?= podman
FYNE_CROSS_NAME := Stundenzettel-Generator
FYNE_CROSS_EXTRA ?=
FYNE_CROSS_FLAGS := -engine $(FYNE_CROSS_ENGINE) -app-id $(APP_ID) -icon Icon.png -name $(FYNE_CROSS_NAME) -env GOTOOLCHAIN=auto $(FYNE_CROSS_EXTRA)

SECRETS_DIR := .secrets
KEYSTORE := $(SECRETS_DIR)/android-release.jks
KEYSTORE_ENV := $(SECRETS_DIR)/github-secrets.env
KEYSTORE_B64 := $(SECRETS_DIR)/android-release.jks.b64
KEYSTORE_ALIAS := stundenzettel
KEYSTORE_DNAME := CN=Stundenzettel-Generator, O=matousjobanek, C=AT

.PHONY: android-keystore build clean run tidy fmt package-windows package-darwin package-linux package-android package-android-signed package-all release

# Creates a release keystore in .secrets/ (gitignored). Does not overwrite.
# Copy the four values from github-secrets.env into GitHub → Settings → Secrets and variables → Actions.
android-keystore:
	@if [ -e "$(KEYSTORE)" ]; then \
		echo "Refusing to overwrite $(KEYSTORE) — that would break APK updates. Back it up and delete it only if you intend a new app identity."; \
		exit 1; \
	fi
	@command -v keytool >/dev/null || { echo "keytool not found (install a JRE/JDK)."; exit 1; }
	@mkdir -p "$(SECRETS_DIR)"
	@umask 077; \
	set -e; \
	storepass=$$(openssl rand -base64 24 | tr -d '/+=' | head -c 32); \
	keytool -genkeypair -noprompt \
		-keystore "$(KEYSTORE)" \
		-storetype PKCS12 \
		-alias "$(KEYSTORE_ALIAS)" \
		-keyalg RSA -keysize 2048 -validity 10000 \
		-storepass "$$storepass" -keypass "$$storepass" \
		-dname "$(KEYSTORE_DNAME)"; \
	base64 -w0 "$(KEYSTORE)" > "$(KEYSTORE_B64)"; \
	printf '%s\n' \
		"ANDROID_KEYSTORE_BASE64=<paste contents of $(KEYSTORE_B64)>" \
		"ANDROID_KEY_ALIAS=$(KEYSTORE_ALIAS)" \
		"ANDROID_KEYSTORE_PASSWORD=$$storepass" \
		"ANDROID_KEY_PASSWORD=$$storepass" \
		> "$(KEYSTORE_ENV)"
	@chmod 600 "$(KEYSTORE)" "$(KEYSTORE_ENV)" "$(KEYSTORE_B64)"
	@echo
	@echo "Wrote $(KEYSTORE)"
	@echo "Wrote $(KEYSTORE_ENV) and $(KEYSTORE_B64)"
	@echo "Back these files up offline (password manager / encrypted copy) before the first signed APK."
	@echo
	@echo "GitHub → repo → Settings → Secrets and variables → Actions → New repository secret:"
	@echo "  ANDROID_KEYSTORE_BASE64   (one line from $(KEYSTORE_B64))"
	@echo "  ANDROID_KEY_ALIAS         (from $(KEYSTORE_ENV))"
	@echo "  ANDROID_KEYSTORE_PASSWORD (from $(KEYSTORE_ENV))"
	@echo "  ANDROID_KEY_PASSWORD      (from $(KEYSTORE_ENV))"

fmt:
	gofmt -w .
	go mod tidy

build:
	go build -o $(APP_NAME) .

run: build
	./$(APP_NAME)

tidy:
	go mod tidy

package-windows:
	fyne-cross windows $(FYNE_CROSS_FLAGS)

package-darwin:
	fyne-cross darwin $(FYNE_CROSS_FLAGS)

package-linux:
	fyne-cross linux $(FYNE_CROSS_FLAGS)

package-android:
	fyne-cross android $(FYNE_CROSS_FLAGS)

package-android-signed:
	@test -f "$(KEYSTORE)" && test -f "$(KEYSTORE_ENV)" || { echo "run make android-keystore first"; exit 1; }
	@storepass=$$(grep '^ANDROID_KEYSTORE_PASSWORD=' "$(KEYSTORE_ENV)" | cut -d= -f2-); \
	keypass=$$(grep '^ANDROID_KEY_PASSWORD=' "$(KEYSTORE_ENV)" | cut -d= -f2-); \
	alias=$$(grep '^ANDROID_KEY_ALIAS=' "$(KEYSTORE_ENV)" | cut -d= -f2-); \
	fyne-cross android $(FYNE_CROSS_FLAGS) \
		-keystore "$(KEYSTORE)" \
		-keystore-pass "$$storepass" \
		-key-pass "$$keypass" \
		-key-name "$$alias"

package-all: package-windows package-darwin package-linux

# Bump FyneApp.toml, commit, and tag locally. Example: make release VERSION=1.0.1
release:
ifndef VERSION
	$(error VERSION=X.Y.Z is required, e.g. make release VERSION=1.0.1)
endif
	./scripts/release.sh $(VERSION)

clean:
	rm -f $(APP_NAME)
	rm -rf fyne-cross
