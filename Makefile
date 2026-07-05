.PHONY: run build clean frontend

frontend:
	cd web && npm run build
	rm -rf internal/webui/dist
	cp -r web/dist internal/webui/

build: frontend
	go build -o bin/pagebound ./cmd/server

run: build
	./bin/pagebound

clean:
	rm -rf bin/pagebound internal/webui/dist/
