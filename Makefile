gen-metaplex:
	go build && rm -rf ./generated/metaplex_token_metadata && \
	./anchor-go -program-id metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s \
	-type-id uint8 \
	-src idl/metaplex/token-metadata-1.14.0.json \
	-dst ./generated/metaplex_token_metadata \
	-pkg metaplex-token-metadata
