#
# Display help.
#
.PHONY: help
help:
	@echo 'Available tasks:'
	@echo '  make help     -- Display this help message'
	@echo '  make generate -- Run code generation'
	@echo '  make build    -- Build the `ktnh` binary'
	@echo '  make clean    -- Remove build artifacts'
	@echo '  make test     -- Run tests'
	@echo '  make license  -- Collect license information and save it to `./licenses/`'
	@echo '  make release  -- Release the ktnh binary'

#
# Run `go generate`.
#
.PHONY: generate
generate:
	@echo 'Generating code...'

	go generate ./internal/pkg/cfn/gen/

#
# Build the binary.
#
.PHONY: build
build: clean
	@echo 'Building ktnh binary...'

	go build \
		-ldflags "\
			-X github.com/quickguard-oss/koreru-toki-no-hiho/cmd.version=$$( git describe --tags --always --dirty 2> /dev/null || echo 'dev' ) \
			-X github.com/quickguard-oss/koreru-toki-no-hiho/cmd.commit=$$( git rev-parse --short 'HEAD' 2> /dev/null || echo 'HEAD' ) \
			-X github.com/quickguard-oss/koreru-toki-no-hiho/cmd.built=$$( date -u '+%FT%TZ' ) \
			-X github.com/quickguard-oss/koreru-toki-no-hiho/cmd.versionOverridden=true \
		" \
		-v \
		-o ktnh ./

#
# Remove build artifacts.
#
.PHONY: clean
clean:
	@echo 'Cleaning build artifacts...'

	rm -rf \
	  ./ktnh \
	  ./dist/ \
	  ./licenses/

#
# Run tests.
#
.PHONY: test
test: generate
	@echo 'Running tests...'

	go test -v ./...

#
# Collect license information.
#
.PHONY: license
license:
	@echo 'Collecting license information...'

	go tool go-licenses save ./ \
	  --force \
	  --save_path ./licenses/

#
# Release the binary.
#
.PHONY: release
release: clean license
	@echo 'Releasing ktnh binary...'

	go tool goreleaser release --clean
