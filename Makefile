VERSION ?= dev
LDFLAGS := -s -w -X tvpkg/cmd.version=$(VERSION)
GO      ?= go

.PHONY: all build vet test install clean release

all: build

build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o tvpkg .

install: build
	install -Dm755 tvpkg $(PREFIX)/bin/tvpkg
	ln -sf tvpkg $(PREFIX)/bin/tvp

vet:
	@test -z "$$(gofmt -l .)" || (gofmt -l .; exit 1)
	$(GO) vet ./...

test:
	$(GO) test ./...

# Cross-compiles tvpkg for the four Termux arches. Requires an Android NDK:
# export NDK_BIN=<ndk>/toolchains/llvm/prebuilt/linux-x86_64/bin (aarch64 also
# builds fine with CGO disabled; the other arches need NDK clang for external
# linking, see .github/workflows/release.yml).
release:
	@set -eu; \
	BIN="$${NDK_BIN:-}"; \
	export GOOS=android CGO_ENABLED=1; \
	build() { goarch=$$1; shift; goarm=$$1; shift; g386=$$1; shift; triple=$$1; \
		[ -n "$$goarm" ] && export GOARM=$$goarm || unset GOARM; \
		[ -n "$$g386" ] && export GO386=$$g386 || unset GO386; \
		[ -n "$$BIN" ] && export CC="$$BIN/$${triple}24-clang" || unset CC; \
		echo "== $$triple ($goarch)"; \
		$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o /data/data/com.termux/files/usr/tmp/tvpkg-$$goarch .; }; \
	build arm64 "" "" aarch64-linux-android; \
	build arm 7 "" armv7a-linux-androideabi; \
	build amd64 "" "" x86_64-linux-android; \
	build 386 "" sse2 i686-linux-android

clean:
	rm -f tvpkg
	rm -rf stage dist