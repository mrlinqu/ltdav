# Директория, в которой хранятся исполняемые
# файлы проекта и зависимости, необходимые для сборки.
LOCAL_BIN := $(CURDIR)/bin

ifndef BIN_DIR
BIN_DIR = $(LOCAL_BIN)
endif

# Версия Go. Используется для проброса в LDFLAGS и проверках.
GO_MAJOR_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
GO_PATCH_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f3)

# Полная версия Go, подставляемая в LDFLAGS (название историческое).
GO_VERSION_SHORT := $(GO_MAJOR_VERSION).$(GO_MINOR_VERSION).$(GO_PATCH_VERSION)

# Минимально поддерживаемая версия Go.
GO_MIN_SUPPORTED_MAJOR_VERSION := 1
GO_MIN_SUPPORTED_MINOR_VERSION := 20

# App version is sanitized CI branch name, if available.
# Otherwise git branch or commit hash is used.
APP_VERSION := $(if $(CI_COMMIT_REF_SLUG),$(CI_COMMIT_REF_SLUG),$(if $(GIT_BRANCH),$(GIT_BRANCH),$(GIT_HASH)))

# CI_PROJECT_ID is set in Gitlab CI, but we will fallback to app name if not present
CI_PROJECT_ID ?= catalog-api

# Короткий хеш коммита.
GIT_HASH := $(shell git log --format="%h" --max-count 1 2> /dev/null)

# Последнее коммит-сообщение в логе, завернутое в base64 для передачи с помощью ldflags.
GIT_LOG := $(shell git log --decorate --oneline --max-count 1 2> /dev/null | base64 | tr -d '\n')

# Текущая ветка в git, в которой происходит сборка приложения.
GIT_BRANCH := $(shell git branch --show-current)

# Момент времени, в который было собрано приложение
BUILD_TS := $(shell date +%FT%T%z)

# Переменные, переопределяемые в приложении на этапе сборки.
# https://pkg.go.dev/cmd/go@go1.21.1#hdr-Compile_packages_and_dependencies
LDFLAGS = \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.Name=catalog-api' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.ProjectID=$(CI_PROJECT_ID)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.Version=$(APP_VERSION)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.GoVersion=$(GO_VERSION_SHORT)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.BuildDate=$(BUILD_TS)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.GitLog=$(GIT_LOG)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.GitHash=$(GIT_HASH)' \
    -X 'github.com/mrlinqu/ltdav/internal/config/app.GitBranch=$(GIT_BRANCH)'

.build: .validate-min-go-version
	$(BUILD_ENVPARMS) go build -o="$(BIN_DIR)/" -ldflags "$(LDFLAGS)" ./cmd/ltdav

build: .build ## Запустить сборку приложения

.run: .validate-min-go-version
	go run \
		-ldflags "$(LDFLAGS)" \
		./cmd/ltdav

run: .run ## Запустить приложение локально в режиме разработки

# Шорткат для golangci-lint
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint

# Требуемая актуальная версия линтера.
GOLANGCI_TAG ?= 1.55.2

# Поиск локальной версии golangci-lint.
ifneq ($(shell command -v $(GOLANGCI_BIN)),)
# Найден локально установленный golangci-lint.
GOLANGCI_BIN_VERSION := $(shell $(GOLANGCI_BIN) --version 2> /dev/null)

# Парсинг версии локально установленного golangci-lint.
ifneq ($(GOLANGCI_BIN_VERSION),)
GOLANGCI_BIN_VERSION_SHORT := $(shell echo "$(GOLANGCI_BIN_VERSION)" | sed -En 's/.*([0-9]+\.[0-9]+\.[0-9]+) built.*/\1/p' | sed -E 's/v(.*)/\1/g') # golangci-lint, в зависимости от версии, может отдавать версию как с 'v' так и без
else
# Не получилось узнать версию локально установленного golangci-lint.
GOLANGCI_BIN_VERSION_SHORT := 0
endif

# Сортируем версии. Если версия локально установленного golangci-lint
# совпадает с требуемой - используем локальную версию.
ifneq ("$(GOLANGCI_TAG)", "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_BIN_VERSION_SHORT)))")
GOLANGCI_BIN :=
endif

endif

# Проверка глобального golangci-lint.
ifneq ($(shell command -v golangci-lint 2> /dev/null),)
# Найден глобально установленный golangci-lint.
GOLANGCI_VERSION := $(shell golangci-lint --version 2> /dev/null)

# Парсинг версии глобально установленного golangci-lint.
ifneq ($(GOLANGCI_VERSION),)
GOLANGCI_VERSION_SHORT := $(shell echo "$(GOLANGCI_VERSION)" | sed -En 's/.*([0-9]+\.[0-9]+\.[0-9]+) built.*/\1/p' | sed -E 's/v(.*)/\1/g') # golangci-lint, в зависимости от версии, может отдавать версию как с 'v' так и без
else
# Не получилось узнать версию глобально установленного golangci-lint.
GOLANGCI_VERSION_SHORT := 0
endif

# Сортируем версии. Если версия глобально установленного golangci-lint
# совпадает с требуемой - используем глобальную версию.
# В противном случае будет произведена установка golangci-lint
# из исходников в локальную директорию с исполняемыми файлами.
ifeq ("$(GOLANGCI_TAG)", "$(word 1, $(sort $(GOLANGCI_TAG) $(GOLANGCI_VERSION_SHORT)))")
GOLANGCI_BIN := $(shell command -v golangci-lint 2> /dev/null)
endif

endif

install-lint: export GOBIN := $(LOCAL_BIN)
install-lint: ## Установить golangci-lint в текущую директорию с исполняемыми файлами
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	@# Если в предыдущих шагах не был найден исполняемый
	@# файл линтера - устанавливаем его из исходников локально.
	$(info Installing golangci-lint v$(GOLANGCI_TAG))
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
# Устанавливаем текущий путь для исполняемого файла линтера.
GOLANGCI_BIN := $(LOCAL_BIN)/golangci-lint
else
	$(info Golangci-lint is already installed to $(GOLANGCI_BIN))
endif

.lint: install-lint
	$(info Running lint against changed files...)
	$(GOLANGCI_BIN) run \
		--new-from-rev=origin/master \
		--config=.golangci.pipeline.yaml \
		./...

lint: .lint ## Запуск golangci-lint на изменениях, отличающихся от мастера

.lint-full: install-lint
	$(info Running lint against all project files...)
	$(GOLANGCI_BIN) run \
		--config=.golangci.pipeline.yaml \
		./...

lint-full: .lint-full ## Запуск golangci-lint по всем файлам проекта

.deps:
	$(info Installing dependencies...)
	go mod download

deps: .deps ## Установить зависимости (go mod download)

# Валидация минимально поддерживаемой версии Go.
.validate-min-go-version:
	$(call _validate_go_version_func,\
		$(GO_MIN_SUPPORTED_MAJOR_VERSION),\
		$(GO_MIN_SUPPORTED_MINOR_VERSION))

# Объявляем, что текущие команды не являются файлами и
# инструктируем Makefile не искать изменений в файловой системе.
.PHONY: \
	all \
	.install-lint \
	install-lint \
	.lint \
	.validate-min-go-version \
	lint \
	.lint-full \
	lint-full \
	.deps \
	deps \
	.test \
	test \
	.build \
	build \
	.run \
	run
