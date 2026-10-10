.PHONY: validate

validate: ## Merge data/, check it against schema/dataset.schema.json and check references
	python3 scripts/validate.py
