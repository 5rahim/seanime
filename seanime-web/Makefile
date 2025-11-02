build-web-only:
	rm -rf ../web
	npm run build
	cp -r out ../web

build-web:
	rm -rf ../web
	rm -rf ../web-denshi
	npm run build
	cp -r out ../web
	npm run build:denshi
	cp -r out-denshi ../web-denshi

build-denshi:
	rm -rf ../web-denshi
	npm run build:denshi
	cp -r out-denshi ../web-denshi
	rm -rf ../seanime-denshi/web-denshi
	cp -r ../web-denshi ../seanime-denshi/web-denshi

.PHONY: build-web build-denshi
