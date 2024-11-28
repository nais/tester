generate:
	cd ./internal/webui/ui && npm i && npm run build
	cp -r ./internal/webui/ui/dist/* ./internal/webui/static/
