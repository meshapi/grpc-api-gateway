start:
	@go run ./server

clean:
	@rm -fr api

generate:
	@rm -fr vendor && mkdir vendor
	@cp -r ../../api/meshapi vendor/
	@buf generate proto

install:
	@make -C ../.. install
