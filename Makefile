install: format
	go install github.com/tchap/git-trunk
	go install github.com/tchap/git-trunk/bin/hooks/commit-msg

format:
	go fmt ./...