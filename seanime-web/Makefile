build-web:
	rm -rf ../web
	rm -rf ../web-desktop
	npm run build
	cp -r out ../web
	npm run build:desktop
	cp -r out-desktop ../web-desktop

.PHONY: build-web
