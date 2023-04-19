bin: $(shell find . -name 'webui' -prune -o -type f -name *.go) webui
	env GOOS=linux GOARCH=amd64 go build && gzip -k -f loghub

webui: $(shell find webui -name 'node_modules' -prune -o  -name 'dist' -prune -o -type f)
	pushd webui && (yarn build; popd)
