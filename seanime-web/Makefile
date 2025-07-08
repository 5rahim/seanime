build-web-only:
	rm -rf ../web
	npm run build
	cp -r out ../web

build-web:
	rm -rf ../web
	rm -rf ../web-desktop
	npm run build
	cp -r out ../web
	npm run build:desktop
	cp -r out-desktop ../web-desktop

build-denshi:
	rm -rf ../web-denshi
	npm run build:denshi
	cp -r out-denshi ../web-denshi
	rm -rf ../seanime-denshi/web-denshi
	cp -r ../web-denshi ../seanime-denshi/web-denshi

.PHONY: build-web build-denshi
