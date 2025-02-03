.ONESHELL:

run/api:
	@echo "Run API";
	@go run main.go;
	@echo "Stop Running API";

run/web:
	@echo "Run Web";
	@cd web; npm run dev;
	@echo "Stop Running Web";

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
