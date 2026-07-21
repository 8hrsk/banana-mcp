APP_NAME := banana-mcp
BUILD_DIR := release

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: all clean build release $(PLATFORMS)

all: clean release

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(APP_NAME)

build:
	go build -o $(APP_NAME) .

release: $(PLATFORMS)

$(PLATFORMS):
	@mkdir -p $(BUILD_DIR)/$(os)-$(arch)
	@echo "Building for $(os)/$(arch)..."
	GOOS=$(os) GOARCH=$(arch) go build -o $(BUILD_DIR)/$(os)-$(arch)/$(APP_NAME)$(if $(filter windows,$(os)),.exe,) .
