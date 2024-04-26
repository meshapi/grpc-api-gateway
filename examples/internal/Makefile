start:
	@go run ./server

clean:
	@rm -fr api
	@rm -fr gen

generate:
	@rm -fr vendor && mkdir vendor
	@cp -r ../../api/meshapi vendor/
	@buf generate proto

install:
	@make -C ../.. install
