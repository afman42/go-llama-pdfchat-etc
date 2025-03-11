.ONESHELL:

run/api:
	@echo "Run API";
	@go run main.go;
	@echo "Stop Running API";

run/web:
	@echo "Run Web";
	@cd web; npm run dev;
	@echo "Stop Running Web";

run/preview_linux: 
	# Prevent error compile linux: pattern dist embed is not found
	@make build/web-staging; 
	@make build/linux;
	@make preview/api_linux;

build/rmfldr:
	@echo "Remove Folder";
	rm -rf ./bin/;
	@echo "Finish Remove Folder";

build/linux:
	@echo "Build Binary linux";
	GOOS=linux GOARCH=amd64 go build -ldflags="-s" -o=./bin/linux_amd64/tmp/app main.go;
	@echo "Build Binary Done";

build/web:
	@echo "Build Dist Web";
	cd web/; rm -rf dist/;
	cd web/; npm run build;
	@echo "Build Dist Web Done";

build/web-staging:
	@echo "Build Dist Web Staging";
	cd ./web/; rm -rf dist/;
	cd ./web/; npm run build:staging;
	@echo "Build Dist Web Staging Done";

build/compress_linux:
	@echo "Start Compress file linux";
	./upx ./bin/linux_amd64/tmp/app -o  ./bin/linux_amd64/app;
	@echo "Finish Compress file linux";

build: build/web build/rmfldr build/linux build/compress_linux; 

deploy:
	caprover deploy -h $$CAPROVER_HOST -p $$CAPROVER_PASSWORD -t deploy.tar -a $$CAPROVER_APP_NAME_GO_LLAMA -n $$CAPROVER_MACHINE_NAME;

deploy/tar:
	rm -f deploy.tar;
	tar -zcvf deploy.tar ./bin/linux_amd64/app ./web/dist/ Dockerfile captain-definition .env.prod ./pdf2txt;

deploy/prod: build deploy/tar deploy

#npmi l="" || npmi l="lib lib"
npmi:
	@echo "Install lib";
	@cd web; 
	@npm install $$l;
	@echo "Finish Install Lib: $l";

#npmu u="" || npmu u="lib lib"
npmu:
	@echo "Uninstall lib";
	@cd web; npm uninstall $$u;
	@echo "Finish Uninstall Lib: $u";

preview/api_linux:
	@echo "Preview";
	./bin/linux_amd64/tmp/app -mode preview;
