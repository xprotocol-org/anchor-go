gen-metaplex:
	go build && rm -rf ./metaplex && \
	./anchor-go -program-id metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s \
	-type-id uint8 \
	-src idl/metaplex/token-metadata.json \
	-dst ./generated/metaplex \
	-pkg metaplex
