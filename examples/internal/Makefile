start:
	@go run ./server

clean:
	@rm -fr api

generate:
	@rm -fr vendor && mkdir vendor
	@cp -rl ../../api/meshapi vendor/
	@buf generate proto

install:
	@make -C ../.. install
