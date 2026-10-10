.PHONY: validate test-go test-typescript

validate: ## Merge data/, check it against schema/dataset.schema.json and check references
	python3 scripts/validate.py

test-go: ## Run the Go dataset loader tests
	go test ./go/...

test-typescript: ## Run the TypeScript dataset loader tests
	npm --prefix typescript ci
	npm --prefix typescript test
