build-web:
	npm run build
	npm run build:desktop

move-web-dist:
	rm -rf ../web
	cp -r out ../web
	rm -rf ../web-desktop
	cp -r out-desktop ../web-desktop

.PHONY: build-web move-web-dist
