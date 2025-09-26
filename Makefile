ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

# 版本信息变量
APP_NAME    = "edge-device"
GIT_COMMIT  = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH  = $(shell git branch --show-current 2>/dev/null || echo "unknown")
BUILD_TIME  = $(shell date '+%Y-%m-%d %H:%M:%S')
GO_VERSION  = $(shell go version | cut -d' ' -f3)

# 编译时的ldflags
LDFLAGS = -X 'demo/internal/consts.BuildTime=$(BUILD_TIME)' \
          -X 'demo/internal/consts.GitCommit=$(GIT_COMMIT)' \
          -X 'demo/internal/consts.GitBranch=$(GIT_BRANCH)' \
          -X 'demo/internal/consts.GoVersion=$(GO_VERSION)'

# 构建目标 - 自动检测平台
build:
ifeq ($(OS),Windows_NT)
	@echo "🔨 Building $(APP_NAME) for Windows with version info..."
	@go build -ldflags "$(LDFLAGS)" -o $(APP_NAME).exe .
	@echo "✅ Build completed: $(APP_NAME).exe"
else
	@echo "🔨 Building $(APP_NAME) for Linux with version info..."
	@go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .
	@echo "✅ Build completed: $(APP_NAME)"
endif

# Windows构建目标
build-windows:
	@echo "🔨 Building $(APP_NAME) for Windows with version info..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME).exe .
	@echo "✅ Build completed: $(APP_NAME).exe"

# Linux构建目标
build-linux:
	@echo "🔨 Building $(APP_NAME) for Linux with version info..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) .
	@echo "✅ Build completed: $(APP_NAME)"

# 运行程序
run: build
ifeq ($(OS),Windows_NT)
	@echo "🚀 Running $(APP_NAME)..."
	@./$(APP_NAME).exe
else
	@echo "🚀 Running $(APP_NAME)..."
	@./$(APP_NAME)
endif

# 清理构建文件
clean:
	@echo "🧹 Cleaning build files..."
	@if exist $(APP_NAME).exe del $(APP_NAME).exe
	@if exist $(APP_NAME) del $(APP_NAME)
	@echo "✅ Clean completed"

include ./hack/hack-cli.mk
include ./hack/hack.mk
