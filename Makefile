# Makefile for cw-ica-controller

optimize:
	@echo "Compiling optimized wasm..."
	@docker run --rm -v "$(shell pwd)":/code \
		--mount type=volume,source="$(shell basename "$(shell pwd)")_cache",target=/code/target \
		--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
		cosmwasm/optimizer:0.15.1

.PHONY: optimize
