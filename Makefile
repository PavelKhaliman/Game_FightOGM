.PHONY: run build test smoke

ifeq ($(OS),Windows_NT)
build:
	powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/build.ps1
run:
	go run ./cmd/fightogm
else
build:
	go test ./...
	go build ./...
	mkdir -p dist/assets/models dist/assets/audio
	go build -trimpath -o dist/FightOGM ./cmd/fightogm
	cp assets/models/*.glb assets/models/special_animation_manifest.json dist/assets/models/
	cp README.md THIRD_PARTY_NOTICES.md dist/
	cp -R licenses dist/
run:
	go run ./cmd/fightogm
endif

test:
	go test ./...

smoke:
	go run ./cmd/fightogm -smoke-test
