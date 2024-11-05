generate:
	cd ./internal/webui/ui && npm run build
	cp -r ./internal/webui/ui/dist/* ./internal/webui/static/
