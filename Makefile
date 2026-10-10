.PHONY: validate test-go

validate: ## Merge data/, check it against schema/dataset.schema.json and check references
	python3 scripts/validate.py

test-go: ## Run the Go dataset loader tests
	go test ./go/...
