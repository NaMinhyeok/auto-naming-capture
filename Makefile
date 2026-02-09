.PHONY: build build-ocr build-go run clean

# Build everything
build: build-ocr build-go
	@echo "Build complete!"

# Build Swift OCR helper
build-ocr:
	@echo "Building OCR helper..."
	swiftc -O -o ocr-helper/ocr-helper ocr-helper/main.swift \
		-framework Vision -framework CoreGraphics -framework ImageIO

# Build Go binary
build-go:
	@echo "Building Go binary..."
	CGO_ENABLED=1 go build -o auto-naming-capture .

# Run the app
run: build
	./auto-naming-capture

# Clean build artifacts
clean:
	rm -f auto-naming-capture
	rm -f ocr-helper/ocr-helper
